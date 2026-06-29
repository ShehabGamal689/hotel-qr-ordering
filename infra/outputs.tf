output "ecr_backend_url" {
description = "URL for the backend ECR repository"
value       = module.ecs.backend_repo_url
}

output "ecr_frontend_url" {
description = "URL for the frontend ECR repository"
value       = module.ecs.frontend_repo_url
}

output "database_endpoint" {
description = "Endpoint for the PostgreSQL RDS instance"
value       = module.db.postgres_endpoint
}

output "redis_endpoint" {
description = "Endpoint for the ElastiCache Redis instance"
value       = module.db.redis_endpoint
}
