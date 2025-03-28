// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iso

import (
	"context"
	"fmt"

	parallelscommon "github.com/Parallels/packer-plugin-parallels/builder/parallels/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// This step sets the device boot order for the virtual machine.
//
// Uses:
//
//	driver Driver
//	ui packersdk.Ui
//	vmName string
//
// Produces:
type stepSetBootOrder struct{}

func (s *stepSetBootOrder) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(parallelscommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	// Set new boot order
	ui.Say("Setting the boot order...")
	command := []string{
		"set", vmName,
		"--device-bootorder", "hdd0 cdrom0",
	}

	if err := driver.Prlctl(command...); err != nil {
		err := fmt.Errorf("Error setting the boot order: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepSetBootOrder) Cleanup(state multistep.StateBag) {}
