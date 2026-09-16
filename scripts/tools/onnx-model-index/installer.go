/*
Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.
*/
package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// InstallerPath is the checksum-pinned registry that decides what may be installed from the
// mirror. It is read as a description of the artifacts, never as the source of their checksums for
// anything that verifies a download - those stay pinned in the file itself, in git.
const InstallerPath = "scripts/dist/download-models.sh"

// InstallerModel is one row of that registry: name|url|fallback|sha256|type|dir|file.
type InstallerModel struct {
	Name     string
	URL      string
	Fallback string
	SHA256   string
	Type     string
	Dir      string
	File     string
}

var (
	installerVar  = regexp.MustCompile(`^([A-Z0-9_]+)="([^"]*)"`)
	installerRow  = regexp.MustCompile(`^([a-z0-9_.-]+)\|`)
	installerRefs = regexp.MustCompile(`\$\{([A-Z0-9_]+)\}`)
)

// ReadInstaller parses the model rows of the installer registry, keyed by artifact filename.
//
// The format is documented in the script itself, and a row that does not have seven fields is an
// error rather than something to skip: a silently dropped row would leave an artifact looking
// undescribed, which is the one signal this tool exists to produce.
func ReadInstaller(path string) (map[string]InstallerModel, error) {
	body, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	vars := make(map[string]string)
	models := make(map[string]InstallerModel)

	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), `\`))

		if match := installerVar.FindStringSubmatch(line); match != nil {
			vars[match[1]] = expandInstallerVars(match[2], vars)
			continue
		}

		if !installerRow.MatchString(line) {
			continue
		}

		fields := strings.Split(strings.Trim(line, `"`), "|")

		if len(fields) != 7 {
			return nil, fmt.Errorf("%s: row %q has %d fields, want 7", path, fields[0], len(fields))
		}

		model := InstallerModel{
			Name:     fields[0],
			URL:      expandInstallerVars(fields[1], vars),
			Fallback: expandInstallerVars(fields[2], vars),
			SHA256:   fields[3],
			Type:     fields[4],
			Dir:      fields[5],
			File:     fields[6],
		}

		// Only single-file entries name an artifact; a "zip" row installs a directory and has no
		// file of its own to index.
		if model.Type != "file" || model.File == "" {
			continue
		}

		models[model.File] = model
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("%s: no model rows found", path)
	}

	return models, nil
}

// expandInstallerVars substitutes the ${NAME} references the registry uses for its base URLs.
func expandInstallerVars(value string, vars map[string]string) string {
	return installerRefs.ReplaceAllStringFunc(value, func(ref string) string {
		if resolved, found := vars[installerRefs.FindStringSubmatch(ref)[1]]; found {
			return resolved
		}

		return ref
	})
}
