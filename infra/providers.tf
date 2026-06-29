terraform {
required_version = ">= 1.0"

backend "s3" {
bucket       = "hotelqr"
key          = "ecs/terraform.tfstate"
region       = "us-east-1"
encrypt      = true
use_lockfile = true
}

required_providers {
aws = {
source  = "hashicorp/aws"
version = "~> 5.0"
}
local = {
source  = "hashicorp/local"
version = "~> 2.0"
}
}
} 

provider "aws" {
region = "us-east-1"
}
