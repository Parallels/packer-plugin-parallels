// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package macvm

import (
	"fmt"
	"os"

	parallelscommon "github.com/Parallels/packer-plugin-parallels/builder/parallels/common"
	"github.com/hashicorp/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/shutdowncommand"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// Config is the configuration structure for the builder.
type Config struct {
	common.PackerConfig                 `mapstructure:",squash"`
	parallelscommon.OutputConfig        `mapstructure:",squash"`
	parallelscommon.PrlctlConfig        `mapstructure:",squash"`
	parallelscommon.PrlctlPostConfig    `mapstructure:",squash"`
	parallelscommon.PrlctlVersionConfig `mapstructure:",squash"`
	parallelscommon.SSHConfig           `mapstructure:",squash"`
	shutdowncommand.ShutdownConfig      `mapstructure:",squash"`
	bootcommand.BootConfig              `mapstructure:",squash"`

	// Screens and it's boot configs
	// A screen is considered matched if all the matching strings are present in the screen.
	// The first matching screen will be considered & boot config of that screen will be used.
	// If matching strings are empty, then it is considered as empty screen,
	// empty screen has some special meaning, which will be considered when none of the other screens are matched.
	// You can use this screen to make system wait for some time / execute a common boot command etc.
	// The empty screen boot command will be executed repeatedly until a non-empty screen is found.
	// If more than one empty screen is found, then it is considered as an error.
	BootScreenConfig parallelscommon.BootScreensConfig `mapstructure:"boot_screen_config" required:"false"`
	// OCR library to use. Two options are currently supported: "tesseract" and "vision".
	// "tesseract" uses the Tesseract OCR library to recognize text. A manual installation of -
	// Tesseract is required for this to work.
	// "vision" uses the Apple Vision library to recognize text, which is included in macOS. It might -
	// cause problems in macOS 13 or older VMs.
	// Defaults to "vision".
	OCRLibrary string `mapstructure:"ocr_library" required:"false"`
	// The path to a MACVM directory that acts as the source
	// of this build.
	SourcePath string `mapstructure:"source_path" required:"true"`
	// This is the name of the MACVM directory for the new
	// virtual machine, without the file extension. By default this is
	// "packer-BUILDNAME", where "BUILDNAME" is the name of the build.
	VMName string `mapstructure:"vm_name" required:"false"`
	// If this is "false" the MAC address of the first
	// NIC will reused when imported else a new MAC address will be generated
	// by Parallels. Defaults to "false".
	ReassignMAC bool `mapstructure:"reassign_mac" required:"false"`

	ctx              interpolate.Context
	screenConfigsMap map[string]parallelscommon.BootScreenConfig
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(c, &config.DecodeOpts{
		PluginType:         parallelscommon.BuilderId,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
				"prlctl",
				"prlctl_post",
				"parallels_tools_guest_path",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	if c.VMName == "" {
		c.VMName = fmt.Sprintf("packer-%s-{{timestamp}}", c.PackerBuildName)
	}

	// Prepare the errors
	var errs *packersdk.MultiError
	errs = packersdk.MultiErrorAppend(errs, c.OutputConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packersdk.MultiErrorAppend(errs, c.PrlctlConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.PrlctlPostConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.PrlctlVersionConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.BootConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.ShutdownConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.SSHConfig.Prepare(&c.ctx)...)

	fmt.Fprintln(os.Stderr, "Screen count is : ", len(c.BootScreenConfig))

	emptyScreenCount := 0
	c.screenConfigsMap = make(map[string]parallelscommon.BootScreenConfig)
	for _, screenConfig := range c.BootScreenConfig {
		errs = packersdk.MultiErrorAppend(errs, screenConfig.Prepare(&c.ctx)...)
		if len(screenConfig.MatchingStrings) == 0 {
			emptyScreenCount++
		}

		if screenConfig.ScreenName == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("screen_name should not be empty"))
			continue
		}

		if _, exists := c.screenConfigsMap[screenConfig.ScreenName]; exists {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("multiple screens with same name: %s", screenConfig.ScreenName))
			continue
		}
		c.screenConfigsMap[screenConfig.ScreenName] = screenConfig
	}

	if emptyScreenCount > 1 {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("more than one empty screen config found."+
			"only one empty screen config is allowed"))
	}

	if c.OCRLibrary == "" {
		c.OCRLibrary = "vision"
	} else if c.OCRLibrary != "tesseract" && c.OCRLibrary != "vision" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("invalid ocr_library: %s", c.OCRLibrary))
	}

	if c.SourcePath == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("source_path is required"))
	} else {
		if _, err := os.Stat(c.SourcePath); err != nil {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("source_path is invalid: %s", err))
		}
	}

	// Warnings
	var warnings []string
	if c.ShutdownCommand == "" {
		warnings = append(warnings,
			"A shutdown_command was not specified. Without a shutdown command, Packer\n"+
				"will forcibly halt the virtual machine, which may result in data loss.")
	}

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}
