// Copyright (c) Parallels International GmBH
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown

package common

/*
#cgo CFLAGS: -x objective-c -Wno-incompatible-pointer-types
#cgo LDFLAGS: -framework Foundation -framework CoreImage -framework Vision -framework AppKit

#import "vision_ocr.mm"
*/
import "C"

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"unsafe"
)

// VisionOCRWrapper is a struct that acts as an Adapter to Apple's Vision Framework for Optical Character Recognition.
type VisionOCRWrapper struct {
	referenceScalingFactor float32
	ScreenConfigs          map[string]BootScreenConfig
	OCRImpl                *C.OCRImpl
}

func NewVisionOCRWrapper(ScreenConfigs map[string]BootScreenConfig) *VisionOCRWrapper {
	ocrRunner := VisionOCRWrapper{
		referenceScalingFactor: 0.0,
		ScreenConfigs:          ScreenConfigs,
		OCRImpl:                C.newOCRImpl(),
	}

	for _, screenConfig := range ScreenConfigs {
		bootScreenConfig := C.newBootScreenConfig()

		// Add the screen name
		C.setScreenName(bootScreenConfig, C.CString(screenConfig.ScreenName))
		for _, matchingString := range screenConfig.MatchingStrings {
			lowerCase := strings.ToLower(matchingString)
			C.addMatchingString(bootScreenConfig, C.CString(lowerCase))
		}

		C.addBootScreenConfig(ocrRunner.OCRImpl, bootScreenConfig)
	}

	return &ocrRunner
}

func (c *VisionOCRWrapper) detectScreenFromImageUsingVisionAPI(imagePath string, useRefScalingFactor bool, screenName chan<- string, err chan<- error) {
	errorBuffer := C.CString("")
	defer C.free(unsafe.Pointer(errorBuffer))
	textBuffer := C.CString("")
	defer C.free(unsafe.Pointer(textBuffer))
	detectedScreenName := C.CString("")
	defer C.free(unsafe.Pointer(detectedScreenName))

	refScalingFactor := C.double(0.0)
	if useRefScalingFactor {
		refScalingFactor = C.double(c.referenceScalingFactor)
	}

	bestScalingFactor := C.double(0.0)

	C.detectScreenFromImage(c.OCRImpl, C.CString(imagePath), refScalingFactor, &textBuffer, &errorBuffer, &bestScalingFactor, &detectedScreenName)
	result := C.GoString(detectedScreenName)
	if len(result) == 0 {
		screenName <- ""
		err <- errors.New(C.GoString(errorBuffer))
		return
	}
	c.referenceScalingFactor = float32(bestScalingFactor)

	// Printing the textBuffer helps users to fine tune their screen configurations
	if C.GoString(textBuffer) != "" {
		log.Printf("Best detected text: %s", C.GoString(textBuffer))
	}

	screenName <- result
	err <- nil
}

// Extracts text from the image at the given path. Uses "Accurate" recognition level.
func (c *VisionOCRWrapper) IdentifyCurrentScreen(imagePath string) (bootScreenConfig BootScreenConfig, err error) {
	// A reference scaling factor will be tried first.
	// If it fails, the scaling factor will be set to 0.0 and the OCR will be tried again.
	useRefScalingFactor := true
	screenConfig := BootScreenConfig{}

	for {
		// Use Vision Framework for OCR
		chanScreenName := make(chan string)
		chanErr := make(chan error)
		go c.detectScreenFromImageUsingVisionAPI(imagePath, useRefScalingFactor, chanScreenName, chanErr)
		screenName := <-chanScreenName
		err := <-chanErr

		fmt.Println("Detected screen name", screenName)
		if err != nil || len(screenName) == 0 {
			fmt.Println("Screen not detected (error if any:", err, ")")
			return BootScreenConfig{}, err
		}

		exists := false
		screenConfig, exists = c.ScreenConfigs[screenName]
		if !exists {
			return BootScreenConfig{}, errors.New("detected screen not found in the configuration, please report this bug")
		}

		// Did we detect an empty screen ? Try again without scaling factor to be sure
		if useRefScalingFactor && (screenConfig.MatchingStrings == nil || len(screenConfig.MatchingStrings) == 0) {
			useRefScalingFactor = false
		} else {
			break
		}
	}

	fmt.Println("Detected screen name 2", screenConfig.ScreenName)
	return screenConfig, nil
}

func (c *VisionOCRWrapper) RemoveBootScreenConfigIfExist(screenName string) {
	C.deleteBootScreenConfig(c.OCRImpl, C.CString(screenName))
}
