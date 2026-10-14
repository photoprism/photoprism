package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestCreateMarker(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPhoto(router)
		CreateMarker(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0y11")

		assert.Equal(t, http.StatusOK, r.Code)

		photoUid := gjson.Get(r.Body.String(), "UID").String()
		fileUid := gjson.Get(r.Body.String(), "Files.0.UID").String()
		markerUid := gjson.Get(r.Body.String(), "Files.0.Markers.0.UID").String()

		assert.NotEmpty(t, photoUid)
		assert.NotEmpty(t, fileUid)
		assert.NotEmpty(t, markerUid)

		u := "/api/v1/markers"

		frm := form.Marker{
			FileUID:       fileUid,
			MarkerType:    "face",
			X:             0.303519,
			Y:             0.260742,
			W:             0.548387,
			H:             0.365234,
			SubjSrc:       "",
			MarkerName:    "",
			MarkerReview:  false,
			MarkerInvalid: false,
		}

		if b, err := json.Marshal(frm); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("POST %s", u)
			r = PerformRequestWithBody(app, "POST", u, string(b))
		}

		assert.Equal(t, http.StatusCreated, r.Code)
		newUID := gjson.Get(r.Body.String(), "UID").String()
		assert.NotEmpty(t, newUID)
		assert.Equal(t, "/api/v1/markers/"+newUID, r.Header().Get("Location"))
	})
	t.Run("SuccessWithName", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPhoto(router)
		CreateMarker(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0y11")

		assert.Equal(t, http.StatusOK, r.Code)

		photoUid := gjson.Get(r.Body.String(), "UID").String()
		fileUid := gjson.Get(r.Body.String(), "Files.0.UID").String()
		markerUid := gjson.Get(r.Body.String(), "Files.0.Markers.0.UID").String()

		assert.NotEmpty(t, photoUid)
		assert.NotEmpty(t, fileUid)
		assert.NotEmpty(t, markerUid)

		u := "/api/v1/markers"

		frm := form.Marker{
			FileUID:       fileUid,
			MarkerType:    "face",
			X:             0.303519,
			Y:             0.260742,
			W:             0.548387,
			H:             0.365234,
			SubjSrc:       "manual",
			MarkerName:    "Jens Mander",
			MarkerReview:  false,
			MarkerInvalid: false,
		}

		if b, err := json.Marshal(frm); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("POST %s", u)
			r = PerformRequestWithBody(app, "POST", u, string(b))
		}

		assert.Equal(t, http.StatusCreated, r.Code)
		newUID := gjson.Get(r.Body.String(), "UID").String()
		assert.NotEmpty(t, newUID)
		assert.Equal(t, "/api/v1/markers/"+newUID, r.Header().Get("Location"))
		assert.Equal(t, "Jens Mander", gjson.Get(r.Body.String(), "Name").String())
		assert.Equal(t, "manual", gjson.Get(r.Body.String(), "SubjSrc").String())
	})
	t.Run("InvalidArea", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPhoto(router)
		CreateMarker(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0y11")

		assert.Equal(t, http.StatusOK, r.Code)

		photoUid := gjson.Get(r.Body.String(), "UID").String()
		fileUid := gjson.Get(r.Body.String(), "Files.0.UID").String()
		markerUid := gjson.Get(r.Body.String(), "Files.0.Markers.0.UID").String()

		assert.NotEmpty(t, photoUid)
		assert.NotEmpty(t, fileUid)
		assert.NotEmpty(t, markerUid)

		u := "/api/v1/markers"

		frm := form.Marker{
			FileUID:       fileUid,
			MarkerType:    "face",
			X:             0.5,
			Y:             0.5,
			W:             0,
			H:             0,
			SubjSrc:       "",
			MarkerName:    "",
			MarkerReview:  false,
			MarkerInvalid: false,
		}

		if b, err := json.Marshal(frm); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("POST %s", u)
			r = PerformRequestWithBody(app, "POST", u, string(b))
		}

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestUpdateMarker(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPhoto(router)
		UpdateMarker(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0y11")

		assert.Equal(t, http.StatusOK, r.Code)

		photoUid := gjson.Get(r.Body.String(), "UID").String()
		fileUid := gjson.Get(r.Body.String(), "Files.0.UID").String()
		markerUid := gjson.Get(r.Body.String(), "Files.0.Markers.0.UID").String()

		assert.NotEmpty(t, photoUid)
		assert.NotEmpty(t, fileUid)
		assert.NotEmpty(t, markerUid)

		u := fmt.Sprintf("/api/v1/markers/%s", markerUid)

		var m = form.Marker{
			SubjSrc:       "manual",
			MarkerInvalid: true,
			MarkerName:    "Foo",
		}

		if b, err := json.Marshal(m); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("PUT %s", u)
			r = PerformRequestWithBody(app, "PUT", u, string(b))
		}

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("NonPrimaryFile", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/ms6sg6b1wowu1000", "test")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("BadRequestFileAndPhotouidNotMatching", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/ms6sg6b1wowu1000", "test")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("FileNotExisting", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/1112", "test")

		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("MarkerNotExisting", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/1112", "test")

		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("EmptyPhotouid", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/ms6sg6b1wowu1000", "test")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("UpdateClusterWithExistingSubject", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		var m = form.Marker{
			SubjSrc:       "manual",
			MarkerInvalid: false,
			MarkerName:    "Actress A",
		}

		if b, err := json.Marshal(m); err != nil {
			t.Fatal(err)
		} else {
			r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/ms6sg6b1wowuy666", string(b))

			assert.Equal(t, http.StatusOK, r.Code)

			ClearMarkerSubject(router)

			r = PerformRequestWithBody(app, "DELETE", "/api/v1/markers/ms6sg6b1wowuy666/subject", "")

			assert.Equal(t, http.StatusOK, r.Code)
		}
	})
	t.Run("UpdateClusterWithExistingSubjectTwo", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		var m = form.Marker{
			SubjSrc:       "manual",
			MarkerInvalid: false,
			MarkerName:    "Actress A",
		}

		if b, err := json.Marshal(m); err != nil {
			t.Fatal(err)
		} else {
			r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/ms6sg6b1wowuy666", string(b))

			assert.Equal(t, http.StatusOK, r.Code)

			ClearMarkerSubject(router)

			r = PerformRequestWithBody(app, "DELETE", "/api/v1/markers/ms6sg6b1wowuy666/subject", "")

			assert.Equal(t, http.StatusOK, r.Code)
		}
	})
	t.Run("InvalidBody", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		var m = struct {
			ID      int
			Type    string
			Src     int
			Name    int
			SubjUID string
			SubjSrc string
			FaceID  string
		}{ID: 8,
			Type:    "face",
			Src:     123,
			Name:    456,
			SubjUID: "js6sg6b1h1njaaac",
			SubjSrc: "manual",
			FaceID:  "GMH5NISEEULNJL6RATITOA3TMZXMTMCI"}
		if b, err := json.Marshal(m); err != nil {
			t.Fatal(err)
		} else {
			r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/ms6sg6b1wowuy666", string(b))

			assert.Equal(t, http.StatusBadRequest, r.Code)
		}
	})
}

