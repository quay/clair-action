package datastore

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/klauspost/compress/zstd"
)

func DownloadDB(ctx context.Context, dbURL string, dest string) error {
	cl := &http.Client{
		Timeout: 10 * time.Minute,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dbURL, nil)
	if err != nil {
		return err
	}
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	default:
		return fmt.Errorf("received status code %d trying to download DB", resp.StatusCode)
	}
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("could not open file to house DB: %v", err)
	}
	defer f.Close()

	gr, err := zstd.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read .zst file: %v", err)
	}
	// TODO(crozzy): do something with this
	_ = resp.Header.Get("x-amz-meta-checksum")

	_, err = io.Copy(f, gr)
	if err != nil {
		return fmt.Errorf("error copying DB to file: %v", err)
	}
	return nil
}
