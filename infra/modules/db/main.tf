variable "vpc_id" {}
variable "private_subnets" {}
variable "db_subnet_group_name" {}
variable "ecs_security_group_id" {}

module "db" {
source  = "terraform-aws-modules/rds/aws"
version = "~> 6.0"

identifier = "hotel-qr-postgres"

engine               = "postgres"
engine_version       = "15"
family               = "postgres15"
major_engine_version = "15"
instance_class       = "db.t4g.micro"

allocated_storage     = 20
max_allocated_storage = 100

db_name  = "hotelqrdb"
username = "hoteladmin"
port     = 5432


password = "HotelPass2026"
manage_master_user_password = false
apply_immediately = true

multi_az               = false
db_subnet_group_name   = var.db_subnet_group_name
vpc_security_group_ids = [aws_security_group.db_sg.id]


skip_final_snapshot = true
}


resource "aws_security_group" "db_sg" {
name        = "hotel-qr-db-sg"
description = "Allow inbound traffic from ECS tasks"
vpc_id      = var.vpc_id

ingress {
description     = "PostgreSQL from ECS"
from_port       = 5432
to_port         = 5432
protocol        = "tcp"
security_groups = [var.ecs_security_group_id]
}
}


resource "aws_elasticache_subnet_group" "redis_subnet" {
name       = "hotel-qr-redis-subnet"
subnet_ids = var.private_subnets
}

resource "aws_elasticache_cluster" "redis" {
cluster_id           = "hotel-qr-redis"
engine               = "redis"
node_type            = "cache.t4g.micro"
num_cache_nodes      = 1
parameter_group_name = "default.redis7"
engine_version       = "7.1"
port                 = 6379
subnet_group_name    = aws_elasticache_subnet_group.redis_subnet.name
security_group_ids   = [aws_security_group.redis_sg.id]
}


resource "aws_security_group" "redis_sg" {
name        = "hotel-qr-redis-sg"
vpc_id      = var.vpc_id

ingress {
description     = "Redis from ECS"
from_port       = 6379
to_port         = 6379
protocol        = "tcp"
security_groups = [var.ecs_security_group_id]
}
}
