package webdav

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime/debug"
	"strings"
	"time"

	"github.com/emersion/go-webdav"

	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/safe"
)

// Client represents a webdav client.
type Client struct {
	client   *webdav.Client
	ctx      context.Context
	endpoint *url.URL
	timeout  time.Duration
	mkdir    map[string]bool
	cidrs    []*net.IPNet
}

// clientUrl returns the validated server url including username and password, if specified.
func clientUrl(serverUrl, user, pass string) (*url.URL, error) {
	result, err := safe.URL(serverUrl)

	if err != nil {
		return nil, err
	}

	// Set user and password if provided.
	if user != "" {
		result.User = url.UserPassword(user, pass)
	}

	return result, nil
}

// newTransferHTTPClient returns an HTTP client with connection-level safeguards but no total transfer deadline.
func newTransferHTTPClient(cidrs []*net.IPNet) *http.Client {
	client := service.NewHTTPClient(0, cidrs)
	transport, ok := client.Transport.(*http.Transport)

	if !ok || transport == nil {
		return client
	}

	transport = transport.Clone()

	if baseDial := transport.DialContext; baseDial != nil {
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			if transferConnectTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, transferConnectTimeout)
				defer cancel()
			}

			return baseDial(ctx, network, addr)
		}
	}

	transport.TLSHandshakeTimeout = transferTLSHandshakeTimeout
	transport.IdleConnTimeout = transferIdleConnTimeout
	transport.ExpectContinueTimeout = transferExpectContinueTimeout
	client.Transport = transport

	return client
}

// NewClient creates a new WebDAV client for the specified endpoint.
func NewClient(serverUrl, user, pass string, timeout Timeout, servicesCIDR string) (*Client, error) {
	endpoint, err := clientUrl(serverUrl, user, pass)

	if err != nil {
		return nil, err
	}

	allowedCIDRs, err := service.ParseCIDRs(servicesCIDR)

	if err != nil {
		return nil, err
	}

	if validateErr := service.ValidateURLHost(endpoint, allowedCIDRs, 5*time.Second); validateErr != nil {
		return nil, validateErr
	}

	serverUrl = endpoint.String()

	log.Debugf("webdav: connecting to %s", clean.Log(serverUrl))

	client, err := webdav.NewClient(newTransferHTTPClient(allowedCIDRs), serverUrl)

	if err != nil {
		return nil, err
	}

	// Create a new webdav.Client wrapper.
	result := &Client{
		client:   client,
		ctx:      context.Background(),
		endpoint: endpoint,
		timeout:  Durations[timeout],
		mkdir:    make(map[string]bool, 128),
		cidrs:    allowedCIDRs,
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
	client, err := webdav.NewClient(service.NewHTTPClient(timeout, c.cidrs), c.endpoint.String())

	if err != nil {
		return c.client
	}

	return client
}

// effectiveTimeout returns the effective request timeout used by WebDAV calls.
func (c *Client) effectiveTimeout(timeout time.Duration) time.Duration {
	if timeout < 0 {
		return -1
	} else if timeout == 0 {
		return c.timeout
	}

	return timeout
}

// timeoutContext returns a request context bounded by the configured timeout when applicable.
func (c *Client) timeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout = c.effectiveTimeout(timeout); timeout > 0 {
		return context.WithTimeout(c.ctx, timeout)
	}

	return c.ctx, func() {}
}

// timeoutRequest returns a timeout-aware client and context for outbound WebDAV calls.
func (c *Client) timeoutRequest(timeout time.Duration) (*webdav.Client, context.Context, context.CancelFunc) {
	ctx, cancel := c.timeoutContext(timeout)
	return c.withTimeout(timeout), ctx, cancel
}

// readDirContext returns the contents of the specified directory using the provided request context.
func (c *Client) readDirContext(ctx context.Context, dir string, recursive bool, timeout time.Duration) ([]webdav.FileInfo, error) {
	dir = trimPath(dir)
	return c.withTimeout(timeout).ReadDir(ctx, dir, recursive)
}

// appendUniqueEntries adds new WebDAV entries only once while preserving response order.
func appendUniqueEntries(result []webdav.FileInfo, found []webdav.FileInfo, seen map[string]bool) []webdav.FileInfo {
	for _, entry := range found {
		entryPath := trimPath(entry.Path)

		if seen[entryPath] {
			continue
		}

		seen[entryPath] = true
		result = append(result, entry)
	}

	return result
}

// appendTraversalDirs adds unseen child directories from a non-recursive PROPFIND response to the queue.
func appendTraversalDirs(queue []string, current string, found []webdav.FileInfo, seen map[string]bool) []string {
	current = trimPath(current)

	for _, entry := range found {
		if !entry.IsDir {
			continue
		}

		entryPath := trimPath(entry.Path)

		if entryPath == "" || entryPath == current || isHiddenPath(entryPath) || seen[entryPath] {
			continue
		}

		seen[entryPath] = true
		queue = append(queue, entryPath)
	}

	return queue
}

// readDirFallback traverses directories with repeated non-recursive PROPFIND requests.
func (c *Client) readDirFallback(ctx context.Context, dir string, timeout time.Duration) (result []webdav.FileInfo, requests int, err error) {
	queue := []string{trimPath(dir)}
	traversed := map[string]bool{trimPath(dir): true}
	seenEntries := map[string]bool{}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		requests++

		found, readErr := c.readDirContext(ctx, current, false, timeout)

		if readErr != nil {
			return result, requests, readErr
		}

		result = appendUniqueEntries(result, found, seenEntries)
		queue = appendTraversalDirs(queue, current, found, traversed)
	}

	return result, requests, nil
}

