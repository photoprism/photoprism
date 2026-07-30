package face

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// Environment variables that configure the face embedding model benchmark.
const (
	// FaceDatasetEnv points to a directory of identity subdirectories, each holding
	// the images of one person. Subdirectories may be grouped one level deeper to
	// benchmark subsets such as "asian" or "children" separately.
	FaceDatasetEnv = "PHOTOPRISM_TEST_FACE_DATASET"
	// FaceModelsEnv limits the benchmark to a comma-separated list of model names.
	FaceModelsEnv = "PHOTOPRISM_TEST_FACE_MODELS"
	// FaceMaxImagesEnv caps how many images per identity are used.
	FaceMaxImagesEnv = "PHOTOPRISM_TEST_FACE_MAX_IMAGES"
	// FaceMaxNegativesEnv caps how many cross-identity pairs are scored.
	FaceMaxNegativesEnv = "PHOTOPRISM_TEST_FACE_MAX_NEGATIVES"
	// FaceThreadsEnv sets the inference thread count.
	FaceThreadsEnv = "PHOTOPRISM_TEST_FACE_THREADS"
)

// benchmarkSubject holds the embeddings computed for one identity, one entry per image.
type benchmarkSubject struct {
	name       string
	embeddings []Embedding
}

// benchmarkSubset is a named group of identities that is evaluated on its own.
type benchmarkSubset struct {
	name string
}

// benchmarkImage is a dataset image together with the identity it belongs to.
type benchmarkImage struct {
	subset  string
	subject string
	file    string
}

// benchmarkStats records how a model performed on the whole dataset.
type benchmarkStats struct {
	model      ModelName
	dims       int
	sizeBytes  int64
	perFace    time.Duration
	embedded   int
	failed     int
	bySubset   map[string]evalResult
	subsetKeys []string
}

// envInt returns an integer environment variable, or the fallback when unset or invalid.
func envInt(name string, fallback int) int {
	if v, err := strconv.Atoi(strings.TrimSpace(os.Getenv(name))); err == nil && v > 0 {
		return v
	}

	return fallback
}

// benchmarkModelNames returns the models to benchmark, honoring FaceModelsEnv and
// skipping any whose weights are not installed.
func benchmarkModelNames(t *testing.T, modelsPath string) []ModelName {
	t.Helper()

	requested := EmbeddingModelNames()

	if list := strings.TrimSpace(os.Getenv(FaceModelsEnv)); list != "" {
		requested = nil

		for _, name := range strings.Split(list, ",") {
			if name = NormalizeModelName(name); name != "" {
				requested = append(requested, name)
			}
		}
	}

	var result []ModelName

	for _, name := range requested {
		m := FindEmbeddingModel(name)

		if m == nil {
			t.Logf("benchmark: %s is not a known model, skipping", name)
			continue
		}

		if !m.Installed(modelsPath) {
			t.Logf("benchmark: %s is not installed, skipping", name)
			continue
		}

		result = append(result, name)
	}

	return result
}

// collectDatasetImages walks a dataset directory and returns its images grouped by
// subset and identity. A directory holding images is treated as one identity, so
// both <dataset>/<person> and <dataset>/<subset>/<person> layouts work.
func collectDatasetImages(root string, maxPerSubject int) ([]benchmarkImage, error) {
	var result []benchmarkImage

	var walk func(dir, subset string, depth int) error

	walk = func(dir, subset string, depth int) error {
		entries, err := os.ReadDir(dir)

		if err != nil {
			return err
		}

		var images []string
		var subdirs []string

		for _, entry := range entries {
			name := entry.Name()

			if strings.HasPrefix(name, ".") {
				continue
			}

			if entry.IsDir() {
				subdirs = append(subdirs, name)
				continue
			}

			switch fs.Extensions[strings.ToLower(filepath.Ext(name))] {
			case fs.ImageJpeg, fs.ImagePng:
				images = append(images, name)
			}
		}

		sort.Strings(images)
		sort.Strings(subdirs)

		// A directory holding images is an identity, and its name is the subject.
		if len(images) > 0 {
			if maxPerSubject > 0 && len(images) > maxPerSubject {
				images = images[:maxPerSubject]
			}

			subject := filepath.Base(dir)

			for _, name := range images {
				result = append(result, benchmarkImage{subset: subset, subject: subject, file: filepath.Join(dir, name)})
			}

			return nil
		}

		// A directory below the root that holds no images of its own groups identities,
		// which is how optional subsets such as "asian" or "children" are recognized.
		if depth == 1 && subset == "" {
			subset = filepath.Base(dir)
		}

		for _, name := range subdirs {
			if err := walk(filepath.Join(dir, name), subset, depth+1); err != nil {
				return err
			}
		}

		return nil
	}

	if err := walk(root, "", 0); err != nil {
		return nil, err
	}

	return result, nil
}