func TestClearMarkerSubject(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPhoto(router)
		ClearMarkerSubject(router)

		photoResp := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0y11")

		if photoResp == nil {
			t.Fatal("response is nil")
		}

		assert.Equal(t, http.StatusOK, photoResp.Code)

		if photoResp.Body.String() == "" {
			t.Fatal("body is empty")
		}

		photoUid := gjson.Get(photoResp.Body.String(), "UID").String()
		fileUid := gjson.Get(photoResp.Body.String(), "Files.0.UID").String()
		markerUid := gjson.Get(photoResp.Body.String(), "Files.0.Markers.0.UID").String()

		assert.NotEmpty(t, photoUid)
		assert.NotEmpty(t, fileUid)
		assert.NotEmpty(t, markerUid)

		u := fmt.Sprintf("/api/v1/markers/%s/subject", markerUid)

		// t.Logf("DELETE %s", u)

		resp := PerformRequestWithBody(app, "DELETE", u, "")

		assert.Equal(t, http.StatusOK, resp.Code)
	})
	t.Run("NonPrimaryFile", func(t *testing.T) {
		app, router, _ := NewApiTest()

		ClearMarkerSubject(router)

		r := PerformRequestWithBody(app, "DELETE", "/api/v1/markers/ms6sg6b1wowu1000/subject", "")

		assert.Equal(t, http.StatusOK, r.Code)
	})
}

