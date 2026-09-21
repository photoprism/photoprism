/*
Command check-log-rendering inventories ordinary logger arguments for rendering review.

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.
*/
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/photoprism/photoprism/pkg/fs"
)

const modulePath = "github.com/photoprism/photoprism/"

// finding identifies an argument requiring rendering review.
type finding struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Call     string `json:"call"`
	Argument int    `json:"argument"`
	Rule     string `json:"rule"`
	Line     int    `json:"-"`
}

// entry records the multiplicity of one review identity.
type entry struct {
	Finding finding `json:"finding"`
	Count   int     `json:"count"`
}

// baseline records the versioned set of reviewed call arguments.
type baseline struct {
	Version int     `json:"version"`
	Entries []entry `json:"entries"`
}

// source holds a parsed source file and its import names.
type source struct {
	path    string
	file    *ast.File
	imports map[string]string
}

// main runs the command with the process streams.
func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run inventories calls and optionally compares or records an explicit baseline.
func run(args []string, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("check-log-rendering", flag.ContinueOnError)
	flags.SetOutput(errOut)
	list := flags.Bool("list", false, "print every review candidate")
	update := flags.Bool("update", false, "record candidates in the explicit baseline file")
	baselinePath := flags.String("baseline", "", "optional per-call baseline file")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if *update && *baselinePath == "" {
		fmt.Fprintln(errOut, "check-log-rendering: -update requires -baseline")
		return 2
	}
	roots := flags.Args()
	optional := len(roots) == 0
	if optional {
		roots = []string{"internal", "pkg", "cmd", "plus/internal", "pro/internal", "portal/internal"}
	}
	findings, err := scan(roots, optional)
	if err != nil {
		fmt.Fprintln(errOut, "check-log-rendering:", err)
		return 2
	}
	current := counts(findings)
	if *list {
		for _, f := range findings {
			printFinding(out, f)
		}
	}
	if *update {
		if err = writeBaseline(*baselinePath, current); err != nil {
			fmt.Fprintln(errOut, "check-log-rendering:", err)
			return 2
		}
		fmt.Fprintf(out, "Recorded %d rendering review candidate(s).\n", len(findings))
		return 0
	}
	if *baselinePath != "" {
		recorded, readErr := readBaseline(*baselinePath)
		if readErr != nil {
			fmt.Fprintln(errOut, "check-log-rendering:", readErr)
			return 2
		}
		added := additions(current, recorded)
		if len(added) > 0 {
			if !*list {
				for _, f := range findings {
					key := f
					key.Line = 0
					if added[key] > 0 {
						printFinding(out, f)
					}
				}
			}
			fmt.Fprintf(errOut, "check-log-rendering: %d new rendering review identity/identities.\n", len(added))
			return 1
		}
	}
	fmt.Fprintf(out, "Rendering inventory: %d candidate(s); this is a scoped review aid, not a log-safety verdict.\n", len(findings))
	return 0
}

// printFinding writes a source position without printing the source expression.
func printFinding(out io.Writer, f finding) {
	fmt.Fprintf(out, "%s:%d: %s argument %d: %s review\n", f.File, f.Line, f.Function, f.Argument, f.Rule)
}

