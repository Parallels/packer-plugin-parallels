Type: `parallels-macvm`
Artifact BuilderId: `packer.parallels`

This Parallels builder is able to create [Parallels Desktop for
Mac](https://www.parallels.com/products/desktop/) virtual machines and export
them in the MACVM format, starting from an existing MACVM (exported macOS virtual
machine image).

The builder builds a virtual machine by importing an existing MACVM file. It then
boots this image, runs provisioners on this new VM, and exports that VM to
create the image. The imported machine is deleted prior to finishing the build.

## Basic Example

Here is a basic example. This starts a MACVM and runs provisioners on it, then
exports that VM to create a new MACVM, if SSH is already enabled in the system.
To enable Remote Login (SSH) in the system and to install Parallels Tools inside
MACVM, you can use boot_command. You can refer to [packer examples](https://github.com/Parallels/packer-examples/tree/main/macos).

```hcl
source "parallels-macvm" "macvm_automated" {
  boot_command         = ["<wait>"]
  boot_wait            = "20s"
  shutdown_command     = "sudo shutdown -h now"
  source_path          = "/Users/user/Parallels/Sonoma.macvm"
  ssh_password         = "parallels"
  ssh_username         = "parallels"
  vm_name              = "macvm_automated"
}

build {
  sources = ["source.parallels-macvm.macvm_automated"]

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

### Required:

<!-- Code generated from the comments of the Config struct in builder/parallels/macvm/config.go; DO NOT EDIT MANUALLY -->

- `source_path` (string) - The path to a MACVM directory that acts as the source
  of this build.

<!-- End of code generated from the comments of the Config struct in builder/parallels/macvm/config.go; -->


### Optional:

<!-- Code generated from the comments of the Config struct in builder/parallels/macvm/config.go; DO NOT EDIT MANUALLY -->

- `boot_screen_config` (parallelscommon.BootScreensConfig) - Screens and it's boot configs
  A screen is considered matched if all the matching strings are present in the screen.
  The first matching screen will be considered & boot config of that screen will be used.
  If matching strings are empty, then it is considered as empty screen,
  empty screen has some special meaning, which will be considered when none of the other screens are matched.
  You can use this screen to make system wait for some time / execute a common boot command etc.
  The empty screen boot command will be executed repeatedly until a non-empty screen is found.
  If more than one empty screen is found, then it is considered as an error.

- `ocr_library` (string) - OCR library to use. Two options are currently supported: "tesseract" and "vision".
  "tesseract" uses the Tesseract OCR library to recognize text. A manual installation of -
  Tesseract is required for this to work.
  "vision" uses the Apple Vision library to recognize text, which is included in macOS. It might -
  cause problems in macOS 13 or older VMs.
  Defaults to "vision".

- `vm_name` (string) - This is the name of the MACVM directory for the new
  virtual machine, without the file extension. By default this is
  "packer-BUILDNAME", where "BUILDNAME" is the name of the build.

- `reassign_mac` (bool) - If this is "false" the MAC address of the first
  NIC will reused when imported else a new MAC address will be generated
  by Parallels. Defaults to "false".

<!-- End of code generated from the comments of the Config struct in builder/parallels/macvm/config.go; -->


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

- `shutdown_timeout` (string) - The amount of time to wait after executing the
  `shutdown_command` for the virtual machine to actually shut down. If it
  doesn't shut down in this time, it is an error. By default, the timeout is
  "5m", or five minutes.

## Parallels Tools

Parallels Tools iso will be mounted automatically in the macOS VM. You can
install Parallels Tools using the boot command or provisioner. You can refer
to [packer examples](https://github.com/Parallels/packer-examples/tree/main/macos).

## Boot Command

The `boot_command` specifies the keys to type when the virtual machine is first
booted. This command is typed after `boot_wait`.

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


## prlctl Commands

In order to perform extra customization of the virtual machine, a template can
define extra calls to `prlctl` to perform.
[prlctl](http://download.parallels.com/desktop/v9/ga/docs/en_US/Parallels%20Command%20Line%20Reference%20Guide.pdf)
is the command-line interface to Parallels Desktop. It can be used to configure
the virtual machine, such as set RAM, CPUs, etc.

Extra `prlctl` commands are defined in the template in the `prlctl` section. An
example is shown below that sets the memory and number of CPUs within the
virtual machine:

```json
{
  "prlctl": [
    ["set", "{{.Name}}", "--memsize", "1024"],
    ["set", "{{.Name}}", "--cpus", "2"]
  ]
}
```

The value of `prlctl` is an array of commands to execute. These commands are
executed in the order defined. So in the above example, the memory will be set
followed by the CPUs.

Each command itself is an array of strings, where each string is an argument to
`prlctl`. Each argument is treated as a [configuration
template](/packer/docs/templates/legacy_json_templates/engine). The only available
variable is `Name` which is replaced with the unique name of the VM, which is
required for many `prlctl` calls.
