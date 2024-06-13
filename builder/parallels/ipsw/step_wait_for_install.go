// Copyright (c) Parallels International GmBH
// SPDX-License-Identifier: MPL-2.0

package ipsw

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	parallelscommon "github.com/Parallels/packer-plugin-parallels/builder/parallels/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// This step waits for installation to complete
//
//	vmName string - The name of the VM
type stepWaitForInstall struct {
	vmName string
}

func (s *stepWaitForInstall) FileContainsString(filePath string, str []string) (bool, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for _, s := range str {
			if strings.Contains(line, s) {
				return true, nil
			}
		}
	}

	// Check for errors
	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

func (s *stepWaitForInstall) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	s.vmName = config.VMName

	// Calculate the log file full path
	vmLogPath := fmt.Sprintf("%s/%s.macvm/parallels.log", config.OutputDir, s.vmName)

	ui.Say("Waiting for the installation to complete...")
	// check installation failed/success every 10 seconds
	for {
		contains, err := s.FileContainsString(vmLogPath, []string{"Installation succeeded"})
		if err != nil {
			err := fmt.Errorf("error reading log file: %s", err)
			state.Put("error", err)
			return multistep.ActionHalt
		}

		if contains {
			ui.Say("Installation completed successfully")
			break
		}

		contains, err = s.FileContainsString(vmLogPath, []string{"Cancelled(1)", "Installation failed"})
		if err != nil {
			return multistep.ActionHalt
		}

		if contains {
			err := fmt.Errorf("installation failed")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Wait for 10 seconds
		select {
		case <-ctx.Done():
			return multistep.ActionHalt
		case <-time.After(10 * time.Second):
		}

	}

	return multistep.ActionContinue
}

func (s *stepWaitForInstall) Cleanup(state multistep.StateBag) {
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
