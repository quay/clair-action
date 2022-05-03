package datastore

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func TestDownloadDB(t *testing.T) {
	f, err := os.Create(filepath.Join("testdata/matcher.zst"))
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zstd.NewWriter(f)
	if err != nil {
		t.Fatal(err)
	}

	_, err = zr.Write([]byte("a db"))
	if err != nil {
		t.Fatal(err)
	}
	zr.Close()
	f.Close()

	ctx := context.TODO()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/matcher.zst")
	}))

	err = DownloadDB(ctx, srv.URL, "testdata/matcher")
	if err != nil {
		t.Fatal(err)
	}
	os.Remove("testdata/matcher")
	os.Remove("testdata/matcher.zst")
}
