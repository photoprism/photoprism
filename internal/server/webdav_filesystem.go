package server

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/net/webdav"

	"github.com/photoprism/photoprism/pkg/fs"
)

// webDAVMutationEntries bounds filesystem entries inspected before directory mutations.
const webDAVMutationEntries = 100000

// webDAVFileSystem applies the shared path policy to a native WebDAV mount.
type webDAVFileSystem struct {
	base webdav.Dir
	root string
}

// newWebDAVFileSystem returns the filesystem boundary for one configured mount.
func newWebDAVFileSystem(dir string) *webDAVFileSystem {
	if dir == "" {
		dir = "."
	}

	return &webDAVFileSystem{base: webdav.Dir(dir), root: dir}
}

// allowed checks logical path components without resolving operator-managed links.
func (d *webDAVFileSystem) allowed(name string) bool {
	policy := fs.ReservedPathPolicy{AllowIgnoreNames: true}
	return !policy.HasReservedComponent(path.Clean("/" + name))
}

// canWriteManagedFiles reports effective full photo authority for managed-file mutations.
func canWriteManagedFiles(ctx context.Context) bool {
	allowed, _ := ctx.Value(webDAVManagedWriteKey{}).(bool)
	return allowed
}

// managedFileName identifies file names requiring full photo authority to modify.
func managedFileName(name string) bool {
	return fs.FileType(name) == fs.SidecarYaml || fs.IsIgnoreFileName(name)
}

// managedFile recognizes logical YAML and ignore-file names, excluding directories.
func (d *webDAVFileSystem) managedFile(ctx context.Context, name string) bool {
	name = path.Clean("/" + name)

	if !managedFileName(name) {
		return false
	}

	info, err := d.base.Stat(ctx, name)
	return err != nil || !info.IsDir()
}

// Stat returns metadata only for visible paths.
func (d *webDAVFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	if !d.allowed(name) {
		return nil, os.ErrNotExist
	}
	return d.base.Stat(ctx, name)
}

// OpenFile opens eligible files and filters directory enumeration.
func (d *webDAVFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	if !d.allowed(name) {
		return nil, os.ErrNotExist
	}
	if flag&os.O_CREATE != 0 && strings.Contains(name, "\\") {
		// LOCK creates a placeholder for a name that does not exist yet, which is the one
		// create the method gate ahead of the handler does not see.
		return nil, os.ErrPermission
	}
	if flag&(os.O_WRONLY|os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_APPEND) != 0 && !canWriteManagedFiles(ctx) && d.managedFile(ctx, name) {
		return nil, os.ErrPermission
	}
	f, err := d.base.OpenFile(ctx, name, flag, perm)
	if err != nil {
		return nil, err
	}
	return &webDAVFile{File: f, fs: d, name: name}, nil
}

// Mkdir creates eligible collections.
func (d *webDAVFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	if !d.allowed(name) {
		return os.ErrPermission
	}
	return d.base.Mkdir(ctx, name, perm)
}

// RemoveAll refuses trees containing protected entries before removing them.
func (d *webDAVFileSystem) RemoveAll(ctx context.Context, name string) error {
	if err := d.checkMutation(ctx, name, webDAVMutationEntries); err != nil {
		return err
	}
	return d.base.RemoveAll(ctx, name)
}

// Rename validates both trees before moving a file or collection.
func (d *webDAVFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	if !canWriteManagedFiles(ctx) && managedFileName(path.Clean("/"+newName)) {
		info, err := os.Lstat(filepath.Join(d.root, filepath.FromSlash(path.Clean("/"+oldName))))
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return os.ErrPermission
		}
	}

	if err := d.checkMutation(ctx, oldName, webDAVMutationEntries); err != nil {
		return err
	}
	if err := d.checkMutation(ctx, newName, webDAVMutationEntries); err != nil {
		return err
	}
	return d.base.Rename(ctx, oldName, newName)
}

// checkMutation scans raw entries in bounded batches without following directory links.
func (d *webDAVFileSystem) checkMutation(ctx context.Context, name string, limit int) error {
	return d.checkTree(ctx, name, limit, false)
}

