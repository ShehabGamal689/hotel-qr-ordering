
resource "aws_ecs_task_definition" "backend" {
family                   = "hotel-qr-backend"
requires_compatibilities = ["FARGATE"]
network_mode             = "awsvpc"
cpu                      = 256
memory                   = 512
execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn 
task_role_arn            = aws_iam_role.backend_task_role.arn

container_definitions = jsonencode([
{
name      = "backend"
image     = "${aws_ecr_repository.backend.repository_url}:latest"
cpu       = 256
memory    = 512
essential = true
portMappings = [
{
containerPort = 8080
hostPort      = 8080
protocol      = "tcp"
}
]
environment = [
{ name = "DB_HOST", value = split(":", var.database_endpoint)[0] },
{ name = "DB_PORT", value = "5432" },
{ name = "DB_USER", value = "hoteladmin" },
{ name = "DB_NAME", value = "hotelqrdb" },
{ name = "DB_PASSWORD", value = var.database_password },
{ name = "REDIS_HOST", value = var.redis_endpoint },
{ name = "REDIS_PORT", value = "6379" } ,
{ name = "STORAGE_BUCKET_NAME", value = var.s3_bucket_name },
{ name = "STORAGE_REGION", value = "us-east-1" } ,
{ name = "ORIGIN_WHITELIST", value = "https://devopsnawy.qzz.io" } , 
{ name = "GUEST_PORTAL_BASE_URL", value = "https://devopsnawy.qzz.io" }
]
logConfiguration = {
logDriver = "awslogs"
options = {
"awslogs-group"         = aws_cloudwatch_log_group.backend_logs.name
"awslogs-region"        = "us-east-1"
"awslogs-stream-prefix" = "ecs"
}
}
}
])
}


resource "aws_ecs_task_definition" "frontend" {
family                   = "hotel-qr-frontend"
requires_compatibilities = ["FARGATE"]
network_mode             = "awsvpc"
cpu                      = 256
memory                   = 512
execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn

container_definitions = jsonencode([
{
name      = "frontend"
image     = "${aws_ecr_repository.frontend.repository_url}:latest"
cpu       = 256
memory    = 512
essential = true
portMappings = [
{
containerPort = 80
hostPort      = 80
protocol      = "tcp"
}
]

environment = [
        { name = "PORT", value = "80" } ,
        { name = "NEXT_PUBLIC_API_URL", value = "https://devopsnawy.qzz.io/api/v1" }
      ]

logConfiguration = {
logDriver = "awslogs"
options = {
"awslogs-group"         = aws_cloudwatch_log_group.frontend_logs.name
"awslogs-region"        = "us-east-1"
"awslogs-stream-prefix" = "ecs"
}
}
}
])
}


resource "aws_ecs_service" "backend" {
name            = "hotel-qr-backend-svc"
cluster         = module.ecs_cluster.cluster_id
task_definition = aws_ecs_task_definition.backend.arn
desired_count   = 1
launch_type     = "FARGATE"

network_configuration {
subnets          = var.private_subnets
security_groups  = [aws_security_group.ecs_tasks_sg.id]
assign_public_ip = false
}

load_balancer {
target_group_arn = var.backend_target_group_arn
container_name   = "backend"
container_port   = 8080
}
}


resource "aws_ecs_service" "frontend" {
name            = "hotel-qr-frontend-svc"
cluster         = module.ecs_cluster.cluster_id
task_definition = aws_ecs_task_definition.frontend.arn
desired_count   = 1
launch_type     = "FARGATE"

network_configuration {
subnets          = var.private_subnets
security_groups  = [aws_security_group.ecs_tasks_sg.id]
assign_public_ip = false
}

load_balancer {
target_group_arn = var.frontend_target_group_arn
container_name   = "frontend"
container_port   = 80
}
}
