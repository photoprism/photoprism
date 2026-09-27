package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// srcTestMarker stores a face marker with the given subject source and name on a fixture file.
func srcTestMarker(t *testing.T, subjSrc, name string) *entity.Marker {
	t.Helper()

	m := entity.Marker{
		MarkerUID:  rnd.GenerateUID('m'),
		FileUID:    entity.FileFixtures.Get("exampleDNGFile.dng").FileUID,
		MarkerType: entity.MarkerFace,
		MarkerSrc:  entity.SrcImage,
		SubjSrc:    subjSrc,
		MarkerName: name,
		W:          0.1,
		H:          0.1,
	}

	require.NoError(t, entity.UnscopedDb().Create(&m).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(entity.Marker{}, "marker_uid = ?", m.MarkerUID) })

	return &m
}

// srcTestCleanup removes the people a test may have created by name.
func srcTestCleanup(t *testing.T, names ...string) {
	t.Cleanup(func() {
		for _, name := range names {
			entity.UnscopedDb().Delete(entity.Subject{}, "subj_name = ?", name)
		}
	})
}

// TestUpdateMarker_SubjectSrc pins which subject sources a request may name a face with.
func TestUpdateMarker_SubjectSrc(t *testing.T) {
	app, router, _ := NewApiTest()
	UpdateMarker(router)

	const name = "Marker Src Gate Person"
	srcTestCleanup(t, name)

	t.Run("WebUI", func(t *testing.T) {
		m := srcTestMarker(t, entity.SrcAuto, "")
		r := PerformRequestWithBody(app, http.MethodPut, "/api/v1/markers/"+m.MarkerUID, `{"SubjSrc":"manual","Name":"`+name+`"}`)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, entity.SrcManual, gjson.Get(r.Body.String(), "SubjSrc").String())
		assert.Equal(t, entity.SrcManual, entity.FindMarker(m.MarkerUID).SubjSrc)
		assert.Equal(t, name, entity.FindMarker(m.MarkerUID).MarkerName)
	})
	t.Run("Batch", func(t *testing.T) {
		m := srcTestMarker(t, entity.SrcAuto, "")
		r := PerformRequestWithBody(app, http.MethodPut, "/api/v1/markers/"+m.MarkerUID, `{"SubjSrc":"batch","Name":"`+name+`"}`)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, entity.SrcBatch, entity.FindMarker(m.MarkerUID).SubjSrc)
	})
	t.Run("Rejected", func(t *testing.T) {
		for _, src := range []string{entity.SrcAdmin, entity.SrcVision, entity.SrcXmp, entity.SrcMeta, entity.SrcImage, entity.SrcMarker, "zzz"} {
			m := srcTestMarker(t, entity.SrcAuto, "")
			r := PerformRequestWithBody(app, http.MethodPut, "/api/v1/markers/"+m.MarkerUID, `{"SubjSrc":"`+src+`","Name":"`+name+`"}`)
			assert.Equal(t, http.StatusBadRequest, r.Code, src)

			stored := entity.FindMarker(m.MarkerUID)
			require.NotNil(t, stored, src)
			assert.Equal(t, entity.SrcAuto, stored.SubjSrc, src)
			assert.Empty(t, stored.MarkerName, src)
			assert.Empty(t, stored.SubjUID, src)
		}
	})
	t.Run("RejectedShapes", func(t *testing.T) {
		for _, body := range []string{
			`{"SubjSrc":"Manual","Name":"` + name + `"}`,
			`{"SubjSrc":" manual","Name":"` + name + `"}`,
			`{"SubjSrc":"manual","SubjSrc":"admin","Name":"` + name + `"}`,
			`{"subjsrc":"admin","Name":"` + name + `"}`,
		} {
			m := srcTestMarker(t, entity.SrcAuto, "")
			r := PerformRequestWithBody(app, http.MethodPut, "/api/v1/markers/"+m.MarkerUID, body)
			assert.Equal(t, http.StatusBadRequest, r.Code, body)
			assert.Empty(t, entity.FindMarker(m.MarkerUID).MarkerName, body)
		}
	})
	t.Run("RejectedSourceChange", func(t *testing.T) {
		m := srcTestMarker(t, entity.SrcManual, name)
		r := PerformRequestWithBody(app, http.MethodPut, "/api/v1/markers/"+m.MarkerUID, `{"SubjSrc":"admin","Name":"`+name+`"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, entity.SrcManual, entity.FindMarker(m.MarkerUID).SubjSrc)
	})
	t.Run("RejectedNameChange", func(t *testing.T) {
		m := srcTestMarker(t, entity.SrcXmp, "Marker Src Xmp Person")
		r := PerformRequestWithBody(app, http.MethodPut, "/api/v1/markers/"+m.MarkerUID, `{"Name":"`+name+`"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code, "the stored source is not one a person may name with")
		assert.Equal(t, "Marker Src Xmp Person", entity.FindMarker(m.MarkerUID).MarkerName)
	})
	t.Run("AutoSrc", func(t *testing.T) {
		m := srcTestMarker(t, entity.SrcAuto, "")
		r := PerformRequestWithBody(app, http.MethodPut, "/api/v1/markers/"+m.MarkerUID, `{"SubjSrc":"","Name":"`+name+`"}`)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())

		stored := entity.FindMarker(m.MarkerUID)
		require.NotNil(t, stored)
		assert.Equal(t, entity.SrcAuto, stored.SubjSrc)
		assert.Empty(t, stored.MarkerName, "the name is not applied")
	})
	t.Run("AutoSrcOnXmp", func(t *testing.T) {
		m := srcTestMarker(t, entity.SrcXmp, "Marker Src Xmp Person")
		r := PerformRequestWithBody(app, http.MethodPut, "/api/v1/markers/"+m.MarkerUID, `{"SubjSrc":"","Name":"`+name+`"}`)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, "Marker Src Xmp Person", entity.FindMarker(m.MarkerUID).MarkerName, "the name is not applied")
		assert.Equal(t, entity.SrcXmp, entity.FindMarker(m.MarkerUID).SubjSrc)
	})
	t.Run("BlankNameOnXmpReject", func(t *testing.T) {
		m := srcTestMarker(t, entity.SrcXmp, "Marker Src Xmp Person")
		r := PerformRequestWithBody(app, http.MethodPut, "/api/v1/markers/"+m.MarkerUID, `{"Invalid":true,"Name":""}`)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.True(t, entity.FindMarker(m.MarkerUID).MarkerInvalid)
		assert.Equal(t, "Marker Src Xmp Person", entity.FindMarker(m.MarkerUID).MarkerName)
	})
	t.Run("SpacedNameOnXmp", func(t *testing.T) {
		m := srcTestMarker(t, entity.SrcXmp, "Marker Src Xmp Person")
		r := PerformRequestWithBody(app, http.MethodPut, "/api/v1/markers/"+m.MarkerUID, `{"Invalid":true,"Name":" Marker Src Xmp Person "}`)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.True(t, entity.FindMarker(m.MarkerUID).MarkerInvalid)
	})
	t.Run("InvalidAuto", func(t *testing.T) {
		m := srcTestMarker(t, entity.SrcAuto, "")
		r := PerformRequestWithBody(app, http.MethodPut, "/api/v1/markers/"+m.MarkerUID, `{"Invalid":true}`)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())

		stored := entity.FindMarker(m.MarkerUID)
		require.NotNil(t, stored)
		assert.True(t, stored.MarkerInvalid)
		assert.Equal(t, entity.SrcAuto, stored.SubjSrc)
	})
	t.Run("InvalidXmp", func(t *testing.T) {
		m := srcTestMarker(t, entity.SrcXmp, "Marker Src Xmp Person")
		r := PerformRequestWithBody(app, http.MethodPut, "/api/v1/markers/"+m.MarkerUID, `{"Invalid":true}`)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())

		stored := entity.FindMarker(m.MarkerUID)
		require.NotNil(t, stored)
		assert.True(t, stored.MarkerInvalid)
		assert.Equal(t, entity.SrcXmp, stored.SubjSrc, "source unchanged")
	})
}

