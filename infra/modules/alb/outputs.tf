output "alb_dns_name" {
value = aws_lb.main.dns_name
}

output "alb_zone_id" {
value = aws_lb.main.zone_id
}

output "alb_security_group_id" {
value = aws_security_group.alb_sg.id
}

output "backend_target_group_arn" {
value = aws_lb_target_group.backend.arn
}

output "frontend_target_group_arn" {
value = aws_lb_target_group.frontend.arn
}
