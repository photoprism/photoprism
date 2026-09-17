package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/webdav"

	"github.com/photoprism/photoprism/pkg/fs"
)

// TestWebDAVReservedPaths checks reads, writes, listings, and ordinary hidden-file workflows.
func TestWebDAVReservedPaths(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	require.NoError(t, conf.CreateDirectories())
	root := conf.OriginalsPath()
	// put writes a populated filesystem control beneath the mounted directory.
	put := func(name, content string) {
		t.Helper()
		filename := filepath.Join(root, name)
		require.NoError(t, fs.MkdirAll(filepath.Dir(filename)))
		require.NoError(t, os.WriteFile(filename, []byte(content), fs.ModeFile))
	}
	names := append(fs.ReservedPathNames(), ".ENV.production", ".GiT", ".bash_history-04218.tmp", ".python_history", ".mylogin.cnf", ".custom.CNF")
	for _, name := range names {
		put(name+"/photo.jpg", "reserved-control")
	}
	put("gitfile/.git", "reserved-control")
	protectedFiles := []string{fs.PPStorageFilename, fs.SigningKeyFile, fs.JoinTokenFile, fs.ClientSecretFile}
	for _, name := range protectedFiles {
		put("controls/"+name, "reserved-control")
	}
	put("visible.jpg", "visible-control")
	put(".hidden/photo.jpg", "visible-control")
	put(".DS_Store", "visible-control")
	put(".ppignore", "visible-control")
	put("%2egit/photo.jpg", "literal-control")
	router := setupWebDAVRouter(conf)
	base := conf.BaseUri(WebDAVOriginals)
	// request exercises the production mount and preflight with authenticated requests.
	request := func(method, name, body, destination string) *httptest.ResponseRecorder {
		req := webDAVRequest(method, base+name, body)
		req.Header.Set("Depth", "infinity")
		if destination != "" {
			req.Header.Set("Destination", base+destination)
			req.Header.Set("Overwrite", "T")
		}
		out := httptest.NewRecorder()
		router.ServeHTTP(out, req)
		return out
	}
	for i, name := range names {
		t.Run(fmt.Sprintf("Reserved%02d", i), func(t *testing.T) {
			for _, method := range []string{"GET", "HEAD", "POST", "PROPFIND", "OPTIONS", "PUT", "MKCOL", "DELETE", "COPY", "MOVE", "LOCK", "UNLOCK", "PROPPATCH"} {
				out := request(method, "/"+name+"/photo.jpg", "replacement", "/destination.jpg")
				assert.Equal(t, http.StatusNotFound, out.Code, method)
				assert.NotContains(t, out.Body.String(), "reserved-control")
			}
			for _, method := range []string{"COPY", "MOVE"} {
				out := request(method, "/visible.jpg", "", "/"+name+"/replacement.jpg")
				assert.Equal(t, http.StatusForbidden, out.Code, method)
			}
			assert.FileExists(t, filepath.Join(root, name, "photo.jpg"))
			assert.NoFileExists(t, filepath.Join(root, name, "replacement.jpg"))
		})
	}
	t.Run("EncodedAndFileNames", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, request("GET", "/%2egit/photo.jpg", "", "").Code)
		assert.Equal(t, http.StatusNotFound, request("GET", "/gitfile/.git", "", "").Code)
		for _, name := range protectedFiles {
			for _, method := range []string{"GET", "PUT", "DELETE", "MOVE"} {
				assert.Equal(t, http.StatusNotFound, request(method, "/controls/"+name, "replacement", "/moved.txt").Code)
			}

			data, err := os.ReadFile(filepath.Join(root, "controls", name)) //nolint:gosec // Test reads its temporary filesystem control.

			require.NoError(t, err)
			assert.Equal(t, "reserved-control", string(data), name)
		}

		allowed := request("GET", "/%252egit/photo.jpg", "", "")
		assert.Equal(t, http.StatusOK, allowed.Code)
		assert.Equal(t, "literal-control", allowed.Body.String())
	})
	t.Run("PrivilegedWrites", func(t *testing.T) {
		d := newWebDAVFileSystem(root)
		ctx := context.WithValue(context.Background(), webDAVManagedWriteKey{}, true)

		for _, name := range protectedFiles {
			_, err := d.OpenFile(ctx, "controls/"+name, os.O_WRONLY|os.O_TRUNC, fs.ModeFile)
			assert.ErrorIs(t, err, os.ErrNotExist, name)
			assert.ErrorIs(t, d.RemoveAll(ctx, "controls/"+name), os.ErrPermission, name)
		}
	})
	t.Run("Listings", func(t *testing.T) {
		for _, depth := range []string{"1", "infinity"} {
			req := webDAVRequest("PROPFIND", base+"/", "")
			req.Header.Set("Depth", depth)
			out := httptest.NewRecorder()
			router.ServeHTTP(out, req)
			require.Equal(t, http.StatusMultiStatus, out.Code)
			for _, name := range names {
				assert.NotContains(t, out.Body.String(), base+"/"+name+"/")
			}
			assert.Contains(t, out.Body.String(), "visible.jpg")
			assert.Contains(t, out.Body.String(), ".hidden/")
		}
	})
	t.Run("OrdinaryDotFiles", func(t *testing.T) {
		for _, name := range []string{".DS_Store", ".ppignore", ".hidden/photo.jpg"} {
			out := request("GET", "/"+name, "", "")
			require.Equal(t, http.StatusOK, out.Code)
			assert.Equal(t, "visible-control", out.Body.String())
		}
		require.Equal(t, http.StatusCreated, request("PUT", "/._upload.jpg", "staging-control", "").Code)
		require.Equal(t, http.StatusCreated, request("MOVE", "/._upload.jpg", "", "/finished.jpg").Code)
		assert.Equal(t, "staging-control", request("GET", "/finished.jpg", "", "").Body.String())
	})
}

