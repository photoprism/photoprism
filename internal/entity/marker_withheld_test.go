package entity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

// withheldTestFileUID is a file uid no fixture uses, so the rows these tests add cannot change
// what another test in this package reads.
const withheldTestFileUID = "fs6sg6bw45bnpriv"

// createWithheldSubject adds a person whose name is withheld under the given name, since no
// fixture carries either flag.
func createWithheldSubject(t *testing.T, name string, hidden bool) *Subject {
	subj := NewSubject(name, SubjPerson, SrcManual)
	require.NotNil(t, subj)

	if hidden {
		subj.SubjHidden = true
	} else {
		subj.SubjPrivate = true
	}

	require.NoError(t, subj.Create())

	t.Cleanup(func() {
		UnscopedDb().Unscoped().Delete(&Subject{}, "subj_uid = ?", subj.SubjUID)
		SubjNames.Unset(subj.SubjUID)
	})

	return subj
}

// withheldMarkerFile stores one marker for a person whose name is withheld and one for a person
// whose name is not, and returns a file that has not loaded either yet. The second marker is the
// positive control: it must survive every case that withholds the first.
func withheldMarkerFile(t *testing.T, hidden bool) (f *File, withheld, public *Subject) {
	name := "Private Paula"

	if hidden {
		name = "Hidden Hannah"
	}

	withheld = createWithheldSubject(t, name, hidden)
	public = SubjectFixtures.Pointer("john-doe")

	markers := Markers{
		{FileUID: withheldTestFileUID, MarkerType: MarkerFace, MarkerSrc: SrcImage, Thumb: "publicthumb",
			SubjUID: public.SubjUID, SubjSrc: SrcManual, MarkerName: public.SubjName,
			X: 0.1, Y: 0.1, W: 0.1, H: 0.1, Size: 160, Score: 80},
		{FileUID: withheldTestFileUID, MarkerType: MarkerFace, MarkerSrc: SrcImage, Thumb: "privatethumb",
			SubjUID: withheld.SubjUID, SubjSrc: SrcManual, MarkerName: withheld.SubjName,
			X: 0.5, Y: 0.5, W: 0.1, H: 0.1, Size: 160, Score: 80},
	}

	for i := range markers {
		created, err := CreateMarkerIfNotExists(&markers[i])
		require.NoError(t, err)
		require.NotEmpty(t, created.MarkerUID)

		uid := created.MarkerUID

		t.Cleanup(func() {
			UnscopedDb().Unscoped().Delete(&Marker{}, "marker_uid = ?", uid)
		})
	}

	return &File{FileUID: withheldTestFileUID, FileName: "2020/01/private.jpg", InstanceID: "xmp-instance-id"}, withheld, public
}

// markerIdentities returns a name/uid pair per marker the serialized file carries.
func markerIdentities(t *testing.T, f *File) []string {
	b, err := json.Marshal(f)
	require.NoError(t, err)

	var res struct {
		Markers []struct {
			Name    string
			SubjUID string
		}
	}

	require.NoError(t, json.Unmarshal(b, &res))

	out := make([]string, 0, len(res.Markers))

	for _, m := range res.Markers {
		out = append(out, m.Name+"/"+m.SubjUID)
	}

	return out
}

func TestFile_MarkersForJSON(t *testing.T) {
	t.Run("Unrestricted", func(t *testing.T) {
		f, withheld, public := withheldMarkerFile(t, false)
		markers := *f.MarkersForJSON()
		require.Len(t, markers, 2)
		assert.Equal(t, public.SubjUID, markers[0].SubjUID)
		assert.Equal(t, withheld.SubjUID, markers[1].SubjUID)
	})
	t.Run("PrivateOmitted", func(t *testing.T) {
		f, withheld, public := withheldMarkerFile(t, false)
		f.OmitWithheldPeople = true

		markers := *f.MarkersForJSON()
		require.Len(t, markers, 1, "the public marker is kept, so an empty list is a failure")
		assert.Equal(t, public.SubjUID, markers[0].SubjUID)
		assert.Equal(t, public.SubjName, markers[0].MarkerName)
		assert.NotEqual(t, withheld.SubjUID, markers[0].SubjUID)
	})
	t.Run("HiddenOmitted", func(t *testing.T) {
		f, withheld, public := withheldMarkerFile(t, true)
		f.OmitWithheldPeople = true

		markers := *f.MarkersForJSON()
		require.Len(t, markers, 1, "the public marker is kept, so an empty list is a failure")
		assert.Equal(t, public.SubjUID, markers[0].SubjUID)
		assert.NotEqual(t, withheld.SubjUID, markers[0].SubjUID)
	})
	t.Run("NoFileUID", func(t *testing.T) {
		f := &File{OmitWithheldPeople: true}
		assert.Empty(t, *f.MarkersForJSON())
	})
	t.Run("OmitMarkers", func(t *testing.T) {
		f, _, _ := withheldMarkerFile(t, false)
		f.OmitMarkers = true
		f.OmitWithheldPeople = true
		assert.Empty(t, *f.MarkersForJSON())
	})
}

