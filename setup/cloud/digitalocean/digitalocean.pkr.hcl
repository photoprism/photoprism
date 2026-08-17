## Packer template for building the PhotoPrism 1-Click App image for DigitalOcean.
##
## The "digitalocean" builder ships as a plugin, so run "packer init" before the
## first build. Legacy JSON templates cannot declare required_plugins, which is
## why this template is HCL2:
##
##   export DIGITALOCEAN_TOKEN="<api token>"
##   packer init digitalocean.pkr.hcl
##   packer build digitalocean.pkr.hcl
##
## Download packer from https://developer.hashicorp.com/packer/install if needed,
## or run it from the official container image without installing it:
##
##   docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp \
##     -e DIGITALOCEAN_TOKEN -v "$PWD:/build" -w /build \
##     hashicorp/packer:1.16.0 build digitalocean.pkr.hcl

packer {
  required_plugins {
    digitalocean = {
      source  = "github.com/digitalocean/digitalocean"
      version = "~> 1.4.1"
    }
  }
}

variable "do_token" {
  type      = string
  default   = env("DIGITALOCEAN_TOKEN")
  sensitive = true
}

variable "base_image" {
  type        = string
  default     = "ubuntu-26-04-x64"
  description = "DigitalOcean base image slug. Must be a distribution the Marketplace image check accepts."
}

variable "region" {
  type    = string
  default = "fra1"
}

variable "size" {
  type        = string
  default     = "s-1vcpu-1gb"
  description = "Build Droplet size. Keep the smallest plan so the image can run on every plan."
}

variable "apt_packages" {
  type    = string
  default = "software-properties-common apt-transport-https ca-certificates openssl ufw curl"
}

variable "application_name" {
  type    = string
  default = "PhotoPrism"
}

variable "application_version" {
  type    = string
  default = "latest"
}

locals {
  timestamp  = regex_replace(timestamp(), "[- TZ:]", "")
  image_name = "photoprism-ce-marketplace-${local.timestamp}"
}

source "digitalocean" "photoprism" {
  api_token     = var.do_token
  image         = var.base_image
  region        = var.region
  size          = var.size
  ssh_username  = "root"
  snapshot_name = local.image_name
}

build {
  sources = ["source.digitalocean.photoprism"]

  provisioner "shell" {
    inline = ["cloud-init status --wait"]
  }

  ## Runs once when a user's Droplet boots for the first time.
  provisioner "file" {
    source      = "install_photoprism.sh"
    destination = "/var/lib/cloud/scripts/per-instance/install_photoprism.sh"
  }

  provisioner "shell" {
    environment_vars = [
      "DEBIAN_FRONTEND=noninteractive",
      "UA_LOG_LEVEL=info",
      "LC_ALL=C",
      "LANG=en_US.UTF-8",
      "LC_CTYPE=en_US.UTF-8",
    ]
    inline = [
      "echo 'Acquire::Retries \"10\";' > /etc/apt/apt.conf.d/80retry",
      "echo 'APT::Install-Recommends \"false\";' > /etc/apt/apt.conf.d/80recommends",
      "echo 'APT::Install-Suggests \"false\";' > /etc/apt/apt.conf.d/80suggests",
      "echo 'APT::Get::Assume-Yes \"true\";' > /etc/apt/apt.conf.d/80forceyes",
      "echo 'APT::Get::Fix-Missing \"true\";' > /etc/apt/apt.conf.d/80fixmissing",
      "echo 'force-confold' > /etc/dpkg/dpkg.cfg.d/force-confold",
      "apt-get -qq update",
      "apt-get -qq dist-upgrade",
      "apt-get -qq install ${var.apt_packages}",
      "apt-get -qq autoclean",
      "apt-get -qq autoremove",
      "chmod 700 /var/lib/cloud/scripts/per-instance/install_photoprism.sh",
    ]
  }

  ## check.sh reports the same verdict DigitalOcean requires for approval, so read
  ## its output in the build log rather than treating a successful build as a pass.
  provisioner "shell" {
    environment_vars = [
      "DEBIAN_FRONTEND=noninteractive",
      "LC_ALL=C",
      "LANG=en_US.UTF-8",
      "LC_CTYPE=en_US.UTF-8",
    ]
    scripts = [
      "init/firewall.sh",
      "init/cleanup.sh",
      "init/check.sh",
    ]
  }
}
