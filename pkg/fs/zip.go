package fs

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Zip compresses one or many files into a single zip archive file.
func Zip(zipName string, files []string, compress bool) (err error) {
	// Create zip file directory if it does not yet exist.
	if zipDir := filepath.Dir(zipName); zipDir != "" && zipDir != "." {
		err = os.MkdirAll(zipDir, ModeDir)

		if err != nil {
			return err
		}
	}

	var newZipFile *os.File

	if newZipFile, err = os.Create(zipName); err != nil {
		return err
	}

	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip archive.
	for _, fileName := range files {
		if err = ZipFile(zipWriter, fileName, "", compress); err != nil {
			return err
		}
	}

	return nil
}

// ZipFile adds a file to a zip archive, optionally with an alias and compression.
func ZipFile(zipWriter *zip.Writer, fileName, fileAlias string, compress bool) (err error) {
	// Open file.
	fileToZip, err := os.Open(fileName)

	if err != nil {
		return err
	}

	// Close file when done.
	defer fileToZip.Close()

	// Get file information.
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	// Create file info header.
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Set filename alias, if any.
	if fileAlias != "" {
		header.Name = fileAlias
	}

	// Set method to deflate to enable compression,
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	if compress {
		header.Method = zip.Deflate
	} else {
		header.Method = zip.Store
	}

	// Write file info header.
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// Copy file to zip.
	_, err = io.Copy(writer, fileToZip)

	// Return error, if any.
	return err
}

// Unzip extracts the contents of a zip file to the target directory.
func Unzip(zipName, dir string) (files []string, err error) {
	zipReader, err := zip.OpenReader(zipName)

	if err != nil {
		return files, err
	}

	defer zipReader.Close()

	for _, zipFile := range zipReader.File {
		// Skip directories like __OSX and potentially malicious file names containing "..".
		if strings.HasPrefix(zipFile.Name, "__") || strings.Contains(zipFile.Name, "..") {
			continue
		}

		fileName, unzipErr := UnzipFile(zipFile, dir)
		if unzipErr != nil {
			return files, unzipErr
		}

		files = append(files, fileName)
	}

	return files, nil
}

// UnzipFile writes a file from a zip archive to the target destination.
func UnzipFile(f *zip.File, dir string) (fileName string, err error) {
	rc, err := f.Open()
	if err != nil {
		return fileName, err
	}

	defer rc.Close()

	// Compose destination file or directory path.
	fileName = filepath.Join(dir, f.Name)

	// Create destination path if it is a directory.
	if f.FileInfo().IsDir() {
		return fileName, MkdirAll(fileName)
	}

	// If it is a file, make sure its destination directory exists.
	var basePath string

	if lastIndex := strings.LastIndex(fileName, string(os.PathSeparator)); lastIndex > -1 {
		basePath = fileName[:lastIndex]
	}

	if err = MkdirAll(basePath); err != nil {
		return fileName, err
	}

	fd, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return fileName, err
	}

	defer fd.Close()

	_, err = io.Copy(fd, rc)
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}
