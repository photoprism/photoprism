/*
Command onnx-model-index generates the machine-readable catalogue published beside the mirrored
ONNX models, so a consumer can learn what each model is and where it came from without downloading
artifacts that total more than a gigabyte.

The directory is the subject and the registries are only a description of it. Building the index
from the registries instead would re-advertise models we deliberately do not mirror, since they
describe non-free weights too.

Nothing at runtime depends on this command, and the index it writes is a catalogue rather than a
trust root: checksums are pinned in scripts/dist/download-models.sh, in git, because a checksum
served from the host that serves the artifact is attested by whoever controls that host.

Reading metadata_props goes through the ONNX Runtime binding, so run this inside the development
container. See specs/intelligence/onnx-model-index.md.

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.
*/
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/onnx"
)

// SchemaVersion is incremented only for a change a consumer could misread. Additive fields do not
// bump it, so a reader that ignores unknown keys keeps working across those.
const SchemaVersion = 1

// NoticeFile names the attribution record published in the same directory.
const NoticeFile = "NOTICE"

// Index is the published document.
type Index struct {
	SchemaVersion int      `json:"schemaVersion"`
	Updated       string   `json:"updated"`
	Notice        string   `json:"notice"`
	Models        []*Entry `json:"models"`
}

// Entry describes one mirrored artifact. It embeds onnx.ModelInfo so the shared fields are the
// ones the runtime itself parses and cannot drift from it; the fields declared here are the
// catalogue's own, which a reader needs and the runtime does not.
type Entry struct {
	Name           string `json:"name"`
	Task           string `json:"task,omitempty"`
	Origin         string `json:"origin"`
	Size           int64  `json:"size"`
	Publisher      string `json:"publisher,omitempty"`
	Source         string `json:"source,omitempty"`
	SourceRevision string `json:"sourceRevision,omitempty"`
	SourceSHA256   string `json:"sourceSHA256,omitempty"`
	Exported       string `json:"exported,omitempty"`
	onnx.ModelInfo
}

// Origins distinguish an artifact we produced from one redistributed unmodified. Only the first
// carries photoprism.* metadata, because embedding properties in the second would end the
// byte-identity the NOTICE asserts.
const (
	OriginExport    = "export"
	OriginPublisher = "publisher"
)

// Tasks name what a model is for.
const (
	TaskLabel         = "label"
	TaskNSFW          = "nsfw"
	TaskFaceDetection = "face-detection"
	TaskFaceEmbedding = "face-embedding"
)

// registryEntry is a model description taken from the code rather than from the artifact.
type registryEntry struct {
	Info *onnx.ModelInfo
	Task string
	Name string
}

func main() {
	dir := flag.String("dir", "", "directory holding the mirrored .onnx artifacts")
	output := flag.String("output", "", "write the index to this file instead of stdout")
	installer := flag.String("installer", InstallerPath, "path of the checksum-pinned installer registry")
	library := flag.String("library", "", "path of the ONNX Runtime shared library, when it is not in a default location")
	flag.Parse()

	if *dir == "" {
		fmt.Fprintln(os.Stderr, "onnx-model-index: -dir is required")
		os.Exit(2)
	}

	// Reading metadata_props goes through the binding, which refuses every model until the shared
	// library is loaded and the environment initialized.
	if err := onnx.EnsureRuntime(*library); err != nil {
		fmt.Fprintf(os.Stderr, "onnx-model-index: %s\n", err)
		os.Exit(1)
	}

	index, err := Generate(*dir, *installer)

	if err != nil {
		fmt.Fprintf(os.Stderr, "onnx-model-index: %s\n", err)
		os.Exit(1)
	}

	encoded, err := json.MarshalIndent(index, "", "  ")

	if err != nil {
		fmt.Fprintf(os.Stderr, "onnx-model-index: %s\n", err)
		os.Exit(1)
	}

	encoded = append(encoded, '\n')

	if *output == "" {
		if _, err = os.Stdout.Write(encoded); err != nil {
			fmt.Fprintf(os.Stderr, "onnx-model-index: %s\n", err)
			os.Exit(1)
		}

		return
	}

	if err = os.WriteFile(*output, encoded, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "onnx-model-index: %s\n", err)
		os.Exit(1)
	}
}

// Generate builds the index from the artifacts in dir. It returns an error naming every file it
// could not describe, because an undescribed artifact is a finding rather than a gap to paper over:
// either nobody registered it, or a registry lost an entry.
func Generate(dir, installerPath string) (*Index, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.onnx"))

	if err != nil {
		return nil, err
	}

	sort.Strings(files)

	if len(files) == 0 {
		return nil, fmt.Errorf("no .onnx artifacts in %s", dir)
	}

	installer, err := ReadInstaller(installerPath)

	if err != nil {
		return nil, err
	}

	registry := Registry()
	index := &Index{
		SchemaVersion: SchemaVersion,
		Updated:       time.Now().UTC().Format(time.RFC3339),
		Notice:        NoticeFile,
	}

	var problems []string

	for _, file := range files {
		entry, entryErr := Describe(file, registry, installer)

		if entryErr != nil {
			problems = append(problems, entryErr.Error())
			continue
		}

		index.Models = append(index.Models, entry)
	}

	if len(problems) > 0 {
		return nil, fmt.Errorf("%d artifact(s) not indexed:\n  %s", len(problems), strings.Join(problems, "\n  "))
	}

	return index, nil
}

