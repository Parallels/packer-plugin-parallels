// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// This step uploads the Parallels Tools ISO to the virtual machine.
//
// Uses:
//
//	communicator packersdk.Communicator
//	parallels_tools_path string
//	ui packersdk.Ui
//
// Produces:
type toolsPathTemplate struct {
	Flavor string
}

// StepUploadParallelsTools is a step that uploads the Parallels Tools ISO
// to the VM.
//
// Uses:
//
//	communicator packersdk.Communicator
//	parallels_tools_path string
//	ui packersdk.Ui
type StepUploadParallelsTools struct {
	ParallelsToolsFlavor    string
	ParallelsToolsGuestPath string
	ParallelsToolsMode      string
	Ctx                     interpolate.Context
}

// Run uploads the Parallels Tools ISO to the VM.
func (s *StepUploadParallelsTools) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packersdk.Communicator)
	ui := state.Get("ui").(packersdk.Ui)

	// If we're attaching then don't do this, since we attached.
	if s.ParallelsToolsMode != ParallelsToolsModeUpload {
		log.Println("Not uploading Parallels Tools since mode is not upload")
		return multistep.ActionContinue
	}

	// Get the Parallels Tools path on the host machine
	parallelsToolsPath := state.Get("parallels_tools_path").(string)

	f, err := os.Open(parallelsToolsPath)
	if err != nil {
		state.Put("error", fmt.Errorf("Error opening Parallels Tools ISO: %s", err))
		return multistep.ActionHalt
	}
	defer f.Close()

	s.Ctx.Data = &toolsPathTemplate{
		Flavor: s.ParallelsToolsFlavor,
	}

	s.ParallelsToolsGuestPath, err = interpolate.Render(s.ParallelsToolsGuestPath, &s.Ctx)
	if err != nil {
		err = fmt.Errorf("Error preparing Parallels Tools path: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Uploading Parallels Tools for '%s' to path: '%s'",
		s.ParallelsToolsFlavor, s.ParallelsToolsGuestPath))
	if err := comm.Upload(s.ParallelsToolsGuestPath, f, nil); err != nil {
		err = fmt.Errorf("Error uploading Parallels Tools: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup does nothing.
func (s *StepUploadParallelsTools) Cleanup(state multistep.StateBag) {}
