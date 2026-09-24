package photoprism

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestNewRegeneratedFaces(t *testing.T) {
	detected := face.Faces{
		{Rows: 100, Cols: 100, Score: 90, Area: face.NewArea("face", 20, 20, 30)},
		{Rows: 100, Cols: 100, Score: 40, Area: face.NewArea("face", 60, 60, 30)},
		{Rows: 100, Cols: 100, Score: 90, Area: face.NewArea("face", 80, 80, 12)},
	}

	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, []int{0}, newRegeneratedFaces(detected, nil, 65, 25, 10))
	})
	t.Run("Claimed", func(t *testing.T) {
		assert.Empty(t, newRegeneratedFaces(detected, map[int]bool{0: true}, 65, 25, 10))
	})
	t.Run("NoScoreThreshold", func(t *testing.T) {
		assert.Equal(t, []int{0, 1}, newRegeneratedFaces(detected, nil, face.NoScoreThreshold, 25, 10))
	})
	t.Run("Retry", func(t *testing.T) {
		// None clears the ordinary size, so the retry floor applies, as it does when indexing.
		assert.Equal(t, []int{0, 2}, newRegeneratedFaces(detected, nil, 65, 40, 10))
		assert.Equal(t, []int{0}, newRegeneratedFaces(detected, nil, 65, 40, 20))
	})
	t.Run("NoRetryAfterClaimedFace", func(t *testing.T) {
		// A claimed face that clears the ordinary size means the first pass found one.
		assert.Empty(t, newRegeneratedFaces(detected, map[int]bool{0: true}, 65, 25, 10))
	})
	t.Run("RetryDisabled", func(t *testing.T) {
		assert.Empty(t, newRegeneratedFaces(detected, nil, 65, 40, 0))
	})
	t.Run("NoDetections", func(t *testing.T) {
		assert.Empty(t, newRegeneratedFaces(nil, nil, 65, 25, 10))
	})
	t.Run("FractionalScore", func(t *testing.T) {
		// Compared as face.Detect compares it: the recorded score against the cutoff as configured.
		faces := face.Faces{{Rows: 100, Cols: 100, Score: 65, Area: face.NewArea("face", 20, 20, 30)}}

		assert.Equal(t, []int{0}, newRegeneratedFaces(faces, nil, 65, 25, 10))
		assert.Empty(t, newRegeneratedFaces(faces, nil, 65.3, 25, 10))
	})
}

func TestFaceRegeneration(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		stats := &FaceRegeneration{}
		stats.add("fs6sg6bw45bn0004", faceRegenerationResult{Updated: 2, Added: 1, Removed: 1, Kept: 3, Failed: 1})
		stats.add("fs6sg6bq45bnlqd0", faceRegenerationResult{Updated: 1})

		assert.Equal(t, int64(2), stats.Files.Load())
		assert.True(t, stats.Processed("fs6sg6bw45bn0004"))
		assert.False(t, stats.Processed("fs6sg6bw45bn0005"))
		assert.Equal(t, "3 updated, 1 added, 1 removed, 3 kept unmatched, 1 failed", stats.String())

		stats.addError()

		assert.Equal(t, "3 updated, 1 added, 1 removed, 3 kept unmatched, 1 failed, 1 file could not be processed", stats.String())
	})
	t.Run("Nil", func(t *testing.T) {
		var stats *FaceRegeneration

		stats.add("fs6sg6bw45bn0004", faceRegenerationResult{Updated: 1})
		stats.addError()

		assert.Equal(t, "", stats.String())
		assert.False(t, stats.Processed("fs6sg6bw45bn0004"))
	})
}