// TestWebDAVReservedAncestors checks refusal before upstream destination replacement.
func TestWebDAVReservedAncestors(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	require.NoError(t, conf.CreateDirectories())
	root := conf.OriginalsPath()
	for name, content := range map[string]string{"tree/.git/config": "reserved-control", "tree/photo.jpg": "visible-control", "destination.jpg": "destination-control", "source.jpg": "source-control", "target/.config/settings": "target-control"} {
		filename := filepath.Join(root, name)
		require.NoError(t, fs.MkdirAll(filepath.Dir(filename)))
		require.NoError(t, os.WriteFile(filename, []byte(content), fs.ModeFile))
	}
	router := setupWebDAVRouter(conf)
	base := conf.BaseUri(WebDAVOriginals)
	for _, tc := range []struct {
		method, source, destination string
		code                        int
	}{
		{"DELETE", "/tree", "", 403}, {"MOVE", "/tree", "/destination.jpg", 403},
		{"COPY", "/source.jpg", "/target", 403}, {"MOVE", "/source.jpg", "/target", 403},
		{"MOVE", "/missing", "/destination.jpg", 404}, {"MOVE", "", "/destination.jpg", 403},
	} {
		req := webDAVRequest(tc.method, base+tc.source, "")
		req.Header.Set("Destination", base+tc.destination)
		req.Header.Set("Overwrite", "T")
		out := httptest.NewRecorder()
		router.ServeHTTP(out, req)
		assert.Equal(t, tc.code, out.Code, tc)
		data, err := os.ReadFile(filepath.Join(root, "destination.jpg")) //nolint:gosec // Test reads its temporary filesystem control.
		require.NoError(t, err)
		assert.Equal(t, "destination-control", string(data))
		assert.FileExists(t, filepath.Join(root, "tree/.git/config"))
		assert.FileExists(t, filepath.Join(root, "target/.config/settings"))
		assert.FileExists(t, filepath.Join(root, "source.jpg"))
	}
	req := webDAVRequest("COPY", base+"/tree", "")
	req.Header.Set("Destination", base+"/copied")
	out := httptest.NewRecorder()
	router.ServeHTTP(out, req)
	require.Equal(t, http.StatusCreated, out.Code, out.Body.String())
	data, err := os.ReadFile(filepath.Join(root, "copied/photo.jpg")) //nolint:gosec // Test reads its temporary filesystem control.
	require.NoError(t, err)
	assert.Equal(t, "visible-control", string(data))
	assert.NoDirExists(t, filepath.Join(root, "copied/.git"))
}

