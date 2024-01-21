package webdav

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime/debug"
	"strings"
	"time"

	"github.com/emersion/go-webdav"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Client represents a webdav client.
type Client struct {
	client   *webdav.Client
	endpoint *url.URL
	timeout  time.Duration
	mkdir    map[string]bool
}

// clientUrl returns the validated server url including username and password, if specified.
func clientUrl(serverUrl, user, pass string) (*url.URL, error) {
	result, err := url.Parse(serverUrl)

	// Check url.
	if err != nil {
		return nil, err
	} else if result == nil {
		return nil, fmt.Errorf("invalid server url")
	}

	// Set user and password if provided.
	if user != "" {
		result.User = url.UserPassword(user, pass)
	}

	return result, nil
}

// NewClient creates a new WebDAV client for the specified endpoint.
func NewClient(serverUrl, user, pass string, timeout Timeout) (*Client, error) {
	// Create a new http.Client without timeout.
	httpClient := &http.Client{}

	endpoint, err := clientUrl(serverUrl, user, pass)

	if err != nil {
		return nil, err
	}

	serverUrl = endpoint.String()

	log.Debugf("webdav: connecting to %s", clean.Log(serverUrl))

	client, err := webdav.NewClient(httpClient, serverUrl)

	if err != nil {
		return nil, err
	}

	// Create a new webdav.Client wrapper.
	result := &Client{
		client:   client,
		endpoint: endpoint,
		timeout:  Durations[timeout],
		mkdir:    make(map[string]bool, 128),
	}

	return result, nil
}

// withTimeout returns a *webdav.Client with specified total request time.
func (c *Client) withTimeout(timeout time.Duration) *webdav.Client {
	if timeout < 0 {
		return c.client
	} else if timeout == 0 {
		timeout = c.timeout
	}

	// Create webdav client with the specified total request time.
	client, err := webdav.NewClient(&http.Client{Timeout: timeout}, c.endpoint.String())

	if err != nil {
		return c.client
	}

	return client
}

// readDirWithTimeout returns the contents of the specified directory with a request time limit if timeout is not negative.
func (c *Client) readDirWithTimeout(dir string, recursive bool, timeout time.Duration) ([]webdav.FileInfo, error) {
	dir = trimPath(dir)
	return c.withTimeout(timeout).Readdir(dir, recursive)
}

// readDirWithTimeout returns the contents of the specified directory without a request timeout.
func (c *Client) readDir(dir string, recursive bool) ([]webdav.FileInfo, error) {
	return c.readDirWithTimeout(dir, recursive, -1)
}

// Files returns information about files in a directory, optionally recursively.
func (c *Client) Files(dir string, recursive bool) (result fs.FileInfos, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("webdav: %s (panic while listing files)\nstack: %s", r, debug.Stack())
		}
	}()

	dir = trimPath(dir)

	found, err := c.readDir(dir, recursive)

	if err != nil {
		return result, err
	}

	result = make(fs.FileInfos, 0, len(found))

	for _, f := range found {
		if f.IsDir || f.Path == "" || strings.HasPrefix(f.Path, ".") {
			continue
		}

		info := fs.WebFileInfo(f, c.endpoint.Path)

		result = append(result, info)
	}

	return result, nil
}

// Directories returns all subdirectories in a path as string slice.
func (c *Client) Directories(dir string, recursive bool, timeout time.Duration) (result fs.FileInfos, err error) {
	dir = trimPath(dir)

	found, err := c.readDirWithTimeout(dir, recursive, timeout)

	if err != nil {
		return result, err
	}

	result = make(fs.FileInfos, 0, len(found))

	for _, f := range found {
		if !f.IsDir || f.Path == "" || strings.HasPrefix(f.Path, ".") {
			continue
		}

		info := fs.WebFileInfo(f, c.endpoint.Path)

		result = append(result, info)
	}

	return result, err
}

// MkdirAll recursively creates remote directories.
func (c *Client) MkdirAll(dir string) (err error) {
	folders := splitPath(dir)

	if len(folders) == 0 {
		return nil
	}

	dir = ""

	for _, folder := range folders {
		dir = path.Join(dir, folder)
		err = c.Mkdir(dir)
	}

	return err
}

