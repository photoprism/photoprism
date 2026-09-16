package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/time/unix"
)

// filteredSessionToken returns the token of a session the people visibility filter applies to:
// full access to photos and no private access to people. CE reaches that combination with a portal
// credential acting for an account, since every CE account role is either full-access or
// shared-only.
func filteredSessionToken(t *testing.T, conf *config.Config) string {
	t.Helper()

	user := entity.UserFixtures.Pointer("alice")

	client := &entity.Client{
		ClientUID:    rnd.GenerateUID(entity.ClientUID),
		UserUID:      user.UserUID,
		UserName:     user.UserName,
		ClientName:   "private-people-probe",
		ClientRole:   acl.RolePortal.String(),
		ClientType:   authn.ClientConfidential,
		AuthProvider: authn.ProviderClient.String(),
		AuthMethod:   authn.MethodOAuth2.String(),
		AuthScope:    "*",
		AuthExpires:  unix.Hour,
		AuthTokens:   1,
		AuthEnabled:  true,
	}

	require.NoError(t, client.Create())
	t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(client) })

	sess := entity.NewSession(conf.SessionMaxAge(), 0)
	sess.SetClient(client)

	require.NoError(t, sess.Create())
	t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(sess) })

	require.False(t, sess.SeesPrivatePeople(), "the portal role is absent from the people table")
	require.False(t, sess.HasSharedAccessOnly(acl.ResourcePhotos))
	require.True(t, sess.IsRegistered())

	return sess.AuthToken()
}

// markPrivate withholds the given person's name for the duration of the test, through whichever
// of the two flags the case asks for.
func markPrivate(t *testing.T, subj *entity.Subject, hidden bool) {
	t.Helper()

	attr := "SubjPrivate"

	if hidden {
		attr = "SubjHidden"
	}

	require.NoError(t, subj.Update(attr, true))
	t.Cleanup(func() { _ = subj.Update(attr, false) })
}

// TestGetPhoto_OmitsWithheldPeople is the named regression for GET /api/v1/photos/{uid}: a session
// that may not see a withheld person receives the picture without that person's markers, and with
// everyone else's.
func TestGetPhoto_OmitsWithheldPeople(t *testing.T) {
	app, router, conf := NewApiTest()
	GetPhoto(router)

	// Public mode resolves every request to the admin visitor, which would hide the filter.
	prevAuthMode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(prevAuthMode) })

	token := filteredSessionToken(t, conf)

	// This picture names two people and also carries markers with no subject at all.
	private := entity.SubjectFixtures.Pointer("actress-1")
	public := entity.SubjectFixtures.Pointer("actor-1")

	// subjects returns the subject uid of every marker the response carries for the picture.
	subjects := func(t *testing.T) []string {
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/photos/ps6sg6be2lvl0yh7", token)
		require.Equal(t, http.StatusOK, r.Code)

		var out []string

		for _, file := range gjson.Get(r.Body.String(), "Files").Array() {
			for _, marker := range file.Get("Markers").Array() {
				out = append(out, marker.Get("SubjUID").String())
			}
		}

		return out
	}

	before := subjects(t)
	require.Contains(t, before, private.SubjUID, "the fixture has to name the person for the case to mean anything")
	require.Contains(t, before, public.SubjUID)

	markPrivate(t, private, false)

	after := subjects(t)
	assert.NotContains(t, after, private.SubjUID)
	assert.Contains(t, after, public.SubjUID, "the other person is kept, so an empty list is a failure")
	assert.Less(t, len(after), len(before), "only that person's markers are left out")
}

