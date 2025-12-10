/*
Package fs provides filesystem related constants and functions.

Copyright (c) 2018 - 2025 PhotoPrism UG. All rights reserved.

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
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"syscall"

	"github.com/photoprism/photoprism/pkg/http/safe"
)

var ignoreCase bool

const (
	// PathSeparator is the filesystem path separator for the current OS.
	PathSeparator = string(filepath.Separator)
	// Home represents the tilde shorthand for the user's home directory.
	Home = "~"
	// HomePath expands Home with a trailing separator.
	HomePath = Home + PathSeparator
)

// Stat returns the os.FileInfo for the given file path, or an error if it does not exist.
func Stat(filePath string) (os.FileInfo, error) {
	if filePath == "" {
		return nil, fmt.Errorf("empty filepath")
	}

	return os.Stat(filePath)
}

// SocketExists returns true if the specified socket exists and is not a regular file or directory.
func SocketExists(socketName string) bool {
	if socketName == "" {
		return false
	}

	// Check if path exists and is a socket.
	if info, err := os.Stat(socketName); err != nil {
		return false
	} else if mode := info.Mode(); info.IsDir() || mode.IsRegular() || mode.Type() != os.ModeSocket {
		return false
	}

	return true
}

// Exists returns true if the specified file system path exists,
// regardless of whether it is a file, directory, or link.
func Exists(fsPath string) bool {
	if fsPath == "" {
		return false
	}

	_, err := os.Stat(fsPath)

	return err == nil
}

// FileExists returns true if a file exists at the specified path.
func FileExists(filePath string) bool {
	if filePath == "" {
		return false
	}

	info, err := os.Stat(filePath)

	return err == nil && !info.IsDir()
}

// FileExistsNotEmpty returns true if file exists, is not a directory, and not empty.
func FileExistsNotEmpty(filePath string) bool {
	if filePath == "" {
		return false
	}

	info, err := os.Stat(filePath)

	return err == nil && !info.IsDir() && info.Size() > 0
}

// FileExistsIsEmpty returns true if the file exists, but is empty.
func FileExistsIsEmpty(filePath string) bool {
	if filePath == "" {
		return false
	}

	info, err := os.Stat(filePath)

	return err == nil && !info.IsDir() && info.Size() == 0
}

// FileSize returns the size of a file in bytes or -1 in case of an error.
func FileSize(filePath string) int64 {
	if filePath == "" {
		return -1
	}

	info, err := os.Stat(filePath)

	if err != nil || info == nil {
		return -1
	} else if info.IsDir() {
		return -1
	}

	return info.Size()
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

// DeviceExists tests if a path exists, and is a device.
func DeviceExists(path string) bool {
	if path == "" {
		return false
	}

	info, err := os.Stat(path)

	if err != nil {
		return false
	}

	m := info.Mode()

	return m&os.ModeDevice != 0 || m&os.ModeCharDevice != 0
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
	if path == "" {
		return false
	} else if !PathExists(path) {
		return false
	}

	return Writable(path)
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

// Download downloads a file from a URL.
func Download(fileName string, url string) error {
	// Preserve existing semantics but with safer network behavior.
	// Allow private IPs by default to avoid breaking intended internal downloads.
	return safe.Download(fileName, url, &safe.Options{AllowPrivate: true})
}

// DirIsEmpty returns true if a directory is empty.
func DirIsEmpty(path string) bool {
	f, err := os.Open(path) //nolint:gosec // path provided by caller; intended to access filesystem

	if err != nil {
		return false
	}

	defer f.Close()

	_, err = f.Readdirnames(1)
	return err == io.EOF
}
