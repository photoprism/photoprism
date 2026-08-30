package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestGetSubject(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetSubject(router)
		r := PerformRequest(app, "GET", "/api/v1/subjects/js6sg6b1h1njaaaa")
		val := gjson.Get(r.Body.String(), "Slug")
		assert.Equal(t, "dangling-subject", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetSubject(router)
		r := PerformRequest(app, "GET", "/api/v1/subjects/xxx1y111h1njaaaa")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Subject not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

// TestGetSubjectDeleted pins that a soft-deleted person reads as not found.
func TestGetSubjectDeleted(t *testing.T) {
	app, router, _ := NewApiTest()
	GetSubject(router)

	m := entity.NewSubject("Deleted Get Subject", entity.SubjPerson, entity.SrcManual)
	require.NotNil(t, m)
	require.NoError(t, m.Create())

	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Subject{}, "subj_uid = ?", m.SubjUID) })

	r := PerformRequest(app, "GET", "/api/v1/subjects/"+m.SubjUID)
	assert.Equal(t, http.StatusOK, r.Code)

	require.NoError(t, m.Delete())

	r = PerformRequest(app, "GET", "/api/v1/subjects/"+m.SubjUID)
	assert.Equal(t, http.StatusNotFound, r.Code)
}

// TestFindSubjectForSession covers the guard the four subject handlers share. Exercised against the
// helper because CE grants people to admin and client only, both of which hold private access, so
// the deny side has no CE role that reaches it through a handler.
func TestFindSubjectForSession(t *testing.T) {
	m := entity.NewSubject("Private Handler Subject", entity.SubjPerson, entity.SrcManual)
	require.NotNil(t, m)
	m.SubjPrivate = true
	require.NoError(t, m.Create())

	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Subject{}, "subj_uid = ?", m.SubjUID) })

	t.Run("PrivateDenied", func(t *testing.T) {
		assert.Nil(t, FindSubjectForSession(m.SubjUID, entity.SessionFixtures.Pointer("visitor")))
	})
	t.Run("PrivateAllowed", func(t *testing.T) {
		assert.NotNil(t, FindSubjectForSession(m.SubjUID, entity.SessionFixtures.Pointer("alice")))
	})
	t.Run("Deleted", func(t *testing.T) {
		// Refused for everyone, including the role that may see a private person.
		require.NoError(t, m.Delete())
		assert.Nil(t, FindSubjectForSession(m.SubjUID, entity.SessionFixtures.Pointer("alice")))
	})
	t.Run("NotFound", func(t *testing.T) {
		assert.Nil(t, FindSubjectForSession("js6sg6b1qekk9jzz", entity.SessionFixtures.Pointer("alice")))
	})
}

