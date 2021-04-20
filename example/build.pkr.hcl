packer {
  required_plugins {
    parallels = {
      version = ">= 0.0.1"
      source  = "github.com/hashicorp/parallels"
    }
  }
}

build {
  sources = ["source.parallels-iso.example"]
}
