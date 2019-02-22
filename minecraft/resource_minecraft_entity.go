package minecraft

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

//EntityResourceData struct
type EntityResourceData struct {
	EntityType string `json:"entityType"`
	Location   []int  `json:"location"`
	CustomName string `json:"customName"`
}

//Entity struct
type Entity struct {
	ID             string              `json:"id"`
	ResourceData   *EntityResourceData `json:"serializedEntityResourceData"`
	ResourceStatus ResourceStatus      `json:"resourceStatus"`
}

//ResourceStatus return
type ResourceStatus string

const (
	ResourceInitializing ResourceStatus = "Initializing"
	ResourceCreating     ResourceStatus = "Creating"
	//ResourceDeleting ResourceStatus ="Deleting"
	//ResourceUpdating ResourceStatus ="Updating"
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

func makeEntityResourceData(d *schema.ResourceData) *EntityResourceData {
	entityType := d.Get("entity_type").(string)
	customName := d.Get("custom_name").(string)
	location := serializeLocation(d.Get("location"))

	EntityResourceData := &EntityResourceData{
		EntityType: entityType,
		CustomName: customName,
		Location:   location,
	}
	return EntityResourceData
}
func resourceMinecraftEntityCreate(d *schema.ResourceData, meta interface{}) error {
	minecraftClient := meta.(*Client)
	r := makeEntityResourceData(d)
	log.Printf("resourceData: %v", r)
	req, err := minecraftClient.newRequest("POST", "entity/", makeEntityResourceData(d))
	if err != nil {
		return err
	}

	entity := &Entity{}
	err = minecraftClient.do(ctx, req, entity)
	if err != nil {
		return err
	}
	d.SetId(entity.ID)
	resourceMinecraftEntityRead(d, meta)
	return nil
}

func resourceMinecraftEntityRead(d *schema.ResourceData, meta interface{}) error {
	minecraftClient := meta.(*Client)
	id := d.Id()
	req, err := minecraftClient.newRequest("GET", "entity/"+id, nil)
	if err != nil {
		return err
	}

	entity := &Entity{}
	err = minecraftClient.do(ctx, req, entity)
	if err != nil {
		return err
	}
	d.Set("custom_name", entity.ResourceData.CustomName)
	d.Set("location", deserializeLocation(entity.ResourceData.Location))
	d.Set("entityType", entity.ResourceData.EntityType)
	return nil
}

func resourceMinecraftEntityUpdate(d *schema.ResourceData, meta interface{}) error {
	minecraftClient := meta.(*Client)
	id := d.Id()
	req, err := minecraftClient.newRequest("PATCH", "entity/"+id, makeEntityResourceData(d))
	if err != nil {
		return err
	}

	entity := &Entity{}
	err = minecraftClient.do(ctx, req, entity)
	if err != nil {
		return err
	}
	return resourceMinecraftEntityRead(d, meta)
}

func resourceMinecraftEntityDelete(d *schema.ResourceData, meta interface{}) error {
	minecraftClient := meta.(*Client)
	id := d.Id()
	req, err := minecraftClient.newRequest("DELETE", "entity/"+id, nil)
	if err != nil {
		return err
	}
	err = minecraftClient.do(ctx, req, nil)
	if err != nil {
		return err
	}
	return nil
}