// scan parses existing roots and resolves package logger declarations before inspecting calls.
func scan(roots []string, optional bool) ([]finding, error) {
	fset := token.NewFileSet()
	files := map[string]source{}
	for _, root := range roots {
		if _, err := os.Stat(root); err != nil {
			if optional && os.IsNotExist(err) && (root == "plus/internal" || root == "pro/internal" || root == "portal/internal") {
				continue
			}
			return nil, err
		}
		info, err := os.Lstat(root)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("source root must not be a symbolic link: %s", root)
		}
		err = filepath.WalkDir(root, func(path string, d iofs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				if filepath.Clean(path) != filepath.Clean(root) && (strings.HasPrefix(d.Name(), ".") || d.Name() == "vendor" || d.Name() == "testdata" || strings.HasPrefix(d.Name(), "zz")) {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			path = filepath.Clean(path)
			if _, ok := files[path]; ok {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf("source input must be a regular file: %s", path)
			}
			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			imports := map[string]string{}
			for _, imp := range file.Imports {
				value, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					return err
				}
				name := filepath.Base(value)
				if imp.Name != nil {
					name = imp.Name.Name
				}
				imports[name] = value
			}
			files[path] = source{path, file, imports}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	loggers := map[string]map[string]bool{}
	for _, src := range files {
		key := filepath.Dir(src.path) + "/" + src.file.Name.Name
		if loggers[key] == nil {
			loggers[key] = map[string]bool{}
		}
		for _, decl := range src.file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				value, ok := spec.(*ast.ValueSpec)
				if !ok || len(value.Names) != len(value.Values) {
					continue
				}
				for i, name := range value.Names {
					if qualified(value.Values[i], src.imports) == modulePath+"internal/event.Log" {
						loggers[key][name.Name] = true
					}
				}
			}
		}
	}
	var found []finding
	for _, src := range files {
		aliases := loggers[filepath.Dir(src.path)+"/"+src.file.Name.Name]
		for _, decl := range src.file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			name := fn.Name.Name
			if fn.Recv != nil {
				name = expression(fset, fn.Recv.List[0].Type) + "." + name
			}
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok || !ordinary(call, src.imports, aliases) {
					return true
				}
				callText := ""
				for i, arg := range call.Args {
					if rendered(arg, src.imports) {
						continue
					}
					rule := "rendering"
					if credential(arg, src.imports) {
						rule = "redaction"
					}
					if callText == "" {
						callText = expression(fset, call)
					}
					found = append(found, finding{File: filepath.ToSlash(src.path), Function: name, Call: callText, Argument: i + 1, Rule: rule, Line: fset.Position(arg.Pos()).Line})
				}
				return true
			})
		}
	}
	sort.Slice(found, func(i, j int) bool {
		if found[i].File != found[j].File {
			return found[i].File < found[j].File
		}
		if found[i].Line != found[j].Line {
			return found[i].Line < found[j].Line
		}
		return found[i].Argument < found[j].Argument
	})
	return found, nil
}

// unparen returns the expression inside any enclosing parentheses.
func unparen(expr ast.Expr) ast.Expr {
	for {
		p, ok := expr.(*ast.ParenExpr)
		if !ok {
			return expr
		}
		expr = p.X
	}
}

// qualified resolves an imported selector without accepting locally shadowed import names.
func qualified(expr ast.Expr, imports map[string]string) string {
	sel, ok := unparen(expr).(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	name, ok := unparen(sel.X).(*ast.Ident)
	if !ok || name.Obj != nil || imports[name.Name] == "" {
		return ""
	}
	return imports[name.Name] + "." + sel.Sel.Name
}

// ordinary recognizes direct ordinary logger calls and unshadowed package logger bindings.
func ordinary(call *ast.CallExpr, imports map[string]string, aliases map[string]bool) bool {
	sel, ok := unparen(call.Fun).(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch sel.Sel.Name {
	case "Print", "Printf", "Println", "Info", "Infof", "Infoln", "Warn", "Warnf", "Warnln", "Warning", "Warningf", "Warningln", "Error", "Errorf", "Errorln", "Fatal", "Fatalf", "Fatalln", "Panic", "Panicf", "Panicln", "Log", "Logf", "Logln":
	default:
		return false
	}
	if qualified(sel.X, imports) == modulePath+"internal/event.Log" {
		return true
	}
	name, ok := unparen(sel.X).(*ast.Ident)
	if !ok || !aliases[name.Name] {
		return false
	}
	if name.Obj == nil {
		return true
	}
	_, ok = name.Obj.Decl.(*ast.ValueSpec)
	// Package declarations have no enclosing function in their object data; compare their initializer.
	if !ok {
		return false
	}
	value := name.Obj.Decl.(*ast.ValueSpec)
	for i, declared := range value.Names {
		if declared.Name == name.Name && len(value.Names) == len(value.Values) {
			return qualified(value.Values[i], imports) == modulePath+"internal/event.Log"
		}
	}
	return false
}

// rendered recognizes literal composition and the documented whole-value renderers.
func rendered(expr ast.Expr, imports map[string]string) bool {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return true
	case *ast.ParenExpr:
		return rendered(e.X, imports)
	case *ast.BinaryExpr:
		return rendered(e.X, imports) && rendered(e.Y, imports)
	case *ast.Ident:
		return false
	case *ast.CallExpr:
		name := qualified(e.Fun, imports)
		switch name {
		case "strings.ToLower", "strings.ToUpper", "strings.TrimSpace":
			return len(e.Args) == 1 && rendered(e.Args[0], imports)
		case "fmt.Sprintf", "fmt.Sprint":
			for _, arg := range e.Args {
				if !rendered(arg, imports) {
					return false
				}
			}
			return true
		case modulePath + "pkg/clean.Log", modulePath + "pkg/clean.LogQuote", modulePath + "pkg/clean.LogLower", modulePath + "pkg/clean.LogNames", modulePath + "pkg/clean.LogUri", modulePath + "pkg/clean.Error", modulePath + "pkg/log/status.Error":
			return !credential(expr, imports)
		}
	}
	return false
}

// credential identifies credential-shaped names outside explicit whole-value redactors.
func credential(expr ast.Expr, imports map[string]string) bool {
	found := false
	ast.Inspect(expr, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			switch qualified(call.Fun, imports) {
			case modulePath + "pkg/clean.UriRedacted", modulePath + "pkg/clean.UriRedactedText", modulePath + "pkg/clean.LogUri", modulePath + "pkg/clean.FileNameRedacted":
				return false
			}
		}
		ident, ok := node.(*ast.Ident)
		if !ok {
			return true
		}
		name := strings.ToLower(ident.Name)
		for _, part := range []string{"password", "passwd", "secret", "token", "authorization", "cookie", "credential", "acckey", "dsn", "proxy", "uri", "url"} {
			if strings.Contains(name, part) {
				found = true
			}
		}
		return true
	})
	return found
}

