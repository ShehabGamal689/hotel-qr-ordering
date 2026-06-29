output "bucket_name" {
  description = "The name of the S3 bucket"
  value       = aws_s3_bucket.hotel_catalog_images.bucket
}

output "bucket_arn" {
  description = "The ARN of the S3 bucket"
  value       = aws_s3_bucket.hotel_catalog_images.arn
}