// benchmarkCrop prepares the inference input for a face, matching what the indexer
// does: landmark-aligned for models that expect it, and a square bounding box crop
// scaled to the model resolution otherwise.
func benchmarkCrop(embedder Embedder, img image.Image, f *Face) (image.Image, error) {
	width, height := embedder.CropSize()

	if embedder.Aligned() {
		return AlignedCrop(img, f, width, height)
	}

	size := f.Area.Scale

	if size < 1 {
		return nil, fmt.Errorf("face has no size")
	}

	bounds := img.Bounds()
	box := image.NewRGBA(image.Rect(0, 0, size, size))
	originX := bounds.Min.X + f.Area.Col - size/2
	originY := bounds.Min.Y + f.Area.Row - size/2

	for y := range size {
		for x := range size {
			box.Set(x, y, img.At(originX+x, originY+y))
		}
	}

	return box, nil
}

// buildPairs scores every within-identity pair and a deterministic sample of
// cross-identity pairs, so results are reproducible across runs.
func buildPairs(subjects []benchmarkSubject, maxNegatives int) []scoredPair {
	var pairs []scoredPair

	for _, s := range subjects {
		for i := 0; i < len(s.embeddings); i++ {
			for j := i + 1; j < len(s.embeddings); j++ {
				if d := s.embeddings[i].Dist(s.embeddings[j]); d >= 0 {
					pairs = append(pairs, scoredPair{dist: d, same: true})
				}
			}
		}
	}

	// Count the candidate negatives first so they can be sampled with an even stride
	// instead of a random generator, which keeps the selection reproducible.
	var candidates int

	for i := range subjects {
		for j := i + 1; j < len(subjects); j++ {
			candidates += len(subjects[i].embeddings) * len(subjects[j].embeddings)
		}
	}

	stride := 1

	if maxNegatives > 0 && candidates > maxNegatives {
		stride = candidates / maxNegatives
	}

	counter := 0

	for i := range subjects {
		for j := i + 1; j < len(subjects); j++ {
			for _, a := range subjects[i].embeddings {
				for _, b := range subjects[j].embeddings {
					if counter%stride == 0 {
						if d := a.Dist(b); d >= 0 {
							pairs = append(pairs, scoredPair{dist: d, same: false})
						}
					}

					counter++
				}
			}
		}
	}

	return pairs
}

