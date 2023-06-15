Type: `parallels-iso`
Artifact BuilderId: `packer.parallels`

The Parallels Packer builder is able to create [Parallels Desktop for
Mac](https://www.parallels.com/products/desktop/) virtual machines and export
them in the PVM format, starting from an ISO image.

The builder builds a virtual machine by creating a new virtual machine from
scratch, booting it, installing an OS, provisioning software within the OS, then
shutting it down. The result of the Parallels builder is a directory containing
all the files necessary to run the virtual machine portably.

## Basic Example

Here is a basic example. This example is not functional. It will start the OS
installer but then fail because we don't provide the preseed file for Ubuntu to
self-install. Still, the example serves to show the basic configuration:

```json
{
  "type": "parallels-iso",
  "guest_os_type": "ubuntu",
  "iso_url": "http://releases.ubuntu.com/12.04/ubuntu-12.04.3-server-amd64.iso",
  "iso_checksum": "2cbe868812a871242cdcdd8f2fd6feb9",
  "parallels_tools_flavor": "lin",
  "ssh_username": "packer",
  "ssh_password": "packer",
  "ssh_timeout": "30s",
  "shutdown_command": "echo 'packer' | sudo -S shutdown -P now"
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
builder. In addition to the options defined there, a private key file
can also be supplied to override the typical auto-generated key:

- `ssh_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with SSH.
  The `~` can be used in path and will be expanded to the home directory
  of current user.


## ISO Configuration Reference

<!-- Code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; DO NOT EDIT MANUALLY -->

By default, Packer will symlink, download or copy image files to the Packer
cache into a "`hash($iso_url+$iso_checksum).$iso_target_extension`" file.
Packer uses [hashicorp/go-getter](https://github.com/hashicorp/go-getter) in
file mode in order to perform a download.

go-getter supports the following protocols:

* Local files
* Git
* Mercurial
* HTTP
* Amazon S3

Examples:
go-getter can guess the checksum type based on `iso_checksum` length, and it is
also possible to specify the checksum type.

In JSON:

```json

	"iso_checksum": "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2",
	"iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```json

	"iso_checksum": "file:ubuntu.org/..../ubuntu-14.04.1-server-amd64.iso.sum",
	"iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```json

	"iso_checksum": "file://./shasums.txt",
	"iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```json

	"iso_checksum": "file:./shasums.txt",
	"iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

In HCL2:

```hcl

	iso_checksum = "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2"
	iso_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```hcl

	iso_checksum = "file:ubuntu.org/..../ubuntu-14.04.1-server-amd64.iso.sum"
	iso_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```hcl

	iso_checksum = "file://./shasums.txt"
	iso_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```hcl

	iso_checksum = "file:./shasums.txt",
	iso_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

<!-- End of code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; -->


### Required:

<!-- Code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; DO NOT EDIT MANUALLY -->

- `iso_checksum` (string) - The checksum for the ISO file or virtual hard drive file. The type of
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

- `iso_url` (string) - A URL to the ISO containing the installation image or virtual hard drive
  (VHD or VHDX) file to clone.

<!-- End of code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; -->


### Optional:

<!-- Code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; DO NOT EDIT MANUALLY -->

- `iso_urls` ([]string) - Multiple URLs for the ISO to download. Packer will try these in order.
  If anything goes wrong attempting to download or while downloading a
  single URL, it will move on to the next. All URLs must point to the same
  file (same checksum). By default this is empty and `iso_url` is used.
  Only one of `iso_url` or `iso_urls` can be specified.

- `iso_target_path` (string) - The path where the iso should be saved after download. By default will
  go in the packer cache, with a hash of the original filename and
  checksum as its name.

- `iso_target_extension` (string) - The extension of the iso file after download. This defaults to `iso`.

<!-- End of code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; -->


### Required:

- `parallels_tools_flavor` (string) - The flavor of the Parallels Tools ISO to
  install into the VM. Valid values are "win", "win-arm", "lin", "lin-arm", "mac", 
  "mac-arm", "os2" and "other". This can be omitted only if `parallels_tools_mode`
  is "disable".

### Optional:

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

- `disk_type` (string) - The type for image file based virtual disk drives,
  defaults to `expand`. Valid options are `expand` (expanding disk) that the
  image file is small initially and grows in size as you add data to it, and
  `plain` (plain disk) that the image file has a fixed size from the moment it
  is created (i.e the space is allocated for the full drive). Plain disks
  perform faster than expanding disks. `skip_compaction` will be set to true
  automatically for plain disks.

- `floppy_files` (array of strings) - A list of files to place onto a floppy
  disk that is attached when the VM is booted. This is most useful for
  unattended Windows installs, which look for an `Autounattend.xml` file on
  removable media. By default, no floppy will be attached. All files listed in
  this setting get placed into the root directory of the floppy and the floppy
  is attached as the first floppy device. Currently, no support exists for
  creating sub-directories on the floppy. Wildcard characters (\*, ?,
  and \[\]) are allowed. Directory names are also allowed, which will add all
  the files found in the directory to the floppy.

- `floppy_dirs` (array of strings) - A list of directories to place onto
  the floppy disk recursively. This is similar to the `floppy_files` option
  except that the directory structure is preserved. This is useful for when
  your floppy disk includes drivers or if you just want to organize it's
  contents as a hierarchy. Wildcard characters (\*, ?, and \[\]) are allowed.

- `floppy_label` (string) - The label to use for the floppy disk that
  is attached when the VM is booted. This is most useful for cloud-init,
  Kickstart or other early initialization tools, which can benefit from labelled floppy disks.
  By default, the floppy label will be 'packer'.

- `guest_os_type` (string) - The guest OS type being installed. By default
  this is "other", but you can get _dramatic_ performance improvements by
  setting this to the proper value. To view all available values for this run
  `prlctl create x --distribution list`. Setting the correct value hints to
  Parallels Desktop how to optimize the virtual hardware to work best with
  that operating system.

- `hard_drive_interface` (string) - The type of controller that the hard
  drives are attached to, defaults to "sata". Valid options are "sata", "ide",
  and "scsi".

- `host_interfaces` (array of strings) - A list of which interfaces on the
  host should be searched for a IP address. The first IP address found on one
  of these will be used as `{{ .HTTPIP }}` in the `boot_command`. Defaults to
  \["en0", "en1", "en2", "en3", "en4", "en5", "en6", "en7", "en8", "en9",
  "ppp0", "ppp1", "ppp2"\].

- `memory` (number) - The amount of memory to use for building the VM in
  megabytes. Defaults to `512` megabytes.

- `output_directory` (string) - This is the path to the directory where the
  resulting virtual machine will be created. This may be relative or absolute.
  If relative, the path is relative to the working directory when `packer`
  is executed. This directory must not exist or be empty prior to running
  the builder. By default this is "output-BUILDNAME" where "BUILDNAME" is the
  name of the build.

- `parallels_tools_guest_path` (string) - The path in the virtual machine to
  upload Parallels Tools. This only takes effect if `parallels_tools_mode`
  is "upload". This is a [configuration
  template](/packer/docs/templates/legacy_json_templates/engine) that has a single
  valid variable: `Flavor`, which will be the value of
  `parallels_tools_flavor`. By default this is `prl-tools-{{.Flavor}}.iso`
  which should upload into the login directory of the user.

- `parallels_tools_mode` (string) - The method by which Parallels Tools are
  made available to the guest for installation. Valid options are "upload",
  "attach", or "disable". If the mode is "attach" the Parallels Tools ISO will
  be attached as a CD device to the virtual machine. If the mode is "upload"
  the Parallels Tools ISO will be uploaded to the path specified by
  `parallels_tools_guest_path`. The default value is "upload".

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

- `skip_compaction` (boolean) - Virtual disk image is compacted at the end of
  the build process using `prl_disk_tool` utility (except for the case that
  `disk_type` is set to `plain`). In certain rare cases, this might corrupt
  the resulting disk image. If you find this to be the case, you can disable
  compaction using this configuration value.

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

	`wget http://{{ .HTTPIP }}:{{ .HTTPPort }}/foo/bar/preseed.cfg`

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

<!-- End of code generated from the comments of the HTTPConfig struct in multistep/commonsteps/http_config.go; -->


## Boot Command

The `boot_command` configuration is very important: it specifies the keys to
type when the virtual machine is first booted in order to start the OS
installer. This command is typed after `boot_wait`, which gives the virtual
machine some time to actually load the ISO.

As documented above, the `boot_command` is an array of strings. The strings are
all typed in sequence. It is an array only to improve readability within the
template.

The boot command is "typed" character for character (using the Parallels
Virtualization SDK, see [Parallels Builder](/packer/plugin/builder/parallels))
simulating a human actually typing the keyboard.

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


Example boot command. This is actually a working boot command used to start an
Ubuntu 12.04 installer:

```text
[
  "<esc><esc><enter><wait>",
  "/install/vmlinuz noapic ",
  "preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/preseed.cfg ",
  "debian-installer=en_US auto locale=en_US kbd-chooser/method=us ",
  "hostname={{ .Name }} ",
  "fb=false debconf/frontend=noninteractive ",
  "keyboard-configuration/modelcode=SKIP keyboard-configuration/layout=USA ",
  "keyboard-configuration/variant=USA console-setup/ask_detect=false ",
  "initrd=/install/initrd.gz -- <enter>;"
]
```

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
