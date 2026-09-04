package search

import (
	"strings"
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFaces(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		results, err := Faces(form.SearchFaces{Unknown: "yes", Order: "added", Markers: true})
		require.NoError(t, err)
		t.Logf("Faces: %#v", results)
		if len(results) == 0 {
			t.Fatal("results are empty")
		} else if results[0].MarkerUID == "" {
			t.Fatal("marker uid is empty")
		}
	})
	t.Run("SearchWithLimit", func(t *testing.T) {
		results, err := Faces(form.SearchFaces{Offset: 3, Order: "subject", Markers: true})
		require.NoError(t, err)
		t.Logf("Faces: %#v", results)
		assert.LessOrEqual(t, 1, len(results))
	})
	t.Run("FindSpecificId", func(t *testing.T) {
		results, err := Faces(form.SearchFaces{UID: "PN6QO5INYTUSAATOFL43LL2ABAV5ACZK", Markers: true})
		require.NoError(t, err)
		t.Logf("Faces: %#v", results)
		assert.LessOrEqual(t, 1, len(results))
	})
	t.Run("ExcludeUnknownHidden", func(t *testing.T) {
		results, err := Faces(form.SearchFaces{Unknown: "no", Hidden: "yes", Order: "samples", Markers: true})
		require.NoError(t, err)
		t.Logf("Faces: %#v", results)
		assert.LessOrEqual(t, 0, len(results))
	})
	t.Run("NotNil", func(t *testing.T) {
		results, err := Faces(form.SearchFaces{Unknown: "no", Hidden: "yes", Order: "samples", Markers: true, Count: 100, Offset: 999999})
		require.NoError(t, err)
		t.Logf("Faces: %#v", results)
		assert.NotNil(t, results)
		assert.Len(t, results, 0)
	})
}

// newSearchFace saves an anonymous cluster so a test can control every marker it holds.
func newSearchFace(t *testing.T) *entity.Face {
	t.Helper()

	m := entity.NewFace("", entity.SrcAuto, face.Embeddings{face.FixtureEmbedding(uint64(len(t.Name())*7919) + uint64(time.Now().UnixNano()))}, face.EmbeddingModelName())

	require.NotNil(t, m)
	require.NoError(t, m.Create())
	t.Cleanup(func() { entity.Db().Delete(m) })

	return m
}

// newSearchMarker saves a face marker of the given cluster, defaulting the fields the search bars
// read so a case states only what it varies.
func newSearchMarker(t *testing.T, faceID string, m entity.Marker) *entity.Marker {
	t.Helper()

	m.FaceID = faceID
	m.MarkerType = entity.MarkerFace

	if m.MarkerSrc == "" {
		m.MarkerSrc = entity.SrcImage
	}

	if m.FileUID == "" {
		m.FileUID = "fs6sg6bw45bnlqdw"
	}

	if m.Thumb == "" {
		m.Thumb = rnd.GenerateUID('m')
	}

	// Marker uids order only to the second and randomly within it, so a case that has to tell a
	// ranking apart from the old "lowest uid wins" sets them rather than relying on insert order.
	if m.MarkerUID == "" {
		m.MarkerUID = rnd.GenerateUID('m')
	}

	require.NoError(t, entity.Db().Create(&m).Error)
	t.Cleanup(func() { entity.Db().Delete(&m) })

	return &m
}

// faceResult returns the search result for one cluster, or nil when the query hides it.
func faceResult(t *testing.T, faceID string) *Face {
	t.Helper()

	results, err := Faces(form.SearchFaces{UID: faceID, Markers: true})
	require.NoError(t, err)

	for i := range results {
		if results[i].ID == faceID {
			return &results[i]
		}
	}

	return nil
}

