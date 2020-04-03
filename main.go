package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/scottwinkler/terraform-provider-minecraft/minecraft"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: minecraft.Provider,
	})
}
