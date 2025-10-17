package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/telepedia/thumbra/config"
	"github.com/telepedia/thumbra/models"
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

// Get an object from S3 returning the metadata about the file, or an error
func (s *S3Client) GetObject(ctx context.Context, key string) (*models.ImageResponse, error) {
	input := &s3.GetObjectInput{
		Bucket: &s.Bucket,
		Key:    &key,
	}

	result, err := s.S3.GetObject(ctx, input)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	resp := &models.ImageResponse{
		Data: data,
	}

	if result.ContentType != nil {
		resp.ContentType = *result.ContentType
	}
	if result.ContentLength != nil {
		resp.Length = *result.ContentLength
	}
	if result.ETag != nil {
		resp.ETag = *result.ETag
	}
	if result.LastModified != nil {
		resp.LastModified = *result.LastModified
	}

	return resp, nil
}

// Send a HEAD request to S3, so that we can get the data about the object
// this allows us to signal to the browser that their cached copy is fine to use
// and avoids us having to send a full GET to S3 which is more expensive
func (s *S3Client) HeadObject(ctx context.Context, key string) (*models.ImageResponse, error) {
	input := &s3.HeadObjectInput{
		Bucket: &s.Bucket,
		Key:    &key,
	}

	result, err := s.S3.HeadObject(ctx, input)
	if err != nil {
		return nil, err
	}

	resp := &models.ImageResponse{
		Data: nil,
	}

	if result.ContentType != nil {
		resp.ContentType = *result.ContentType
	}
	if result.ContentLength != nil {
		resp.Length = *result.ContentLength
	}
	if result.ETag != nil {
		resp.ETag = *result.ETag
	}
	if result.LastModified != nil {
		resp.LastModified = *result.LastModified
	}
	if result.ContentDisposition != nil {
		resp.ContentDisposition = *result.ContentDisposition
	}

	return resp, nil
}

// Put an object (thumb) into S3, mainly a wrapper around the existing s3 client
func (s *S3Client) PutObject(ctx context.Context, key string, data []byte, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket:      &s.Bucket,
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
		ACL:         "public-read",
	}

	_, err := s.S3.PutObject(ctx, input)
	if err != nil {
		return err
	}

	return nil
}
