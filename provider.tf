variable "access_key" {}

variable "secret_key" {}

provider "aws" {
  access_key  = "${var.access_key}"
  secret_key  = "${var.secret_key}"
  region      = "${var.region}"
  max_retries = 7
}

variable "region" {
  description = "AWS Region."
  default     = "us-east-1"
}
