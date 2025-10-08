package tensorflow

import (
	"bufio"
	iofs "io/fs"
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/clean"
)

func loadLabelsFromPath(path string) (labels []string, err error) {
	log.Infof("vision: loading TensorFlow model labels from %s", path)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	// Labels are separated by newlines
	for scanner.Scan() {
		labels = append(labels, scanner.Text())
	}

	err = scanner.Err()

	return labels, err
}

// LoadLabels loads the labels of classification models from the specified path and returns them.
func LoadLabels(modelPath string, expectedLabels int) (labels []string, err error) {

	dir := os.DirFS(modelPath)
	matches, err := iofs.Glob(dir, "labels*.txt")
	if err != nil {
		return nil, err
	}

	for i := range matches {
		loadedLabels, labelsErr := loadLabelsFromPath(filepath.Join(modelPath, matches[i]))

		if labelsErr != nil {
			return nil, labelsErr
		}

		switch expectedLabels - len(loadedLabels) {
		case 0:
			log.Infof("vision: found valid labels in %s", clean.Log(matches[i]))
			return loadedLabels, nil
		case 1:
			log.Infof("vision: found valid labels in %s, but bias needs to be added", clean.Log(matches[i]))
			return append([]string{"background"}, loadedLabels...), nil
		default:
			log.Infof("vision: invalid labels file, expected %d labels and found %d",
				expectedLabels, len(loadedLabels))
		}
	}
	return nil, os.ErrNotExist
}
