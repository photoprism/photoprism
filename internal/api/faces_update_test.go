package api

import (
	"net/http"
	"net/http/httptest"
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
)

// newUpdateFaceFixture creates a visible face without a person or markers, so a request can be
// checked for writes against a known state.
func newUpdateFaceFixture(t *testing.T) *entity.Face {
	t.Helper()

	f := &entity.Face{ID: strings.ToUpper(rnd.Base36(32)), FaceSrc: entity.SrcManual}
	require.NoError(t, f.Create())

	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Face{}, "id = ?", f.ID) })

	return f
}

// newUpdateFaceSubject creates a subject of the given type and removes it when the test ends.
func newUpdateFaceSubject(t *testing.T, name, subjType string) *entity.Subject {
	t.Helper()

	m := entity.NewSubject(name, subjType, entity.SrcManual)
	require.NotNil(t, m)
	require.NoError(t, m.Create())

	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Subject{}, "subj_uid = ?", m.SubjUID) })

	return m
}

// assertFaceUnchanged checks that the face still has no person and is not hidden.
func assertFaceUnchanged(t *testing.T, id string) {
	t.Helper()

	f := entity.FindFace(id)
	require.NotNil(t, f)
	assert.Empty(t, f.SubjUID, "a rejected request must not assign the person")
	assert.False(t, f.FaceHidden, "a rejected request must not change the visibility either")
}

// TestUpdateFace_Subject covers how UpdateFace resolves the person it assigns: an unknown, deleted,
// non-person, or withheld subject is not found, and nothing is written.
func TestUpdateFace_Subject(t *testing.T) {
	app, router, conf := NewApiTest()
	UpdateFace(router)

	prevAuthMode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(prevAuthMode) })

	admin, err := entity.AddClientSession("update-face-admin", conf.SessionMaxAge(), "*",
		authn.GrantPassword, entity.UserFixtures.Pointer("alice"))
	require.NoError(t, err)
	t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(admin) })
	require.True(t, admin.SeesPrivatePeople())

	// CE grants people only to roles that also see private people, so the test installs a role that
	// may update people without that access.
	role := acl.RoleGuest
	prevGrant, hadGrant := acl.Rules[acl.ResourcePeople][role]
	acl.Rules[acl.ResourcePeople][role] = acl.Grant{acl.ActionView: true, acl.ActionSearch: true, acl.ActionUpdate: true}
	t.Cleanup(func() {
		if hadGrant {
			acl.Rules[acl.ResourcePeople][role] = prevGrant
		} else {
			delete(acl.Rules[acl.ResourcePeople], role)
		}
	})

	restricted, err := entity.AddClientSession("update-face-restricted", conf.SessionMaxAge(), "*",
		authn.GrantPassword, entity.UserFixtures.Pointer("guest"))
	require.NoError(t, err)
	t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(restricted) })
	require.False(t, restricted.SeesPrivatePeople())

	// update sends the person together with a visibility change, so a partial write would show.
	update := func(faceId, subjUid, token string) *httptest.ResponseRecorder {
		return AuthenticatedRequestWithBody(app, http.MethodPut, "/api/v1/faces/"+faceId,
			`{"SubjUID": "`+subjUid+`", "Hidden": true}`, token)
	}

	// assertNotFound checks the answer given for every person the request may not assign.
	assertNotFound := func(t *testing.T, subjUid, token string) {
		t.Helper()

		f := newUpdateFaceFixture(t)
		r := update(f.ID, subjUid, token)

		assert.Equal(t, http.StatusNotFound, r.Code)
		assert.Equal(t, "Subject not found", gjson.Get(r.Body.String(), "error").String())
		assert.NotContains(t, r.Body.String(), subjUid)
		assertFaceUnchanged(t, f.ID)
	}

	t.Run("Unknown", func(t *testing.T) {
		assertNotFound(t, rnd.GenerateUID('j'), admin.AuthToken())
	})
	t.Run("Deleted", func(t *testing.T) {
		m := newUpdateFaceSubject(t, "Update Face Deleted", entity.SubjPerson)
		require.NoError(t, m.Delete())
		assertNotFound(t, m.SubjUID, admin.AuthToken())
	})
	t.Run("NotAPerson", func(t *testing.T) {
		m := newUpdateFaceSubject(t, "Update Face Object", "object")
		require.False(t, m.IsPerson())
		assertNotFound(t, m.SubjUID, admin.AuthToken())
	})
	t.Run("PrivateWithheld", func(t *testing.T) {
		m := newUpdateFaceSubject(t, "Update Face Private", entity.SubjPerson)
		markPrivate(t, m, false)
		assertNotFound(t, m.SubjUID, restricted.AuthToken())
	})
	t.Run("HiddenWithheld", func(t *testing.T) {
		m := newUpdateFaceSubject(t, "Update Face Hidden", entity.SubjPerson)
		markPrivate(t, m, true)
		assertNotFound(t, m.SubjUID, restricted.AuthToken())
	})
	t.Run("PrivateAcceptedForAdmin", func(t *testing.T) {
		m := newUpdateFaceSubject(t, "Update Face Private Admin", entity.SubjPerson)
		markPrivate(t, m, false)
		f := newUpdateFaceFixture(t)

		r := update(f.ID, m.SubjUID, admin.AuthToken())
		assert.Equal(t, http.StatusOK, r.Code)

		result := entity.FindFace(f.ID)
		require.NotNil(t, result)
		assert.Equal(t, m.SubjUID, result.SubjUID)
		assert.True(t, result.FaceHidden)
	})
	t.Run("CanonicalUID", func(t *testing.T) {
		// The face stores the UID of the person found, not the string the request sent.
		m := newUpdateFaceSubject(t, "Update Face Canonical", entity.SubjPerson)
		f := newUpdateFaceFixture(t)

		r := update(f.ID, " "+strings.ToUpper(m.SubjUID)+" ", admin.AuthToken())
		assert.Equal(t, http.StatusOK, r.Code)

		result := entity.FindFace(f.ID)
		require.NotNil(t, result)
		assert.Equal(t, m.SubjUID, result.SubjUID)
	})
	t.Run("HiddenOnly", func(t *testing.T) {
		// A request without a person changes the visibility alone, as the web UI sends it.
		f := newUpdateFaceFixture(t)

		r := AuthenticatedRequestWithBody(app, http.MethodPut, "/api/v1/faces/"+f.ID, `{"Hidden": true}`, restricted.AuthToken())
		assert.Equal(t, http.StatusOK, r.Code)

		result := entity.FindFace(f.ID)
		require.NotNil(t, result)
		assert.Empty(t, result.SubjUID)
		assert.True(t, result.FaceHidden)
	})
	t.Run("VisiblePerson", func(t *testing.T) {
		// The positive control for the restricted session: an ordinary person stays assignable.
		m := newUpdateFaceSubject(t, "Update Face Visible", entity.SubjPerson)
		f := newUpdateFaceFixture(t)

		r := update(f.ID, m.SubjUID, restricted.AuthToken())
		assert.Equal(t, http.StatusOK, r.Code)

		result := entity.FindFace(f.ID)
		require.NotNil(t, result)
		assert.Equal(t, m.SubjUID, result.SubjUID)
		assert.True(t, result.FaceHidden)
	})
}
