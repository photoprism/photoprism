package photoprism

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// xmpArea is the shared face rectangle used by the reconcile matrix tests.
var xmpArea = crop.Area{Name: "face", X: 0.308333, Y: 0.206944, W: 0.355556, H: 0.355556}

// xmpFarArea is a non-overlapping rectangle in the opposite corner.
var xmpFarArea = crop.Area{Name: "face", X: 0.90, Y: 0.90, W: 0.05, H: 0.05}

var xmpHashSeq int

// newXmpFile persists a fresh photo and its primary file with a unique hash.
func newXmpFile(t *testing.T) *entity.File {
	t.Helper()

	photo := entity.Photo{PhotoUID: rnd.GenerateUID('p'), PhotoName: "xmp-reconcile", PhotoType: entity.MediaImage}
	require.NoError(t, photo.Save())

	xmpHashSeq++
	hash := fmt.Sprintf("%040x", xmpHashSeq)

	file := &entity.File{
		PhotoID:     photo.ID,
		PhotoUID:    photo.PhotoUID,
		FileUID:     rnd.GenerateUID('f'),
		FileHash:    hash,
		FileName:    "xmp-reconcile/" + hash + ".jpg",
		FileRoot:    entity.RootOriginals,
		FilePrimary: true,
		FileType:    "jpg",
	}
	require.NoError(t, file.Save())

	return file
}

// addXmpMarker persists a marker on the file with the given source/name state.
func addXmpMarker(t *testing.T, file *entity.File, area crop.Area, markerSrc, subjSrc, subjUID, name string, invalid bool) *entity.Marker {
	t.Helper()
	m := entity.NewMarker(*file, area, subjUID, markerSrc, entity.MarkerFace, 100, 65)
	require.NotNil(t, m)
	m.SubjSrc = subjSrc
	m.MarkerName = name
	m.MarkerInvalid = invalid
	require.NoError(t, m.Create())
	return m
}

// newXmpSubject creates or returns a person for reconcile assertions.
func newXmpSubject(t *testing.T, name, src string) *entity.Subject {
	t.Helper()
	s := entity.FirstOrCreateSubject(entity.NewSubject(name, entity.SubjPerson, src))
	require.NotNil(t, s)
	return s
}

// runReconcile loads the file markers, reconciles the regions, and persists
// newly created markers the way the index flow does before re-querying.
func runReconcile(t *testing.T, file *entity.File, faces []meta.Face) entity.Markers {
	t.Helper()
	reconcileXmpFaces(faces, file, file.Markers())
	_, err := file.SaveMarkers()
	require.NoError(t, err)
	saved, err := entity.FindMarkers(file.FileUID)
	require.NoError(t, err)
	return saved
}

func region(name string, a crop.Area) meta.Face {
	return meta.Face{Name: name, X: a.X, Y: a.Y, W: a.W, H: a.H}
}

