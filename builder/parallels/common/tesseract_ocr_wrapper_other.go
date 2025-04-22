// Copyright (c) Parallels International GmBH
// SPDX-License-Identifier: MPL-2.0

//go:build !darwin
// +build !darwin

package common

import "errors"

// This is a dummy implementation of TesseractOCRWrapper for platforms other than macOS
// To avoid validation errors, we need to implement the interface
// but it will not perform any actual OCR operations.

type TesseractOCRWrapper struct {
}

func NewTesseractOCRWrapper(ScreenConfigs map[string]BootScreenConfig) *TesseractOCRWrapper {
	return &TesseractOCRWrapper{}
}

func (c *TesseractOCRWrapper) IdentifyCurrentScreen(imagePath string) (bootScreenConfig BootScreenConfig, err error) {

	return nil, errrors.New("TesseractOCRWrapper is not implemented other than darwin")
}

func (c *TesseractOCRWrapper) RemoveBootScreenConfigIfExist(screenName string) {
	// No operation
}

func executeBinary(binaryPath string, args ...string) (string, error) {
	// This is a dummy implementation for non-macOS platforms
	return "", errors.New("executeBinary is not implemented for non-macOS platforms")
}
