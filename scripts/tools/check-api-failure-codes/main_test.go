package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/photoprism/photoprism/pkg/fs"
)

// header is the start of every fixture package. Its AbortRequestTooLarge uses a literal status, so
// only its name marks it.
const header = `package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var _ = http.StatusOK

// AbortRequestTooLarge writes a 413 response.
func AbortRequestTooLarge(c *gin.Context) {
	c.AbortWithStatus(413)
}
`

// undocumented and documented are Swagger annotations without and with 413.
const undocumented = `
//	@Failure	400,401	{object}	i18n.Response
//	@Router		/api/v1/x [post]
`
const documented = `
//	@Failure	400,401,413	{object}	i18n.Response
//	@Router		/api/v1/x [post]
`

// writePackage writes the named files into a new directory and returns it.
func writePackage(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()

	for name, src := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), fs.ModeFile); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

// fixture writes a package made of header and body and returns its directory.
func fixture(t *testing.T, body string) string {
	t.Helper()
	return writePackage(t, map[string]string{"api.go": header + body})
}

// result returns the handlers found in dir by name.
func result(t *testing.T, dir string) map[string]Handler {
	t.Helper()
	found, err := check(dir)

	if err != nil {
		t.Fatal(err)
	}

	byName := make(map[string]Handler, len(found))

	for _, h := range found {
		byName[h.Name] = h
	}

	return byName
}

// parseFunc parses src as a file and returns its function with the given name.
func parseFunc(t *testing.T, src, name string) *ast.FuncDecl {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "x.go", src, parser.ParseComments)

	if err != nil {
		t.Fatal(err)
	}

	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == name {
			return fn
		}
	}

	t.Fatalf("function %s not found", name)

	return nil
}

func TestRun(t *testing.T) {
	missing := fixture(t, undocumented+`func Direct(router *gin.RouterGroup) {
	router.POST("/x", func(c *gin.Context) { AbortRequestTooLarge(c) })
}
`)
	passing := fixture(t, documented+`func Direct(router *gin.RouterGroup) {
	router.POST("/x", func(c *gin.Context) { AbortRequestTooLarge(c) })
}
`)

	t.Run("Missing", func(t *testing.T) {
		var out, errOut bytes.Buffer
		if code := run([]string{missing}, &out, &errOut); code != 1 {
			t.Fatalf("expected exit code 1, got %d: %s", code, out.String())
		}
		if !strings.Contains(out.String(), "handler Direct can return 413") {
			t.Fatalf("expected a finding: %s", out.String())
		}
	})
	t.Run("Success", func(t *testing.T) {
		var out, errOut bytes.Buffer
		if code := run([]string{passing}, &out, &errOut); code != 0 {
			t.Fatalf("expected exit code 0, got %d: %s", code, out.String())
		}
	})
	t.Run("List", func(t *testing.T) {
		var out, errOut bytes.Buffer
		if code := run([]string{"-list", missing}, &out, &errOut); code != 1 || !strings.Contains(out.String(), "Direct (missing)") {
			t.Fatalf("expected a listed finding and exit code 1, got %d: %s", code, out.String())
		}
	})
	t.Run("InvalidRoot", func(t *testing.T) {
		var out, errOut bytes.Buffer
		if code := run([]string{filepath.Join(t.TempDir(), "absent")}, &out, &errOut); code != 2 {
			t.Fatalf("expected exit code 2, got %d", code)
		}
	})
	t.Run("UnknownOption", func(t *testing.T) {
		var out, errOut bytes.Buffer
		if code := run([]string{"-h"}, &out, &errOut); code != 2 || !strings.Contains(errOut.String(), "unknown option -h") {
			t.Fatalf("expected exit code 2 and an unknown option, got %d: %s", code, errOut.String())
		}
	})
	t.Run("ParseError", func(t *testing.T) {
		var out, errOut bytes.Buffer
		dir := writePackage(t, map[string]string{"api.go": "package api\nfunc {"})
		if code := run([]string{dir}, &out, &errOut); code != 2 {
			t.Fatalf("expected exit code 2, got %d", code)
		}
	})
}

func TestCheck(t *testing.T) {
	t.Run("SkipsTests", func(t *testing.T) {
		dir := writePackage(t, map[string]string{
			"api.go":      header,
			"api_test.go": "package api\n" + undocumented + "func Tested(router *gin.RouterGroup) { AbortRequestTooLarge(nil) }\n",
		})
		if h := result(t, dir); len(h) != 0 {
			t.Fatalf("expected test files to be skipped: %+v", h)
		}
	})
	t.Run("OtherPackage", func(t *testing.T) {
		// An edition package calls the function of the CE package it imports.
		h := result(t, writePackage(t, map[string]string{"api.go": `package api

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/api"
)
` + undocumented + `func Other(router *gin.RouterGroup) {
	router.POST("/x", func(c *gin.Context) { api.AbortRequestTooLarge(c) })
}
`}))
		if !h["Other"].Missing() {
			t.Fatalf("expected Other to be missing 413: %+v", h)
		}
	})
}

