package main

import (
	"os"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// TestGenerateLabelRules verifies aliases and rejects ambiguous sources before writing output.
func TestGenerateLabelRules(t *testing.T) {
	t.Run("DirectAlias", func(t *testing.T) {
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("rules.yml", []byte("fashion:\n  label: portrait\n  threshold: 0.5\ncardigan:\n  see: fashion\n"), fs.ModeFile))
		main()
		data, err := os.ReadFile("rules.go")
		require.NoError(t, err)
		require.Regexp(t, `(?s)"cardigan":\s*\{\s*Label:\s*"portrait",\s*Threshold:\s*0\.500000,`, string(data))
	})
	t.Run("DuplicateLabel", func(t *testing.T) {
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("rules.yml", []byte("cardigan:\n  label: portrait\ncardigan:\n  label: dog\n"), fs.ModeFile))
		require.NoError(t, os.WriteFile("rules.go", []byte("unchanged"), fs.ModeFile))
		require.Panics(t, main)
		data, err := os.ReadFile("rules.go")
		require.NoError(t, err)
		require.Equal(t, "unchanged", string(data))
	})
	t.Run("DuplicateField", func(t *testing.T) {
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("rules.yml", []byte("dog:\n  threshold: 0.2\n  threshold: 0.5\n"), fs.ModeFile))
		require.Panics(t, main)
		require.NoFileExists(t, "rules.go")
	})
	t.Run("UnknownField", func(t *testing.T) {
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("rules.yml", []byte("dog:\n  threhsold: 0.2\n"), fs.ModeFile))
		require.Panics(t, main)
		require.NoFileExists(t, "rules.go")
	})
	t.Run("MissingAlias", func(t *testing.T) {
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("rules.yml", []byte("dog:\n  see: missing\n"), fs.ModeFile))
		require.PanicsWithValue(t, "missing label: missing", main)
		require.NoFileExists(t, "rules.go")
	})
	t.Run("IndirectAlias", func(t *testing.T) {
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("rules.yml", []byte("animal:\n  label: animal\ndog:\n  see: animal\ncanine:\n  see: dog\n"), fs.ModeFile))
		require.PanicsWithValue(t, "classify: alias canine must reference a direct rule", main)
		require.NoFileExists(t, "rules.go")
	})
	t.Run("UppercaseLabel", func(t *testing.T) {
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("rules.yml", []byte("Dog:\n  label: dog\n"), fs.ModeFile))
		require.PanicsWithValue(t, "classify: Dog must be lowercase", main)
		require.NoFileExists(t, "rules.go")
	})
	t.Run("TemplateError", func(t *testing.T) {
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("rules.yml", []byte("dog:\n  label: dog\n"), fs.ModeFile))
		original := packageTemplate
		t.Cleanup(func() { packageTemplate = original })
		packageTemplate = template.Must(template.New("invalid").Parse("{{.MissingField}}"))
		require.Panics(t, main)
	})
}