// TestUpdateMarker_NamedCluster pins that naming one marker in a cluster named after another
// person relabels no other marker.
func TestUpdateMarker_NamedCluster(t *testing.T) {
	app, router, conf := NewApiTest()
	UpdateMarker(router)

	prevAuthMode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(prevAuthMode) })

	sess := entity.NewSession(conf.SessionMaxAge(), 0)
	sess.SetUser(entity.UserFixtures.Pointer("alice"))
	sess.SetScope("files")
	sess.SetProvider(authn.ProviderApplication)
	require.NoError(t, sess.Create())
	t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(sess) })
	require.False(t, sess.SeesPrivatePeople())

	person := entity.NewSubject("Named Cluster Person", entity.SubjPerson, entity.SrcManual)
	require.NotNil(t, person)
	require.NoError(t, person.Create())
	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Subject{}, "subj_uid = ?", person.SubjUID) })
	markPrivate(t, person, false)

	f := entity.NewFace(person.SubjUID, entity.SrcAuto, face.Embeddings{face.FixtureEmbedding(7201)}, face.EmbeddingModelName())
	require.NotNil(t, f)
	require.NoError(t, f.Create())
	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Face{}, "id = ?", f.ID) })

	newMarker := func(subjUID, subjSrc string) string {
		m := entity.Marker{
			MarkerUID:  rnd.GenerateUID('m'),
			FileUID:    entity.FileFixtures.Get("exampleDNGFile.dng").FileUID,
			MarkerType: entity.MarkerFace,
			SubjUID:    subjUID,
			SubjSrc:    subjSrc,
			FaceID:     f.ID,
			FaceDist:   0.1,
			EmbedModel: f.EmbedModel,
			MatchedAt:  entity.TimeStamp(),
			W:          0.1,
			H:          0.1,
		}

		require.NoError(t, entity.UnscopedDb().Create(&m).Error)
		t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Marker{}, "marker_uid = ?", m.MarkerUID) })

		return m.MarkerUID
	}

	auto := newMarker(person.SubjUID, entity.SrcAuto)
	unnamed := newMarker("", entity.SrcAuto)
	rejected := newMarker("", entity.SrcManual)

	b, err := json.Marshal(form.Marker{SubjSrc: entity.SrcManual, MarkerName: "Named Cluster Other"})
	require.NoError(t, err)

	r := AuthenticatedRequestWithBody(app, http.MethodPut, "/api/v1/markers/"+rejected, string(b), sess.AuthToken())
	require.Equal(t, http.StatusOK, r.Code, r.Body.String())

	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Subject{}, "subj_name = ?", "Named Cluster Other") })

	named := entity.FindMarker(rejected)
	require.NotNil(t, named)
	require.NotEmpty(t, named.SubjUID, "the named marker changes")
	assert.NotEqual(t, person.SubjUID, named.SubjUID)

	assert.Equal(t, person.SubjUID, entity.FindMarker(auto).SubjUID)
	assert.Empty(t, entity.FindMarker(unnamed).SubjUID)
	assert.Equal(t, person.SubjUID, entity.FindFace(f.ID).SubjUID)
}

