resource "aws_ecr_repository" "backend" {
name                 = "hotel-qr-backend"
image_tag_mutability = "MUTABLE"
force_delete         = true

image_scanning_configuration {
scan_on_push = true
}
}

resource "aws_ecr_repository" "frontend" {
name                 = "hotel-qr-frontend"
image_tag_mutability = "MUTABLE"
force_delete         = true

image_scanning_configuration {
scan_on_push = true
}
}


module "ecs_cluster" {
source  = "terraform-aws-modules/ecs/aws"
version = "~> 5.0"

cluster_name = "hotel-qr-cluster"

cluster_settings = {
name  = "containerInsights"
value = "enabled"
}

fargate_capacity_providers = {
FARGATE = {
default_capacity_provider_strategy = {
weight = 100
}
}
}
}



resource "aws_security_group" "ecs_tasks_sg" {
name        = "hotel-qr-ecs-tasks-sg"
description = "Security group for ECS tasks"
vpc_id      = var.vpc_id


ingress {
from_port       = 0
to_port         = 0
protocol        = "-1"
security_groups = [var.alb_security_group_id]
}


egress {
from_port   = 0
to_port     = 0
protocol    = "-1"
cidr_blocks = ["0.0.0.0/0"]
}
}



resource "aws_iam_role" "ecs_task_execution_role" {
name = "hotel-qr-ecs-execution-role"

assume_role_policy = jsonencode({
Version = "2012-10-17"
Statement = [
{
Action = "sts:AssumeRole"
Effect = "Allow"
Principal = {
Service = "ecs-tasks.amazonaws.com"
}
}
]
})
}


resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
role       = aws_iam_role.ecs_task_execution_role.name
policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}



resource "aws_iam_role" "backend_task_role" {
  name = "hotel-qr-backend-task-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })
}

data "aws_iam_policy_document" "backend_s3" {
  statement {
    sid       = "AllowS3Uploads"
    effect    = "Allow"
    actions   = [
      "s3:PutObject",
      "s3:GetObject"
    ]
    resources = [
      "${var.s3_bucket_arn}/*"
    ]
  }
}

resource "aws_iam_role_policy" "backend_s3_policy" {
  name   = "backend-s3-upload-policy"
  role   = aws_iam_role.backend_task_role.id
  policy = data.aws_iam_policy_document.backend_s3.json
}


resource "aws_cloudwatch_log_group" "backend_logs" {
name              = "/ecs/hotel-qr-backend"
retention_in_days = 7
}

resource "aws_cloudwatch_log_group" "frontend_logs" {
name              = "/ecs/hotel-qr-frontend"
retention_in_days = 7
}