// checkTree inspects mutation eligibility, optionally omitting invisible COPY source children.
func (d *webDAVFileSystem) checkTree(ctx context.Context, name string, limit int, skipReserved bool) error {
	if !d.allowed(name) {
		return os.ErrPermission
	}
	root := filepath.Join(d.root, filepath.FromSlash(path.Clean("/"+name)))
	remaining := limit
	var visit func(string, string, int) error
	visit = func(filename, logical string, depth int) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		remaining--
		if remaining < 0 || depth > 128 {
			return os.ErrPermission
		}
		if !d.allowed(logical) {
			if skipReserved {
				return nil
			}
			return os.ErrPermission
		}
		info, err := os.Lstat(filename)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if skipReserved && info.Mode()&os.ModeSymlink != 0 {
			info, err = os.Stat(filename)
			if err != nil {
				return err
			}
		}
		if !info.IsDir() {
			if !canWriteManagedFiles(ctx) && d.managedFile(ctx, logical) {
				return os.ErrPermission
			}
			return nil
		}
		dir, err := os.Open(filename) //nolint:gosec // Path follows the checked mount namespace and eligible linked originals.
		if err != nil {
			return err
		}
		defer dir.Close()
		for {
			entries, readErr := dir.ReadDir(128)
			for _, entry := range entries {
				if err = visit(filepath.Join(filename, entry.Name()), path.Join(logical, entry.Name()), depth+1); err != nil {
					return err
				}
			}
			if readErr == io.EOF {
				return nil
			}
			if readErr != nil {
				return readErr
			}
		}
	}
	return visit(root, name, 0)
}

// preflight validates request paths before upstream COPY/MOVE can replace a destination.
func (d *webDAVFileSystem) preflight(r *http.Request, prefix string) int {
	name := strings.TrimPrefix(r.URL.Path, prefix)
	if !d.allowed(name) {
		return http.StatusNotFound
	}
	if WebDAVWriteMethod(r.Method) && r.Method != "MKCOL" && r.Method != "COPY" && !canWriteManagedFiles(r.Context()) && d.managedFile(r.Context(), name) {
		return http.StatusForbidden
	}
	switch r.Method {
	case "DELETE":
		if d.checkMutation(r.Context(), name, webDAVMutationEntries) != nil {
			return http.StatusForbidden
		}
	case "MOVE", "COPY":
		dest, err := url.Parse(r.Header.Get("Destination"))
		if err != nil || r.Header.Get("Destination") == "" {
			return http.StatusBadRequest
		}
		if dest.Host != "" && dest.Host != r.Host {
			return http.StatusBadGateway
		}
		if dest.Path != prefix && !strings.HasPrefix(dest.Path, prefix+"/") {
			return http.StatusNotFound
		}
		target := strings.TrimPrefix(dest.Path, prefix)
		if target == "" {
			return http.StatusBadGateway
		}
		if path.Clean("/"+target) == "/" {
			return http.StatusForbidden
		}
		if r.Method == "MOVE" {
			if path.Clean("/"+name) == "/" {
				return http.StatusForbidden
			}
			if _, err := d.base.Stat(r.Context(), name); err != nil {
				return http.StatusNotFound
			}
		}
		if !d.allowed(target) {
			return http.StatusForbidden
		}
		if !canWriteManagedFiles(r.Context()) {
			info, err := d.base.Stat(r.Context(), name)
			if err != nil {
				return http.StatusNotFound
			}
			if r.Method == "MOVE" {
				info, err = os.Lstat(filepath.Join(d.root, filepath.FromSlash(path.Clean("/"+name))))
				if err != nil {
					return http.StatusNotFound
				}
			}
			if !info.IsDir() && (managedFileName(path.Clean("/"+target)) || d.managedFile(r.Context(), target)) {
				return http.StatusForbidden
			}
			if r.Method == "COPY" && r.Header.Get("Depth") != "0" && info.IsDir() && d.checkTree(r.Context(), name, webDAVMutationEntries, true) != nil {
				return http.StatusForbidden
			}
		}

		if r.Method == "MOVE" && d.checkMutation(r.Context(), name, webDAVMutationEntries) != nil {
			return http.StatusForbidden
		}
		if r.Header.Get("Overwrite") != "F" && d.checkMutation(r.Context(), target, webDAVMutationEntries) != nil {
			return http.StatusForbidden
		}
	}
	return http.StatusOK
}

// webDAVFile filters reserved directory entries while retaining ordinary file operations.
type webDAVFile struct {
	webdav.File
	fs   *webDAVFileSystem
	name string
}

// Readdir returns only visible entries and preserves count and EOF semantics.
func (f *webDAVFile) Readdir(count int) ([]os.FileInfo, error) {
	result := make([]os.FileInfo, 0)
	for {
		batch := 128
		if count > 0 && count-len(result) < batch {
			batch = count - len(result)
		}
		entries, err := f.File.Readdir(batch)
		for _, entry := range entries {
			if f.fs.allowed(path.Join(f.name, entry.Name())) {
				result = append(result, entry)
			}
		}
		if count > 0 && len(result) >= count {
			return result, nil
		}
		if err == io.EOF {
			if count <= 0 || len(result) > 0 {
				return result, nil
			}
			return result, io.EOF
		}
		if err != nil {
			return result, err
		}
	}
}
