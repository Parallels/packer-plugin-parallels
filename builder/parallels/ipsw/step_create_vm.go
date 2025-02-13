// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ipsw

import (
	"context"
	"fmt"
	"strconv"

	parallelscommon "github.com/Parallels/packer-plugin-parallels/builder/parallels/common"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// This step creates the actual virtual machine.
//
// Produces:
//
//	vmName string - The name of the VM
type stepCreateVM struct {
	vmName string
}

func (s *stepCreateVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(parallelscommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)
	ipswPath := state.Get("ipsw_path").(string)
	name := config.VMName

	commands := make([][]string, 3)

	commands[0] = []string{
		"create", name,
		"-o", "macos",
		"--restore-image", ipswPath,
		"--dst", config.OutputDir,
		"--no-hdd",
	}
	commands[1] = []string{
		"set", name,
		"--cpus", strconv.Itoa(config.HWConfig.CpuCount),
	}
	commands[2] = []string{
		"set", name,
		"--memsize", strconv.Itoa(config.HWConfig.MemorySize),
	}

	ui.Say("Creating virtual machine...")
	for _, command := range commands {
		if err := driver.Prlctl(command...); err != nil {
			err := fmt.Errorf("Error creating VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Say("Applying default settings...")
	if err := driver.SetDefaultConfiguration(name); err != nil {
		err := fmt.Errorf("Error VM configuration: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	prlctlCurrVersionStr, verErr := driver.Version()
	if verErr != nil {
		return multistep.ActionHalt
	}
	prlctlCurrVersion, _ := version.NewVersion(prlctlCurrVersionStr)
	v2, _ := version.NewVersion("20.0.0")

	if prlctlCurrVersion.LessThan(v2) {
		// workaround when prlctl capture is not available
		// Currently it is not possible to retrieve window ID of the MacOS VM when it is in headless mode
		// So, we are setting the VM to window mode after setting the default configuration
		command := []string{"set", name, "--startup-view", "window"}
		if err := driver.Prlctl(command...); err != nil {
			err := fmt.Errorf("error setting VM to window mode: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Set the VM name property on the first command
	if s.vmName == "" {
		s.vmName = name
	}

	// Set the final name in the state bag so others can use it
	state.Put("vmName", s.vmName)
	return multistep.ActionContinue
}

func (s *stepCreateVM) Cleanup(state multistep.StateBag) {
	if s.vmName == "" {
		return
	}

	driver := state.Get("driver").(parallelscommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Unregistering virtual machine...")
	if err := driver.Prlctl("unregister", s.vmName); err != nil {
		ui.Error(fmt.Sprintf("Error unregistering virtual machine: %s", err))
	}
}
