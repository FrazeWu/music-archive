package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// InitS3Client initializes and returns an S3 client configured for Oracle Object Storage
func InitS3Client() (*s3.S3, error) {
	// AWS S3 session (compatible with Oracle Object Storage)
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(os.Getenv("S3_REGION")),
		Endpoint:         aws.String(os.Getenv("S3_ENDPOINT")),
		Credentials:      credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
		S3ForcePathStyle: aws.Bool(true), // Required for Oracle Object Storage
	})
	if err != nil {
		return nil, err
	}

	s3Client := s3.New(sess)
	return s3Client, nil
}

// ValidateS3Connection tests the S3 connection and permissions
func ValidateS3Connection(s3Client *s3.S3) error {
	bucket := os.Getenv("S3_BUCKET")

	// Test bucket access
	_, err := s3Client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		log.Printf("S3 validation failed: %v", err)
		return err
	}

	log.Printf("S3 connection validated successfully for bucket: %s", bucket)
	return nil
}
