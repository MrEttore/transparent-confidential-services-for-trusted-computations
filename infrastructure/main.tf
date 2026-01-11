terraform {
  required_version = ">= 1.2"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 4.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Firewall rule to allow attestation traffic to the confidential VM.
module "firewall" {
  source        = "./modules/firewall"
  name          = "allow-attestation"
  network       = var.network
  target_tags   = ["allow-attestation"]
}

# Transparent Service's prod confidential VM based on the golden boot disk image "golden-reference-tee".
module "prod_tee" {
  source               = "./modules/compute"
  instance_name        = "llm-core-tee"
  boot_disk_name       = "llm-core-tee-boot-disk"
  zone                 = var.zone
  network              = var.network
  tags                 = ["allow-attestation"]
  image                = "golden-reference-tee"
  image_project        = var.golden_image_project_id
}
