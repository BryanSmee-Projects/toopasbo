terraform {
  backend "s3" {
    bucket = "toobo-terraform-state"
    key    = "toobo/terraform.tfstate"
    region = "eu-central-1"
  }
}