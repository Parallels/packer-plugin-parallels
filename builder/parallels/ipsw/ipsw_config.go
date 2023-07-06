// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown

package ipsw

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// Adapter class : This class acts same as ISOConfig. IPSWConfig is created
// to clearly communicate to users that macvms supports ipsw format files only.
// By default, Packer will symlink, download or copy image files to the Packer
// cache into a "`hash($ipsw_url+$ipsw_checksum).$ipsw_target_extension`" file.
// Packer uses [hashicorp/go-getter](https://github.com/hashicorp/go-getter) in
// file mode in order to perform a download.
//
// go-getter supports the following protocols:
//
// * Local files
// * Git
// * Mercurial
// * HTTP
// * Amazon S3
//
// Examples:
// go-getter can guess the checksum type based on `ipsw_checksum` length, and it is
// also possible to specify the checksum type.
//
// In JSON:
//
// ```json
//
//	"ipsw_checksum": "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2",
//	"ipsw_url": "server.org/.../UniversalMac_13.4_22F66_Restore.ipsw"
//
// ```
//
// ```json
//
//	"ipsw_checksum": "file:ubuntu.org/..../UniversalMac_13.4_22F66_Restore.ipsw.sum",
//	"ipsw_url": "ubuntu.org/.../UniversalMac_13.4_22F66_Restore.ipsw"
//
// ```
//
// ```json
//
//	"ipsw_checksum": "file://./shasums.txt",
//	"ipsw_url": "ubuntu.org/.../UniversalMac_13.4_22F66_Restore.ipsw"
//
// ```
//
// ```json
//
//	"ipsw_checksum": "file:./shasums.txt",
//	"ipsw_url": "ubuntu.org/.../UniversalMac_13.4_22F66_Restore.ipsw"
//
// ```
//
// In HCL2:
//
// ```hcl
//
//	ipsw_checksum = "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2"
//	ipsw_url = "ubuntu.org/.../UniversalMac_13.4_22F66_Restore.ipsw"
//
// ```
//
// ```hcl
//
//	ipsw_checksum = "file:ubuntu.org/..../UniversalMac_13.4_22F66_Restore.ipsw.sum"
//	ipsw_url = "ubuntu.org/.../UniversalMac_13.4_22F66_Restore.ipsw"
//
// ```
//
// ```hcl
//
//	ipsw_checksum = "file://./shasums.txt"
//	ipsw_url = "ubuntu.org/.../UniversalMac_13.4_22F66_Restore.ipsw"
//
// ```
//
// ```hcl
//
//	ipsw_checksum = "file:./shasums.txt",
//	ipsw_url = "ubuntu.org/.../UniversalMac_13.4_22F66_Restore.ipsw"
//
// ```

type IPSWConfig struct {
	// The checksum for the IPSW file or virtual hard drive file. The type of
	// the checksum is specified within the checksum field as a prefix, ex:
	// "md5:{$checksum}". The type of the checksum can also be omitted and
	// Packer will try to infer it based on string length. Valid values are
	// "none", "{$checksum}", "md5:{$checksum}", "sha1:{$checksum}",
	// "sha256:{$checksum}", "sha512:{$checksum}" or "file:{$path}". Here is a
	// list of valid checksum values:
	//  * md5:090992ba9fd140077b0661cb75f7ce13
	//  * 090992ba9fd140077b0661cb75f7ce13
	//  * sha1:ebfb681885ddf1234c18094a45bbeafd91467911
	//  * ebfb681885ddf1234c18094a45bbeafd91467911
	//  * sha256:ed363350696a726b7932db864dda019bd2017365c9e299627830f06954643f93
	//  * ed363350696a726b7932db864dda019bd2017365c9e299627830f06954643f93
	//  * file:http://releases.ubuntu.com/20.04/SHA256SUMS
	//  * file:file://./local/path/file.sum
	//  * file:./local/path/file.sum
	//  * none
	// Although the checksum will not be verified when it is set to "none",
	// this is not recommended since these files can be very large and
	// corruption does happen from time to time.
	IPSWChecksum string `mapstructure:"ipsw_checksum" required:"true"`
	// A URL to the IPSW containing the installation image or virtual hard drive
	// (VHD or VHDX) file to clone.
	RawSingleIPSWUrl string `mapstructure:"ipsw_url" required:"true"`
	// Multiple URLs for the IPSW to download. Packer will try these in order.
	// If anything goes wrong attempting to download or while downloading a
	// single URL, it will move on to the next. All URLs must point to the same
	// file (same checksum). By default this is empty and `ipsw_url` is used.
	// Only one of `ipsw_url` or `ipsw_urls` can be specified.
	IPSWUrls []string `mapstructure:"ipsw_urls"`
	// The path where the ipsw should be saved after download. By default will
	// go in the packer cache, with a hash of the original filename and
	// checksum as its name.
	TargetPath string `mapstructure:"ipsw_target_path"`
}

func (c *IPSWConfig) Prepare(ctx *interpolate.Context) (warnings []string, errs []error) {
	config := new(commonsteps.ISOConfig)

	config.ISOChecksum = c.IPSWChecksum
	config.RawSingleISOUrl = c.RawSingleIPSWUrl
	config.ISOUrls = c.IPSWUrls
	config.TargetPath = c.TargetPath
	//config.TargetExtension = c.TargetExtension : Target extension should be ipsw for macvm to install

	ipswWarnings, ipswErrs := config.Prepare(ctx)

	c.IPSWChecksum = config.ISOChecksum
	c.RawSingleIPSWUrl = config.RawSingleISOUrl
	c.IPSWUrls = config.ISOUrls
	c.TargetPath = config.TargetPath

	return ipswWarnings, ipswErrs
}