// TestSearchFolders_RequiresWholeLibraryAccess pins the gate on the directory listing. It names
// every file in the library, and its response is cached across sessions, so library reach alone is
// not enough - the File Browser, Index and Import pages all ask for more before rendering it.
func TestSearchFolders_RequiresWholeLibraryAccess(t *testing.T) {
	t.Run("Grants", func(t *testing.T) {
		assert.True(t, acl.Rules.Allow(acl.ResourceFiles, acl.RoleAdmin, acl.AccessAll))
		assert.True(t, acl.Rules.Allow(acl.ResourceFiles, acl.RoleClient, acl.AccessAll))
		// Shared-only roles reach neither, so the listing was never theirs.
		assert.False(t, acl.Rules.Allow(acl.ResourceFiles, acl.RoleGuest, acl.AccessAll))
		assert.False(t, acl.Rules.Allow(acl.ResourceFiles, acl.RoleVisitor, acl.AccessAll))
	})

	app, router, conf := NewApiTest()
	SearchFoldersOriginals(router)

	prevAuthMode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(prevAuthMode) })

	t.Run("Admin", func(t *testing.T) {
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/folders/originals?files=true", AuthenticateAdmin(app, router))
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("Unauthenticated", func(t *testing.T) {
		r := PerformRequest(app, http.MethodGet, "/api/v1/folders/originals?files=true")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
}

// TestWriteHandlers_OmitWithheldPeople pins that a write response is shaped like the by-uid read.
// Six handlers answer with the edited picture, and a credential can be admitted on photos while
// people is outside its scope, so the session that may not see a person can still reach them.
//
// It creates its own markers on a picture no other test in this package reads, and restores what it
// writes: a fixture picture mutated here is a later test's failure.
func TestWriteHandlers_OmitWithheldPeople(t *testing.T) {
	const (
		photoUID = "ps6sg6be2lvl0y14" // Photo07, unreferenced elsewhere in this package
		fileUID  = "fs6sg6bqhhinlplr" // its primary file
		label    = "Withheld Write Probe"
	)

	app, router, conf := NewApiTest()
	UpdatePhoto(router)
	PhotoPrimary(router)
	AddPhotoLabel(router)
	UpdatePhotoLabel(router)
	RemovePhotoLabel(router)

	prevAuthMode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(prevAuthMode) })

	// An app password of an admin account, scoped to photos alone.
	sess, err := entity.AddClientSession("withheld-write-probe", conf.SessionMaxAge(), "photos",
		authn.GrantPassword, entity.UserFixtures.Pointer("alice"))
	require.NoError(t, err)
	t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(sess) })

	require.False(t, sess.InsufficientScope(acl.ResourcePhotos, acl.Permissions{acl.ActionUpdate}),
		"the scope admits it on photos")
	require.False(t, sess.SeesPrivatePeople(), "and people is outside that scope")

	withheld := entity.SubjectFixtures.Pointer("actress-1")
	public := entity.SubjectFixtures.Pointer("actor-1")

	for i, subj := range []*entity.Subject{withheld, public} {
		m := entity.Marker{
			FileUID: fileUID, MarkerType: entity.MarkerFace, MarkerSrc: entity.SrcImage,
			Thumb:   fmt.Sprintf("writeprobe%d", i),
			SubjUID: subj.SubjUID, SubjSrc: entity.SrcManual, MarkerName: subj.SubjName,
			X: float32(i+1) / 10, Y: 0.4, W: 0.1, H: 0.1, Size: 200, Score: 80,
		}

		created, createErr := entity.CreateMarkerIfNotExists(&m)
		require.NoError(t, createErr)

		uid := created.MarkerUID

		t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(&entity.Marker{}, "marker_uid = ?", uid) })
	}

	// Restore the picture and drop the probe label, so nothing downstream reads this state.
	photo := entity.Photo{PhotoUID: photoUID}
	require.NoError(t, entity.UnscopedDb().First(&photo, "photo_uid = ?", photoUID).Error)

	wasFavorite := photo.PhotoFavorite

	t.Cleanup(func() {
		_ = entity.UnscopedDb().Model(&entity.Photo{}).Where("photo_uid = ?", photoUID).
			UpdateColumn("photo_favorite", wasFavorite).Error
		entity.UnscopedDb().Unscoped().
			Exec("DELETE FROM photos_labels WHERE photo_id = ? AND label_id IN (SELECT id FROM labels WHERE label_name = ?)",
				photo.ID, label)
		entity.UnscopedDb().Unscoped().Delete(&entity.Label{}, "label_name = ?", label)
	})

	subjects := func(t *testing.T, body string) []string {
		var out []string

		for _, f := range gjson.Get(body, "Files").Array() {
			for _, m := range f.Get("Markers").Array() {
				if uid := m.Get("SubjUID").String(); uid != "" {
					out = append(out, uid)
				}
			}
		}

		return out
	}

	before := func(t *testing.T) []string {
		r := AuthenticatedRequestWithBody(app, http.MethodPut, "/api/v1/photos/"+photoUID, `{"Favorite": true}`, sess.AuthToken())
		require.Equal(t, http.StatusOK, r.Code)

		return subjects(t, r.Body.String())
	}(t)

	require.Contains(t, before, withheld.SubjUID, "the markers this test created have to arrive")
	require.Contains(t, before, public.SubjUID)

	markPrivate(t, withheld, false)

	// Run in order: the label handlers need the label the first one adds, and each asserts the same
	// thing about its own response. Left out and covered structurally instead, by
	// TestPhotoResponses_AllRedact: three that rewrite a file header, unstack a picture or delete a
	// file, and the handlers loading through query.PhotoByUID, whose response carries no files and
	// so no marker to withhold.
	steps := []struct {
		name, method, path, body string
	}{
		{"UpdatePhoto", http.MethodPut, "/api/v1/photos/" + photoUID, `{"Favorite": true}`},
		{"PhotoPrimary", http.MethodPost, "/api/v1/photos/" + photoUID + "/files/" + fileUID + "/primary", ""},
		{"AddPhotoLabel", http.MethodPost, "/api/v1/photos/" + photoUID + "/label", `{"Name": "` + label + `", "Uncertainty": 25}`},
		{"UpdatePhotoLabel", http.MethodPut, "/api/v1/photos/" + photoUID + "/label/%d", `{"Uncertainty": 30}`},
		{"RemovePhotoLabel", http.MethodDelete, "/api/v1/photos/" + photoUID + "/label/%d", ""},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			path := step.path

			if strings.Contains(path, "%d") {
				l := entity.Label{}
				require.NoError(t, entity.UnscopedDb().First(&l, "label_name = ?", label).Error)
				path = fmt.Sprintf(path, l.ID)
			}

			r := AuthenticatedRequestWithBody(app, step.method, path, step.body, sess.AuthToken())
			require.Equal(t, http.StatusOK, r.Code, r.Body.String())

			after := subjects(t, r.Body.String())
			assert.NotContains(t, after, withheld.SubjUID)
			assert.Contains(t, after, public.SubjUID, "the other person is kept, so an empty list is a failure")
		})
	}
}

