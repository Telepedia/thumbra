package services

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/telepedia/thumbra/models"
)

// Encode and decode a PNG image to test helper functions
func TestDecodeEncodePNG(t *testing.T) {
	// 10x10 red square for testing
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode png: %v", err)
	}

	dec, err := decodeImage(bytes.NewReader(buf.Bytes()), "png")
	if err != nil {
		t.Fatalf("decodeImage failed: %v", err)
	}
	if dec.Bounds().Dx() != 10 || dec.Bounds().Dy() != 10 {
		t.Fatalf("decoded image has wrong dimensions: %v", dec.Bounds())
	}

	tmp := t.TempDir()
	outPath := filepath.Join(tmp, "out.png")
	f, err := os.Create(outPath)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if err := encodeImage(f, dec, "png"); err != nil {
		t.Fatalf("encodeImage failed: %v", err)
	}
	f.Close()
}

// Test thumbnailing an image
func TestThumbnailImage(t *testing.T) {
	// 100x50 png
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode png: %v", err)
	}

	obj := &models.ImageResponse{Data: buf.Bytes()}
	svc := &ImageService{}

	// this is larger than the original width, therefore expect an error
	tr := models.ThumbnailRequest{Filename: "foo.png", Width: "200"}
	if _, err := svc.ThumbnailImage(tr, obj); err == nil {
		t.Fatalf("expected error when requesting width larger than original, got nil")
	}

	// not actually interested in the image, just check that it returned a valid path
	tr2 := models.ThumbnailRequest{Filename: "foo.png", Width: "50"}
	path, err := svc.ThumbnailImage(tr2, obj)
	if err != nil {
		t.Fatalf("ThumbnailImage failed: %v", err)
	}
	// cleanup
	os.Remove(path)
}