// TestSubjectHandlersRefuseDeleted pins the guard at the handlers rather than at the helper, since
// a read that refuses and three writes that do not is the state this exists to prevent. Deleted is
// the case CE can reach: both roles holding people also hold private access.
func TestSubjectHandlersRefuseDeleted(t *testing.T) {
	app, router, _ := NewApiTest()

	GetSubject(router)
	UpdateSubject(router)
	LikeSubject(router)
	DislikeSubject(router)

	m := entity.NewSubject("Deleted Handler Subject", entity.SubjPerson, entity.SrcManual)
	require.NotNil(t, m)
	require.NoError(t, m.Create())

	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Subject{}, "subj_uid = ?", m.SubjUID) })
	require.NoError(t, m.Delete())

	uid := m.SubjUID

	t.Run("Get", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, PerformRequest(app, "GET", "/api/v1/subjects/"+uid).Code)
	})
	t.Run("Update", func(t *testing.T) {
		r := PerformRequestWithBody(app, "PUT", "/api/v1/subjects/"+uid, `{"Favorite": true}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("Like", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, PerformRequest(app, "POST", "/api/v1/subjects/"+uid+"/like").Code)
	})
	t.Run("Dislike", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, PerformRequest(app, "DELETE", "/api/v1/subjects/"+uid+"/like").Code)
	})
}

func TestLikeSubject(t *testing.T) {
	t.Run("InvalidSubject", func(t *testing.T) {
		app, router, _ := NewApiTest()
		LikeSubject(router)
		r := PerformRequest(app, "POST", "/api/v1/subjects/8775789/like")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("ExistingSubject", func(t *testing.T) {
		app, router, _ := NewApiTest()

		// Register routes.
		GetSubject(router)
		LikeSubject(router)

		r := PerformRequest(app, "GET", "/api/v1/subjects/js6sg6b2h8njw0sx")
		t.Log(r.Body.String())
		val := gjson.Get(r.Body.String(), "Slug")
		assert.Equal(t, "joe-biden", val.String())
		val2 := gjson.Get(r.Body.String(), "Favorite")
		assert.Equal(t, "false", val2.String())

		r2 := PerformRequest(app, "POST", "/api/v1/subjects/js6sg6b2h8njw0sx/like")
		t.Log(r2.Body.String())
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/subjects/js6sg6b2h8njw0sx")
		t.Log(r3.Body.String())
		val3 := gjson.Get(r3.Body.String(), "Slug")
		assert.Equal(t, "joe-biden", val3.String())
		val4 := gjson.Get(r3.Body.String(), "Favorite")
		assert.Equal(t, "true", val4.String())
	})
}

func TestDislikeSubject(t *testing.T) {
	t.Run("InvalidSubject", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DislikeSubject(router)
		r := PerformRequest(app, "DELETE", "/api/v1/subjects/8775789/like")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("ExistingSubject", func(t *testing.T) {
		app, router, _ := NewApiTest()

		// Register routes.
		GetSubject(router)
		DislikeSubject(router)

		r := PerformRequest(app, "GET", "/api/v1/subjects/js6sg6b1qekk9jx8")
		t.Log(r.Body.String())
		val := gjson.Get(r.Body.String(), "Slug")
		assert.Equal(t, "john-doe", val.String())
		val2 := gjson.Get(r.Body.String(), "Favorite")
		assert.Equal(t, "true", val2.String())

		r2 := PerformRequest(app, "DELETE", "/api/v1/subjects/js6sg6b1qekk9jx8/like")
		t.Log(r2.Body.String())
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/subjects/js6sg6b1qekk9jx8")
		t.Log(r3.Body.String())
		val3 := gjson.Get(r3.Body.String(), "Slug")
		assert.Equal(t, "john-doe", val3.String())
		val4 := gjson.Get(r3.Body.String(), "Favorite")
		assert.Equal(t, "false", val4.String())
	})
}

func TestUpdateSubject(t *testing.T) {
	t.Run("SuccessfulRequestPerson", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateSubject(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/subjects/js6sg6b1qekk9jx8", `{"Name": "Updated Name"}`)
		val := gjson.Get(r.Body.String(), "Name")
		assert.Equal(t, "Updated Name", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateSubject(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/subjects/js6sg6b1qekk9jx8", `{"Name": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateSubject(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/subjectss/xxx", `{"Name": "Updated Name"}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("SetVerified", func(t *testing.T) {
		// The flag has to round-trip through the API, because that is the only way a person can
		// set it: nothing automatic may, or it stops meaning that somebody vouched for the name.
		app, router, _ := NewApiTest()

		SearchSubjects(router)
		UpdateSubject(router)

		const uid = "js6sg6b1qekk9jx8"

		r := PerformRequestWithBody(app, "PUT", "/api/v1/subjects/"+uid, `{"Verified": true}`)
		assert.Equal(t, http.StatusOK, r.Code)
		assert.True(t, gjson.Get(r.Body.String(), "Verified").Bool())

		// And the list the people views read has to carry it, or the checkbox cannot show it.
		s := PerformRequest(app, "GET", "/api/v1/subjects?count=100&uid="+uid)
		assert.Equal(t, http.StatusOK, s.Code)
		assert.True(t, gjson.Get(s.Body.String(), "0.Verified").Bool())

		r = PerformRequestWithBody(app, "PUT", "/api/v1/subjects/"+uid, `{"Verified": false}`)
		assert.Equal(t, http.StatusOK, r.Code)
		assert.False(t, gjson.Get(r.Body.String(), "Verified").Bool())
	})
	t.Run("SetBirthday", func(t *testing.T) {
		// The date is entered here and nowhere else, so the round trip through the API and the list
		// the people views read is the whole feature - a value the search omits cannot be edited.
		app, router, _ := NewApiTest()

		SearchSubjects(router)
		UpdateSubject(router)

		const uid = "js6sg6b1qekk9jx8"

		r := PerformRequestWithBody(app, "PUT", "/api/v1/subjects/"+uid, `{"Birthday": "1990-08-01T00:00:00Z"}`)
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "1990-08-01T00:00:00Z", gjson.Get(r.Body.String(), "Birthday").String())

		s := PerformRequest(app, "GET", "/api/v1/subjects?count=100&uid="+uid)
		assert.Equal(t, http.StatusOK, s.Code)
		assert.Equal(t, "1990-08-01T00:00:00Z", gjson.Get(s.Body.String(), "0.Birthday").String())

		// Replacing one date with another, which setting the first cannot reach: the form is built
		// from the subject, so both have to own their value for the update to be seen.
		r = PerformRequestWithBody(app, "PUT", "/api/v1/subjects/"+uid, `{"Birthday": "1991-09-02T00:00:00Z"}`)
		assert.Equal(t, http.StatusOK, r.Code)

		s = PerformRequest(app, "GET", "/api/v1/subjects?count=100&uid="+uid)
		assert.Equal(t, http.StatusOK, s.Code)
		assert.Equal(t, "1991-09-02T00:00:00Z", gjson.Get(s.Body.String(), "0.Birthday").String())

		// A date entered for the wrong person has to be removable again.
		r = PerformRequestWithBody(app, "PUT", "/api/v1/subjects/"+uid, `{"Birthday": null}`)
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "", gjson.Get(r.Body.String(), "Birthday").String())

		// Present as null rather than left out. The client tracks the keys a response carried, so an
		// omitted one is never sent back and the next date entered would go nowhere. Exists() is the
		// discriminator: true for an explicit null, false for an absent key.
		assert.True(t, gjson.Get(r.Body.String(), "Birthday").Exists())

		s = PerformRequest(app, "GET", "/api/v1/subjects?count=100&uid="+uid)
		assert.Equal(t, http.StatusOK, s.Code)
		assert.True(t, gjson.Get(s.Body.String(), "0.Birthday").Exists())
	})
	t.Run("ImplausibleBirthday", func(t *testing.T) {
		// A value the client has to correct, so it answers 400: reporting it as a server fault would
		// bury real ones in the logs and tell the user nothing they can act on.
		app, router, _ := NewApiTest()
		UpdateSubject(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/subjects/js6sg6b1qekk9jx8", `{"Birthday": "2999-01-01T00:00:00Z"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
		r = PerformRequestWithBody(app, "PUT", "/api/v1/subjects/js6sg6b1qekk9jx8", `{"Birthday": "0190-01-01T00:00:00Z"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("SetManualCover", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetSubject(router)
		UpdateSubject(router)

		const hash = "6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c-0910162fd2fd"
		r := PerformRequestWithBody(app, "PUT", "/api/v1/subjects/js6sg6b2h8njw0sx", `{"Thumb":"`+hash+`","ThumbSrc":"manual"}`)
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, hash, gjson.Get(r.Body.String(), "Thumb").String())
		assert.Equal(t, "manual", gjson.Get(r.Body.String(), "ThumbSrc").String())
	})
}
