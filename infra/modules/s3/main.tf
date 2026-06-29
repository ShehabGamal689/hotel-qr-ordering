resource "aws_s3_bucket" "hotel_catalog_images" {
  bucket = "hotel-qr-catalog-images-prod" # Note: S3 bucket names must be globally unique!
}

resource "aws_s3_bucket_public_access_block" "public_access" {
  bucket = aws_s3_bucket.hotel_catalog_images.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_policy" "public_read" {
  bucket = aws_s3_bucket.hotel_catalog_images.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "PublicReadGetObject"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetObject"
        Resource  = "${aws_s3_bucket.hotel_catalog_images.arn}/*"
      }
    ]
  })
  depends_on = [aws_s3_bucket_public_access_block.public_access]
}


