package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/HugoSmits86/nativewebp"
	"github.com/aws/smithy-go"
	"github.com/disintegration/imaging"
	"github.com/telepedia/thumbra/models"
	"github.com/telepedia/thumbra/storage"
)

type ImageService struct {
	s3Client *storage.S3Client
}

var (
	ErrImageNotFound = fmt.Errorf("image not found")
	ErrWidthTooLarge = errors.New("requested width exceeds original image width, caller should return original image")
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

	s3Key := is.s3KeyForImage(req)
	return is.fetchObject(ctx, s3Key)
}

// Get the metdata about an image from S3; this can handle both archives and latest images
// it does not handle thumbnails, but it should!
func (is *ImageService) GetImageMetadata(req models.ImageRequest) (*models.ImageResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s3Key := is.s3KeyForImage(req)

	return is.headObjectByKey(ctx, s3Key)
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

	return is.headObjectByKey(ctx, s3Key)
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

	return is.fetchObject(ctx, s3Key)
}

// Take the original image and generate a thumbnail, storing it in the temp dir and returning the path
// the caller is responsible for uploading the thumbnail to S3 and deleting the temporary file
// @TODO: investigate whether this function should upload the thumbnail to S3 itself
func (is *ImageService) ThumbnailImage(req models.ThumbnailRequest, obj *models.ImageResponse) (string, error) {
	// find out what the file type is from the extension
	ext := strings.ToLower(filepath.Ext(req.Filename))
	format := strings.TrimPrefix(ext, ".")

	// Decode the image
	img, err := decodeImage(bytes.NewReader(obj.Data), format)
	if err != nil {
		return "", fmt.Errorf("failed to decode original image: %w", err)
	}

	origWidth := img.Bounds().Dx()
	requestWidth, err := strconv.Atoi(req.Width)
	if err != nil {
		return "", fmt.Errorf("error when converting the width to an int: %s", req.Width)
	}

	if requestWidth > origWidth {
		return "", ErrWidthTooLarge
	}

	// do the actual resizing, obviously and write it to the temp directory
	thumb := imaging.Resize(img, requestWidth, 0, imaging.Lanczos)

	tmpFile, err := os.CreateTemp("", "thumb-*"+ext)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	if err := encodeImage(tmpFile, thumb, format); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return tmpFile.Name(), nil
}

// Decode an image and return it
func decodeImage(r io.Reader, format string) (image.Image, error) {
	switch format {
	case "webp":
		return nativewebp.Decode(r)
	case "jpg", "jpeg":
		return jpeg.Decode(r)
	case "png":
		return png.Decode(r)
	case "gif":
		return gif.Decode(r)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}
}

// Encode the image with the specified format
func encodeImage(w io.Writer, img image.Image, format string) error {
	switch format {
	case "webp":
		return nativewebp.Encode(w, img, nil)
	case "jpg", "jpeg":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 85})
	case "png":
		return png.Encode(w, img)
	case "gif":
		return gif.Encode(w, img, nil)
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}
}

func (is *ImageService) UploadThumbnail(req models.ThumbnailRequest, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read thumbnail file: %w", err)
	}

	// try get the content type from the extension
	ext := strings.ToLower(filepath.Ext(req.Filename))
	contentType := getContentType(ext)

	var key string

	if req.Revision == "latest" {
		key = is.s3KeyForThumbnail(req)
	} else {
		key = is.s3KeyForThumbnail(req)
	}

	// Upload to S3
	err = is.s3Client.PutObject(context.Background(), key, data, contentType)
	if err != nil {
		return fmt.Errorf("failed to upload thumbnail to S3: %w", err)
	}

	return nil
}

// utility function to get the S3 key for either latest or archive images
func (is *ImageService) s3KeyForImage(req models.ImageRequest) string {
	if req.Revision == "latest" {
		return req.GetS3Key()
	}
	return req.GetArchiveKey()
}

// utility function to get the S3 key for either latest or archive thumbnails
func (is *ImageService) s3KeyForThumbnail(req models.ThumbnailRequest) string {
	if req.Revision == "latest" {
		return req.GetS3ThumbKey()
	}
	return req.GetThumbArchiveKey()
}

// wrapper around S3 GetObject that returns errors that Thumbra can understand
func (is *ImageService) fetchObject(ctx context.Context, key string) (*models.ImageResponse, error) {
	obj, err := is.s3Client.GetObject(ctx, key)
	if err != nil {
		if err == storage.ErrObjectNotFound {
			return nil, ErrImageNotFound
		}
		return nil, fmt.Errorf("failed to retrieve image from S3: %w", err)
	}
	return obj, nil
}

// Wrapper for S3 HeadObject to get metadata about an object
// returns errors that Thumbra can understand (since the S3 api is weird)
func (is *ImageService) headObjectByKey(ctx context.Context, key string) (*models.ImageResponse, error) {
	metadata, err := is.s3Client.HeadObject(ctx, key)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NotFound" {
			return nil, ErrImageNotFound
		}
		if err == storage.ErrObjectNotFound {
			return nil, ErrImageNotFound
		}
		return nil, fmt.Errorf("failed to retrieve image metadata from S3: %w", err)
	}
	return metadata, nil
}

func getContentType(ext string) string {
	switch strings.TrimPrefix(ext, ".") {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	default:
		// might need to forgoe this if we can't understand the
		// conent type, return errrorrrrrrrr?
		return "application/octet-stream"
	}
}
