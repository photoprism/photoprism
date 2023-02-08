package webdav

import (
	"context"
	"golang.org/x/net/webdav"
	"os"
)

type ReadOnlyDir string

func (d ReadOnlyDir) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return os.ErrPermission
}

func (d ReadOnlyDir) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	if flag != os.O_RDONLY {
		return nil, os.ErrPermission
	}
	return webdav.Dir(d).OpenFile(ctx, name, flag, perm)
}

func (d ReadOnlyDir) RemoveAll(ctx context.Context, name string) error {
	return os.ErrPermission
}

func (d ReadOnlyDir) Rename(ctx context.Context, oldName, newName string) error {
	return os.ErrPermission
}

func (d ReadOnlyDir) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return webdav.Dir(d).Stat(ctx, name)
}
