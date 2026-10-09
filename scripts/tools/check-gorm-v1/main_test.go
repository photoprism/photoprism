package main

import (
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScan(t *testing.T) {
	t.Run("Sample", func(t *testing.T) {
		findings, err := scan([]string{"./testdata/sample"}, []string{"testdata"})
		require.NoError(t, err)

		counts := make(map[string]int)
		details := make(map[string][]string)
		files := make(map[string]int)

		for _, f := range findings {
			files[f.File]++
			counts[f.Pattern]++
			details[f.Pattern] = append(details[f.Pattern], f.Detail)
		}

		// The test file counts once although it is loaded with each package variant, and the
		// generated file not at all.
		assert.Equal(t, map[string]int{"testdata/sample/sample.go": 16, "testdata/sample/sample_test.go": 1}, files)
		assert.Equal(t, map[string]int{
			PatternRecordNotFound: 3,
			PatternQueryExpr:      1,
			PatternSubQuery:       1,
			PatternGetName:        1,
			PatternScope:          2,
			PatternDeletedAt:      2,
			PatternCount:          3,
			PatternBoolLiteral:    4,
		}, counts)
		assert.Equal(t, []string{"Count(*int)", "Count method value", "Count(*uint)"}, details[PatternCount])
		assert.Equal(t, []string{"photo_private = 0", "file_missing=1", "`photo_private` = 1", `"photos"."file_missing"<>0`}, details[PatternBoolLiteral])
	})
	t.Run("OutsideRoots", func(t *testing.T) {
		findings, err := scan([]string{"./testdata/sample"}, []string{"internal"})
		require.NoError(t, err)
		assert.Empty(t, findings)
	})
}

func TestCheck(t *testing.T) {
	t.Run("MissingRoots", func(t *testing.T) {
		findings, scanned, err := check([]string{"plus-missing/internal", "pro-missing/internal"})
		require.NoError(t, err)
		assert.Empty(t, findings)
		assert.Empty(t, scanned)
	})
}

func TestIdentPattern(t *testing.T) {
	gormTypes := types.NewPackage(gormPkg, "gorm")
	otherTypes := types.NewPackage("example.com/other", "other")
	timeTypes := types.NewPackage("time", "time")
	timeType := types.NewNamed(types.NewTypeName(token.NoPos, timeTypes, "Time", nil), types.NewStruct(nil, nil), nil)
	sig := types.NewSignatureType(nil, nil, nil, nil, nil, false)

	t.Run("Methods", func(t *testing.T) {
		assert.Equal(t, PatternRecordNotFound, identPattern(types.NewFunc(token.NoPos, gormTypes, "RecordNotFound", sig)))
		assert.Equal(t, PatternGetName, identPattern(types.NewFunc(token.NoPos, gormTypes, "GetName", sig)))
		assert.Equal(t, "", identPattern(types.NewFunc(token.NoPos, gormTypes, "Where", sig)))
		assert.Equal(t, "", identPattern(types.NewFunc(token.NoPos, otherTypes, "RecordNotFound", sig)))
	})
	t.Run("Scope", func(t *testing.T) {
		assert.Equal(t, PatternScope, identPattern(types.NewTypeName(token.NoPos, gormTypes, "Scope", nil)))
		assert.Equal(t, "", identPattern(types.NewTypeName(token.NoPos, otherTypes, "Scope", nil)))
	})
	t.Run("DeletedAt", func(t *testing.T) {
		assert.Equal(t, PatternDeletedAt, identPattern(types.NewField(token.NoPos, otherTypes, "DeletedAt", types.NewPointer(timeType), false)))
		assert.Equal(t, "", identPattern(types.NewField(token.NoPos, otherTypes, "DeletedAt", timeType, false)))
		assert.Equal(t, "", identPattern(types.NewVar(token.NoPos, otherTypes, "DeletedAt", types.NewPointer(timeType))))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "", identPattern(nil))
	})
}

