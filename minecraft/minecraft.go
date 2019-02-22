package minecraft

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-querystring/query"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/svanharmelen/jsonapi"
	"golang.org/x/time/rate"
)

const (
	//userAgent = "go-minecraft"

	// DefaultAddress of Minecraft Server plugin.
	DefaultAddress = "http://localhost:8080"
	// DefaultBasePath on which the API is served.
	DefaultBasePath = "/"
)

var (
	// ErrUnauthorized is returned when a receiving a 401.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrResourceNotFound is returned when a receiving a 404.
	ErrResourceNotFound = errors.New("resource not found")
)

// Config provides configuration details to the API client.
type Config struct {
	// The address of the Terraform Enterprise API.
	Address string

	// The base path on which the API is served.
	BasePath string

	// Headers that will be added to every request.
	Headers http.Header

	// A custom HTTP client to use.
	HTTPClient *http.Client
}

// DefaultConfig returns a default config structure.
func DefaultConfig() *Config {
	config := &Config{
		Address:    os.Getenv("MINECRAFT_ADDRESS"),
		BasePath:   DefaultBasePath,
		Headers:    make(http.Header),
		HTTPClient: cleanhttp.DefaultPooledClient(),
	}

	// Set the default address if none is given.
	if config.Address == "" {
		config.Address = DefaultAddress
	}

	// Set the default user agent.
	config.Headers.Set("Content-Type", "application/json")

	return config
}

// Client is the Terraform Enterprise API client. It provides the basic
// connectivity and configuration for accessing the TFE API.
type Client struct {
	baseURL *url.URL
	headers http.Header
	http    *http.Client
	limiter *rate.Limiter
}

// NewClient creates a new Terraform Enterprise API client.
func NewClient(cfg *Config) (*Client, error) {
	config := DefaultConfig()

	// Layer in the provided config for any non-blank values.
	if cfg != nil {
		if cfg.Address != "" {
			config.Address = cfg.Address
		}
		if cfg.BasePath != "" {
			config.BasePath = cfg.BasePath
		}
		for k, v := range cfg.Headers {
			config.Headers[k] = v
		}
		if cfg.HTTPClient != nil {
			config.HTTPClient = cfg.HTTPClient
		}
	}

	// Parse the address to make sure its a valid URL.
	baseURL, err := url.Parse(config.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %v", err)
	}

	baseURL.Path = config.BasePath
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}

	// Create the client.
	client := &Client{
		baseURL: baseURL,
		headers: config.Headers,
		http:    config.HTTPClient,
	}

	return client, nil
}

// newRequest creates an API request. A relative URL path can be provided in
// path, in which case it is resolved relative to the apiVersionPath of the
// Client. Relative URL paths should always be specified without a preceding
// slash.
// If v is supplied, the value will be JSONAPI encoded and included as the
// request body. If the method is GET, the value will be parsed and added as
// query parameters.
func (c *Client) newRequest(method, path string, v interface{}) (*http.Request, error) {
	u, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	var body io.Reader
	switch method {
	case "GET":
		if v != nil {
			q, err := query.Values(v)
			if err != nil {
				return nil, err
			}
			u.RawQuery = q.Encode()
		}
	case "DELETE", "PATCH", "POST":
		if v != nil {
			dat, _ := json.Marshal(v)
			log.Printf("[DEBUG] body: " + string(dat))
			body = bytes.NewReader(dat)
		}
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	// Set the default headers.
	for k, v := range c.headers {
		req.Header[k] = v
	}

	return req, nil
}

// do sends an API request and returns the API response. The API response
// is JSONAPI decoded and the document's primary data is stored in the value
// pointed to by v, or returned as an error if an API error has occurred.

// If v implements the io.Writer interface, the raw response body will be
// written to v, without attempting to first decode it.
//
// The provided ctx must be non-nil. If it is canceled or times out, ctx.Err()
// will be returned.
func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) error {
	// Add the context to the request.
	req = req.WithContext(ctx)
	log.Printf("[DEBUG] request: %v", req)
	// Execute the request and check the response.
	resp, err := c.http.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return err
		}
	}
	defer resp.Body.Close()
	log.Printf("[DEBUG] response: %v", resp)
	// Basic response checking.
	if err := checkResponseCode(resp); err != nil {
		return err
	}

	// Return here if decoding the response isn't needed.
	if v == nil {
		return nil
	}
	/*
		// If v implements io.Writer, write the raw response body.
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			return err
		}

		// Get the value of v so we can test if it's a struct.
		dst := reflect.Indirect(reflect.ValueOf(v))

		// Return an error if v is not a struct or an io.Writer.
		if dst.Kind() != reflect.Struct {
			return fmt.Errorf("v must be a struct or an io.Writer")
		}
		return jsonapi.UnmarshalPayload(resp.Body, v)*/
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	return json.Unmarshal(buf.Bytes(), v)
}

// checkResponseCode can be used to check the status code of an HTTP request.
func checkResponseCode(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode <= 299 {
		return nil
	}

	switch r.StatusCode {
	case 401:
		return ErrUnauthorized
	case 404:
		return ErrResourceNotFound
	}

	// Decode the error payload.
	errPayload := &jsonapi.ErrorsPayload{}
	err := json.NewDecoder(r.Body).Decode(errPayload)
	if err != nil || len(errPayload.Errors) == 0 {
		log.Printf("[DEBUG] resp status: %v", r.Status)
		return fmt.Errorf(r.Status)
	}

	// Parse and format the errors.
	var errs []string
	for _, e := range errPayload.Errors {
		if e.Detail == "" {
			errs = append(errs, e.Title)
		} else {
			errs = append(errs, fmt.Sprintf("%s\n\n%s", e.Title, e.Detail))
		}
	}

	return fmt.Errorf(strings.Join(errs, "\n"))
}
