output "zone_id" {
description = "The ID of the Route53 Hosted Zone"
value       = aws_route53_zone.main.zone_id
}

output "certificate_arn" {
description = "The ARN of the validated ACM certificate"
value       = aws_acm_certificate_validation.cert.certificate_arn
}
