package dl

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestRemuxOptionsFromInfo(t *testing.T) {
	info := Info{
		Title:       " My Title ",
		AltTitle:    "Alt",
		Description: "  Desc  ",
		Artist:      "Artist Name",
		Timestamp:   float64(time.Date(2024, 12, 31, 23, 59, 58, 0, time.UTC).Unix()),
	}
	opt := RemuxOptionsFromInfo("ffmpeg", fs.VideoMp4, info, "https://example.com/v")
	if opt.Title != "My Title" {
		t.Fatalf("Title mismatch: %q", opt.Title)
	}
	if opt.Description != "Desc" {
		t.Fatalf("Description mismatch: %q", opt.Description)
	}
	if opt.Author != "Artist Name" {
		t.Fatalf("Author mismatch: %q", opt.Author)
	}
	if opt.Comment != "https://example.com/v" {
		t.Fatalf("Comment mismatch: %q", opt.Comment)
	}
	if opt.Created.IsZero() {
		t.Fatalf("Created timestamp should be set")
	}
}

func TestCreatedFromInfo_UploadDateFallback(t *testing.T) {
	info := Info{
		Title:      "X",
		UploadDate: "20211120",
	}
	created := CreatedFromInfo(info)
	if created.IsZero() {
		t.Fatalf("expected created time from UploadDate fallback")
	}
	if got, want := created.UTC().Format(time.RFC3339), "2021-11-20T00:00:00Z"; got != want {
		t.Fatalf("created mismatch: got %s want %s", got, want)
	}
}

func TestCreatedFromInfo_ReleaseDateFallback(t *testing.T) {
	info := Info{
		Title:       "Y",
		ReleaseDate: "20190501",
	}
	created := CreatedFromInfo(info)
	if created.IsZero() {
		t.Fatalf("expected created time from ReleaseDate fallback")
	}
	if got, want := created.UTC().Format(time.RFC3339), "2019-05-01T00:00:00Z"; got != want {
		t.Fatalf("created mismatch: got %s want %s", got, want)
	}
}
