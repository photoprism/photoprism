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

// fixture writes a source-only package for the scanner.
func fixture(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	writeFixture(t, filepath.Join(root, "logger.go"), `package fixture
import "github.com/photoprism/photoprism/internal/event"
var log = event.Log
`)
	writeFixture(t, filepath.Join(root, "calls.go"), `package fixture
import (
 "github.com/photoprism/photoprism/internal/event"
 "github.com/photoprism/photoprism/pkg/clean"
 "strings"
 "fmt"
)
`+body)
	return root
}

// writeFixture writes source or baseline data to a test path.
func writeFixture(t *testing.T, path, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), fs.ModeFile); err != nil {
		t.Fatal(err)
	}
}

// TestScan exercises production-shaped calls and separated operator controls.
func TestScan(t *testing.T) {
	t.Run("KnownArguments", func(t *testing.T) {
		root := fixture(t, `func review() {
log.Infof("folder: %s", a.AlbumFilter)
log.Infof("import: %s", folder.Path)
log.Warnf("transcode: %s", relName)
log.Errorf("import: %s", err.Error())
log.Error(result.Err)
log.Infof("connection: %s", clean.Log(serviceURL))
log.Infof("account: %s", password)
}`)
		found, err := scan([]string{root, root}, false)
		if err != nil || len(found) != 7 {
			t.Fatalf("found %d: %v: %+v", len(found), err, found)
		}
		if found[5].Rule != "redaction" || found[6].Rule != "redaction" {
			t.Fatalf("credential review missing: %+v", found)
		}
		if found[0].Function != "review" || found[0].Argument != 2 || found[0].Line == 0 {
			t.Fatal(found[0])
		}
	})
	t.Run("RenderedAndSeparated", func(t *testing.T) {
		root := fixture(t, `func review() {
log.Infof("folder: %s", clean.Log(folder.Path))
log.Warnf("transcode: %s", clean.LogQuote(relName))
log.Error(clean.Error(err))
log.Info(strings.ToLower(clean.Error(err)))
log.Info(fmt.Sprintf("file: %s", clean.Log(relName)))
log.Info("file: " + clean.Log(relName))
log.Info(clean.Log(clean.UriRedacted(serviceURL)))
log.Info(clean.Log(clean.FileNameRedacted(archiveToken)))
event.SystemError([]string{"worker", "failed"}, clean.ErrorFull(err))
event.SystemLog.Error(err)
log.Debug(err)
}`)
		found, err := scan([]string{root}, false)
		if err != nil || len(found) != 0 {
			t.Fatalf("%+v: %v", found, err)
		}
	})
	t.Run("WholeExpression", func(t *testing.T) {
		root := fixture(t, `func review() {
log.Info(clean.Log(name) + raw)
log.Info(fmt.Sprintf("%s %s", clean.Log(name), raw))
log.Info(arbitrary(clean.Log(name)))
log.Error(clean.ErrorFull(err))
log.Info(clean.UriRedacted(serviceURL) + token)
log.Info(clean.Log(password))
}`)
		found, err := scan([]string{root}, false)
		if err != nil || len(found) != 6 {
			t.Fatalf("%+v: %v", found, err)
		}
	})
	t.Run("AliasesAndShadowing", func(t *testing.T) {
		root := fixture(t, `func ordinary() { event.Log.Warn(err) }
func shadow(log other) { log.Error(err) }
func shadowClean() { clean := other; log.Info(clean.Log(name)) }
func shadowBuiltin() { var true = name; log.Info(true) }
func shadowInitializer() { var other, log = event.Log, alternate; log.Info(name) }
func (x *Thing) report() { log.Error(err) }
`)
		writeFixture(t, filepath.Join(root, "alias.go"), `package fixture
import e "github.com/photoprism/photoprism/internal/event"
import c "github.com/photoprism/photoprism/pkg/clean"
var app = e.Log
var system = e.SystemLog
func alias() { app.Error(err); e.Log.Info(c.Log(name)); system.Error(err) }
`)
		found, err := scan([]string{root}, false)
		if err != nil || len(found) != 5 {
			t.Fatalf("%+v: %v", found, err)
		}
	})
	t.Run("ExclusionsAndErrors", func(t *testing.T) {
		root := fixture(t, `func review() { log.Error(err) }`)
		writeFixture(t, filepath.Join(root, "calls_test.go"), "not go")
		if err := os.Mkdir(filepath.Join(root, "testdata"), fs.ModeDir); err != nil {
			t.Fatal(err)
		}
		writeFixture(t, filepath.Join(root, "testdata", "bad.go"), "not go")
		found, err := scan([]string{root}, false)
		if err != nil || len(found) != 1 {
			t.Fatalf("%+v %v", found, err)
		}
		writeFixture(t, filepath.Join(root, "bad.go"), "not go")
		if _, err = scan([]string{root}, false); err == nil {
			t.Fatal("accepted malformed source")
		}
		if _, err = scan([]string{filepath.Join(root, "missing")}, false); err == nil {
			t.Fatal("accepted missing explicit root")
		}
	})
}

