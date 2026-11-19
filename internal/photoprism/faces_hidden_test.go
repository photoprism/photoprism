package photoprism

import (
	"context"
	"errors"
	"image"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestHiddenFacesRecord(t *testing.T) {
	t.Run("RoundTrip", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "sub", "faces-hidden.json")
		record := hiddenFacesRecord{Clusters: []hiddenCluster{{ID: "A", Markers: []string{"m1", "m2"}}}, Target: face.ModelSFace,
			Committed: true, CreatedAt: time.Date(2026, 11, 19, 7, 0, 0, 0, time.UTC)}

		require.NoError(t, writeHiddenFacesRecord(fileName, record))

		read, found, err := readHiddenFacesRecord(fileName)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, record, read)

		// Only the record itself is left, the staged file is published or removed.
		entries, err := os.ReadDir(filepath.Dir(fileName))
		require.NoError(t, err)
		assert.Len(t, entries, 1)

		require.NoError(t, removeHiddenFacesRecord(fileName))
		require.NoError(t, removeHiddenFacesRecord(fileName))
		assert.NoFileExists(t, fileName)
	})
	t.Run("Missing", func(t *testing.T) {
		_, found, err := readHiddenFacesRecord(filepath.Join(t.TempDir(), "faces-hidden.json"))
		assert.NoError(t, err)
		assert.False(t, found)
	})
	t.Run("Invalid", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "faces-hidden.json")
		require.NoError(t, os.WriteFile(fileName, []byte("{"), 0o600))

		_, found, err := readHiddenFacesRecord(fileName)
		assert.True(t, found)
		assert.ErrorContains(t, err, "remove it to continue")
	})
	t.Run("NotARegularFile", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "faces-hidden.json")
		require.NoError(t, os.Mkdir(fileName, 0o700))

		_, found, err := readHiddenFacesRecord(fileName)
		assert.True(t, found)
		assert.ErrorContains(t, err, "not a regular file")
	})
	t.Run("NoFilename", func(t *testing.T) {
		assert.Error(t, writeHiddenFacesRecord("", hiddenFacesRecord{}))
	})
}

func TestRecordHiddenFaces(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "faces-hidden.json")
	clusters := []hiddenCluster{{ID: "A", Markers: []string{"m1"}}}

	require.NoError(t, recordHiddenFaces(fileName, face.ModelSFace, clusters))

	record, found, err := readHiddenFacesRecord(fileName)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, clusters, record.Clusters)
	assert.False(t, record.Committed)

	// Nothing hidden any more removes what an earlier run recorded.
	require.NoError(t, recordHiddenFaces(fileName, face.ModelSFace, nil))
	assert.NoFileExists(t, fileName)
}

func TestCommitHiddenFacesRecord(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "faces-hidden.json")

	require.NoError(t, commitHiddenFacesRecord(fileName), "nothing to commit")
	assert.NoFileExists(t, fileName)

	require.NoError(t, recordHiddenFaces(fileName, face.ModelSFace, []hiddenCluster{{ID: "A", Markers: []string{"m1"}}}))
	require.NoError(t, commitHiddenFacesRecord(fileName))

	record, _, err := readHiddenFacesRecord(fileName)
	require.NoError(t, err)
	assert.True(t, record.Committed)
	assert.Equal(t, []string{"m1"}, record.markers())
}

func TestMergeHiddenClusters(t *testing.T) {
	merged := mergeHiddenClusters(
		[]hiddenCluster{{ID: "A", Markers: []string{"a1"}}, {ID: "B", Markers: []string{"b1"}}},
		[]hiddenCluster{{ID: "A", Markers: []string{"a1", "a2"}}, {ID: "C", Markers: []string{"c1"}}})

	assert.Equal(t, []hiddenCluster{{ID: "A", Markers: []string{"a1", "a2"}}, {ID: "B", Markers: []string{"b1"}},
		{ID: "C", Markers: []string{"c1"}}}, merged)
	assert.Empty(t, mergeHiddenClusters(nil, nil))
}

func TestHiddenFaceClusters(t *testing.T) {
	newMigrateTestConfig(t, "hiddenfaceclusters")
	hidden := newHiddenTestCluster(t, "HIDDENFACECLUSTERS0000000000000000000000", true, 2)
	newHiddenTestCluster(t, "VISIBLEFACECLUSTERS000000000000000000000", false, 1)

	clusters, err := hiddenFaceClusters()
	require.NoError(t, err)
	require.Len(t, clusters, 1)
	assert.Equal(t, "HIDDENFACECLUSTERS0000000000000000000000", clusters[0].ID)
	assert.ElementsMatch(t, hidden, clusters[0].Markers)
}

