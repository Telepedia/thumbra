package services

import "github.com/telepedia/thumbra/storage"

type ImageService struct {
	s3Client *storage.S3Client
}

func NewImageService(s3Client *storage.S3Client) *ImageService {
	return &ImageService{s3Client: s3Client}
}
