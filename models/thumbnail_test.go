package models

import "testing"

// Test that a thumbnail request returns the appropriate S3 thumb key
func TestGetS3ThumbKey(t *testing.T) {
	tr := &ThumbnailRequest{
		Wiki:     "metawiki",
		Hash1:    "a",
		Hash2:    "a0",
		Filename: "foo.png",
		Width:    "200",
	}

	expectedKey := "metawiki/thumb/a/a0/foo.png/200px-foo.png"
	if got := tr.GetS3ThumbKey(); got != expectedKey {
		t.Fatalf("GetS3ThumbKey() = %q, but we expected %q", got, expectedKey)
	}
}

// Test that an archive thumbnail request returns the appropriate thumb archive key
func TestGetThumbArchiveKey(t *testing.T) {
	tr := &ThumbnailRequest{
		Wiki:     "metawiki",
		Hash1:    "a",
		Hash2:    "a0",
		Filename: "foo.png",
		Width:    "200",
		Revision: "20251021233101",
	}

	expectedKey := "metawiki/thumb/archive/a/a0/20251021233101!foo.png/200px-foo.png"
	if got := tr.GetThumbArchiveKey(); got != expectedKey {
		t.Fatalf("GetThumbArchiveKey() = %q, want %q", got, expectedKey)
	}
}