// TestFile_MarshalJSONOmitsWithheldPeople is the named regression for GET /api/v1/photos/{uid} and
// GET /api/v1/files/{hash}, which both serialize their markers through this path.
func TestFile_MarshalJSONOmitsWithheldPeople(t *testing.T) {
	t.Run("FullAccess", func(t *testing.T) {
		f, withheld, _ := withheldMarkerFile(t, false)
		f.RedactForSession(adminSession(), acl.ResourcePhotos)
		assert.Contains(t, markerIdentities(t, f), withheld.SubjName+"/"+withheld.SubjUID)
	})
	t.Run("NilSession", func(t *testing.T) {
		f, withheld, _ := withheldMarkerFile(t, false)
		f.RedactForSession(nil, acl.ResourcePhotos)
		assert.Contains(t, markerIdentities(t, f), withheld.SubjName+"/"+withheld.SubjUID)
	})
	t.Run("LibraryAccessWithoutWithheldPeople", func(t *testing.T) {
		f, withheld, public := withheldMarkerFile(t, false)
		f.RedactForSession(libraryNoWithheldPeopleSession(), acl.ResourcePhotos)

		names := markerIdentities(t, f)

		require.Len(t, names, 1, "the public marker is kept, so an empty list is a failure")
		assert.Equal(t, public.SubjName+"/"+public.SubjUID, names[0])
		assert.NotContains(t, names, withheld.SubjName+"/"+withheld.SubjUID)
	})
	t.Run("SharedOnly", func(t *testing.T) {
		f, _, _ := withheldMarkerFile(t, false)
		f.RedactForSession(sharedOnlySession(), acl.ResourcePhotos)
		assert.Empty(t, markerIdentities(t, f), "a shared-only session loses every marker")
	})
}

// TestFile_MarkersForJSONDoesNotPersist proves the response shaping stays in the response: a file
// serialized for a session that may not see a withheld person still holds and saves that marker.
func TestFile_MarkersForJSONDoesNotPersist(t *testing.T) {
	f, withheld, _ := withheldMarkerFile(t, false)
	f.RedactForSession(libraryNoWithheldPeopleSession(), acl.ResourcePhotos)

	require.Len(t, *f.MarkersForJSON(), 1, "the withheld marker is left out of the response")
	require.Len(t, *f.Markers(), 2, "the loaded markers stay complete")

	_, err := json.Marshal(f)
	require.NoError(t, err)

	if _, err = f.SaveMarkers(); err != nil {
		t.Fatal(err)
	}

	stored, err := FindMarkers(withheldTestFileUID)
	require.NoError(t, err)
	require.Len(t, stored, 2)

	found := false

	for _, m := range stored {
		if m.SubjUID == withheld.SubjUID {
			found = true
			assert.Equal(t, withheld.SubjName, m.MarkerName)
		}
	}

	assert.True(t, found, "the withheld marker is still stored with its name and subject")
}

// nameOnlyMarker stores a marker that carries a name without a subject link, which is the shape
// query.CreateMarkerSubjects exists to reconcile and which two fixtures are already in.
func nameOnlyMarker(t *testing.T, name string) {
	marker := Marker{FileUID: withheldTestFileUID, MarkerType: MarkerFace, MarkerSrc: SrcImage,
		Thumb: "nameonlythumb", MarkerName: name, SubjSrc: SrcManual, SubjUID: "",
		X: 0.8, Y: 0.8, W: 0.1, H: 0.1, Size: 200, Score: 80}

	created, err := CreateMarkerIfNotExists(&marker)
	require.NoError(t, err)
	require.Empty(t, created.SubjUID, "the stored marker has no subject link")

	t.Cleanup(func() {
		UnscopedDb().Unscoped().Delete(&Marker{}, "marker_uid = ?", created.MarkerUID)
	})
}

