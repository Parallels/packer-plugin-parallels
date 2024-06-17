#import <Foundation/Foundation.h>
#import <CoreImage/CoreImage.h>
#import <Vision/Vision.h>
#import <AppKit/AppKit.h>

@interface BootScreenConfig : NSObject
@property(readonly) NSMutableArray<NSString*>* matchingStrings;
@end

@implementation BootScreenConfig
- (id) init
{
	if (self = [super init]) {
		_matchingStrings = [[NSMutableArray alloc] init];
	}
	return self;
}

- (void) addMatchingString:(const char*)matchingString
{
	[_matchingStrings addObject:[NSString stringWithUTF8String:matchingString]];
}
@end

@interface OCRImpl : NSObject
@end

@implementation OCRImpl
{
	//Array of matching strings of screen configurations - for verification purposes
	NSMutableArray<BootScreenConfig*>* screenConfigs;
}
- (id) init
{
	if (self = [super init]) {
		screenConfigs = [[NSMutableArray alloc] init];
	}
	return self;
}

- (void) addBootScreenConfig:(BootScreenConfig*)config
{
	[screenConfigs addObject:config];
}

- (NSMutableArray<NSMutableArray<NSNumber *>*>*) createCheckBoxesForScreenConfigs
{
	NSMutableArray* ret = [[NSMutableArray alloc] init];
	for (NSArray* config in screenConfigs)
	{
		NSMutableArray* checks = [[NSMutableArray alloc] init];
		BootScreenConfig *bsc = (BootScreenConfig*)config;
		for (int i = 0; i < bsc.matchingStrings.count; i++)
		{
			[checks addObject:[[NSNumber alloc] initWithBool:NO]];
		}

		[ret addObject:checks];
	}

	return ret;
}

- (int) checkForMatches:(NSString*)recognizedText checkBox:(NSMutableArray<NSMutableArray<NSNumber *>*>*)checkBox
{
	int emptyBootScreenIndex = -1;
	for (int i = 0; i < screenConfigs.count; i++)
	{
		if (screenConfigs[i].matchingStrings.count == 0)
		{
			emptyBootScreenIndex = i;
			continue;
		}

		bool match = true;
		for (int j = 0; j < screenConfigs[i].matchingStrings.count; j++)
			if (![checkBox[i][j] boolValue])
			{
				if ([recognizedText containsString:screenConfigs[i].matchingStrings[j]])
					checkBox[i][j] = @YES;
				else
					match = false;
			}

		if (match)
			return i;
	}

	return emptyBootScreenIndex;
}

-(NSArray<VNRecognizedTextObservation *>*) recognizeText:(CIImage*)image errorBuffer:(char **)errorBuffer
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
	//request.revision = VNRecognizeTextRequestRevision3;

	NSError *error;
	VNImageRequestHandler *docRequestHandler = [[VNImageRequestHandler alloc] initWithCIImage:image options:@{}];
	[docRequestHandler performRequests:@[request] error:&error];

	//Wait for processing to complete
	dispatch_semaphore_wait(sem, DISPATCH_TIME_FOREVER);

	return textObservations;
}

- (CIFilter*) scaleFilter:(CIImage*)image factor:(NSNumber*)factor
{
	CIFilter *scaleFilter = [CIFilter filterWithName:@"CILanczosScaleTransform"];
	[scaleFilter setValue:image forKey:kCIInputImageKey];
	[scaleFilter setValue:factor forKey:kCIInputScaleKey];
	[scaleFilter setValue:@1.0 forKey:kCIInputAspectRatioKey]; // Maintain aspect ratio

	return scaleFilter;
}


/**
 imagePath - image to run OCR
 refScaleFactor [should be 0.0 or above] - Represents the image scale factor to improve OCR accuracy.
 if 0.0, OCR accuracy will be checked by resizing image based on width from 1000-2000px, best result will be returned.
 if >0.0, OCR accuracy will be checked from refScaleFactor ± 10% and best result will be returned
 errorBuffer - returns with error text if any
 topConfidenceScaleFactor - returns the scale factor resulted in best OCR accuracy. You can use this as refScaleFactor for next calls with similar images

 @returns index of matching boot screen. '-1/empty screen index' if none of them are matching/error
 */
