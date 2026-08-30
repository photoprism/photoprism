package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
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

		// Replacing one date with another, which is not the same path as setting the first: the form
		// is built from the subject, so a shared pointer would let the request edit the subject in
		// place and the update would compare equal to itself and write nothing.
		r = PerformRequestWithBody(app, "PUT", "/api/v1/subjects/"+uid, `{"Birthday": "1991-09-02T00:00:00Z"}`)
		assert.Equal(t, http.StatusOK, r.Code)

		s = PerformRequest(app, "GET", "/api/v1/subjects?count=100&uid="+uid)
		assert.Equal(t, http.StatusOK, s.Code)
		assert.Equal(t, "1991-09-02T00:00:00Z", gjson.Get(s.Body.String(), "0.Birthday").String())

		// A date entered for the wrong person has to be removable again.
		r = PerformRequestWithBody(app, "PUT", "/api/v1/subjects/"+uid, `{"Birthday": null}`)
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "", gjson.Get(r.Body.String(), "Birthday").String())
	})
	t.Run("ImplausibleBirthday", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateSubject(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/subjects/js6sg6b1qekk9jx8", `{"Birthday": "2999-01-01T00:00:00Z"}`)
		assert.Equal(t, http.StatusInternalServerError, r.Code)
		r = PerformRequestWithBody(app, "PUT", "/api/v1/subjects/js6sg6b1qekk9jx8", `{"Birthday": "0190-01-01T00:00:00Z"}`)
		assert.Equal(t, http.StatusInternalServerError, r.Code)
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
