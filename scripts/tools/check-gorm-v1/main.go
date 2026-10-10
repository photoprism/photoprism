/*
Command check-gorm-v1 counts code that depends on GORM v1 behavior, so that the sites to change when
switching to GORM v2 do not grow while both are in reach.

It type-checks the packages, including their tests, and reports uses of RecordNotFound,
IsRecordNotFoundError, QueryExpr, SubQuery, Dialect.GetName, NewScope and gorm.Scope, DeletedAt fields
of type *time.Time, Count destinations other than *int64, and string literals that compare a boolean
column with 0 or 1.
The sites are held in baseline.txt as a count per file and pattern, and the check fails when a group
holds more than it records. Each edition repository keeps the entries of its files in a baseline.txt of
its own, at the same path. Run with -update after removing sites to record the lower numbers, and
-list to print every finding. An update keeps the entries of directories it did not scan.

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.
*/
package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"golang.org/x/tools/go/packages"

	"github.com/photoprism/photoprism/pkg/fs"
)

// defaultRoots are searched when none are named. The edition directories are separate
// repositories that a clone may not have, so a missing one is skipped rather than reported.
var defaultRoots = []string{"internal", "pkg", "plus/internal", "pro/internal", "portal/internal"}

const (
	gormPkg    = "github.com/jinzhu/gorm"
	modulePath = "github.com/photoprism/photoprism"
	entityPkg  = modulePath + "/internal/entity"
)

// Patterns name the GORM v1 dependencies the check counts.
const (
	PatternRecordNotFound = "RecordNotFound"
	PatternQueryExpr      = "QueryExpr"
	PatternSubQuery       = "SubQuery"
	PatternGetName        = "Dialect.GetName"
	PatternScope          = "gorm.Scope"
	PatternDeletedAt      = "DeletedAt *time.Time"
	PatternCount          = "Count into non-int64"
	PatternBoolLiteral    = "bool column = 0/1"
)

// gormMethods maps the GORM v1 functions and methods the check counts to their pattern.
var gormMethods = map[string]string{
	"RecordNotFound":        PatternRecordNotFound,
	"IsRecordNotFoundError": PatternRecordNotFound,
	"NewScope":              PatternScope,
	"QueryExpr":             PatternQueryExpr,
	"SubQuery":              PatternSubQuery,
	"GetName":               PatternGetName,
}

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

	findings, scanned, err := check(roots)

	if err != nil {
		fmt.Fprintf(os.Stderr, "check-gorm-v1: %s\n", err)
		os.Exit(2)
	}

	if list {
		for _, f := range findings {
			fmt.Println(f.String())
		}
	}

	keys := keysOf(findings)
	baseline, err := readBaseline()

	if err != nil {
		fmt.Fprintf(os.Stderr, "check-gorm-v1: %s\n", err)
		os.Exit(2)
	}

	if update {
		if err = writeBaseline(mergeUnscanned(keys, baseline, scanned)); err != nil {
			fmt.Fprintf(os.Stderr, "check-gorm-v1: %s\n", err)
			os.Exit(2)
		}

		fmt.Printf("Recorded %d GORM v1 site(s) in %s.\n", len(findings), strings.Join(baselineFiles(), ", "))

		return
	}

	if added := exceeded(keys, baseline); len(added) > 0 {
		extra := 0

		for key := range added {
			extra += keys[key] - baseline[key]
		}

		// The baseline carries no line, so every site in a group that grew is listed.
		for _, f := range findings {
			if !added[f.key()] {
				continue
			}

			if f.first(findings) {
				fmt.Printf("%s: %s, recorded %d, found %d\n", f.File, f.Pattern, baseline[f.key()], keys[f.key()])
			}

			fmt.Printf("\t%s\n", f.Position)
		}

		fmt.Fprintf(os.Stderr, "\ncheck-gorm-v1: %d site(s) more than the baseline records, in %d group(s).\n"+
			"Use an API that works with GORM v1 and v2, or run with -update if the site is meant to stay.\n",
			extra, len(added))
		os.Exit(1)
	}

	fmt.Printf("GORM v1 sites checked, %d recorded (no increase over the baseline).\n", len(findings))
}

// baselineFile records how many sites each file is known to have, relative to the repository root.
const baselineFile = "scripts/tools/check-gorm-v1/baseline.txt"

// editions are the directories of separate repositories whose entries are kept in their own baseline file.
// An edition directory may be a symbolic link to its repository.
var editions = []string{"plus", "pro", "portal"}

// baselinePath returns the baseline file that records the sites of a slash-separated file path.
func baselinePath(file string) string {
	if edition, _, found := strings.Cut(file, "/"); found && slices.Contains(editions, edition) {
		return path.Join(edition, baselineFile)
	}

	return baselineFile
}

