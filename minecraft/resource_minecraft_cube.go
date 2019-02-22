package minecraft

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

//CubeRequest struct
type CubeRequest struct {
	MaterialID string `json:"materialId"`
	Location   []int  `json:"location"`
	Dimensions []int  `json:"cubeDimensions"`
}

//Cube struct
type Cube struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Location     []int    `json:"location"`
	Dimensions   []int    `json:"cubeDimensions"`
	PreviousData []string `json:"previousData"`
	CurrentData  []string `json:"currentData"`
	MaterialID   string   `json:"materialId"`
}

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
					},
				},
			},
			"dimensions": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"length": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"width": {
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

func deserializeCubeDimensions(dimensions []int) []interface{} {
	l := make([]interface{}, 1)
	m := make(map[string]interface{})
	m["height"] = dimensions[0]
	m["width"] = dimensions[1]
	m["length"] = dimensions[2]
	l[0] = m
	return l
}

func serializeCubeDimensions(dimensions interface{}) []int {
	var height int
	var width int
	var length int
	d := dimensions.([]interface{})[0].(map[string]interface{})
	h, _ := d["height"]
	height = h.(int)
	w, _ := d["width"]
	width = w.(int)
	l, _ := d["length"]
	length = l.(int)
	return []int{height, width, length}
}

func deserializeLocation(location []int) []interface{} {
	l := make([]interface{}, 1)
	m := make(map[string]interface{})
	m["x"] = location[0]
	m["y"] = location[1]
	m["z"] = location[2]
	l[0] = m
	return l
}

func serializeLocation(location interface{}) []int {
	var xPos int
	var yPos int
	var zPos int
	l := location.([]interface{})[0].(map[string]interface{})
	x, _ := l["x"]
	xPos = x.(int)
	y, _ := l["y"]
	yPos = y.(int)
	z, _ := l["z"]
	zPos = z.(int)
	return []int{xPos, yPos, zPos}
}

func makeCubeRequest(d *schema.ResourceData) *CubeRequest {
	// Get the material id
	materialID := d.Get("material_id").(string)
	dimensions := serializeCubeDimensions(d.Get("dimensions"))
	location := serializeLocation(d.Get("location"))

	cubeRequest := &CubeRequest{
		MaterialID: materialID,
		Dimensions: dimensions,
		Location:   location,
	}
	return cubeRequest
}
func resourceMinecraftCubeCreate(d *schema.ResourceData, meta interface{}) error {
	minecraftClient := meta.(*Client)

	req, err := minecraftClient.newRequest("POST", "cube/", makeCubeRequest(d))
	if err != nil {
		return err
	}

	cube := &Cube{}
	err = minecraftClient.do(ctx, req, cube)
	if err != nil {
		return err
	}
	d.SetId(cube.ID)
	resourceMinecraftCubeRead(d, meta)
	return nil
}

func resourceMinecraftCubeRead(d *schema.ResourceData, meta interface{}) error {
	minecraftClient := meta.(*Client)
	id := d.Id()
	req, err := minecraftClient.newRequest("GET", "cube/"+id, nil)
	if err != nil {
		return err
	}

	cube := &Cube{}
	err = minecraftClient.do(ctx, req, cube)
	if err != nil {
		return err
	}
	d.Set("material_id", cube.MaterialID)
	log.Printf("loc: %v", cube.Location)
	d.Set("location", deserializeLocation(cube.Location))
	log.Printf("dim: %v", cube.Dimensions)
	d.Set("dimensions", deserializeCubeDimensions(cube.Dimensions))
	d.Set("current_data", cube.CurrentData)
	d.Set("previous_data", cube.PreviousData)
	for _, element := range cube.CurrentData {
		if element != cube.MaterialID {
			d.Set("dirty", true)
			break
		}
	}
	d.Set("current_data", cube.CurrentData)
	d.Set("type", cube.Type)
	return nil
}

func resourceMinecraftCubeUpdate(d *schema.ResourceData, meta interface{}) error {
	minecraftClient := meta.(*Client)
	id := d.Id()
	req, err := minecraftClient.newRequest("PATCH", "cube/"+id, makeCubeRequest(d))
	if err != nil {
		return err
	}

	cube := &Cube{}
	err = minecraftClient.do(ctx, req, cube)
	if err != nil {
		return err
	}
	return resourceMinecraftCubeRead(d, meta)
}

func resourceMinecraftCubeDelete(d *schema.ResourceData, meta interface{}) error {
	minecraftClient := meta.(*Client)
	id := d.Id()
	req, err := minecraftClient.newRequest("DELETE", "cube/"+id, nil)
	if err != nil {
		return err
	}
	err = minecraftClient.do(ctx, req, nil)
	if err != nil {
		return err
	}
	return nil
}
