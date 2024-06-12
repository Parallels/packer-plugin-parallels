// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// This step creates the virtual disk that will be used as the
// hard drive for the virtual machine.
type StepScreenBasedBoot struct {
	ScreenConfigs []BootScreenConfig
	OCRLibrary    string
	VmName        string
	Ctx           interpolate.Context
}

func (s *StepScreenBasedBoot) executeBinary(binaryPath string, args ...string) (string, error) {
	// Run the binary with the specified path and arguments
	cmd := exec.Command(binaryPath, args...)

	// Capture the output
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Convert output to string and return
	result := string(output)
	return result, nil
}

func (s *StepScreenBasedBoot) detectScreen(text string) BootScreenConfig {
	text = strings.ToLower(text)
	emptyBootConfig := BootScreenConfig{}

	for _, screenConfig := range s.ScreenConfigs {
		// Empty screens are not considered for matching
		if (screenConfig.MatchingStrings == nil) || (len(screenConfig.MatchingStrings) == 0) {
			emptyBootConfig = screenConfig
			continue
		}

		// Check if all screenConfig.MatchingStrings are present in text. If all are present, return the screenConfig
		matching := true
		for _, matchingString := range screenConfig.MatchingStrings {
			if !strings.Contains(text, matchingString) {
				matching = false
				break
			}
		}

		if matching {
			return screenConfig
		}
	}

	return emptyBootConfig
}

func (s *StepScreenBasedBoot) executeBootCommand(bootConfig bootcommand.BootConfig, ctx context.Context, state multistep.StateBag) error {
	// Execute the boot command
	step := StepTypeBootCommand{
		BootWait:       bootConfig.BootWait,
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

func (s *StepScreenBasedBoot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// We don't have any screen configs, so no wasting of time here.
	if (s.ScreenConfigs == nil) || (len(s.ScreenConfigs) == 0) {
		return multistep.ActionContinue
	}

	// Retrieve the window ID
	windowIDDetector := WindowIDDetector{}
	windowId, err := windowIDDetector.DetectWindowId(s.VmName, state)
	if err != nil || windowId == 0 {
		log.Println("Error retrieving window ID:", err)
		return multistep.ActionHalt
	}

	ui := state.Get("ui").(packersdk.Ui)
	// Window ID option
	windowIDOption := "-l" + fmt.Sprint(windowId)
	// Path to the binary you want to execute
	binaryPath := "/usr/sbin/screencapture"
	// Create a temporary file to store the screenshot
	file, err := os.CreateTemp("", "screenshot*.png")
	if err != nil {
		log.Println("Error creating temporary file:", err)
		return multistep.ActionHalt
	}

	file.Close()
	defer os.Remove(file.Name())
	// Command-line arguments to pass to the binary
	args := []string{"-x", windowIDOption, file.Name()}

	ui.Say("Starting screen based boot...")

	prevTime := time.Now()
	minDelay := 1 * time.Second
	lastScreenName := ""
	ocrRunner := NewOCRRunner()
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
		_, err := s.executeBinary(binaryPath, args...)
		if err != nil {
			log.Println("Error:", err)
			return multistep.ActionHalt
		}

		// Use OCR to detect the text in the screenshot
		text, err := ocrRunner.RecognizeText(file.Name(), s.OCRLibrary)
		if err != nil {
			fmt.Println("Error:", err)
			return multistep.ActionHalt
		}

		log.Println("Recognized text:", text)
		// Detect the screen based on the text
		screenConfig := s.detectScreen(text)
		if screenConfig.ScreenName == "" {
			log.Printf("No matching screen found for text on screen : %s.\n", text)
			return multistep.ActionHalt
		}

		if screenConfig.ScreenName == lastScreenName {
			continue
		}

		lastScreenName = screenConfig.ScreenName
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