// baselineFiles returns the main baseline file and that of every edition directory that exists.
func baselineFiles() []string {
	files := []string{baselineFile}

	for _, edition := range editions {
		if info, err := os.Stat(edition); err == nil && info.IsDir() {
			files = append(files, path.Join(edition, baselineFile))
		}
	}

	return files
}

// finding is one site that depends on GORM v1.
type finding struct {
	File     string
	Position string
	Line     int
	Pattern  string
	Detail   string
}

// String renders a finding the way a compiler reports one.
func (f finding) String() string {
	return fmt.Sprintf("%s: %s: %s", f.Position, f.Pattern, f.Detail)
}

// key returns the baseline key of a finding.
func (f finding) key() string {
	return f.File + "\t" + f.Pattern
}

// first reports whether f is the earliest listed finding of its key.
func (f finding) first(findings []finding) bool {
	for _, other := range findings {
		if other.key() == f.key() {
			return other == f
		}
	}

	return false
}

// check loads the packages below every root that exists and returns the findings, sorted by
// location, and the roots it loaded.
func check(roots []string) (findings []finding, scanned []string, err error) {
	patterns := make([]string, 0, len(roots))

	for _, root := range roots {
		if _, statErr := os.Stat(root); statErr != nil { //nolint:gosec // G703: roots come from the command line or the default list
			continue
		}

		patterns = append(patterns, "./"+filepath.ToSlash(filepath.Clean(root))+"/...")
		scanned = append(scanned, root)
	}

	if len(scanned) == 0 {
		return nil, nil, nil
	}

	findings, err = scan(patterns, scanned)

	return findings, scanned, err
}

// scan loads the packages that match the patterns and returns the findings in files below the
// scanned roots, sorted by location. The entity package is always loaded for its bool columns.
func scan(patterns, scanned []string) (findings []finding, err error) {
	// Dependencies are type-checked from source, so the check does not read the export data of the
	// installed Go release.
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes |
			packages.NeedTypesInfo | packages.NeedImports | packages.NeedDeps,
		Tests:      true,
		BuildFlags: []string{"-tags=slow,develop,integration"},
	}

	pkgs, err := packages.Load(cfg, append([]string{entityPkg}, patterns...)...)

	if err != nil {
		return nil, err
	}

	wd, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	var columns *regexp.Regexp

	for _, pkg := range pkgs {
		if isScratch(pkg.PkgPath) {
			continue
		} else if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("%s: %s", pkg.PkgPath, pkg.Errors[0])
		}

		if pkg.PkgPath == entityPkg && columns == nil {
			columns = boolColumnPattern(boolColumns(pkg.Types))
		}
	}

	if columns == nil {
		return nil, fmt.Errorf("cannot load %s", entityPkg)
	}

	seen := make(map[string]bool)

	for _, pkg := range pkgs {
		if isScratch(pkg.PkgPath) {
			continue
		}

		for _, file := range pkg.Syntax {
			rel, relErr := filepath.Rel(wd, pkg.Fset.Position(file.Pos()).Filename)

			if relErr != nil {
				return nil, relErr
			}

			rel = filepath.ToSlash(rel)

			// A package with tests is loaded more than once, and generated files are not edited by hand.
			if seen[rel] || !underRoot(rel, scanned) || ast.IsGenerated(file) {
				continue
			}

			seen[rel] = true
			findings = append(findings, checkFile(pkg.Fset, pkg.TypesInfo, file, rel, columns)...)
		}
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].File != findings[j].File {
			return findings[i].File < findings[j].File
		}

		return findings[i].Line < findings[j].Line
	})

	return findings, nil
}

// isScratch reports whether a package is in the gitignored internal/zz* scratch directories, which must
// not fail the check.
func isScratch(pkgPath string) bool {
	return strings.HasPrefix(pkgPath, modulePath+"/internal/zz")
}

