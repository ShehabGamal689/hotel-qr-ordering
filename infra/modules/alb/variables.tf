variable "vpc_id" {
description = "The ID of the VPC"
type        = string
}

variable "public_subnets" {
description = "List of public subnets for the ALB"
type        = list(string)
}

variable "certificate_arn" {
description = "The ARN of the validated ACM certificate"
type        = string
}