// newHiddenTestCluster creates a cluster with markers in the isolated migration database; an empty ID
// creates markers that belong to no cluster.
func newHiddenTestCluster(t *testing.T, id string, hidden bool, markers int) (markerUIDs []string) {
	t.Helper()

	if id != "" {
		cluster := &entity.Face{ID: id, FaceSrc: entity.SrcAuto, FaceHidden: hidden, EmbedModel: face.ModelFaceNet,
			EmbeddingJSON: []byte("[0.1,0.2]")}
		require.NoError(t, entity.Db().Create(cluster).Error)
	}

	for range markers {
		m := &entity.Marker{MarkerUID: rnd.GenerateUID('m'), FileUID: rnd.GenerateUID('f'),
			MarkerType: entity.MarkerFace, MarkerSrc: entity.SrcImage, FaceID: id, W: 0.1, H: 0.1}
		require.NoError(t, entity.Db().Create(m).Error)
		markerUIDs = append(markerUIDs, m.MarkerUID)
	}

	return markerUIDs
}

// faceHidden reports whether the cluster with the specified ID is hidden.
func faceHidden(t *testing.T, id string) bool {
	t.Helper()

	var f entity.Face
	require.NoError(t, entity.UnscopedDb().First(&f, "id = ?", id).Error)

	return f.FaceHidden
}

func TestCarriedHiddenClusters(t *testing.T) {
	// unhid still exists and is visible now; detached markers of it were re-embedded by the stopped run.
	unhid := hiddenCluster{ID: "CARRIEDUNHID0000000000000000000000000000",
		Markers: append(newHiddenTestCluster(t, "", false, 2), "ms6sg6b1zzzzzz96")}
	stillHidden := hiddenCluster{ID: "CARRIEDHIDDEN000000000000000000000000000"}
	gone := hiddenCluster{ID: "CARRIEDGONE00000000000000000000000000000", Markers: []string{"ms6sg6b1zzzzzz95"}}

	setup := func(t *testing.T, name string) string {
		t.Helper()

		c := newMigrateTestConfig(t, name)
		newHiddenTestCluster(t, unhid.ID, false, 0)
		stillHidden.Markers = newHiddenTestCluster(t, stillHidden.ID, true, 1)

		return c.FacesHiddenFile()
	}

	t.Run("NotCommitted", func(t *testing.T) {
		// A cluster a person un-hid since is dropped with every marker it had, attached or not.
		fileName := setup(t, "carriedhiddennotcommitted")
		require.NoError(t, writeHiddenFacesRecord(fileName, hiddenFacesRecord{Clusters: []hiddenCluster{unhid, stillHidden, gone}}))

		carried, err := carriedHiddenClusters(fileName)
		require.NoError(t, err)
		assert.Equal(t, []hiddenCluster{stillHidden, gone}, carried)
	})
	t.Run("Committed", func(t *testing.T) {
		// After the finalize every recorded cluster was replaced, so every one still waits.
		fileName := setup(t, "carriedhiddencommitted")
		require.NoError(t, writeHiddenFacesRecord(fileName, hiddenFacesRecord{Clusters: []hiddenCluster{unhid, gone}, Committed: true}))

		carried, err := carriedHiddenClusters(fileName)
		require.NoError(t, err)
		assert.Equal(t, []hiddenCluster{unhid, gone}, carried)
	})
	t.Run("NoRecord", func(t *testing.T) {
		fileName := setup(t, "carriedhiddennone")

		carried, err := carriedHiddenClusters(fileName)
		assert.NoError(t, err)
		assert.Empty(t, carried)
	})
}

