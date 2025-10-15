package utils

import (
	"fmt"
	"strings"

	"github.com/telepedia/thumbra/models"
)

var (
	ErrInvalidWiki     = fmt.Errorf("invalid wiki name")
	ErrInvalidHash     = fmt.Errorf("invalid hash structure")
	ErrInvalidFileName = fmt.Errorf("invalid file name")
	ErrInvalidRevision = fmt.Errorf("invalid revision")
	ErrInvalidWidth    = fmt.Errorf("invalid width")
	ErrInvalidHeight   = fmt.Errorf("invalid height")
)

// validate that the request is valid and correctly formed
func ValidateImageRequest(req models.ImageRequest) error {
	if req.Wiki == "" {
		return ErrInvalidWiki
	}

	if req.Hash1 == "" || len(req.Hash1) > 2 {
		return ErrInvalidHash
	}

	if req.Hash2 == "" || len(req.Hash2) > 3 || len(req.Hash2) < 2 || !strings.HasPrefix(req.Hash2, req.Hash1) {
		return ErrInvalidHash
	}

	if req.Filename == "" {
		return ErrInvalidFileName
	}

	if req.Revision == "" {
		return ErrInvalidRevision
	}

	// we need to check that the revision is valid here in MediaWiki format
	// but I can't deal with doing that rn so maybe later

	return nil
}

// validate that a request to get a thumbnail is valid and correctly formed
// it will need to contain the width - the height is currently not supported as the image will be scaled
// and the aspect ratio retained
func ValidateThumbnailRequest(req models.ThumbnailRequest) error {
	if req.Wiki == "" {
		return ErrInvalidWiki
	}

	if req.Hash1 == "" || len(req.Hash1) > 2 {
		return ErrInvalidHash
	}

	if req.Hash2 == "" || len(req.Hash2) > 3 || len(req.Hash2) < 2 || !strings.HasPrefix(req.Hash2, req.Hash1) {
		return ErrInvalidHash
	}

	if req.Filename == "" {
		return ErrInvalidFileName
	}

	if req.Revision == "" {
		return ErrInvalidRevision
	}

	if req.Width == "" {
		return ErrInvalidWidth
	}

	// we need to check that the revision is valid here in MediaWiki format
	// but I can't deal with doing that rn so maybe later

	return nil
}
