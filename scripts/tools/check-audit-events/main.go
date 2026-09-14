/*
Command check-audit-events reports event calls that build their message from a value instead of
passing it as an argument.

The first argument of event.Audit* and event.System* is joined into the format string those calls
render, so a value among the segments is interpreted rather than printed. The convention is that
segments are structure and values are arguments. go vet cannot check it, because the format string
is assembled inside event.Format rather than passed at the call site.

A local variable holding a constant reads the same here as one that does not, so the remaining call
sites are a worklist rather than a verdict. They are held in baseline.txt, one line per
call, and the check fails on any call that is not recorded there - so fixing one site and adding
another does not net out. The rule set is
ours, so unlike a third-party linter the count moves only when this repository does. Run with
-update after fixing sites to record the lower numbers, and -list to print every finding.

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.
*/
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	iofs "io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/photoprism/photoprism/pkg/fs"
)

// defaultRoots are searched when none are named. The edition directories are separate
// repositories that a clone may not have, so a missing one is skipped rather than reported.
var defaultRoots = []string{"internal", "pkg", "plus/internal", "pro/internal", "portal/internal"}

// allowedPrefixes are the segment expressions that cannot carry a format verb or a field
// separator: the status, authn, acl and i18n exports are constants, except status.Error, which is
// safe for a different reason - clean.Error maps both. A future export of these packages that is
// neither would be waved through, so keep that property when adding one.
var allowedPrefixes = []string{"status.", "authn.", "acl.", "string(acl.", "i18n."}

// allowedExact are the client-address helpers, which resolve through clean.IP.
var allowedExact = []string{"clientIp", "clientIP", "ClientIP(c)", "api.ClientIP(c)", "header.ClientIP(c)", "baseapi.ClientIP(c)"}

// allowedCalls are the error renderers, which belong in the segment list rather than the argument
// list: they map the format verb and the field separator, which the value helpers do not. Matched
// after one wrapping call is removed, so strings.ToLower(clean.Error(err)) passes and a
// concatenation that merely contains one does not.
var allowedCalls = []string{"clean.Error", "clean.ErrorFull", "status.Error"}

func main() {
	update, list := false, false
	roots := make([]string, 0, len(os.Args))

	for _, arg := range os.Args[1:] {
		switch arg {
		case "-update":
			update = true
		case "-list":
			list = true
		default:
			roots = append(roots, arg)
		}
	}

	if len(roots) == 0 {
		roots = defaultRoots
	}

	findings, err := check(roots)

	if err != nil {
		fmt.Fprintf(os.Stderr, "check-audit-events: %s\n", err)
		os.Exit(2)
	}

	if list {
		for _, f := range findings {
			fmt.Println(f.String())
		}
	}

	keys := keysOf(findings)

	if update {
		if err = writeBaseline(keys); err != nil {
			fmt.Fprintf(os.Stderr, "check-audit-events: %s\n", err)
			os.Exit(2)
		}

		fmt.Printf("Recorded %d audit event call(s) to review in %s.\n", len(findings), baselineFile)

		return
	}

	baseline, err := readBaseline()

	if err != nil {
		fmt.Fprintf(os.Stderr, "check-audit-events: %s\n", err)
		os.Exit(2)
	}

	if added := exceeded(keys, baseline); len(added) > 0 {
		for _, f := range findings {
			if added[f.key()] {
				fmt.Println(f.String())
			}
		}

		fmt.Fprintf(os.Stderr, "\ncheck-audit-events: %d call(s) are not recorded in %s.\n"+
			"Pass the value as an argument behind a \"%%s\" segment, or run with -update if it is meant to stay.\n",
			len(added), baselineFile)
		os.Exit(1)
	}

	fmt.Printf("Audit event calls checked, %d to review (no increase over %s).\n", len(findings), baselineFile)
}

// baselineFile records how many calls each file is known to have, relative to this source.
const baselineFile = "scripts/tools/check-audit-events/baseline.txt"

// finding is one segment that holds a value.
type finding struct {
	File     string
	Position string
	Call     string
	Segment  string
}

// String renders a finding the way a compiler reports one.
func (f finding) String() string {
	return fmt.Sprintf("%s: %s carries %s in its segment list", f.Position, f.Call, f.Segment)
}

// check walks every root that exists and returns the findings, sorted by position.
func check(roots []string) (findings []finding, err error) {
	for _, root := range roots {
		if _, statErr := os.Stat(root); statErr != nil {
			continue
		}

		found, checkErr := checkRoot(root)

		if checkErr != nil {
			return nil, checkErr
		}

		findings = append(findings, found...)
	}

	sort.Slice(findings, func(i, j int) bool { return findings[i].Position < findings[j].Position })

	return findings, nil
}

// key identifies a call without its line, so moving code does not churn the baseline while a new
// call in a file that already has some still shows up.
func (f finding) key() string {
	return f.File + "\t" + f.Call + "\t" + f.Segment
}

// keysOf counts the findings by key, since one file may hold the same call more than once.
func keysOf(findings []finding) map[string]int {
	keys := make(map[string]int)

	for _, f := range findings {
		keys[f.key()]++
	}

	return keys
}