// TestWebDAVFileSystemLinks checks operator-managed links and logical path exclusions.
func TestWebDAVFileSystemLinks(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".config", "originals")
	require.NoError(t, fs.MkdirAll(filepath.Join(root, ".ssh")))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".ssh/key"), []byte("reserved-control"), fs.ModeFile))
	outside := filepath.Join(t.TempDir(), "original.jpg")
	require.NoError(t, os.WriteFile(outside, []byte("linked-control"), fs.ModeFile))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "ordinary-link")))
	require.NoError(t, os.Symlink(filepath.Join(root, ".ssh/key"), filepath.Join(root, "reserved-link")))
	require.NoError(t, os.Symlink(filepath.Join(root, ".ssh"), filepath.Join(root, "reserved-dir")))
	require.NoError(t, os.Symlink(filepath.Join(root, "missing"), filepath.Join(root, "broken")))
	d := newWebDAVFileSystem(root)
	ctx := context.Background()
	f, err := d.OpenFile(ctx, "ordinary-link", os.O_RDONLY, 0)
	require.NoError(t, err)
	data, err := io.ReadAll(f)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	assert.Equal(t, "linked-control", string(data))
	for _, name := range []string{".ssh/key", ".ssh/new.txt"} {
		_, err := d.Stat(ctx, name)
		assert.Error(t, err, name)
		_, err = d.OpenFile(ctx, name, os.O_CREATE|os.O_WRONLY, fs.ModeFile)
		assert.Error(t, err, name)
	}
	require.NoError(t, d.Mkdir(ctx, "ordinary", fs.ModeDir))
	dir, err := d.OpenFile(ctx, "/", os.O_RDONLY, 0)
	require.NoError(t, err)
	defer dir.Close()
	var names []string
	for {
		entries, err := dir.Readdir(1)
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.NotEmpty(t, entries)
	}
	assert.ElementsMatch(t, []string{"ordinary-link", "ordinary", "reserved-link", "reserved-dir", "broken"}, names)
}

// TestWebDAVMutationBudget checks bounded scans and cancellation before mutation.
func TestWebDAVMutationBudget(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a", "b", "c"} {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte("control"), fs.ModeFile))
	}
	d := newWebDAVFileSystem(root)
	assert.NoError(t, d.checkMutation(context.Background(), "/", 4))
	assert.Error(t, d.checkMutation(context.Background(), "/", 3))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.ErrorIs(t, d.checkMutation(ctx, "/", 4), context.Canceled)
	assert.Equal(t, 100000, webDAVMutationEntries)
	// A canceled mutation leaves the selected directory intact.
	require.NoError(t, fs.MkdirAll(filepath.Join(root, "folder")))
	assert.Error(t, d.RemoveAll(ctx, "folder"))
	assert.DirExists(t, filepath.Join(root, "folder"))
}

// TestWebDAVNoSymlinkCreation checks regular uploads and unsupported link-creation methods.
func TestWebDAVNoSymlinkCreation(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	require.NoError(t, conf.CreateDirectories())
	root := conf.OriginalsPath()
	require.NoError(t, os.WriteFile(filepath.Join(root, "target.txt"), []byte("target-control"), fs.ModeFile))
	router := setupWebDAVRouter(conf)
	base := conf.BaseUri(WebDAVOriginals)

	for _, method := range []string{"LINK", "MKLINK", "BIND"} {
		req := webDAVRequest(method, base+"/link.txt", "target.txt")
		out := httptest.NewRecorder()
		router.ServeHTTP(out, req)
		assert.Contains(t, []int{http.StatusNotFound, http.StatusMethodNotAllowed}, out.Code)
		assert.NoFileExists(t, filepath.Join(root, "link.txt"))
	}

	req := webDAVRequest("PUT", base+"/link.txt", "target.txt")
	out := httptest.NewRecorder()
	router.ServeHTTP(out, req)
	require.Equal(t, http.StatusCreated, out.Code)
	info, err := os.Lstat(filepath.Join(root, "link.txt"))
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular())
	data, err := os.ReadFile(filepath.Join(root, "target.txt")) //nolint:gosec // Test reads its temporary filesystem control.
	require.NoError(t, err)
	assert.Equal(t, "target-control", string(data))
}

