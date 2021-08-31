package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"

	"github.com/hashicorp/packer-plugin-parallels/builder/parallels/iso"
	"github.com/hashicorp/packer-plugin-parallels/builder/parallels/pvm"
	"github.com/hashicorp/packer-plugin-parallels/version"
)

func main() {
	pps := plugin.NewSet()
	pps.SetVersion(version.PluginVersion)
	pps.RegisterBuilder("iso", new(iso.Builder))
	pps.RegisterBuilder("pvm", new(pvm.Builder))
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