// TestFile_MarkersForJSONOmitsUnlinkedNames pins that identity is classified on the name as well as
// the link. A marker names a person in its own column, and that column is what a response
// discloses, so a marker with a name and no link is withheld on the name.
func TestFile_MarkersForJSONOmitsUnlinkedNames(t *testing.T) {
	withheld := createWithheldSubject(t, "Unlinked Ulla", false)
	nameOnlyMarker(t, withheld.SubjName)

	t.Run("Unrestricted", func(t *testing.T) {
		f := &File{FileUID: withheldTestFileUID}
		require.Len(t, *f.MarkersForJSON(), 1)
	})
	t.Run("Withheld", func(t *testing.T) {
		f := &File{FileUID: withheldTestFileUID, OmitWithheldPeople: true}
		assert.Empty(t, *f.MarkersForJSON())
	})
	t.Run("Serialized", func(t *testing.T) {
		f := &File{FileUID: withheldTestFileUID}
		f.RedactForSession(libraryNoWithheldPeopleSession(), acl.ResourcePhotos)

		for _, id := range markerIdentities(t, f) {
			assert.NotContains(t, id, withheld.SubjName)
			assert.NotContains(t, id, withheld.SubjUID, "the subject is not resolved on the way out either")
		}
	})
	t.Run("GeneratedNames", func(t *testing.T) {
		markers, err := FindMarkers(withheldTestFileUID)
		require.NoError(t, err)
		require.Len(t, markers, 1)
		assert.Empty(t, markers.SubjectNames())
	})
	t.Run("VisibleNameKept", func(t *testing.T) {
		markers := Markers{{MarkerType: MarkerFace, MarkerName: "Jens Mander", SubjSrc: SrcManual}}
		assert.Equal(t, []string{"Jens Mander"}, markers.SubjectNames(), "an unlinked visible name is kept")
	})
}

// TestFindMarkers pins that the shared loader stays unscoped, since indexing and face clustering
// read through it and a filtered list would starve clustering of rows.
func TestFindMarkers(t *testing.T) {
	_, withheld, _ := withheldMarkerFile(t, false)

	markers, err := FindMarkers(withheldTestFileUID)
	require.NoError(t, err)
	require.Len(t, markers, 2)

	found := false

	for _, m := range markers {
		if m.SubjUID == withheld.SubjUID {
			found = true
		}
	}

	assert.True(t, found, "the withheld person's marker is still loaded for indexing")
}

func TestFindVisibleMarkers(t *testing.T) {
	t.Run("IncludeWithheld", func(t *testing.T) {
		_, withheld, _ := withheldMarkerFile(t, false)

		markers, err := FindVisibleMarkers(withheldTestFileUID, false)
		require.NoError(t, err)
		require.Len(t, markers, 2)
		assert.Equal(t, withheld.SubjUID, markers[1].SubjUID)
	})
	t.Run("OmitWithheld", func(t *testing.T) {
		_, _, public := withheldMarkerFile(t, false)

		markers, err := FindVisibleMarkers(withheldTestFileUID, true)
		require.NoError(t, err)
		require.Len(t, markers, 1)
		assert.Equal(t, public.SubjUID, markers[0].SubjUID)
	})
	t.Run("KeepsMarkersWithoutSubject", func(t *testing.T) {
		marker := Marker{FileUID: withheldTestFileUID, MarkerType: MarkerFace, MarkerSrc: SrcImage,
			Thumb: "nosubjectthumb", SubjSrc: SrcAuto, X: 0.9, Y: 0.9, W: 0.1, H: 0.1}

		created, err := CreateMarkerIfNotExists(&marker)
		require.NoError(t, err)

		t.Cleanup(func() {
			UnscopedDb().Unscoped().Delete(&Marker{}, "marker_uid = ?", created.MarkerUID)
		})

		markers, err := FindVisibleMarkers(withheldTestFileUID, true)
		require.NoError(t, err)
		require.Len(t, markers, 1)
		assert.Empty(t, markers[0].SubjUID)
	})
}

