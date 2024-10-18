// Copyright (c) Parallels International GmBH
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type BootScreenConfig

package common

import (
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type BootScreensConfig []BootScreenConfig

type BootScreenConfig struct {
	// BootConfig for this screen
	bootcommand.BootConfig `mapstructure:",squash"`
	// Screen name to identify
	ScreenName string `mapstructure:"screen_name"`
	// Strings present in the screen to identify this screen
	MatchingStrings []string `mapstructure:"matching_strings"`
	// Specifies if the current screen is the last screen
	// Screen based boot will stop after this screen
	IsLastScreen bool `mapstructure:"is_last_screen"`
	// If true, the screen will be deleted after first execution
	// Default value is false
	ExecuteOnlyOnce bool `mapstructure:"execute_only_once"`
}

func (c *BootScreenConfig) Prepare(ctx *interpolate.Context) (errs []error) {
	// Prepare bootconfig first
	errs = append(errs, c.BootConfig.Prepare(ctx)...)

	if len(c.ScreenName) == 0 {
		errs = append(errs,
			fmt.Errorf("screen name should not be empty"))
	}

	if !c.ExecuteOnlyOnce {
		c.ExecuteOnlyOnce = false
	}

	if len(errs) > 0 {
		return errs
	}

	// Convert all matching strings to lowercase
	for i, matchingString := range c.MatchingStrings {
		c.MatchingStrings[i] = strings.ToLower(matchingString)
	}

	return nil
}

func (c *BootScreenConfig) FlatBootCommand() string {
	return strings.Join(c.BootCommand, "")
}
