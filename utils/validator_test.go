package utils

import (
	"testing"

	"github.com/telepedia/thumbra/models"
)

func TestValidateImageRequest(t *testing.T) {
	good := models.ImageRequest{
		Wiki:     "metawiki",
		Hash1:    "a",
		Hash2:    "a0",
		Filename: "foo.png",
		Revision: "latest",
	}

	if err := ValidateImageRequest(good); err != nil {
		t.Fatalf("expected nil error for valid request, got %v", err)
	}

	bad := models.ImageRequest{}
	if err := ValidateImageRequest(bad); err == nil {
		t.Fatalf("expected error for invalid request, got nil")
	}
}

func TestValidateThumbnailRequest(t *testing.T) {
	good := models.ThumbnailRequest{
		Wiki:     "metawiki",
		Hash1:    "a",
		Hash2:    "a0",
		Filename: "foo.png",
		Revision: "latest",
		Width:    "200",
	}

	if err := ValidateThumbnailRequest(good); err != nil {
		t.Fatalf("expected nil error for valid thumbnail request, got %v", err)
	}

	bad := models.ThumbnailRequest{}
	if err := ValidateThumbnailRequest(bad); err == nil {
		t.Fatalf("expected error for invalid thumbnail request, got nil")
	}
}
