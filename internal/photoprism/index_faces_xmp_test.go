package photoprism

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/fs"
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
	_, err := reconcileXmpFaces(meta.FaceRegions{Faces: faces, Declared: true}, file, file.Markers())
	require.NoError(t, err)
	_, err = file.SaveMarkers()
	require.NoError(t, err)
	saved, err := entity.FindMarkers(file.FileUID)
	require.NoError(t, err)
	return saved
}

// mustCollectXmpFaces collects an authoritative XMP face set for tests.
func mustCollectXmpFaces(t *testing.T, m *MediaFile) []meta.Face {
	t.Helper()

	regions, err := collectXmpFaces(m)
	require.NoError(t, err)

	return regions.Faces
}

func region(name string, a crop.Area) meta.Face {
	return meta.Face{Name: name, X: a.X, Y: a.Y, W: a.W, H: a.H}
}

// hasXmpMarkerName reports whether markers contain an imported XMP face name.
func hasXmpMarkerName(markers entity.Markers, name string) bool {
	for i := range markers {
		if markers[i].MarkerSrc == entity.SrcXmp && markers[i].MarkerName == name {
			return true
		}
	}

	return false
}

// newXmpIndexConfig returns an isolated config and installs it for XMP discovery.
func newXmpIndexConfig(t *testing.T, name string) *config.Config {
	t.Helper()
	t.Setenv("PHOTOPRISM_TEST_DSN", filepath.Join(t.TempDir(), name+".db"))

	c := config.NewMinimalTestConfigWithDb(name, filepath.Join(t.TempDir(), "storage"))
	oldConfig := Config()
	SetConfig(c)
	t.Cleanup(func() {
		SetConfig(oldConfig)
		oldConfig.RegisterDb()
	})

	return c
}

// writeHeicXmpFace copies a HEIC sample and embeds one MWG-RS face region.
func writeHeicXmpFace(t *testing.T, c *config.Config, destName string) {
	t.Helper()

	source, err := NewMediaFile(filepath.Join(c.SamplesPath(), "iphone_7.heic"))
	require.NoError(t, err)
	require.NoError(t, source.Copy(destName, false))

	// #nosec G204 -- the test binary and destination come from an isolated config.
	cmd := exec.Command(c.ExifToolBin(),
		"-overwrite_original",
		"-RegionName=HeicAlice",
		"-RegionType=Face",
		"-RegionAreaX=0.5",
		"-RegionAreaY=0.4",
		"-RegionAreaW=0.1",
		"-RegionAreaH=0.15",
		"-RegionAreaUnit=normalized",
		destName,
	)
	output, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "embedding HEIC XMP failed: %s", output)
}

// writeNamedXmp copies the sidecar fixture and replaces its person name.
func writeNamedXmp(t *testing.T, destName, name string) {
	t.Helper()

	xmpData, err := os.ReadFile("testdata/xmp-faces/sidecar.jpg.xmp") //nolint:gosec // test fixture
	require.NoError(t, err)
	xmpData = []byte(strings.ReplaceAll(string(xmpData), "Cara", name))
	require.NoError(t, os.WriteFile(destName, xmpData, fs.ModeFile)) //nolint:gosec // isolated test path
}

// emptyRegionListXmp is the sidecar fixture with its region container kept but
// emptied, which is what an editor writes once the user deletes the last face.
const emptyRegionListXmp = `<?xpacket begin='' id='W5M0MpCehiHzreSzNTczkc9d'?>
<x:xmpmeta xmlns:x='adobe:ns:meta/'>
<rdf:RDF xmlns:rdf='http://www.w3.org/1999/02/22-rdf-syntax-ns#'>
 <rdf:Description rdf:about=''
  xmlns:MP='http://ns.microsoft.com/photo/1.2/'
  xmlns:MPRI='http://ns.microsoft.com/photo/1.2/t/RegionInfo#'>
  <MP:RegionInfo rdf:parseType='Resource'>
   <MPRI:Regions><rdf:Bag/></MPRI:Regions>
  </MP:RegionInfo>
 </rdf:Description>
</rdf:RDF>
</x:xmpmeta>
<?xpacket end='w'?>`

