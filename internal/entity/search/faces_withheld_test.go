package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

// newWithheldSearchSubject adds a person whose name is withheld, since no fixture carries either
// flag.
func newWithheldSearchSubject(t *testing.T, name string, hidden bool) *entity.Subject {
	t.Helper()

	m := entity.NewSubject(name, entity.SubjPerson, entity.SrcManual)
	require.NotNil(t, m)

	if hidden {
		m.SubjHidden = true
	} else {
		m.SubjPrivate = true
	}

	require.NoError(t, m.Create())

	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.Subject{}, "subj_uid = ?", m.SubjUID)
		entity.SubjNames.Unset(m.SubjUID)
	})

	return m
}

// newNamedSearchFace saves a cluster assigned to the given person, plus a marker large enough to
// represent it, and returns the cluster.
func newNamedSearchFace(t *testing.T, subj *entity.Subject) *entity.Face {
	t.Helper()

	f := newSearchFace(t)

	require.NoError(t, f.Update("SubjUID", subj.SubjUID))

	f.SubjUID = subj.SubjUID

	newSearchMarker(t, f.ID, entity.Marker{Size: 300, Score: 90, FaceDist: 0.2,
		SubjUID: subj.SubjUID, SubjSrc: entity.SrcManual, MarkerName: subj.SubjName})

	return f
}

// filteredSession returns a session the people visibility filter applies to: full access to
// photos and no private access to people. CE reaches it with a portal credential for an account.
func filteredSession() *entity.Session {
	s := &entity.Session{}
	s.SetClient(&entity.Client{ClientRole: acl.RolePortal.String(), AuthProvider: authn.ProviderClient.String()})
	s.SetUser(entity.UserFixtures.Pointer("alice"))

	return s
}

// holdsFace reports whether the results include the given cluster.
func holdsFace(results FaceResults, faceID string) bool {
	for i := range results {
		if results[i].ID == faceID {
			return true
		}
	}

	return false
}