// TestCreateMarker_SubjectSrc pins that a new marker is named only with a source a person may set.
func TestCreateMarker_SubjectSrc(t *testing.T) {
	app, router, _ := NewApiTest()
	CreateMarker(router)

	const name = "Marker Src Create Person"
	srcTestCleanup(t, name)

	fileUID := entity.FileFixtures.Get("exampleDNGFile.dng").FileUID
	body := func(subjSrc, name string) string {
		return `{"FileUID":"` + fileUID + `","Type":"face","Src":"manual","X":0.1,"Y":0.1,"W":0.2,"H":0.2,"SubjSrc":"` + subjSrc + `","Name":"` + name + `"}`
	}
	cleanup := func(r string) {
		if uid := gjson.Get(r, "UID").String(); uid != "" {
			t.Cleanup(func() { entity.UnscopedDb().Delete(entity.Marker{}, "marker_uid = ?", uid) })
		}
	}

	t.Run("WebUI", func(t *testing.T) {
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/markers", body("", ""))
		cleanup(r.Body.String())
		assert.Equal(t, http.StatusCreated, r.Code, r.Body.String())
	})
	t.Run("Manual", func(t *testing.T) {
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/markers", body(entity.SrcManual, name))
		cleanup(r.Body.String())
		assert.Equal(t, http.StatusCreated, r.Code, r.Body.String())
		assert.Equal(t, name, gjson.Get(r.Body.String(), "Name").String())
	})
	t.Run("BlankNameAdmin", func(t *testing.T) {
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/markers", body(entity.SrcAdmin, ""))
		cleanup(r.Body.String())
		require.Equal(t, http.StatusCreated, r.Code, r.Body.String())
		stored := entity.FindMarker(gjson.Get(r.Body.String(), "UID").String())
		require.NotNil(t, stored)
		assert.Equal(t, entity.SrcAuto, stored.SubjSrc, "a blank name stores no source")
	})
	t.Run("Admin", func(t *testing.T) {
		var before int
		require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).Where("file_uid = ?", fileUID).Count(&before).Error)

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/markers", body(entity.SrcAdmin, name))
		cleanup(r.Body.String())
		assert.Equal(t, http.StatusBadRequest, r.Code, r.Body.String())

		var after int
		require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).Where("file_uid = ?", fileUID).Count(&after).Error)
		assert.Equal(t, before, after, "nothing stored")
	})
}

