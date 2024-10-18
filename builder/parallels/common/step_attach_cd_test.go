// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func TestStepAttachCD_impl(t *testing.T) {
	var _ multistep.Step = new(StepAttachCD)
}

func TestStepAttachCD(t *testing.T) {
	state := testState(t)
	step := new(StepAttachCD)

	// Create a temporary file for our CD
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Close()
	defer os.Remove(tf.Name())

	state.Put("cd_path", tf.Name())
	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if !driver.DeviceAddCDROMCalled {
		t.Fatal("DeviceAddCDROM not called")
	}
	if driver.DeviceAddCDROMName != "foo" {
		t.Fatal("DeviceAddCDROM name isn't what we specified (foo)")
	}
	if driver.DeviceAddCDROMResult != "cdrom0" {
		t.Fatal("DeviceAddCDROM didn't return cdrom0")
	}
	if driver.DeviceAddCDROMErr != nil {
		t.Fatal("DeviceAddCDROM returned an error")
	}

}

func TestStepAttachCD_noCD(t *testing.T) {
	state := testState(t)
	step := new(StepAttachCD)

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if driver.DeviceAddCDROMCalled {
		t.Fatal("DeviceAddCDROM has been called")
	}
}
