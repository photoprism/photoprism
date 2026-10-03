/*
Command check-api-failure-codes reports REST handlers that can answer 413 Request Entity Too Large
without listing 413 in their Swagger @Failure annotation.

A handler is a function with a *gin.RouterGroup parameter. It can answer 413 if its body, including
function literals, refers to AbortRequestTooLarge or net/http's StatusRequestEntityTooLarge, or to a
package-level function or function variable of the same package that does, directly or through
further ones. Local names and fields do not count, and methods are not followed, because a method
name alone does not say which type it belongs to.

A handler with a @Router annotation passes if one of its @Failure lines lists 413; a handler without
one is not part of the Swagger document, so it is listed but not checked. The check scans
internal/api and the api packages of the editions that are present, the directories Swagger is
generated from, and exits with 1 on a finding and 2 on an error. Run with -list to print every
handler that can answer 413.

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.
*/
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// defaultRoots are the API packages Swagger is generated from. The edition directories are
// separate repositories that a clone may not have, so a missing one is skipped rather than reported.
var defaultRoots = []string{"internal/api", "plus/internal/api", "pro/internal/api", "portal/internal/api"}

// abortFunc is the function that answers a request with 413.
const abortFunc = "AbortRequestTooLarge"

// statusConst is the net/http constant for 413.
const statusConst = "StatusRequestEntityTooLarge"

// Handler describes a handler that can answer 413.
type Handler struct {
	Name       string
	Pos        token.Position
	Swagger    bool
	Documented bool
}

// Missing reports whether the handler is part of the Swagger document but does not list 413.
func (h Handler) Missing() bool {
	return h.Swagger && !h.Documented
}

// pkgInfo holds the declarations of a package that the rule needs.
type pkgInfo struct {
	funcs    map[string]ast.Node
	decls    []*ast.FuncDecl
	httpPkg  map[string]bool
	ginPkg   map[string]bool
	topLevel map[*ast.Object]bool //nolint:staticcheck // Parser scopes suffice here.
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run checks the roots named in args, or the default roots, and returns the exit code.
func run(args []string, stdout, stderr io.Writer) int {
	list := false
	roots := make([]string, 0, len(args))

	for _, arg := range args {
		switch {
		case arg == "-list":
			list = true
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(stderr, "check-api-failure-codes: unknown option %s\n", arg)
			return 2
		default:
			roots = append(roots, arg)
		}
	}

	explicit := len(roots) > 0

	if !explicit {
		roots = defaultRoots
	}

	var found []Handler

	for _, root := range roots {
		if info, err := os.Stat(root); err != nil || !info.IsDir() { //nolint:gosec // Roots are named by the developer running the check.
			if explicit {
				fmt.Fprintf(stderr, "check-api-failure-codes: %s is not a directory\n", root)
				return 2
			}

			continue
		}

		handlers, err := check(root)

		if err != nil {
			fmt.Fprintf(stderr, "check-api-failure-codes: %s\n", err)
			return 2
		}

		found = append(found, handlers...)
	}

	missing := 0

	for _, h := range found {
		if list {
			status := "documented"

			if !h.Swagger {
				status = "no swagger"
			} else if !h.Documented {
				status = "missing"
			}

			fmt.Fprintf(stdout, "%s: %s (%s)\n", h.Pos, h.Name, status)
		}

		if h.Missing() {
			missing++

			if !list {
				fmt.Fprintf(stdout, "%s: handler %s can return 413 but its @Failure list omits it\n", h.Pos, h.Name)
			}
		}
	}

	if missing > 0 {
		fmt.Fprintf(stdout, "%d of %d handlers that can return 413 do not document it.\n", missing, len(found))
		return 1
	}

	fmt.Fprintf(stdout, "API failure codes checked, %d handlers that can return 413 document it.\n", len(found))

	return 0
}

// check parses the non-test Go files in dir and returns the handlers that can answer 413.
func check(dir string) ([]Handler, error) {
	fset := token.NewFileSet()
	names, err := filepath.Glob(filepath.Join(dir, "*.go"))

	if err != nil {
		return nil, err
	}

	var files []*ast.File

	for _, name := range names {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		f, parseErr := parser.ParseFile(fset, name, nil, parser.ParseComments)

		if parseErr != nil {
			return nil, parseErr
		}

		files = append(files, f)
	}

	return handlers(fset, files), nil
}

// collect returns the functions, function variables and import names of a package.
func collect(files []*ast.File) pkgInfo {
	info := pkgInfo{funcs: make(map[string]ast.Node), httpPkg: make(map[string]bool), ginPkg: make(map[string]bool), topLevel: make(map[*ast.Object]bool)} //nolint:staticcheck // Parser scopes suffice here.

	for _, f := range files {
		if f.Scope != nil { //nolint:staticcheck // Parser scopes suffice here.
			for _, obj := range f.Scope.Objects { //nolint:staticcheck // Parser scopes suffice here.
				info.topLevel[obj] = true
			}
		}

		for _, imp := range f.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			name := filepath.Base(path)

			if imp.Name != nil {
				name = imp.Name.Name
			}

			switch path {
			case "net/http":
				info.httpPkg[name] = true
			case "github.com/gin-gonic/gin":
				info.ginPkg[name] = true
			}
		}

		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Body != nil && d.Recv == nil {
					info.funcs[d.Name.Name] = d.Body
					info.decls = append(info.decls, d)
				}
			case *ast.GenDecl:
				if d.Tok != token.VAR {
					continue
				}

				for _, spec := range d.Specs {
					vs, ok := spec.(*ast.ValueSpec)

					if !ok {
						continue
					}

					for i, name := range vs.Names {
						if i < len(vs.Values) {
							if lit, isLit := vs.Values[i].(*ast.FuncLit); isLit {
								info.funcs[name.Name] = lit.Body
							}
						}
					}
				}
			}
		}
	}

	return info
}

