package tensorflow

import (
	"path/filepath"

	tf "github.com/wamuir/graft/tensorflow"

	"github.com/photoprism/photoprism/pkg/clean"
)

// SavedModel loads a saved TensorFlow model from the specified path.
func SavedModel(modelPath string, tags []string) (model *tf.SavedModel, err error) {
	log.Infof("tensorflow: loading %s", clean.Log(filepath.Base(modelPath)))

	if len(tags) == 0 {
		tags = []string{"serve"}
	}

	return tf.LoadSavedModel(modelPath, tags, nil)
}
