// Copyright (c) Parallels International GmBH
// SPDX-License-Identifier: MPL-2.0

package common

import "errors"

type OCRWrapper interface {
	// IdentifyCurrentScreen takes a screenshot of the current screen and returns the BootScreenConfig that matches the screen.
	IdentifyCurrentScreen(imagePath string) (bootScreenConfig BootScreenConfig, err error)
	// Removes the screen config with specified name if exist
	RemoveBootScreenConfigIfExist(screenName string)
}

func NewOCRWrapper(OCRLibrary string, ScreenConfigs map[string]BootScreenConfig) (OCRWrapper, error) {
	if OCRLibrary == "vision" {
		return NewVisionOCRWrapper(ScreenConfigs), nil
	} else if OCRLibrary == "tesseract" {
		return NewTesseractOCRWrapper(ScreenConfigs), nil
	} else {
		return nil, errors.New("invalid OCR library")
	}
}
