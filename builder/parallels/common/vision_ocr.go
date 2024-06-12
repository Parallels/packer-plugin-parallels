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

NSArray<VNRecognizedTextObservation *> * recognizeText(CIImage *image, char **errorBuffer)
{
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

	return textObservations;
}

CIFilter* scaleFilter(CIImage* image, NSNumber* factor)
{
	// Create the CILanczosScaleTransform filter
	CIFilter *scaleFilter = [CIFilter filterWithName:@"CILanczosScaleTransform"];
	[scaleFilter setValue:image forKey:kCIInputImageKey];
	[scaleFilter setValue:factor forKey:kCIInputScaleKey]; // Scale down to 70%
	[scaleFilter setValue:@1.0 forKey:kCIInputAspectRatioKey]; // Maintain aspect ratio

	return scaleFilter;
}

// imagePath - image to run OCR
// refScaleFactor [should be 0.0 or above] - Represents the image scale factor to improve OCR accuracy.
//			if 0.0, OCR accuracy will be checked by resizing image based on width from 1000-2000px, best result will be returned.
//			if >0.0, OCR accuracy will be checked from refScaleFactor ± 5% and best result will be returned
// detectedText - returns best result, multiple lines will be appended with whitespace
// errorBuffer - returns with error text if any
// topConfidenceScaleFactor - returns the scale factor resulted in best OCR accuracy. You can use this as refScaleFactor for next calls with similar images
bool detectTextFromImage(const char *imagePath, double refScaleFactor, char** detectedText, char **errorBuffer, double* topConfidenceScaleFactor)
{
 @autoreleasepool {
	 //Creating the image
	 NSString* imagePathStr = [[NSString alloc] initWithCString:imagePath encoding:kUnicodeUTF8Format];
	 NSURL* url = [[NSURL alloc] initFileURLWithPath:imagePathStr];
	 CIImage* image = [[CIImage alloc] initWithContentsOfURL:url];
	 if (!image)
	 {
		 *errorBuffer = strdup("an image does not exist at the path");
		 return false;
	 }
	 // Get the extent of the image
	 CGRect extent = [image extent];
	 // Extract the width and height
	 CGFloat width = CGRectGetWidth(extent);
	 //Need to iterate width between 1000-2000px to get accurate OCR
	 CGFloat begin = 1000 / width;
	 CGFloat end = 2000 / width;
	 if (refScaleFactor > 0.0)
	 {
		 begin = refScaleFactor - 0.05;
		 end = refScaleFactor + 0.05;
	 }
	 CGFloat topConfidence = 0.0;
	 CGFloat topConfidenceScalingFactor = 0.0;
	 NSString* topConfidenceString = NSString.string;
	 for (; begin < end; begin += 0.01)
	 {
		 NSNumber* scaleFactor = [[NSNumber alloc] initWithFloat:begin];
		 // Get the scaled image
		 CIFilter* filter = scaleFilter(image, scaleFactor);
		 CIImage* scaledImage = [filter outputImage];
		 NSArray<VNRecognizedTextObservation *> * textObservations = recognizeText(scaledImage, errorBuffer);
		 if (!textObservations)
			 return false;
		 CGFloat currentConfidence = 0.0;
		 NSString* recognizedText = NSString.string;
		 for (VNRecognizedTextObservation* observation in textObservations)
		 {
			 NSArray<VNRecognizedText*>* text = [observation topCandidates:1];
			 if (text && text[0] && text[0].string.length > 0)
			 {
				 recognizedText = [recognizedText stringByAppendingString:text[0].string];
				 recognizedText = [recognizedText stringByAppendingString:@" "];
				 currentConfidence += observation.confidence;
			 }
		 }
		 if (currentConfidence > topConfidence)
		 {
			 topConfidence = currentConfidence;
			 topConfidenceScalingFactor = begin;
			 topConfidenceString = recognizedText;
		 }
	 }
	 *detectedText = strdup(topConfidenceString.UTF8String);
	 if (topConfidenceScaleFactor)
	     *topConfidenceScaleFactor = topConfidenceScalingFactor;
	 return true;
 }
}
*/
import "C"

import (
	"errors"
	"log"
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
	referenceScalingFactor float32
}

func NewOCRRunner() *OCRRunner {
	return &OCRRunner{referenceScalingFactor: 0.0}
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

	refScalingFactor := C.double(c.referenceScalingFactor)
	bestScalingFactor := C.double(0.0)
	log.Printf("refScalingFactor: %f", c.referenceScalingFactor)
	result := C.detectTextFromImage(C.CString(imagePath), refScalingFactor, &detectedText, &errorBuffer, &bestScalingFactor)
	if !result {
		text <- ""
		err <- errors.New(C.GoString(errorBuffer))
		return
	}
	c.referenceScalingFactor = float32(bestScalingFactor)

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