// Recognizing a picture response by provenance rather than by shape: a handler that loads a picture
// and then serializes anything mentioning it has to shape it first, whatever the variable is called
// and whether the payload is the entity or a map wrapping it.
//
// Three boundaries of this approach, stated so they are not discovered later:
//
//   - The scope is the top-level function. Every registrar today registers one route, so
//     per-function and per-handler coincide; one registering two picture routes would let one
//     handler's call vouch for the other's response.
//   - The operand scan reads string literals too, so a handler loading into a variable named after
//     a map key in the same payload is flagged by the key alone. That fails loudly rather than
//     hiding anything.
//   - It recognizes a picture by the loaders below. One obtained another way - entity.FindPhoto, a
//     helper, a search result - is invisible to both the pattern and the floor.
var (
	// pictureLoader matches the assignment that gives a handler its picture.
	pictureLoader = regexp.MustCompile(`(\w+)\s*,?\s*(?:\w+\s*)?:?=\s*query\.Photo(?:PreloadBy)?(?:By)?UID\(`)

	// picturePersist matches a call that writes the picture somewhere durable, which the shaping
	// has to follow rather than precede.
	picturePersist = regexp.MustCompile(`SaveSidecarYaml\(&(\w+)\)|(\w+)\.(?:Save|Updates?)\(`)

	// jsonResponse matches a line that writes a JSON response.
	jsonResponse = regexp.MustCompile(`c\.(?:Indented)?JSON\(`)

	// funcStart matches a top-level function, which is where the scan resets.
	funcStart = regexp.MustCompile(`^func `)

	// pictureOperand matches an identifier used as a value rather than as a receiver, so a derived
	// payload such as m.Links() is not mistaken for the picture itself.
	pictureOperand = regexp.MustCompile(`\b([A-Za-z_]\w*)\b\.?`)
)