func TestFaces_replayHiddenFaces(t *testing.T) {
	loaded := face.NormalizeModelName(face.ConfiguredModel())

	// committedRecord records the markers of a replacement cluster as the hidden cluster it replaced.
	committedRecord := func(t *testing.T, c interface{ FacesHiddenFile() string }, replacement, target string) {
		t.Helper()

		markers := newHiddenTestCluster(t, replacement, false, 2)
		require.NoError(t, writeHiddenFacesRecord(c.FacesHiddenFile(), hiddenFacesRecord{
			Clusters: []hiddenCluster{{ID: "REPLAYHIDDENREPLACED000000000000000000000", Markers: markers}},
			Target:   target, Committed: true, CreatedAt: time.Now().UTC()}))
	}

	t.Run("AppliedOnce", func(t *testing.T) {
		c := newMigrateTestConfig(t, "replayhiddenapplied")
		committedRecord(t, c, "REPLAYHIDDENREPLACEMENT00000000000000000", loaded)

		hidden, err := NewFaces(c).replayHiddenFaces()
		require.NoError(t, err)
		assert.Equal(t, 1, hidden)
		assert.True(t, faceHidden(t, "REPLAYHIDDENREPLACEMENT00000000000000000"))
		assert.NoFileExists(t, c.FacesHiddenFile(), "a record is applied once")
	})
	t.Run("ReplacedMeanwhile", func(t *testing.T) {
		// A migration that writes its own record while the replay runs keeps it.
		c := newMigrateTestConfig(t, "replayhiddenreplaced")
		committedRecord(t, c, "REPLAYHIDDENREPLACEDBYRUN000000000000000", loaded)

		restore := migrationRestoreHiddenFaces
		migrationRestoreHiddenFaces = func(markers []string) (int, error) {
			require.NoError(t, recordHiddenFaces(c.FacesHiddenFile(), loaded, []hiddenCluster{{ID: "NEWRUN", Markers: []string{"m1"}}}))
			return restore(markers)
		}
		t.Cleanup(func() { migrationRestoreHiddenFaces = restore })

		_, err := NewFaces(c).replayHiddenFaces()
		require.NoError(t, err)

		record, found, err := readHiddenFacesRecord(c.FacesHiddenFile())
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, []string{"m1"}, record.markers())
	})
	t.Run("NotCommitted", func(t *testing.T) {
		// A run that did not finalize carries its record forward; a faces pass never applies it.
		c := newMigrateTestConfig(t, "replayhiddennotcommitted")
		markers := newHiddenTestCluster(t, "REPLAYHIDDENNOTCOMMITTED0000000000000000", false, 2)
		require.NoError(t, writeHiddenFacesRecord(c.FacesHiddenFile(), hiddenFacesRecord{
			Clusters: []hiddenCluster{{ID: "REPLAYHIDDENOLD0000000000000000000000000", Markers: markers}}, Target: loaded}))

		hidden, err := NewFaces(c).replayHiddenFaces()
		require.NoError(t, err)
		assert.Zero(t, hidden)
		assert.False(t, faceHidden(t, "REPLAYHIDDENNOTCOMMITTED0000000000000000"))
		assert.FileExists(t, c.FacesHiddenFile())
	})
	t.Run("EmbeddingsBlocked", func(t *testing.T) {
		// A pass that did not cluster has built no replacements to hide.
		c := newMigrateTestConfig(t, "replayhiddenblocked")
		committedRecord(t, c, "REPLAYHIDDENBLOCKED000000000000000000000", loaded)

		face.BlockEmbeddings("test")
		t.Cleanup(face.UnblockEmbeddings)

		hidden, err := NewFaces(c).replayHiddenFaces()
		require.NoError(t, err)
		assert.Zero(t, hidden)
		assert.False(t, faceHidden(t, "REPLAYHIDDENBLOCKED000000000000000000000"))
		assert.FileExists(t, c.FacesHiddenFile())
	})
	t.Run("OtherModelLoaded", func(t *testing.T) {
		c := newMigrateTestConfig(t, "replayhiddenothermodel")
		committedRecord(t, c, "REPLAYHIDDENOTHERMODEL00000000000000000", "othermodel")

		hidden, err := NewFaces(c).replayHiddenFaces()
		require.NoError(t, err)
		assert.Zero(t, hidden)
		assert.False(t, faceHidden(t, "REPLAYHIDDENOTHERMODEL00000000000000000"))
		assert.FileExists(t, c.FacesHiddenFile())
	})
	t.Run("NoRecord", func(t *testing.T) {
		c := newMigrateTestConfig(t, "replayhiddennone")

		hidden, err := NewFaces(c).replayHiddenFaces()
		assert.NoError(t, err)
		assert.Zero(t, hidden)
	})
	t.Run("InvalidRecordIsKept", func(t *testing.T) {
		c := newMigrateTestConfig(t, "replayhiddeninvalid")
		require.NoError(t, os.WriteFile(c.FacesHiddenFile(), []byte("{"), 0o600))

		_, err := NewFaces(c).replayHiddenFaces()
		assert.Error(t, err)
		assert.FileExists(t, c.FacesHiddenFile())
	})
}

