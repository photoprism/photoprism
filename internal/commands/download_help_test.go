package commands

import (
	"testing"
)

func TestDownloadCommand_HelpFlagsAndArgs(t *testing.T) {
	if DownloadCommand.ArgsUsage != "[url]..." {
		t.Fatalf("ArgsUsage mismatch: %q", DownloadCommand.ArgsUsage)
	}
	// Verify new flags are present by name
	want := map[string]bool{
		"cookies":    false,
		"add-header": false,
		"dl-method":  false,
		"file-remux": false,
	}
	for _, f := range DownloadCommand.Flags {
		name := f.Names()[0]
		if _, ok := want[name]; ok {
			want[name] = true
		}
	}
	for k, ok := range want {
		if !ok {
			t.Fatalf("missing flag: %s", k)
		}
	}
}
