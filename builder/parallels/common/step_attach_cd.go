// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"fmt"
	"log"
	"regexp"

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
	cdRomType   string `mapstructure:"cdrom_type"`
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
	//Use CLI instead of driver wrapper so we can control iface
	//Also the cli lets us get back the cdrom device for cleanup
	/*// Attaching the cd using the driver instead of cli
	cdrom, err := driver.DeviceAddCDROM(vmName, cdPath)
	if err != nil {
		err = fmt.Errorf("Error attaching cd: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	// Track the device name so that we can can delete later
	s.cdromDevice = cdrom*/

	//Attaching the cd using the cli
	// just force sata for now and don't let people choose
	s.cdRomType = "sata"
	command := []string{
		"set", vmName,
		"--device-add", "cdrom",
		"--image", cdPath,
		"--iface", s.cdRomType,
		"--enable", "--connect",
	}

	out, err := driver.PrlctlGet(command...)
	if err != nil {
		state.Put("error", fmt.Errorf("Error adding cd: %s", err))
		return multistep.ActionHalt
	}

	deviceRe := regexp.MustCompile(`\s+(cdrom\d+)\s+`)
	matches := deviceRe.FindStringSubmatch(out)
	if matches == nil {
		state.Put("error", fmt.Errorf("Could not determine cdrom device name in the output: %s", out))
		return multistep.ActionHalt
	}

	deviceName := matches[1]
	s.cdromDevice = deviceName

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
