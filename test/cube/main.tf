provider "minecraft" {
    hostname = "http://localhost:8080"
}

resource "minecraft_cube" "diamonds" {
    location = {
        x = 100
        y = 10
        z = 100
    }
    material_id = "DIAMOND_BLOCK"
    dimensions = {
        length = 5
        width = 5
        height = 5
    }
}