// TestRendered verifies the expression boundary independently of logger discovery.
func TestRendered(t *testing.T) {
	imports := map[string]string{"clean": modulePath + "pkg/clean", "status": modulePath + "pkg/log/status", "fmt": "fmt", "strings": "strings"}
	for _, tc := range []struct {
		name, input string
		want        bool
	}{
		{"Literal", `"value"`, true}, {"BoolIdentifier", `true`, false}, {"Parentheses", `(clean.Log(name))`, true},
		{"Names", `clean.LogNames(names)`, true}, {"Lower", `clean.LogLower(name)`, true},
		{"Status", `status.Error(err)`, true}, {"Unknown", `safe(name)`, false},
		{"Mixed", `clean.Log(name)+raw`, false}, {"Credential", `clean.Log(accessToken)`, false},
		{"RedactedOnly", `clean.UriRedacted(proxy)`, false}, {"RedactedRendered", `clean.Log(clean.UriRedacted(proxy))`, true}, {"RawError", `err.Error()`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := parser.ParseExpr(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if got := rendered(e, imports); got != tc.want {
				t.Fatalf("%s: %v", tc.input, got)
			}
			if expression(token.NewFileSet(), e) == "?" {
				t.Fatal("expression failed")
			}
		})
	}
}

// TestCountsAndAdditions pins per-call identity, multiplicity and stable line movement.
func TestCountsAndAdditions(t *testing.T) {
	a := finding{File: "a.go", Function: "one", Call: `log.Error(err)`, Argument: 1, Rule: "rendering", Line: 10}
	b := a
	b.Line = 30
	c := a
	c.Function = "two"
	recorded := counts([]finding{a, a})
	moved := counts([]finding{b, b})
	if len(additions(moved, recorded)) != 0 {
		t.Fatal("line movement changed identity")
	}
	added := additions(counts([]finding{c}), recorded)
	c.Line = 0
	if added[c] != 1 {
		t.Fatal("unrelated removals hid new call")
	}
	a.Line = 0
	if additions(counts([]finding{b, b, b}), recorded)[a] != 1 {
		t.Fatal("multiplicity lost")
	}
}

// TestBaseline checks round trips and rejects unusable accounting files.
func TestBaseline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "baseline.json")
	f := finding{File: "file.go", Function: "test", Call: `log.Error(err)`, Argument: 1, Rule: "rendering"}
	t.Run("RoundTrip", func(t *testing.T) {
		if err := writeBaseline(path, map[finding]int{f: 2}); err != nil {
			t.Fatal(err)
		}
		first, err := os.ReadFile(path) // #nosec G304 -- test reads its own temporary baseline.
		if err != nil {
			t.Fatal(err)
		}
		if err = writeBaseline(path, map[finding]int{f: 2}); err != nil {
			t.Fatal(err)
		}
		second, _ := os.ReadFile(path) // #nosec G304 -- test reads its own temporary baseline.
		if !bytes.Equal(first, second) {
			t.Fatal("unstable output")
		}
		got, err := readBaseline(path)
		if err != nil || got[f] != 2 {
			t.Fatalf("%+v %v", got, err)
		}
	})
	t.Run("Invalid", func(t *testing.T) {
		valid := `{"file":"file.go","function":"test","call":"log.Error(err)","argument":1,"rule":"rendering"}`
		for _, data := range []string{`{`, `{"version":2}`, `{"version":1,"extra":0}`, `{"version":1} {}`, `{"version":1,"entries":[{"finding":` + valid + `,"count":0}]}`, `{"version":1,"entries":[{"finding":` + valid + `,"count":1},{"finding":` + valid + `,"count":1}]}`} {
			writeFixture(t, path, data)
			if _, err := readBaseline(path); err == nil {
				t.Fatalf("accepted %s", data)
			}
		}
		if _, err := readBaseline(path + "missing"); err == nil {
			t.Fatal("missing accepted")
		}
		if err := writeBaseline(filepath.Join(path, "child"), map[finding]int{}); err == nil {
			t.Fatal("invalid output accepted")
		}
	})
}

