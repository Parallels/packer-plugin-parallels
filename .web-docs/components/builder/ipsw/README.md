Type: `parallels-ipsw`
Artifact BuilderId: `packer.parallels`

The Parallels Packer builder is able to create [Parallels Desktop for
Mac](https://www.parallels.com/products/desktop/) virtual machines and export
them in the MACVM format, starting from an IPSW image. This IPSW builder is
applicable for macOS VMs in Apple Silicon Chip systems only.

The builder builds a virtual machine by creating a new virtual machine from
scratch, booting it, installing an OS, provisioning software within the OS, then
shutting it down. The result of the Parallels builder is a directory containing
all the files necessary to run the virtual machine portably.

## Basic Example

Here is a basic example. This example is not functional. It will start the OS
installer and then select the country. It is up to you to add the rest of the
boot commands to actually install the OS. You can refer to [packer examples](https://github.com/Parallels/packer-examples/tree/main/macos).

```hcl
source "parallels-ipsw" "macvm_automated" {
  boot_command         = ["<wait><enter><wait2s><enter><wait20s>", # Wait for boot
                          "<leftShiftOn><tab><leftShiftOff><spacebar><wait5s>", # Select country
                          ]
  boot_wait            = "4m"
  shutdown_command     = "sudo shutdown -h now"
  ipsw_url             = "https://updates.cdn-apple.com/2023SpringFCS/fullrestores/042-01877/2F49A9FE-7033-41D0-9D0C-64EFCE6B4C22/UniversalMac_13.4.1_22F82_Restore.ipsw"
  ipsw_checksum        = "md5:acd17423a6de261121454513f0a2b814"
  ssh_password         = "parallels"
  ssh_username         = "parallels"
  vm_name              = "macOS"
  cpus                 = "4"
  memory               = "8192"
  disk_size            = "50000"
}

build {
  sources = ["source.parallels-ipsw.macvm_automated"]

  provisioner "shell" {
    inline = ["echo 'Running provisioner script'", "# Additional commands here"]
  }

}
```

It is important to add a `shutdown_command`. By default Packer halts the virtual
machine and the file system may not be sync'd. Thus, changes made in a
provisioner might not be saved.

## Configuration Reference

There are many configuration options available for the Parallels builder. They
are organized below into two categories: required and optional. Within each
category, the available options are alphabetized and described.

In addition to the options listed here, a
[communicator](/packer/docs/templates/legacy_json_templates/communicator) can be configured for this
builder. Setting communicator to "none" disables the communicator. The default
communicator is "ssh".

## IPSW Configuration Reference

By default, Packer will symlink, download or copy image files to the Packer
cache into a "`hash($ipsw_url+$ipsw_checksum).ipsw`" file.
Packer uses [hashicorp/go-getter](https://github.com/hashicorp/go-getter) in
file mode in order to perform a download.

go-getter supports the following protocols:

* Local files
* Git
* Mercurial
* HTTP
* Amazon S3

Examples:
go-getter can guess the checksum type based on `ipsw_checksum` length, and it is
also possible to specify the checksum type.

In JSON:

```json

	"ipsw_checksum": "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2",
	"ipsw_url": "apple.com/.../UniversalMac_13.5_22G74_Restore.ipsw"

```

```json

	"ipsw_checksum": "file:apple.com/.../UniversalMac_13.5_22G74_Restore.ipsw.sum",
	"ipsw_url": "apple.com/.../UniversalMac_13.5_22G74_Restore.ipsw"

```

```json

	"ipsw_checksum": "file://./shasums.txt",
	"ipsw_url": "apple.com/.../UniversalMac_13.5_22G74_Restore.ipsw"

```

```json

	"ipsw_checksum": "file:./shasums.txt",
	"ipsw_url": "apple.com/.../UniversalMac_13.5_22G74_Restore.ipsw"

```

In HCL2:

```hcl

	ipsw_checksum = "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2"
	ipsw_url = "apple.com/.../UniversalMac_13.5_22G74_Restore.ipsw"

```

```hcl

	ipsw_checksum = "file:apple.com/.../UniversalMac_13.5_22G74_Restore.ipsw.sum"
	ipsw_url = "apple.com/.../UniversalMac_13.5_22G74_Restore.ipsw"

```

```hcl

	ipsw_checksum = "file://./shasums.txt"
	ipsw_url = "apple.com/.../UniversalMac_13.5_22G74_Restore.ipsw"

```

```hcl

	ipsw_checksum = "file:./shasums.txt",
	ipsw_url = "apple.com/.../UniversalMac_13.5_22G74_Restore.ipsw"

```

### Required:

<!-- Code generated from the comments of the IPSWConfig struct in builder/parallels/ipsw/ipsw_config.go; DO NOT EDIT MANUALLY -->

- `ipsw_checksum` (string) - The checksum for the IPSW file or virtual hard drive file. The type of
  the checksum is specified within the checksum field as a prefix, ex:
  "md5:{$checksum}". The type of the checksum can also be omitted and
  Packer will try to infer it based on string length. Valid values are
  "none", "{$checksum}", "md5:{$checksum}", "sha1:{$checksum}",
  "sha256:{$checksum}", "sha512:{$checksum}" or "file:{$path}". Here is a
  list of valid checksum values:
   * md5:090992ba9fd140077b0661cb75f7ce13
   * 090992ba9fd140077b0661cb75f7ce13
   * sha1:ebfb681885ddf1234c18094a45bbeafd91467911
   * ebfb681885ddf1234c18094a45bbeafd91467911
   * sha256:ed363350696a726b7932db864dda019bd2017365c9e299627830f06954643f93
   * ed363350696a726b7932db864dda019bd2017365c9e299627830f06954643f93
   * file:http://releases.ubuntu.com/20.04/SHA256SUMS
   * file:file://./local/path/file.sum
   * file:./local/path/file.sum
   * none
  Although the checksum will not be verified when it is set to "none",
  this is not recommended since these files can be very large and
  corruption does happen from time to time.

- `ipsw_url` (string) - A URL to the IPSW containing the installation image or virtual hard drive
  (VHD or VHDX) file to clone.

<!-- End of code generated from the comments of the IPSWConfig struct in builder/parallels/ipsw/ipsw_config.go; -->


### Optional:

<!-- Code generated from the comments of the IPSWConfig struct in builder/parallels/ipsw/ipsw_config.go; DO NOT EDIT MANUALLY -->

- `ipsw_urls` ([]string) - Multiple URLs for the IPSW to download. Packer will try these in order.
  If anything goes wrong attempting to download or while downloading a
  single URL, it will move on to the next. All URLs must point to the same
  file (same checksum). By default this is empty and `ipsw_url` is used.
  Only one of `ipsw_url` or `ipsw_urls` can be specified.

- `ipsw_target_path` (string) - The path where the ipsw should be saved after download. By default will
  go in the packer cache, with a hash of the original filename and
  checksum as its name.

<!-- End of code generated from the comments of the IPSWConfig struct in builder/parallels/ipsw/ipsw_config.go; -->


- `boot_command` (array of strings) - This is an array of commands to type
  when the virtual machine is first booted. The goal of these commands should
  be to type just enough to initialize the operating system installer. Special
  keys can be typed as well, and are covered in the section below on the
  boot command. If this is not specified, it is assumed the installer will
  start itself.

- `boot_wait` (string) - The time to wait after booting the initial virtual
  machine before typing the `boot_command`. The value of this should be
  a duration. Examples are "5s" and "1m30s" which will cause Packer to wait
  five seconds and one minute 30 seconds, respectively. If this isn't
  specified, the default is 10 seconds.

- `cpus` (number) - The number of cpus to use for building the VM.
  Defaults to `1`.

- `disk_size` (number) - The size, in megabytes, of the hard disk to create
  for the VM. By default, this is 40000 (about 40 GB).

- `host_interfaces` (array of strings) - A list of which interfaces on the
  host should be searched for a IP address. The first IP address found on one
  of these will be used as `{{ .HTTPIP }}` in the `boot_command`. Defaults to
  \["en0", "en1", "en2", "en3", "en4", "en5", "en6", "en7", "en8", "en9",
  "en10", "en11", "en12", "en13", "en14", "en15", "en16", "en17", "en18", "en19", "en20",
  "ppp0", "ppp1", "ppp2"\].

- `memory` (number) - The amount of memory to use for building the VM in
  megabytes. Defaults to `512` megabytes.

- `output_directory` (string) - This is the path to the directory where the
  resulting virtual machine will be created. This may be relative or absolute.
  If relative, the path is relative to the working directory when `packer`
  is executed. This directory must not exist or be empty prior to running
  the builder. By default this is "output-BUILDNAME" where "BUILDNAME" is the
  name of the build.

- `prlctl` (array of array of strings) - Custom `prlctl` commands to execute
  in order to further customize the virtual machine being created. The value
  of this is an array of commands to execute. The commands are executed in the
  order defined in the template. For each command, the command is defined
  itself as an array of strings, where each string represents a single
  argument on the command-line to `prlctl` (but excluding `prlctl` itself).
  Each arg is treated as a [configuration
  template](/packer/docs/templates/legacy_json_templates/engine), where the `Name`
  variable is replaced with the VM name. More details on how to use `prlctl`
  are below.

- `prlctl_post` (array of array of strings) - Identical to `prlctl`, except
  that it is run after the virtual machine is shutdown, and before the virtual
  machine is exported.

- `prlctl_version_file` (string) - The path within the virtual machine to
  upload a file that contains the `prlctl` version that was used to create
  the machine. This information can be useful for provisioning. By default
  this is ".prlctl_version", which will generally upload it into the
  home directory.

- `shutdown_command` (string) - The command to use to gracefully shut down the
  machine once all the provisioning is done. By default this is an empty
  string, which tells Packer to just forcefully shut down the machine.

- `sound` (boolean) - Specifies whether to enable the sound device when
  building the VM. Defaults to `false`.

- `shutdown_timeout` (string) - The amount of time to wait after executing the
  `shutdown_command` for the virtual machine to actually shut down. If it
  doesn't shut down in this time, it is an error. By default, the timeout is
  "5m", or five minutes.

- `usb` (boolean) - Specifies whether to enable the USB bus when building
  the VM. Defaults to `false`.

- `vm_name` (string) - This is the name of the PVM directory for the new
  virtual machine, without the file extension. By default this is
  "packer-BUILDNAME", where "BUILDNAME" is the name of the build.

## Http directory configuration reference

<!-- Code generated from the comments of the HTTPConfig struct in multistep/commonsteps/http_config.go; DO NOT EDIT MANUALLY -->

Packer will create an http server serving `http_directory` when it is set, a
random free port will be selected and the architecture of the directory
referenced will be available in your builder.

Example usage from a builder:

```
wget http://{{ .HTTPIP }}:{{ .HTTPPort }}/foo/bar/preseed.cfg
```

<!-- End of code generated from the comments of the HTTPConfig struct in multistep/commonsteps/http_config.go; -->


### Optional:

<!-- Code generated from the comments of the HTTPConfig struct in multistep/commonsteps/http_config.go; DO NOT EDIT MANUALLY -->

- `http_directory` (string) - Path to a directory to serve using an HTTP server. The files in this
  directory will be available over HTTP that will be requestable from the
  virtual machine. This is useful for hosting kickstart files and so on.
  By default this is an empty string, which means no HTTP server will be
  started. The address and port of the HTTP server will be available as
  variables in `boot_command`. This is covered in more detail below.

- `http_content` (map[string]string) - Key/Values to serve using an HTTP server. `http_content` works like and
  conflicts with `http_directory`. The keys represent the paths and the
  values contents, the keys must start with a slash, ex: `/path/to/file`.
  `http_content` is useful for hosting kickstart files and so on. By
  default this is empty, which means no HTTP server will be started. The
  address and port of the HTTP server will be available as variables in
  `boot_command`. This is covered in more detail below.
  Example:
  ```hcl
    http_content = {
      "/a/b"     = file("http/b")
      "/foo/bar" = templatefile("${path.root}/preseed.cfg", { packages = ["nginx"] })
    }
  ```

- `http_port_min` (int) - These are the minimum and maximum port to use for the HTTP server
  started to serve the `http_directory`. Because Packer often runs in
  parallel, Packer will choose a randomly available port in this range to
  run the HTTP server. If you want to force the HTTP server to be on one
  port, make this minimum and maximum port the same. By default the values
  are `8000` and `9000`, respectively.

- `http_port_max` (int) - HTTP Port Max

- `http_bind_address` (string) - This is the bind address for the HTTP server. Defaults to 0.0.0.0 so that
  it will work with any network interface.

- `http_network_protocol` (string) - Defines the HTTP Network protocol. Valid options are `tcp`, `tcp4`, `tcp6`,
  `unix`, and `unixpacket`. This value defaults to `tcp`.

<!-- End of code generated from the comments of the HTTPConfig struct in multistep/commonsteps/http_config.go; -->


## Boot Command

The `boot_command` configuration is very important: it specifies the keys to
type when the virtual machine is first booted in order to start the OS
installer. This command is typed after `boot_wait`, which gives the virtual
machine some time to actually load the restore image.

As documented above, the `boot_command` is an array of strings. The strings are
all typed in sequence. It is an array only to improve readability within the
template.

The boot command is "typed" character for character (using the Parallels
Virtualization SDK, see [Parallels Builder](/packer/plugin/builder/parallels) in Parallels Desktop version
18 or before, or using the 'prlctl send-key-event from Parallels Desktop
version 19') simulating a human actually typing the keyboard.

There are a set of special keys available. If these are in your boot
command, they will be replaced by the proper key:

- `<bs>` - Backspace

- `<del>` - Delete

- `<enter> <return>` - Simulates an actual "enter" or "return" keypress.

- `<esc>` - Simulates pressing the escape key.

- `<tab>` - Simulates pressing the tab key.

- `<f1> - <f12>` - Simulates pressing a function key.

- `<up> <down> <left> <right>` - Simulates pressing an arrow key.

- `<spacebar>` - Simulates pressing the spacebar.

- `<insert>` - Simulates pressing the insert key.

- `<home> <end>` - Simulates pressing the home and end keys.

- `<pageUp> <pageDown>` - Simulates pressing the page up and page down keys.

- `<menu>` - Simulates pressing the Menu key.

- `<leftAlt> <rightAlt>` - Simulates pressing the alt key.

- `<leftCtrl> <rightCtrl>` - Simulates pressing the ctrl key.

- `<leftShift> <rightShift>` - Simulates pressing the shift key.

- `<leftSuper> <rightSuper>` - Simulates pressing the ⌘ or Windows key.

- `<wait> <wait5> <wait10>` - Adds a 1, 5 or 10 second pause before
  sending any additional keys. This is useful if you have to generally wait
  for the UI to update before typing more.

- `<waitXX>` - Add an arbitrary pause before sending any additional keys. The
  format of `XX` is a sequence of positive decimal numbers, each with
  optional fraction and a unit suffix, such as `300ms`, `1.5h` or `2h45m`.
  Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. For example
  `<wait10m>` or `<wait1m20s>`

### On/Off variants

Any printable keyboard character, and of these "special" expressions, with the
exception of the `<wait>` types, can also be toggled on or off. For example, to
simulate ctrl+c, use `<leftCtrlOn>c<leftCtrlOff>`. Be sure to release them,
otherwise they will be held down until the machine reboots.

To hold the `c` key down, you would use `<cOn>`. Likewise, `<cOff>` to release.

### Templates inside boot command

In addition to the special keys, each command to type is treated as a
[template engine](/packer/docs/templates/legacy_json_templates/engine). The
available variables are:

- `HTTPIP` and `HTTPPort` - The IP and port, respectively of an HTTP server
  that is started serving the directory specified by the `http_directory`
  configuration parameter or the content specified in the `http_content` map. If
  `http_directory` or `http_content` isn't specified, these will be blank!
- `Name` - The name of the VM.

For more examples of various boot commands, see the sample projects from our
[community templates page](/community-tools#templates).

For more examples of various boot commands, see the sample projects from our
[community templates page](/community-tools#templates).

## prlctl Commands

In order to perform extra customization of the virtual machine, a template can
define extra calls to `prlctl` to perform.
[prlctl](http://download.parallels.com/desktop/v9/ga/docs/en_US/Parallels%20Command%20Line%20Reference%20Guide.pdf)
is the command-line interface to Parallels Desktop. It can be used to configure
the advanced virtual machine options.

Extra `prlctl` commands are defined in the template in the `prlctl` section. In the
example below `prlctl` is used to explicitly enable the adaptive hypervisor, and
disable 3d acceleration:

```json
{
  "prlctl": [
    ["set", "{{.Name}}", "--3d-accelerate", "off"],
    ["set", "{{.Name}}", "--adaptive-hypervisor", "on"]
  ]
}
```

The value of `prlctl` is an array of commands to execute. These commands are
executed in the order defined. So in the above example, 3d acceleration will be disabled
first, followed by the command which enables the adaptive hypervisor.

Each command itself is an array of strings, where each string is an argument to
`prlctl`. Each argument is treated as a [template engine](/packer/docs/templates/legacy_json_templates/engine). The only available
variable is `Name` which is replaced with the unique name of the VM, which is
required for many `prlctl` calls.

## VM Configuration

<!-- Code generated from the comments of the VMConfig struct in builder/parallels/common/vm_config.go; DO NOT EDIT MANUALLY -->

VMConfig contains various configuration options for the VM.

<!-- End of code generated from the comments of the VMConfig struct in builder/parallels/common/vm_config.go; -->


### Optional:

<!-- Code generated from the comments of the VMConfig struct in builder/parallels/common/vm_config.go; DO NOT EDIT MANUALLY -->

- `startup_view` (string) - StartupView specifies the view to be shown when the VM starts.
  Possible values are: same, window, coherence, fullscreen, modality, headless.
  MacOS VMs in Apple Silicon Chip Macs do not support coherence, fullscreen and modality.

- `on_window_close` (string) - OnWindowClose specifies the action to be taken on VM when the VM window is closed.
  Possible values are: suspend, shutdown, stop, ask, keep-running.

<!-- End of code generated from the comments of the VMConfig struct in builder/parallels/common/vm_config.go; -->


## BootScreen Configuration

### Optional:

<!-- Code generated from the comments of the BootScreenConfig struct in builder/parallels/common/boot_screen_config.go; DO NOT EDIT MANUALLY -->

- `screen_name` (string) - Screen name to identify

- `matching_strings` ([]string) - Strings present in the screen to identify this screen

- `is_last_screen` (bool) - Specifies if the current screen is the last screen
  Screen based boot will stop after this screen

- `execute_only_once` (bool) - If true, the screen will be deleted after first execution
  Default value is false

<!-- End of code generated from the comments of the BootScreenConfig struct in builder/parallels/common/boot_screen_config.go; -->