func TestMatchRegeneratedFaces(t *testing.T) {
	detected := face.Faces{
		{Rows: 100, Cols: 100, Score: 90, Area: face.NewArea("face", 30, 30, 20)},
		{Rows: 100, Cols: 100, Score: 90, Area: face.NewArea("face", 70, 70, 20)},
	}

	t.Run("ValidBeforeRejected", func(t *testing.T) {
		// The rejected box lies inside the face and fits the detection better than the shifted valid
		// marker, but covers too little of it to block it, so the valid marker keeps its face.
		valid := entity.Markers{{MarkerUID: "valid", X: 0.22, Y: 0.20, W: 0.2, H: 0.2}}
		rejected := entity.Markers{{MarkerUID: "rejected", MarkerInvalid: true, X: 0.24, Y: 0.24, W: 0.12, H: 0.12}}

		assignments, claimed := matchRegeneratedFaces(valid, rejected, detected)

		assert.Equal(t, map[string]int{"valid": 0}, assignments)
		assert.Equal(t, map[int]bool{0: true}, claimed)
	})
	t.Run("RejectedBlocksValid", func(t *testing.T) {
		// A rejected marker covering the detection keeps it from a valid marker, as an ordinary
		// index would not add it next to the rejected one.
		valid := entity.Markers{{MarkerUID: "valid", X: 0.21, Y: 0.21, W: 0.2, H: 0.2}}
		rejected := entity.Markers{{MarkerUID: "rejected", MarkerInvalid: true, X: 0.20, Y: 0.20, W: 0.2, H: 0.2}}

		assignments, claimed := matchRegeneratedFaces(valid, rejected, detected)

		assert.Equal(t, map[string]int{"rejected": 0}, assignments)
		assert.True(t, claimed[0])
		assert.False(t, claimed[1])
	})
	t.Run("RejectedOnlyTakesWhatItBlocks", func(t *testing.T) {
		// A small rejected box inside an unmarked face does not keep it from being added.
		rejected := entity.Markers{{MarkerUID: "rejected", MarkerInvalid: true, X: 0.64, Y: 0.64, W: 0.1, H: 0.1}}

		assignments, claimed := matchRegeneratedFaces(nil, rejected, detected)

		assert.Empty(t, assignments)
		assert.Empty(t, claimed)
	})
	t.Run("NoMarkers", func(t *testing.T) {
		assignments, claimed := matchRegeneratedFaces(nil, nil, detected)

		assert.Empty(t, assignments)
		assert.Empty(t, claimed)
	})
}

func TestUnregeneratedMarkers(t *testing.T) {
	all, err := query.FaceMarkerFiles("")
	require.NoError(t, err)
	require.Greater(t, len(all), 1)

	total := 0

	for _, n := range all {
		total += n
	}

	t.Run("NoneProcessed", func(t *testing.T) {
		markers, files, countErr := unregeneratedMarkers("", &FaceRegeneration{})

		require.NoError(t, countErr)
		assert.Equal(t, len(all), files)
		assert.Equal(t, total, markers)
	})
	t.Run("PartlyProcessed", func(t *testing.T) {
		stats := &FaceRegeneration{}
		skipped := ""

		for fileUID := range all {
			if skipped == "" {
				skipped = fileUID
				continue
			}

			stats.add(fileUID, faceRegenerationResult{})
		}

		markers, files, countErr := unregeneratedMarkers("", stats)

		require.NoError(t, countErr)
		assert.Equal(t, 1, files)
		assert.Equal(t, all[skipped], markers)
	})
	t.Run("AllProcessed", func(t *testing.T) {
		stats := &FaceRegeneration{}

		for fileUID := range all {
			stats.add(fileUID, faceRegenerationResult{})
		}

		markers, files, countErr := unregeneratedMarkers("", stats)

		require.NoError(t, countErr)
		assert.Zero(t, markers)
		assert.Zero(t, files)
	})
}

func TestFaceRegenerationResult_Changed(t *testing.T) {
	assert.False(t, faceRegenerationResult{}.Changed())
	assert.False(t, faceRegenerationResult{Kept: 1, Failed: 1}.Changed())
	assert.True(t, faceRegenerationResult{Updated: 1}.Changed())
	assert.True(t, faceRegenerationResult{Added: 1}.Changed())
	assert.True(t, faceRegenerationResult{Removed: 1}.Changed())
}

func TestXmpAnchoredMarkers(t *testing.T) {
	markers := entity.Markers{{MarkerUID: "m1", X: 0.3, Y: 0.2, W: 0.1, H: 0.15}}

	t.Run("NoImport", func(t *testing.T) {
		assert.Empty(t, xmpAnchoredMarkers(&MediaFile{}, markers, false))
	})
	t.Run("NoMarkers", func(t *testing.T) {
		assert.Empty(t, xmpAnchoredMarkers(&MediaFile{}, nil, true))
	})
	t.Run("Region", func(t *testing.T) {
		m, err := NewMediaFile("testdata/xmp-faces/sidecar.jpg")
		require.NoError(t, err)

		// The sidecar's region is where m1 is, so the region matches it.
		assert.Equal(t, map[string]bool{"m1": true}, xmpAnchoredMarkers(m, markers, true))
		assert.Empty(t, xmpAnchoredMarkers(m, entity.Markers{{MarkerUID: "m2", X: 0.7, Y: 0.7, W: 0.1, H: 0.1}}, true))
	})
}

