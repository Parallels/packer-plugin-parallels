# For full specification on the configuration of this file visit:
# https://github.com/hashicorp/integration-template#metadata-configuration
integration {
  name = "Parallels"
  description = "The Parallels plugin can be used with HashiCorp Packer to create custom images on Parallels."
  identifier = "packer/Parallels/parallels"
  component {
    type = "builder"
    name = "Parallels PVM"
    slug = "pvm"
  }
  component {
    type = "builder"
    name = "Parallels ISO"
    slug = "iso"
  }
}