// TestRun covers advisory output, opt-in accounting and command failures.
func TestRun(t *testing.T) {
	root := fixture(t, `func review() { log.Error(err) }`)
	path := filepath.Join(t.TempDir(), "baseline.json")
	var out, errOut bytes.Buffer
	invoke := func(args ...string) int { out.Reset(); errOut.Reset(); return run(args, &out, &errOut) }
	if code := invoke("-list", root); code != 0 || !strings.Contains(out.String(), "argument 1") || strings.Contains(out.String(), "log.Error(err)") {
		t.Fatalf("%d %s %s", code, &out, &errOut)
	}
	if code := invoke("-update", root); code != 2 {
		t.Fatal(code)
	}
	if code := invoke("-unknown"); code != 2 {
		t.Fatal(code)
	}
	if code := invoke("-h"); code != 0 {
		t.Fatal(code)
	}
	if code := invoke("-baseline", path, root); code != 2 {
		t.Fatal(code)
	}
	if code := invoke("-update", "-baseline", path, root); code != 0 {
		t.Fatalf("%d %s", code, &errOut)
	}
	if code := invoke("-baseline", path, root); code != 0 {
		t.Fatal(code)
	}
	writeFixture(t, filepath.Join(root, "calls.go"), `package fixture
func changed() { log.Error(other) }`)
	if code := invoke("-baseline", path, root); code != 1 || !strings.Contains(out.String(), "changed argument 1") {
		t.Fatalf("%d %s %s", code, &out, &errOut)
	}
	if code := invoke("-list", "-baseline", path, root); code != 1 {
		t.Fatal(code)
	}
	if code := invoke("-update", "-baseline", filepath.Join(path, "child"), root); code != 2 {
		t.Fatal(code)
	}
	if code := invoke(root + "missing"); code != 2 {
		t.Fatal(code)
	}
}

// TestQualified checks imported selectors and non-package expressions.
func TestQualified(t *testing.T) {
	imports := map[string]string{"c": modulePath + "pkg/clean"}
	for _, tc := range []struct{ input, want string }{{"c.Log", modulePath + "pkg/clean.Log"}, {"missing.Log", ""}, {"value", ""}, {"obj.child.Log", ""}} {
		e, err := parser.ParseExpr(tc.input)
		if err != nil {
			t.Fatal(err)
		}
		if got := qualified(e, imports); got != tc.want {
			t.Fatalf("%s: %q", tc.input, got)
		}
	}
}

// TestCredential distinguishes names from explicit redacted values.
func TestCredential(t *testing.T) {
	imports := map[string]string{"clean": modulePath + "pkg/clean"}
	for _, tc := range []struct {
		input string
		want  bool
	}{{"service.AccKey", true}, {"clean.Log(password)", true}, {"clean.UriRedacted(uri)", false}, {"clean.FileNameRedacted(token)", false}, {"clean.UriRedacted(uri)+token", true}, {"folder.Path", false}} {
		e, err := parser.ParseExpr(tc.input)
		if err != nil {
			t.Fatal(err)
		}
		if got := credential(e, imports); got != tc.want {
			t.Fatalf("%s: %v", tc.input, got)
		}
	}
}

// TestScanExplicitRoot keeps explicit source directories in scope regardless of spelling.
func TestScanExplicitRoot(t *testing.T) {
	root := fixture(t, `func review() { log.Error(err) }`)
	t.Chdir(root)
	for _, path := range []string{".", "./"} {
		found, err := scan([]string{path}, false)
		if err != nil || len(found) != 1 {
			t.Fatalf("%s: %+v %v", path, found, err)
		}
	}
	if err := os.Mkdir(".hidden", fs.ModeDir); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, ".hidden/calls.go", `package hidden
import "github.com/photoprism/photoprism/internal/event"
func review() { event.Log.Error(err) }`)
	found, err := scan([]string{"./.hidden"}, false)
	if err != nil || len(found) != 1 {
		t.Fatalf("%+v %v", found, err)
	}
	found, err = scan([]string{"."}, false)
	if err != nil || len(found) != 1 {
		t.Fatalf("hidden traversal: %+v %v", found, err)
	}
}