// pictureFinding names a response the rule rejects, and why.
type pictureFinding struct {
	Line   int
	Reason string
}

// pictureOperands returns the identifiers a line passes by value, dropping those it only calls a
// method or reads a field on.
func pictureOperands(line string) (names []string) {
	for _, m := range pictureOperand.FindAllStringSubmatch(line, -1) {
		if strings.HasSuffix(m[0], ".") {
			continue
		}

		names = append(names, m[1])
	}

	return names
}

// unshapedPictureResponses returns what the rule rejects in src: a picture serialized without being
// shaped for the session, and a shaping placed above a call that persists the same picture - which
// would write the reduced entity to a sidecar rather than only to the response.
func unshapedPictureResponses(src string) (findings []pictureFinding) {
	all := strings.Split(src, "\n")

	// Function boundaries, so each response is judged against its own function.
	starts := []int{0}

	for i, line := range all {
		if funcStart.MatchString(line) {
			starts = append(starts, i)
		}
	}

	starts = append(starts, len(all))

	for b := 0; b < len(starts)-1; b++ {
		body := all[starts[b]:starts[b+1]]
		offset := starts[b]

		// Which variables in this function hold a picture.
		loaded := map[string]bool{}

		for _, line := range body {
			for _, m := range pictureLoader.FindAllStringSubmatch(line, -1) {
				loaded[m[1]] = true
			}
		}

		if len(loaded) == 0 {
			continue
		}

		// Where each picture is shaped, if it is.
		shapedAt := map[string]int{}

		for i, line := range body {
			for name := range loaded {
				if _, seen := shapedAt[name]; !seen && strings.Contains(line, name+".RedactForSession") {
					shapedAt[name] = i
				}
			}
		}

		served := map[string]bool{}

		for i, line := range body {
			if !jsonResponse.MatchString(line) {
				continue
			}

			for _, name := range pictureOperands(line) {
				if !loaded[name] {
					continue
				}

				served[name] = true

				if _, shaped := shapedAt[name]; !shaped {
					findings = append(findings, pictureFinding{offset + i + 1,
						"serializes a picture without shaping it for the session"})
				}

				break
			}
		}

		// A shaping above a write persists the reduced picture, not just the response.
		for name, at := range shapedAt {
			if !served[name] {
				continue
			}

			for i := at + 1; i < len(body); i++ {
				for _, m := range picturePersist.FindAllStringSubmatch(body[i], -1) {
					if m[1] != name && m[2] != name {
						continue
					}

					findings = append(findings, pictureFinding{offset + at + 1,
						"shapes a picture above a call that persists it, so the reduction reaches storage"})

					i = len(body)

					break
				}
			}
		}
	}

	return findings
}

