// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
)

// BuilderId is the common builder ID to all of these artifacts.
const BuilderId = "packer.parallels"

// These are the extensions of files and directories that are unnecessary for the function
// of a Parallels virtual machine.
var unnecessaryFiles = []string{"\\.log$", "\\.backup$", "\\.Backup$", "\\.app"}

// Artifact is the result of running the parallels builder, namely a set
// of files associated with the resulting machine.
type artifact struct {
	APIEndpoint     string
	SourceImageList string
	dir             string
	f               []string

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

// NewArtifact returns a Parallels artifact containing the files
// in the given directory.
func NewArtifact(dir string, generatedData map[string]interface{}, apiEndpoint string, sourceImageList string) (packersdk.Artifact, error) {
	files := make([]string, 0, 5)
	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				// It's possible that the path in question existed prior to visit being invoked, but has
				// since been deleted. This happens, for example, if you have the VM console open while
				// building. Opening the console creates a <vm name>.app directory which is gone by the time
				// visit is invoked, resulting in err being a "file not found" error. In this case, skip the
				// entire path and continue to the next one.
				return filepath.SkipDir
			}
			return err
		}
		for _, unnecessaryFile := range unnecessaryFiles {
			if unnecessary, _ := regexp.MatchString(unnecessaryFile, path); unnecessary {
				return os.RemoveAll(path)
			}
		}

		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	}

	if err := filepath.Walk(dir, visit); err != nil {
		return nil, err
	}

	return &artifact{
		APIEndpoint:     apiEndpoint,
		SourceImageList: sourceImageList,
		dir:             dir,
		f:               files,
		StateData:       generatedData,
	}, nil
}

func (*artifact) BuilderId() string {
	return BuilderId
}

func (a *artifact) Files() []string {
	return a.f
}

func (*artifact) Id() string {
	return "VM"
}

func (a *artifact) String() string {
	return fmt.Sprintf("VM files in directory: %s", a.dir)
}

func (a *artifact) State(name string) interface{} {
	log.Printf("[TRACE] Artifact State is %s", name)

	if name == image.ArtifactStateURI {
		return a.buildHCPackerRegistryMetadata()
	}
	return a.StateData[name]
}

func (a *artifact) Destroy() error {
	return os.RemoveAll(a.dir)
}

func (a *artifact) buildHCPackerRegistryMetadata() interface{} {
	region := a.APIEndpoint
	if region == "" {
		region = "local"
	}

	log.Printf("[TRACE] Region is %s", a.APIEndpoint)
	log.Printf("[TRACE] SourceId is %s", a.SourceImageList)

	img, err := image.FromArtifact(a, image.WithRegion(region), image.WithSourceID(a.SourceImageList))

	if err != nil {
		log.Printf("[TRACE] error encountered when creating HCP Packer registry image for artifact: %s", err)
		return nil
	}

	return img
}