// TestBaselineOrdering makes serialized order independent of input order.
func TestBaselineOrdering(t *testing.T) {
	path := filepath.Join(t.TempDir(), "baseline.json")
	a := finding{File: "a.go", Function: "one", Call: `log.Error(err)`, Argument: 1, Rule: "rendering"}
	b := a
	b.File = "b.go"
	current := map[finding]int{b: 1, a: 2}
	if err := writeBaseline(path, current); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path) // #nosec G304 -- test reads its own temporary baseline.
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Index(data, []byte(`"a.go"`)) > bytes.Index(data, []byte(`"b.go"`)) {
		t.Fatal("not sorted")
	}
	got, err := readBaseline(path)
	if err != nil || len(got) != 2 || got[a] != 2 || got[b] != 1 {
		t.Fatalf("%+v %v", got, err)
	}
}

// TestOrdinary checks direct channel and method selection.
func TestOrdinary(t *testing.T) {
	imports := map[string]string{"e": modulePath + "internal/event"}
	for _, tc := range []struct {
		input string
		want  bool
	}{{"e.Log.Info(value)", true}, {"e.SystemLog.Info(value)", false}, {"e.Log.Debug(value)", false}, {"e.Log.Log(level, value)", true}, {"other.Info(value)", false}, {"print(value)", false}} {
		e, err := parser.ParseExpr(tc.input)
		if err != nil {
			t.Fatal(err)
		}
		call := e.(*ast.CallExpr)
		if got := ordinary(call, imports, map[string]bool{}); got != tc.want {
			t.Fatalf("%s: %v", tc.input, got)
		}
	}
}

// TestPrintFinding keeps source expressions out of routine output.
func TestPrintFinding(t *testing.T) {
	var out bytes.Buffer
	printFinding(&out, finding{File: "file.go", Line: 12, Function: "review", Call: `log.Info("example literal")`, Argument: 2, Rule: "rendering"})
	if got := out.String(); got != "file.go:12: review argument 2: rendering review\n" {
		t.Fatal(got)
	}
}

// TestUnparen retains direct calls inside parentheses.
func TestUnparen(t *testing.T) {
	root := fixture(t, `func review() { (log).Error(err); (event.Log).Error(err); log.Info((clean.Log)(name)) }`)
	found, err := scan([]string{root}, false)
	if err != nil || len(found) != 2 {
		t.Fatalf("%+v %v", found, err)
	}
	e, err := parser.ParseExpr("((name))")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := unparen(e).(*ast.Ident); !ok {
		t.Fatal("not unwrapped")
	}
}

// TestScanSymlinkRoot rejects a source root that would not be traversed.
func TestScanSymlinkRoot(t *testing.T) {
	root := fixture(t, `func review() { log.Error(err) }`)
	link := filepath.Join(t.TempDir(), "source")
	if err := os.Symlink(root, link); err != nil {
		t.Fatal(err)
	}
	if _, err := scan([]string{link}, false); err == nil {
		t.Fatal("symlink root accepted")
	}
}

// TestBaselinePrivatePublication preserves the link target and publishes private accounting.
func TestBaselinePrivatePublication(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	path := filepath.Join(root, "baseline.json")
	writeFixture(t, target, "unchanged")
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if err := writeBaseline(path, map[finding]int{}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target) // #nosec G304 -- test reads its own temporary target.
	if err != nil || string(data) != "unchanged" {
		t.Fatalf("%s %v", data, err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != fs.ModeSecretFile {
		t.Fatal(info.Mode())
	}
	if err := os.Chmod(path, fs.ModeFile); err != nil {
		t.Fatal(err)
	}
	if err := writeBaseline(path, map[finding]int{}); err != nil {
		t.Fatal(err)
	}
	info, err = os.Stat(path)
	if err != nil || info.Mode().Perm() != fs.ModeSecretFile {
		t.Fatalf("%v %v", info, err)
	}
	if err := writeBaseline(root, map[finding]int{}); err == nil {
		t.Fatal("directory replacement accepted")
	}
	leftovers, err := filepath.Glob(filepath.Join(root, ".log-rendering-*"))
	if err != nil || len(leftovers) != 0 {
		t.Fatalf("%v %v", leftovers, err)
	}
}

// TestScanNonRegularSource rejects child source links before parsing their targets.
func TestScanNonRegularSource(t *testing.T) {
	root := fixture(t, `func review() { log.Error(err) }`)
	if err := os.Symlink("calls.go", filepath.Join(root, "linked.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := scan([]string{root}, false); err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("link source: %v", err)
	}
}