// handlers returns the handlers in files that can answer 413, sorted by position.
func handlers(fset *token.FileSet, files []*ast.File) []Handler {
	info := collect(files)
	aborting := abortingFuncs(info)

	var result []Handler

	for _, fn := range info.decls {
		if !isHandler(fn, info.ginPkg) || !aborting[fn.Name.Name] {
			continue
		}

		result = append(result, Handler{Name: fn.Name.Name, Pos: fset.Position(fn.Pos()), Swagger: hasRouter(fn.Doc), Documented: documents413(fn.Doc)})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Pos.Filename != result[j].Pos.Filename {
			return result[i].Pos.Filename < result[j].Pos.Filename
		}

		return result[i].Pos.Line < result[j].Pos.Line
	})

	return result
}

// abortingFuncs returns the names of the functions and function variables that can answer 413,
// directly or through others of the package, computed to a fixpoint.
func abortingFuncs(info pkgInfo) map[string]bool {
	aborting := make(map[string]bool)

	for changed := true; changed; {
		changed = false

		for name, body := range info.funcs {
			if !aborting[name] && refersToAborting(body, aborting, info) {
				aborting[name] = true
				changed = true
			}
		}
	}

	return aborting
}

// refersToAborting reports whether node refers to AbortRequestTooLarge, qualified or not, to
// StatusRequestEntityTooLarge of net/http, or to a package-level name in the aborting set, by call
// or by value. A local name, a parameter, a composite literal key, and the name after a dot of
// anything else do not count.
func refersToAborting(node ast.Node, aborting map[string]bool, info pkgInfo) (found bool) {
	skipped := make(map[*ast.Ident]bool)

	ast.Inspect(node, func(n ast.Node) bool {
		if found {
			return false
		}

		switch x := n.(type) {
		case *ast.CompositeLit:
			for _, elt := range x.Elts {
				if kv, ok := elt.(*ast.KeyValueExpr); ok {
					if key, isIdent := kv.Key.(*ast.Ident); isIdent {
						skipped[key] = true
					}
				}
			}
		case *ast.SelectorExpr:
			skipped[x.Sel] = true

			if pkg, ok := x.X.(*ast.Ident); ok && info.httpPkg[pkg.Name] && x.Sel.Name == statusConst {
				found = true
			} else if x.Sel.Name == abortFunc {
				found = true
			}
		case *ast.Ident:
			if skipped[x] {
				return true
			}

			// A name declared inside a function resolves to a local object.
			if x.Obj != nil && x.Obj.Kind != ast.Fun && !info.topLevel[x.Obj] { //nolint:staticcheck // Parser scopes suffice here.
				return true
			}

			if x.Name == abortFunc || aborting[x.Name] {
				found = true
			}
		}

		return !found
	})

	return found
}

// isHandler reports whether fn takes a *gin.RouterGroup parameter, given the names gin is imported as.
func isHandler(fn *ast.FuncDecl, ginPkg map[string]bool) bool {
	for _, field := range fn.Type.Params.List {
		star, ok := field.Type.(*ast.StarExpr)

		if !ok {
			continue
		}

		if sel, ok := star.X.(*ast.SelectorExpr); ok && sel.Sel.Name == "RouterGroup" {
			if pkg, ok := sel.X.(*ast.Ident); ok && ginPkg[pkg.Name] {
				return true
			}
		}
	}

	return false
}

// hasRouter reports whether doc carries a Swagger @Router annotation.
func hasRouter(doc *ast.CommentGroup) bool {
	return len(annotations(doc, "@Router")) > 0
}

// documents413 reports whether one of the @Failure lines in doc lists 413.
func documents413(doc *ast.CommentGroup) bool {
	for _, fields := range annotations(doc, "@Failure") {
		if len(fields) > 1 && slices.Contains(strings.Split(fields[1], ","), "413") {
			return true
		}
	}

	return false
}

// annotations returns the whitespace-separated fields of the doc lines that start with tag.
func annotations(doc *ast.CommentGroup, tag string) (result [][]string) {
	if doc == nil {
		return nil
	}

	for _, c := range doc.List {
		if fields := strings.Fields(strings.TrimPrefix(c.Text, "//")); len(fields) > 0 && fields[0] == tag {
			result = append(result, fields)
		}
	}

	return result
}
