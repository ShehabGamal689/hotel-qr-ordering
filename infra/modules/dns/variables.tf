variable "domain_name" {
description = "The root domain name for the application"
type        = string
}

variable "registrar_api_key" {
description = "API key for the domain registrar"
type        = string
sensitive   = true
}

variable "cluster_name" {
description = "Name of the cluster to tag the ACM certificate"
type        = string
}

variable "alb_dns_name" {
description = "DNS name of the ALB to alias to"
type        = string
}

variable "alb_zone_id" {
description = "Zone ID of the ALB to alias to"
type        = string
}
