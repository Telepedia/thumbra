package models

import (
	"fmt"
	"time"
)

type ImageRequest struct {
	Wiki     string
	Hash1    string
	Hash2    string
	Filename string
	Revision string
}

// metadata associated with this image in S3 such as the
// length, content type etc (we pull this from S3 to avoid calculating
// ourselves and wasting processing time since the response will
// already contain it)
type ImageResponse struct {
	Data               []byte
	ContentType        string
	Length             int64
	ETag               string
	LastModified       time.Time
	ContentDisposition string
}

// Helper to convert the URL parameters to the file path (albeit virtual)
// file location in S3
// Typically this will return something like: /metawiki/a/a0/foo.png
// this can be used ONLY for originals - since thumbs and archives will
// be in different locations
func (ir *ImageRequest) GetS3Key() string {
	return fmt.Sprintf("%s/%s/%s/%s", ir.Wiki, ir.Hash1, ir.Hash2, ir.Filename)
}

// Helper to convert the URL into a S3 path, taking into account files will
// be stored in the archive table for revisions, and have names like TIMESTAMP!FILENAME.EXT
func (ir *ImageRequest) GetArchiveKey() string {
	filename := ir.Revision + "!" + ir.Filename
	return fmt.Sprintf("%s/archive/%s/%s/%s", ir.Wiki, ir.Hash1, ir.Hash2, filename)
}