func TestIndex_RegenerateFaces(t *testing.T) {
	t.Run("MissingFile", func(t *testing.T) {
		ind := NewIndex(Config(), NewConvert(Config()), NewFiles(), NewPhotos())

		_, err := ind.regenerateFaces(nil, &entity.File{}, false)
		require.Error(t, err)

		_, err = ind.regenerateFaces(&MediaFile{}, nil, false)
		require.Error(t, err)
	})
}

// regenerateTestMarker is what a regeneration may change on a marker, and what it must keep.
type regenerateTestMarker struct {
	X, Y, W, H             float32
	Thumb, DetectModel     string
	Embeddings, Landmarks  string
	Score, Size            int
	Invalid, Review        bool
	SubjUID, SubjSrc, Name string
	UpdatedAt              int64
}

// regenerateTestMarkers returns the face markers of the files as a map keyed by marker uid.
func regenerateTestMarkers(t *testing.T, fileUIDs ...string) map[string]regenerateTestMarker {
	t.Helper()

	result := make(map[string]regenerateTestMarker)

	for _, fileUID := range fileUIDs {
		markers, err := entity.FindMarkers(fileUID)
		require.NoError(t, err)

		for _, m := range markers {
			if m.MarkerType != entity.MarkerFace {
				continue
			}

			result[m.MarkerUID] = regenerateTestMarker{
				X: m.X, Y: m.Y, W: m.W, H: m.H, Thumb: m.Thumb, DetectModel: m.DetectModel,
				Embeddings: string(m.EmbeddingsJSON), Landmarks: string(m.LandmarksJSON), Score: m.Score, Size: m.Size,
				Invalid: m.MarkerInvalid, Review: m.MarkerReview, SubjUID: m.SubjUID, SubjSrc: m.SubjSrc, Name: m.MarkerName,
				UpdatedAt: m.UpdatedAt.UnixNano(),
			}
		}
	}

	return result
}

// regenerateTestXmp returns a sidecar that names face regions, in the Microsoft Photo format.
func regenerateTestXmp(names []string, markers ...entity.Marker) string {
	regions := ""

	for i, m := range markers {
		regions += fmt.Sprintf(`
     <rdf:li rdf:parseType='Resource'>
      <MPReg:PersonDisplayName>%s</MPReg:PersonDisplayName>
      <MPReg:Rectangle>%f, %f, %f, %f</MPReg:Rectangle>
     </rdf:li>`, names[i], m.X, m.Y, m.W, m.H)
	}

	return fmt.Sprintf(`<?xpacket begin='' id='W5M0MpCehiHzreSzNTczkc9d'?>
<x:xmpmeta xmlns:x='adobe:ns:meta/'>
<rdf:RDF xmlns:rdf='http://www.w3.org/1999/02/22-rdf-syntax-ns#'>
 <rdf:Description rdf:about=''
  xmlns:MP='http://ns.microsoft.com/photo/1.2/'
  xmlns:MPRI='http://ns.microsoft.com/photo/1.2/t/RegionInfo#'
  xmlns:MPReg='http://ns.microsoft.com/photo/1.2/t/Region#'>
  <MP:RegionInfo rdf:parseType='Resource'>
   <MPRI:Regions>
    <rdf:Bag>%s
    </rdf:Bag>
   </MPRI:Regions>
  </MP:RegionInfo>
 </rdf:Description>
</rdf:RDF>
</x:xmpmeta>
<?xpacket end='w'?>`, regions)
}

