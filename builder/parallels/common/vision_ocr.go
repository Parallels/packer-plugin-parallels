// Copyright (c) Parallels International GmBH
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown

package common

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework CoreImage -framework Vision

#import <Foundation/Foundation.h>
#import <CoreImage/CoreImage.h>
#import <Vision/Vision.h>

bool detectTextFromImage(const char *imagePath, char** detectedText, char **errorBuffer)
{
	@autoreleasepool {
		//Creating the image
		NSString *imagePathStr = [[NSString alloc] initWithCString:imagePath encoding:kUnicodeUTF8Format];
		NSURL* url = [[NSURL alloc] initFileURLWithPath:imagePathStr];
		CIImage *image = [[CIImage alloc] initWithContentsOfURL:url];

		if (!image)
		{
			*errorBuffer = strdup("an image does not exist at the path");
			return false;
		}

		//Recognizing text
		dispatch_semaphore_t sem = dispatch_semaphore_create(0);
		__block NSArray<VNRecognizedTextObservation *> *textObservations;
		VNRecognizeTextRequest *request = [[VNRecognizeTextRequest alloc] initWithCompletionHandler:^(VNRequest * _Nonnull request, NSError * _Nullable error) {
			if (error)
			{
				*errorBuffer = strdup(error.localizedDescription.UTF8String);
				dispatch_semaphore_signal(sem);
				return;
			}

			textObservations = request.results;
			dispatch_semaphore_signal(sem);
		}];
		request.recognitionLevel = VNRequestTextRecognitionLevelAccurate;

		NSError *error;
		VNImageRequestHandler *docRequestHandler = [[VNImageRequestHandler alloc] initWithCIImage:image options:@{}];
		[docRequestHandler performRequests:@[request] error:&error];

		//Wait for processing to complete
		dispatch_semaphore_wait(sem, DISPATCH_TIME_FOREVER);

		if (!textObservations)
			return false;

		NSString* finalString = [[NSString alloc] init];
		for (VNRecognizedTextObservation* observation in textObservations)
		{
			NSArray<VNRecognizedText*>* text = [observation topCandidates:1];
			if (text && text[0] && text[0].string.length > 0)
			{
				finalString = [finalString stringByAppendingString:text[0].string];
				finalString = [finalString stringByAppendingString:@" "];
			}
		}

		*detectedText = strdup(finalString.UTF8String);
		return true;
	}
}
*/
import "C"

import (
	"errors"
	"os/exec"
	"strings"
	"unsafe"
)

func (c *OCRRunner) executeBinary(binaryPath string, args ...string) (string, error) {
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

// OCRRunner is a struct that acts as an Adapter to Apple's Vision Framework for Optical Character Recognition.
type OCRRunner struct {
}

func (c *OCRRunner) detectTextFromImageUsingTesseract(imagePath string) (string, error) {
	tesseractPath, err := exec.LookPath("tesseract")
	if err != nil {
		return "", errors.New("tesseract binary not found in PATH. Please install tesseract.")
	}

	// Execute binary 'tesseract imagePath stdout' to extract text from image
	detectedText, execErr := c.executeBinary(tesseractPath, imagePath, "stdout")
	if execErr != nil {
		return "", execErr
	}

	detectedText = strings.ReplaceAll(detectedText, "\n", " ")
	return detectedText, nil
}

func (c *OCRRunner) detectTextFromImageUsingVisionAPI(imagePath string, text chan<- string, err chan<- error) {
	detectedText := C.CString("")
	defer C.free(unsafe.Pointer(detectedText))
	errorBuffer := C.CString("")
	defer C.free(unsafe.Pointer(errorBuffer))

	result := C.detectTextFromImage(C.CString(imagePath), &detectedText, &errorBuffer)
	if !result {
		text <- ""
		err <- errors.New(C.GoString(errorBuffer))
		return
	}

	text <- C.GoString(detectedText)
	err <- nil
}

// Extracts text from the image at the given path. Uses "Accurate" recognition level.
func (c *OCRRunner) RecognizeText(imagePath string, library string) (text string, err error) {
	if library == "tesseract" {
		// Use Tesseract for OCR
		return c.detectTextFromImageUsingTesseract(imagePath)
	}

	if library == "vision" {
		// Use Vision Framework for OCR
		chanText := make(chan string)
		chanErr := make(chan error)
		go c.detectTextFromImageUsingVisionAPI(imagePath, chanText, chanErr)
		return <-chanText, <-chanErr
	}

	return "", errors.New("invalid OCR library")
}
