/*
Package entity implementing sharing with WebDAV servers.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki
*/
package webdav

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/studio-b12/gowebdav"
)

var log = event.Log

type Client struct {
	client *gowebdav.Client
}

// Connect creates a new WebDAV client.
func Connect(url, user, pass string) Client {
	clt := gowebdav.NewClient(url, user, pass)

	result := Client{client: clt}

	return result
}

// Files returns all files in path as string slice.
func (c Client) Files(path string) (result []string, err error) {
	files, err := c.client.ReadDir(path)

	if err != nil {
		return result, err
	}

	for _, file := range files {
		if !file.Mode().IsRegular() { continue }
		result = append(result, fmt.Sprintf("%s/%s", path, file.Name()))
	}

	return result, nil
}

// Directories returns all sub directories in path as string slice.
func (c Client) Directories(path string) (result []string, err error) {
	files, err := c.client.ReadDir(path)

	if err != nil {
		return result, err
	}

	for _, file := range files {
		if !file.Mode().IsDir() { continue }
		result = append(result, fmt.Sprintf("%s/%s", path, file.Name()))
	}

	return result, nil
}

// Download downloads a single file to the given location.
func (c Client) Download(from, to string) error {
	dir := path.Dir(to)
	dirInfo, err := os.Stat(dir)

	if err != nil {
		// Create directory
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("webdav: can't create %s (%s)", dir, err)
		}
	} else if !dirInfo.IsDir() {
		return fmt.Errorf("webdav: %s is not a directory", dir)
	}

	bytes, err := c.client.Read(from)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(to, bytes, 0644)
}

// DownloadDir downloads all files from a remote to a local directory.
func (c Client) DownloadDir(from, to string, recursive bool) (errs []error) {
	files, err := c.Files(from)

	if err != nil {
		return append(errs, err)
	}

	for _, file := range files {
		dest := to + string(os.PathSeparator) + file

		if _, err := os.Stat(dest); err == nil {
			// File exists
			msg := fmt.Errorf("webdav: %s exists", dest)
			errs = append(errs, msg)
			log.Error(msg)
			continue
		}

		if err := c.Download(file, dest); err != nil {
			msg := fmt.Errorf("webdav: %s", err)
			errs = append(errs, msg)
			log.Error(msg)
			continue
		}
	}

	if !recursive {
		return errs
	}

	dirs, err := c.Directories(from)

	for _, dir := range dirs {
		errs = append(errs, c.DownloadDir(dir, to, true)...)
	}

	return errs
}

// Upload uploads a single file to the remote server.
func (c Client) Upload(from, to string) error {
	file, err := os.Open(from)

	if err != nil {
		return err
	}

	defer file.Close()

	return c.client.WriteStream(to, file, 0644)
}

// Delete deletes a single file or directory on a remote server.
func (c Client) Delete(path string) error {
	return c.client.Remove(path)
}
