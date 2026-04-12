package repository

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/serhatkilbas/lms-poc/internal/domain"
	"io"
)

type s3MediaStorage struct {
	client *s3.Client
	presignClient *s3.PresignClient
	bucket string
}

func NewS3MediaStorage(accessKey, secretKey, endpoint, bucket string) (domain.MediaStorage, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpoint,
			HostnameImmutable: true,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(client)

	return &s3MediaStorage{
		client: client,
		presignClient: presignClient,
		bucket: bucket,
	}, nil
}

func (s *s3MediaStorage) UploadFile(ctx context.Context, filePath, fileName string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return s.UploadStream(ctx, file, fileName, "")
}

func (s *s3MediaStorage) UploadStream(ctx context.Context, reader io.Reader, fileName, contentType string) (string, error) {
	uploader := manager.NewUploader(s.client)
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileName),
		Body:   reader,
	}

	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	result, err := uploader.Upload(ctx, input)
	if err != nil {
		return "", err
	}

	return result.Location, nil
}

func (s *s3MediaStorage) GetPresignedURL(ctx context.Context, fileName string, expires time.Duration) (string, error) {
	request, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileName),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expires
	})
	if err != nil {
		return "", err
	}

	return request.URL, nil
}

func (s *s3MediaStorage) UploadDirectory(ctx context.Context, localDir, remotePrefix string) error {
	files, err := os.ReadDir(localDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		localFilePath := filepath.Join(localDir, file.Name())
		remotePath := filepath.Join(remotePrefix, file.Name())

		_, err := s.UploadFile(ctx, localFilePath, remotePath)
		if err != nil {
			return err
		}
	}

	return nil
}