// Mkdir creates a single remote directory.
func (c *Client) Mkdir(dir string) error {
	dir = trimPath(dir)

	if dir == "" || dir == "." || dir == ".." {
		// Ignore.
		return nil
	} else if c.mkdir[dir] {
		// Dir was already created.
		return nil
	}

	c.mkdir[dir] = true

	err := c.client.Mkdir(dir)

	if err == nil {
		return nil
	} else if strings.Contains(err.Error(), "already exists") {
		return nil
	}

	return err
}

// Upload uploads a single file to the remote server.
func (c *Client) Upload(src, dest string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("webdav: %s (panic while uploading)\nstack: %s", r, debug.Stack())
		}
	}()

	dest = trimPath(dest)

	if !fs.FileExists(src) {
		return fmt.Errorf("file %s not found", clean.Log(path.Base(src)))
	}

	f, err := os.OpenFile(src, os.O_RDONLY, 0)

	if err != nil {
		log.Errorf("webdav: %s", clean.Error(err))
		return fmt.Errorf("webdav: failed to read %s", clean.Log(path.Base(src)))
	}

	defer f.Close()

	var writer io.WriteCloser

	writer, err = c.client.Create(dest)

	if err != nil {
		log.Errorf("webdav: %s", clean.Error(err))
		return fmt.Errorf("webdav: failed to write %s", clean.Log(dest))
	}

	defer writer.Close()

	if _, err = io.Copy(writer, f); err != nil {
		log.Errorf("webdav: %s", clean.Error(err))
		return fmt.Errorf("webdav: failed to upload %s", clean.Log(dest))
	}

	return nil
}

// Download downloads a single file to the given location.
func (c *Client) Download(src, dest string, force bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("webdav: %s (panic)\nstack: %s", r, clean.Log(src))
			err = fmt.Errorf("webdav: unexpected error while downloading %s", clean.Log(src))
		}
	}()

	src = trimPath(src)

	// Skip if file already exists.
	if _, err := os.Stat(dest); err == nil && !force {
		return fmt.Errorf("webdav: download skipped, %s already exists", clean.Log(dest))
	}

	dir := path.Dir(dest)
	dirInfo, err := os.Stat(dir)

	if err != nil {
		// Create local storage path.
		if err = fs.MkdirAll(dir); err != nil {
			return fmt.Errorf("webdav: cannot create folder %s (%s)", clean.Log(dir), err)
		}
	} else if !dirInfo.IsDir() {
		return fmt.Errorf("webdav: %s is not a folder", clean.Log(dir))
	}

	var reader io.ReadCloser

	// Start download.
	reader, err = c.client.Open(src)

	// Error?
	if err != nil {
		log.Errorf("webdav: %s", clean.Error(err))
		return fmt.Errorf("webdav: failed to download %s", clean.Log(src))
	}

	defer reader.Close()

	f, err := os.OpenFile(dest, os.O_TRUNC|os.O_RDWR|os.O_CREATE, fs.ModeFile)

	if err != nil {
		log.Errorf("webdav: %s", clean.Error(err))
		return fmt.Errorf("webdav: failed to create %s", clean.Log(path.Base(dest)))
	}

	defer f.Close()

	if _, err = f.ReadFrom(reader); err != nil {
		log.Errorf("webdav: %s", clean.Error(err))
		return fmt.Errorf("webdav: failed writing to %s", clean.Log(path.Base(dest)))
	}

	return nil
}

// DownloadDir downloads all files from a remote to a local directory.
func (c *Client) DownloadDir(src, dest string, recursive, force bool) (errs []error) {
	src = trimPath(src)

	files, err := c.Files(src, recursive)

	if err != nil {
		return append(errs, err)
	}

	for _, file := range files {
		dest := path.Join(dest, file.Abs)

		if _, err = os.Stat(dest); err == nil {
			// File already exists.
			msg := fmt.Errorf("webdav: %s already exists", clean.Log(dest))
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

	return errs
}

// Delete deletes a single file or directory on a remote server.
func (c *Client) Delete(dir string) error {
	dir = trimPath(dir)
	return c.client.RemoveAll(dir)
}
