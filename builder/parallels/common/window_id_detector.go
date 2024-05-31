// Copyright (c) Parallels International GmBH
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown

package common

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework CoreGraphics

#import <Foundation/Foundation.h>
#import <CoreGraphics/CoreGraphics.h>

int retrieveWindowId(int pid)
{
	@autoreleasepool {
		// Get a list of all windows
		CFArrayRef windowList = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly, kCGNullWindowID);
		// Iterate through the window list
		for (NSDictionary *windowInfo in (__bridge NSArray *)windowList) {
			NSNumber *pidOfProcess = windowInfo[(id)kCGWindowOwnerPID];
			// Check if the pid is same
			if (pidOfProcess.intValue == pid) {
				// Get the window ID
				NSNumber *windowID = windowInfo[(id)kCGWindowNumber];
				return windowID.intValue;
			}
		}

		// Release the window list
		CFRelease(windowList);
	}

	return 0;
}
*/
import "C"
import (
	"bytes"
	"errors"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

// VisionOCR is a struct that acts as an Adapter to Apple's Vision Framework for Optical Character Recognition.
type WindowIDDetector struct {
}

func (s *WindowIDDetector) retrieveVMUUID(vmName string, state multistep.StateBag) (string, error) {
	driver := state.Get("driver").(Driver)
	result, err := driver.PrlctlGet("list", vmName, "--no-header", "-o", "uuid")
	if err != nil {
		return "", err
	}

	stdoutString := strings.TrimSpace(result)
	if len(stdoutString) == 0 {
		return "", errors.New("no uuid found for : " + vmName)
	}

	return stdoutString, nil
}

func (s *WindowIDDetector) getVMProcessID(uuid string) (int, error) {
	c1 := exec.Command("ps", "aux")
	c2 := exec.Command("grep", "prl_macvm_app")
	c3 := exec.Command("grep", uuid)
	c4 := exec.Command("awk", "{print $2}")

	r1, w1 := io.Pipe()
	c1.Stdout = w1
	c2.Stdin = r1

	r2, w2 := io.Pipe()
	c2.Stdout = w2
	c3.Stdin = r2

	r3, w3 := io.Pipe()
	c3.Stdout = w3
	c4.Stdin = r3

	var buf bytes.Buffer
	c4.Stdout = &buf

	// Start the commands
	err := c1.Start()
	if err != nil {
		return 0, err
	}

	err = c2.Start()
	if err != nil {
		return 0, err
	}

	err = c3.Start()
	if err != nil {
		return 0, err
	}

	err = c4.Start()
	if err != nil {
		return 0, err
	}

	err = c1.Wait()
	if err != nil {
		return 0, err
	}

	err = w1.Close()
	if err != nil {
		return 0, err
	}

	err = c2.Wait()
	if err != nil {
		return 0, err
	}

	err = w2.Close()
	if err != nil {
		return 0, err
	}

	err = c3.Wait()
	if err != nil {
		return 0, err
	}

	err = w3.Close()
	if err != nil {
		return 0, err
	}

	err = c4.Wait()
	if err != nil {
		return 0, err
	}

	// Capture the output
	output, err := buf.ReadString('\n')
	if err != nil {
		return 0, err
	}

	// Convert output to int and return
	return strconv.Atoi(strings.TrimSpace(output))
}

// DetectWindowId is a method that detects the window ID of a given VM.
func (c *WindowIDDetector) DetectWindowId(vmName string, state multistep.StateBag) (id int, err error) {
	// Get the UUID of the VM
	uuid, err := c.retrieveVMUUID(vmName, state)
	if err != nil {
		return 0, err
	}
	log.Println("VM UUID: ", uuid)
	// Get the process ID of the VM
	pid, err := c.getVMProcessID(uuid)
	if err != nil {
		return 0, err
	}
	log.Println("VM process ID: ", pid)
	// Get the window ID of the VM
	id = int(C.retrieveWindowId(C.int(pid)))
	if id == 0 {
		return 0, errors.New("no window id found for the process")
	}

	log.Println("VM window ID: ", id)
	return id, nil
}
