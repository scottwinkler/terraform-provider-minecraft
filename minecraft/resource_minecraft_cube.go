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

func resourceMinecraftCube() *schema.Resource {
	return &schema.Resource{
		Create: resourceMinecraftCubeCreate,
		Read:   resourceMinecraftCubeRead,
		Update: resourceMinecraftCubeUpdate,
		Delete: resourceMinecraftCubeDelete,

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
						"length_x": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"height_y": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"width_z": {
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

func resourceMinecraftCubeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*sdk.Client)

	// Create a context
	ctx := context.Background()

	// Create a new cube
	options := sdk.ShapeCreateOptions{
		Location:   deserializeLocation(d.Get("location")),
		ShapeType:  sdk.ShapeTypeCube,
		Material:   d.Get("material").(string),
		Dimensions: deserializeCubeDimensions(d.Get("dimensions")),
	}

	shape, err := conn.Shapes.Create(ctx, options)
	log.Printf("%v", err)
	log.Printf("%v", shape)
	d.SetId(shape.ID)
	resourceMinecraftCubeRead(d, meta)
	return nil
}

func resourceMinecraftCubeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*sdk.Client)

	// Create a context
	ctx := context.Background()

	id := d.Id()
	var shape *sdk.Shape
	// Wait until resource is in a valid state
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		shape, err := conn.Shapes.Read(ctx, id)
		if err != nil {
			log.Printf("[DEBUG] Error reading Shape: %s", err)
			return resource.NonRetryableError(err)
		}
		if shape.Status != sdk.ResourceStatusReady {
			log.Printf("[DEBUG] Shape not in ready state: %s", shape.Status)
			return resource.RetryableError(errors.New("invalid state"))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error reading Shape: %s", err)
	}

	d.Set("material", shape.Material)
	d.Set("location", serializeLocation(shape.Location))
	d.Set("dimensions", serializeCubeDimensions(shape.Dimensions.(*sdk.CubeDimensions)))
	d.Set("previous_data", shape.PreviousData)
	d.Set("shape_type", shape.ShapeType)
	return nil
}

func resourceMinecraftCubeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*sdk.Client)

	// Create a context
	ctx := context.Background()

	// Update cube to current settings
	options := sdk.ShapeUpdateOptions{
		Location:   deserializeLocation(d.Get("location")),
		ShapeType:  sdk.ShapeTypeCube,
		Material:   d.Get("material").(string),
		Dimensions: deserializeCubeDimensions(d.Get("dimensions")),
	}

	id := d.Id()
	conn.Shapes.Update(ctx, id, options)
	return resourceMinecraftCubeRead(d, meta)
}

func resourceMinecraftCubeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*sdk.Client)
	id := d.Id()

	// Create a context
	ctx := context.Background()

	conn.Shapes.Delete(ctx, id)
	return nil
}

func deserializeCubeDimensions(dimensions interface{}) *sdk.CubeDimensions {
	d := dimensions.([]interface{})[0].(map[string]interface{})
	x, _ := d["length_x"]
	lengthX := x.(int)
	y, _ := d["height_y"]
	heightY := y.(int)
	z, _ := d["width_z"]
	widthZ := z.(int)
	return sdk.NewCubeDimensions(lengthX, heightY, widthZ)
}

func serializeCubeDimensions(dimensions *sdk.CubeDimensions) interface{} {
	d := make([]interface{}, 1)
	m := make(map[string]interface{})
	m["length_x"] = dimensions.LengthX
	m["height_y"] = dimensions.HeightY
	m["width_z"] = dimensions.WidthZ
	d[0] = m
	return d
}
