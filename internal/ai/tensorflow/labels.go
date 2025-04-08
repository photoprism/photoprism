package tensorflow

import (
	"bufio"
	"os"
)

// LoadLabels loads the labels of classification models from the specified path and returns them.
func LoadLabels(modelPath string) (labels []string, err error) {
	modelLabels := modelPath + "/labels.txt"

	log.Infof("tensorflow: loading model labels from labels.txt")

	f, err := os.Open(modelLabels)

	if err != nil {
		return labels, err
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
