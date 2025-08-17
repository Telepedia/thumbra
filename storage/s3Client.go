package storage

import (
	"context"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/telepedia/thumbra/config"
)

type S3Client struct {
	S3     *s3.Client
	Bucket string
}

func New(cfg *config.Config) *S3Client {
	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithRegion(cfg.S3.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.S3.AccessKey,
				cfg.S3.SecretKey,
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
		Bucket: cfg.S3.Bucket,
	}
}