func TestReconcileXmpFaces(t *testing.T) {
	config.TestConfig()

	t.Run("Case1_NoMarker_CreatesXmp", func(t *testing.T) {
		file := newXmpFile(t)
		markers := runReconcile(t, file, []meta.Face{region("Alice", xmpArea)})
		require.Len(t, markers, 1)
		m := markers[0]
		assert.Equal(t, entity.SrcXmp, m.MarkerSrc)
		assert.Equal(t, entity.SrcXmp, m.SubjSrc)
		assert.Equal(t, "Alice", m.MarkerName)
		assert.NotEmpty(t, m.SubjUID)
		assert.InDelta(t, xmpArea.X, m.X, 0.001)
	})

	t.Run("Case2_AiAutoName_XmpWins", func(t *testing.T) {
		file := newXmpFile(t)
		auto := newXmpSubject(t, "AutoAlice", entity.SrcAuto)
		ai := addXmpMarker(t, file, xmpArea, entity.SrcImage, entity.SrcAuto, auto.SubjUID, "AutoAlice", false)

		markers := runReconcile(t, file, []meta.Face{region("Alice2", xmpArea)})
		require.Len(t, markers, 1)
		m := markers[0]
		assert.Equal(t, ai.MarkerUID, m.MarkerUID, "must reuse the AI marker")
		assert.Equal(t, entity.SrcImage, m.MarkerSrc, "box origin must stay SrcImage")
		assert.Equal(t, entity.SrcXmp, m.SubjSrc)
		assert.Equal(t, "Alice2", m.MarkerName)
		assert.InDelta(t, xmpArea.X, m.X, 0.001, "AI box unchanged")
	})

	t.Run("Case3_ManualName_XmpIgnored", func(t *testing.T) {
		file := newXmpFile(t)
		manual := newXmpSubject(t, "ManualAlice", entity.SrcManual)
		addXmpMarker(t, file, xmpArea, entity.SrcImage, entity.SrcManual, manual.SubjUID, "ManualAlice", false)

		markers := runReconcile(t, file, []meta.Face{region("Alicia", xmpArea)})
		require.Len(t, markers, 1)
		m := markers[0]
		assert.Equal(t, entity.SrcManual, m.SubjSrc, "manual name must win")
		assert.Equal(t, "ManualAlice", m.MarkerName)
	})

	t.Run("Case4_AiNoName_LinkExisting", func(t *testing.T) {
		file := newXmpFile(t)
		bob := newXmpSubject(t, "Bob4", entity.SrcMarker)
		ai := addXmpMarker(t, file, xmpArea, entity.SrcImage, "", "", "", false)

		// XMP casing differs from the Person; the Person must not be renamed.
		markers := runReconcile(t, file, []meta.Face{region("bob4", xmpArea)})
		require.Len(t, markers, 1)
		m := markers[0]
		assert.Equal(t, ai.MarkerUID, m.MarkerUID)
		assert.Equal(t, bob.SubjUID, m.SubjUID, "must link to existing Person")
		assert.Equal(t, entity.SrcImage, m.MarkerSrc)
		assert.Equal(t, entity.SrcXmp, m.SubjSrc)
		refreshed := entity.FindSubjectByName("Bob4", false)
		require.NotNil(t, refreshed)
		assert.Equal(t, "Bob4", refreshed.SubjName, "Person name must not be globally renamed to lowercase")
	})

	t.Run("Case5_AiNoName_NewPerson", func(t *testing.T) {
		file := newXmpFile(t)
		ai := addXmpMarker(t, file, xmpArea, entity.SrcImage, "", "", "", false)
		assert.Nil(t, entity.FindSubjectByName("Cara5", false))

		markers := runReconcile(t, file, []meta.Face{region("Cara5", xmpArea)})
		require.Len(t, markers, 1)
		m := markers[0]
		assert.Equal(t, ai.MarkerUID, m.MarkerUID)
		assert.Equal(t, entity.SrcImage, m.MarkerSrc)
		assert.Equal(t, entity.SrcXmp, m.SubjSrc)
		assert.NotEmpty(t, m.SubjUID)
		assert.NotNil(t, entity.FindSubjectByName("Cara5", false), "new Person must be created")
	})

	t.Run("Case6_RejectedMarker_StaysRejected", func(t *testing.T) {
		file := newXmpFile(t)
		rejected := addXmpMarker(t, file, xmpArea, entity.SrcImage, "", "", "", true)

		markers := runReconcile(t, file, []meta.Face{region("Ghost", xmpArea)})
		require.Len(t, markers, 1, "no new marker may be created over a rejected one")
		m := markers[0]
		assert.Equal(t, rejected.MarkerUID, m.MarkerUID)
		assert.True(t, m.MarkerInvalid, "must stay rejected")
		assert.Empty(t, m.MarkerName)
		assert.Empty(t, m.SubjUID)
	})

	t.Run("Case7_Idempotent", func(t *testing.T) {
		file := newXmpFile(t)
		first := runReconcile(t, file, []meta.Face{region("Dave", xmpArea)})
		require.Len(t, first, 1)
		second := runReconcile(t, file, []meta.Face{region("Dave", xmpArea)})
		require.Len(t, second, 1, "re-import must not duplicate")
		assert.Equal(t, first[0].MarkerUID, second[0].MarkerUID)
		assert.Equal(t, "Dave", second[0].MarkerName)
	})

	t.Run("Case8_DuplicateRegions_SingleMarker", func(t *testing.T) {
		file := newXmpFile(t)
		markers := runReconcile(t, file, []meta.Face{region("Eve", xmpArea), region("Eve", xmpArea)})
		require.Len(t, markers, 1, "duplicate regions must collapse to one marker")
	})

	t.Run("Unnamed_CreatesReviewMarker", func(t *testing.T) {
		file := newXmpFile(t)
		markers := runReconcile(t, file, []meta.Face{region("", xmpArea)})
		require.Len(t, markers, 1)
		m := markers[0]
		assert.Equal(t, entity.SrcXmp, m.MarkerSrc)
		assert.True(t, m.MarkerReview, "unnamed region must be flagged for review")
		assert.Empty(t, m.SubjUID, "no Person for an unnamed region")
	})

	t.Run("NamedNoOverlap_CreatesSecondMarker", func(t *testing.T) {
		file := newXmpFile(t)
		addXmpMarker(t, file, xmpArea, entity.SrcImage, "", "", "", false)
		markers := runReconcile(t, file, []meta.Face{region("Faraway", xmpFarArea)})
		require.Len(t, markers, 2, "a non-overlapping named region must add a marker")
	})

	t.Run("ValidAndRejectedOverlap_NamesValid", func(t *testing.T) {
		file := newXmpFile(t)
		addXmpMarker(t, file, xmpArea, entity.SrcImage, "", "", "", true) // rejected
		ai := addXmpMarker(t, file, xmpArea, entity.SrcImage, "", "", "", false)

		markers := runReconcile(t, file, []meta.Face{region("Hank", xmpArea)})
		var named *entity.Marker
		for i := range markers {
			if markers[i].MarkerUID == ai.MarkerUID {
				named = &markers[i]
			}
		}
		require.NotNil(t, named)
		assert.Equal(t, "Hank", named.MarkerName, "valid marker must be named even when a rejected marker also overlaps")
		assert.Equal(t, entity.SrcXmp, named.SubjSrc)
	})

	t.Run("NoClustering_XmpMarkerHasNoEmbedding", func(t *testing.T) {
		file := newXmpFile(t)
		markers := runReconcile(t, file, []meta.Face{region("Grace", xmpArea)})
		require.Len(t, markers, 1)
		assert.Empty(t, markers[0].EmbeddingsJSON, "XMP marker must carry no embedding")
		assert.Empty(t, markers[0].FaceID, "XMP marker must not seed a shared face")
	})

	t.Run("StaleLink_RepairedAndPersisted", func(t *testing.T) {
		file := newXmpFile(t)

		// A Person named "AliceRepair" exists independently of the marker.
		alice := newXmpSubject(t, "AliceRepair", entity.SrcMarker)

		// A prior import left the marker named but unlinked (empty SubjUID), so
		// the marker->Person link is broken in the database.
		stale := addXmpMarker(t, file, xmpArea, entity.SrcImage, entity.SrcXmp, "", "AliceRepair", false)
		require.Empty(t, stale.SubjUID, "precondition: marker starts unlinked")

		runReconcile(t, file, []meta.Face{region("AliceRepair", xmpArea)})

		// Re-query from the DB: the in-memory repair must be durably persisted.
		saved := entity.FindMarker(stale.MarkerUID)
		require.NotNil(t, saved)
		assert.Equal(t, "AliceRepair", saved.MarkerName)
		assert.Equal(t, alice.SubjUID, saved.SubjUID, "stale marker must be relinked to the existing Person and persisted")
		assert.Equal(t, entity.SrcXmp, saved.SubjSrc)
	})

	t.Run("StaleLink_MissingPerson_CreatesRatherThanDetaches", func(t *testing.T) {
		file := newXmpFile(t)

		// The marker is named but unlinked and no matching Person exists yet.
		stale := addXmpMarker(t, file, xmpArea, entity.SrcImage, entity.SrcXmp, "", "BobRepair", false)
		require.Empty(t, stale.SubjUID, "precondition: marker starts unlinked")
		require.Nil(t, entity.FindSubjectByName("BobRepair", false), "precondition: no Person yet")

		runReconcile(t, file, []meta.Face{region("BobRepair", xmpArea)})

		// The repair must create the Person and persist the link, not blank it.
		saved := entity.FindMarker(stale.MarkerUID)
		require.NotNil(t, saved)
		assert.NotEmpty(t, saved.SubjUID, "marker must be linked to a freshly created Person")
		assert.Equal(t, entity.SrcXmp, saved.SubjSrc)
		require.NotNil(t, entity.FindSubjectByName("BobRepair", false), "the XMP import must create the Person")
	})
}

