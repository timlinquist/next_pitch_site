package models

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// UploadVideo uploads a video file to S3 and returns the S3 URL and temporary URL
func UploadVideo(file io.Reader, filename string) (string, string, error) {
	// Get AWS credentials from environment
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("AWS_S3_BUCKET")

	if accessKeyID == "" || secretAccessKey == "" || bucketName == "" {
		return "", "", fmt.Errorf("AWS credentials or bucket name environment variables are not set")
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
		return "", "", fmt.Errorf("unable to load AWS config: %v", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

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

	// Generate a presigned URL that expires in 1 hour
	presignClient := s3.NewPresignClient(client)
	presignResult, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Hour
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return key, presignResult.URL, nil
}

// DeleteVideo deletes a video from S3
func DeleteVideo(key string) error {
	// Get AWS credentials from environment
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("AWS_S3_BUCKET")

	if accessKeyID == "" || secretAccessKey == "" || bucketName == "" {
		return fmt.Errorf("AWS credentials or bucket name environment variables are not set")
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
		return fmt.Errorf("unable to load AWS config: %v", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

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