// Files returns information about files in a directory, optionally recursively.
func (c *Client) Files(dir string, recursive bool) (result fs.FileInfos, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("webdav: %s (panic while listing files)\nstack: %s", r, debug.Stack())
		}
	}()

	dir = trimPath(dir)
	ctx, cancel := c.timeoutContext(0)
	defer cancel()

	found, err := c.readDirContext(ctx, dir, recursive, 0)

	if err != nil {
		return result, err
	}

	result = make(fs.FileInfos, 0, len(found))

	for _, f := range found {
		if f.IsDir || f.Path == "" || isHiddenPath(f.Path) {
			continue
		}

		info := fs.WebFileInfo(f, c.endpoint.Path)

		result = append(result, info)
	}

	return result, nil
}

// Directories returns all subdirectories in a path and falls back to iterative Depth: 1 traversal when needed.
func (c *Client) Directories(dir string, recursive bool, timeout time.Duration) (result fs.FileInfos, err error) {
	dir = trimPath(dir)
	ctx, cancel := c.timeoutContext(timeout)
	defer cancel()

	found, err := c.readDirContext(ctx, dir, recursive, timeout)

	if err != nil && recursive {
		started := time.Now()

		if fallback, requests, fallbackErr := c.readDirFallback(ctx, dir, timeout); fallbackErr == nil {
			log.Infof("webdav: recursive PROPFIND failed for %s, using iterative Depth: 1 fallback after %d requests [%s] (%s)", clean.Log(path.Join("/", dir)), requests, time.Since(started).Round(time.Millisecond), clean.Error(err))
			found = fallback
			err = nil
		} else {
			log.Warnf("webdav: recursive PROPFIND failed for %s (%s)", clean.Log(path.Join("/", dir)), clean.Error(err))
			log.Debugf("webdav: Depth: 1 fallback failed for %s after %d requests [%s] (%s)", clean.Log(path.Join("/", dir)), requests, time.Since(started).Round(time.Millisecond), clean.Error(fallbackErr))
		}
	}

	if err != nil {
		return result, err
	}

	result = make(fs.FileInfos, 0, len(found))

	for _, f := range found {
		if !f.IsDir || f.Path == "" || isHiddenPath(f.Path) {
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
	client, ctx, cancel := c.timeoutRequest(0)
	defer cancel()

	err := client.Mkdir(ctx, dir)

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

	f, err := os.OpenFile(src, os.O_RDONLY, 0) //nolint:gosec // path provided by caller; read-only

	if err != nil {
		log.Errorf("webdav: %s", clean.Error(err))
		return fmt.Errorf("webdav: failed to read %s", clean.Log(path.Base(src)))
	}

	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Debugf("webdav: %s (close source file)", clean.Error(closeErr))
		}
	}()

	var writer io.WriteCloser
	writer, err = c.client.Create(c.ctx, dest)

	if err != nil {
		log.Errorf("webdav: %s", clean.Error(err))
		return fmt.Errorf("webdav: failed to write %s", clean.Log(dest))
	}

	if _, err = io.Copy(writer, f); err != nil {
		_ = writer.Close()
		log.Errorf("webdav: %s", clean.Error(err))
		return fmt.Errorf("webdav: failed to upload %s", clean.Log(dest))
	}

	if closeErr := writer.Close(); closeErr != nil {
		log.Errorf("webdav: %s", clean.Error(closeErr))
		return fmt.Errorf("webdav: failed to finalize upload %s", clean.Log(dest))
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
	if fs.Exists(dest) && !force {
		return fmt.Errorf("webdav: download skipped, %s already exists", clean.Log(dest))
	}

	dir := path.Dir(dest)
	dirInfo, err := fs.Stat(dir)

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
	reader, err = c.client.Open(c.ctx, src)

	// Error?
	if err != nil {
		log.Errorf("webdav: %s", clean.Error(err))
		return fmt.Errorf("webdav: failed to download %s", clean.Log(src))
	}

	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Debugf("webdav: %s (close source stream)", clean.Error(closeErr))
		}
	}()

	f, err := os.OpenFile(dest, os.O_TRUNC|os.O_RDWR|os.O_CREATE, fs.ModeFile) //nolint:gosec // dest provided by caller

	if err != nil {
		log.Errorf("webdav: %s", clean.Error(err))
		return fmt.Errorf("webdav: failed to create %s", clean.Log(path.Base(dest)))
	}

	if _, err = f.ReadFrom(reader); err != nil {
		_ = f.Close()
		log.Errorf("webdav: %s", clean.Error(err))
		return fmt.Errorf("webdav: failed writing to %s", clean.Log(path.Base(dest)))
	}

	if closeErr := f.Close(); closeErr != nil {
		log.Errorf("webdav: %s", clean.Error(closeErr))
		return fmt.Errorf("webdav: failed to finalize %s", clean.Log(path.Base(dest)))
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
		fileName := path.Join(dest, file.Abs)

		// Check if file already exists.
		if fs.Exists(fileName) {
			msg := fmt.Errorf("webdav: %s already exists", clean.Log(fileName))
			log.Warn(msg)
			errs = append(errs, msg)
			continue
		}

		// Download file from remote server.
		if err = c.Download(file.Abs, fileName, force); err != nil {
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
	client, ctx, cancel := c.timeoutRequest(0)
	defer cancel()
	return client.RemoveAll(ctx, dir)
}
