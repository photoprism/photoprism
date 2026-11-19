package photoprism

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// hiddenFacesRecordLimit bounds the record that is read back, which lists one marker UID per entry.
const hiddenFacesRecordLimit = 64 << 20

// migrationRestoreHiddenFaces hides the replacement clusters of the recorded markers; tests replace it to
// stop a migration where a crash would, after the finalize and before the clusters are hidden again.
var migrationRestoreHiddenFaces = query.RestoreHiddenFaces

// hiddenCluster is a cluster a person hid, with the markers that carry the decision to its replacement.
type hiddenCluster struct {
	ID      string   `json:"id"`
	Markers []string `json:"markers"`
}

// hiddenFacesRecord lists the clusters a person hid, so the decisions survive a migration that stops
// before it has hidden their replacements: re-embedding detaches the markers from the clusters.
type hiddenFacesRecord struct {
	Clusters []hiddenCluster `json:"clusters"`
	Target   string          `json:"target"`
	// Committed is set once the finalize has replaced the clusters, which is when a faces pass may apply it.
	Committed bool      `json:"committed"`
	CreatedAt time.Time `json:"createdAt"`
}

// markers returns the UIDs of every recorded marker.
func (r hiddenFacesRecord) markers() (result []string) {
	for _, c := range r.Clusters {
		result = append(result, c.Markers...)
	}

	return result
}

// writeHiddenFacesRecord publishes the record through a staged file, so a reader never sees it partly
// written, and syncs the file and its folder, since it has to outlive a crash right after the finalize.
func writeHiddenFacesRecord(fileName string, record hiddenFacesRecord) (err error) {
	if fileName == "" {
		return errors.New("missing filename")
	}

	b, err := json.Marshal(record)

	if err != nil {
		return err
	}

	dir := filepath.Dir(fileName)

	if err = fs.MkdirAll(dir); err != nil {
		return err
	}

	f, err := fs.OpenStageFile(fileName)

	if err != nil {
		return err
	}

	staged := f.Name()

	defer func() {
		if removeErr := os.Remove(staged); removeErr != nil && !os.IsNotExist(removeErr) && err == nil {
			err = removeErr
		}
	}()

	if _, err = f.Write(b); err != nil {
		_ = f.Close()
		return err
	} else if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	} else if err = f.Close(); err != nil {
		return err
	} else if err = fs.PublishFile(staged, fileName, true); err != nil {
		return err
	}

	// Best effort: not every filesystem can sync a folder, and the file itself is synced already.
	if d, dirErr := os.Open(dir); dirErr == nil { //nolint:gosec // dir is derived from the storage path
		_ = d.Sync()
		_ = d.Close()
	}

	return nil
}

// readHiddenFacesRecord returns the record a migration left, and false when there is none. Anything but
// a regular file of bounded size is refused.
func readHiddenFacesRecord(fileName string) (record hiddenFacesRecord, found bool, err error) {
	info, err := os.Lstat(fileName)

	switch {
	case os.IsNotExist(err):
		return record, false, nil
	case err != nil:
		return record, false, err
	case !info.Mode().IsRegular():
		return record, true, fmt.Errorf("%s is not a regular file, remove it to continue without the hidden clusters it records",
			clean.Log(filepath.Base(fileName)))
	}

	f, err := os.Open(fileName) //nolint:gosec // path is derived from the storage path

	if err != nil {
		return record, true, err
	}

	defer f.Close()

	b, err := io.ReadAll(io.LimitReader(f, hiddenFacesRecordLimit+1))

	if err != nil {
		return record, true, err
	} else if len(b) > hiddenFacesRecordLimit {
		return record, true, fmt.Errorf("%s is too large, remove it to continue without the hidden clusters it records",
			clean.Log(filepath.Base(fileName)))
	} else if err = json.Unmarshal(b, &record); err != nil {
		return record, true, fmt.Errorf("%s is invalid (%s), remove it to continue without the hidden clusters it records",
			clean.Log(filepath.Base(fileName)), err)
	}

	return record, true, nil
}

