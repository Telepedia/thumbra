package storage

import (
	"context"
	"fmt"
	"io"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/telepedia/thumbra/config"
)

type S3Client struct {
	S3     *s3.Client
	Bucket string
}

var (
	ErrObjectNotFound = fmt.Errorf("object not found")
)

func New(cfg config.S3Config) *S3Client {
	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretKey,
				"",
			),
		),
	)
	if err != nil {
		panic(err)
	}

	s3Client := s3.NewFromConfig(awsCfg)

	return &S3Client{
		S3:     s3Client,
		Bucket: cfg.Bucket,
	}
}

func (s *S3Client) GetObject(ctx context.Context, key string) ([]byte, string, error) {
	input := &s3.GetObjectInput{
		Bucket: &s.Bucket,
		Key:    &key,
	}

	result, err := s.S3.GetObject(ctx, input)
	if err != nil {
		return nil, "", err
	}

	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := *result.ContentType
	return data, contentType, nil
}