// identityChangingEmbedder embeds like oneHotEmbedder and adds a named marker on its first call, as a
// people edit made during the run would, which makes the finalize roll back.
type identityChangingEmbedder struct {
	oneHotEmbedder
	add func()
}

// Run adds the named marker once and returns the next one-hot vector.
func (e *identityChangingEmbedder) Run(img image.Image) face.Embeddings {
	if e.add != nil {
		e.add()
		e.add = nil
	}

	return e.oneHotEmbedder.Run(img)
}

// TestFaces_migrateHiddenClusters pins that the hide decisions survive a run that is interrupted,
// refused, rolled back or stopped before hiding the replacements, and that the run completing the
// migration applies them, while a cluster a person un-hid between the runs stays visible.
func TestFaces_migrateHiddenClusters(t *testing.T) {
	// hiddenLibrary is the library of TestFaces_migrate with three markers in a hidden cluster. With
	// checkpointed > 0, a first run records them, re-embeds that many markers and is interrupted.
	hiddenLibrary := func(t *testing.T, name string, checkpointed int) (w *Faces, hiddenID string, uids []string) {
		t.Helper()

		c := newMigrateTestConfig(t, name)
		addMigrateTestLibrary(t, c, 13, false)

		var markers entity.Markers
		require.NoError(t, entity.UnscopedDb().Order("file_uid").Limit(3).Find(&markers).Error)
		require.Len(t, markers, 3)

		hiddenID = "MIGRATEHIDDEN" + rnd.Base36(27)
		require.NoError(t, entity.Db().Create(&entity.Face{ID: hiddenID, FaceSrc: entity.SrcAuto, FaceHidden: true,
			EmbedModel: face.ModelSFace, EmbeddingJSON: []byte("[0.1,0.2,0.3,0.4]")}).Error)

		for _, m := range markers {
			require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).Where("marker_uid = ?", m.MarkerUID).
				UpdateColumn("face_id", hiddenID).Error)
			uids = append(uids, m.MarkerUID)
		}

		w = NewFaces(c)

		if checkpointed > 0 {
			interruptedMigration(t, w, 13, checkpointed)
			requireHiddenRecord(t, w, false, uids)
		}

		return w, hiddenID, uids
	}

	// run migrates the library to FaceNet and returns the markers the run restored hidden clusters from.
	run := func(t *testing.T, w *Faces, embedder face.Embedder) (restored []string, called bool, err error) {
		t.Helper()

		restore := migrationRestoreHiddenFaces
		migrationRestoreHiddenFaces = func(markers []string) (int, error) {
			restored, called = markers, true
			return restore(markers)
		}
		t.Cleanup(func() { migrationRestoreHiddenFaces = restore })

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		_, err = w.migrate(context.Background(), plan, embedder,
			FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		return restored, called, err
	}

	// completes runs the migration to its end and asserts it restores the hidden markers and removes the record.
	completes := func(t *testing.T, w *Faces, hidden []string) {
		t.Helper()

		restored, called, err := run(t, w, &oneHotEmbedder{dims: 4})
		require.NoError(t, err)
		require.True(t, called)
		assert.ElementsMatch(t, hidden, restored, "the completing run restores the recorded decisions")
		assert.NoFileExists(t, w.conf.FacesHiddenFile())
	}

	t.Run("Interrupted", func(t *testing.T) {
		// Re-embedding detached the checkpointed markers from the hidden cluster; only the record names them.
		w, _, hidden := hiddenLibrary(t, "migratehiddeninterrupted", 5)
		completes(t, w, hidden)
	})
	t.Run("Refused", func(t *testing.T) {
		// A face found and not embedded is a fault of the run, which refuses the finalize.
		w, _, hidden := hiddenLibrary(t, "migratehiddenrefused", 12)

		_, called, err := run(t, w, &missingFaceEmbedder{oneHotEmbedder{dims: 4}})
		var aborted *FacesMigrateAbortedError
		require.ErrorAs(t, err, &aborted)
		assert.False(t, called)
		requireHiddenRecord(t, w, false, hidden)

		completes(t, w, hidden)
	})
	t.Run("RolledBack", func(t *testing.T) {
		w, _, hidden := hiddenLibrary(t, "migratehiddenrolledback", 0)

		var file entity.File
		require.NoError(t, entity.UnscopedDb().Order("file_uid").First(&file).Error)

		_, called, err := run(t, w, &identityChangingEmbedder{oneHotEmbedder: oneHotEmbedder{dims: 4}, add: func() {
			addMigrateTestMarker(t, file.FileUID, entity.SrcManual, "Added During The Run")
		}})
		var rerun *FacesMigrateRerunError
		require.ErrorAs(t, err, &rerun, "the finalize rolls back")
		assert.False(t, called)
		requireHiddenRecord(t, w, false, hidden)

		completes(t, w, hidden)
	})
	t.Run("StoppedBeforeRestoring", func(t *testing.T) {
		w, _, hidden := hiddenLibrary(t, "migratehiddenstopped", 12)

		restore := migrationRestoreHiddenFaces
		migrationRestoreHiddenFaces = func([]string) (int, error) { return 0, errors.New("stopped") }

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		_, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4},
			FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})
		migrationRestoreHiddenFaces = restore
		require.ErrorContains(t, err, "stopped")
		requireHiddenRecord(t, w, true, hidden)

		completes(t, w, hidden)
	})
	t.Run("UnhiddenBetweenRuns", func(t *testing.T) {
		// The interrupted run detached the markers; the person then un-hides the cluster, which still exists.
		w, hiddenID, _ := hiddenLibrary(t, "migratehiddenunhid", 5)
		require.NoError(t, entity.UnscopedDb().Model(&entity.Face{}).Where("id = ?", hiddenID).
			UpdateColumn("face_hidden", false).Error)

		completes(t, w, nil)
	})
	t.Run("CommittedUntilRestored", func(t *testing.T) {
		w, _, _ := hiddenLibrary(t, "migratehiddenrestored", 12)

		restore := migrationRestoreHiddenFaces
		var committed, called bool
		migrationRestoreHiddenFaces = func(markers []string) (int, error) {
			record, _, err := readHiddenFacesRecord(w.conf.FacesHiddenFile())
			assert.NoError(t, err)
			committed, called = record.Committed, true
			return restore(markers)
		}
		t.Cleanup(func() { migrationRestoreHiddenFaces = restore })

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		_, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4},
			FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})
		require.NoError(t, err)

		require.True(t, called)
		assert.True(t, committed, "the record is committed and kept until the clusters are hidden")
		assert.NoFileExists(t, w.conf.FacesHiddenFile())
	})
	t.Run("NothingHidden", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migratehiddennone")
		addMigrateTestLibrary(t, c, 13, false)
		completes(t, NewFaces(c), nil)
	})
}

