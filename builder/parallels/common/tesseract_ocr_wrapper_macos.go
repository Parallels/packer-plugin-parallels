// Copyright (c) Parallels International GmBH
// SPDX-License-Identifier: MPL-2.0

//go:build darwin
// +build darwin

package common

import (
	"errors"
	"log"
	"os/exec"
	"strings"
)

type TesseractOCRWrapper struct {
	ScreenConfigs map[string]BootScreenConfig
}

func NewTesseractOCRWrapper(ScreenConfigs map[string]BootScreenConfig) *TesseractOCRWrapper {
	tesseractOCRWrapper := TesseractOCRWrapper{
		ScreenConfigs: ScreenConfigs,
	}

	return &tesseractOCRWrapper
}

func executeBinary(binaryPath string, args ...string) (string, error) {
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

func (s *TesseractOCRWrapper) detectScreen(text string) BootScreenConfig {
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

func (c *TesseractOCRWrapper) detectTextFromImageUsingTesseract(imagePath string) (string, error) {
	tesseractPath, err := exec.LookPath("tesseract")
	if err != nil {
		return "", errors.New("tesseract binary not found in PATH. Try installing tesseract")
	}

	// Execute binary 'tesseract imagePath stdout' to extract text from image
	detectedText, execErr := executeBinary(tesseractPath, imagePath, "stdout")
	if execErr != nil {
		return "", execErr
	}

	detectedText = strings.ReplaceAll(detectedText, "\n", " ")
	return detectedText, nil
}

func (c *TesseractOCRWrapper) IdentifyCurrentScreen(imagePath string) (bootScreenConfig BootScreenConfig, err error) {
	text, err := c.detectTextFromImageUsingTesseract(imagePath)
	if err != nil {
		return BootScreenConfig{}, err
	}

	log.Println("Detected text: ", text)
	return c.detectScreen(text), nil
}

func (c *TesseractOCRWrapper) RemoveBootScreenConfigIfExist(screenName string) {
	delete(c.ScreenConfigs, screenName)
}