// TestReconcileXmpFaces_OverwritesClusteredAiFace models the realistic flow of
// an AI-recognized, auto-clustered person spanning two photos whose name is
// later overwritten by a different XMP tag on one of them. The import must
// relabel only the tagged marker: the sibling marker on the other photo, the
// shared Face cluster, and the original person's global name must stay intact
// (an XMP name labels only its own marker — no XMP-driven clustering in v1).
func TestReconcileXmpFaces_OverwritesClusteredAiFace(t *testing.T) {
	config.TestConfig()

	// The AI recognized "AutoAlice" and clustered her across two photos.
	autoPerson := newXmpSubject(t, "AutoAlice", entity.SrcAuto)

	emb := make(face.Embedding, len(face.NullEmbedding))
	emb[0], emb[1] = 2, 1

	cluster := &entity.Face{
		ID:            "XMPCLUSTER0000000000000000000000000000000A",
		FaceSrc:       entity.SrcAuto,
		SubjUID:       autoPerson.SubjUID,
		EmbeddingJSON: (face.Embeddings{emb}).JSON(),
		Samples:       2,
	}
	require.NoError(t, entity.Db().Create(cluster).Error)

	// clustered persists a detected AI marker tied to the shared face cluster.
	clustered := func(file *entity.File) *entity.Marker {
		m := entity.NewMarker(*file, xmpArea, autoPerson.SubjUID, entity.SrcImage, entity.MarkerFace, 100, 100)
		require.NotNil(t, m)
		m.SubjSrc = entity.SrcAuto
		m.MarkerName = "AutoAlice"
		m.FaceID = cluster.ID
		m.FaceDist = 0.1
		m.SetEmbeddings(face.Embeddings{emb})
		require.NoError(t, m.Create())
		return m
	}

	file1 := newXmpFile(t)
	file2 := newXmpFile(t)
	markerA := clustered(file1)
	markerB := clustered(file2)

	// A different XMP tag lands on the first photo's face.
	runReconcile(t, file1, []meta.Face{region("XmpBob", xmpArea)})

	// Tagged marker: relabeled to the XMP person, box + detection origin kept.
	gotA := entity.FindMarker(markerA.MarkerUID)
	require.NotNil(t, gotA)
	assert.Equal(t, "XmpBob", gotA.MarkerName, "tagged marker must adopt the XMP name")
	assert.Equal(t, entity.SrcXmp, gotA.SubjSrc)
	assert.Equal(t, entity.SrcImage, gotA.MarkerSrc, "detection box origin must stay SrcImage")
	assert.InDelta(t, xmpArea.X, gotA.X, 0.001, "AI box must be unchanged")
	xmpBob := entity.FindSubjectByName("XmpBob", false)
	require.NotNil(t, xmpBob, "a fresh person must be created for the XMP name")
	assert.Equal(t, xmpBob.SubjUID, gotA.SubjUID)
	assert.NotEqual(t, autoPerson.SubjUID, gotA.SubjUID, "tagged marker must point to the new person")
	// The marker keeps its cluster link (FaceID) while pointing at a different
	// subject — a known v1 limitation: XMP renames do not re-cluster the face.
	assert.Equal(t, cluster.ID, gotA.FaceID, "cluster link is intentionally left intact in v1")

	// Sibling marker on the other photo must be completely untouched.
	gotB := entity.FindMarker(markerB.MarkerUID)
	require.NotNil(t, gotB)
	assert.Equal(t, "AutoAlice", gotB.MarkerName, "sibling marker name must not change")
	assert.Equal(t, autoPerson.SubjUID, gotB.SubjUID, "sibling marker must keep the auto person")
	assert.Equal(t, entity.SrcAuto, gotB.SubjSrc, "sibling marker source must stay SrcAuto")

	// The shared cluster and the original person's global name must be intact.
	gotFace := entity.FindFace(cluster.ID)
	require.NotNil(t, gotFace)
	assert.Equal(t, autoPerson.SubjUID, gotFace.SubjUID, "shared face cluster must not be reassigned")
	gotPerson := entity.FindSubject(autoPerson.SubjUID)
	require.NotNil(t, gotPerson)
	assert.Equal(t, "AutoAlice", gotPerson.SubjName, "auto person must not be globally renamed")
}

