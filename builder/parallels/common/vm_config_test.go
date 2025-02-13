// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

func TestVMConfigPrepare_Prlctl(t *testing.T) {
	// Test with empty
	c := new(VMConfig)
	errs := c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if c.StartupView != "" {
		t.Fatalf("bad: %#v", c.StartupView)
	}

	if c.OnWindowClose != "" {
		t.Fatalf("bad: %#v", c.OnWindowClose)
	}

	c = new(VMConfig)
	c.OnWindowClose = "something"

	errs = c.Prepare(interpolate.NewContext())
	if len(errs) == 0 {
		t.Fatalf("err: %#v", errs)
	}

	c = new(VMConfig)
	c.StartupView = "something"

	errs = c.Prepare(interpolate.NewContext())
	if len(errs) == 0 {
		t.Fatalf("err: %#v", errs)
	}

	supportedStartupViewValues := []string{"same", "window", "coherence", "fullscreen", "modality", "headless"}
	for _, value := range supportedStartupViewValues {
		c = new(VMConfig)
		c.StartupView = value
		errs = c.Prepare(interpolate.NewContext())
		if len(errs) > 0 {
			t.Fatalf("err: %#v", errs)
		}
	}

	supportedOnWindowCloseValues := []string{"suspend", "shutdown", "stop", "ask", "keep-running"}
	for _, value := range supportedOnWindowCloseValues {
		c = new(VMConfig)
		c.OnWindowClose = value
		errs = c.Prepare(interpolate.NewContext())
		if len(errs) > 0 {
			t.Fatalf("err: %#v", errs)
		}
	}
}
