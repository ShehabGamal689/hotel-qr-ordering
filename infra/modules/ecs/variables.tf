variable "vpc_id" {}
variable "public_subnets" {}
variable "private_subnets" {}
variable "alb_security_group_id" {}
variable "backend_target_group_arn" {}
variable "frontend_target_group_arn" {}
variable "database_endpoint" {}
variable "redis_endpoint" {}
variable "database_password" {
  description = "The password for the RDS database"
  type        = string
  sensitive   = true
}
variable "s3_bucket_name" {
  description = "Name of the S3 bucket for catalog images"
  type        = string
}

variable "s3_bucket_arn" {
  description = "ARN of the S3 bucket for catalog images"
  type        = string
}