func TestCollectXmpFaces(t *testing.T) {
	c := config.TestConfig()
	if c.ExifToolBin() == "" {
		t.Skip("exiftool not configured")
	}

	t.Run("Embedded", func(t *testing.T) {
		m, err := NewMediaFile("testdata/xmp-faces/embedded-mwg.jpg")
		require.NoError(t, err)
		// Generate the ExifTool JSON cache the index flow relies on before reading.
		require.NoError(t, m.CreateExifToolJson(NewConvert(c)))
		faces := collectXmpFaces(m)
		require.Len(t, faces, 1)
		assert.Equal(t, "Alice", faces[0].Name)
		assert.InDelta(t, 0.45, faces[0].X, 0.01)
	})

	t.Run("Sidecar", func(t *testing.T) {
		m, err := NewMediaFile("testdata/xmp-faces/sidecar.jpg")
		require.NoError(t, err)
		faces := collectXmpFaces(m)
		require.GreaterOrEqual(t, len(faces), 1)
		found := false
		for _, f := range faces {
			if f.Name == "Cara" {
				found = true
			}
		}
		assert.True(t, found, "sidecar region 'Cara' must be collected, got %+v", faces)
	})

	t.Run("RotatedOrientation6", func(t *testing.T) {
		m, err := NewMediaFile("testdata/xmp-faces/rotated-o6.jpg")
		require.NoError(t, err)
		require.NoError(t, m.CreateExifToolJson(NewConvert(c)))
		faces := collectXmpFaces(m)
		require.Len(t, faces, 1)
		f := faces[0]
		// center (0.5,0.4) size (0.1,0.15) -> TL (0.45,0.325); rotateRect(...,6)
		// = (1-0.325-0.15, 0.45, 0.15, 0.1) = (0.525,0.45,0.15,0.1).
		assert.Equal(t, "Rita", f.Name)
		assert.InDelta(t, 0.525, f.X, 0.01)
		assert.InDelta(t, 0.45, f.Y, 0.01)
		assert.InDelta(t, 0.15, f.W, 0.01)
		assert.InDelta(t, 0.10, f.H, 0.01)
	})
}
