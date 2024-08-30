variable "resource_group_name" {
  description = "The name of the resource group"
  type        = string
  default     = "toobo-rg"
}

variable "location" {
  description = "The Azure region to deploy resources into"
  type        = string
  default     = "West Europe"
}

variable "config_secret_name" {
  description = "The name of the secret in the key vault"
  type        = string
  default     = "toobo-config"
}