- (int) detectScreenFrom:(const char *)imagePath ScaleFactor:(double)refScaleFactor BestRecognizedText:(char**)textBuffer ErrorBuffer:(char **)errorBuffer TopScaleFactor:(double*)topConfidenceScaleFactor
{
	@autoreleasepool {
		//Creating the image
		NSString* imagePathStr = [[NSString alloc] initWithCString:imagePath encoding:kUnicodeUTF8Format];
		NSURL* url = [[NSURL alloc] initFileURLWithPath:imagePathStr];
		CIImage* image = [[CIImage alloc] initWithContentsOfURL:url];
		if (!image)
		{
			*errorBuffer = strdup("an image does not exist at the path");
			return -1;
		}

		// Get the extent of the image
		CGRect extent = [image extent];
		// Extract the width and height
		CGFloat width = CGRectGetWidth(extent);
		// Need to iterate width between 1000-2000px to get accurate OCR
		CGFloat begin = 1000 / width;
		CGFloat end = 2000 / width;
		if (refScaleFactor > 0.0)
		{
			begin = refScaleFactor - 0.1;
			end = refScaleFactor + 0.1;
		}

		int result = -1;
		CGFloat topConfidence = 0.0;
		CGFloat topConfidenceScalingFactor = 0.0;
		NSString *bestRecognizedText = NSString.string;
		NSMutableArray<NSMutableArray<NSNumber *>*>* checkBox = [self createCheckBoxesForScreenConfigs];

		for (; begin < end; begin += 0.01)
		{
			NSNumber* scaleFactor = [[NSNumber alloc] initWithFloat:begin];
			// Get the scaled image
			CIFilter* filter = [self scaleFilter:image factor:scaleFactor];
			CIImage* scaledImage = [filter outputImage];
			NSArray<VNRecognizedTextObservation *> * textObservations = [self recognizeText:scaledImage errorBuffer:errorBuffer];
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

			result = [self checkForMatches:[recognizedText lowercaseString] checkBox:checkBox];
			if (result != -1 && screenConfigs[result].matchingStrings.count != 0)
			{
				topConfidenceScalingFactor = begin;
				bestRecognizedText = @"";
				break;
			}

			if (currentConfidence > topConfidence)
			{
				topConfidence = currentConfidence;
				topConfidenceScalingFactor = begin;
				bestRecognizedText = recognizedText;
			}
		}

		if (topConfidenceScaleFactor)
			*topConfidenceScaleFactor = topConfidenceScalingFactor;
		*textBuffer = strdup(bestRecognizedText.UTF8String);
		return result;
	}
}

@end

#ifdef __cplusplus
extern "C" {
#endif


static OCRImpl* newOCRImpl()
{
	OCRImpl* impl = [[OCRImpl alloc] init];
	return impl;
}

static BootScreenConfig* newBootScreenConfig()
{
	BootScreenConfig* bsc = [[BootScreenConfig alloc] init];
	return bsc;
}

static int detectScreenFromImage(OCRImpl* impl, const char* imagePath, double refScaleFactor, char** bestRecognizedText, char** errorBuffer, double* topConfidenceScaleFactor)
{
	return [impl detectScreenFrom:imagePath ScaleFactor:refScaleFactor BestRecognizedText:bestRecognizedText ErrorBuffer:errorBuffer TopScaleFactor:topConfidenceScaleFactor];
}

static void addMatchingString(BootScreenConfig *config, const char* mStr)
{
	[config addMatchingString:mStr];
}

static void addBootScreenConfig(OCRImpl* impl, BootScreenConfig* config)
{
	[impl addBootScreenConfig:config];
}

#ifdef __cplusplus
}
#endif