// TestFaces_ResetAndReindex_Regenerate runs "faces reset --detector" through the real faces-only
// index, on markers left in the state an earlier detector would leave them in.
func TestFaces_ResetAndReindex_Regenerate(t *testing.T) {
	useTestDb(t, "faces-regenerate")
	cfg := config.NewMinimalTestConfigWithDb("faces-regenerate", filepath.Join(t.TempDir(), "storage"))
	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
		_ = oldCfg.ConfigureFaceDetector(0)
	})

	require.NoError(t, cfg.ConfigureFaceDetector(0))

	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())

	// index adds a sample picture to the originals and indexes it with face detection.
	index := func(t *testing.T, sample string, importFaceTags bool) (fileUID, photoUID string) {
		t.Helper()

		dst := filepath.Join(cfg.OriginalsPath(), "regenerate", sample)

		if !fs.FileExists(dst) {
			src, err := NewMediaFile(filepath.Join("..", "ai", "face", "testdata", sample))
			require.NoError(t, err)
			require.NoError(t, src.Copy(dst, false))
		}

		m, err := NewMediaFile(dst)
		require.NoError(t, err)

		opt := IndexOptionsSingle(cfg)
		opt.Rescan = true
		opt.DetectFaces = !importFaceTags
		opt.ImportFaceTags = importFaceTags

		result := ind.MediaFile(m, opt, "", "")
		require.True(t, result.Success(), "index must succeed: %v", result.Err)

		return result.FileUID, result.PhotoUID
	}

	namedFileUID, namedPhotoUID := index(t, "17.jpg", false)
	addedFileUID, addedPhotoUID := index(t, "1.jpg", false)
	xmpFileUID, _ := index(t, "16.jpg", false)

	named, err := entity.FindMarkers(namedFileUID)
	require.NoError(t, err)

	added, err := entity.FindMarkers(addedFileUID)
	require.NoError(t, err)

	xmpMarkers, err := entity.FindMarkers(xmpFileUID)
	require.NoError(t, err)

	// Fails rather than skips, so a missing model cannot turn this into a test that asserts nothing.
	require.GreaterOrEqual(t, len(named), 2, "the face detector and embedder must be installed, see make dep")
	require.GreaterOrEqual(t, len(added), 2, "the face detector and embedder must be installed, see make dep")
	require.NotEmpty(t, xmpMarkers, "the face detector and embedder must be installed, see make dep")

	// A sidecar names the first face in 16.jpg with a region wider than the detected face. The
	// earlier detector's box lies in between, so the region names it, while a box moved onto the
	// face would cover too little of the region for the name to stay on it.
	scaled := func(m entity.Marker, f float32) entity.Marker {
		m.X, m.Y, m.W, m.H = m.X-m.W*(f-1)/2, m.Y-m.H*(f-1)/2, m.W*f, m.H*f
		return m
	}

	xmpBox := scaled(xmpMarkers[0], 1.4142)
	require.NoError(t, entity.Db().Model(&entity.Marker{}).Where("marker_uid = ?", xmpMarkers[0].MarkerUID).
		Updates(entity.Values{"detect_model": "centerface", "x": xmpBox.X, "y": xmpBox.Y, "w": xmpBox.W, "h": xmpBox.H}).Error)
	// A second region names a detector marker where there is no face, which no detection matches.
	var xmpFile entity.File
	require.NoError(t, entity.Db().Where("file_uid = ?", xmpFileUID).First(&xmpFile).Error)

	sidecarUnmatched := entity.NewMarker(xmpFile, crop.NewArea("face", 0.01, 0.9, 0.06, 0.06), "", entity.SrcImage, entity.MarkerFace, 40, 90)
	require.NoError(t, sidecarUnmatched.Create())

	require.NoError(t, fs.WriteString(filepath.Join(cfg.OriginalsPath(), "regenerate", "16.jpg.xmp"),
		regenerateTestXmp([]string{"Cara", "Dana"}, scaled(xmpMarkers[0], 1.6), *sidecarUnmatched)))
	index(t, "16.jpg", true)
	cfg.Options().XMPFaces = true

	xmpMarkerCount := len(xmpMarkers) + 1

	sidecarNamed := entity.FindMarker(xmpMarkers[0].MarkerUID)
	require.NotNil(t, sidecarNamed)
	require.Equal(t, "Cara", sidecarNamed.MarkerName, "the region must name the detected marker")
	require.Equal(t, entity.SrcXmp, sidecarNamed.SubjSrc)
	require.Equal(t, "Dana", entity.FindMarker(sidecarUnmatched.MarkerUID).MarkerName, "the region must name the marker")

	var file, addedFile entity.File
	require.NoError(t, entity.Db().Where("file_uid = ?", namedFileUID).First(&file).Error)
	require.NoError(t, entity.Db().Where("file_uid = ?", addedFileUID).First(&addedFile).Error)

	subj := entity.NewSubject("Regenerate Test", entity.SubjPerson, entity.SrcManual)
	require.NoError(t, subj.Create())

	// The markers an earlier detector left: shifted boxes, a foreign detect_model, a person's name
	// on one, and a rejection on the other.
	foreign := entity.Values{"detect_model": "centerface", "x": named[0].X + 0.02, "y": named[0].Y + 0.02}
	foreign["subj_uid"], foreign["subj_src"], foreign["marker_name"] = subj.SubjUID, entity.SrcManual, subj.SubjName
	require.NoError(t, entity.Db().Model(&entity.Marker{}).Where("marker_uid = ?", named[0].MarkerUID).Updates(foreign).Error)
	require.NoError(t, entity.Db().Model(&entity.Marker{}).Where("marker_uid = ?", named[1].MarkerUID).
		Updates(entity.Values{"detect_model": "centerface", "x": named[1].X - 0.02, "marker_invalid": true}).Error)

	// A face the earlier detector missed, and a small rejected box inside a face that is still valid,
	// whose marker the earlier detector placed slightly off.
	require.NoError(t, added[1].Delete())
	require.NoError(t, entity.Db().Model(&entity.Marker{}).Where("marker_uid = ?", added[0].MarkerUID).
		Updates(entity.Values{"detect_model": "centerface", "x": added[0].X + 0.02}).Error)

	inside := added[0].CropArea()
	inside.X, inside.Y, inside.W, inside.H = inside.X+inside.W*0.2, inside.Y+inside.H*0.2, inside.W*0.6, inside.H*0.6
	rejectedInside := entity.NewMarker(addedFile, inside, "", entity.SrcImage, entity.MarkerFace, 40, 90)
	rejectedInside.MarkerInvalid = true
	require.NoError(t, rejectedInside.Create())

	// A small rejected box inside the missed face, which must not keep it from being added.
	missed := added[1].CropArea()
	missed.X, missed.Y, missed.W, missed.H = missed.X+missed.W*0.2, missed.Y+missed.H*0.2, missed.W*0.55, missed.H*0.55
	rejectedSmall := entity.NewMarker(addedFile, missed, "", entity.SrcImage, entity.MarkerFace, 40, 90)
	rejectedSmall.MarkerInvalid = true
	require.NoError(t, rejectedSmall.Create())

	// Markers where there is no face: only the unnamed automatic one may be removed.
	newMarker := func(x, y float32, src string) *entity.Marker {
		m := entity.NewMarker(file, crop.NewArea("face", x, y, 0.05, 0.05), "", src, entity.MarkerFace, 40, 90)
		require.NotNil(t, m)
		return m
	}

	unnamed := newMarker(0.9, 0.9, entity.SrcImage)
	require.NoError(t, unnamed.Create())

	keptNamed := newMarker(0.01, 0.9, entity.SrcImage)
	keptNamed.SubjUID, keptNamed.MarkerName, keptNamed.SubjSrc, keptNamed.DetectModel = subj.SubjUID, subj.SubjName, entity.SrcManual, "centerface"
	keptRejected := newMarker(0.2, 0.9, entity.SrcImage)
	keptRejected.MarkerInvalid = true
	keptManual := newMarker(0.4, 0.9, entity.SrcManual)
	keptSidecar := newMarker(0.6, 0.9, entity.SrcXmp)
	keptSubject := newMarker(0.8, 0.01, entity.SrcImage)
	keptSubject.SubjUID, keptSubject.SubjSrc = subj.SubjUID, entity.SrcManual

	kept := []*entity.Marker{keptNamed, keptRejected, keptManual, keptSidecar, keptSubject, rejectedInside, rejectedSmall, sidecarUnmatched}

	for _, m := range kept[:5] {
		require.NoError(t, m.Create())
	}

	// A cluster a person created, which the default scope alone would keep.
	cluster := entity.NewFace(subj.SubjUID, entity.SrcManual, face.RandomEmbeddings(1, face.RegularFace), face.EmbeddingModelName())
	require.NoError(t, cluster.Create())

	w := NewFaces(cfg)
	before := regenerateTestMarkers(t, namedFileUID, addedFileUID, xmpFileUID)

	t.Run("EmbeddingErrorLeavesMarkers", func(t *testing.T) {
		m, mediaErr := NewMediaFile(filepath.Join(cfg.OriginalsPath(), "regenerate", "17.jpg"))
		require.NoError(t, mediaErr)

		visionConfig := vision.Config
		vision.Config = nil
		t.Cleanup(func() { vision.Config = visionConfig })

		f := file
		_, regenErr := ind.regenerateFaces(m, &f, false)

		require.Error(t, regenErr)
		assert.Equal(t, before, regenerateTestMarkers(t, namedFileUID, addedFileUID, xmpFileUID))
	})
	t.Run("BlockedEmbeddingsAreRefused", func(t *testing.T) {
		t.Cleanup(face.UnblockEmbeddings)
		face.BlockEmbeddings("12 marker(s) use facenet, but this instance is configured for sface")

		_, blockedErr := w.resetAndReindex(string(face.DetectorYuNet), ind, false, "regenerate")

		require.Error(t, blockedErr)
		assert.Equal(t, before, regenerateTestMarkers(t, namedFileUID, addedFileUID, xmpFileUID))
		assert.NotNil(t, entity.FindFace(cluster.ID), "nothing may be removed when the request is refused")
	})

	stats, err := w.resetAndReindex(string(face.DetectorYuNet), ind, false, "regenerate")
	require.NoError(t, err)
	require.NotNil(t, stats)

	after := regenerateTestMarkers(t, namedFileUID, addedFileUID, xmpFileUID)

	t.Run("Summary", func(t *testing.T) {
		assert.Equal(t, int64(3), stats.Files.Load())
		assert.Equal(t, int64(0), stats.FailedFiles.Load())
		assert.GreaterOrEqual(t, stats.Updated.Load(), int64(2))
		assert.Equal(t, int64(1), stats.Added.Load())
		assert.Equal(t, int64(1), stats.Removed.Load())
		assert.Equal(t, int64(len(kept)), stats.Kept.Load())
		assert.Equal(t, int64(0), stats.Failed.Load())
	})
	t.Run("EveryMatchedMarkerIsRegenerated", func(t *testing.T) {
		for _, uid := range []string{named[0].MarkerUID, named[1].MarkerUID, added[0].MarkerUID} {
			assert.Equal(t, string(face.DetectorYuNet), after[uid].DetectModel, uid)
			assert.NotEmpty(t, after[uid].Embeddings, uid)
		}

		assert.Equal(t, "centerface", after[keptNamed.MarkerUID].DetectModel, "an unmatched marker keeps its vector")
	})
	t.Run("NameIsKept", func(t *testing.T) {
		m, ok := after[named[0].MarkerUID]
		require.True(t, ok)

		assert.Equal(t, subj.SubjUID, m.SubjUID)
		assert.Equal(t, entity.SrcManual, m.SubjSrc)
		assert.Equal(t, subj.SubjName, m.Name)
		assert.InDelta(t, named[0].X, m.X, 0.005, "the box must be the new detection's")
		assert.InDelta(t, named[0].Y, m.Y, 0.005, "the box must be the new detection's")
	})
	t.Run("RejectedFaceStaysRejected", func(t *testing.T) {
		m, ok := after[named[1].MarkerUID]
		require.True(t, ok)

		assert.True(t, m.Invalid)
		assert.Equal(t, string(face.DetectorYuNet), m.DetectModel)
		assert.InDelta(t, named[1].X-0.02, m.X, 0.0001, "a rejected marker must not move")

		n := 0

		for _, other := range regenerateTestMarkers(t, namedFileUID) {
			if !other.Invalid && crop.NewArea("face", other.X, other.Y, other.W, other.H).OverlapPercent(named[1].CropArea()) > face.OverlapThresholdFloor {
				n++
			}
		}

		assert.Zero(t, n, "no valid marker may be added for a rejected face")
	})
	t.Run("RejectedBoxInsideValidFace", func(t *testing.T) {
		m, ok := after[added[0].MarkerUID]
		require.True(t, ok, "the valid marker must keep its face")
		assert.False(t, m.Invalid)

		r, ok := after[rejectedInside.MarkerUID]
		require.True(t, ok)
		assert.True(t, r.Invalid)
		assert.Equal(t, before[rejectedInside.MarkerUID].X, r.X)
	})
	t.Run("MissedFaceIsAdded", func(t *testing.T) {
		n := 0

		for uid, m := range regenerateTestMarkers(t, addedFileUID) {
			if uid != added[0].MarkerUID && !m.Invalid {
				n++
			}
		}

		assert.Equal(t, 1, n)

		var photo entity.Photo
		require.NoError(t, entity.Db().Where("photo_uid = ?", addedPhotoUID).First(&photo).Error)
		assert.Equal(t, 2, photo.PhotoFaces)
	})
	t.Run("UnmatchedMarkers", func(t *testing.T) {
		assert.NotContains(t, after, unnamed.MarkerUID, "an unmatched unnamed marker is removed")

		for _, m := range kept {
			assert.Contains(t, after, m.MarkerUID)
		}
	})
	t.Run("FaceCountIsRecomputed", func(t *testing.T) {
		var photo entity.Photo
		require.NoError(t, entity.Db().Where("photo_uid = ?", namedPhotoUID).First(&photo).Error)
		assert.Equal(t, entity.ValidFaceCount(namedFileUID), photo.PhotoFaces)
		assert.Equal(t, 5, photo.PhotoFaces)
	})
	t.Run("SidecarNameIsKept", func(t *testing.T) {
		m, ok := after[xmpMarkers[0].MarkerUID]
		require.True(t, ok)

		assert.Equal(t, "Cara", m.Name, "the sidecar name must stay on the marker it named")
		assert.Equal(t, string(face.DetectorYuNet), m.DetectModel)
		assert.Equal(t, xmpBox.X, m.X, "the marker a sidecar named must not move")
		assert.Len(t, regenerateTestMarkers(t, xmpFileUID), xmpMarkerCount, "no marker may be added for the region")
	})
	t.Run("ClustersAreRemoved", func(t *testing.T) {
		assert.Nil(t, entity.FindFace(cluster.ID))
	})
	t.Run("SmallRejectedBoxKeepsNoFace", func(t *testing.T) {
		r, ok := after[rejectedSmall.MarkerUID]
		require.True(t, ok)
		assert.True(t, r.Invalid)
		assert.Equal(t, before[rejectedSmall.MarkerUID].DetectModel, r.DetectModel, "it must not take the face's detection")
	})
	t.Run("RunTwice", func(t *testing.T) {
		again, againErr := w.resetAndReindex(string(face.DetectorYuNet), ind, false, "regenerate")
		require.NoError(t, againErr)

		assert.Equal(t, fmt.Sprintf("0 updated, 0 added, 0 removed, %d kept unmatched, 0 failed", len(kept)), again.String())
		assert.Equal(t, after, regenerateTestMarkers(t, namedFileUID, addedFileUID, xmpFileUID))
	})
	t.Run("AllKeepsSidecarRegion", func(t *testing.T) {
		// Names are cleared first, so the sidecar region is what keeps the marker it names in place.
		_, allErr := w.resetAndReindex(string(face.DetectorYuNet), ind, true, "regenerate")
		require.NoError(t, allErr)

		markers := regenerateTestMarkers(t, xmpFileUID)
		m, ok := markers[xmpMarkers[0].MarkerUID]
		require.True(t, ok)

		assert.Equal(t, "Cara", m.Name)
		assert.Equal(t, xmpBox.X, m.X, "the marker a sidecar region matches must not move")
		assert.Len(t, markers, xmpMarkerCount, "no marker may be added for the region")

		u, ok := markers[sidecarUnmatched.MarkerUID]
		require.True(t, ok, "an unmatched marker a sidecar region names must be kept")
		assert.Equal(t, "Dana", u.Name)
	})
	t.Run("UnreachedFilesAreReported", func(t *testing.T) {
		// A file the index does not reach, because its original is gone, keeps its marker.
		unreached := entity.File{FileUID: rnd.GenerateUID(entity.FileUID), PhotoID: addedFile.PhotoID, PhotoUID: addedFile.PhotoUID,
			FileName: "regenerate/unreached.jpg", FileRoot: entity.RootOriginals, FileHash: rnd.GenerateUID('h'), FileType: "jpg", FilePrimary: true}
		require.NoError(t, entity.Db().Create(&unreached).Error)

		marker := entity.NewMarker(unreached, crop.NewArea("face", 0.4, 0.4, 0.1, 0.1), "", entity.SrcImage, entity.MarkerFace, 40, 90)
		require.NoError(t, marker.Create())

		t.Cleanup(func() {
			_ = entity.UnscopedDb().Delete(marker).Error
			_ = entity.UnscopedDb().Delete(&unreached).Error
		})

		_, unreachedErr := w.resetAndReindex(string(face.DetectorYuNet), ind, false, "regenerate")

		require.Error(t, unreachedErr)
		assert.True(t, strings.HasPrefix(unreachedErr.Error(), "faces: could not regenerate 1 marker in 1 file the index did not reach"))
	})
}
