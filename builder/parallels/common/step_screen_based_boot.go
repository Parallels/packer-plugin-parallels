// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

const macScreenCaptureBinaryPath = "/usr/sbin/screencapture"

// This step creates the virtual disk that will be used as the
// hard drive for the virtual machine.
type StepScreenBasedBoot struct {
	ScreenConfigs map[string]BootScreenConfig
	OCRLibrary    string
	VmName        string
	Ctx           interpolate.Context
}

func (s *StepScreenBasedBoot) executeBootCommand(bootConfig bootcommand.BootConfig, ctx context.Context, state multistep.StateBag) error {
	// Execute the boot command
	step := StepTypeBootCommand{
		BootWait:       -1,
		BootCommand:    bootConfig.FlatBootCommand(),
		HostInterfaces: []string{},
		VMName:         s.VmName,
		Ctx:            s.Ctx,
		GroupInterval:  bootConfig.BootGroupInterval,
	}

	resultAction := step.Run(ctx, state)
	if resultAction == multistep.ActionContinue {
		return nil
	}

	// Report any errors.
	if rawErr, ok := state.GetOk("error"); ok {
		return rawErr.(error)
	}

	return errors.New("boot command failed")
}

func (s *StepScreenBasedBoot) captureScreenPD(state multistep.StateBag, fileName string) error {
	driver := state.Get("driver").(Driver)
	return driver.Prlctl("capture", s.VmName, "--file", fileName)
}

func (s *StepScreenBasedBoot) captureScreenMac(windowId string, fileName string) error {
	// Window ID option
	windowIDOption := "-l" + fmt.Sprint(windowId)
	// Command-line arguments to pass to the binary
	args := []string{"-x", windowIDOption, fileName}
	_, err := executeBinary(macScreenCaptureBinaryPath, args...)
	return err
}

func (s *StepScreenBasedBoot) captureScreen(state multistep.StateBag, windowId string, fileName string) error {
	if windowId == "-1" { // PD version is above 20.0.0
		return s.captureScreenPD(state, fileName)
	}
	return s.captureScreenMac(windowId, fileName)
}

func (s *StepScreenBasedBoot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// We don't have any screen configs, so no wasting of time here.
	if (s.ScreenConfigs == nil) || (len(s.ScreenConfigs) == 0) {
		return multistep.ActionContinue
	}

	driver := state.Get("driver").(Driver)
	prlctlCurrVersionStr, verErr := driver.Version()
	if verErr != nil {
		log.Println("Error retrieving prlctl version:", verErr)
		return multistep.ActionHalt
	}
	prlctlCurrVersion, _ := version.NewVersion(prlctlCurrVersionStr)
	v2, _ := version.NewVersion("20.0.0")

	// From PD20.0.0, 'prlctl capture' command is available for macOS VMs
	windowId := -1
	if prlctlCurrVersion.LessThan(v2) {
		// Retrieve the window ID
		windowIDDetector := WindowIDDetector{}
		var err error
		windowId, err = windowIDDetector.DetectWindowId(s.VmName, state)
		if err != nil || windowId == 0 {
			log.Println("Error retrieving window ID:", err)
			return multistep.ActionHalt
		}
	}

	ui := state.Get("ui").(packersdk.Ui)
	// Create a temporary file to store the screenshot
	file, err := os.CreateTemp("", "screenshot*.png")
	if err != nil {
		log.Println("Error creating temporary file:", err)
		return multistep.ActionHalt
	}
	file.Close()
	defer os.Remove(file.Name())

	ui.Say("Starting screen based boot...")

	prevTime := time.Now()
	minDelay := 1 * time.Second
	lastScreen := BootScreenConfig{}
	ocrWrapper, _ := NewOCRWrapper(s.OCRLibrary, s.ScreenConfigs)
	for {
		log.Println("Checking screen...")

		// Checking for interrupt
		if state.Get(multistep.StateCancelled) != nil {
			return multistep.ActionHalt
		}

		// Check if the minimum delay has passed
		if time.Since(prevTime) < minDelay {
			time.Sleep(minDelay - time.Since(prevTime))
		}
		prevTime = time.Now()

		// Capturing the screenshot
		err := s.captureScreen(state, fmt.Sprint(windowId), file.Name())
		if err != nil {
			log.Println("Error: '", err, "' while capturing the screenshot.")
			ui.Error("Error capturing the screenshot. Make sure you have the necessary permissions.")
			return multistep.ActionHalt
		}

		screenConfig, err := ocrWrapper.IdentifyCurrentScreen(file.Name())
		if err != nil {
			log.Println("Error:", err)
			return multistep.ActionHalt
		}

		// If the screen is the same as the last screen and the last screen is not empty, skip the boot command
		if screenConfig.ScreenName == lastScreen.ScreenName && len(lastScreen.MatchingStrings) > 0 {
			continue
		}

		// If we switched the screen & the last screen is execute only once, then remove the screen
		if len(lastScreen.MatchingStrings) > 0 && lastScreen.ExecuteOnlyOnce {
			ocrWrapper.RemoveBootScreenConfigIfExist(lastScreen.ScreenName)
		}

		lastScreen = screenConfig
		ui.Say("Screen changed to " + screenConfig.ScreenName)
		// Execute the boot command
		bootCommand := screenConfig.FlatBootCommand()
		if bootCommand != "" {
			err := s.executeBootCommand(screenConfig.BootConfig, ctx, state)
			if err != nil {
				log.Println("Error:", err)
				return multistep.ActionHalt
			}
		}

		if screenConfig.IsLastScreen {
			break
		}

	}

	return multistep.ActionContinue
}

func (s *StepScreenBasedBoot) Cleanup(state multistep.StateBag) {}