// checkFile returns the GORM v1 sites in one type-checked file.
func checkFile(fset *token.FileSet, info *types.Info, file *ast.File, rel string, columns *regexp.Regexp) (findings []finding) {
	add := func(pos token.Pos, pattern, detail string) {
		p := fset.Position(pos)
		findings = append(findings, finding{File: rel, Position: p.String(), Line: p.Line, Pattern: pattern, Detail: detail})
	}

	// counted holds the Count selectors of direct calls, so that a method value is reported once.
	counted := make(map[*ast.Ident]bool)

	ast.Inspect(file, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.Ident:
			obj := info.Uses[n]

			if obj == nil {
				obj = info.Defs[n]
			}

			if pattern := identPattern(obj); pattern != "" {
				add(n.Pos(), pattern, n.Name)
			} else if isGormCount(obj) && !counted[n] {
				add(n.Pos(), PatternCount, "Count method value")
			}
		case *ast.CallExpr:
			if sel, ok := n.Fun.(*ast.SelectorExpr); ok && isGormCount(info.Uses[sel.Sel]) {
				counted[sel.Sel] = true
			}

			if dest, ok := countDest(info, n); ok {
				add(n.Pos(), PatternCount, "Count("+dest+")")
			}
		case *ast.BasicLit:
			if n.Kind != token.STRING {
				return true
			}

			if s, err := strconv.Unquote(n.Value); err == nil {
				for _, m := range boolLiterals(columns, s) {
					add(n.Pos(), PatternBoolLiteral, m)
				}
			}
		}

		return true
	})

	return findings
}

// isGormCount reports whether obj is the Count method of gorm.DB.
func isGormCount(obj types.Object) bool {
	fn, ok := obj.(*types.Func)

	return ok && fn.Name() == "Count" && fn.Pkg() != nil && fn.Pkg().Path() == gormPkg
}

// identPattern returns the pattern an identifier's object belongs to, or an empty string.
func identPattern(obj types.Object) string {
	if obj == nil {
		return ""
	}

	switch obj := obj.(type) {
	case *types.Func:
		if obj.Pkg() != nil && obj.Pkg().Path() == gormPkg {
			return gormMethods[obj.Name()]
		}
	case *types.TypeName:
		if obj.Pkg() != nil && obj.Pkg().Path() == gormPkg && obj.Name() == "Scope" {
			return PatternScope
		}
	case *types.Var:
		if obj.IsField() && obj.Name() == "DeletedAt" && obj.Type().String() == "*time.Time" {
			return PatternDeletedAt
		}
	}

	return ""
}

// countDest returns the type of a (*gorm.DB).Count destination that is not *int64, for method calls
// and method expressions.
func countDest(info *types.Info, call *ast.CallExpr) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)

	if !ok || !isGormCount(info.Uses[sel.Sel]) || len(call.Args) == 0 {
		return "", false
	}

	t := info.TypeOf(call.Args[len(call.Args)-1])

	if t == nil {
		return "unknown", true
	} else if t.String() == "*int64" {
		return "", false
	}

	return t.String(), true
}

// boolColumns returns the column names of the bool fields of the structs a package declares.
func boolColumns(pkg *types.Package) []string {
	found := make(map[string]bool)
	scope := pkg.Scope()

	for _, name := range scope.Names() {
		tn, ok := scope.Lookup(name).(*types.TypeName)

		if !ok {
			continue
		}

		st, ok := tn.Type().Underlying().(*types.Struct)

		if !ok {
			continue
		}

		for i := 0; i < st.NumFields(); i++ {
			if column := boolColumn(st.Field(i), st.Tag(i)); column != "" {
				found[column] = true
			}
		}
	}

	columns := make([]string, 0, len(found))

	for column := range found {
		columns = append(columns, column)
	}

	sort.Strings(columns)

	return columns
}

// boolColumn returns the column name of an exported bool field that GORM maps, or an empty string.
func boolColumn(field *types.Var, tag string) string {
	if !field.Exported() || field.Embedded() {
		return ""
	} else if basic, ok := field.Type().(*types.Basic); !ok || basic.Kind() != types.Bool {
		return ""
	}

	settings := reflect.StructTag(tag).Get("gorm")

	for _, setting := range strings.Split(settings, ";") {
		key, value, _ := strings.Cut(strings.TrimSpace(setting), ":")

		switch strings.ToLower(strings.TrimSpace(key)) {
		case "-":
			return ""
		case "column":
			return strings.TrimSpace(value)
		}
	}

	return gorm.ToColumnName(field.Name())
}

// boolColumnPattern returns a pattern that matches a comparison of one of the columns with 0 or 1.
func boolColumnPattern(columns []string) *regexp.Regexp {
	quoted := make([]string, len(columns))

	for i, column := range columns {
		quoted[i] = regexp.QuoteMeta(column)
	}

	return regexp.MustCompile("(?i)(?:[`\"]?\\w+[`\"]?\\.)?[`\"]?(?:" + strings.Join(quoted, "|") + ")[`\"]?\\s*(?:=|<>|!=)\\s*[01]")
}

// boolLiterals returns the matches of the pattern in s that are not part of a longer name or number.
func boolLiterals(re *regexp.Regexp, s string) (found []string) {
	for _, loc := range re.FindAllStringIndex(s, -1) {
		if loc[0] > 0 && isNameByte(s[loc[0]-1]) || loc[1] < len(s) && isNameByte(s[loc[1]]) {
			continue
		}

		found = append(found, s[loc[0]:loc[1]])
	}

	return found
}