// TestBenchmarkEmbeddingModels compares the installed face embedding models on a
// labeled dataset. It is skipped unless PHOTOPRISM_TEST_FACE_DATASET points to a
// directory of identity subdirectories.
func TestBenchmarkEmbeddingModels(t *testing.T) {
	datasetPath := strings.TrimSpace(os.Getenv(FaceDatasetEnv))

	if datasetPath == "" {
		t.Skipf("benchmark: set %s to a dataset directory to run this comparison", FaceDatasetEnv)
	}

	maxPerSubject := envInt(FaceMaxImagesEnv, 10)
	maxNegatives := envInt(FaceMaxNegativesEnv, 200000)
	threads := envInt(FaceThreadsEnv, 2)

	images, err := collectDatasetImages(datasetPath, maxPerSubject)
	require.NoError(t, err)
	require.NotEmpty(t, images, "no images found in %s", datasetPath)

	modelNames := benchmarkModelNames(t, embeddingModelsPath)
	require.NotEmpty(t, modelNames, "no embedding models installed")

	useTestDetector(t)

	// Every model sees the same detections, so differences come from the embedding
	// model rather than from detector variance between runs.
	embedders := make(map[ModelName]Embedder, len(modelNames))
	stats := make(map[ModelName]*benchmarkStats, len(modelNames))

	for _, name := range modelNames {
		m := FindEmbeddingModel(name)

		var embedder Embedder
		var initErr error

		if m.Runtime == RuntimeONNX {
			embedder, initErr = NewONNXEmbedder(EmbedderSettings{
				Name:      name,
				Model:     m,
				ModelPath: m.FilePath(embeddingModelsPath),
				Threads:   threads,
			})
		} else {
			tfModel := NewModel(name, m.FilePath(embeddingModelsPath), t.TempDir(), m.Width, nil, false)
			initErr = tfModel.Init()
			embedder = tfModel
		}

		if initErr != nil {
			t.Logf("benchmark: failed to load %s (%s), skipping", name, initErr)
			continue
		}

		embedders[name] = embedder
		stats[name] = &benchmarkStats{
			model:     name,
			dims:      embedder.Dims(),
			sizeBytes: modelSizeBytes(m.FilePath(embeddingModelsPath)),
			bySubset:  make(map[string]evalResult),
		}

		t.Cleanup(func() { _ = embedder.Close() })
	}

	require.NotEmpty(t, embedders, "no embedding models could be loaded")

	// subset -> subject -> model -> embeddings
	collected := make(map[string]map[string]map[ModelName][]Embedding)
	elapsed := make(map[ModelName]time.Duration, len(embedders))
	detectFailed := 0

	for _, item := range images {
		img, _, decodeErr := fs.DecodeImageFile(item.file)

		if decodeErr != nil {
			t.Logf("benchmark: failed to decode %s (%s)", filepath.Base(item.file), decodeErr)
			detectFailed++
			continue
		}

		faces, detectErr := Detect(item.file, SizeThreshold)

		if detectErr != nil || len(faces) == 0 {
			detectFailed++
			continue
		}

		// Use the largest detection, which is the subject of a portrait.
		best := 0

		for i := range faces {
			if faces[i].Area.Scale > faces[best].Area.Scale {
				best = i
			}
		}

		f := &faces[best]

		for name, embedder := range embedders {
			crop, cropErr := benchmarkCrop(embedder, img, f)

			if cropErr != nil {
				stats[name].failed++
				continue
			}

			start := time.Now()
			embeddings := embedder.Run(crop)
			elapsed[name] += time.Since(start)

			if embeddings.Empty() {
				stats[name].failed++
				continue
			}

			stats[name].embedded++

			if collected[item.subset] == nil {
				collected[item.subset] = make(map[string]map[ModelName][]Embedding)
			}

			if collected[item.subset][item.subject] == nil {
				collected[item.subset][item.subject] = make(map[ModelName][]Embedding)
			}

			collected[item.subset][item.subject][name] = append(collected[item.subset][item.subject][name], embeddings[0])
		}
	}

	for name := range embedders {
		if stats[name].embedded > 0 {
			stats[name].perFace = elapsed[name] / time.Duration(stats[name].embedded)
		}
	}

	// PhotoPrism accepts an automatic match up to this distance from a cluster centroid,
	// so it shows whether the current thresholds still hold for a candidate model.
	legacyDist := ClusterRadius + MatchDist

	subsets := make([]benchmarkSubset, 0, len(collected))

	for subsetName := range collected {
		subsets = append(subsets, benchmarkSubset{name: subsetName})
	}

	sort.Slice(subsets, func(i, j int) bool { return subsets[i].name < subsets[j].name })

	for _, name := range sortedModelNames(embedders) {
		for i := range subsets {
			var subjects []benchmarkSubject

			for subject, byModel := range collected[subsets[i].name] {
				if len(byModel[name]) > 0 {
					subjects = append(subjects, benchmarkSubject{name: subject, embeddings: byModel[name]})
				}
			}

			sort.Slice(subjects, func(a, b int) bool { return subjects[a].name < subjects[b].name })

			key := subsets[i].name

			if key == "" {
				key = "all"
			}

			stats[name].bySubset[key] = evaluatePairs(buildPairs(subjects, maxNegatives), legacyDist)
			stats[name].subsetKeys = append(stats[name].subsetKeys, key)
		}
	}

	reportBenchmark(t, datasetPath, len(images), detectFailed, threads, legacyDist, sortedModelNames(embedders), stats)
}

