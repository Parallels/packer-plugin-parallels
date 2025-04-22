// Copyright (c) Parallels International GmBH
// SPDX-License-Identifier: MPL-2.0

//go:build !darwin
// +build !darwin

package common

import "errors"

// This is a dummy implementation of VisionOCRWrapper for platforms other than macOS
// To avoid validation errors, we need to implement the interface
// but it will not perform any actual OCR operations.
type VisionOCRWrapper struct {
}

func (c *VisionOCRWrapper) IdentifyCurrentScreen(imagePath string) (bootScreenConfig BootScreenConfig, err error) {
	return nil, errors.New("VisionOCRWrapper is not implemented other than darwin")
}

func (c *VisionOCRWrapper) RemoveBootScreenConfigIfExist(screenName string) {
	// No operation
}

func NewVisionOCRWrapper(ScreenConfigs map[string]BootScreenConfig) *VisionOCRWrapper {
	return &VisionOCRWrapper{}
}
