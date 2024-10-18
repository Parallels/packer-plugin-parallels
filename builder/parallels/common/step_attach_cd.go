// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepAttachCD is a step that attaches a cd to the virtual machine.
//
// Uses:
//
//	driver Driver
//	ui packersdk.Ui
//	vmName string
//
// Produces:
type StepAttachCD struct {
	cdPath      string
	cdromDevice string
}

// Run adds a virtual CD to the VM and attaches the image.
// If the image is not specified, then this step will be skipped.
func (s *StepAttachCD) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// Determine if we even have a cd to attach
	var cdPath string
	if cdPathRaw, ok := state.GetOk("cd_path"); ok {
		cdPath = cdPathRaw.(string)
	} else {
		log.Println("No cd found, not attaching.")
		return multistep.ActionContinue
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Attaching cd...")
	// Attaching the cd using the driver
	cdrom, err := driver.DeviceAddCDROM(vmName, cdPath)
	if err != nil {
		err = fmt.Errorf("error attaching cd: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	// Track the device name so that we can can delete later
	s.cdromDevice = cdrom

	return multistep.ActionContinue
}

// Cleanup removes the virtual FDD device attached to the VM.
func (s *StepAttachCD) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(Driver)
	vmName := state.Get("vmName").(string)

	if s.cdPath == "" {
		return
	}

	log.Println("Detaching cd disk...")
	command := []string{
		"set", vmName,
		"--device-del", s.cdromDevice,
	}
	_ = driver.Prlctl(command...)
}
