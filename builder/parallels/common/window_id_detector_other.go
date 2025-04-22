//go:build linux || windows
// +build linux windows

package common

import (
	"errors"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

// WindowIDDetector is a dummy struct for other platforms
type WindowIDDetector struct{}

func (s *WindowIDDetector) retrieveVMUUID(vmName string, state multistep.StateBag) (string, error) {
	return "", errors.New("retrieveVMUUID is not implemented other than darwin")
}

func (s *WindowIDDetector) getVMProcessID(uuid string) (int, error) {
	return 0, errors.New("getVMProcessID is not implemented other than darwin")
}

func (c *WindowIDDetector) DetectWindowId(vmName string, state multistep.StateBag) (int, error) {
	return 0, errors.New("DetectWindowId is not implemented other than darwin")
}