// TestWebDAVFileSystemMutations checks the filesystem boundary independently of HTTP preflight.
func TestWebDAVFileSystemMutations(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, fs.MkdirAll(filepath.Join(root, "tree/.config")))
	require.NoError(t, os.WriteFile(filepath.Join(root, "tree/.config/settings"), []byte("reserved-control"), fs.ModeFile))
	d := newWebDAVFileSystem(root)
	ctx := context.Background()
	assert.Error(t, d.RemoveAll(ctx, "tree"))
	assert.Error(t, d.Rename(ctx, "tree", "renamed"))
	assert.Error(t, d.Mkdir(ctx, ".git", fs.ModeDir))
	_, err := d.OpenFile(ctx, "tree/.config/new", os.O_CREATE|os.O_WRONLY, fs.ModeFile)
	assert.Error(t, err)
	require.NoError(t, d.Mkdir(ctx, "ordinary", fs.ModeDir))
	f, err := d.OpenFile(ctx, "ordinary/file", os.O_CREATE|os.O_WRONLY, fs.ModeFile)
	require.NoError(t, err)
	_, err = f.Write([]byte("ordinary-control"))
	require.NoError(t, err)
	require.NoError(t, f.Close())
	assert.Error(t, d.Rename(ctx, "ordinary/file", "tree/.config/replaced"))
	require.NoError(t, d.Rename(ctx, "ordinary/file", "ordinary/moved"))
	info, err := d.Stat(ctx, "ordinary/moved")
	require.NoError(t, err)
	assert.EqualValues(t, len("ordinary-control"), info.Size())
	require.NoError(t, d.RemoveAll(ctx, "ordinary"))
	assert.NoDirExists(t, filepath.Join(root, "ordinary"))
	assert.FileExists(t, filepath.Join(root, "tree/.config/settings"))
	require.NoError(t, os.Symlink("cycle", filepath.Join(root, "cycle")))
	assert.True(t, d.allowed("cycle"))
	_, err = d.Stat(ctx, "cycle")
	assert.Error(t, err)
}

