provider "minecraft" {
    hostname = "http://localhost:8080"
}
/*
resource "minecraft_entity" "sheep" {
    count = 3
    location = {
        x = 10
        y = 5
        z = 10
    }
    entity_type = "SHEEP"
}
*/
resource "minecraft_entity" "pig" {
    location = {
        x = 9
        y = 5
        z = 9
    }
    entity_type = "PIG"
   custom_name = "piggy"
}