func TestHandlers(t *testing.T) {
	t.Run("Direct", func(t *testing.T) {
		h := result(t, fixture(t, undocumented+`func Direct(router *gin.RouterGroup) {
	router.POST("/x", func(c *gin.Context) { AbortRequestTooLarge(c) })
}
`))
		if !h["Direct"].Missing() {
			t.Fatalf("expected Direct to be missing 413: %+v", h)
		}
	})
	t.Run("Documented", func(t *testing.T) {
		h := result(t, fixture(t, documented+`func Documented(router *gin.RouterGroup) {
	router.POST("/x", func(c *gin.Context) { AbortRequestTooLarge(c) })
}
`))
		if !h["Documented"].Documented || h["Documented"].Missing() {
			t.Fatalf("expected Documented to pass: %+v", h)
		}
	})
	t.Run("NoSwagger", func(t *testing.T) {
		h := result(t, fixture(t, `
// Registrar registers routes without a Swagger annotation.
func Registrar(router *gin.RouterGroup) {
	router.POST("/x", func(c *gin.Context) { AbortRequestTooLarge(c) })
}
`))
		if r, ok := h["Registrar"]; !ok || r.Swagger || r.Missing() {
			t.Fatalf("expected Registrar to be listed but not checked: %+v", h)
		}
	})
	t.Run("NotHandler", func(t *testing.T) {
		h := result(t, fixture(t, undocumented+`func Helper(c *gin.Context) { AbortRequestTooLarge(c) }
`))
		if _, ok := h["Helper"]; ok {
			t.Fatalf("expected Helper not to be a handler: %+v", h)
		}
	})
}

func TestAbortingFuncs(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "api.go", header+`
func inner(c *gin.Context) { AbortRequestTooLarge(c) }

func outer(c *gin.Context) { inner(c) }

var handle = func(c *gin.Context) { outer(c) }

func status(c *gin.Context) { c.JSON(http.StatusRequestEntityTooLarge, nil) }

func plain(c *gin.Context) { c.JSON(http.StatusOK, nil) }
`, parser.ParseComments)

	if err != nil {
		t.Fatal(err)
	}

	aborting := abortingFuncs(collect([]*ast.File{f}))

	// AbortRequestTooLarge itself is matched by name, so only its callers join the set.
	for _, name := range []string{"inner", "outer", "handle", "status"} {
		if !aborting[name] {
			t.Errorf("expected %s to be aborting", name)
		}
	}

	if aborting["plain"] {
		t.Error("expected plain not to be aborting")
	}
}

func TestRefersToAborting(t *testing.T) {
	aborting := map[string]bool{"UpdateLink": true}
	info := pkgInfo{httpPkg: map[string]bool{"http": true, "nethttp": true}}
	body := func(t *testing.T, src string) ast.Node {
		t.Helper()
		return parseFunc(t, "package api\n\nfunc F() {\n"+src+"\n}\n", "F").Body
	}

	t.Run("Call", func(t *testing.T) {
		if !refersToAborting(body(t, "UpdateLink(nil)"), aborting, info) {
			t.Fatal("expected a call to be found")
		}
	})
	t.Run("Value", func(t *testing.T) {
		if !refersToAborting(body(t, "_ = UpdateLink"), aborting, info) {
			t.Fatal("expected a value reference to be found")
		}
	})
	t.Run("AliasedStatus", func(t *testing.T) {
		if !refersToAborting(body(t, "_ = nethttp.StatusRequestEntityTooLarge"), aborting, info) {
			t.Fatal("expected the aliased status to be found")
		}
	})
	t.Run("OtherStatus", func(t *testing.T) {
		if refersToAborting(body(t, "_ = other.StatusRequestEntityTooLarge"), aborting, info) {
			t.Fatal("expected a constant of another package not to count")
		}
	})
	t.Run("Method", func(t *testing.T) {
		if refersToAborting(body(t, "share.UpdateLink(nil)"), aborting, info) {
			t.Fatal("expected a method name not to count")
		}
	})
	t.Run("CompositeKey", func(t *testing.T) {
		if refersToAborting(body(t, "_ = opts{UpdateLink: 1}"), aborting, info) {
			t.Fatal("expected a composite literal key not to count")
		}
	})
	t.Run("Local", func(t *testing.T) {
		if refersToAborting(body(t, "UpdateLink := 1\n_ = UpdateLink"), aborting, info) {
			t.Fatal("expected a local name not to count")
		}
	})
}

func TestHandlers_FuncVar(t *testing.T) {
	// A handler that registers a function variable of its own file.
	h := result(t, fixture(t, `
var handleBig = func(c *gin.Context) { AbortRequestTooLarge(c) }
`+undocumented+`func Big(router *gin.RouterGroup) {
	router.POST("/x", handleBig)
}
`))

	if !h["Big"].Missing() {
		t.Fatalf("expected Big to be missing 413: %+v", h)
	}
}

