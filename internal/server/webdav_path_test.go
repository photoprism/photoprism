package server

import "testing"

func TestIsWebDAVPath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		basePaths []string
		want      bool
	}{
		{name: "RootOriginals", path: "/originals", want: true},
		{name: "RootImportChild", path: "/import/folder/file.jpg", want: true},
		{name: "ProxyOriginals", path: "/i/acme/originals/", want: true},
		{name: "ProxyImport", path: "/i/acme/import/a.jpg", want: true},
		{name: "CustomBaseOriginals", path: "/instance-a/originals/album", basePaths: []string{"/instance-a/originals", "/instance-a/import"}, want: true},
		{name: "CustomBaseImport", path: "/instance-a/import", basePaths: []string{"/instance-a/originals", "/instance-a/import"}, want: true},
		{name: "LibraryPath", path: "/library/browse", want: false},
		{name: "ProxyNonDAV", path: "/i/acme/library", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsWebDAVPath(tc.path, tc.basePaths...); got != tc.want {
				t.Fatalf("IsWebDAVPath(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}
