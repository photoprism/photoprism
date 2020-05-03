package fs

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// ReadLines returns all lines in a text file as string slice.
func ReadLines(fileName string) (lines []string, err error) {
	file, err := os.Open(fileName)

	if err != nil {
		return lines, err
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		line, _, err := reader.ReadLine()

		if err == io.EOF {
			break
		} else if err != nil {
			return lines, err
		}

		lines = append(lines, strings.TrimSpace(string(line)))
	}

	return lines, nil
}
