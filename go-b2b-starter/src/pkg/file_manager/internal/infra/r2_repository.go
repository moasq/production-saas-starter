package infra

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

	fileconfig "github.com/moasq/go-b2b-starter/pkg/file_manager/config"
	"github.com/moasq/go-b2b-starter/pkg/file_manager/domain"
)

type r2Repository struct {
	client     *s3.Client
	bucketName string
}

func NewR2Repository(cfg *fileconfig.Config) (domain.R2Repository, error) {
	// Create custom AWS config for R2
	r2Cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.R2.Region), // Always "auto" for R2
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.R2.AccessKeyID,
			cfg.R2.SecretAccessKey,
			"", // No session token needed
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 config: %w", err)
	}

	// Create S3 client with R2 endpoint
	client := s3.NewFromConfig(r2Cfg, func(o *s3.Options) {
		// R2 endpoint format: https://<account_id>.r2.cloudflarestorage.com
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com",
			cfg.R2.AccountID))

		// R2 only supports path-style URLs (not virtual-host style)
		o.UsePathStyle = true
	})

	repo := &r2Repository{
		client:     client,
		bucketName: cfg.R2.BucketName,
	}

	// Ensure bucket exists (R2 doesn't auto-create buckets)
	if err := repo.ensureBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure R2 bucket exists: %w", err)
	}

	return repo, nil
}

// ensureBucket checks if bucket exists (R2 requires manual bucket creation)
func (r *r2Repository) ensureBucket(ctx context.Context) error {
	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(r.bucketName),
	})

	if err != nil {
		return fmt.Errorf("bucket '%s' does not exist in R2. Please create it manually in Cloudflare dashboard: %w",
			r.bucketName, err)
	}

	return nil
}

func (r *r2Repository) UploadObject(ctx context.Context, objectKey string, content io.Reader, size int64, contentType string) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r.bucketName),
		Key:           aws.String(objectKey),
		Body:          content,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})

	if err != nil {
		return fmt.Errorf("failed to upload object to R2: %w", err)
	}

	return nil
}

// DownloadObject downloads a file from R2
func (r *r2Repository) DownloadObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	result, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get object from R2: %w", err)
	}

	return result.Body, nil
}

func (r *r2Repository) DeleteObject(ctx context.Context, objectKey string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		return fmt.Errorf("failed to delete object from R2: %w", err)
	}

	return nil
}

// GetPresignedURL generates a presigned URL for temporary access
func (r *r2Repository) GetPresignedURL(ctx context.Context, objectKey string, expiryHours int) (string, error) {
	// Create presign client
	presignClient := s3.NewPresignClient(r.client)

	// Generate presigned URL
	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expiryHours) * time.Hour
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate R2 presigned URL: %w", err)
	}

	return request.URL, nil
}

// ObjectExists checks if an object exists in R2
func (r *r2Repository) ObjectExists(ctx context.Context, objectKey string) (bool, error) {
	_, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		// Check if error is "NotFound" using AWS SDK error handling
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "NoSuchKey" {
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to check R2 object existence: %w", err)
	}

	return true, nil
}
