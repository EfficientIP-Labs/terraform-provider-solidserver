# Configure the Cloudflare provider using the required_providers stanza
# required with Terraform 0.13 and beyond. You may optionally use version
# directive to prevent breaking changes occurring unannounced.


terraform {
  required_providers {
    solidserver = {
      source = "EfficientIP-Labs/solidserver"
      version = "~> 1.1.0"
    }
  }
}

provider "solidserver" {
    username = "username"
    password = "password"
    host  = "192.168.0.1"
    sslverify = "false"
}