// modelSizeBytes returns the total size of a model file or SavedModel directory.
func modelSizeBytes(path string) (size int64) {
	if info, err := os.Stat(path); err != nil {
		return 0
	} else if !info.IsDir() {
		return info.Size()
	}

	entries, err := os.ReadDir(path)

	if err != nil {
		return 0
	}

	for _, entry := range entries {
		if info, infoErr := entry.Info(); infoErr == nil && !info.IsDir() {
			size += info.Size()
		}
	}

	return size
}

// sortedModelNames returns the loaded model names in a stable order.
func sortedModelNames(embedders map[ModelName]Embedder) []ModelName {
	names := make([]ModelName, 0, len(embedders))

	for name := range embedders {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// reportBenchmark prints the comparison as Markdown tables that can be pasted into docs.
func reportBenchmark(t *testing.T, datasetPath string, images, detectFailed, threads int, legacyDist float64, names []ModelName, stats map[ModelName]*benchmarkStats) {
	t.Helper()

	var b strings.Builder

	b.WriteString("\n### Face Embedding Model Comparison\n\n")
	fmt.Fprintf(&b, "- Dataset: `%s`\n", datasetPath)
	fmt.Fprintf(&b, "- Images: %d (%d without a usable detection)\n", images, detectFailed)
	fmt.Fprintf(&b, "- Inference threads: %d\n", threads)
	fmt.Fprintf(&b, "- Legacy accept distance (ClusterRadius + MatchDist): %.2f\n\n", legacyDist)

	b.WriteString("| Model | Dim | Model Size | ms/Face | Embedded | Failed |\n")
	b.WriteString("|:------|----:|-----------:|--------:|---------:|-------:|\n")

	for _, name := range names {
		s := stats[name]
		fmt.Fprintf(&b, "| %s | %d | %.1f MB | %.1f | %d | %d |\n",
			name, s.dims, float64(s.sizeBytes)/(1024*1024), float64(s.perFace.Microseconds())/1000, s.embedded, s.failed)
	}

	subsetKeys := map[string]bool{}

	for _, name := range names {
		for key := range stats[name].bySubset {
			subsetKeys[key] = true
		}
	}

	keys := make([]string, 0, len(subsetKeys))

	for key := range subsetKeys {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		fmt.Fprintf(&b, "\n#### Subset: %s\n\n", key)
		b.WriteString("| Model | Pairs (same/diff) | AUC | EER | Acc | Best d | TAR@FAR=1% | d | TAR@FAR=0.1% | d | Mean d same/diff | TAR/FAR at legacy d |\n")
		b.WriteString("|:------|:------------------|----:|----:|----:|-------:|-----------:|--:|-------------:|--:|:-----------------|:--------------------|\n")

		for _, name := range names {
			r, found := stats[name].bySubset[key]

			if !found || r.Positives == 0 {
				continue
			}

			fmt.Fprintf(&b, "| %s | %d/%d | %.4f | %.4f | %.4f | %.3f | %.4f | %.3f | %.4f | %.3f | %.3f / %.3f | %.4f / %.4f |\n",
				name, r.Positives, r.Negatives, r.AUC, r.EER, r.Accuracy, r.AccThresh,
				r.TAR1e2, r.Thresh1e2, r.TAR1e3, r.Thresh1e3, r.MeanSame, r.MeanDiff, r.LegacyTAR, r.LegacyFAR)
		}
	}

	t.Log(b.String())
}