// TestWebDAVYamlWrites checks metadata write authority without reducing permitted reads.
func TestWebDAVYamlWrites(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, fs.MkdirAll(filepath.Join(root, "folder.yml")))
	require.NoError(t, os.WriteFile(filepath.Join(root, "folder.yml/keep.txt"), []byte("destination-control"), fs.ModeFile))
	require.NoError(t, fs.MkdirAll(filepath.Join(root, "album")))
	for name, content := range map[string]string{"note.yml": "yaml-control", "album/photo.yaml": "nested-control", "image.jpg": "image-control", "destination.txt": "destination-control"} {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(content), fs.ModeFile))
	}
	require.NoError(t, os.Symlink(filepath.Join(root, "album"), filepath.Join(root, "album-alias")))
	d := newWebDAVFileSystem(root)
	restricted := context.Background()
	privileged := context.WithValue(restricted, webDAVManagedWriteKey{}, true)
	assert.False(t, canWriteManagedFiles(restricted))
	assert.True(t, canWriteManagedFiles(privileged))
	for _, name := range []string{"note.yml", "new.YAML", "trailing.yml/"} {
		assert.True(t, d.managedFile(restricted, name))
		_, err := d.OpenFile(restricted, name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fs.ModeFile)
		assert.Error(t, err)
	}
	f, err := d.OpenFile(restricted, "note.yml", os.O_RDONLY, 0)
	require.NoError(t, err)
	data, err := io.ReadAll(f)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	assert.Equal(t, "yaml-control", string(data))
	assert.Error(t, d.RemoveAll(restricted, "note.yml"))
	assert.Error(t, d.Rename(restricted, "image.jpg", "promoted.yml"))
	assert.Error(t, d.Rename(restricted, "album", "moved"))
	for _, tc := range []struct {
		method, source, dest string
		status               int
	}{
		{"PUT", "/note.yml", "", 403},
		{"PUT", "/trailing.yaml/", "", 403},
		{"MOVE", "/image.jpg", "/folder.yml", 403},
		{"COPY", "/image.jpg", "/folder.yml", 403},
		{"MOVE", "/image.jpg", "/trailing.yaml/", 403},
		{"COPY", "/image.jpg", "/trailing.yaml/", 403},
		{"MOVE", "/album-alias", "/link.yml", 403}, {"MOVE", "/image.jpg", "/promoted.yaml", 403},
		{"MOVE", "/album", "/destination.txt", 403}, {"COPY", "/album", "/destination.txt", 403},
		{"COPY", "/album-alias", "/destination.txt", 403},
		{"COPY", "/note.yml", "/ordinary.txt", 200},
	} {
		req := httptest.NewRequest(tc.method, "/originals"+tc.source, nil)
		req.Header.Set("Destination", "/originals"+tc.dest)
		req.Header.Set("Overwrite", "T")
		assert.Equal(t, tc.status, d.preflight(req, "/originals"), tc)
	}
	copyEmpty := httptest.NewRequest("COPY", "/originals/album", nil)
	copyEmpty.Header.Set("Depth", "0")
	copyEmpty.Header.Set("Destination", "/originals/empty")
	assert.Equal(t, http.StatusOK, d.preflight(copyEmpty, "/originals"))
	copyHandler := &webdav.Handler{Prefix: "/originals", FileSystem: d, LockSystem: webdav.NewMemLS()}
	copyResponse := httptest.NewRecorder()
	copyHandler.ServeHTTP(copyResponse, copyEmpty)
	require.Equal(t, http.StatusCreated, copyResponse.Code)
	children, err := os.ReadDir(filepath.Join(root, "empty"))
	require.NoError(t, err)
	assert.Empty(t, children)
	data, err = os.ReadFile(filepath.Join(root, "folder.yml/keep.txt")) //nolint:gosec // Test reads its temporary filesystem control.
	require.NoError(t, err)
	assert.Equal(t, "destination-control", string(data))
	data, err = os.ReadFile(filepath.Join(root, "destination.txt")) //nolint:gosec // Test reads its temporary filesystem control.
	require.NoError(t, err)
	assert.Equal(t, "destination-control", string(data))
	require.NoError(t, d.Mkdir(restricted, "folder.yaml", fs.ModeDir))
	assert.False(t, d.managedFile(restricted, "folder.yaml"))
	f, err = d.OpenFile(privileged, "new.yaml", os.O_CREATE|os.O_WRONLY, fs.ModeFile)
	require.NoError(t, err)
	_, err = f.Write([]byte("privileged-control"))
	require.NoError(t, err)
	require.NoError(t, f.Close())
	require.NoError(t, d.Rename(privileged, "new.yaml", "renamed.yml"))
	require.NoError(t, d.RemoveAll(privileged, "renamed.yml"))
}

// TestManagedFileName checks file types and exact ignore-configuration names.
func TestManagedFileName(t *testing.T) {
	for _, name := range []string{"photo.yml", "photo.YAML", ".ppignore", "album/.PPIGNORE", ".dockerignore", ".gitignore", ".rsyncignore", ".ignore"} {
		assert.True(t, managedFileName(name), name)
	}
	for _, name := range []string{"photo.json", "photo.txt", "ppignore", ".ppignore.txt", ".ppignore-backup"} {
		assert.False(t, managedFileName(name), name)
	}
}

