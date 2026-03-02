package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// assertExitCode asserts that an error is a CLI exit error with the expected code.
func assertExitCode(t *testing.T, err error, code int) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected exit code %d, got nil error", code)
	}

	ec, ok := err.(cli.ExitCoder)
	if !ok {
		t.Fatalf("expected ExitCoder, got %T", err)
	}

	assert.Equal(t, code, ec.ExitCode())
}

func TestIndexCommandRejectsInvalidSubfolderPath(t *testing.T) {
	output, err := RunWithTestContext(IndexCommand, []string{"index", ".."})
	assertExitCode(t, err, 2)

	assert.Contains(t, err.Error(), "invalid subfolder path")
	assert.NotContains(t, output, "indexing originals in")
}

func TestImportCopyCommandRejectsInvalidDestinationPath(t *testing.T) {
	tests := []struct {
		Name string
		Cmd  *cli.Command
		Args []string
	}{
		{Name: "Import", Cmd: ImportCommand, Args: []string{"mv", "--dest", ".."}},
		{Name: "Copy", Cmd: CopyCommand, Args: []string{"cp", "--dest", ".."}},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			output, err := RunWithTestContext(tt.Cmd, tt.Args)
			assertExitCode(t, err, 2)

			assert.Contains(t, err.Error(), "invalid destination path")
			assert.NotContains(t, output, "media files from")
		})
	}
}

func TestImportCopyCommandRejectsOriginalsAsSourcePath(t *testing.T) {
	conf := get.Config()
	if conf == nil {
		t.Fatal("config is nil")
	}

	tests := []struct {
		Name string
		Cmd  *cli.Command
		Args []string
	}{
		{Name: "Import", Cmd: ImportCommand, Args: []string{"mv", conf.OriginalsPath()}},
		{Name: "Copy", Cmd: CopyCommand, Args: []string{"cp", conf.OriginalsPath()}},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			output, err := RunWithTestContext(tt.Cmd, tt.Args)
			assertExitCode(t, err, 2)

			assert.Contains(t, err.Error(), "source path is identical with originals")
			assert.NotContains(t, output, "media files from")
		})
	}
}
