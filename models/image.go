package models

import "fmt"

type ImageRequest struct {
	Wiki     string
	Hash1    string
	Hash2    string
	Filename string
	Revision string
}

// Helper to convert the URL parameters to the file path (albeit virtual)
// file location in S3
// Typically this will return something like: /metawiki/a/a0/foo.png
// this can be used ONLY for originals - since thumbs and archives will
// be in different locations
func (ir *ImageRequest) GetS3Key() string {
	return fmt.Sprintf("%s/%s/%s/%s", ir.Wiki, ir.Hash1, ir.Hash2, ir.Filename)
}
