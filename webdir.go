package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/zncoder/check"
	"github.com/zncoder/mygo"
	"golang.org/x/net/webdav"
)

func main() {
	port := flag.Int("p", 5556, "port")
	rw := flag.Bool("rw", false, "read-write mode")
	mygo.ParseFlag("dir")

	dir := flag.Arg(0)
	dir = check.V(filepath.Abs(dir)).F("abs", "dir", dir)
	check.T(mygo.IsDir(dir)).F("not a directory", "dir", dir)

	var fsys webdav.FileSystem
	if *rw {
		fsys = webdav.Dir(dir)
	} else {
		fsys = ReadonlyDir{webdav.Dir(dir)}
	}

	handler := &webdav.Handler{
		FileSystem: fsys,
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			check.E(err).I(fs.ErrNotExist).L("webdav err", "url", r.URL)
		},
	}

	http.Handle("/", handler)

	check.L("Serving", "dir", dir, "port", *port)
	check.E(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)).F("listen", "port", *port)
}

type ReadonlyDir struct {
	webdav.Dir
}

func (d ReadonlyDir) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return fs.ErrPermission
}

func (d ReadonlyDir) RemoveAll(ctx context.Context, name string) error {
	return fs.ErrPermission
}

func (d ReadonlyDir) Rename(ctx context.Context, oldName, newName string) error {
	return fs.ErrPermission
}

func (d ReadonlyDir) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	// TODO: make ReadonlyDir appears as a read-only file system
	// if flag&(os.O_RDONLY|os.O_RDWR) != 0 {
	// 	flag = os.O_RDONLY
	// } else {
	// 	return nil, fs.ErrPermission
	// }
	return d.Dir.OpenFile(ctx, name, flag, perm)
}
