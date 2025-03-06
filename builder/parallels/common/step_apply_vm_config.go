// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepApplyVMConfig is a step that applies some of the VM configuration provided by user
type StepApplyVMConfig struct {
	CustomVMConfig VMConfig
}

func (s *StepApplyVMConfig) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	vmName := state.Get("vmName").(string)
	ui := state.Get("ui").(packersdk.Ui)

	var commands [][]string

	if s.CustomVMConfig.StartupView != "" {
		commands = append(commands, []string{"set", vmName, "--startup-view", s.CustomVMConfig.StartupView})
	}

	if s.CustomVMConfig.OnWindowClose != "" {
		commands = append(commands, []string{"set", vmName, "--on-window-close", s.CustomVMConfig.OnWindowClose})
	}

	for _, command := range commands {
		if err := driver.Prlctl(command...); err != nil {
			err := fmt.Errorf("error creating VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepApplyVMConfig) Cleanup(state multistep.StateBag) {
}
