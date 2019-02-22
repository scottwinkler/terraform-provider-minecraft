package minecraft

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

//CylinderRequest struct
type CylinderRequest struct {
	MaterialID string `json:"materialId"`
	Location   []int  `json:"location"`
	Dimensions []int  `json:"cylinderDimensions"`
}

//Cylinder struct
type Cylinder struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Location []int  `json:"location"`
	//Dimensions   []int    `json:"cylinderDimensions"`
	PreviousData []string `json:"previousData"`
	CurrentData  []string `json:"currentData"`
	MaterialID   string   `json:"materialId"`
}

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
					},
				},
			},
			"dimensions": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"radius": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"height": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"material_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
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
			"current_data": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"dirty": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func deserializeCylinderDimensions(dimensions []int) []interface{} {
	l := make([]interface{}, 1)
	m := make(map[string]interface{})
	m["radius"] = dimensions[0]
	m["height"] = dimensions[1]
	l[0] = m
	return l
}

func serializeCylinderDimensions(dimensions interface{}) []int {
	var radius int
	var height int
	d := dimensions.([]interface{})[0].(map[string]interface{})
	r, _ := d["radius"]
	radius = r.(int)
	h, _ := d["height"]
	height = h.(int)
	return []int{radius, height}
}

func makeCylinderRequest(d *schema.ResourceData) *CylinderRequest {
	// Get the material id
	materialID := d.Get("material_id").(string)
	dimensions := serializeCylinderDimensions(d.Get("dimensions"))
	location := serializeLocation(d.Get("location"))

	cylinderRequest := &CylinderRequest{
		MaterialID: materialID,
		Dimensions: dimensions,
		Location:   location,
	}
	return cylinderRequest
}
func resourceMinecraftCylinderCreate(d *schema.ResourceData, meta interface{}) error {
	minecraftClient := meta.(*Client)

	req, err := minecraftClient.newRequest("POST", "cylinder/", makeCylinderRequest(d))
	if err != nil {
		return err
	}

	cylinder := &Cylinder{}
	err = minecraftClient.do(ctx, req, cylinder)
	if err != nil {
		return err
	}
	d.SetId(cylinder.ID)
	resourceMinecraftCylinderRead(d, meta)
	return nil
}

func resourceMinecraftCylinderRead(d *schema.ResourceData, meta interface{}) error {
	minecraftClient := meta.(*Client)
	id := d.Id()
	req, err := minecraftClient.newRequest("GET", "cylinder/"+id, nil)
	if err != nil {
		return err
	}

	cylinder := &Cylinder{}
	err = minecraftClient.do(ctx, req, cylinder)
	if err != nil {
		return err
	}
	d.Set("material_id", cylinder.MaterialID)
	log.Printf("loc: %v", cylinder.Location)
	d.Set("location", deserializeLocation(cylinder.Location))
	//log.Printf("dim: %v", cylinder.Dimensions)
	//d.Set("dimensions", deserializeCylinderDimensions(cylinder.Dimensions))
	d.Set("current_data", cylinder.CurrentData)
	d.Set("previous_data", cylinder.PreviousData)
	for _, element := range cylinder.CurrentData {
		if element != cylinder.MaterialID {
			d.Set("dirty", true)
			break
		}
	}
	d.Set("current_data", cylinder.CurrentData)
	d.Set("type", cylinder.Type)
	return nil
}

func resourceMinecraftCylinderUpdate(d *schema.ResourceData, meta interface{}) error {
	minecraftClient := meta.(*Client)
	id := d.Id()
	req, err := minecraftClient.newRequest("PATCH", "cylinder/"+id, makeCylinderRequest(d))
	if err != nil {
		return err
	}

	cylinder := &Cylinder{}
	err = minecraftClient.do(ctx, req, cylinder)
	if err != nil {
		return err
	}
	return resourceMinecraftCylinderRead(d, meta)
}

func resourceMinecraftCylinderDelete(d *schema.ResourceData, meta interface{}) error {
	minecraftClient := meta.(*Client)
	id := d.Id()
	req, err := minecraftClient.newRequest("DELETE", "cylinder/"+id, nil)
	if err != nil {
		return err
	}
	err = minecraftClient.do(ctx, req, nil)
	if err != nil {
		return err
	}
	return nil
}