// Registry collects the model descriptions the code already carries, keyed by artifact filename.
// Only registries whose models may be redistributed are included; a non-free entry is skipped here
// so that it can be reported as a mirroring error if its file is present.
func Registry() map[string]registryEntry {
	known := make(map[string]registryEntry)

	for name, model := range face.EmbeddingModels {
		if model.ONNX == nil || model.ONNX.File == "" {
			continue
		}

		known[model.ONNX.File] = registryEntry{Info: model.ONNX, Task: TaskFaceEmbedding, Name: name}
	}

	for _, detector := range face.Detectors {
		if detector.ONNX == nil || detector.ONNX.File == "" {
			continue
		}

		known[detector.ONNX.File] = registryEntry{Info: detector.ONNX, Task: TaskFaceDetection, Name: detector.Name}
	}

	return known
}

// Describe builds one entry from the artifact's own metadata, from a registry, or from both.
// When both describe it they must agree, because preferring one side would hide the disagreement.
func Describe(path string, registry map[string]registryEntry, installer map[string]InstallerModel) (*Entry, error) {
	file := filepath.Base(path)

	stat, err := os.Stat(path)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}

	digest, err := FileSHA256(path)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}

	metadata, err := onnx.Metadata(path)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}

	known, inRegistry := registry[file]
	installed, inInstaller := installer[file]

	if len(metadata) == 0 && !inRegistry && !inInstaller {
		return nil, fmt.Errorf("%s: no embedded metadata, no registry entry and no installer row", file)
	}

	if inInstaller && installed.SHA256 != digest {
		return nil, fmt.Errorf("%s: installer pins %s but the file is %s", file, installed.SHA256, digest)
	}

	if inRegistry {
		if known.Info.SHA256 != "" && known.Info.SHA256 != digest {
			return nil, fmt.Errorf("%s: registry pins %s but the file is %s", file, known.Info.SHA256, digest)
		}
	}

	entry := &Entry{
		Size:           stat.Size(),
		SourceRevision: metadata["sourceRevision"],
		SourceSHA256:   metadata["sourceSHA256"],
		Exported:       metadata["exported"],
	}

	entry.File = file
	entry.SHA256 = digest
	entry.Source = metadata["source"]
	entry.License = metadata["license"]
	entry.Quantization = metadata["quantization"]

	if len(metadata) > 0 {
		entry.Origin = OriginExport
		entry.Name = strings.TrimSuffix(file, ".onnx")
	} else {
		entry.Origin = OriginPublisher
	}

	if inRegistry {
		entry.Name = known.Name
		entry.Task = known.Task

		if entry.License == "" {
			entry.License = known.Info.License
		} else if known.Info.License != "" && entry.License != known.Info.License {
			return nil, fmt.Errorf("%s: embedded license %q disagrees with registry %q",
				file, entry.License, known.Info.License)
		}

		entry.Input = known.Info.Input
		entry.Output = known.Info.Output
	}

	if inInstaller {
		// The installer name is what an operator types, so it is preferred over a filename stem.
		entry.Name = installed.Name

		// The fallback is the publisher's own copy, which is where a publisher-hosted artifact
		// records its provenance; our own exports have no upstream equivalent and leave it empty.
		if entry.Source == "" {
			entry.Source = installed.Fallback
		}
	}

	if entry.Task == "" {
		entry.Task = TaskFromMetadata(metadata)
	}

	// The SCRFD and ArcFace weights are licensed for non-commercial research only. Their presence
	// on the mirror is a redistribution fault rather than a catalogue entry, so this is checked on
	// the resolved license and holds whichever source described the artifact.
	if entry.License == face.LicenseNonFree {
		return nil, fmt.Errorf("%s: licensed %s and MUST NOT be mirrored or indexed", file, face.LicenseNonFree)
	}

	if entry.License == "" || entry.License == face.LicenseUnknown {
		return nil, fmt.Errorf("%s: no license recorded in the artifact, a registry or the installer", file)
	}

	entry.Publisher = PublisherFromSource(entry.Source)

	return entry, nil
}

// TaskFromMetadata derives the task of a model PhotoPrism exported. An NSFW export declares how its
// output reduces to one probability; a classifier declares the ImageNet-1k output width instead.
func TaskFromMetadata(metadata map[string]string) string {
	if metadata["reduction"] != "" {
		return TaskNSFW
	}

	if metadata["outputWidth"] == "1000" {
		return TaskLabel
	}

	return ""
}

// PublisherFromSource returns the account a checkpoint was published under, which is the first path
// segment on the hosts we take checkpoints from. It returns an empty string for anything else,
// because a guessed publisher in a licensing record is worse than none.
func PublisherFromSource(source string) string {
	if source == "" {
		return ""
	}

	parsed, err := url.Parse(source)

	if err != nil {
		return ""
	}

	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")

	switch parsed.Host {
	case "huggingface.co", "raw.githubusercontent.com", "github.com":
		if len(segments) > 0 {
			return segments[0]
		}
	case "media.githubusercontent.com":
		// Large-file URLs are served under a /media prefix, so the account is one segment deeper.
		if len(segments) > 1 && segments[0] == "media" {
			return segments[1]
		}
	}

	return ""
}

// FileSHA256 returns the hex-encoded SHA-256 of a file, streaming it so a multi-hundred-megabyte
// artifact does not have to be held in memory.
func FileSHA256(path string) (string, error) {
	f, err := os.Open(path)

	if err != nil {
		return "", err
	}

	defer f.Close()

	digest := sha256.New()

	if _, err = io.Copy(digest, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", digest.Sum(nil)), nil
}
