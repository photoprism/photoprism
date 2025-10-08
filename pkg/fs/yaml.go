package fs

import (
	"path/filepath"
)

// YamlFilePath returns the appropriate YAML file name to use. This can be either
// the absolute path of the custom file name passed as the first argument, the default
// name with a ".yml" extension if it already exists, or the default name with a ".yaml"
// extension if a ".yml" file does not exist. This facilitates the transition from ".yml"
// to the new default YAML file extension, ".yaml".
func YamlFilePath(yamlName, yamlDir, customFileName string) string {
	// Return custom file name with absolute path.
	if customFileName != "" {
		return Abs(customFileName)
	}

	// If the file already exists, return the file path with the legacy "*.yml" extension.
	if filePathYml := filepath.Join(yamlDir, yamlName+ExtYml); FileExists(filePathYml) {
		return filePathYml
	}

	// Return file path with ".yaml" extension.
	return filepath.Join(yamlDir, yamlName+ExtYaml)
}
