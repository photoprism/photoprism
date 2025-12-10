package fs

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// MaxUnzipEntries caps the number of entries extracted by Unzip. It may be tuned via config/env later.
var MaxUnzipEntries = 100000

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

	if newZipFile, err = os.Create(zipName); err != nil { //nolint:gosec // zipName provided by caller
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
	fileToZip, err := os.Open(fileName) //nolint:gosec // fileName provided by caller

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
// totalSizeLimit: 0 means unlimited; -1 also means unlimited (reserved for backward compatibility).
func Unzip(zipName, dir string, fileSizeLimit, totalSizeLimit int64) (files []string, skipped []string, err error) {
	zipReader, err := zip.OpenReader(zipName)

	if err != nil {
		return files, skipped, err
	}

	defer zipReader.Close()

	// Treat 0 as no limit; negative also unlimited.
	if totalSizeLimit == 0 {
		totalSizeLimit = -1
	}

	entryLimit := MaxUnzipEntries

	for i, zipFile := range zipReader.File {
		if entryLimit > 0 && i >= entryLimit {
			return files, skipped, fmt.Errorf("zip entry limit exceeded (%d)", entryLimit)
		}

		// Skip directories like __OSX and potentially malicious file names containing "..".
		if strings.HasPrefix(zipFile.Name, "__") || strings.Contains(zipFile.Name, "..") ||
			fileSizeLimit > 0 && zipFile.UncompressedSize64 > uint64(fileSizeLimit) {
			skipped = append(skipped, zipFile.Name)
			continue
		}

		if zipFile.UncompressedSize64 > uint64(math.MaxInt64) {
			skipped = append(skipped, zipFile.Name)
			continue
		}

		if totalSizeLimit > 0 {
			entrySize := int64(zipFile.UncompressedSize64) //nolint:gosec // safe: capped by check above

			totalSizeLimit -= entrySize

			if totalSizeLimit < 1 {
				skipped = append(skipped, zipFile.Name)
				totalSizeLimit = 0
				continue
			}
		}

		fileName, unzipErr := unzipFileWithLimit(zipFile, dir, fileSizeLimit)
		if unzipErr != nil {
			return files, skipped, unzipErr
		}

		files = append(files, fileName)
	}

	return files, skipped, nil
}

// UnzipFile writes a file from a zip archive to the target destination.
func UnzipFile(f *zip.File, dir string) (fileName string, err error) {
	return unzipFileWithLimit(f, dir, 0)
}

// unzipFileWithLimit writes a file from a zip archive to the target destination while applying a size limit.
func unzipFileWithLimit(f *zip.File, dir string, fileSizeLimit int64) (fileName string, err error) {
	rc, err := f.Open()
	if err != nil {
		return fileName, err
	}

	defer rc.Close()

	// Compose destination file or directory path with safety checks.
	if fileName, err = safeJoin(dir, f.Name); err != nil {
		return fileName, err
	}

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

	fd, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode()) //nolint:gosec // destination derived from safeJoin
	if err != nil {
		return fileName, err
	}

	defer fd.Close()

	limit := fileSizeLimit

	if limit <= 0 {
		switch {
		case f.UncompressedSize64 == 0:
			limit = math.MaxInt64
		case f.UncompressedSize64 > uint64(math.MaxInt64):
			return fileName, fmt.Errorf("zip entry too large")
		default:
			limit = int64(f.UncompressedSize64) //nolint:gosec // safe: capped above
		}
	}

	written, copyErr := io.CopyN(fd, rc, limit)
	if copyErr != nil && !errors.Is(copyErr, io.EOF) && !errors.Is(copyErr, io.ErrUnexpectedEOF) {
		return fileName, copyErr
	}

	// Abort if the entry exceeded the configured limit.
	if written >= limit && (fileSizeLimit > 0 || f.UncompressedSize64 > 0) {
		// Drain a single byte to see if more data remains (indicating truncation).
		var b [1]byte
		if _, extraErr := rc.Read(b[:]); extraErr == nil {
			return fileName, fmt.Errorf("zip entry exceeds limit")
		}
	}

	return fileName, nil
}

// safeJoin joins a base directory with a relative name and ensures
// that the resulting path stays within the base directory. Absolute
// paths and Windows-style volume names are rejected.
func safeJoin(baseDir, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("invalid zip path")
	}

	// Normalize separators so mixed '/' and '\\' are handled consistently.
	name = strings.ReplaceAll(name, "\\", "/")

	// Reject Windows-style volume names even on non-Windows platforms.
	if len(name) >= 2 && name[1] == ':' && ((name[0] >= 'A' && name[0] <= 'Z') || (name[0] >= 'a' && name[0] <= 'z')) {
		return "", fmt.Errorf("invalid zip path: absolute or volume path not allowed")
	}

	if filepath.IsAbs(name) || filepath.VolumeName(name) != "" {
		return "", fmt.Errorf("invalid zip path: absolute or volume path not allowed")
	}

	cleaned := filepath.Clean(name)
	base := filepath.Clean(baseDir)

	dest := filepath.Join(base, cleaned)
	rel, err := filepath.Rel(base, dest)
	if err != nil {
		return "", fmt.Errorf("invalid zip path: %w", err)
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid zip path: outside target directory")
	}

	return dest, nil
}