// TestWebDAVIgnoreWrites preserves ignore settings across direct and recursive mutations.
func TestWebDAVIgnoreWrites(t *testing.T) {
	for _, tc := range []struct{ label, name string }{
		{"PhotoPrism", ".ppignore"}, {"Git", ".gitignore"}, {"Docker", ".dockerignore"},
		{"Rsync", ".rsyncignore"}, {"Generic", ".ignore"},
	} {
		t.Run(tc.label, func(t *testing.T) {
			ignoreName := tc.name
			root := t.TempDir()
			for _, dir := range []string{"album", "target/" + ignoreName, "empty"} {
				require.NoError(t, fs.MkdirAll(filepath.Join(root, dir)))
			}
			for _, name := range []string{ignoreName, "album/" + ignoreName, "source.txt", "target/" + ignoreName + "/keep.txt"} {
				require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte("original-control"), fs.ModeFile))
			}
			require.NoError(t, os.Symlink(filepath.Join(root, "album"), filepath.Join(root, "album-alias")))
			d := newWebDAVFileSystem(root)
			restricted := context.Background()
			privileged := context.WithValue(restricted, webDAVManagedWriteKey{}, true)
			for _, name := range []string{ignoreName, "new/" + strings.ToUpper(ignoreName), "trailing/" + ignoreName + "/"} {
				assert.True(t, d.managedFile(restricted, name), name)
				_, err := d.OpenFile(restricted, name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fs.ModeFile)
				assert.ErrorIs(t, err, os.ErrPermission, name)
			}
			assert.ErrorIs(t, d.Rename(restricted, "source.txt", "empty/"+ignoreName), os.ErrPermission)
			assert.ErrorIs(t, d.RemoveAll(restricted, "album"), os.ErrPermission)
			handler := &webdav.Handler{Prefix: "/originals", FileSystem: d, LockSystem: webdav.NewMemLS()}
			for _, tc := range []struct{ method, source, destination string }{
				{"PUT", "/" + ignoreName, ""}, {"DELETE", "/" + ignoreName, ""},
				{"MOVE", "/" + ignoreName, "/moved.txt"},
				{"MOVE", "/source.txt", "/empty/" + ignoreName}, {"COPY", "/source.txt", "/target/" + ignoreName},
				{"COPY", "/source.txt", "/empty/" + ignoreName + "/"}, {"MOVE", "/source.txt", "/" + ignoreName},
				{"DELETE", "/album", ""}, {"MOVE", "/album", "/moved"},
				{"COPY", "/album", "/copied"}, {"COPY", "/album-alias", "/copied"},
				{"COPY", "/empty", "/album"}, {"MOVE", "/empty", "/album"},
			} {
				req := httptest.NewRequest(tc.method, "/originals"+tc.source, nil)
				req.Header.Set("Destination", "/originals"+tc.destination)
				req.Header.Set("Overwrite", "T")
				assert.Equal(t, http.StatusForbidden, d.preflight(req, "/originals"), tc)
			}
			for _, name := range []string{ignoreName, "album/" + ignoreName, "target/" + ignoreName + "/keep.txt"} {
				req := httptest.NewRequest("GET", "/originals/"+name, nil)
				require.Equal(t, http.StatusOK, d.preflight(req, "/originals"))
				out := httptest.NewRecorder()
				handler.ServeHTTP(out, req)
				require.Equal(t, http.StatusOK, out.Code)
				assert.Equal(t, "original-control", out.Body.String())
			}
			f, err := d.OpenFile(privileged, ignoreName, os.O_WRONLY|os.O_TRUNC, fs.ModeFile)
			require.NoError(t, err)
			_, err = f.Write([]byte("updated-control"))
			require.NoError(t, err)
			require.NoError(t, f.Close())
			require.NoError(t, d.Rename(privileged, "source.txt", "empty/"+ignoreName))
			require.NoError(t, d.RemoveAll(privileged, "album"))
			data, err := os.ReadFile(filepath.Join(root, ignoreName)) //nolint:gosec // Test reads its temporary filesystem control.
			require.NoError(t, err)
			assert.Equal(t, "updated-control", string(data))
		})
	}
}

// TestWebDAVFileSystemRelativeRoot checks consistent native roots for relative mounts.
func TestWebDAVFileSystemRelativeRoot(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, dir := range []string{"", ".", "relative"} {
		t.Run(fmt.Sprintf("Root%q", dir), func(t *testing.T) {
			if dir != "" {
				require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
			}
			d := newWebDAVFileSystem(dir)
			assert.Equal(t, string(d.base), d.root)
			name := filepath.Join(d.root, "control.txt")
			require.NoError(t, os.WriteFile(name, []byte("control"), fs.ModeFile))
			require.NoError(t, d.Rename(context.Background(), "/control.txt", "/renamed.txt"))
			assert.FileExists(t, filepath.Join(d.root, "renamed.txt"))
			require.NoError(t, d.RemoveAll(context.Background(), "/renamed.txt"))
		})
	}
}