// TestMarkerSubjectSrcAccepted pins the naming gate for each subject source.
func TestMarkerSubjectSrcAccepted(t *testing.T) {
	t.Run("Unchanged", func(t *testing.T) {
		assert.True(t, markerSubjectSrcAccepted("Jane", entity.SrcXmp, form.Marker{MarkerName: "Jane", SubjSrc: entity.SrcXmp}))
	})
	t.Run("Auto", func(t *testing.T) {
		assert.True(t, markerSubjectSrcAccepted("", entity.SrcAuto, form.Marker{MarkerName: "Jane", SubjSrc: entity.SrcAuto}))
	})
	t.Run("Accepted", func(t *testing.T) {
		for _, src := range []string{entity.SrcBatch, entity.SrcManual} {
			assert.True(t, markerSubjectSrcAccepted("", entity.SrcAuto, form.Marker{MarkerName: "Jane", SubjSrc: src}), src)
		}
	})
	t.Run("Rejected", func(t *testing.T) {
		for _, src := range []string{entity.SrcAdmin, entity.SrcVision, entity.SrcXmp, entity.SrcMeta, entity.SrcMarker, entity.SrcImage, entity.SrcDefault, "zzz"} {
			assert.False(t, markerSubjectSrcAccepted("", entity.SrcAuto, form.Marker{MarkerName: "Jane", SubjSrc: src}), src)
		}
	})
	t.Run("SourceChange", func(t *testing.T) {
		assert.False(t, markerSubjectSrcAccepted("Jane", entity.SrcManual, form.Marker{MarkerName: "Jane", SubjSrc: entity.SrcAdmin}))
	})
	t.Run("BlankName", func(t *testing.T) {
		assert.True(t, markerSubjectSrcAccepted("Jane", entity.SrcXmp, form.Marker{MarkerName: " ", SubjSrc: entity.SrcAdmin}))
	})
	t.Run("NameChange", func(t *testing.T) {
		assert.False(t, markerSubjectSrcAccepted("Jane", entity.SrcXmp, form.Marker{MarkerName: "Joan", SubjSrc: entity.SrcXmp}))
	})
}
