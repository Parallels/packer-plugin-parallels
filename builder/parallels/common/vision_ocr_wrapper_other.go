// Copyright (c) Parallels International GmBH
// SPDX-License-Identifier: MPL-2.0

package common

// This is a dummy implementation of OCRWrapper for platforms other than macOS
// To avoid validation errors, we need to implement the interface
// but it will not perform any actual OCR operations.
type OCRWrapperOther struct {
}

func (c *OCRWrapperOther) IdentifyCurrentScreen(imagePath string) (bootScreenConfig BootScreenConfig, err error) {
	return BootScreenConfig{}, nil
}

func (c *OCRWrapperOther) RemoveBootScreenConfigIfExist(screenName string) {
	// No operation
}

func NewOCRWrapperOthers() (OCRWrapper, error) {
	return &OCRWrapperOther{}, nil
}
