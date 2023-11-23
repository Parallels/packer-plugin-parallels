
The Parallels Packer Plugin is able to create [Parallels Desktop for
Mac](https://www.parallels.com/products/desktop/) virtual machines and export
them in the PVM format.

### Installation

To install this plugin, copy and paste this code into your Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    parallels = {
      version = ">= 1.1.5"
      source  = "github.com/parallels/parallels"
    }
  }
}
```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/parallels/parallels
```

### Requirements for Parallels Desktop v18 or Below

In addition to [Parallels Desktop for
Mac](https://www.parallels.com/products/desktop/) this requires the [Parallels
Virtualization SDK](https://www.parallels.com/downloads/desktop/) only if you are using Parallels Desktop for Mac v18 or below.

The SDK can be installed by downloading and following the instructions in the dmg.

Parallels Desktop for Mac 9 and later is supported, from PD 11 Pro or Business edition is required.


### Components

#### Builders

The plugin comes with multiple builders able to create Parallels machines,
depending on the strategy you want to use to build the image. Packer supports
the following Parallels builders:

- [parallels-iso](/packer/integrations/parallels/latest/components/builder/iso) - Starts from an ISO
  file, creates a brand new Parallels VM, installs an OS, provisions software
  within the OS, then exports that machine to create an image. This is best
  for people who want to start from scratch.

- [parallels-pvm](/packer/integrations/parallels/latest/components/builder/pvm) - This builder imports
  an existing PVM file, runs provisioners on top of that VM, and exports that
  machine to create an image. This is best if you have an existing Parallels
  VM export you want to use as the source. As an additional benefit, you can
  feed the artifact of this builder back into itself to iterate on a machine.

- [parallels-ipsw](/packer/integrations/parallels/latest/components/builder/ipsw) - Starts from an IPSW
  file, creates a brand new Parallels Mac OS VM, installs an OS, provisions software
  within the OS, then exports that machine to create an image. This is best
  for people who want to start from scratch.

- [parallels-macvm](/packer/integrations/parallels/latest/components/builder/macvm) - This builder imports
  an existing Mac VM file, runs provisioners on top of that VM, and exports that
  machine to create an image. This is best if you have an existing Parallels
  Mac VM export you want to use as the source. As an additional benefit, you can
  feed the artifact of this builder back into itself to iterate on a machine.

