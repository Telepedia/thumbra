package models

import "fmt"

type ThumbnailRequest struct {
	Wiki     string
	Hash1    string
	Hash2    string
	Filename string
	Revision string
	Width    string
}

// Get the s3 key for the latest thumbnail
// will be in s3 in something like /{wiki}/thumb/{hash1}/{hash2}/{filename}/{width}px-{filename}
func (ir *ThumbnailRequest) GetS3ThumbKey() string {
	thumbnailName := ir.Width + "px-" + ir.Filename
	return fmt.Sprintf("%s/thumb/%s/%s/%s/%s", ir.Wiki, ir.Hash1, ir.Hash2, ir.Filename, thumbnailName)
}

// Return the s3 key for an archived thumbnail
// will be in s3 in something like /{wiki}/thumb/archive/{hash1}/{hash2}/20250818122033!{filename}/{width}px-{filename}
func (ir *ThumbnailRequest) GetThumbArchiveKey() string {
	filename := ir.Revision + "!" + ir.Filename
	thumbnailName := ir.Width + "px-" + ir.Filename
	return fmt.Sprintf("%s/thumb/archive/%s/%s/%s/%s", ir.Wiki, ir.Hash1, ir.Hash2, filename, thumbnailName)
}
