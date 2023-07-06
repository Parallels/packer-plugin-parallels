// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// This step imports an PVM VM into Parallels.
type StepImport struct {
	Name        string
	SourcePath  string
	vmName      string
	OutputDir   string
	ReassignMAC bool
}

func (s *StepImport) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say(fmt.Sprintf("Importing VM: %s", s.SourcePath))
	if err := driver.Import(s.Name, s.SourcePath, s.OutputDir, s.ReassignMAC); err != nil {
		err := fmt.Errorf("Error importing VM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.vmName = s.Name
	state.Put("vmName", s.Name)
	return multistep.ActionContinue
}

func (s *StepImport) Cleanup(state multistep.StateBag) {

	if s.vmName == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Unregistering virtual machine...")
	if err := driver.Prlctl("unregister", s.vmName); err != nil {
		ui.Error(fmt.Sprintf("Error unregistering virtual machine: %s", err))
	}
}
