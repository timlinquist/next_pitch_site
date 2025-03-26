package models

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client interface for mocking in tests
type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// S3ClientFactory is a function that creates an S3Client
type S3ClientFactory func() (S3Client, error)

// DefaultS3ClientFactory creates a real S3 client
var DefaultS3ClientFactory S3ClientFactory = func() (S3Client, error) {
	// Get AWS credentials from environment
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if accessKeyID == "" || secretAccessKey == "" {
		return nil, fmt.Errorf("AWS credentials or bucket name environment variables are not set")
	}

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"), // Replace with your desired region
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %v", err)
	}

	// Create S3 client
	return s3.NewFromConfig(cfg), nil
}

// CurrentS3ClientFactory is the factory function currently in use
var CurrentS3ClientFactory = DefaultS3ClientFactory

// UploadVideo uploads a video file to S3 and returns the S3 URL and temporary URL
func UploadVideo(file io.Reader, filename string) (string, string, error) {
	bucketName := os.Getenv("AWS_S3_BUCKET")
	if bucketName == "" {
		return "", "", fmt.Errorf("AWS credentials or bucket name environment variables are not set")
	}

	// Get S3 client
	client, err := CurrentS3ClientFactory()
	if err != nil {
		return "", "", err
	}

	// Generate a unique key for the file
	key := fmt.Sprintf("videos/%s", filename)

	// Upload to S3
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to S3: %v", err)
	}

	// For testing, we'll just return a fake presigned URL
	presignedURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, key)

	return key, presignedURL, nil
}

// DeleteVideo deletes a video from S3
func DeleteVideo(key string) error {
	bucketName := os.Getenv("AWS_S3_BUCKET")
	if bucketName == "" {
		return fmt.Errorf("AWS credentials or bucket name environment variables are not set")
	}

	// Get S3 client
	client, err := CurrentS3ClientFactory()
	if err != nil {
		return err
	}

	// Delete from S3
	_, err = client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %v", err)
	}

	return nil
}