// TestUnshapedPictureResponses covers the check below, which cannot be proven by mutating the
// package: it reads the source at run time, so a compile-time overlay does not reach it.
func TestUnshapedPictureResponses(t *testing.T) {
	reasons := func(src string) []string {
		var out []string

		for _, f := range unshapedPictureResponses(src) {
			out = append(out, fmt.Sprintf("%d:%s", f.Line, f.Reason))
		}

		return out
	}

	t.Run("Shaped", func(t *testing.T) {
		assert.Empty(t, reasons(`func h() {
		p, err := query.PhotoPreloadByUID(uid)
		p.RedactForSession(s)
		c.JSON(http.StatusOK, p)
}`))
	})
	t.Run("Unshaped", func(t *testing.T) {
		assert.Equal(t, []string{"4:serializes a picture without shaping it for the session"},
			reasons(`func h() {
		p, err := query.PhotoPreloadByUID(uid)
		PublishPhotoEvent(StatusUpdated, uid)
		c.JSON(http.StatusOK, p)
}`))
	})
	// The two shapes the previous pattern missed: any variable name, and a wrapped payload.
	t.Run("WrappedPayload", func(t *testing.T) {
		assert.NotEmpty(t, reasons(`func h() {
		m, err := query.PhotoByUID(id)
		c.JSON(http.StatusOK, gin.H{"photo": m})
}`))
		assert.Empty(t, reasons(`func h() {
		m, err := query.PhotoByUID(id)
		m.RedactForSession(s)
		c.JSON(http.StatusOK, gin.H{"photo": m})
}`))
	})
	t.Run("Indented", func(t *testing.T) {
		assert.NotEmpty(t, reasons(`func h() {
		p, err := query.PhotoPreloadByUID(uid)
		c.IndentedJSON(http.StatusOK, p)
}`))
	})
	// Position, not just presence: a reduced picture must not reach the sidecar on disk.
	t.Run("ShapedAboveAWrite", func(t *testing.T) {
		assert.Equal(t, []string{"3:shapes a picture above a call that persists it, so the reduction reaches storage"},
			reasons(`func h() {
		p, err := query.PhotoPreloadByUID(uid)
		p.RedactForSession(s)
		SaveSidecarYaml(&p)
		c.JSON(http.StatusOK, p)
}`))
		assert.Empty(t, reasons(`func h() {
		p, err := query.PhotoPreloadByUID(uid)
		SaveSidecarYaml(&p)
		p.RedactForSession(s)
		c.JSON(http.StatusOK, p)
}`), "the order every handler uses")
	})
	t.Run("ShapedAboveASave", func(t *testing.T) {
		assert.NotEmpty(t, reasons(`func h() {
		m, err := query.PhotoByUID(id)
		m.RedactForSession(s)
		m.Save()
		c.JSON(http.StatusOK, gin.H{"photo": m})
}`))
	})
	// Judged per function, so one handler's call cannot vouch for another's response.
	t.Run("PerFunction", func(t *testing.T) {
		assert.Equal(t, []string{"8:serializes a picture without shaping it for the session"},
			reasons(`func a() {
		p, err := query.PhotoByUID(uid)
		p.RedactForSession(s)
		c.JSON(http.StatusOK, p)
}
func b() {
		p, err := query.PhotoByUID(uid)
		c.JSON(http.StatusOK, p)
}`))
	})
	// A derived payload is not the picture: this serializes the share links, not the entity.
	t.Run("DerivedPayload", func(t *testing.T) {
		assert.Empty(t, reasons(`func h() {
		m, err := query.PhotoByUID(uid)
		c.JSON(http.StatusOK, m.Links())
}`))
	})
	t.Run("NoPictureLoaded", func(t *testing.T) {
		assert.Empty(t, reasons(`func h() {
		m := entity.Marker{}
		c.JSON(http.StatusOK, m)
}`))
	})
}

// TestPhotoResponses_AllRedact pins the rule rather than each handler: a handler that loads a
// picture shapes it for the session before serializing it. It is what covers ChangeFileOrientation,
// PhotoUnstack and DeleteFile, which a behavioral test cannot drive without rewriting a file header,
// unstacking a picture or deleting a file - and what fails when a handler is added later without the
// call.
//
// It recognizes a response by where the picture came from, so a wrapped payload or a differently
// named variable is still checked. It reads source text, so it proves the call is present rather
// than that it works; the subtests of TestWriteHandlers_OmitWithheldPeople prove the latter.
func TestPhotoResponses_AllRedact(t *testing.T) {
	files, err := filepath.Glob("*.go")
	require.NoError(t, err)

	checked := 0

	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		src, readErr := os.ReadFile(name) //nolint:gosec // a package source file this test globbed
		require.NoError(t, readErr)

		checked += len(pictureLoader.FindAllString(string(src), -1))

		for _, f := range unshapedPictureResponses(string(src)) {
			t.Errorf("%s:%d %s", name, f.Line, f.Reason)
		}
	}

	// Counts the loads rather than the responses, so the floor moves when handlers stop obtaining a
	// picture through the loaders above. It does not move for a handler that obtains one some other
	// way, which the existing loads keep the count above - that boundary is the doc comment's third.
	assert.GreaterOrEqual(t, checked, 20, "every handler that loads a picture is checked")
}
