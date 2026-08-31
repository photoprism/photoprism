package report

import (
	"encoding/json"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestRowsToObjectsAndJSONExport(t *testing.T) {
	rows := [][]string{
		{"Alice", "30", "extra"}, // extra value should be ignored
		{"Bob"},                  // missing values default to ""
		{"Carol", "27"},
	}
	cols := []string{"First Name", "Age", "!@#$%"}

	objs := RowsToObjects(rows, cols)
	if assert.Len(t, objs, 3) {
		assert.Equal(t, map[string]string{"first_name": "Alice", "age": "30", "col": "extra"}, objs[0])
		assert.Equal(t, map[string]string{"first_name": "Bob", "age": "", "col": ""}, objs[1])
		assert.Equal(t, map[string]string{"first_name": "Carol", "age": "27", "col": ""}, objs[2])
	}

	// JSONExport should marshal the same shape
	s, err := JSONExport(rows, cols)
	assert.NoError(t, err)

	var back []map[string]string
	assert.NoError(t, json.Unmarshal([]byte(s), &back))
	assert.Equal(t, objs, back)

	// Columns that canonicalize to one key are suffixed rather than collapsed, so a table holding
	// two of the same name - a marker source beside a subject source - exports both.
	rows = [][]string{{"x", "y"}}
	cols = []string{"A-A", "A A"}
	objs = RowsToObjects(rows, cols)
	assert.Equal(t, map[string]string{"a_a": "x", "a_a_2": "y"}, objs[0])
}

func TestCliFormatStrict(t *testing.T) {
	// Helper to build a cli.Context with flags
	newCtx := func(setFlags func(ctx *cli.Context)) *cli.Context {
		app := &cli.App{Flags: CliFlags}
		fs := flag.NewFlagSet("test", 0)
		// Register app flags into the stdlib flagset
		for _, fl := range app.Flags {
			_ = fl.Apply(fs)
		}
		ctx := cli.NewContext(app, fs, nil)
		if setFlags != nil {
			setFlags(ctx)
		}
		return ctx
	}

	// Default
	fmt, err := CliFormatStrict(newCtx(nil))
	assert.NoError(t, err)
	assert.Equal(t, Format(Default), fmt)

	// Individual flags
	fmt, err = CliFormatStrict(newCtx(func(ctx *cli.Context) { _ = ctx.Set("json", "true") }))
	assert.NoError(t, err)
	assert.Equal(t, Format(JSON), fmt)

	fmt, err = CliFormatStrict(newCtx(func(ctx *cli.Context) { _ = ctx.Set("md", "true") }))
	assert.NoError(t, err)
	assert.Equal(t, Format(Markdown), fmt)

	fmt, err = CliFormatStrict(newCtx(func(ctx *cli.Context) { _ = ctx.Set("csv", "true") }))
	assert.NoError(t, err)
	assert.Equal(t, Format(CSV), fmt)

	fmt, err = CliFormatStrict(newCtx(func(ctx *cli.Context) { _ = ctx.Set("tsv", "true") }))
	assert.NoError(t, err)
	assert.Equal(t, Format(TSV), fmt)

	// Multiple flags → usage error with exit code 2
	_, err = CliFormatStrict(newCtx(func(ctx *cli.Context) {
		_ = ctx.Set("json", "true")
		_ = ctx.Set("csv", "true")
	}))
	if assert.Error(t, err) {
		if exit, ok := err.(cli.ExitCoder); ok {
			assert.Equal(t, 2, exit.ExitCode())
		} else {
			t.Fatalf("expected cli.ExitCoder, got %T", err)
		}
	}
}

func TestUniqueKeys(t *testing.T) {
	t.Run("Distinct", func(t *testing.T) {
		assert.Equal(t, []string{"a", "b"}, uniqueKeys([]string{"A", "B"}))
	})
	t.Run("Repeated", func(t *testing.T) {
		// A marker source beside a subject source, both titled "Src" for the reader.
		assert.Equal(t, []string{"src", "src_2", "src_3"}, uniqueKeys([]string{"Src", "Src", "Src"}))
	})
}

func TestRowsToObjectsKeys(t *testing.T) {
	rows := [][]string{{"x", "y"}}

	t.Run("GivenKeys", func(t *testing.T) {
		objs := RowsToObjectsKeys(rows, []string{"Src", "Src"}, []string{"marker_src", "subj_src"})
		assert.Equal(t, map[string]string{"marker_src": "x", "subj_src": "y"}, objs[0])
	})
	t.Run("FallsBackToTheHeadings", func(t *testing.T) {
		objs := RowsToObjectsKeys(rows, []string{"A", "B"}, nil)
		assert.Equal(t, map[string]string{"a": "x", "b": "y"}, objs[0])
	})
	t.Run("PartialKeysFallBackForTheRest", func(t *testing.T) {
		// Short rather than wrong: the named ones are used and the rest keep a heading-derived key.
		objs := RowsToObjectsKeys(rows, []string{"A", "B"}, []string{"first"})
		assert.Equal(t, map[string]string{"first": "x", "b": "y"}, objs[0])
	})
}
