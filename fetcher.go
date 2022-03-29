package main

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/quay/claircore"
	"github.com/quay/claircore/libindex"
)

type LocalFetchArena struct{}

func (*LocalFetchArena) Init(wc *http.Client, root string) {

}

func (n *LocalFetchArena) FetchOne(ctx context.Context, l *claircore.Layer) (do func() error) {
	return func() error {
		var buf bytes.Buffer
		err := compress(l.URI, &buf)
		if err != nil {
			return err
		}
		fileName := "/tmp/" + l.Hash.String()
		fileToWrite, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, os.FileMode(0600))
		if err != nil {
			return err
		}
		if _, err := io.Copy(fileToWrite, &buf); err != nil {
			return err
		}

		l.SetLocal(fileName)
		return nil
	}
}

func (*LocalFetchArena) Forget(digest string) error {
	return os.Remove("/tmp/" + digest)
}

func (n *LocalFetchArena) Fetcher() *libindex.FetchProxy {
	return libindex.NewFetchProxy(n)
}

func (*LocalFetchArena) Close(ctx context.Context) error {
	return nil
}

func compress(src string, buf io.Writer) error {
	tw := tar.NewWriter(buf)

	info, err := os.Stat(src)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(src)
	}

	if err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		var link string
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			if link, err = os.Readlink(path); err != nil {
				return err
			}
		}

		// Try and open the file before writing the header in case permission
		// is denied.
		fh, err := os.Open(path)
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				fmt.Printf("permission denied opening file %v\n", err)
				return nil
			}
			return err
		}
		defer fh.Close()

		header, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}

		header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, src))
		if err = tw.WriteHeader(header); err != nil {
			return err
		}

		if _, err = io.Copy(tw, fh); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}
	return nil
}