// TestUpdateMarker_Correction pins that naming a face in a cluster after another existing person
// reports it to the cluster and answers with the face moved to a face of that person.
func TestUpdateMarker_Correction(t *testing.T) {
	app, router, _ := NewApiTest()
	UpdateMarker(router)

	carol := entity.NewSubject("Correction Api Carol", entity.SubjPerson, entity.SrcManual)
	require.NoError(t, carol.Create())
	dave := entity.NewSubject("Correction Api Dave", entity.SubjPerson, entity.SrcManual)
	require.NoError(t, dave.Create())
	t.Cleanup(func() {
		entity.UnscopedDb().Delete(&entity.Subject{}, "subj_uid IN (?)", []string{carol.SubjUID, dave.SubjUID})
	})

	f := entity.NewFace(carol.SubjUID, entity.SrcAuto, face.Embeddings{face.FixtureEmbedding(7601)}, face.EmbeddingModelName())
	require.NoError(t, f.Create())
	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Face{}, "id = ?", f.ID) })

	dist := 0.6 * f.AcceptDist()
	m := entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		FileUID:        entity.FileFixtures.Get("exampleDNGFile.dng").FileUID,
		MarkerType:     entity.MarkerFace,
		SubjUID:        carol.SubjUID,
		SubjSrc:        entity.SrcAuto,
		FaceID:         f.ID,
		FaceDist:       dist,
		EmbeddingsJSON: face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), dist, 1)}.JSON(),
		EmbedModel:     f.EmbedModel,
		Size:           face.ClusterSizeThreshold,
		Score:          face.ClusterScore("") + 10,
		MatchedAt:      entity.TimeStamp(),
		W:              0.1,
		H:              0.1,
	}

	require.NoError(t, entity.UnscopedDb().Create(&m).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Marker{}, "marker_uid = ?", m.MarkerUID) })
	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Face{}, "subj_uid = ?", dave.SubjUID) })

	b, err := json.Marshal(form.Marker{SubjSrc: entity.SrcManual, MarkerName: dave.SubjName})
	require.NoError(t, err)

	r := PerformRequestWithBody(app, http.MethodPut, "/api/v1/markers/"+m.MarkerUID, string(b))
	require.Equal(t, http.StatusOK, r.Code, r.Body.String())
	assert.Equal(t, dave.SubjUID, gjson.Get(r.Body.String(), "SubjUID").String())

	faceID := gjson.Get(r.Body.String(), "FaceID").String()
	require.NotEmpty(t, faceID)
	assert.NotEqual(t, f.ID, faceID)

	if own := entity.FindFace(faceID); assert.NotNil(t, own) {
		assert.Equal(t, dave.SubjUID, own.SubjUID)
	}

	cluster := entity.FindFace(f.ID)
	require.NotNil(t, cluster)
	assert.Equal(t, carol.SubjUID, cluster.SubjUID)
	assert.Equal(t, 1, cluster.Collisions)
}

// TestUpdateMarker_Review covers approve, reject, and compatible review payloads.
func TestUpdateMarker_Review(t *testing.T) {
	app, router, _ := NewApiTest()
	UpdateMarker(router)
	for _, tc := range []struct {
		name, body      string
		review, invalid bool
	}{
		{"Approve", `{"Review":false,"Invalid":false}`, false, false},
		{"ReviewOnly", `{"Review":false}`, false, true},
		{"Reject", `{"Review":false,"Invalid":true}`, false, true},
		{"Alias", `{"MarkerReview":false}`, false, true},
		{"Omitted", `{}`, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := entity.Marker{MarkerUID: rnd.GenerateUID('m'), FileUID: entity.FileFixtures.Get("exampleDNGFile.dng").FileUID,
				MarkerType: entity.MarkerFace, MarkerSrc: entity.SrcManual, MarkerReview: true, MarkerInvalid: true, W: 0.1, H: 0.1}
			require.NoError(t, entity.UnscopedDb().Create(&m).Error)
			t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Marker{}, "marker_uid = ?", m.MarkerUID) })
			r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/"+m.MarkerUID, tc.body)
			require.Equal(t, http.StatusOK, r.Code, r.Body.String())
			assert.Equal(t, tc.review, gjson.Get(r.Body.String(), "Review").Bool())
			assert.Equal(t, tc.invalid, gjson.Get(r.Body.String(), "Invalid").Bool())
			stored := entity.FindMarker(m.MarkerUID)
			require.NotNil(t, stored)
			assert.Equal(t, tc.review, stored.MarkerReview)
			assert.Equal(t, tc.invalid, stored.MarkerInvalid)
		})
	}
}

