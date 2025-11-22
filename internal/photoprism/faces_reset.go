package photoprism

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/fs"
)

// runFacesReindex delegates face-only indexing to the supplied Index instance; tests may override it.
var runFacesReindex = func(index *Index, opt IndexOptions) (fs.Done, int, error) {
	if index == nil {
		return nil, 0, fmt.Errorf("faces: index service unavailable")
	}

	found, updated := index.Start(opt)
	return found, updated, nil
}

// Reset removes automatically added face clusters, marker matches, and dangling subjects.
func (w *Faces) Reset() (err error) {
	// Remove automatically added subject and face references from the markers table.
	if removed, err := query.ResetFaceMarkerMatches(); err != nil {
		return fmt.Errorf("faces: %s (reset markers)", err)
	} else {
		log.Infof("faces: removed %d face matches", removed)
	}

	// Remove automatically added face clusters from the index.
	if removed, err := query.RemoveAutoFaceClusters(); err != nil {
		return fmt.Errorf("faces: %s (reset faces)", err)
	} else {
		log.Infof("faces: removed %d face clusters", removed)
	}

	// Remove dangling marker subjects.
	if removed, err := query.RemoveOrphanSubjects(); err != nil {
		return fmt.Errorf("faces: %s (reset subjects)", err)
	} else {
		log.Infof("faces: removed %d dangling subjects", removed)
	}

	return nil
}

// ResetAndReindex resets face data and optionally regenerates markers with the specified engine.
func (w *Faces) ResetAndReindex(engine string, index *Index) error {
	trimmed := strings.TrimSpace(engine)
	lowered := strings.ToLower(trimmed)
	if lowered != "" {
		parsed := face.ParseEngine(lowered)
		if parsed == face.EngineAuto && !strings.EqualFold(trimmed, string(face.EngineAuto)) {
			return fmt.Errorf("faces: unsupported detection engine %q", engine)
		}
	}

	if err := w.Reset(); err != nil {
		return err
	}

	if lowered == "" {
		return nil
	}

	if w.conf == nil {
		return fmt.Errorf("faces: configuration not available")
	}

	engineName := face.ParseEngine(lowered)
	w.conf.Options().FaceEngine = engineName

	if err := face.ConfigureEngine(face.EngineSettings{
		Name: w.conf.FaceEngine(),
		ONNX: face.ONNXOptions{
			ModelPath: w.conf.FaceEngineModelPath(),
			Threads:   w.conf.FaceEngineThreads(),
		},
	}); err != nil {
		return err
	}

	convert := w.conf.Settings().Index.Convert && w.conf.SidecarWritable()
	opt := IndexOptionsFacesOnly(w.conf)
	opt.Convert = convert

	found, updated, err := runFacesReindex(index, opt)
	if err != nil {
		return err
	}

	log.Infof("faces: regenerated %s using %s engine (%s scanned)", english.Plural(updated, "file", "files"), w.conf.FaceEngine(), english.Plural(len(found), "file", "files"))

	return nil
}
