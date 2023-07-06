// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"

	"github.com/Parallels/packer-plugin-parallels/builder/parallels/ipsw"
	"github.com/Parallels/packer-plugin-parallels/builder/parallels/iso"
	"github.com/Parallels/packer-plugin-parallels/builder/parallels/macvm"
	"github.com/Parallels/packer-plugin-parallels/builder/parallels/pvm"
	"github.com/Parallels/packer-plugin-parallels/version"
)

func main() {
	pps := plugin.NewSet()
	pps.SetVersion(version.PluginVersion)
	pps.RegisterBuilder("iso", new(iso.Builder))
	pps.RegisterBuilder("pvm", new(pvm.Builder))
	pps.RegisterBuilder("macvm", new(macvm.Builder))
	pps.RegisterBuilder("ipsw", new(ipsw.Builder))
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