// TestCreateMarker_Review accepts both review names when creating an unnamed region.
func TestCreateMarker_Review(t *testing.T) {
	app, router, _ := NewApiTest()
	CreateMarker(router)
	for _, key := range []string{"Review", "MarkerReview"} {
		t.Run(key, func(t *testing.T) {
			body := fmt.Sprintf(`{"FileUID":%q,"Type":"face","Src":"manual","X":0.2,"Y":0.2,"W":0.1,"H":0.1,%q:true}`, entity.FileFixtures.Get("exampleDNGFile.dng").FileUID, key)
			r := PerformRequestWithBody(app, "POST", "/api/v1/markers", body)
			require.Equal(t, http.StatusCreated, r.Code, r.Body.String())
			uid := gjson.Get(r.Body.String(), "UID").String()
			t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Marker{}, "marker_uid = ?", uid) })
			assert.True(t, gjson.Get(r.Body.String(), "Review").Bool())
			stored := entity.FindMarker(uid)
			require.NotNil(t, stored)
			assert.True(t, stored.MarkerReview)
		})
	}
}

// TestCreateMarker_ManualFace validates the source and type of hand-drawn regions.
func TestCreateMarker_ManualFace(t *testing.T) {
	app, router, _ := NewApiTest()
	CreateMarker(router)
	for _, tc := range []struct {
		name, fields string
		status       int
	}{
		{"WebUI", `,"Type":"face","Src":"manual","Review":false,"Invalid":false`, http.StatusCreated},
		{"Defaults", ``, http.StatusCreated},
		{"NullDefaults", `,"Src":null,"Type":null`, http.StatusCreated},
		{"Xmp", `,"Src":"xmp"`, http.StatusBadRequest},
		{"Image", `,"Src":"image"`, http.StatusBadRequest},
		{"OtherType", `,"Type":"object"`, http.StatusBadRequest},
		{"EmptySource", `,"Src":""`, http.StatusBadRequest},
		{"EmptyType", `,"Type":""`, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := entity.FileFixtures.Get("exampleDNGFile.dng")
			before, err := entity.FindMarkers(file.FileUID)
			require.NoError(t, err)
			body := fmt.Sprintf(`{"FileUID":%q,"X":0.2,"Y":0.2,"W":0.1,"H":0.1%s}`, file.FileUID, tc.fields)
			r := PerformRequestWithBody(app, "POST", "/api/v1/markers", body)
			require.Equal(t, tc.status, r.Code, r.Body.String())
			after, err := entity.FindMarkers(file.FileUID)
			require.NoError(t, err)
			if tc.status != http.StatusCreated {
				assert.Len(t, after, len(before))
				return
			}
			uid := gjson.Get(r.Body.String(), "UID").String()
			t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Marker{}, "marker_uid = ?", uid) })
			stored := entity.FindMarker(uid)
			require.NotNil(t, stored)
			assert.Equal(t, entity.SrcManual, stored.MarkerSrc)
			assert.Equal(t, entity.MarkerFace, stored.MarkerType)
			assert.False(t, stored.DetectedFace())
			assert.Equal(t, "manual", gjson.Get(r.Body.String(), "Src").String())
			assert.Equal(t, "face", gjson.Get(r.Body.String(), "Type").String())
		})
	}
}

