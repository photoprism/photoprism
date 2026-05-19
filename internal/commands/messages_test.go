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
