package node

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/fs"
)

// ApplyOptionsUpdate persists the provided cluster.OptionsUpdate to options.yml
// and returns true when a write occurred.
func ApplyOptionsUpdate(conf *config.Config, update cluster.OptionsUpdate) (bool, error) {
	if conf == nil || update.IsZero() {
		return false, nil
	}

	fileName := conf.OptionsYaml()
	if err := fs.MkdirAll(filepath.Dir(fileName)); err != nil {
		return false, err
	}

	var existing map[string]any
	if fs.FileExists(fileName) {
		if b, err := os.ReadFile(fileName); err == nil && len(b) > 0 {
			_ = yaml.Unmarshal(b, &existing)
		}
	}
	if existing == nil {
		existing = make(map[string]any)
	}

	update.Visit(func(key string, value any) {
		existing[key] = value
	})

	b, err := yaml.Marshal(existing)
	if err != nil {
		return false, err
	}

	if err := os.WriteFile(fileName, b, fs.ModeFile); err != nil {
		return false, err
	}

	return true, nil
}
