package photoprism

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// TestFaces_Match exercises the end-to-end matching flow with a loaded test configuration.
func TestFaces_Match(t *testing.T) {
	c := config.TestConfig()

	m := NewFaces(c)

	opt := FacesOptions{
		Force:     true,
		Threshold: 1,
	}

	r, err := m.Match(opt)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(r)
}

// TestBuildFaceCandidates validates that we drop non-matchable faces when building the index.
func TestBuildFaceCandidates(t *testing.T) {
	regular := entity.NewFace("", entity.SrcAuto, face.RandomEmbeddings(3, face.RegularFace))
	require.NotNil(t, regular)

	children := entity.NewFace("", entity.SrcAuto, face.RandomEmbeddings(3, face.ChildrenFace))
	require.NotNil(t, children)

	faces := entity.Faces{*regular, *children}

	index := buildFaceIndex(faces)

	require.Len(t, index.fallback, 1)
	require.Equal(t, regular.ID, index.fallback[0].ref.ID)
}

// TestSelectBestFace ensures the best candidate is returned after indexing.
func TestSelectBestFace(t *testing.T) {
	markerEmb := face.RandomEmbeddings(1, face.RegularFace)

	matchFace := entity.NewFace("", entity.SrcAuto, markerEmb)
	require.NotNil(t, matchFace)

	// Force a different face that should not be a better match.
	otherEmb := face.RandomEmbeddings(4, face.RegularFace)
	otherFace := entity.NewFace("", entity.SrcAuto, otherEmb)
	require.NotNil(t, otherFace)

	faces := entity.Faces{*matchFace, *otherFace}

	index := buildFaceIndex(faces)
	require.Len(t, index.fallback, 2)

	best, dist := selectBestFace(markerEmb, index)
	require.NotNil(t, best)
	require.Equal(t, matchFace.ID, best.ID)
	require.InDelta(t, 0.0, dist, 1e-9)
}

func TestFacesMatchRespectsVeto(t *testing.T) {
	conf := config.TestConfig()
	w := NewFaces(conf)

	var marker entity.Marker
	require.NoError(t, entity.Db().Where("marker_type = ? AND marker_invalid = 0 AND face_id <> ''", entity.MarkerFace).Take(&marker).Error)

	origFaceID := marker.FaceID
	require.NotEqual(t, "", origFaceID)

	var face entity.Face
	require.NoError(t, entity.Db().Where("id = ?", origFaceID).Take(&face).Error)

	_, err := marker.ClearFace()
	require.NoError(t, err)

	stats := make(map[*entity.Face]*faceMatchStats)
	faces := entity.Faces{face}

	w.rememberVeto(marker.MarkerUID)
	_, err = w.MatchFaces(faces, false, nil, stats)
	require.NoError(t, err)

	require.NoError(t, entity.Db().Where("marker_uid = ?", marker.MarkerUID).Take(&marker).Error)
	require.Equal(t, "", marker.FaceID)

	// restore original assignment to keep fixtures consistent
	dist := minMarkerDistance(face.Embedding(), marker.Embeddings())
	_, err = marker.SetFace(&face, dist)
	require.NoError(t, err)
	w.clearVeto(marker.MarkerUID)
}
