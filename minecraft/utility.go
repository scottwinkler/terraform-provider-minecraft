package minecraft

import (
	sdk "github.com/scottwinkler/go-minecraft"
)

func deserializeLocation(location interface{}) *sdk.Location {
	l := location.([]interface{})[0].(map[string]interface{})
	x, _ := l["x"]
	xPos := x.(int)
	y, _ := l["y"]
	yPos := y.(int)
	z, _ := l["z"]
	zPos := z.(int)
	w, _ := l["world"]
	world := w.(string)
	return sdk.NewLocation(xPos, yPos, zPos, world)
}

func serializeLocation(location *sdk.Location) []interface{} {
	l := make([]interface{}, 1)
	m := make(map[string]interface{})
	m["x"] = location.X
	m["y"] = location.Y
	m["z"] = location.Z
	m["world"] = location.World
	l[0] = m
	return l
}