// TestMarkers_SubjectNamesOmitsWithheldPeople is the named regression for generated titles,
// captions and search keywords, which are all derived from this list.
func TestMarkers_SubjectNamesOmitsWithheldPeople(t *testing.T) {
	public := SubjectFixtures.Pointer("john-doe")

	names := func(t *testing.T, withheld *Subject) []string {
		markers := Markers{
			{MarkerType: MarkerFace, SubjUID: public.SubjUID, SubjSrc: SrcManual, MarkerName: public.SubjName},
			{MarkerType: MarkerFace, SubjUID: withheld.SubjUID, SubjSrc: SrcManual, MarkerName: withheld.SubjName},
		}

		return markers.SubjectNames()
	}

	t.Run("Private", func(t *testing.T) {
		withheld := createWithheldSubject(t, "Private Nadia", false)
		assert.Equal(t, []string{public.SubjName}, names(t, withheld))
	})
	t.Run("Hidden", func(t *testing.T) {
		withheld := createWithheldSubject(t, "Hidden Nadia", true)
		assert.Equal(t, []string{public.SubjName}, names(t, withheld))
	})
	t.Run("Visible", func(t *testing.T) {
		other := SubjectFixtures.Pointer("joe-biden")
		markers := Markers{
			{MarkerType: MarkerFace, SubjUID: public.SubjUID, SubjSrc: SrcManual, MarkerName: public.SubjName},
			{MarkerType: MarkerFace, SubjUID: other.SubjUID, SubjSrc: SrcManual, MarkerName: other.SubjName},
		}

		assert.Equal(t, []string{public.SubjName, other.SubjName}, markers.SubjectNames())
	})
}

func TestSubject_NameWithheld(t *testing.T) {
	t.Run("Private", func(t *testing.T) {
		assert.True(t, (&Subject{SubjPrivate: true}).NameWithheld())
	})
	t.Run("Hidden", func(t *testing.T) {
		assert.True(t, (&Subject{SubjHidden: true}).NameWithheld())
	})
	t.Run("Visible", func(t *testing.T) {
		assert.False(t, SubjectFixtures.Pointer("john-doe").NameWithheld())
	})
	t.Run("Excluded", func(t *testing.T) {
		assert.False(t, (&Subject{SubjExcluded: true}).NameWithheld(), "excluded is not part of this check")
	})
}

func TestMarkers_SubjectUIDs(t *testing.T) {
	t.Run("FaceMarkersWithSubject", func(t *testing.T) {
		markers := Markers{
			{MarkerType: MarkerFace, SubjUID: "js6sg6b1qekk9jx8"},
			{MarkerType: MarkerFace, SubjUID: ""},
			{MarkerType: MarkerLabel, SubjUID: "ls6sg6b1wowuy3c3"},
		}

		assert.Equal(t, []string{"js6sg6b1qekk9jx8"}, markers.SubjectUIDs())
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, Markers{}.SubjectUIDs())
	})
}

func TestFindWithheldPeople(t *testing.T) {
	t.Run("ByUID", func(t *testing.T) {
		withheld := createWithheldSubject(t, "Private Otto", false)

		w, err := FindWithheldPeople()
		require.NoError(t, err)
		assert.True(t, w.Withholds(withheld.SubjUID, ""))
		assert.False(t, w.Withholds(SubjectFixtures.Get("john-doe").SubjUID, ""))
	})
	t.Run("ByName", func(t *testing.T) {
		withheld := createWithheldSubject(t, "Hidden Otto", true)

		w, err := FindWithheldPeople()
		require.NoError(t, err)
		assert.True(t, w.Withholds("", withheld.SubjName))
		assert.False(t, w.Withholds("", "John Doe"))
	})
	// Matched the same way on both drivers, which is why the comparison is in Go: subj_name is a
	// case-insensitive VARCHAR on MariaDB and a case-sensitive one on SQLite.
	t.Run("NameCaseIgnored", func(t *testing.T) {
		createWithheldSubject(t, "Private Casey", false)

		w, err := FindWithheldPeople()
		require.NoError(t, err)
		assert.True(t, w.Withholds("", "private casey"))
		assert.True(t, w.Withholds("", "PRIVATE CASEY"))
	})
	t.Run("EitherIsEnough", func(t *testing.T) {
		withheld := createWithheldSubject(t, "Private Nils", false)

		w, err := FindWithheldPeople()
		require.NoError(t, err)
		assert.True(t, w.Withholds(withheld.SubjUID, "John Doe"))
		assert.True(t, w.Withholds("js6sg6b1qekk9jx8", withheld.SubjName))
	})
	t.Run("Visible", func(t *testing.T) {
		w, err := FindWithheldPeople()
		require.NoError(t, err)
		assert.False(t, w.Withholds("js6sg6b1qekk9jx8", "John Doe"))
		assert.False(t, w.Withholds("", ""))
	})
}