func TestBoolColumn(t *testing.T) {
	pkg := types.NewPackage("example.com/entity", "entity")
	field := func(name string, typ types.Type) *types.Var {
		return types.NewField(token.NoPos, pkg, name, typ, false)
	}
	boolType := types.Typ[types.Bool]

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "photo_private", boolColumn(field("PhotoPrivate", boolType), `json:"Private"`))
		assert.Equal(t, "label_nsfw", boolColumn(field("LabelNSFW", boolType), ""))
	})
	t.Run("ColumnTag", func(t *testing.T) {
		assert.Equal(t, "webdav", boolColumn(field("WebDAV", boolType), `gorm:"column:webdav;" json:"WebDAV"`))
		assert.Equal(t, "file_hdr", boolColumn(field("FileHDR", boolType), `gorm:"index; column:file_hdr"`))
	})
	t.Run("Ignored", func(t *testing.T) {
		assert.Equal(t, "", boolColumn(field("Selected", boolType), `gorm:"-" json:"Selected"`))
		assert.Equal(t, "", boolColumn(field("count", boolType), ""))
		assert.Equal(t, "", boolColumn(field("PhotoCount", types.Typ[types.Int]), ""))
		assert.Equal(t, "", boolColumn(types.NewField(token.NoPos, pkg, "Embedded", boolType, true), ""))
	})
}

func TestBoolLiterals(t *testing.T) {
	re := boolColumnPattern([]string{"photo_private", "file_missing"})

	t.Run("Match", func(t *testing.T) {
		assert.Equal(t, []string{"photo_private = 0"}, boolLiterals(re, "photo_private = 0"))
		assert.Equal(t, []string{"p.photo_private<>1", "file_missing=0"}, boolLiterals(re, "(p.photo_private<>1,file_missing=0)"))
		assert.Equal(t, []string{"PHOTO_PRIVATE != 1"}, boolLiterals(re, "WHERE PHOTO_PRIVATE != 1 AND x = 1"))
		assert.Equal(t, []string{"`p`.`photo_private` = 0"}, boolLiterals(re, "`p`.`photo_private` = 0"))
	})
	t.Run("NoMatch", func(t *testing.T) {
		assert.Empty(t, boolLiterals(re, "photo_private = TRUE"))
		assert.Empty(t, boolLiterals(re, "photo_private = 10"))
		assert.Empty(t, boolLiterals(re, "photo_private = 1.5"))
		assert.Empty(t, boolLiterals(re, "xphoto_private = 1"))
		assert.Empty(t, boolLiterals(re, "photo_private_count = 1"))
		assert.Empty(t, boolLiterals(re, "photo_private = ?"))
	})
}

func TestBoolColumnPattern(t *testing.T) {
	t.Run("QuotesNames", func(t *testing.T) {
		re := boolColumnPattern([]string{"a.b"})
		assert.Empty(t, boolLiterals(re, "axb = 1"))
		assert.Equal(t, []string{"a.b = 1"}, boolLiterals(re, "a.b = 1"))
	})
	t.Run("Compiles", func(t *testing.T) {
		assert.IsType(t, &regexp.Regexp{}, boolColumnPattern(nil))
	})
}

func TestIsScratch(t *testing.T) {
	t.Run("Scratch", func(t *testing.T) {
		assert.True(t, isScratch("github.com/photoprism/photoprism/internal/zzscratch"))
		assert.True(t, isScratch("github.com/photoprism/photoprism/internal/zz/sub"))
	})
	t.Run("Package", func(t *testing.T) {
		assert.False(t, isScratch("github.com/photoprism/photoprism/internal/api"))
		assert.False(t, isScratch("github.com/photoprism/photoprism/internal/api/zz"))
		assert.False(t, isScratch("github.com/photoprism/photoprism/pkg/zzx"))
		assert.False(t, isScratch("github.com/photoprism/photoprism/pro/internal/zzx"))
	})
}