// TestFacesRepresentativeMarker covers which marker People shows for a cluster, and which clusters
// the page shows at all. Both were decided by literals that no configuration reached.
func TestFacesRepresentativeMarker(t *testing.T) {
	t.Run("PicksTheLargestFace", func(t *testing.T) {
		f := newSearchFace(t)
		// The worst candidate carries the lowest uid, so picking by uid returns the wrong marker.
		newSearchMarker(t, f.ID, entity.Marker{MarkerUID: "ms6sg6bw45bnl001", Size: 120, Score: 95, FaceDist: 0.2})
		want := newSearchMarker(t, f.ID, entity.Marker{MarkerUID: "ms6sg6bw45bnl002", Size: 300, Score: 80, FaceDist: 0.2})

		got := faceResult(t, f.ID)
		require.NotNil(t, got)
		assert.Equal(t, want.MarkerUID, got.MarkerUID)
	})
	t.Run("PicksTheMostConfidentOfEqualSize", func(t *testing.T) {
		f := newSearchFace(t)
		newSearchMarker(t, f.ID, entity.Marker{MarkerUID: "ms6sg6bw45bnl011", Size: 200, Score: 80, FaceDist: 0.2})
		want := newSearchMarker(t, f.ID, entity.Marker{MarkerUID: "ms6sg6bw45bnl012", Size: 200, Score: 95, FaceDist: 0.2})

		got := faceResult(t, f.ID)
		require.NotNil(t, got)
		assert.Equal(t, want.MarkerUID, got.MarkerUID)
	})
	t.Run("ShowsAClusterClusteringWouldAccept", func(t *testing.T) {
		// Between FACE_CLUSTER_SIZE and the 80 px this query used to demand, which hid twenty of
		// sixty-two clusters on one real library.
		f := newSearchFace(t)
		want := newSearchMarker(t, f.ID, entity.Marker{Size: face.ClusterSizeThreshold, Score: 80, FaceDist: 0.2})

		got := faceResult(t, f.ID)
		require.NotNil(t, got, "a cluster whose markers clustering accepts has to be visible")
		assert.Equal(t, want.MarkerUID, got.MarkerUID)
	})
	t.Run("ShowsAClusterWhoseMarkersAreAllFarFromTheCentroid", func(t *testing.T) {
		// Beyond the old 0.64 bound but inside what a cluster of this model accepts, which is
		// where most assigned markers of a wide cluster sit.
		f := newSearchFace(t)
		dist := face.AcceptDist(face.ClusterRadius) - 0.01
		require.Greater(t, dist, 0.64)
		want := newSearchMarker(t, f.ID, entity.Marker{Size: 200, Score: 80, FaceDist: dist})

		got := faceResult(t, f.ID)
		require.NotNil(t, got)
		assert.Equal(t, want.MarkerUID, got.MarkerUID)
	})
	t.Run("HidesAClusterBelowTheClusteringSize", func(t *testing.T) {
		f := newSearchFace(t)
		newSearchMarker(t, f.ID, entity.Marker{Size: face.ClusterSizeThreshold - 1, Score: 80, FaceDist: 0.2})

		assert.Nil(t, faceResult(t, f.ID), "the size bar still applies, at the clustering value")
	})
	t.Run("IgnoresInvalidAndUnmeasuredMarkers", func(t *testing.T) {
		// Size defaults to -1 for a marker whose file dimensions are unknown, so it fails the bar
		// rather than passing an unset value through.
		f := newSearchFace(t)
		newSearchMarker(t, f.ID, entity.Marker{Size: 400, Score: 95, FaceDist: 0.2, MarkerInvalid: true})
		newSearchMarker(t, f.ID, entity.Marker{Size: -1, Score: 95, FaceDist: 0.2})
		want := newSearchMarker(t, f.ID, entity.Marker{Size: 120, Score: 80, FaceDist: 0.2})

		got := faceResult(t, f.ID)
		require.NotNil(t, got)
		assert.Equal(t, want.MarkerUID, got.MarkerUID)
	})
}

// TestRepresentativeMarkerJoin covers the predicate the three unknown variants used to duplicate,
// which had already drifted between the copies.
func TestRepresentativeMarkerJoin(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		join, args := representativeMarkerJoin("faces", "yes")
		assert.Contains(t, join, "m2.subj_uid = ''")
		assert.NotContains(t, join, "m2.subj_uid <> ''")
		assert.Equal(t, strings.Count(join, "?"), len(args))
	})
	t.Run("Known", func(t *testing.T) {
		join, args := representativeMarkerJoin("faces", "no")
		assert.Contains(t, join, "m2.subj_uid <> ''")
		// Dropped with the predicate that required it: automatic assignment never writes a name,
		// so a person named once had no marker that could represent their cluster here.
		assert.NotContains(t, join, "marker_name")
		assert.Equal(t, strings.Count(join, "?"), len(args))
	})
	t.Run("Any", func(t *testing.T) {
		join, args := representativeMarkerJoin("faces", "")
		assert.NotContains(t, join, "m2.subj_uid")
		assert.Equal(t, strings.Count(join, "?"), len(args))
	})
	// Every literal this query used to carry is gone; the bars come from the face configuration.
	t.Run("CarriesNoThresholdLiterals", func(t *testing.T) {
		join, _ := representativeMarkerJoin("faces", "")
		for _, literal := range []string{"0.64", "80", "15", "MIN(", "GROUP BY"} {
			assert.NotContains(t, join, literal)
		}
	})
}

// TestFacesSampleOrder pins the ranking People > New reads: clusters largest first by the number of
// embeddings their centroid was built from, which is the size each had when it was formed.
func TestFacesSampleOrder(t *testing.T) {
	t.Run("DescendingBySamples", func(t *testing.T) {
		results, err := Faces(form.SearchFaces{Order: "samples"})
		require.NoError(t, err)
		require.Greater(t, len(results), 1, "ordering needs at least two clusters")

		ranked := false

		for i := 1; i < len(results); i++ {
			assert.GreaterOrEqual(t, results[i-1].Samples, results[i].Samples)

			if results[i-1].Samples != results[i].Samples {
				ranked = true
			}
		}

		// Equal counts throughout would satisfy the loop above without ordering anything.
		assert.True(t, ranked, "fixtures must differ in samples for the order to be tested")
	})
	t.Run("DefaultIsTheSameOrder", func(t *testing.T) {
		bySamples, err := Faces(form.SearchFaces{Order: "samples"})
		require.NoError(t, err)

		byDefault, err := Faces(form.SearchFaces{})
		require.NoError(t, err)
		require.Equal(t, len(bySamples), len(byDefault))

		for i := range bySamples {
			assert.Equal(t, bySamples[i].ID, byDefault[i].ID)
		}
	})
}