// TestSubject_SaveFormRefreshesOnPrivateChange covers the pictures that were titled before the
// flag was set: clearing checked_at is what has the maintenance worker generate the title again.
func TestSubject_SaveFormRefreshesOnPrivateChange(t *testing.T) {
	subj := SubjectFixtures.Pointer("actress-1")

	// Photo of 19800101_000002_D640C559, which carries a marker for this person.
	photo := PhotoFixtures.Pointer("19800101_000002_D640C559")

	checked := Now()
	require.NoError(t, photo.Update("CheckedAt", &checked))

	frm, err := form.NewSubject(subj)
	require.NoError(t, err)

	frm.SubjPrivate = !subj.SubjPrivate

	changed, err := subj.SaveForm(frm)
	require.NoError(t, err)
	require.True(t, changed)

	t.Cleanup(func() {
		revert, revertErr := form.NewSubject(subj)
		require.NoError(t, revertErr)
		revert.SubjPrivate = false
		_, _ = subj.SaveForm(revert)
		// Restored too, or a later reader of this column depends on test order.
		_ = photo.Update("CheckedAt", &checked)
	})

	stored := FindPhoto(*photo)
	require.NotNil(t, stored)
	assert.Nil(t, stored.CheckedAt, "the picture is flagged for metadata maintenance")
}

func TestVisiblePeopleFilter(t *testing.T) {
	t.Run("WithNames", func(t *testing.T) {
		joins, cond := VisiblePeopleFilter("markers", true)
		require.Len(t, joins, 2)
		assert.Contains(t, joins[0], "LEFT JOIN subjects markers_subj ON markers_subj.subj_uid = markers.subj_uid")
		assert.Contains(t, joins[1], "LEFT JOIN subjects markers_named ON markers_named.subj_name = markers.marker_name")
		// Positive form, so a row with no person joined is kept rather than dropped on a NULL.
		assert.Contains(t, cond, "markers_subj.subj_uid IS NULL")
		assert.Contains(t, cond, "markers_subj.subj_private = 0 AND markers_subj.subj_hidden = 0")
		assert.Contains(t, cond, "markers_named.subj_uid IS NULL")
	})
	t.Run("WithoutNames", func(t *testing.T) {
		joins, cond := VisiblePeopleFilter("faces", false)
		require.Len(t, joins, 1, "the faces table has no name column")
		assert.Contains(t, joins[0], "faces_subj.subj_uid = faces.subj_uid")
		assert.NotContains(t, cond, "marker_name")
		assert.NotContains(t, cond, "faces_named")
	})
}

func TestSession_SeesPrivatePeople(t *testing.T) {
	t.Run("NilUnrestricted", func(t *testing.T) {
		var s *Session
		assert.True(t, s.SeesPrivatePeople())
	})
	t.Run("Admin", func(t *testing.T) {
		assert.True(t, adminSession().SeesPrivatePeople())
	})
	t.Run("Guest", func(t *testing.T) {
		assert.False(t, sharedOnlySession().SeesPrivatePeople())
	})
	// The scope is read alongside the role, since these names travel on other resources: an app
	// password admitted on photos or files would otherwise reach them without people in its scope.
	t.Run("ScopeWithoutPeople", func(t *testing.T) {
		assert.False(t, SessionFixtures.Pointer("alice_app_password_webdav").SeesPrivatePeople())
		assert.False(t, SessionFixtures.Pointer("visitor_token_metrics").SeesPrivatePeople())
	})
	t.Run("UnrestrictedScope", func(t *testing.T) {
		assert.True(t, SessionFixtures.Pointer("alice_app_password_full_access").SeesPrivatePeople(),
			"the option the app-password dialog offers for full access carries no restriction")
		assert.True(t, SessionFixtures.Pointer("alice").SeesPrivatePeople(), "a browser session has no scope")
	})
	t.Run("ScopeNamingPeople", func(t *testing.T) {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("alice"))
		s.SetScope("people photos")
		assert.True(t, s.SeesPrivatePeople())

		s.SetScope("photos")
		assert.False(t, s.SeesPrivatePeople(), "a resource the names travel on is not people")
	})
	t.Run("PortalClientForAccount", func(t *testing.T) {
		s := libraryNoWithheldPeopleSession()
		assert.False(t, s.SeesPrivatePeople(), "the portal role is absent from the people table")
		assert.False(t, s.HasSharedAccessOnly(acl.ResourcePhotos), "so no marker is dropped for sharing")
		assert.True(t, s.IsRegistered(), "and the shared-only redaction does not apply")
	})
}