// TestUserFaces_OmitsWithheldPeople is the named regression for GET /api/v1/faces. Every case
// keeps a visible cluster beside the withheld one, so an empty result cannot pass as a filtered
// one.
func TestUserFaces_OmitsWithheldPeople(t *testing.T) {
	private := newWithheldSearchSubject(t, "Private Priya", false)
	privateFace := newNamedSearchFace(t, private)
	publicFace := newNamedSearchFace(t, entity.SubjectFixtures.Pointer("john-doe"))

	frm := func() form.SearchFaces {
		return form.SearchFaces{Markers: true, Unknown: "no", Count: 1000}
	}

	t.Run("Unscoped", func(t *testing.T) {
		results, err := Faces(frm())
		require.NoError(t, err)
		assert.True(t, holdsFace(results, privateFace.ID), "internal and CLI use is not scoped")
	})
	t.Run("NoSession", func(t *testing.T) {
		results, err := UserFaces(frm(), nil)
		require.NoError(t, err)
		assert.True(t, holdsFace(results, privateFace.ID))
	})
	t.Run("Admin", func(t *testing.T) {
		results, err := UserFaces(frm(), entity.SessionFixtures.Pointer("alice"))
		require.NoError(t, err)
		assert.True(t, holdsFace(results, privateFace.ID))
	})
	t.Run("WithoutPrivatePeople", func(t *testing.T) {
		results, err := UserFaces(frm(), filteredSession())
		require.NoError(t, err)
		assert.True(t, holdsFace(results, publicFace.ID), "the public cluster is kept")
		assert.False(t, holdsFace(results, privateFace.ID))
	})
	t.Run("DeniedRole", func(t *testing.T) {
		results, err := UserFaces(frm(), entity.SessionFixtures.Pointer("visitor"))
		require.NoError(t, err)
		assert.True(t, holdsFace(results, publicFace.ID))
		assert.False(t, holdsFace(results, privateFace.ID))
	})
	t.Run("HiddenPerson", func(t *testing.T) {
		hidden := newWithheldSearchSubject(t, "Hidden Hugo", true)
		hiddenFace := newNamedSearchFace(t, hidden)

		results, err := UserFaces(frm(), filteredSession())
		require.NoError(t, err)
		assert.True(t, holdsFace(results, publicFace.ID))
		assert.False(t, holdsFace(results, hiddenFace.ID))
	})
	// markers is off by default, so the clusters are then filtered on their own subject rather
	// than through the representative-marker join.
	t.Run("WithoutMarkers", func(t *testing.T) {
		results, err := UserFaces(form.SearchFaces{Unknown: "no", Count: 1000}, filteredSession())
		require.NoError(t, err)
		assert.True(t, holdsFace(results, publicFace.ID))
		assert.False(t, holdsFace(results, privateFace.ID))

		results, err = UserFaces(form.SearchFaces{UID: privateFace.ID}, filteredSession())
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

// TestUserFaces_OmitsWithheldPeopleByID is the named regression for GET /api/v1/faces/{id}, which
// answers through the uid branch and returns the cluster together with its marker.
func TestUserFaces_OmitsWithheldPeopleByID(t *testing.T) {
	private := newWithheldSearchSubject(t, "Private Quentin", false)
	privateFace := newNamedSearchFace(t, private)
	publicFace := newNamedSearchFace(t, entity.SubjectFixtures.Pointer("john-doe"))

	t.Run("Admin", func(t *testing.T) {
		results, err := UserFaces(form.SearchFaces{UID: privateFace.ID, Markers: true}, entity.SessionFixtures.Pointer("alice"))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, private.SubjUID, results[0].SubjUID)
	})
	t.Run("WithoutPrivatePeople", func(t *testing.T) {
		results, err := UserFaces(form.SearchFaces{UID: privateFace.ID, Markers: true}, filteredSession())
		require.NoError(t, err)
		assert.Empty(t, results, "the handler answers 404 on an empty result")

		results, err = UserFaces(form.SearchFaces{UID: publicFace.ID, Markers: true}, filteredSession())
		require.NoError(t, err)
		require.Len(t, results, 1, "a public cluster still resolves by id")
	})
	// Without markers the cluster is filtered on its own subject rather than through the
	// representative-marker join, so this is what pins the faces-table predicate on this branch.
	t.Run("WithoutMarkers", func(t *testing.T) {
		results, err := UserFaces(form.SearchFaces{UID: privateFace.ID}, filteredSession())
		require.NoError(t, err)
		assert.Empty(t, results)

		results, err = UserFaces(form.SearchFaces{UID: publicFace.ID}, filteredSession())
		require.NoError(t, err)
		require.Len(t, results, 1)
	})
}

// TestUserFaces_RepresentativeMarker pins that the representative is checked on its own subject,
// since a marker may name a person its cluster is not named after.
func TestUserFaces_RepresentativeMarker(t *testing.T) {
	private := newWithheldSearchSubject(t, "Private Rosa", false)
	public := entity.SubjectFixtures.Pointer("john-doe")

	f := newSearchFace(t)
	require.NoError(t, f.Update("SubjUID", public.SubjUID))

	// The private person's marker is the larger one, so it wins the ranking unless excluded.
	newSearchMarker(t, f.ID, entity.Marker{Size: 400, Score: 95, FaceDist: 0.2,
		SubjUID: private.SubjUID, SubjSrc: entity.SrcManual, MarkerName: private.SubjName})
	fallback := newSearchMarker(t, f.ID, entity.Marker{Size: 200, Score: 90, FaceDist: 0.2,
		SubjUID: public.SubjUID, SubjSrc: entity.SrcManual, MarkerName: public.SubjName})

	t.Run("Admin", func(t *testing.T) {
		results, err := UserFaces(form.SearchFaces{UID: f.ID, Markers: true}, entity.SessionFixtures.Pointer("alice"))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, private.SubjName, results[0].MarkerName)
	})
	t.Run("WithoutPrivatePeople", func(t *testing.T) {
		results, err := UserFaces(form.SearchFaces{UID: f.ID, Markers: true}, filteredSession())
		require.NoError(t, err)
		require.Len(t, results, 1, "the cluster itself is not private, so it stays")
		assert.Equal(t, fallback.MarkerUID, results[0].MarkerUID)
		assert.Equal(t, public.SubjName, results[0].MarkerName)
	})
}

// TestUserFaces_KeepsUnknownFaces pins People > New: it filters on the subj_uid column, so the
// unnamed clusters it lists are outside what this filter touches.
func TestUserFaces_KeepsUnknownFaces(t *testing.T) {
	f := newSearchFace(t)
	newSearchMarker(t, f.ID, entity.Marker{Size: 300, Score: 90, FaceDist: 0.2})

	results, err := UserFaces(form.SearchFaces{Unknown: "yes", Markers: true, Count: 1000}, filteredSession())
	require.NoError(t, err)
	assert.True(t, holdsFace(results, f.ID))
}
