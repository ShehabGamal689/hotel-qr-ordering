output "backend_repo_url" {
value = aws_ecr_repository.backend.repository_url
}

output "frontend_repo_url" {
value = aws_ecr_repository.frontend.repository_url
}

output "cluster_name" {
value = module.ecs_cluster.cluster_name
}

output "ecs_security_group_id" {
value = aws_security_group.ecs_tasks_sg.id
}
