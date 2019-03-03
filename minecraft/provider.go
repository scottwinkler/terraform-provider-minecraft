package minecraft

import (
	"context"
	"net/url"

	"github.com/hashicorp/terraform/helper/mutexkv"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	sdk "github.com/scottwinkler/go-minecraft"
)

// ctx is used as default context.Context when making API calls.
var ctx = context.Background()

//Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"hostname": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MINECRAFT_HOSTNAME", nil),
				Description: "Minecraft Hostname",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"minecraft_cube":     resourceMinecraftCube(),
			"minecraft_cylinder": resourceMinecraftCylinder(),
			"minecraft_entity":   resourceMinecraftEntity(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	// Parse the hostname for comparison,
	hostname, _ := d.Get("hostname").(string)

	// Get the full Minecraft service address.
	address, _ := url.Parse(hostname)

	// Create a new Minecraft client config
	cfg := &sdk.Config{
		Address: address.String(),
	}

	// Create a new Minecraft client.
	return sdk.NewClient(cfg)
}

// This is a global MutexKV for use within this plugin.
var minecraftMutexKV = mutexkv.NewMutexKV()