// ratingOnlyXmp is a well-formed sidecar that declares no region container, as
// written by editors that only store develop settings or a rating.
const ratingOnlyXmp = `<?xpacket begin='' id='W5M0MpCehiHzreSzNTczkc9d'?>
<x:xmpmeta xmlns:x='adobe:ns:meta/'>
<rdf:RDF xmlns:rdf='http://www.w3.org/1999/02/22-rdf-syntax-ns#'>
 <rdf:Description rdf:about='' xmlns:xmp='http://ns.adobe.com/xap/1.0/' xmp:Rating='4'/>
</rdf:RDF>
</x:xmpmeta>
<?xpacket end='w'?>`

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
		changes, err := reconcileXmpFaces(meta.FaceRegions{Faces: []meta.Face{region("Dave", xmpArea)}, Declared: true}, file, file.Markers())
		require.NoError(t, err)
		assert.Zero(t, changes, "unchanged XMP must not report marker changes")
		_, err = file.SaveMarkers()
		require.NoError(t, err)
		second, err := entity.FindMarkers(file.FileUID)
		require.NoError(t, err)
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

	t.Run("RemovedRegionDeletesPureXmpMarker", func(t *testing.T) {
		file := newXmpFile(t)
		stale := addXmpMarker(t, file, xmpArea, entity.SrcXmp, entity.SrcXmp, "", "Removed", false)

		markers := runReconcile(t, file, nil)
		assert.Empty(t, markers)
		assert.Nil(t, entity.FindMarker(stale.MarkerUID))
	})

	t.Run("MovedRegionReplacesPureXmpMarker", func(t *testing.T) {
		file := newXmpFile(t)
		first := runReconcile(t, file, []meta.Face{region("Moved", xmpArea)})
		require.Len(t, first, 1)

		second := runReconcile(t, file, []meta.Face{region("Moved", xmpFarArea)})
		require.Len(t, second, 1)
		assert.NotEqual(t, first[0].MarkerUID, second[0].MarkerUID)
		assert.InDelta(t, xmpFarArea.X, second[0].X, 0.001)
	})
	t.Run("OverlappingMovedRegionUpdatesPureXmpMarker", func(t *testing.T) {
		file := newXmpFile(t)
		first := runReconcile(t, file, []meta.Face{region("Moved Nearby", xmpArea)})
		require.Len(t, first, 1)

		movedArea := xmpArea
		movedArea.X += 0.12
		second := runReconcile(t, file, []meta.Face{region("Moved Nearby", movedArea)})
		require.Len(t, second, 1)
		assert.Equal(t, first[0].MarkerUID, second[0].MarkerUID)
		assert.InDelta(t, movedArea.X, second[0].X, 0.001)

		saved := entity.FindMarker(second[0].MarkerUID)
		require.NotNil(t, saved)
		assert.InDelta(t, movedArea.X, saved.X, 0.001, "updated XMP geometry must be persisted")
	})

	t.Run("RenamedRegionUpdatesPureXmpMarker", func(t *testing.T) {
		file := newXmpFile(t)
		first := runReconcile(t, file, []meta.Face{region("Before", xmpArea)})
		require.Len(t, first, 1)

		second := runReconcile(t, file, []meta.Face{region("After", xmpArea)})
		require.Len(t, second, 1)
		assert.Equal(t, first[0].MarkerUID, second[0].MarkerUID)
		assert.Equal(t, "After", second[0].MarkerName)
	})

	t.Run("UnnamedRegionClearsPriorXmpName", func(t *testing.T) {
		file := newXmpFile(t)
		first := runReconcile(t, file, []meta.Face{region("Named", xmpArea)})
		require.Len(t, first, 1)

		second := runReconcile(t, file, []meta.Face{region("", xmpArea)})
		require.Len(t, second, 1)
		assert.Equal(t, first[0].MarkerUID, second[0].MarkerUID)
		assert.Empty(t, second[0].MarkerName)
		assert.Empty(t, second[0].SubjUID)
		assert.True(t, second[0].MarkerReview)
	})

	t.Run("RejectedXmpMarkerIsPreserved", func(t *testing.T) {
		file := newXmpFile(t)
		rejected := addXmpMarker(t, file, xmpArea, entity.SrcXmp, entity.SrcXmp, "", "Rejected", true)

		markers := runReconcile(t, file, nil)
		require.Len(t, markers, 1)
		assert.Equal(t, rejected.MarkerUID, markers[0].MarkerUID)
		assert.True(t, markers[0].MarkerInvalid)
	})

	t.Run("ManualXmpMarkerIsPreserved", func(t *testing.T) {
		file := newXmpFile(t)
		manual := newXmpSubject(t, "Manual Preserve", entity.SrcManual)
		kept := addXmpMarker(t, file, xmpArea, entity.SrcXmp, entity.SrcManual, manual.SubjUID, manual.SubjName, false)

		markers := runReconcile(t, file, nil)
		require.Len(t, markers, 1)
		assert.Equal(t, kept.MarkerUID, markers[0].MarkerUID)
		assert.Equal(t, entity.SrcManual, markers[0].SubjSrc)
	})

	t.Run("RemovedXmpNameClearsUnclusteredAiMarker", func(t *testing.T) {
		file := newXmpFile(t)
		xmpSubject := newXmpSubject(t, "Temporary XMP", entity.SrcXmp)
		stale := addXmpMarker(t, file, xmpArea, entity.SrcImage, entity.SrcXmp, xmpSubject.SubjUID, xmpSubject.SubjName, false)

		markers := runReconcile(t, file, nil)
		require.Len(t, markers, 1)
		assert.Equal(t, stale.MarkerUID, markers[0].MarkerUID)
		assert.Empty(t, markers[0].MarkerName)
		assert.Empty(t, markers[0].SubjUID)
		assert.Empty(t, markers[0].SubjSrc)
		assert.True(t, markers[0].MarkerReview)
	})

	t.Run("RemovedXmpNameRestoresClusteredAiName", func(t *testing.T) {
		file := newXmpFile(t)
		autoSubject := newXmpSubject(t, "Restored Auto", entity.SrcAuto)
		xmpSubject := newXmpSubject(t, "Temporary Cluster XMP", entity.SrcXmp)
		xmpHashSeq++
		cluster := &entity.Face{
			ID:      fmt.Sprintf("XMPRESTORE%032x", xmpHashSeq),
			FaceSrc: entity.SrcAuto,
			SubjUID: autoSubject.SubjUID,
			Samples: 1,
		}
		require.NoError(t, entity.Db().Create(cluster).Error)

		stale := addXmpMarker(t, file, xmpArea, entity.SrcImage, entity.SrcXmp, xmpSubject.SubjUID, xmpSubject.SubjName, false)
		stale.FaceID = cluster.ID
		require.NoError(t, stale.Update("FaceID", cluster.ID))

		markers := runReconcile(t, file, nil)
		require.Len(t, markers, 1)
		assert.Equal(t, "Restored Auto", markers[0].MarkerName)
		assert.Equal(t, autoSubject.SubjUID, markers[0].SubjUID)
		assert.Equal(t, entity.SrcAuto, markers[0].SubjSrc)
		assert.False(t, markers[0].MarkerReview)
	})

	t.Run("OverlappingRegionsKeepDistinctMarkers", func(t *testing.T) {
		file := newXmpFile(t)
		// Two overlapping face rectangles: each region must bind its own marker
		// instead of both collapsing onto the first, which would delete the
		// second as stale.
		areaA := crop.Area{Name: "face", X: 0.40, Y: 0.40, W: 0.10, H: 0.10}
		areaB := crop.Area{Name: "face", X: 0.45, Y: 0.40, W: 0.10, H: 0.10}
		markerA := addXmpMarker(t, file, areaA, entity.SrcXmp, entity.SrcXmp, "", "Ann", false)
		markerB := addXmpMarker(t, file, areaB, entity.SrcXmp, entity.SrcXmp, "", "Bea", false)

		markers := runReconcile(t, file, []meta.Face{region("Ann", areaA), region("Bea", areaB)})
		require.Len(t, markers, 2, "overlapping regions must not delete a distinct marker")
		survived := map[string]bool{}
		for _, m := range markers {
			survived[m.MarkerUID] = true
		}
		assert.True(t, survived[markerA.MarkerUID] && survived[markerB.MarkerUID], "both original markers must survive")
	})

	t.Run("PartialSetDoesNotDeleteUnmatched", func(t *testing.T) {
		file := newXmpFile(t)
		stale := addXmpMarker(t, file, xmpArea, entity.SrcXmp, entity.SrcXmp, "", "Keep", false)

		changed, err := reconcileXmpFaces(meta.FaceRegions{Declared: true, Partial: true}, file, file.Markers())
		require.NoError(t, err)
		assert.Zero(t, changed, "a partial parse must not report deletions")
		_, err = file.SaveMarkers()
		require.NoError(t, err)
		assert.NotNil(t, entity.FindMarker(stale.MarkerUID), "a partial parse must not delete unmatched markers")
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
		faces := mustCollectXmpFaces(t, m)
		require.Len(t, faces, 1)
		assert.Equal(t, "Alice", faces[0].Name)
		assert.InDelta(t, 0.45, faces[0].X, 0.01)
	})

	t.Run("Sidecar", func(t *testing.T) {
		m, err := NewMediaFile("testdata/xmp-faces/sidecar.jpg")
		require.NoError(t, err)
		faces := mustCollectXmpFaces(t, m)
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
		faces := mustCollectXmpFaces(t, m)
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

// TestCollectXmpFaces_AuthoritativeSidecar verifies sidecar selection and fallback metadata.
func TestCollectXmpFaces_AuthoritativeSidecar(t *testing.T) {
	c := newXmpIndexConfig(t, "collect-authoritative-xmp")

	t.Run("NewestSidecarWins", func(t *testing.T) {
		dir := filepath.Join(c.OriginalsPath(), "newest")
		require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
		imageName := filepath.Join(dir, "photo.jpg")
		genericName := filepath.Join(dir, "photo.xmp")
		fullName := filepath.Join(dir, "photo.jpg.xmp")
		require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg", imageName, false))
		writeNamedXmp(t, genericName, "New Generic")
		writeNamedXmp(t, fullName, "Old Full")

		oldStamp := time.Unix(1700000000, 0)
		newStamp := oldStamp.Add(time.Second)
		require.NoError(t, os.Chtimes(fullName, oldStamp, oldStamp))
		require.NoError(t, os.Chtimes(genericName, newStamp, newStamp))

		m, err := NewMediaFile(imageName)
		require.NoError(t, err)
		faces := mustCollectXmpFaces(t, m)
		require.Len(t, faces, 1)
		assert.Equal(t, "New Generic", faces[0].Name)
	})

	t.Run("EqualTimeFullNameWins", func(t *testing.T) {
		dir := filepath.Join(c.OriginalsPath(), "tie")
		require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
		imageName := filepath.Join(dir, "photo.jpg")
		genericName := filepath.Join(dir, "photo.xmp")
		fullName := filepath.Join(dir, "photo.jpg.xmp")
		require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg", imageName, false))
		writeNamedXmp(t, genericName, "Generic")
		writeNamedXmp(t, fullName, "Full Name")

		stamp := time.Unix(1700000000, 0)
		require.NoError(t, os.Chtimes(genericName, stamp, stamp))
		require.NoError(t, os.Chtimes(fullName, stamp, stamp))

		m, err := NewMediaFile(imageName)
		require.NoError(t, err)
		faces := mustCollectXmpFaces(t, m)
		require.Len(t, faces, 1)
		assert.Equal(t, "Full Name", faces[0].Name)
	})

	t.Run("SidecarOverridesEmbedded", func(t *testing.T) {
		dir := filepath.Join(c.OriginalsPath(), "override")
		require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
		imageName := filepath.Join(dir, "photo.jpg")
		xmpName := imageName + fs.ExtXMP
		require.NoError(t, fs.Copy("testdata/xmp-faces/embedded-mwg.jpg", imageName, false))
		writeNamedXmp(t, xmpName, "Sidecar Person")

		m, err := NewMediaFile(imageName)
		require.NoError(t, err)
		faces := mustCollectXmpFaces(t, m)
		require.Len(t, faces, 1)
		assert.Equal(t, "Sidecar Person", faces[0].Name)
		assert.False(t, hasFaceName(faces, "Alice"), "embedded faces must not be merged when a sidecar exists")
	})

	t.Run("MalformedSidecarReturnsError", func(t *testing.T) {
		dir := filepath.Join(c.OriginalsPath(), "malformed")
		require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
		imageName := filepath.Join(dir, "photo.jpg")
		require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg", imageName, false))
		require.NoError(t, os.WriteFile(imageName+fs.ExtXMP, []byte("<broken"), fs.ModeFile)) //nolint:gosec // isolated test path

		m, err := NewMediaFile(imageName)
		require.NoError(t, err)
		_, err = collectXmpFaces(m)
		require.Error(t, err)
	})

	t.Run("MissingSidecarOrientationUsesSource", func(t *testing.T) {
		dir := filepath.Join(c.OriginalsPath(), "orientation")
		require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
		imageName := filepath.Join(dir, "photo.jpg")
		require.NoError(t, fs.Copy("testdata/xmp-faces/rotated-o6.jpg", imageName, false))
		writeNamedXmp(t, imageName+fs.ExtXMP, "Fallback Orientation")

		m, err := NewMediaFile(imageName)
		require.NoError(t, err)
		faces := mustCollectXmpFaces(t, m)
		require.Len(t, faces, 1)
		assert.Equal(t, "Fallback Orientation", faces[0].Name)
		assert.InDelta(t, 0.65, faces[0].X, 0.01)
		assert.InDelta(t, 0.30, faces[0].Y, 0.01)
		assert.InDelta(t, 0.15, faces[0].W, 0.01)
		assert.InDelta(t, 0.10, faces[0].H, 0.01)
	})
}

// TestApplyXmpFaces_MalformedSidecarPreservesMarkers verifies non-destructive errors.
func TestApplyXmpFaces_MalformedSidecarPreservesMarkers(t *testing.T) {
	c := newXmpIndexConfig(t, "apply-malformed-xmp")
	dir := filepath.Join(c.OriginalsPath(), "malformed-preserve")
	require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
	imageName := filepath.Join(dir, "photo.jpg")
	require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg", imageName, false))
	require.NoError(t, os.WriteFile(imageName+fs.ExtXMP, []byte("<broken"), fs.ModeFile)) //nolint:gosec // isolated test path

	m, err := NewMediaFile(imageName)
	require.NoError(t, err)
	file := newXmpFile(t)
	existing := addXmpMarker(t, file, xmpArea, entity.SrcXmp, entity.SrcXmp, "", "Keep Me", false)

	saved, _, err := ApplyXmpFaces(m, file)
	require.Error(t, err)
	assert.False(t, saved)
	assert.NotNil(t, entity.FindMarker(existing.MarkerUID), "parse errors must not delete existing XMP markers")
}

// TestCollectXmpFaces_RelatedRawSidecars verifies RAW sidecar naming and locations.
func TestCollectXmpFaces_RelatedRawSidecars(t *testing.T) {
	c := newXmpIndexConfig(t, "collect-related-raw-xmp")

	tests := []struct {
		name    string
		xmpName func(string) string
	}{
		{
			name: "BaseNameInOriginals",
			xmpName: func(dir string) string {
				return filepath.Join(c.OriginalsPath(), dir, "face.xmp")
			},
		},
		{
			name: "FullNameInOriginals",
			xmpName: func(dir string) string {
				return filepath.Join(c.OriginalsPath(), dir, "face.dng.xmp")
			},
		},
		{
			name: "FullNameInSidecar",
			xmpName: func(dir string) string {
				return filepath.Join(c.SidecarPath(), dir, "face.dng.xmp")
			},
		},
		{
			name: "BaseNameInHidden",
			xmpName: func(dir string) string {
				return filepath.Join(c.OriginalsPath(), dir, fs.PPHiddenPathname, "face.xmp")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := strings.ToLower(tc.name)
			rawName := filepath.Join(c.OriginalsPath(), dir, "face.dng")
			previewName := filepath.Join(c.SidecarPath(), dir, "face.dng.jpg")

			raw, err := NewMediaFile(filepath.Join(c.SamplesPath(), "canon_eos_6d.dng"))
			require.NoError(t, err)
			require.NoError(t, raw.Copy(rawName, false))
			require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg", previewName, false))
			require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg.xmp", tc.xmpName(dir), false))

			preview, err := NewMediaFile(previewName)
			require.NoError(t, err)

			faces := mustCollectXmpFaces(t, preview)
			require.True(t, hasFaceName(faces, "Cara"), "related RAW sidecar face not collected: %+v", faces)
		})
	}
}

// hasFaceName reports whether parsed XMP faces contain the specified name.
func hasFaceName(faces []meta.Face, name string) bool {
	for i := range faces {
		if faces[i].Name == name {
			return true
		}
	}

	return false
}

// TestIndexRelated_XmpFacesFromLogicalSource verifies source-to-primary reconciliation.
func TestIndexRelated_XmpFacesFromLogicalSource(t *testing.T) {
	t.Run("HeicEmbedded", func(t *testing.T) {
		c := newXmpIndexConfig(t, "index-related-heic-xmp")
		if c.ExifToolBin() == "" {
			t.Skip("exiftool not configured")
		}

		dir := filepath.Join(c.OriginalsPath(), "heic")
		heicName := filepath.Join(dir, "face.heic")
		previewName := filepath.Join(dir, "face.jpg")
		writeHeicXmpFace(t, c, heicName)
		require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg", previewName, false))

		main, err := NewMediaFile(heicName)
		require.NoError(t, err)
		related, err := main.RelatedFiles(false)
		require.NoError(t, err)
		require.True(t, related.Main.IsHeic())

		opt := IndexOptionsSingle(c)
		opt.Convert = false
		opt.DetectFaces = false
		opt.DetectNsfw = false
		opt.GenerateLabels = false
		opt.ImportFaceTags = true

		result := IndexRelated(related, NewIndex(c, NewConvert(c), NewFiles(), NewPhotos()), opt)
		require.True(t, result.Success(), "HEIC indexing failed: %v", result.Err)

		primary, err := entity.PrimaryFile(result.PhotoUID)
		require.NoError(t, err)
		require.True(t, primary.FilePrimary)
		require.Equal(t, "heic/face.jpg", primary.FileName)

		markers, err := entity.FindMarkers(primary.FileUID)
		require.NoError(t, err)
		require.True(t, hasXmpMarkerName(markers, "HeicAlice"), "embedded HEIC face not imported: %+v", markers)
	})

	rawSidecars := []struct {
		name     string
		fileName string
	}{
		{name: "BaseName", fileName: "face.xmp"},
		{name: "FullName", fileName: "face.dng.xmp"},
	}

	for _, tc := range rawSidecars {
		t.Run("RawSidecar"+tc.name, func(t *testing.T) {
			c := newXmpIndexConfig(t, "index-related-raw-"+strings.ToLower(tc.name))
			dir := filepath.Join(c.OriginalsPath(), strings.ToLower(tc.name))
			rawName := filepath.Join(dir, "face.dng")
			previewName := filepath.Join(dir, "face.jpg")
			xmpName := filepath.Join(dir, tc.fileName)

			raw, err := NewMediaFile(filepath.Join(c.SamplesPath(), "canon_eos_6d.dng"))
			require.NoError(t, err)
			require.NoError(t, raw.Copy(rawName, false))
			require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg", previewName, false))
			require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg.xmp", xmpName, false))

			main, err := NewMediaFile(rawName)
			require.NoError(t, err)
			related, err := main.RelatedFiles(false)
			require.NoError(t, err)
			require.True(t, related.Main.IsRaw())

			opt := IndexOptionsSingle(c)
			opt.Convert = false
			opt.DetectFaces = false
			opt.DetectNsfw = false
			opt.GenerateLabels = false
			opt.ImportFaceTags = true
			if tc.name == "BaseName" {
				opt.ImportFaceTags = false
			}

			ind := NewIndex(c, NewConvert(c), NewFiles(), NewPhotos())
			result := IndexRelated(related, ind, opt)
			require.True(t, result.Success(), "RAW indexing failed: %v", result.Err)

			primary, err := entity.PrimaryFile(result.PhotoUID)
			require.NoError(t, err)
			markers, err := entity.FindMarkers(primary.FileUID)
			require.NoError(t, err)

			if tc.name == "BaseName" {
				require.False(t, hasXmpMarkerName(markers, "Cara"), "disabled XMP import must not create markers")

				related, err = main.RelatedFiles(false)
				require.NoError(t, err)
				opt.ImportFaceTags = true
				opt.Rescan = true
				result = IndexRelated(related, ind, opt)
				require.True(t, result.Success(), "RAW XMP rescan failed: %v", result.Err)

				markers, err = entity.FindMarkers(primary.FileUID)
				require.NoError(t, err)
			}

			require.True(t, hasXmpMarkerName(markers, "Cara"), "RAW sidecar face not imported: %+v", markers)

			if tc.fileName != "face.dng.xmp" {
				return
			}

			xmpData, err := os.ReadFile(xmpName) //nolint:gosec // test reads its temporary XMP fixture
			require.NoError(t, err)
			xmpData = []byte(strings.ReplaceAll(string(xmpData), "Cara", "Dana"))
			require.NoError(t, os.WriteFile(xmpName, xmpData, fs.ModeFile)) //nolint:gosec // test rewrites its temporary XMP fixture
			xmpStamp := time.Now().Add(2 * time.Second)
			require.NoError(t, os.Chtimes(xmpName, xmpStamp, xmpStamp))

			xmp, err := NewMediaFile(xmpName)
			require.NoError(t, err)
			incremental := RelatedFiles{Main: main, Files: MediaFiles{xmp}}
			opt.Rescan = false

			result = IndexRelated(incremental, ind, opt)
			require.True(t, result.Success(), "incremental RAW sidecar indexing failed: %v", result.Err)

			markers, err = entity.FindMarkers(primary.FileUID)
			require.NoError(t, err)
			require.Len(t, markers, 1)
			assert.Equal(t, "Dana", markers[0].MarkerName)
		})
	}
}

// TestApplyDetectedFaces covers the deferred vision/metadata worker path: when
// XMP face-tag import is enabled it reconciles XMP regions onto the primary
// file's markers even without AI detection, and it stays a no-op when the
// toggle is off, the media file is missing, or no work is required.
func TestApplyDetectedFaces(t *testing.T) {
	c := config.TestConfig()
	t.Cleanup(func() { c.Options().XMPFaces = false })

	// hasXmpMarker reports whether a SrcXmp marker with the given name exists.
	hasXmpMarker := func(markers entity.Markers, name string) bool {
		for i := range markers {
			if markers[i].MarkerSrc == entity.SrcXmp && markers[i].MarkerName == name {
				return true
			}
		}
		return false
	}

	t.Run("ImportsWhenEnabledAndDetectionDeferred", func(t *testing.T) {
		c.Options().XMPFaces = true

		file := newXmpFile(t)
		m, err := NewMediaFile("testdata/xmp-faces/sidecar.jpg")
		require.NoError(t, err)

		// No detected faces: the worker path must still import the XMP region.
		saved, count, err := ApplyDetectedFaces(m, file, nil)
		require.NoError(t, err)
		assert.True(t, saved, "XMP-only import must persist markers")
		assert.GreaterOrEqual(t, count, 1, "the imported face must be counted")

		markers, err := entity.FindMarkers(file.FileUID)
		require.NoError(t, err)
		require.True(t, hasXmpMarker(markers, "Cara"), "sidecar region 'Cara' must import via ApplyDetectedFaces, got %+v", markers)
		for i := range markers {
			if markers[i].MarkerName == "Cara" {
				assert.NotEmpty(t, markers[i].SubjUID, "imported name must link to a Person")
				assert.Equal(t, entity.SrcXmp, markers[i].SubjSrc)
			}
		}
	})

	t.Run("SkipsWhenDisabled", func(t *testing.T) {
		c.Options().XMPFaces = false

		file := newXmpFile(t)
		m, err := NewMediaFile("testdata/xmp-faces/sidecar.jpg")
		require.NoError(t, err)

		saved, count, err := ApplyDetectedFaces(m, file, nil)
		require.NoError(t, err)
		assert.False(t, saved, "no persistence when the toggle is off and no faces detected")
		assert.Equal(t, 0, count)

		markers, err := entity.FindMarkers(file.FileUID)
		require.NoError(t, err)
		assert.Empty(t, markers, "no markers may be created when the toggle is off")
	})

	t.Run("SkipsWhenMediaFileNil", func(t *testing.T) {
		c.Options().XMPFaces = true

		file := newXmpFile(t)

		// A nil media file disables XMP import even when the toggle is on,
		// because there is nothing to read the embedded XMP or sidecar from.
		saved, _, err := ApplyDetectedFaces(nil, file, nil)
		require.NoError(t, err)
		assert.False(t, saved)

		markers, err := entity.FindMarkers(file.FileUID)
		require.NoError(t, err)
		assert.Empty(t, markers)
	})

	t.Run("NilFile", func(t *testing.T) {
		_, _, err := ApplyDetectedFaces(nil, nil, nil)
		assert.Error(t, err, "a nil file must be rejected")
	})
}

// sampleDetectedFace returns a minimal AI-detected face with a single embedding.
func sampleDetectedFace() face.Face {
	return face.Face{
		Rows:       480,
		Cols:       720,
		Score:      45,
		Area:       face.NewArea("face", 250, 200, 10),
		Embeddings: face.Embeddings{{0.1, 0.2, 0.3}},
	}
}

// TestApplyDetectedFaces_MalformedSidecarKeepsDetectedFaces verifies that a
// broken XMP sidecar does not discard AI-detected faces: the XMP error is logged
// and the detected marker is still persisted.
func TestApplyDetectedFaces_MalformedSidecarKeepsDetectedFaces(t *testing.T) {
	c := newXmpIndexConfig(t, "apply-detected-malformed-xmp")
	c.Options().XMPFaces = true

	dir := filepath.Join(c.OriginalsPath(), "detected-malformed")
	require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
	imageName := filepath.Join(dir, "photo.jpg")
	require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg", imageName, false))
	require.NoError(t, os.WriteFile(imageName+fs.ExtXMP, []byte("<broken"), fs.ModeFile)) //nolint:gosec // isolated test path

	m, err := NewMediaFile(imageName)
	require.NoError(t, err)

	file := newXmpFile(t)
	saved, count, err := ApplyDetectedFaces(m, file, face.Faces{sampleDetectedFace()})
	require.NoError(t, err, "a malformed sidecar must not abort saving detected faces")
	assert.True(t, saved, "detected faces must be persisted despite the XMP error")
	assert.GreaterOrEqual(t, count, 1)

	markers, err := entity.FindMarkers(file.FileUID)
	require.NoError(t, err)
	found := false
	for i := range markers {
		if markers[i].MarkerSrc == entity.SrcImage {
			found = true
		}
	}
	assert.True(t, found, "the detected face marker must be saved, got %+v", markers)
}

// TestIsXmpFaceSource covers the supported-still-image predicate.
func TestIsXmpFaceSource(t *testing.T) {
	c := config.TestConfig()
	jpg, err := NewMediaFile(c.SamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
	require.NoError(t, err)
	assert.True(t, isXmpFaceSource(jpg), "a still JPEG is an XMP face source")
	video, err := NewMediaFile(c.SamplesPath() + "/gopher-video.mp4")
	require.NoError(t, err)
	assert.False(t, isXmpFaceSource(video), "a video is not an XMP face source")
	assert.False(t, isXmpFaceSource(nil), "a nil media file is not an XMP face source")
}

// TestIsFullNameSidecar covers full-name vs generic sidecar matching.
func TestIsFullNameSidecar(t *testing.T) {
	m, err := NewMediaFile(config.TestConfig().SamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
	require.NoError(t, err)
	full := m.FileName() + fs.ExtXMP
	generic := strings.TrimSuffix(m.FileName(), filepath.Ext(m.FileName())) + fs.ExtXMP
	assert.True(t, isFullNameSidecar(full, m), "photo.jpg.xmp must match the full-name form")
	assert.False(t, isFullNameSidecar(generic, m), "photo.xmp must not match the full-name form")
	assert.False(t, isFullNameSidecar(full, nil), "a nil source never matches")
}

// TestFaceOptions covers the nil-source zero value and the MetaData mirror.
func TestFaceOptions(t *testing.T) {
	assert.Equal(t, meta.FaceOptions{}, faceOptions(nil), "a nil source yields zero options")
	m, err := NewMediaFile(config.TestConfig().SamplesPath() + "/iphone_7.heic")
	require.NoError(t, err)
	got := faceOptions(m)
	data := m.MetaData()
	assert.Equal(t, data.Width, got.Width)
	assert.Equal(t, data.Height, got.Height)
	assert.Equal(t, data.Orientation, got.Orientation)
}

// TestXmpFaceSources covers the logical still-image source set: a nil input and
// a video yield nil, a standalone image resolves to itself, and a preview also
// includes its distinct RAW original.
func TestXmpFaceSources(t *testing.T) {
	c := newXmpIndexConfig(t, "xmp-face-sources")

	t.Run("NilReturnsNil", func(t *testing.T) {
		assert.Nil(t, xmpFaceSources(nil), "a nil media file has no XMP face sources")
	})

	t.Run("VideoReturnsNil", func(t *testing.T) {
		video, err := NewMediaFile(filepath.Join(c.SamplesPath(), "gopher-video.mp4"))
		require.NoError(t, err)
		assert.Nil(t, xmpFaceSources(video), "a video is not an XMP face source")
	})

	t.Run("StandaloneImageResolvesToSelf", func(t *testing.T) {
		dir := filepath.Join(c.OriginalsPath(), "standalone")
		require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
		imageName := filepath.Join(dir, "photo.jpg")
		require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg", imageName, false))

		m, err := NewMediaFile(imageName)
		require.NoError(t, err)
		sources := xmpFaceSources(m)
		require.Len(t, sources, 1, "a standalone still image is its own only source")
		assert.Equal(t, m.FileName(), sources[0].FileName())
	})

	t.Run("PreviewIncludesLogicalRawSource", func(t *testing.T) {
		dir := "preview-raw"
		rawName := filepath.Join(c.OriginalsPath(), dir, "face.dng")
		previewName := filepath.Join(c.SidecarPath(), dir, "face.dng.jpg")

		raw, err := NewMediaFile(filepath.Join(c.SamplesPath(), "canon_eos_6d.dng"))
		require.NoError(t, err)
		require.NoError(t, raw.Copy(rawName, false))
		require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg", previewName, false))

		preview, err := NewMediaFile(previewName)
		require.NoError(t, err)
		sources := xmpFaceSources(preview)
		require.Len(t, sources, 2, "a preview and its distinct RAW original must both be returned")
		assert.Equal(t, preview.FileName(), sources[0].FileName())
		assert.True(t, sources[1].IsRaw(), "the second source must be the logical RAW original")
	})
}

// TestApplyXmpFaces covers the guard branches and the success path: a nil file
// is rejected, a nil media file or an empty file hash is a no-op, and a still
// image with a sidecar region imports and persists a named marker.
func TestApplyXmpFaces(t *testing.T) {
	c := newXmpIndexConfig(t, "apply-xmp-faces")

	t.Run("NilFile", func(t *testing.T) {
		saved, count, err := ApplyXmpFaces(nil, nil)
		assert.Error(t, err, "a nil file must be rejected")
		assert.False(t, saved)
		assert.Equal(t, 0, count)
	})

	t.Run("NilMediaFileNoOp", func(t *testing.T) {
		file := newXmpFile(t)
		saved, count, err := ApplyXmpFaces(nil, file)
		require.NoError(t, err, "a nil media file must be a no-op, not an error")
		assert.False(t, saved)
		assert.Equal(t, 0, count)
		markers, err := entity.FindMarkers(file.FileUID)
		require.NoError(t, err)
		assert.Empty(t, markers, "a no-op must not create markers")
	})

	t.Run("EmptyFileHashNoOp", func(t *testing.T) {
		m, err := NewMediaFile("testdata/xmp-faces/sidecar.jpg")
		require.NoError(t, err)
		saved, count, err := ApplyXmpFaces(m, &entity.File{})
		require.NoError(t, err, "an empty file hash must be a no-op, not an error")
		assert.False(t, saved)
		assert.Equal(t, 0, count)
	})

	t.Run("Success", func(t *testing.T) {
		dir := filepath.Join(c.OriginalsPath(), "apply-success")
		require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
		imageName := filepath.Join(dir, "photo.jpg")
		require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg", imageName, false))
		require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg.xmp", imageName+fs.ExtXMP, false))

		m, err := NewMediaFile(imageName)
		require.NoError(t, err)
		file := newXmpFile(t)

		saved, count, err := ApplyXmpFaces(m, file)
		require.NoError(t, err)
		assert.True(t, saved, "a sidecar region must persist a marker")
		assert.GreaterOrEqual(t, count, 1, "the imported face must be counted")

		markers, err := entity.FindMarkers(file.FileUID)
		require.NoError(t, err)
		assert.True(t, hasXmpMarkerName(markers, "Cara"), "sidecar region 'Cara' must import, got %+v", markers)
	})
}

// TestCollectXmpFaces_UndeclaredRegions verifies that a sidecar which tracks no
// face regions is not read as an authoritative "this image has no faces".
func TestCollectXmpFaces_UndeclaredRegions(t *testing.T) {
	c := newXmpIndexConfig(t, "collect-undeclared-xmp")
	if c.ExifToolBin() == "" {
		t.Skip("exiftool not configured")
	}

	t.Run("RatingOnlySidecarFallsBackToEmbedded", func(t *testing.T) {
		dir := filepath.Join(c.OriginalsPath(), "rating-only")
		require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
		imageName := filepath.Join(dir, "photo.jpg")
		require.NoError(t, fs.Copy("testdata/xmp-faces/embedded-mwg.jpg", imageName, false))

		m, err := NewMediaFile(imageName)
		require.NoError(t, err)
		require.NoError(t, m.CreateExifToolJson(NewConvert(c)))
		require.NoError(t, os.WriteFile(imageName+fs.ExtXMP, []byte(ratingOnlyXmp), fs.ModeFile)) //nolint:gosec // isolated test path

		regions, err := collectXmpFaces(m)
		require.NoError(t, err)
		assert.True(t, regions.Declared, "the embedded packet declares the regions")
		require.Len(t, regions.Faces, 1, "a region-less sidecar must not suppress the embedded regions")
		assert.Equal(t, "Alice", regions.Faces[0].Name)
	})
	t.Run("EmptyRegionListStaysAuthoritative", func(t *testing.T) {
		dir := filepath.Join(c.OriginalsPath(), "empty-list")
		require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
		imageName := filepath.Join(dir, "photo.jpg")
		require.NoError(t, fs.Copy("testdata/xmp-faces/embedded-mwg.jpg", imageName, false))

		m, err := NewMediaFile(imageName)
		require.NoError(t, err)
		require.NoError(t, m.CreateExifToolJson(NewConvert(c)))
		require.NoError(t, os.WriteFile(imageName+fs.ExtXMP, []byte(emptyRegionListXmp), fs.ModeFile)) //nolint:gosec // isolated test path

		regions, err := collectXmpFaces(m)
		require.NoError(t, err)
		assert.True(t, regions.Declared, "an empty region container still declares the region set")
		assert.Empty(t, regions.Faces, "an emptied region list must win over the embedded packet")
	})
}

// TestApplyXmpFaces_RegionlessSidecarKeepsMarkers verifies that an ordinary
// rating-only sidecar never deletes previously imported face markers.
func TestApplyXmpFaces_RegionlessSidecarKeepsMarkers(t *testing.T) {
	c := newXmpIndexConfig(t, "regionless-keeps-markers")
	dir := filepath.Join(c.OriginalsPath(), "regionless-keep")
	require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
	imageName := filepath.Join(dir, "photo.jpg")
	require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg", imageName, false))
	require.NoError(t, os.WriteFile(imageName+fs.ExtXMP, []byte(ratingOnlyXmp), fs.ModeFile)) //nolint:gosec // isolated test path

	m, err := NewMediaFile(imageName)
	require.NoError(t, err)
	file := newXmpFile(t)
	existing := addXmpMarker(t, file, xmpArea, entity.SrcXmp, entity.SrcXmp, "", "Cara", false)

	saved, _, err := ApplyXmpFaces(m, file)
	require.NoError(t, err)
	assert.False(t, saved, "a region-less sidecar has nothing to reconcile")
	assert.NotNil(t, entity.FindMarker(existing.MarkerUID), "a region-less sidecar must not delete imported markers")
}

// TestReconcileXmpFaces_UndeclaredSkipsSweep verifies the delete-sweep gate.
func TestReconcileXmpFaces_UndeclaredSkipsSweep(t *testing.T) {
	config.TestConfig()

	t.Run("UndeclaredKeepsUnmatched", func(t *testing.T) {
		file := newXmpFile(t)
		stale := addXmpMarker(t, file, xmpArea, entity.SrcXmp, entity.SrcXmp, "", "Keep", false)
		changed, err := reconcileXmpFaces(meta.FaceRegions{}, file, file.Markers())
		require.NoError(t, err)
		assert.Zero(t, changed, "an undeclared set must not report deletions")
		assert.NotNil(t, entity.FindMarker(stale.MarkerUID), "an undeclared set must not delete unmatched markers")
	})
	t.Run("DeclaredEmptyDeletesUnmatched", func(t *testing.T) {
		file := newXmpFile(t)
		stale := addXmpMarker(t, file, xmpArea, entity.SrcXmp, entity.SrcXmp, "", "Drop", false)
		changed, err := reconcileXmpFaces(meta.FaceRegions{Declared: true}, file, file.Markers())
		require.NoError(t, err)
		assert.Equal(t, 1, changed, "a declared empty set must delete the stale marker")
		assert.Nil(t, entity.FindMarker(stale.MarkerUID), "a declared empty set must delete unmatched markers")
	})
}

// TestApplyXmpFaceRegions verifies the collect-free apply entry point.
func TestApplyXmpFaceRegions(t *testing.T) {
	config.TestConfig()

	t.Run("NilFile", func(t *testing.T) {
		saved, count, err := applyXmpFaceRegions(meta.FaceRegions{Declared: true}, nil)
		require.Error(t, err)
		assert.False(t, saved)
		assert.Equal(t, 0, count)
	})
	t.Run("EmptyFileHashNoOp", func(t *testing.T) {
		saved, count, err := applyXmpFaceRegions(meta.FaceRegions{Declared: true}, &entity.File{})
		require.NoError(t, err)
		assert.False(t, saved)
		assert.Equal(t, 0, count)
	})
	t.Run("Success", func(t *testing.T) {
		file := newXmpFile(t)
		regions := meta.FaceRegions{Faces: []meta.Face{region("Nina", xmpArea)}, Declared: true}
		saved, count, err := applyXmpFaceRegions(regions, file)
		require.NoError(t, err)
		assert.True(t, saved)
		assert.GreaterOrEqual(t, count, 1)
		markers, err := entity.FindMarkers(file.FileUID)
		require.NoError(t, err)
		assert.True(t, hasXmpMarkerName(markers, "Nina"), "the region must import, got %+v", markers)
	})
}

// TestCollectXmpFaces_UnreadableSourceSkipped verifies that a related source
// whose metadata cannot be parsed does not discard the primary file's own
// regions, and that the resulting set is not treated as authoritative.
func TestCollectXmpFaces_UnreadableSourceSkipped(t *testing.T) {
	c := newXmpIndexConfig(t, "collect-unreadable-source")
	if c.ExifToolBin() == "" {
		t.Skip("exiftool not configured")
	}

	dir := filepath.Join(c.OriginalsPath(), "unreadable-source")
	require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
	imageName := filepath.Join(dir, "photo.jpg")
	rawName := filepath.Join(dir, "photo.cr2")
	require.NoError(t, fs.Copy("testdata/xmp-faces/embedded-mwg.jpg", imageName, false))
	require.NoError(t, os.WriteFile(rawName, []byte("not a raw image"), fs.ModeFile)) //nolint:gosec // isolated test path

	m, err := NewMediaFile(imageName)
	require.NoError(t, err)
	require.NoError(t, m.CreateExifToolJson(NewConvert(c)))

	// Seeding pins the source list to exactly these two files, so the test does
	// not depend on how RelatedFiles happens to resolve the group.
	raw, err := NewMediaFile(rawName)
	require.NoError(t, err)
	require.True(t, isXmpFaceSource(raw), "the unreadable sibling must qualify as a source")
	m.SetRelatedMain(raw)

	regions, err := collectXmpFaces(m)
	require.NoError(t, err, "an unreadable source must not fail the import")
	require.Len(t, regions.Faces, 1, "the primary file's own regions must survive")
	assert.Equal(t, "Alice", regions.Faces[0].Name)
	assert.True(t, regions.Declared)
	assert.True(t, regions.Partial, "an unread source must suppress the delete sweep")
}

// TestRestoreMarkerName verifies that dropping an obsolete XMP name restores the
// marker's clustered name when one exists and flags it for review otherwise.
func TestRestoreMarkerName(t *testing.T) {
	config.TestConfig()

	t.Run("NilMarker", func(t *testing.T) {
		changed, err := restoreMarkerName(nil)
		require.NoError(t, err)
		assert.False(t, changed)
	})
	t.Run("NonXmpSubjSrcNoOp", func(t *testing.T) {
		file := newXmpFile(t)
		m := addXmpMarker(t, file, xmpArea, entity.SrcXmp, entity.SrcManual, "", "Manual Name", false)
		changed, err := restoreMarkerName(m)
		require.NoError(t, err)
		assert.False(t, changed, "a manually assigned name must not be restored away")
		assert.Equal(t, "Manual Name", m.MarkerName)
	})
	t.Run("ClearsWithoutClusteredFace", func(t *testing.T) {
		file := newXmpFile(t)
		m := addXmpMarker(t, file, xmpArea, entity.SrcXmp, entity.SrcXmp, "", "Gone", false)
		changed, err := restoreMarkerName(m)
		require.NoError(t, err)
		assert.True(t, changed)
		assert.Empty(t, m.MarkerName)
		assert.Empty(t, m.SubjUID)
		assert.Empty(t, m.SubjSrc)
		assert.True(t, m.MarkerReview, "an unnamed marker must be flagged for review")

		saved := entity.FindMarker(m.MarkerUID)
		require.NotNil(t, saved)
		assert.Empty(t, saved.MarkerName, "the cleared name must be persisted")
	})
	t.Run("RestoresClusteredName", func(t *testing.T) {
		autoPerson := newXmpSubject(t, "ClusteredCara", entity.SrcAuto)
		cluster := &entity.Face{
			ID:            "XMPRESTORE00000000000000000000000000000001",
			FaceSrc:       entity.SrcAuto,
			SubjUID:       autoPerson.SubjUID,
			EmbeddingJSON: (face.Embeddings{face.NullEmbedding}).JSON(),
			Samples:       2,
		}
		require.NoError(t, entity.Db().Create(cluster).Error)

		file := newXmpFile(t)
		m := addXmpMarker(t, file, xmpArea, entity.SrcXmp, entity.SrcXmp, "", "XmpName", false)
		m.FaceID = cluster.ID
		require.NoError(t, m.Updates(entity.Values{"face_id": m.FaceID}))

		changed, err := restoreMarkerName(m)
		require.NoError(t, err)
		assert.True(t, changed)
		assert.Equal(t, "ClusteredCara", m.MarkerName)
		assert.Equal(t, autoPerson.SubjUID, m.SubjUID)
		assert.Equal(t, entity.SrcAuto, m.SubjSrc)
		assert.False(t, m.MarkerReview, "a restored clustered name needs no review")
	})
	t.Run("Idempotent", func(t *testing.T) {
		// The second call is short-circuited by the SubjSrc guard: a restore
		// always clears SubjSrc away from SrcXmp, so a restored marker is never
		// a candidate again.
		file := newXmpFile(t)
		m := addXmpMarker(t, file, xmpArea, entity.SrcXmp, entity.SrcXmp, "", "Gone", false)
		first, err := restoreMarkerName(m)
		require.NoError(t, err)
		require.True(t, first)
		second, err := restoreMarkerName(m)
		require.NoError(t, err)
		assert.False(t, second, "restoring an already restored marker must report no change")
	})
}
