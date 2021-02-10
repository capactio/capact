# this file is used during image build to speedup the running time
terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "3.55.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
    azurerm = {
      source = "hashicorp/azurerm"
      version = "2.46.1"
    }
  }
}
