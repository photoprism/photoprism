package commands

import "testing"

func TestFormatCount(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		singular string
		plural   string
		want     string
	}{
		{name: "Zero", count: 0, singular: "error", plural: "errors", want: "0 errors"},
		{name: "One", count: 1, singular: "download", plural: "downloads", want: "1 download"},
		{name: "Many", count: 2, singular: "file", plural: "files", want: "2 files"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatCount(tt.count, tt.singular, tt.plural); got != tt.want {
				t.Fatalf("formatCount() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatFailedCount(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		singular string
		plural   string
		want     string
	}{
		{name: "One", count: 1, singular: "download", plural: "downloads", want: "1 download failed"},
		{name: "Many", count: 2, singular: "file", plural: "files", want: "2 files failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatFailedCount(tt.count, tt.singular, tt.plural); got != tt.want {
				t.Fatalf("formatFailedCount() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatVideoSummary(t *testing.T) {
	t.Run("DryRun", func(t *testing.T) {
		want := "transcode: would process 1 file, skipped 0 files"
		if got := formatVideoSummary("transcode", true, 1, 0, 0, 0); got != want {
			t.Fatalf("formatVideoSummary() = %q, want %q", got, want)
		}
	})
	t.Run("Run", func(t *testing.T) {
		want := "trim: processed 2 files, skipped 1 file, 0 files failed"
		if got := formatVideoSummary("trim", false, 0, 2, 1, 0); got != want {
			t.Fatalf("formatVideoSummary() = %q, want %q", got, want)
		}
	})
}
