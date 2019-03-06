package minecraft

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	sdk "github.com/scottwinkler/go-minecraft"
	"github.com/scottwinkler/terraform/helper/resource"
)

func resourceMinecraftCylinder() *schema.Resource {
	return &schema.Resource{
		Create: resourceMinecraftCylinderCreate,
		Read:   resourceMinecraftCylinderRead,
		Update: resourceMinecraftCylinderUpdate,
		Delete: resourceMinecraftCylinderDelete,

		Schema: map[string]*schema.Schema{
			"location": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"x": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"y": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"z": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"world": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"dimensions": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"height": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"radius": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"material": {
				Type:     schema.TypeString,
				Required: true,
			},
			"shape_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"previous_data": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceMinecraftCylinderCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*sdk.Client)

	// Create a context
	ctx := context.Background()

	// Create a new cylinder
	options := sdk.ShapeCreateOptions{
		Location:   deserializeLocation(d.Get("location")),
		ShapeType:  sdk.ShapeTypeCylinder,
		Material:   d.Get("material").(string),
		Dimensions: deserializeCylinderDimensions(d.Get("dimensions")),
	}

	shape, _ := conn.Shapes.Create(ctx, options)
	d.SetId(shape.ID)
	resourceMinecraftCylinderRead(d, meta)
	return nil
}

func resourceMinecraftCylinderRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*sdk.Client)

	// Create a context
	ctx := context.Background()

	id := d.Id()
	var shape *sdk.Shape
	var err error
	// Wait until resource is in a valid state
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		shape, err = conn.Shapes.Read(ctx, id)
		if err != nil {
			log.Printf("[DEBUG] Error reading Shape: %s", err)
			return resource.NonRetryableError(err)
		}
		if shape.Status != sdk.ResourceStatusReady {
			log.Printf("[DEBUG] Shape not in ready state. Current state is: %s", shape.Status)
			return resource.RetryableError(errors.New("invalid state"))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error reading Shape: %s", err)
	}

	d.Set("material", shape.Material)
	d.Set("location", serializeLocation(shape.Location))
	d.Set("dimensions", serializeCylinderDimensions(shape.Dimensions.(*sdk.CylinderDimensions)))
	d.Set("previous_data", shape.PreviousData)
	d.Set("shape_type", shape.ShapeType)
	return nil
}

func resourceMinecraftCylinderUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*sdk.Client)

	// Create a context
	ctx := context.Background()

	// Update cylinder to current settings
	options := sdk.ShapeUpdateOptions{
		Location:   deserializeLocation(d.Get("location")),
		ShapeType:  sdk.ShapeTypeCylinder,
		Material:   d.Get("material").(string),
		Dimensions: deserializeCylinderDimensions(d.Get("dimensions")),
	}

	id := d.Id()
	conn.Shapes.Update(ctx, id, options)
	return resourceMinecraftCylinderRead(d, meta)
}

func resourceMinecraftCylinderDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*sdk.Client)
	id := d.Id()

	// Create a context
	ctx := context.Background()

	conn.Shapes.Delete(ctx, id)

	// Wait until resource is finished deleting
	resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.Shapes.Read(ctx, id)
		// A 404 error indicates success
		if err != nil {
			if err == sdk.ErrResourceNotFound {
				return nil
			}
		}
		log.Printf("[DEBUG] Shape deleting...")
		return resource.RetryableError(errors.New("invalid state"))
	})
	return nil
}

func deserializeCylinderDimensions(dimensions interface{}) *sdk.CylinderDimensions {
	d := dimensions.([]interface{})[0].(map[string]interface{})
	h, _ := d["height"]
	height := h.(int)
	r, _ := d["radius"]
	radius := r.(int)
	return sdk.NewCylinderDimensions(height, radius)
}

func serializeCylinderDimensions(dimensions *sdk.CylinderDimensions) interface{} {
	d := make([]interface{}, 1)
	m := make(map[string]interface{})
	m["height"] = dimensions.Height
	m["radius"] = dimensions.Radius
	d[0] = m
	return d
}
