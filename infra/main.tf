module "vpc" {
source = "./modules/vpc"
}

module "alb" {
source          = "./modules/alb"
vpc_id          = module.vpc.vpc_id
public_subnets  = module.vpc.public_subnets
certificate_arn = module.dns.certificate_arn
}

module "ecs" {
source                    = "./modules/ecs"
vpc_id                    = module.vpc.vpc_id
public_subnets            = module.vpc.public_subnets
private_subnets           = module.vpc.private_subnets
alb_security_group_id     = module.alb.alb_security_group_id
backend_target_group_arn  = module.alb.backend_target_group_arn
frontend_target_group_arn = module.alb.frontend_target_group_arn
database_endpoint         = module.db.postgres_endpoint
redis_endpoint            = module.db.redis_endpoint
database_password = "HotelPass2026"
s3_bucket_name =            module.s3.bucket_name
s3_bucket_arn  =            module.s3.bucket_arn
}

module "db" {
source                = "./modules/db"
vpc_id                = module.vpc.vpc_id
private_subnets       = module.vpc.private_subnets
db_subnet_group_name  = module.vpc.database_subnet_group_name
ecs_security_group_id = module.ecs.ecs_security_group_id
}

module "dns" {
source            = "./modules/dns"
domain_name       = "devopsnawy.qzz.io"
registrar_api_key = var.registrar_api_key
cluster_name      = "hotel-qr"
alb_dns_name      = module.alb.alb_dns_name
alb_zone_id       = module.alb.alb_zone_id
}

module "s3" {
  source = "./modules/s3"
}
