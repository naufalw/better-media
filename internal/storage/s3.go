package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	Client     *s3.Client
	Downloader *manager.Downloader
	Uploader   *manager.Uploader
	BucketName string
}

func NewS3Client(bucketName, endpoint, region string) (*S3Client, error) {
	accessKeyId := os.Getenv("S3_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("S3_ACCESS_KEY_SECRET")

	// TODO!!! This is hacky, to detect minio or localstack
	isMinIO := strings.Contains(endpoint, "minio") || strings.Contains(endpoint, "localhost")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		config.WithRegion(region),
	)

	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = isMinIO
	})

	return &S3Client{
		Client:     client,
		Downloader: manager.NewDownloader(client),
		Uploader:   manager.NewUploader(client),

		BucketName: bucketName,
	}, nil

}

func (s *S3Client) DownloadFile(ctx context.Context, objectKey, localPath string) error {
	file, err := os.Create(localPath)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = s.Downloader.Download(ctx, file, &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
	})

	return err
}

func (s *S3Client) UploadFile(ctx context.Context, localPath, objectKey string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = s.Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})

	return err
}

func (s *S3Client) GeneratePresignedPut(ctx context.Context, objectKey string, validDuration time.Duration) (*v4.PresignedHTTPRequest, error) {
	presignClient := s3.NewPresignClient(s.Client)
	presignResult, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(validDuration))

	if err != nil {
		return nil, err
	}

	return presignResult, nil
}

func (s *S3Client) GeneratePresignedGet(ctx context.Context, objectKey string, validDuration time.Duration) (*v4.PresignedHTTPRequest, error) {
	presignClient := s3.NewPresignClient(s.Client)
	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(validDuration))

	if err != nil {
		return nil, err
	}

	return presignResult, nil
}

func (s *S3Client) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	fmt.Println("Getting object from S3:", key)
	fmt.Println("Bucket Name:", s.BucketName)
	output, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		fmt.Println("Error getting object:", err)
		return nil, err
	}
	return output.Body, nil
}