// isNameByte reports whether b can continue an SQL identifier or number.
func isNameByte(b byte) bool {
	return b == '_' || b == '.' || b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

// keysOf counts the findings per baseline key.
func keysOf(findings []finding) map[string]int {
	keys := make(map[string]int)

	for _, f := range findings {
		keys[f.key()]++
	}

	return keys
}

// mergeUnscanned returns the counts with the baseline entries of files outside the scanned roots
// added, so an update from a clone without the edition repositories does not drop their entries.
func mergeUnscanned(keys, baseline map[string]int, scanned []string) map[string]int {
	merged := make(map[string]int, len(keys)+len(baseline))

	for key, n := range keys {
		merged[key] = n
	}

	for key, n := range baseline {
		file, _, _ := strings.Cut(key, "\t")

		if !underRoot(file, scanned) {
			merged[key] = n
		}
	}

	return merged
}

// underRoot reports whether a slash-separated file path is one of the roots or below one.
func underRoot(file string, roots []string) bool {
	for _, root := range roots {
		root = strings.TrimSuffix(filepath.ToSlash(filepath.Clean(root)), "/")

		if root == "." || file == root || strings.HasPrefix(file, root+"/") {
			return true
		}
	}

	return false
}

// readBaseline returns the counts recorded in all baseline files.
func readBaseline() (map[string]int, error) {
	counts := make(map[string]int)

	for _, name := range baselineFiles() {
		recorded, err := readBaselineFile(name)

		if err != nil {
			return nil, err
		}

		for key, n := range recorded {
			counts[key] = n
		}
	}

	return counts, nil
}

// readBaselineFile returns the counts recorded in one baseline file, or none if it does not exist.
func readBaselineFile(name string) (map[string]int, error) {
	if fs.IsSymlink(name) {
		return nil, fmt.Errorf("%s is a symbolic link", name)
	}

	b, err := os.ReadFile(name) //nolint:gosec // G304: fixed baseline paths below the working directory

	if os.IsNotExist(err) {
		return map[string]int{}, nil
	} else if err != nil {
		return nil, err
	}

	return parseBaseline(name, string(b))
}

// parseBaseline reads lines of a count followed by a tab and a key, skipping comments and blank lines.
func parseBaseline(name, s string) (map[string]int, error) {
	counts := make(map[string]int)

	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		n, key, found := strings.Cut(line, "\t")

		if !found {
			return nil, fmt.Errorf("%s: cannot read %q", name, line)
		}

		c, convErr := strconv.Atoi(n)

		if convErr != nil {
			return nil, fmt.Errorf("%s: cannot read %q", name, line)
		}

		counts[key] = c
	}

	return counts, nil
}

// writeBaseline writes the counts sorted by key, each to the baseline file of its repository.
// Entries of an edition whose directory does not exist are not written. The main file is written
// last, so a failed write does not drop the edition entries it may still hold.
func writeBaseline(keys map[string]int) error {
	files := baselineFiles()
	parts := make(map[string]map[string]int, len(files))

	for _, name := range files {
		if fs.IsSymlink(name) {
			return fmt.Errorf("%s is a symbolic link", name)
		} else if err := os.MkdirAll(filepath.Dir(name), fs.ModeDir); err != nil {
			return err
		}

		parts[name] = make(map[string]int)
	}

	for key, n := range keys {
		file, _, _ := strings.Cut(key, "\t")

		if part, ok := parts[baselinePath(file)]; ok {
			part[key] = n
		}
	}

	for _, name := range append(files[1:], files[0]) {
		if err := os.WriteFile(name, []byte(formatBaseline(parts[name])), fs.ModeFile); err != nil {
			return err
		}
	}

	return nil
}

// formatBaseline renders the counts sorted by key, below a header that names the columns.
func formatBaseline(keys map[string]int) string {
	sorted := make([]string, 0, len(keys))

	for key := range keys {
		sorted = append(sorted, key)
	}

	sort.Strings(sorted)

	var b strings.Builder

	b.WriteString("# Code that depends on GORM v1: count, file, pattern.\n")
	b.WriteString("# Regenerate in the CE repository with: go run ./scripts/tools/check-gorm-v1 -update\n")

	for _, key := range sorted {
		fmt.Fprintf(&b, "%d\t%s\n", keys[key], key)
	}

	return b.String()
}

// exceeded returns the keys whose count is higher than the baseline records.
func exceeded(keys, baseline map[string]int) map[string]bool {
	over := make(map[string]bool)

	for key, n := range keys {
		if n > baseline[key] {
			over[key] = true
		}
	}

	return over
}
