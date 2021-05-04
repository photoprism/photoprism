package fs

import (
	"archive/zip"
	"io"
	"os"
	"strings"
)

// ZipFiles compresses one or many files into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: files is a list of files to add to the zip.
func Zip(filename string, files []string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = AddToZip(zipWriter, file); err != nil {
			return err
		}
	}

	return nil
}

func AddToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)

	if err != nil {
		return err
	}

	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

// Extract Zip file in destination directory
func Unzip(src, dest string) (fileNames []string, err error) {
	r, err := zip.OpenReader(src)

	if err != nil {
		return fileNames, err
	}

	defer r.Close()

	for _, f := range r.File {
		// Skip directories like __OSX and potentially malicious file names containing "..".
		if strings.HasPrefix(f.Name, "__") || strings.Contains(f.Name, "..") {
			continue
		}

		fn, err := copyToFile(f, dest)
		if err != nil {
			return fileNames, err
		}

		fileNames = append(fileNames, fn)
	}

	return fileNames, nil
}
