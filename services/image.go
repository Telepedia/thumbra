package services

import (
	"context"
	"fmt"
	"log"
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

// construct a new image service
func NewImageService(s3Client *storage.S3Client) *ImageService {
	return &ImageService{s3Client: s3Client}
}

// Get the original image from S3. This is used to return the image when the user
// requests either the original image, or a thumbnail that does not exist so that we can generate
// a thumbnail for it
func (is *ImageService) GetOriginalImage(req models.ImageRequest) (*models.ImageResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var s3Key string

	if req.Revision == "latest" {
		s3Key = req.GetS3Key()
	} else {
		s3Key = req.GetArchiveKey()
	}

	obj, err := is.s3Client.GetObject(ctx, s3Key)
	if err != nil {
		if err == storage.ErrObjectNotFound {
			return nil, ErrImageNotFound
		}
		return nil, fmt.Errorf("failed to retrieve image from S3: %w", err)
	}

	return obj, nil
}

// Get the metdata about an image from S3; this can handle both archives and latest images
// it does not handle thumbnails, but it should!
func (is *ImageService) GetImageMetadata(req models.ImageRequest) (*models.ImageResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var s3Key string

	if req.Revision == "latest" {
		s3Key = req.GetS3Key()
	} else {
		s3Key = req.GetArchiveKey()
		log.Println("S3Key", s3Key)
	}

	metadata, err := is.s3Client.HeadObject(ctx, s3Key)
	if err != nil {
		if err == storage.ErrObjectNotFound {
			return nil, ErrImageNotFound
		}
		return nil, fmt.Errorf("failed to retrieve image metadata from S3: %w", err)
	}

	return metadata, nil
}

// Get the metdata about an image from S3; this can handle both archives and latest images
// it does not handle thumbnails, but it should!
func (is *ImageService) GetThumbnailMetadata(req models.ThumbnailRequest) (*models.ImageResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var s3Key string

	if req.Revision == "latest" {
		s3Key = req.GetS3ThumbKey()
	} else {
		s3Key = req.GetThumbArchiveKey()
		log.Println("S3Key", s3Key)
	}

	metadata, err := is.s3Client.HeadObject(ctx, s3Key)
	if err != nil {
		if err == storage.ErrObjectNotFound {
			return nil, ErrImageNotFound
		}
		return nil, fmt.Errorf("failed to retrieve image metadata from S3: %w", err)
	}

	return metadata, nil
}

// Get the original image from S3. This is used to return the image when the user
// requests either the original image, or a thumbnail that does not exist so that we can generate
// a thumbnail for it
func (is *ImageService) GetThumbnail(req models.ThumbnailRequest) (*models.ImageResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var s3Key string

	if req.Revision == "latest" {
		s3Key = req.GetS3ThumbKey()
	} else {
		s3Key = req.GetThumbArchiveKey()
	}

	obj, err := is.s3Client.GetObject(ctx, s3Key)
	if err != nil {
		if err == storage.ErrObjectNotFound {
			return nil, ErrImageNotFound
		}
		return nil, fmt.Errorf("failed to retrieve image from S3: %w", err)
	}

	return obj, nil
}