func TestIsHandler(t *testing.T) {
	gin := map[string]bool{"gin": true}

	t.Run("Pointer", func(t *testing.T) {
		if !isHandler(parseFunc(t, "package api\nfunc H(router *gin.RouterGroup) {}\n", "H"), gin) {
			t.Fatal("expected a handler")
		}
	})
	t.Run("Alias", func(t *testing.T) {
		if !isHandler(parseFunc(t, "package api\nfunc H(router *g.RouterGroup) {}\n", "H"), map[string]bool{"g": true}) {
			t.Fatal("expected a handler with an aliased import")
		}
	})
	t.Run("Value", func(t *testing.T) {
		if isHandler(parseFunc(t, "package api\nfunc H(router gin.RouterGroup) {}\n", "H"), gin) {
			t.Fatal("expected a non-pointer parameter not to count")
		}
	})
	t.Run("OtherPackage", func(t *testing.T) {
		if isHandler(parseFunc(t, "package api\nfunc H(router *echo.RouterGroup) {}\n", "H"), gin) {
			t.Fatal("expected a RouterGroup of another package not to count")
		}
	})
}

func TestHasRouter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if !hasRouter(parseFunc(t, "package api\n"+undocumented+"func H() {}\n", "H").Doc) {
			t.Fatal("expected a @Router annotation")
		}
	})
	t.Run("Prose", func(t *testing.T) {
		if hasRouter(parseFunc(t, "package api\n// H has no @Router annotation.\nfunc H() {}\n", "H").Doc) {
			t.Fatal("expected prose not to count")
		}
	})
	t.Run("Nil", func(t *testing.T) {
		if hasRouter(nil) {
			t.Fatal("expected false for a missing doc comment")
		}
	})
}

func TestDocuments413(t *testing.T) {
	doc := func(t *testing.T, lines string) *ast.CommentGroup {
		t.Helper()
		return parseFunc(t, "package api\n"+lines+"func H() {}\n", "H").Doc
	}

	t.Run("Success", func(t *testing.T) {
		if !documents413(doc(t, documented)) {
			t.Fatal("expected 413 to be found")
		}
	})
	t.Run("SecondLine", func(t *testing.T) {
		if !documents413(doc(t, "//\t@Failure\t400\t{object}\ti18n.Response\n//\t@Failure\t413\t{object}\ti18n.Response\n")) {
			t.Fatal("expected 413 on a second @Failure line to be found")
		}
	})
	t.Run("Missing", func(t *testing.T) {
		if documents413(doc(t, undocumented)) {
			t.Fatal("expected 413 to be missing")
		}
	})
	t.Run("OtherTag", func(t *testing.T) {
		if documents413(doc(t, "//\t@Success\t413\t{object}\ti18n.Response\n")) {
			t.Fatal("expected 413 under another tag not to count")
		}
	})
	t.Run("Nil", func(t *testing.T) {
		if documents413(nil) {
			t.Fatal("expected false for a missing doc comment")
		}
	})
}

func TestAnnotations(t *testing.T) {
	fn := parseFunc(t, "package api\n"+documented+"func H() {}\n", "H")

	t.Run("Success", func(t *testing.T) {
		if got := annotations(fn.Doc, "@Failure"); len(got) != 1 || got[0][1] != "400,401,413" {
			t.Fatalf("unexpected fields: %v", got)
		}
	})
	t.Run("Nil", func(t *testing.T) {
		if got := annotations(nil, "@Failure"); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})
}

func TestHandler_Missing(t *testing.T) {
	t.Run("Missing", func(t *testing.T) {
		if !(Handler{Swagger: true}).Missing() {
			t.Fatal("expected a Swagger handler without 413 to be missing")
		}
	})
	t.Run("Documented", func(t *testing.T) {
		if (Handler{Swagger: true, Documented: true}).Missing() {
			t.Fatal("expected a documented handler not to be missing")
		}
	})
	t.Run("NoSwagger", func(t *testing.T) {
		if (Handler{}).Missing() {
			t.Fatal("expected a handler without Swagger not to be missing")
		}
	})
}

func TestCollect(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "api.go", `package api

import (
	nethttp "net/http"

	g "github.com/gin-gonic/gin"
)

var handle = func(c *g.Context) {}

var count = 1

func (w writer) Method() {}

func Plain() {}
`, parser.ParseComments)

	if err != nil {
		t.Fatal(err)
	}

	info := collect([]*ast.File{f})

	if !info.httpPkg["nethttp"] || !info.ginPkg["g"] {
		t.Fatalf("expected aliased imports: %+v %+v", info.httpPkg, info.ginPkg)
	}

	if _, ok := info.funcs["handle"]; !ok {
		t.Fatal("expected the function variable")
	}

	if _, ok := info.funcs["count"]; ok {
		t.Fatal("expected a non-function variable to be skipped")
	}

	if _, ok := info.funcs["Method"]; ok {
		t.Fatal("expected methods to be skipped")
	}

	if _, ok := info.funcs["Plain"]; !ok || len(info.decls) != 1 {
		t.Fatalf("expected one function declaration: %+v", info.decls)
	}
}
