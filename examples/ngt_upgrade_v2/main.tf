terraform {
  required_providers {
    nutanix = {
      source  = "nutanix/nutanix"
      version = "2.0"
    }
  }
}

#defining nutanix configuration
provider "nutanix" {
  username = var.nutanix_username
  password = var.nutanix_password
  endpoint = var.nutanix_endpoint
  port     = 9440
  insecure = true
}


resource "nutanix_ngt_upgrade_v2" "example" {
  ext_id = "<VM UUID>"

  reboot_preference {
    schedule_type = "IMMEDIATE"
  }
}
