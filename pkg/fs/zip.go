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

// ErrArchiveSymlink identifies an unsupported symbolic-link archive entry.
var ErrArchiveSymlink = errors.New("archive symlinks are not supported")

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

	defer func() {
		err = errors.Join(err, newZipFile.Close())
	}()

	zipWriter := zip.NewWriter(newZipFile)
	defer func() {
		err = errors.Join(err, zipWriter.Close())
	}()

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
	defer func() {
		err = errors.Join(err, fileToZip.Close())
	}()

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

// Unzip extracts the contents of a zip file to the target directory, and optional filters select entries.
// A negative total size limit is unlimited and 0 leaves no budget. Each file must leave at least one byte
// of the budget, and once one does not fit, no later file entry is extracted.
func Unzip(zipName, dir string, fileSizeLimit, totalSizeLimit int64, filters ...func(string, bool) bool) (files []string, skipped []string, err error) {
	zipReader, err := zip.OpenReader(zipName)

	if err != nil {
		return files, skipped, err
	}

	defer func() {
		err = errors.Join(err, zipReader.Close())
	}()

	entryLimit := MaxUnzipEntries

	for i, zipFile := range zipReader.File {
		if entryLimit > 0 && i >= entryLimit {
			return files, skipped, fmt.Errorf("zip entry limit exceeded (%d)", entryLimit)
		}

		// Skip directories like __OSX and potentially malicious file names containing "..".
		skipEntry := strings.HasPrefix(zipFile.Name, "__") || strings.Contains(zipFile.Name, "..") || HasReservedComponent(zipFile.Name) || zipFile.Mode()&os.ModeSymlink != 0

		for _, filter := range filters {
			if filter != nil && !filter(zipFile.Name, zipFile.FileInfo().IsDir()) {
				skipEntry = true
				break
			}
		}

		if !skipEntry && fileSizeLimit > 0 {
			if zipFile.UncompressedSize64 > uint64(math.MaxInt64) {
				skipEntry = true
			} else if int64(zipFile.UncompressedSize64) > fileSizeLimit { //nolint:gosec // bounded by MaxInt64 check above
				skipEntry = true
			}
		}

		if skipEntry {
			skipped = append(skipped, zipFile.Name)
			continue
		}

		if zipFile.UncompressedSize64 > uint64(math.MaxInt64) {
			skipped = append(skipped, zipFile.Name)
			continue
		}

		writeLimit := fileSizeLimit

		if totalSizeLimit >= 0 && !zipFile.FileInfo().IsDir() {
			if int64(zipFile.UncompressedSize64) >= totalSizeLimit { //nolint:gosec // safe: capped by check above
				skipped = append(skipped, zipFile.Name)
				totalSizeLimit = 0
				continue
			}

			if writeLimit <= 0 || writeLimit > totalSizeLimit-1 {
				writeLimit = totalSizeLimit - 1
			}
		}

		fileName, written, unzipErr := unzipEntry(zipFile, dir, writeLimit)

		if errors.Is(unzipErr, ErrReservedPath) || errors.Is(unzipErr, ErrArchiveSymlink) {
			skipped = append(skipped, zipFile.Name)
			continue
		}

		if unzipErr != nil {
			return files, skipped, unzipErr
		}

		if totalSizeLimit > 0 {
			totalSizeLimit -= written
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
	fileName, _, err = unzipEntry(f, dir, fileSizeLimit)
	return fileName, err
}

// unzipEntry writes a file from a zip archive to the target destination and returns the number of bytes written.
// A positive limit caps the write, otherwise the declared size does. An entry whose data is incomplete, exceeds
// the limit, or does not match its checksum returns an error, and its file is removed. New files get ModeFile.
func unzipEntry(f *zip.File, dir string, limit int64) (fileName string, written int64, err error) {
	if HasReservedComponent(f.Name) {
		return "", 0, ErrReservedPath
	} else if f.Mode()&os.ModeSymlink != 0 {
		return "", 0, ErrArchiveSymlink
	}

	var opened, complete bool

	// Registered first so that it runs after the deferred closes below.
	defer func() {
		if opened && (!complete || err != nil) {
			if removeErr := os.Remove(fileName); removeErr != nil && !os.IsNotExist(removeErr) {
				err = errors.Join(err, removeErr)
			}
		}
	}()

	rc, err := f.Open()
	if err != nil {
		return fileName, 0, err
	}

	defer func() {
		err = errors.Join(err, rc.Close())
	}()

	// Compose destination file or directory path with safety checks.
	if fileName, err = SafeJoin(dir, f.Name); err != nil {
		return fileName, 0, err
	}

	// Create destination path if it is a directory.
	if f.FileInfo().IsDir() {
		return fileName, 0, MkdirAll(fileName)
	}

	// If it is a file, make sure its destination directory exists.
	var basePath string

	if lastIndex := strings.LastIndex(fileName, string(os.PathSeparator)); lastIndex > -1 {
		basePath = fileName[:lastIndex]
	}

	if err = MkdirAll(basePath); err != nil {
		return fileName, 0, err
	}

	fd, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, ModeFile) //nolint:gosec // destination derived from SafeJoin

	if err != nil {
		return fileName, 0, err
	}

	opened = true

	defer func() {
		err = errors.Join(err, fd.Close())
	}()

	capped := limit > 0

	if !capped {
		switch {
		case f.UncompressedSize64 == 0:
			limit = math.MaxInt64
		case f.UncompressedSize64 > uint64(math.MaxInt64):
			return fileName, 0, fmt.Errorf("zip entry too large")
		default:
			limit = int64(f.UncompressedSize64) //nolint:gosec // safe: capped above
			capped = true
		}
	}

	written, copyErr := io.CopyN(fd, rc, limit)

	if copyErr != nil && !errors.Is(copyErr, io.EOF) {
		return fileName, written, copyErr
	}

	// Read past the limit to detect remaining data and to let the archive reader verify the entry.
	if capped && written >= limit {
		var b [1]byte

		if n, extraErr := rc.Read(b[:]); n > 0 || extraErr == nil {
			return fileName, written, fmt.Errorf("zip entry exceeds limit")
		} else if !errors.Is(extraErr, io.EOF) {
			return fileName, written, extraErr
		}
	}

	complete = true

	return fileName, written, nil
}
