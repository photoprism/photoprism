package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// SaveClusterOptionsUpdate persists a cluster options update to options.yml,
// reloads in-memory options, and returns true when values changed.
func (c *Config) SaveClusterOptionsUpdate(update cluster.OptionsUpdate) (bool, error) {
	if c == nil || c.options == nil || update.IsZero() {
		return false, nil
	}

	if err := validateClusterOptionsUpdate(update); err != nil {
		return false, err
	}

	fileName, values, err := c.loadOptionsYAML()

	if err != nil {
		return false, err
	}

	changed := false
	changed = applyOptionString(values, "ClusterUUID", update.ClusterUUID) || changed
	changed = applyOptionString(values, "ClusterCIDR", update.ClusterCIDR) || changed
	changed = applyOptionString(values, "NodeClientID", update.NodeClientID) || changed
	changed = applyOptionString(values, "JWKSUrl", update.JWKSUrl) || changed
	changed = applyOptionString(values, "NodeUUID", update.NodeUUID) || changed
	changed = applyOptionString(values, "DatabaseDriver", update.DatabaseDriver) || changed
	changed = applyOptionString(values, "DatabaseDSN", update.DatabaseDSN) || changed
	changed = applyOptionString(values, "DatabaseServer", update.DatabaseServer) || changed
	changed = applyOptionString(values, "DatabaseName", update.DatabaseName) || changed
	changed = applyOptionString(values, "DatabaseUser", update.DatabaseUser) || changed
	changed = applyOptionString(values, "DatabasePassword", update.DatabasePassword) || changed

	if !changed {
		return false, nil
	}

	b, err := yaml.Marshal(values)

	if err != nil {
		return false, err
	}

	if err = os.WriteFile(fileName, b, fs.ModeConfigFile); err != nil {
		return false, err
	}

	if err = c.options.Load(fileName); err != nil {
		return true, err
	}

	return true, nil
}

// loadOptionsYAML loads options.yml into a writable map and returns its file path.
func (c *Config) loadOptionsYAML() (string, Values, error) {
	fileName := c.OptionsYaml()
	if fileName == "" {
		return "", nil, fmt.Errorf("invalid options.yml filename")
	}

	if err := fs.MkdirAll(filepath.Dir(fileName)); err != nil {
		return fileName, nil, err
	}

	values := Values{}

	if !fs.FileExists(fileName) {
		return fileName, values, nil
	}

	b, err := os.ReadFile(fileName) //nolint:gosec // path derived from config directory
	if err != nil || len(b) == 0 {
		return fileName, values, err
	}

	if err = yaml.Unmarshal(b, &values); err != nil {
		return fileName, nil, fmt.Errorf("failed parsing %s: %w", fileName, err)
	}

	if values == nil {
		values = Values{}
	}

	return fileName, values, nil
}

// applyOptionString sets a string value in the options map and reports changes.
func applyOptionString(values Values, key string, value *string) bool {
	if values == nil || value == nil {
		return false
	}

	if current, ok := values[key]; ok {
		if currentString, ok := current.(string); ok && currentString == *value {
			return false
		}
	}

	values[key] = *value
	return true
}

// validateClusterOptionsUpdate validates cluster-managed option updates.
func validateClusterOptionsUpdate(update cluster.OptionsUpdate) error {
	if update.ClusterUUID != nil && !rnd.IsUUID(*update.ClusterUUID) {
		return fmt.Errorf("invalid cluster UUID")
	}

	if update.NodeUUID != nil && !rnd.IsUUID(*update.NodeUUID) {
		return fmt.Errorf("invalid node UUID")
	}

	return nil
}