// TestCreateMarker_XmpReconcile preserves hand-drawn regions through an authoritative XMP update.
func TestCreateMarker_XmpReconcile(t *testing.T) {
	app, router, _ := NewApiTest()
	CreateMarker(router)
	file := entity.FileFixtures.Get("exampleDNGFile.dng")
	file.ID = 0
	file.FileUID = rnd.GenerateUID('f')
	file.FileName = file.FileUID + ".jpg"
	require.NoError(t, entity.UnscopedDb().Create(&file).Error)
	t.Cleanup(func() {
		entity.UnscopedDb().Delete(&entity.Marker{}, "file_uid = ?", file.FileUID)
		entity.UnscopedDb().Delete(&file)
	})
	body := fmt.Sprintf(`{"FileUID":%q,"Type":"face","Src":"manual","X":0.2,"Y":0.2,"W":0.1,"H":0.1}`, file.FileUID)
	r := PerformRequestWithBody(app, "POST", "/api/v1/markers", body)
	require.Equal(t, http.StatusCreated, r.Code, r.Body.String())
	uid := gjson.Get(r.Body.String(), "UID").String()
	imported := entity.NewMarker(file, crop.NewArea("face", 0.8, 0.8, 0.1, 0.1), "", entity.SrcXmp, entity.MarkerFace, 100, 100)
	require.NotNil(t, imported)
	require.NoError(t, imported.Create())
	imageName := filepath.Join(t.TempDir(), "photo.jpg")
	require.NoError(t, fs.Copy("../photoprism/testdata/xmp-faces/sidecar.jpg", imageName, false))
	const emptyRegions = `<x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description rdf:about="" xmlns:MP="http://ns.microsoft.com/photo/1.2/" xmlns:MPRI="http://ns.microsoft.com/photo/1.2/t/RegionInfo#"><MP:RegionInfo rdf:parseType="Resource"><MPRI:Regions><rdf:Bag/></MPRI:Regions></MP:RegionInfo></rdf:Description></rdf:RDF></x:xmpmeta>`
	require.NoError(t, os.WriteFile(imageName+fs.ExtXMP, []byte(emptyRegions), fs.ModeFile)) //nolint:gosec // Isolated test sidecar.
	media, err := photoprism.NewMediaFile(imageName)
	require.NoError(t, err)
	saved, _, err := photoprism.ApplyXmpFaces(media, &file)
	require.NoError(t, err)
	assert.True(t, saved)
	assert.Nil(t, entity.FindMarker(imported.MarkerUID), "the authoritative sidecar must reconcile imported regions")
	kept := entity.FindMarker(uid)
	require.NotNil(t, kept)
	assert.Equal(t, entity.SrcManual, kept.MarkerSrc)
	assert.Equal(t, entity.MarkerFace, kept.MarkerType)
}

// TestCreateMarker_FileHash requires a file hash before creating a region.
func TestCreateMarker_FileHash(t *testing.T) {
	app, router, _ := NewApiTest()
	router.Use(gin.Recovery())
	CreateMarker(router)
	file := entity.FileFixtures.Get("exampleDNGFile.dng")
	file.ID = 0
	file.FileUID = rnd.GenerateUID('f')
	file.FileHash = ""
	file.FileName = file.FileUID + ".jpg"
	require.NoError(t, entity.UnscopedDb().Create(&file).Error)
	t.Cleanup(func() {
		entity.UnscopedDb().Delete(&entity.Marker{}, "file_uid = ?", file.FileUID)
		entity.UnscopedDb().Delete(&file)
	})
	body := fmt.Sprintf(`{"FileUID":%q,"Type":"face","Src":"manual","X":0.2,"Y":0.2,"W":0.1,"H":0.1}`, file.FileUID)
	r := PerformRequestWithBody(app, "POST", "/api/v1/markers", body)
	assert.Equal(t, http.StatusNotFound, r.Code, r.Body.String())
	markers, err := entity.FindMarkers(file.FileUID)
	require.NoError(t, err)
	assert.Empty(t, markers)
}
