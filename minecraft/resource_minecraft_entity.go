package minecraft

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	sdk "github.com/scottwinkler/go-minecraft"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func resourceMinecraftEntity() *schema.Resource {
	return &schema.Resource{
		Create: resourceMinecraftEntityCreate,
		Read:   resourceMinecraftEntityRead,
		Update: resourceMinecraftEntityUpdate,
		Delete: resourceMinecraftEntityDelete,

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
			"entity_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"custom_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
		},
	}
}

func resourceMinecraftEntityCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*sdk.Client)

	// Create a context
	ctx := context.Background()

	// Create a new entity
	options := sdk.EntityCreateOptions{
		Location:   deserializeLocation(d.Get("location")),
		EntityType: d.Get("entity_type").(string),
		CustomName: d.Get("custom_name").(string),
	}

	entity, _ := conn.Entities.Create(ctx, options)
	d.SetId(entity.ID)
	resourceMinecraftEntityRead(d, meta)
	return nil
}

func resourceMinecraftEntityRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*sdk.Client)

	// Create a context
	ctx := context.Background()

	id := d.Id()
	var entity *sdk.Entity
	var err error
	// Wait until resource is in a valid state
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		entity, err = conn.Entities.Read(ctx, id)
		if err != nil {
			log.Printf("[DEBUG] Error reading Entity: %s", err)
			return resource.NonRetryableError(err)
		}
		if entity.Status != sdk.ResourceStatusReady {
			log.Printf("[DEBUG] Entity not in ready state. Current state is: %s", entity.Status)
			return resource.RetryableError(errors.New("invalid state"))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error reading Entity: %s", err)
	}

	d.Set("location", serializeLocation(entity.Location))
	d.Set("custom_name", entity.CustomName)
	d.Set("entity_type", entity.EntityType)
	return nil
}

func resourceMinecraftEntityUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*sdk.Client)

	// Create a context
	ctx := context.Background()

	// Update entity to current settings
	options := sdk.EntityUpdateOptions{
		Location:   deserializeLocation(d.Get("location")),
		EntityType: d.Get("entity_type").(string),
		CustomName: d.Get("custom_name").(string),
	}

	id := d.Id()
	conn.Entities.Update(ctx, id, options)
	return resourceMinecraftEntityRead(d, meta)
}

func resourceMinecraftEntityDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*sdk.Client)
	id := d.Id()

	// Create a context
	ctx := context.Background()

	conn.Entities.Delete(ctx, id)

	// Wait until resource is finished deleting
	resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.Entities.Read(ctx, id)
		// A 404 error indicates success
		if err != nil {
			if err == sdk.ErrResourceNotFound {
				return nil
			}
		}
		log.Printf("[DEBUG] Entity deleting...")
		return resource.RetryableError(errors.New("invalid state"))

	})
	return nil
}