// TestMarker_RedactForSession pins the shape a marker write answers with, which a read of the same
// marker has to match.
func TestMarker_RedactForSession(t *testing.T) {
	withheld := createWithheldSubject(t, "Redacted Rita", false)
	visible := SubjectFixtures.Pointer("john-doe")

	marker := func(subj *Subject) *Marker {
		return &Marker{MarkerUID: "mrs6sg6bwbjkbeez", MarkerType: MarkerFace,
			SubjUID: subj.SubjUID, SubjSrc: SrcManual, MarkerName: subj.SubjName}
	}

	t.Run("Withheld", func(t *testing.T) {
		m := marker(withheld).RedactForSession(libraryNoWithheldPeopleSession())
		assert.Empty(t, m.SubjUID)
		assert.Empty(t, m.MarkerName)
		assert.Empty(t, m.SubjSrc)
		assert.Equal(t, "mrs6sg6bwbjkbeez", m.MarkerUID, "only the identity is cleared")
	})
	t.Run("VisiblePerson", func(t *testing.T) {
		m := marker(visible).RedactForSession(libraryNoWithheldPeopleSession())
		assert.Equal(t, visible.SubjUID, m.SubjUID)
		assert.Equal(t, visible.SubjName, m.MarkerName)
	})
	t.Run("SessionSeesThem", func(t *testing.T) {
		m := marker(withheld).RedactForSession(adminSession())
		assert.Equal(t, withheld.SubjUID, m.SubjUID)
	})
	t.Run("NilSession", func(t *testing.T) {
		m := marker(withheld).RedactForSession(nil)
		assert.Equal(t, withheld.SubjUID, m.SubjUID)
	})
	t.Run("NilMarker", func(t *testing.T) {
		var m *Marker
		assert.Nil(t, m.RedactForSession(adminSession()))
	})
}

// adminSession returns a session holding full access to people.
func adminSession() *Session {
	s := &Session{}
	s.SetUser(UserFixtures.Pointer("alice"))

	return s
}

// libraryNoWithheldPeopleSession returns a session the filter applies to: full access to photos
// and no private access to people. CE reaches it with a portal credential acting for an account,
// since a credential on its own is not registered and would lose every marker instead.
func libraryNoWithheldPeopleSession() *Session {
	s := &Session{}
	s.SetClient(&Client{ClientRole: acl.RolePortal.String(), AuthProvider: authn.ProviderClient.String()})
	s.SetUser(UserFixtures.Pointer("alice"))

	return s
}

// sharedOnlySession returns a session limited to shared content.
func sharedOnlySession() *Session {
	s := &Session{}
	s.SetUser(UserFixtures.Pointer("guest"))

	return s
}

// TestPhoto_GenerateCaptionIgnoresTitleSource pins that an automatic caption is refreshed even when
// the title is not automatic. The caption has its own source to answer for, so a title somebody
// typed must not leave a generated caption naming a withheld person behind.
func TestPhoto_GenerateCaptionIgnoresTitleSource(t *testing.T) {
	withheld := createWithheldSubject(t, "Caption Carla", false)
	nameOnlyMarker(t, withheld.SubjName)

	markers, err := FindMarkers(withheldTestFileUID)
	require.NoError(t, err)
	require.Len(t, markers, 1)
	require.Empty(t, markers.SubjectNames(), "the withheld name is out of the list captions derive from")

	t.Run("ManualTitleAutoCaption", func(t *testing.T) {
		p := &Photo{PhotoUID: "ps6sg6be2lvl0zz1", TitleSrc: SrcManual, PhotoTitle: "My Own Title",
			CaptionSrc: SrcAuto, PhotoCaption: withheld.SubjName + ", A, B & C"}

		// Refused for the title, which is the point: the caption still has to be refreshed.
		require.Error(t, p.GenerateTitle(classify.Labels{}))
		assert.Equal(t, "My Own Title", p.PhotoTitle, "a title somebody typed is left alone")
		assert.NotContains(t, p.PhotoCaption, withheld.SubjName)
	})
	t.Run("ManualCaptionKept", func(t *testing.T) {
		p := &Photo{PhotoUID: "ps6sg6be2lvl0zz2", TitleSrc: SrcManual, PhotoTitle: "My Own Title",
			CaptionSrc: SrcManual, PhotoCaption: "a caption somebody wrote"}

		require.Error(t, p.GenerateTitle(classify.Labels{}))
		assert.Equal(t, "a caption somebody wrote", p.PhotoCaption, "a caption somebody wrote is left alone")
	})
}
