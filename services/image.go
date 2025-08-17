package services

import (
	"context"
	"fmt"
	"time"

	"github.com/telepedia/thumbra/models"
	"github.com/telepedia/thumbra/storage"
)

type ImageService struct {
	s3Client *storage.S3Client
}

var (
	ErrImageNotFound = fmt.Errorf("image not found")
)

func NewImageService(s3Client *storage.S3Client) *ImageService {
	return &ImageService{s3Client: s3Client}
}

func (is *ImageService) GetOriginalImage(req models.ImageRequest) ([]byte, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s3Key := req.GetS3Key()
	data, contentType, err := is.s3Client.GetObject(ctx, s3Key)
	if err != nil {
		if err == storage.ErrObjectNotFound {
			return nil, "", ErrImageNotFound
		}
		return nil, "", fmt.Errorf("failed to retrieve image from S3: %w", err)
	}

	return data, contentType, nil
}
