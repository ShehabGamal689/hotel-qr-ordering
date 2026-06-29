resource "aws_route53_zone" "main" {
name = var.domain_name
}


resource "null_resource" "update_registrar_ns" {
triggers = {
ns_records = join(",", aws_route53_zone.main.name_servers)
}

depends_on = [aws_route53_zone.main]

provisioner "local-exec" {
command = <<EOT
curl -X PUT "https://domain-api.digitalplat.org/api/v1/domains/${var.domain_name}/nameservers" \
-H "Authorization: Bearer ${var.registrar_api_key}" \
-H "Content-Type: application/json" \
-H "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36" \
-d '{
"nameservers": [
"${aws_route53_zone.main.name_servers[0]}",
"${aws_route53_zone.main.name_servers[1]}",
"${aws_route53_zone.main.name_servers[2]}",
"${aws_route53_zone.main.name_servers[3]}"
]
}'
EOT
}
}


resource "time_sleep" "wait_for_dns_propagation" {
depends_on      = [null_resource.update_registrar_ns]
create_duration = "60s"
}

resource "aws_acm_certificate" "cert" {
domain_name       = var.domain_name
validation_method = "DNS"


subject_alternative_names = ["*.${var.domain_name}"]

lifecycle {
create_before_destroy = true
}

tags = {
Name = "${var.cluster_name}-cert"
}
}

resource "aws_route53_record" "cert_validation" {
for_each = {
for dvo in aws_acm_certificate.cert.domain_validation_options : dvo.domain_name => {
name   = dvo.resource_record_name
record = dvo.resource_record_value
type   = dvo.resource_record_type
}
}
allow_overwrite = true
name            = each.value.name
records         = [each.value.record]
ttl             = 60
type            = each.value.type
zone_id         = aws_route53_zone.main.zone_id
}

resource "aws_acm_certificate_validation" "cert" {
depends_on = [
time_sleep.wait_for_dns_propagation,
aws_route53_record.cert_validation
]

certificate_arn         = aws_acm_certificate.cert.arn
validation_record_fqdns = [for record in aws_route53_record.cert_validation : record.fqdn]
}


resource "aws_route53_record" "app_domain" {
zone_id = aws_route53_zone.main.zone_id
name    = var.domain_name
type    = "A"

alias {
name                   = var.alb_dns_name
zone_id                = var.alb_zone_id
evaluate_target_health = true
}
}
