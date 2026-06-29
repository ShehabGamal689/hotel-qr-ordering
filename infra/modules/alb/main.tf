resource "aws_security_group" "alb_sg" {
name        = "hotel-qr-alb-sg"
description = "Allow HTTP and HTTPS inbound traffic"
vpc_id      = var.vpc_id

ingress {
from_port   = 80
to_port     = 80
protocol    = "tcp"
cidr_blocks = ["0.0.0.0/0"]
}

ingress {
from_port   = 443
to_port     = 443
protocol    = "tcp"
cidr_blocks = ["0.0.0.0/0"]
}

egress {
from_port   = 0
to_port     = 0
protocol    = "-1"
cidr_blocks = ["0.0.0.0/0"]
}
}


resource "aws_lb" "main" {
name               = "hotel-qr-alb"
internal           = false
load_balancer_type = "application"
security_groups    = [aws_security_group.alb_sg.id]
subnets            = var.public_subnets
}


resource "aws_lb_listener" "http" {
load_balancer_arn = aws_lb.main.arn
port              = 80
protocol          = "HTTP"

default_action {
type = "redirect"
redirect {
port        = "443"
protocol    = "HTTPS"
status_code = "HTTP_301"
}
}
}


resource "aws_lb_target_group" "frontend" {
name        = "hotel-qr-frontend-tg"
port        = 80
protocol    = "HTTP"
vpc_id      = var.vpc_id
target_type = "ip" 

health_check {
path                = "/"
healthy_threshold   = 2
unhealthy_threshold = 10
interval            = 30
timeout             = 5
matcher             = "200-399"
}
}


resource "aws_lb_target_group" "backend" {
name        = "hotel-qr-backend-tg"
port        = 8080
protocol    = "HTTP"
vpc_id      = var.vpc_id
target_type = "ip" # Required for Fargate

health_check {
path                = "/health" # Ensure your Go app has a /health endpoint!
healthy_threshold   = 2
unhealthy_threshold = 10
interval            = 30
timeout             = 5
}
}

resource "aws_lb_listener" "https" {
load_balancer_arn = aws_lb.main.arn
port              = 443
protocol          = "HTTPS"
certificate_arn   = var.certificate_arn


default_action {
type             = "forward"
target_group_arn = aws_lb_target_group.frontend.arn
}
}

resource "aws_lb_listener_rule" "backend_api" {
listener_arn = aws_lb_listener.https.arn
priority     = 100

action {
type             = "forward"
target_group_arn = aws_lb_target_group.backend.arn
}

condition {
path_pattern {
values = ["/api/*"] # Any URL starting with /api/ goes to the Go Backend
}
}
}
