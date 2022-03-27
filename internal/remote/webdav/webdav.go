/*

Package webdav implements sharing with WebDAV servers.

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://photoprism.app/trademark>

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/
package webdav

import (
	"fmt"
	"os"
	"path"
	"runtime/debug"
	"time"

	"github.com/studio-b12/gowebdav"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// Global log instance.
var log = event.Log

type Timeout string

// Request Timeout options.
const (
	TimeoutHigh    Timeout = "high"   // 120 * Second
	TimeoutDefault Timeout = ""       // 60 * Second
	TimeoutMedium  Timeout = "medium" // 60 * Second
	TimeoutLow     Timeout = "low"    // 30 * Second
	TimeoutNone    Timeout = "none"   // 0
)

// Second represents a second on which other timeouts are based.
const Second = time.Second

// MaxRequestDuration is the maximum request duration e.g. for recursive retrieval of large remote directory structures.
const MaxRequestDuration = 30 * time.Minute

// Durations maps Timeout options to specific time durations.
var Durations = map[Timeout]time.Duration{
	TimeoutHigh:    120 * Second,
	TimeoutDefault: 60 * Second,
	TimeoutMedium:  60 * Second,
	TimeoutLow:     30 * Second,
	TimeoutNone:    0,
}

// Client represents a gowebdav.Client wrapper.
type Client struct {
	client  *gowebdav.Client
	timeout Timeout
}

// New creates a new WebDAV client.
func New(url, user, pass string, timeout Timeout) Client {
	// Create a new gowebdav.Client instance.
	client := gowebdav.NewClient(url, user, pass)

	// Create a new gowebdav.Client wrapper.
	result := Client{
		client:  client,
		timeout: timeout,
	}

	return result
}

func (c Client) readDir(path string) ([]os.FileInfo, error) {
	if path == "" {
		path = "/"
	}

	return c.client.ReadDir(path)
}

// Files returns all files in a directory as string slice.
func (c Client) Files(dir string) (result fs.FileInfos, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("webdav: %s (panic while listing files)\nstack: %s", r, debug.Stack())
		}
	}()

	files, err := c.readDir(dir)

	if err != nil {
		return result, err
	}

	for _, file := range files {
		if !file.Mode().IsRegular() {
			continue
		}

		info := fs.NewFileInfo(file, dir)

		result = append(result, info)
	}

	return result, nil
}

// Directories returns all subdirectories in a path as string slice.
func (c Client) Directories(root string, recursive bool, timeout time.Duration) (result fs.FileInfos, err error) {
	start := time.Now()

	if timeout == 0 {
		timeout = Durations[c.timeout]
	}

	result, err = c.fetchDirs(root, recursive, start, timeout)

	if time.Now().Sub(start) >= timeout {
		log.Warnf("webdav: read dir timeout reached")
	}

	return result, err
}

// fetchDirs recursively fetches all directories until the timeout is reached.
func (c Client) fetchDirs(root string, recursive bool, start time.Time, timeout time.Duration) (result fs.FileInfos, err error) {
	files, err := c.readDir(root)

	if err != nil {
		return result, err
	}

	if root == "/" {
		root = ""
	}

	for _, file := range files {
		if !file.Mode().IsDir() {
			continue
		}

		info := fs.NewFileInfo(file, root)

		result = append(result, info)

		if recursive && (timeout < time.Second || time.Now().Sub(start) < timeout) {
			subDirs, err := c.fetchDirs(info.Abs, true, start, timeout)

			if err != nil {
				return result, err
			}

			result = append(result, subDirs...)
		}
	}

	return result, nil
}

// Download downloads a single file to the given location.
func (c Client) Download(from, to string, force bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("webdav: %s (panic)\nstack: %s", r, sanitize.Log(from))
			err = fmt.Errorf("webdav: unexpected error while downloading %s", sanitize.Log(from))
		}
	}()

	// Skip if file already exists.
	if _, err := os.Stat(to); err == nil && !force {
		return fmt.Errorf("webdav: download skipped, %s already exists", sanitize.Log(to))
	}

	dir := path.Dir(to)
	dirInfo, err := os.Stat(dir)

	if err != nil {
		// Create local storage path.
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("webdav: cannot create folder %s (%s)", sanitize.Log(dir), err)
		}
	} else if !dirInfo.IsDir() {
		return fmt.Errorf("webdav: %s is not a folder", sanitize.Log(dir))
	}

	var bytes []byte

	// Start download.
	bytes, err = c.client.Read(from)

	// Error?
	if err != nil {
		log.Errorf("webdav: %s", sanitize.Log(err.Error()))
		return fmt.Errorf("webdav: failed downloading %s", sanitize.Log(from))
	}

	// Write data to file and return.
	return os.WriteFile(to, bytes, os.ModePerm)
}

// DownloadDir downloads all files from a remote to a local directory.
func (c Client) DownloadDir(from, to string, recursive, force bool) (errs []error) {
	files, err := c.Files(from)

	if err != nil {
		return append(errs, err)
	}

	for _, file := range files {
		dest := to + string(os.PathSeparator) + file.Abs

		if _, err = os.Stat(dest); err == nil {
			// File already exists.
			msg := fmt.Errorf("webdav: %s already exists", sanitize.Log(dest))
			log.Warn(msg)
			errs = append(errs, msg)
			continue
		}

		if err = c.Download(file.Abs, dest, force); err != nil {
			// Failed to download file.
			errs = append(errs, err)
			log.Error(err)
			continue
		}
	}

	if !recursive {
		return errs
	}

	dirs, err := c.Directories(from, false, MaxRequestDuration)

	for _, dir := range dirs {
		errs = append(errs, c.DownloadDir(dir.Abs, to, true, force)...)
	}

	return errs
}

// CreateDir recursively creates directories if they don't exist.
func (c Client) CreateDir(dir string) error {
	if dir == "" || dir == "/" || dir == "." {
		return nil
	}

	return c.client.MkdirAll(dir, os.ModePerm)
}

// Upload uploads a single file to the remote server.
func (c Client) Upload(from, to string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("webdav: %s (panic while uploading)\nstack: %s", r, debug.Stack())
		}
	}()

	file, err := os.Open(from)

	if err != nil || file == nil {
		return err
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	return c.client.WriteStream(to, file, os.ModePerm)
}

// Delete deletes a single file or directory on a remote server.
func (c Client) Delete(path string) error {
	return c.client.Remove(path)
}
