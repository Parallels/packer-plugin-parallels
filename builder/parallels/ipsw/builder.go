// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package ipsw

import (
	"context"
	"errors"
	"fmt"
	"os"

	parallelscommon "github.com/Parallels/packer-plugin-parallels/builder/parallels/common"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/shutdowncommand"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

const BuilderId = "rickard-von-essen.parallels"

type Builder struct {
	config Config
	runner multistep.Runner
}

type Config struct {
	common.PackerConfig                 `mapstructure:",squash"`
	commonsteps.HTTPConfig              `mapstructure:",squash"`
	bootcommand.BootConfig              `mapstructure:",squash"`
	parallelscommon.OutputConfig        `mapstructure:",squash"`
	parallelscommon.HWConfig            `mapstructure:",squash"`
	parallelscommon.PrlctlConfig        `mapstructure:",squash"`
	parallelscommon.PrlctlPostConfig    `mapstructure:",squash"`
	parallelscommon.PrlctlVersionConfig `mapstructure:",squash"`
	shutdowncommand.ShutdownConfig      `mapstructure:",squash"`
	parallelscommon.SSHConfig           `mapstructure:",squash"`

	// Screens and it's boot configs
	// A screen is considered matched if all the matching strings are present in the screen.
	// The first matching screen will be considered & boot config of that screen will be used.
	// If matching strings are empty, then it is considered as empty screen,
	// which will be considered when none of the other screens are matched (You can use this screen to -
	// make system wait for some time / execute a common boot command etc.).
	// If more than one empty screen is found, then it is considered as an error.
	BootScreenConfig parallelscommon.BootScreensConfig `mapstructure:"boot_screen_config" required:"false"`
	// IPSWConfig is the configuration for the IPSW file
	IPSWConfig IPSWConfig `mapstructure:",squash"`
	// The size, in megabytes, of the hard disk to create
	// for the VM. By default, this is 40000 (about 40 GB).
	DiskSize uint `mapstructure:"disk_size" required:"false"`
	// A list of which interfaces on the
	// host should be searched for a IP address. The first IP address found on one
	// of these will be used as `{{ .HTTPIP }}` in the boot_command. Defaults to
	// ["en0", "en1", "en2", "en3", "en4", "en5", "en6", "en7", "en8", "en9",
	// "en10", "en11", "en12", "en13", "en14", "en15", "en16", "en17", "en18", "en19", "en20",
	// "en10", "en11", "en12", "en13", "en14", "en15", "en16", "en17", "en18", "en19", "en20",
	// "ppp0", "ppp1", "ppp2"].
	HostInterfaces []string `mapstructure:"host_interfaces" required:"false"`
	// This is the name of the PVM directory for the new
	// virtual machine, without the file extension. By default this is
	// "packer-BUILDNAME", where "BUILDNAME" is the name of the build.
	VMName string `mapstructure:"vm_name" required:"false"`

	ctx interpolate.Context
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
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
		return nil, nil, err
	}

	// Accumulate any errors and warnings
	var errs *packersdk.MultiError
	warnings := make([]string, 0)

	ipswWarnings, ipswErrs := b.config.IPSWConfig.Prepare(&b.config.ctx)
	warnings = append(warnings, ipswWarnings...)
	errs = packersdk.MultiErrorAppend(errs, ipswErrs...)

	errs = packersdk.MultiErrorAppend(errs, b.config.HTTPConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(
		errs, b.config.OutputConfig.Prepare(&b.config.ctx, &b.config.PackerConfig)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.HWConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.PrlctlConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.PrlctlPostConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.PrlctlVersionConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.ShutdownConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.SSHConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.BootConfig.Prepare(&b.config.ctx)...)

	if b.config.DiskSize == 0 {
		b.config.DiskSize = 40000
	}

	if len(b.config.HostInterfaces) == 0 {
		b.config.HostInterfaces = []string{
			"en0", "en1", "en2", "en3", "en4", "en5", "en6", "en7",
			"en8", "en9", "ppp0", "ppp1", "ppp2",
		}
	}

	if b.config.VMName == "" {
		b.config.VMName = fmt.Sprintf("packer-%s", b.config.PackerBuildName)
	}

	// Warnings
	if b.config.ShutdownCommand == "" {
		warnings = append(warnings,
			"A shutdown_command was not specified. Without a shutdown command, Packer\n"+
				"will forcibly halt the virtual machine, which may result in data loss.")
	}

	fmt.Fprintln(os.Stderr, "Screen count is : ", len(b.config.BootScreenConfig))

	emptyScreenCount := 0
	for _, screenConfig := range b.config.BootScreenConfig {
		errs = packersdk.MultiErrorAppend(errs, screenConfig.Prepare(&b.config.ctx)...)
		if len(screenConfig.MatchingStrings) == 0 {
			emptyScreenCount++
		}
	}

	if emptyScreenCount > 1 {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("more than one empty screen config found."+
			"only one empty screen config is allowed"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return nil, warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	// Create the driver that we'll use to communicate with Parallels
	driver, err := parallelscommon.NewDriver()
	if err != nil {
		return nil, fmt.Errorf("Failed creating Parallels driver: %s", err)
	}

	steps := []multistep.Step{
		&commonsteps.StepDownload{
			Checksum:    b.config.IPSWConfig.IPSWChecksum,
			Description: "IPSW",
			Extension:   "ipsw",
			ResultKey:   "ipsw_path",
			TargetPath:  b.config.IPSWConfig.TargetPath,
			Url:         b.config.IPSWConfig.IPSWUrls,
		},
		&parallelscommon.StepOutputDir{
			Force: b.config.PackerForce,
			Path:  b.config.OutputDir,
		},
		commonsteps.HTTPServerFromHTTPConfig(&b.config.HTTPConfig),
		new(stepCreateVM),
		&parallelscommon.StepPrlctl{
			Commands: b.config.Prlctl,
			Ctx:      b.config.ctx,
		},
		&parallelscommon.StepRun{},
		&parallelscommon.StepTypeBootCommand{
			BootWait:       b.config.BootWait,
			BootCommand:    b.config.FlatBootCommand(),
			HostInterfaces: b.config.HostInterfaces,
			VMName:         b.config.VMName,
			Ctx:            b.config.ctx,
			GroupInterval:  b.config.BootConfig.BootGroupInterval,
		},
		&parallelscommon.StepScreenBasedBoot{
			ScreenConfigs: b.config.BootScreenConfig,
			VmName:        b.config.VMName,
			Ctx:           b.config.ctx,
		},
		&communicator.StepConnect{
			Config:    &b.config.SSHConfig.Comm,
			Host:      parallelscommon.CommHost(b.config.SSHConfig.Comm.Host()),
			SSHConfig: b.config.SSHConfig.Comm.SSHConfigFunc(),
		},
	}

	// Add post-communitcation steps
	if b.config.SSHConfig.Comm.Type != "none" {
		steps = append(steps, []multistep.Step{
			&parallelscommon.StepUploadVersion{
				Path: b.config.PrlctlVersionFile,
			},
			new(commonsteps.StepProvision),
			&commonsteps.StepCleanupTempKeys{
				Comm: &b.config.SSHConfig.Comm,
			},
			&parallelscommon.StepShutdown{
				Command: b.config.ShutdownCommand,
				Timeout: b.config.ShutdownTimeout,
			},
		}...)
	}

	steps = append(steps, []multistep.Step{
		&parallelscommon.StepPrlctl{
			Commands: b.config.PrlctlPost,
			Ctx:      b.config.ctx,
		},
	}...)

	// Setup the state bag
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Run
	b.runner = commonsteps.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("Build was halted.")
	}

	generatedData := map[string]interface{}{"generated_data": state.Get("generated_data")}
	return parallelscommon.NewArtifact(b.config.OutputDir, generatedData)
}
