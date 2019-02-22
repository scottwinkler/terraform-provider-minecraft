provider "minecraft" {
    hostname = "http://localhost:8080"
}

resource "minecraft_cube" "wall1" {
    location = {
        x = 0
        y = 4
        z = 0
    }
    material_id = "COBBLESTONE"
    dimensions = {
        length = 20
        width = 1
        height = 5
    }
}
resource "minecraft_cube" "wall2" {
    depends_on = ["minecraft_cube.wall1"]
    location = {
        x = 20
        y = 4
        z = 0
    }
    material_id = "COBBLESTONE"
    dimensions = {
        length = 1
        width = 20
        height = 5
    }
}

resource "minecraft_cube" "wall3" {
    depends_on = ["minecraft_cube.wall2"]
    location = {
        x = 0
        y = 4
        z = 19
    }
    material_id = "COBBLESTONE"
    dimensions = {
        length = 20
        width = 1
        height = 5
    }
}

resource "minecraft_cube" "wall4" {
    depends_on = ["minecraft_cube.wall3"]
    location = {
        x = -1
        y = 4
        z = 0
    }
    material_id = "COBBLESTONE"
    dimensions = {
        length = 1
        width = 20
        height = 5
    }
}

resource "minecraft_cube" "door" {
    depends_on = ["minecraft_cube.wall1"]
    location = {
        x = 9
        y = 4
        z = 0
    }
    material_id = "AIR"
    dimensions = {
        length = 3
        width = 1
        height = 3
    }
}

resource "minecraft_cube" "floor" {
    location = {
        x = 0
        y = 3
        z = 0
    }
    material_id = "OAK_PLANKS"
    dimensions = {
        length = 20
        width = 20
        height = 1
    }
}

resource "minecraft_cylinder" "tower1" {
    location = {
        x = 0
        y = 4
        z = 0
    }
    material_id = "COBBLESTONE"
    dimensions = {
        radius = 5
        height = 8
    }
}

resource "minecraft_cylinder" "tower2" {
    location = {
        x = 20
        y = 4
        z = 0
    }
    material_id = "COBBLESTONE"
    dimensions = {
        radius = 5
        height = 8
    }
}

resource "minecraft_cylinder" "tower3" {
    location = {
        x = 0
        y = 4
        z = 19
    }
    material_id = "COBBLESTONE"
    dimensions = {
        radius = 5
        height = 8
    }
}


resource "minecraft_cylinder" "tower4" {
    location = {
        x = 19
        y = 4
        z = 19
    }
    material_id = "COBBLESTONE"
    dimensions = {
        radius = 5
        height = 8
    }
}

resource "minecraft_Entity" "sheep" {
    depends_on = ["minecraft_cube.floor"]
    count = 3
    location = {
        x = 10
        y = 5
        z = 10
    }
    entity = "SHEEP"
}

resource "minecraft_Entity" "pig" {
    depends_on = ["minecraft_cube.floor"]
    location = {
        x = 9
        y = 5
        z = 9
    }
    entity = "PIG"
   custom_name = "piggy"
}