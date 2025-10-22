package models

import "testing"

// Test that an image request returns the appropriate key
func TestGetS3Key(t *testing.T) {
	ir := &ImageRequest{
		Wiki:     "metawiki",
		Hash1:    "a",
		Hash2:    "a0",
		Filename: "foo.png",
	}

	expectedKey := "metawiki/a/a0/foo.png"
	if got := ir.GetS3Key(); got != expectedKey {
		t.Fatalf("GetS3Key() = %q, but we expected %q", got, expectedKey)
	}
}

// Test that an image request with a revision returns the appropriate archive key
func TestGetArchiveKey(t *testing.T) {
	ir := &ImageRequest{
		Wiki:     "metawiki",
		Hash1:    "a",
		Hash2:    "a0",
		Filename: "foo.png",
		Revision: "20251021233101",
	}

	expectedKey := "metawiki/archive/a/a0/20251021233101!foo.png"
	if got := ir.GetArchiveKey(); got != expectedKey {
		t.Fatalf("GetArchiveKey() = %q, but we expected %q", got, expectedKey)
	}
}
