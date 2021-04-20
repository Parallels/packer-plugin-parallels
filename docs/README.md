# Parallels Plugin

## Components

The Parallels Packer Plugin is able to create [Parallels Desktop for
Mac](https://www.parallels.com/products/desktop/) virtual machines and export
them in the PVM format.

The plugin comes with multiple builders able to create Parallels machines,
depending on the strategy you want to use to build the image. Packer supports
the following Parallels builders:

- [parallels-iso](/docs/builders/parallels-iso.mdx) - Starts from an ISO
  file, creates a brand new Parallels VM, installs an OS, provisions software
  within the OS, then exports that machine to create an image. This is best
  for people who want to start from scratch.

- [parallels-pvm](/docs/builders/parallels-pvm.mdx) - This builder imports
  an existing PVM file, runs provisioners on top of that VM, and exports that
  machine to create an image. This is best if you have an existing Parallels
  VM export you want to use as the source. As an additional benefit, you can
  feed the artifact of this builder back into itself to iterate on a machine.

## Requirements

In addition to [Parallels Desktop for
Mac](https://www.parallels.com/products/desktop/) this requires the [Parallels
Virtualization SDK](https://www.parallels.com/downloads/desktop/).

The SDK can be installed by downloading and following the instructions in the
dmg.

Parallels Desktop for Mac 9 and later is supported, from PD 11 Pro or Business
edition is required.

## Installation

### Using pre-built releases

#### Using the `packer init` command

Starting from version 1.7, Packer supports a new `packer init` command allowing
automatic installation of Packer plugins. Read the
[Packer documentation](https://www.packer.io/docs/commands/init) for more information.

To install this plugin, copy and paste this code into your Packer configuration .
Then, run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    parallels = {
      version = ">= 0.0.1"
      source  = "github.com/hashicorp/parallels"
    }
  }
}
```

#### Manual installation

You can find pre-built binary releases of the plugin [here](https://github.com/hashicorp/packer-plugin-name/releases).
Once you have downloaded the latest archive corresponding to your target OS,
uncompress it to retrieve the plugin binary file corresponding to your platform.
To install the plugin, please follow the Packer documentation on
[installing a plugin](https://www.packer.io/docs/extending/plugins/#installing-plugins).


#### From Source

If you prefer to build the plugin from its source code, clone the GitHub
repository locally and run the command `go build` from the root
directory. Upon successful compilation, a `packer-plugin-parallels` plugin
binary file can be found in the root directory.
To install the compiled plugin, please follow the official Packer documentation
on [installing a plugin](https://www.packer.io/docs/extending/plugins/#installing-plugins).