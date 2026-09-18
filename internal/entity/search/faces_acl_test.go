package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// peopleFilterApplies reports whether the role reaches the library without private access to
// people. The two conditions come from different resources, so nothing but this check links them.
func peopleFilterApplies(role acl.Role) bool {
	if !acl.Rules.AllowAny(acl.ResourcePhotos, role, acl.Permissions{acl.AccessAll, acl.AccessLibrary}) {
		return false
	}

	return !acl.Rules.Allow(acl.ResourcePeople, role, acl.AccessPrivate)
}

// roleSession returns a session holding the given role. A client role is paired with an admin
// account, since Session.Grants intersects both and the account is what makes it registered.
func roleSession(role acl.Role, client bool) *entity.Session {
	s := &entity.Session{}

	if client {
		s.SetClient(&entity.Client{ClientRole: role.String(), AuthProvider: authn.ProviderClient.String()})
		s.SetUser(entity.UserFixtures.Pointer("alice"))

		return s
	}

	s.SetUser(&entity.User{ID: 1, UserUID: rnd.GenerateUID(entity.UserUID),
		UserName: "acl-probe", UserRole: role.String()})

	return s
}

// TestPeopleVisibilityRoles pins a relationship the role tables hold across two resources: every
// role reaching the library while denied private access to people has the filter applied on both
// surfaces that resolve a name. It iterates the registered roles rather than naming them, and each
// edition reassigns these tables, so the same check in an edition package covers its own roles.
func TestPeopleVisibilityRoles(t *testing.T) {
	private := newWithheldSearchSubject(t, "Private Sasha", false)
	privateFace := newNamedSearchFace(t, private)
	publicFace := newNamedSearchFace(t, entity.SubjectFixtures.Pointer("john-doe"))

	fileUID := rnd.GenerateUID(entity.FileUID)

	newSearchMarker(t, privateFace.ID, entity.Marker{FileUID: fileUID, Size: 300, Score: 90, FaceDist: 0.2,
		SubjUID: private.SubjUID, SubjSrc: entity.SrcManual, MarkerName: private.SubjName})
	newSearchMarker(t, publicFace.ID, entity.Marker{FileUID: fileUID, Size: 300, Score: 90, FaceDist: 0.2,
		SubjUID: entity.SubjectFixtures.Get("john-doe").SubjUID, SubjSrc: entity.SrcManual, MarkerName: "John Doe"})

	roles := map[acl.Role]bool{}

	for _, role := range acl.UserRoles {
		if _, ok := roles[role]; !ok {
			roles[role] = false
		}
	}

	for _, role := range acl.ClientRoles {
		if isClient, ok := roles[role]; !ok || isClient {
			roles[role] = true
		}
	}

	covered := 0

	for role, client := range roles {
		if !peopleFilterApplies(role) {
			continue
		}

		covered++

		t.Run(role.String(), func(t *testing.T) {
			sess := roleSession(role, client)
			require.False(t, sess.SeesPrivatePeople())

			results, err := UserFaces(form.SearchFaces{Unknown: "no", Count: 1000}, sess)
			require.NoError(t, err)
			assert.True(t, holdsFace(results, publicFace.ID), "the public cluster is kept")
			assert.False(t, holdsFace(results, privateFace.ID))

			f := &entity.File{FileUID: fileUID}
			f.RedactForSession(sess, acl.ResourcePhotos)

			markers := *f.MarkersForJSON()

			for i := range markers {
				assert.NotEqual(t, private.SubjUID, markers[i].SubjUID)
			}

			if !sess.HasSharedAccessOnly(acl.ResourcePhotos) {
				assert.Len(t, markers, 1, "the public marker is kept")
			}
		})
	}

	assert.Positive(t, covered, "no role matches, so this check proves nothing")
}