// requireHiddenRecord asserts the record a run left: its state and the markers it carries.
func requireHiddenRecord(t *testing.T, w *Faces, committed bool, markers []string) {
	t.Helper()

	record, found, err := readHiddenFacesRecord(w.conf.FacesHiddenFile())
	require.NoError(t, err)
	require.True(t, found, "the decisions outlive the run")
	assert.Equal(t, committed, record.Committed)
	assert.ElementsMatch(t, markers, record.markers())
	assert.Equal(t, face.ModelFaceNet, record.Target)
}

// TestFaces_StartReplaysHiddenFaces pins that a faces pass applies a committed record once it clustered.
func TestFaces_StartReplaysHiddenFaces(t *testing.T) {
	loaded := face.NormalizeModelName(face.ConfiguredModel())
	record := hiddenFacesRecord{Clusters: []hiddenCluster{{ID: "STARTREPLAYGONE0000000000000000000000000",
		Markers: []string{"ms6sg6b1zzzzzz97"}}}, Target: loaded, Committed: true, CreatedAt: time.Now().UTC()}

	t.Run("Applied", func(t *testing.T) {
		c := newMigrateTestConfig(t, "startreplayhidden")
		require.NoError(t, writeHiddenFacesRecord(c.FacesHiddenFile(), record))

		require.NoError(t, NewFaces(c).Start(FacesOptions{}))
		assert.NoFileExists(t, c.FacesHiddenFile())
	})
	t.Run("KeptWhileEmbeddingsBlocked", func(t *testing.T) {
		c := newMigrateTestConfig(t, "startreplayblocked")
		require.NoError(t, writeHiddenFacesRecord(c.FacesHiddenFile(), record))

		face.BlockEmbeddings("test")
		t.Cleanup(face.UnblockEmbeddings)

		require.NoError(t, NewFaces(c).Start(FacesOptions{}))
		assert.FileExists(t, c.FacesHiddenFile())
	})
}
