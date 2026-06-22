package repository

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type StorageRepository interface {
	UploadFile(ctx context.Context, filename string, file io.Reader, contentType string) (string, error)
}

type S3StorageRepository struct {
	client     *s3.Client
	bucketName string
	publicURL  string
}

// NewS3StorageRepository instantiates a new client configured for AWS S3 or MinIO.
// DevOps Config Checklist:
//   - Set STORAGE_ENDPOINT to your MinIO container url (e.g. "http://localhost:9000") in dev.
//   - In AWS production, leave STORAGE_ENDPOINT empty to use standard AWS S3 endpoint routing.
//   - Set STORAGE_ACCESS_KEY and STORAGE_SECRET_KEY for static credential exchange.
//   - Set STORAGE_REGION (e.g. "us-east-1") and STORAGE_BUCKET_NAME.
//   - Set STORAGE_PUBLIC_URL_PREFIX if you serve objects through a CDN (CloudFront/Cloudflare) or a proxy.
func NewS3StorageRepository(ctx context.Context) (*S3StorageRepository, error) {
	bucket := os.Getenv("STORAGE_BUCKET_NAME")
	if bucket == "" {
		bucket = "hotel-catalog-images"
	}

	endpoint := os.Getenv("STORAGE_ENDPOINT")
	accessKey := os.Getenv("STORAGE_ACCESS_KEY")
	secretKey := os.Getenv("STORAGE_SECRET_KEY")
	region := os.Getenv("STORAGE_REGION")
	publicURL := os.Getenv("STORAGE_PUBLIC_URL_PREFIX")

	if region == "" {
		region = "us-east-1"
	}

	var cfg aws.Config
	var err error

	if accessKey != "" && secretKey != "" {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
	} else {
		// Use default credentials chain (IAM Roles, AWS CLI environment variables, etc.)
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
	}

	if err != nil {
		return nil, fmt.Errorf("unable to load AWS/MinIO SDK configuration: %w", err)
	}

	var clientOpts []func(*s3.Options)
	if endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true // Mandatory configuration parameter for local MinIO deployments
		})
	}

	client := s3.NewFromConfig(cfg, clientOpts...)

	// Auto-create bucket if endpoint is set (convenient for local MinIO startup)
	if endpoint != "" {
		go func() {
			ctx := context.Background()
			_, createErr := client.CreateBucket(ctx, &s3.CreateBucketInput{
				Bucket: aws.String(bucket),
			})
			if createErr != nil {
				log.Printf("MinIO: Auto-bucket check completed. (Note: Bucket may already exist: %v)", createErr)
			} else {
				log.Printf("MinIO: Created bucket '%s' successfully.", bucket)
			}

			policy := fmt.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Sid": "PublicRead",
						"Effect": "Allow",
						"Principal": "*",
						"Action": ["s3:GetObject"],
						"Resource": ["arn:aws:s3:::%s/*"]
					}
				]
			}`, bucket)
			_, policyErr := client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
				Bucket: aws.String(bucket),
				Policy: aws.String(policy),
			})
			if policyErr != nil {
				log.Printf("MinIO: Failed to set public read policy: %v", policyErr)
			} else {
				log.Printf("MinIO: Configured public read policy for bucket '%s'.", bucket)
			}
		}()
	}

	return &S3StorageRepository{
		client:     client,
		bucketName: bucket,
		publicURL:  publicURL,
	}, nil
}

func (s *S3StorageRepository) UploadFile(ctx context.Context, filename string, file io.Reader, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(filename),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object to storage: %w", err)
	}

	if s.publicURL != "" {
		return fmt.Sprintf("%s/%s", s.publicURL, filename), nil
	}

	endpoint := os.Getenv("STORAGE_ENDPOINT")
	if endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", endpoint, s.bucketName, filename), nil
	}

	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucketName, filename), nil
}