func TestBaseline(t *testing.T) {
	t.Run("RoundTrip", func(t *testing.T) {
		keys := map[string]int{
			"internal/entity/photo.go\tDeletedAt *time.Time": 12,
			"internal/api/a_test.go\tCount into non-int64":   2,
		}

		s := formatBaseline(keys)
		assert.Contains(t, s, "# Regenerate with: go run ./scripts/tools/check-gorm-v1 -update\n")
		assert.Contains(t, s, "2\tinternal/api/a_test.go\tCount into non-int64\n12\tinternal/entity/photo.go\tDeletedAt *time.Time\n")

		parsed, err := parseBaseline(s)
		require.NoError(t, err)
		assert.Equal(t, keys, parsed)
	})
	t.Run("Invalid", func(t *testing.T) {
		_, err := parseBaseline("internal/a.go\tRecordNotFound\n")
		assert.Error(t, err)
		_, err = parseBaseline("x\tinternal/a.go\tRecordNotFound\n")
		assert.Error(t, err)
	})
}

func TestExceeded(t *testing.T) {
	baseline := map[string]int{"a.go\tRecordNotFound": 2, "b.go\tQueryExpr": 1}

	t.Run("Increase", func(t *testing.T) {
		over := exceeded(map[string]int{"a.go\tRecordNotFound": 3, "b.go\tQueryExpr": 1, "c.go\tSubQuery": 1}, baseline)
		assert.Equal(t, map[string]bool{"a.go\tRecordNotFound": true, "c.go\tSubQuery": true}, over)
	})
	t.Run("Decrease", func(t *testing.T) {
		assert.Empty(t, exceeded(map[string]int{"a.go\tRecordNotFound": 1}, baseline))
	})
}

func TestMergeUnscanned(t *testing.T) {
	baseline := map[string]int{
		"internal/a.go\tRecordNotFound":    2,
		"pro/internal/b.go\tQueryExpr":     1,
		"portal/internal/c.go\tgorm.Scope": 3,
	}
	t.Run("EditionsMissing", func(t *testing.T) {
		merged := mergeUnscanned(map[string]int{"internal/a.go\tRecordNotFound": 1}, baseline, []string{"internal", "pkg"})

		assert.Equal(t, map[string]int{
			"internal/a.go\tRecordNotFound":    1,
			"pro/internal/b.go\tQueryExpr":     1,
			"portal/internal/c.go\tgorm.Scope": 3,
		}, merged)
	})
	t.Run("AllScanned", func(t *testing.T) {
		keys := map[string]int{"internal/a.go\tRecordNotFound": 1}

		assert.Equal(t, keys, mergeUnscanned(keys, baseline, []string{"internal", "pro/internal", "portal/internal"}))
	})
}

func TestUnderRoot(t *testing.T) {
	t.Run("Inside", func(t *testing.T) {
		assert.True(t, underRoot("internal/api/a.go", []string{"internal"}))
		assert.True(t, underRoot("pro/internal/api/b.go", []string{"internal", "pro/internal/"}))
		assert.True(t, underRoot("internal/api/a.go", []string{"./internal"}))
		assert.True(t, underRoot("internal/api/a.go", []string{"."}))
	})
	t.Run("Outside", func(t *testing.T) {
		assert.False(t, underRoot("pro/internal/api/b.go", []string{"internal", "pkg"}))
		assert.False(t, underRoot("internalx/a.go", []string{"internal"}))
	})
}

func TestBaselineSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	require.NoError(t, os.WriteFile(target, []byte("1\tinternal/a.go\tRecordNotFound\n"), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, filepath.Dir(baselineFile)), 0o700))
	require.NoError(t, os.Symlink(target, filepath.Join(dir, baselineFile)))
	t.Chdir(dir)

	t.Run("Read", func(t *testing.T) {
		_, err := readBaseline()
		assert.Error(t, err)
	})
	t.Run("Write", func(t *testing.T) {
		assert.Error(t, writeBaseline(map[string]int{"internal/b.go\tQueryExpr": 2}))

		b, err := os.ReadFile(target) //nolint:gosec // G304: test-owned temporary file.
		require.NoError(t, err)
		assert.Equal(t, "1\tinternal/a.go\tRecordNotFound\n", string(b))
	})
}
