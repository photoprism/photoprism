/*
Package fs provides filesystem related constants and functions.

Copyright (c) 2018 - 2024 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package fs

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
)

var ignoreCase bool

const (
	PathSeparator = string(filepath.Separator)
	Home          = "~"
	HomePath      = Home + PathSeparator
)

// FileExists returns true if file exists and is not a directory.
func FileExists(fileName string) bool {
	if fileName == "" {
		return false
	}

	info, err := os.Stat(fileName)

	return err == nil && !info.IsDir()
}

// FileExistsNotEmpty returns true if file exists, is not a directory, and not empty.
func FileExistsNotEmpty(fileName string) bool {
	if fileName == "" {
		return false
	}

	info, err := os.Stat(fileName)

	return err == nil && !info.IsDir() && info.Size() > 0
}

// PathExists tests if a path exists, and is a directory or symlink.
func PathExists(path string) bool {
	if path == "" {
		return false
	}

	info, err := os.Stat(path)

	if err != nil {
		return false
	}

	m := info.Mode()

	return m&os.ModeDir != 0 || m&os.ModeSymlink != 0
}

// Writable checks if the path is accessible for reading and writing.
func Writable(path string) bool {
	if path == "" {
		return false
	}
	return syscall.Access(path, syscall.O_RDWR) == nil
}

// PathWritable tests if a path exists and is writable.
func PathWritable(path string) bool {
	if !PathExists(path) {
		return false
	}

	return Writable(path)
}

// Overwrite overwrites the file with data. Creates file if not present.
func Overwrite(fileName string, data []byte) bool {
	f, err := os.Create(fileName)
	if err != nil {
		return false
	}

	_, err = f.Write(data)
	return err == nil
}

// Abs returns the full path of a file or directory, "~" is replaced with home.
func Abs(name string) string {
	if name == "" {
		return ""
	}

	if len(name) > 2 && name[:2] == HomePath {
		if usr, err := user.Current(); err == nil {
			name = filepath.Join(usr.HomeDir, name[2:])
		}
	}

	result, err := filepath.Abs(name)

	if err != nil {
		panic(err)
	}

	return result
}

// copyToFile copies the zip file to destination
// if the zip file is a directory, a directory is created at the destination.
func copyToFile(f *zip.File, dest string) (fileName string, err error) {
	rc, err := f.Open()
	if err != nil {
		return fileName, err
	}

	defer rc.Close()

	// Store filename/path for returning and using later on
	fileName = filepath.Join(dest, f.Name)

	if f.FileInfo().IsDir() {
		// Make Folder
		return fileName, os.MkdirAll(fileName, ModeDir)
	}

	// Make File
	var fdir string

	if lastIndex := strings.LastIndex(fileName, string(os.PathSeparator)); lastIndex > -1 {
		fdir = fileName[:lastIndex]
	}

	err = os.MkdirAll(fdir, ModeDir)
	if err != nil {
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

// Download downloads a file from a URL.
func Download(fileName string, url string) error {
	if dir := filepath.Dir(fileName); dir == "" || dir == "/" || dir == "." || dir == ".." {
		return fmt.Errorf("invalid path")
	} else if err := os.MkdirAll(dir, ModeDir); err != nil {
		return err
	}

	// Create the file
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// DirIsEmpty returns true if a directory is empty.
func DirIsEmpty(path string) bool {
	f, err := os.Open(path)

	if err != nil {
		return false
	}

	defer f.Close()

	_, err = f.Readdirnames(1)

	if err == io.EOF {
		return true
	}

	return false
}
