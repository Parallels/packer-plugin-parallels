// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"
	"slices"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// VMConfig contains various configuration options for the VM.
type VMConfig struct {
	// StartupView specifies the view to be shown when the VM starts.
	// Possible values are: same, window, coherence, fullscreen, modality, headless.
	// MacOS VMs in Apple Silicon Chip Macs do not support coherence, fullscreen and modality.
	StartupView string `mapstructure:"startup_view" required:"false"`
	// OnWindowClose specifies the action to be taken on VM when the VM window is closed.
	// Possible values are: suspend, shutdown, stop, ask, keep-running.
	OnWindowClose string `mapstructure:"on_window_close" required:"false"`
}

// Prepare sets the default values for SSH communicator properties.
func (c *VMConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	var validStartupView = []string{"", "same", "window", "coherence", "fullscreen", "modality", "headless"}

	if !slices.Contains(validStartupView, c.StartupView) {
		errs = append(errs,
			fmt.Errorf("invalid value for startup-view: %s. Allowed values are : %v", c.StartupView, validStartupView))
	}

	var validOnWindowClose = []string{"", "suspend", "shutdown", "stop", "ask", "keep-running"}
	if !slices.Contains(validOnWindowClose, c.OnWindowClose) {
		errs = append(errs,
			fmt.Errorf("invalid value for on-window-close: %s. Allowed values are : %v", c.OnWindowClose, validOnWindowClose))
	}

	return errs
}
