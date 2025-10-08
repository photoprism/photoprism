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

	// Duplicate column names collide to the same key; last wins
	rows = [][]string{{"x", "y"}}
	cols = []string{"A-A", "A A"}
	objs = RowsToObjects(rows, cols)
	assert.Equal(t, map[string]string{"a_a": "y"}, objs[0])
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
