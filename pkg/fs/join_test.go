package fs

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestSafeJoin(t *testing.T) {
	base := filepath.Clean("/tmp/base")

	if runtime.GOOS == "windows" {
		base = filepath.Clean(`C:\tmp\base`)
	}

	type testCase struct {
		name     string
		baseDir  string
		input    string
		wantErr  bool
		wantPath string
	}

	tests := []testCase{
		{
			name:     "SimpleFile",
			baseDir:  base,
			input:    "a.txt",
			wantPath: filepath.Join(base, "a.txt"),
		},
		{
			name:     "NestedRelative",
			baseDir:  base,
			input:    "nested/dir/file.jpg",
			wantPath: filepath.Join(base, "nested", "dir", "file.jpg"),
		},
		{
			name:    "EmptyName",
			baseDir: base,
			input:   "",
			wantErr: true,
		},
		{
			name:    "DotDotTraversal",
			baseDir: base,
			input:   "../outside.txt",
			wantErr: true,
		},
		{
			name:     "MixedSeparatorsTraversal",
			baseDir:  base,
			input:    `dir\..\inside.txt`,
			wantPath: filepath.Join(base, "inside.txt"),
		},
		{
			name:     "ContainsParentInMiddle",
			baseDir:  base,
			input:    "dir/../inside.txt",
			wantPath: filepath.Join(base, "inside.txt"),
		},
		{
			name:    "AbsoluteUnix",
			baseDir: base,
			input:   "/outside/target.txt",
			wantErr: true,
		},
		{
			name:    "RootedInteriorTraversal",
			baseDir: base,
			// A remote WebDAV href becomes relative once its leading slash is
			// stripped; interior ".." segments must still be rejected.
			input:   "sub/../../../../outside/target.jpg",
			wantErr: true,
		},
		{
			name:    "WindowsVolumeForwardSlash",
			baseDir: base,
			input:   "C:/outside/target.txt",
			wantErr: true,
		},
		{
			name:    "WindowsVolumeBackslash",
			baseDir: base,
			input:   `D:\outside\target.txt`,
			wantErr: true,
		},
		{
			name:     "CleansInsideBase",
			baseDir:  base,
			input:    "sub/../ok.txt",
			wantPath: filepath.Join(base, "ok.txt"),
		},
		{
			name:     "RepeatedSeparators",
			baseDir:  base,
			input:    "dir//file.txt",
			wantPath: filepath.Join(base, "dir", "file.txt"),
		},
		{
			name:    "VolumeNameOnly",
			baseDir: base,
			input:   "C:",
			wantErr: true,
		},
		{
			name:    "RootedBackslashUnix",
			baseDir: base,
			input:   `\outside.txt`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SafeJoin(tc.baseDir, tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got none (path=%q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantPath {
				t.Fatalf("unexpected path: got %q want %q", got, tc.wantPath)
			}
		})
	}
}
