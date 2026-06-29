output "postgres_endpoint" {
value = module.db.db_instance_endpoint
}

output "redis_endpoint" {
value = aws_elasticache_cluster.redis.cache_nodes[0].address
}