// readBaseline returns the recorded count per key, treating an absent file as no entries.
func readBaseline() (map[string]int, error) {
	b, err := os.ReadFile(baselineFile)

	if os.IsNotExist(err) {
		return map[string]int{}, nil
	} else if err != nil {
		return nil, err
	}

	counts := make(map[string]int)

	for _, line := range strings.Split(string(b), "\n") {
		if line = strings.TrimSpace(line); line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		n, key, found := strings.Cut(line, "\t")

		if !found {
			return nil, fmt.Errorf("%s: cannot read %q", baselineFile, line)
		}

		c, convErr := strconv.Atoi(n)

		if convErr != nil {
			return nil, fmt.Errorf("%s: cannot read %q", baselineFile, line)
		}

		counts[key] = c
	}

	return counts, nil
}

// writeBaseline records the current calls, sorted so a diff shows only what moved.
func writeBaseline(keys map[string]int) error {
	sorted := make([]string, 0, len(keys))

	for key := range keys {
		sorted = append(sorted, key)
	}

	sort.Strings(sorted)

	var b strings.Builder

	b.WriteString("# Calls that build an event message from a value: count, file, call, segment.\n")
	b.WriteString("# Regenerate with: go run ./scripts/tools/check-audit-events -update\n")

	for _, key := range sorted {
		fmt.Fprintf(&b, "%d\t%s\n", keys[key], key)
	}

	return os.WriteFile(baselineFile, []byte(b.String()), fs.ModeFile)
}

// exceeded returns the calls that occur more often than the baseline records.
func exceeded(keys, baseline map[string]int) map[string]bool {
	over := make(map[string]bool)

	for key, n := range keys {
		if n > baseline[key] {
			over[key] = true
		}
	}

	return over
}

// checkRoot walks one directory and returns a finding for every offending call.
func checkRoot(root string) (findings []finding, err error) {
	fset := token.NewFileSet()

	err = filepath.WalkDir(root, func(path string, d iofs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			// Scratch packages are gitignored, so a working tree must not fail the check.
			if strings.HasPrefix(d.Name(), "zz") {
				return filepath.SkipDir
			}

			return nil
		} else if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, nil, 0)

		if parseErr != nil {
			return fmt.Errorf("%s: %w", path, parseErr)
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)

			if !ok || !isEventCall(call) || len(call.Args) < 2 {
				return true
			}

			lit, ok := call.Args[0].(*ast.CompositeLit)

			if !ok {
				return true
			}

			system := isSystemCall(call)

			for i, elt := range lit.Elts {
				// System renders the first segment outside Format, so a value there is not
				// interpreted and a verb there is a defect rather than the remedy.
				if system && i == 0 {
					continue
				}

				if isLiteral(elt) || allowed(fset, elt) {
					continue
				}

				findings = append(findings, finding{
					File:     filepath.ToSlash(path),
					Position: fset.Position(elt.Pos()).String(),
					Call:     expr(fset, call.Fun),
					Segment:  expr(fset, elt),
				})
			}

			return true
		})

		return nil
	})

	return findings, err
}

// isLiteral reports whether a segment is written out in the source, including the concatenated
// form a segment too long for one line takes.
func isLiteral(e ast.Expr) bool {
	switch t := e.(type) {
	case *ast.BasicLit:
		return true
	case *ast.BinaryExpr:
		return t.Op == token.ADD && isLiteral(t.X) && isLiteral(t.Y)
	}

	return false
}

// isSystemCall reports whether the call is one of the System helpers, whose first segment is
// rendered as a plain category label rather than through Format.
func isSystemCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)

	return ok && strings.HasPrefix(sel.Sel.Name, "System")
}

// isEventCall reports whether the call is one of the event helpers that render a segment list.
func isEventCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)

	if !ok {
		return false
	}

	pkg, ok := sel.X.(*ast.Ident)

	if !ok || pkg.Name != "event" {
		return false
	}

	return strings.HasPrefix(sel.Sel.Name, "Audit") || strings.HasPrefix(sel.Sel.Name, "System")
}

// allowed reports whether a non-literal segment is one of the accepted constants or helpers.
func allowed(fset *token.FileSet, e ast.Expr) bool {
	s := expr(fset, e)

	for _, p := range allowedPrefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}

	for _, a := range allowedExact {
		if s == a {
			return true
		}
	}

	return isAllowedCall(e) || isAllowedCall(unwrapCall(e))
}

// unwrapCall returns the single argument of a one-argument call, so a renderer wrapped in
// something like strings.ToLower is recognized while a concatenation is not.
func unwrapCall(e ast.Expr) ast.Expr {
	if call, ok := e.(*ast.CallExpr); ok && len(call.Args) == 1 {
		return call.Args[0]
	}

	return nil
}

// isAllowedCall reports whether an expression is a call to one of the error renderers.
func isAllowedCall(e ast.Expr) bool {
	call, ok := e.(*ast.CallExpr)

	if !ok {
		return false
	}

	name := expr(token.NewFileSet(), call.Fun)

	for _, a := range allowedCalls {
		if name == a {
			return true
		}
	}

	return false
}

// expr renders an expression the way it is written in the source.
func expr(fset *token.FileSet, e ast.Expr) string {
	var b bytes.Buffer

	if err := printer.Fprint(&b, fset, e); err != nil {
		return "?"
	}

	return b.String()
}