// removeHiddenFacesRecord removes the record once the decisions it holds have been applied.
func removeHiddenFacesRecord(fileName string) error {
	if err := os.Remove(fileName); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// hiddenFaceClusters returns the hidden clusters and their markers as they are now.
func hiddenFaceClusters() (result []hiddenCluster, err error) {
	markers, err := query.HiddenFaceMarkerClusters()

	if err != nil {
		return nil, err
	}

	for _, m := range markers {
		if n := len(result); n > 0 && result[n-1].ID == m.FaceID {
			result[n-1].Markers = append(result[n-1].Markers, m.MarkerUID)
		} else {
			result = append(result, hiddenCluster{ID: m.FaceID, Markers: []string{m.MarkerUID}})
		}
	}

	return result, nil
}

// carriedHiddenClusters returns the clusters an earlier migration recorded and has not applied yet. Of a
// record whose finalize did not commit, it drops the clusters that still exist and are visible, since a
// person un-hid them since; every cluster of a committed one was replaced and still waits.
func carriedHiddenClusters(fileName string) (clusters []hiddenCluster, err error) {
	record, found, err := readHiddenFacesRecord(fileName)

	if err != nil || !found || len(record.Clusters) == 0 {
		return nil, err
	} else if record.Committed {
		return record.Clusters, nil
	}

	ids := make([]string, 0, len(record.Clusters))

	for _, c := range record.Clusters {
		ids = append(ids, c.ID)
	}

	visible, err := query.VisibleFaceIDs(ids)

	if err != nil {
		return nil, err
	}

	shown := make(map[string]struct{}, len(visible))

	for _, id := range visible {
		shown[id] = struct{}{}
	}

	for _, c := range record.Clusters {
		if _, ok := shown[c.ID]; !ok {
			clusters = append(clusters, c)
		}
	}

	return clusters, nil
}

// mergeHiddenClusters returns the clusters of both lists, with the markers of a cluster in both combined.
func mergeHiddenClusters(clusters, carried []hiddenCluster) []hiddenCluster {
	result := make([]hiddenCluster, 0, len(clusters)+len(carried))
	index := make(map[string]int, len(clusters)+len(carried))
	seen := make(map[string]struct{})

	for _, list := range [][]hiddenCluster{clusters, carried} {
		for _, c := range list {
			i, ok := index[c.ID]

			if !ok {
				index[c.ID] = len(result)
				result = append(result, hiddenCluster{ID: c.ID})
				i = len(result) - 1
			}

			for _, uid := range c.Markers {
				if key := c.ID + "/" + uid; !hasKey(seen, key) {
					seen[key] = struct{}{}
					result[i].Markers = append(result[i].Markers, uid)
				}
			}
		}
	}

	return result
}

// hasKey reports whether the set contains the key.
func hasKey(set map[string]struct{}, key string) bool {
	_, ok := set[key]
	return ok
}

// recordHiddenFaces records the hidden clusters before a migration changes any marker, or removes the
// record when there are none.
func recordHiddenFaces(fileName, target string, clusters []hiddenCluster) error {
	if len(clusters) == 0 {
		return removeHiddenFacesRecord(fileName)
	}

	if err := writeHiddenFacesRecord(fileName, hiddenFacesRecord{Clusters: clusters, Target: target,
		CreatedAt: time.Now().UTC()}); err != nil {
		return fmt.Errorf("faces: %s (record hidden clusters)", err)
	}

	return nil
}

// commitHiddenFacesRecord marks the record as belonging to a finalize that has replaced the clusters.
func commitHiddenFacesRecord(fileName string) error {
	record, found, err := readHiddenFacesRecord(fileName)

	if err != nil || !found {
		return err
	}

	record.Committed = true

	return writeHiddenFacesRecord(fileName, record)
}

// replayHiddenFaces hides the replacement clusters of a committed migration that stopped before hiding
// them, once a faces pass has clustered with the target model loaded, and reports how many it hid. It is
// applied once, like the migration's own restore, and removed unless a migration has replaced it since.
func (w *Faces) replayHiddenFaces() (hidden int, err error) {
	if w == nil || w.conf == nil {
		return 0, nil
	}

	fileName := w.conf.FacesHiddenFile()
	record, found, err := readHiddenFacesRecord(fileName)

	if err != nil || !found || !record.Committed {
		return 0, err
	}

	// Until the target is loaded and embeddings run, the pass did not cluster the recorded markers.
	if face.EmbeddingsBlockedReason() != "" || face.NormalizeModelName(face.ConfiguredModel()) != face.NormalizeModelName(record.Target) {
		return 0, nil
	}

	if hidden, err = migrationRestoreHiddenFaces(record.markers()); err != nil {
		return hidden, err
	}

	log.Infof("faces: hid %d cluster(s) again that were hidden before the migration to %s", hidden, clean.Log(record.Target))

	if current, ok, readErr := readHiddenFacesRecord(fileName); readErr != nil || !ok || !current.CreatedAt.Equal(record.CreatedAt) {
		return hidden, readErr
	}

	return hidden, removeHiddenFacesRecord(fileName)
}
