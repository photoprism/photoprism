package meta

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestXmpSecurity_OversizeFileRejected covers the size-cap branch of
// Load: a 1 MiB + 1 byte file must error with ErrXmpFileTooLarge
// before any parsing happens.
func TestXmpSecurity_OversizeFileRejected(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "oversize.xmp")
	payload := strings.Repeat("A", xmpMaxFileSize+1)
	if err := writeFile(t, tmp, payload); err != nil {
		t.Fatal(err)
	}
	var doc XmpDocument
	err := doc.Load(tmp)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrXmpFileTooLarge), "got %v, want ErrXmpFileTooLarge", err)
}

// TestXmpSecurity_DepthBombRejected covers the depth-cap branch:
// nesting one level deeper than xmpMaxDepth errors with ErrXmpTooDeep;
// documents exactly at the limit are still accepted.
func TestXmpSecurity_DepthBombRejected(t *testing.T) {
	t.Run("ExceedsLimit", func(t *testing.T) {
		nesting := xmpMaxDepth + 1
		body := strings.Repeat("<a>", nesting) + strings.Repeat("</a>", nesting)
		tmp := filepath.Join(t.TempDir(), "deep.xmp")
		if err := writeFile(t, tmp, `<?xml version="1.0"?>`+body); err != nil {
			t.Fatal(err)
		}
		var doc XmpDocument
		err := doc.Load(tmp)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrXmpTooDeep), "got %v, want ErrXmpTooDeep", err)
	})
	t.Run("AtLimitAllowed", func(t *testing.T) {
		// xmpMaxDepth-1 elements puts the deepest one at xmpMaxDepth
		// counting the document node — the highest accepted value.
		nesting := xmpMaxDepth - 1
		body := strings.Repeat("<a>", nesting) + strings.Repeat("</a>", nesting)
		tmp := filepath.Join(t.TempDir(), "at-limit.xmp")
		if err := writeFile(t, tmp, `<?xml version="1.0"?>`+body); err != nil {
			t.Fatal(err)
		}
		var doc XmpDocument
		err := doc.Load(tmp)
		assert.NoError(t, err)
	})
}

// TestXmpSecurity_XXENotResolved guards against a future loader change
// that switches to a less-defensive parser. encoding/xml does not
// resolve external entities by default; this asserts &xxe; is never
// expanded to the contents of /etc/hostname or any other file.
func TestXmpSecurity_XXENotResolved(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "xxe.xmp")
	body := `<?xml version="1.0"?>
<!DOCTYPE foo [ <!ENTITY xxe SYSTEM "file:///etc/hostname"> ]>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:dc="http://purl.org/dc/elements/1.1/">
   <dc:title>&xxe;</dc:title>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`
	if err := writeFile(t, tmp, body); err != nil {
		t.Fatal(err)
	}
	var doc XmpDocument
	if err := doc.Load(tmp); err != nil {
		// encoding/xml errors on unresolved entity references in strict
		// mode — acceptable, proves the entity was not fetched.
		t.Logf("Load rejected XXE document with %v", err)
		return
	}
	// If Load succeeded, the title must be empty or the literal "&xxe;";
	// anything resembling a hostname or path is a regression.
	title := doc.Title()
	assert.NotContains(t, title, "/etc/")
	if title != "" && title != "&xxe;" {
		t.Errorf("unexpected Title %q after XXE attempt — entity may have resolved", title)
	}
}