// expression renders a call for its stable baseline identity.
func expression(fset *token.FileSet, expr ast.Expr) string {
	var b bytes.Buffer
	if err := printer.Fprint(&b, fset, expr); err != nil {
		return "?"
	}
	return b.String()
}

// counts groups candidates without their movable source line numbers.
func counts(findings []finding) map[finding]int {
	result := map[finding]int{}
	for _, f := range findings {
		f.Line = 0
		result[f]++
	}
	return result
}

// additions returns per-identity increases without netting out unrelated removals.
func additions(current, recorded map[finding]int) map[finding]int {
	result := map[finding]int{}
	for f, n := range current {
		if n > recorded[f] {
			result[f] = n - recorded[f]
		}
	}
	return result
}

// readBaseline validates a versioned baseline and rejects duplicate identities.
func readBaseline(path string) (map[finding]int, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- baseline path is explicitly selected by the local operator.
	if err != nil {
		return nil, err
	}
	var b baseline
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&b); err != nil {
		return nil, err
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("expected one baseline document")
	}
	if b.Version != 1 {
		return nil, fmt.Errorf("unsupported baseline version %d", b.Version)
	}
	result := map[finding]int{}
	for _, e := range b.Entries {
		if e.Count < 1 || e.Finding.File == "" || e.Finding.Function == "" || e.Finding.Call == "" || e.Finding.Argument < 1 || (e.Finding.Rule != "rendering" && e.Finding.Rule != "redaction") || result[e.Finding] != 0 {
			return nil, fmt.Errorf("invalid baseline entry")
		}
		result[e.Finding] = e.Count
	}
	return result, nil
}

// writeBaseline records a deterministic per-identity inventory.
func writeBaseline(path string, current map[finding]int) error {
	b := baseline{Version: 1, Entries: make([]entry, 0, len(current))}
	for f, n := range current {
		f.Line = 0
		b.Entries = append(b.Entries, entry{f, n})
	}
	sort.Slice(b.Entries, func(i, j int) bool {
		a, _ := json.Marshal(b.Entries[i].Finding)
		c, _ := json.Marshal(b.Entries[j].Finding)
		return string(a) < string(c)
	})
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".log-rendering-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if err = file.Chmod(fs.ModeSecretFile); err != nil {
		_ = file.Close()
		return err
	}
	if _, err = file.Write(append(data, '\n')); err != nil {
		_ = file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
