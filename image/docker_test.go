package image

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/txtar"
)

func setup(resultsDir string) {
	os.Mkdir(resultsDir, 0700)
}

func teardown(resultsDir string) {
	os.RemoveAll(resultsDir)
}

func TestFromExported(t *testing.T) {
	resultsDir := "testdata/results"
	setup(resultsDir)
	defer teardown(resultsDir)
	ctx := context.TODO()
	di, err := NewDockerLocalImage(ctx, "testdata/algo", resultsDir)
	if err != nil {
		t.Fatalf("got error %v", err)
	}
}

// writeTarFromTxtar converts a txtar-like archive into a tar file on disk.
func writeTarFromTxtar(t *testing.T, txtarPath string) string {
	t.Helper()
	b, err := os.ReadFile(txtarPath)
	if err != nil {
		t.Fatalf("read txtar: %v", err)
	}
	ar := txtar.Parse(b)

	tmpTar := filepath.Join(t.TempDir(), "image-save.tar")
	tf, err := os.Create(tmpTar)
	if err != nil {
		t.Fatalf("create tar: %v", err)
	}
	defer tf.Close()
	tw := tar.NewWriter(tf)
	for _, fe := range ar.Files {
		if fe.Name == "" {
			t.Fatalf("empty file name in txtar")
		}
		h := &tar.Header{
			Name: fe.Name,
			Mode: 0600,
			Size: int64(len(fe.Data)),
		}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatalf("write header %s: %v", fe.Name, err)
		}
		if _, err := io.Copy(tw, bytes.NewReader(fe.Data)); err != nil {
			t.Fatalf("write contents %s: %v", fe.Name, err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	return tmpTar
}

func TestNewDockerLocalImage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		txtarRelPath   string
		wantConfigHash string
		wantLayerHash  string
	}{
		{
			name:           "docker save",
			txtarRelPath:   "testdata/docker_save.txtar",
			wantConfigHash: "sha256:dfd20f40ba9083739c91799f08f4bd57b1c5c840583f9ea6045b8f9cbd3bb539",
			wantLayerHash:  "sha256:e7328e803158cca63d8efdbe1caefb1b51654de77e5fa8691079ad06db1abf75",
		},
		{
			name:           "podman save",
			txtarRelPath:   "testdata/podman_save.txtar",
			wantConfigHash: "sha256:dfd20f40ba9083739c91799f08f4bd57b1c5c840583f9ea6045b8f9cbd3bb539",
			wantLayerHash:  "sha256:e7328e803158cca63d8efdbe1caefb1b51654de77e5fa8691079ad06db1abf75",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			exportTar := writeTarFromTxtar(t, tt.txtarRelPath)
			importDir := t.TempDir()
			ctx := context.Background()
			di, err := NewDockerLocalImage(ctx, exportTar, importDir)
			if err != nil {
				t.Fatalf("NewDockerLocalImage error: %v", err)
			}
			m, err := di.GetManifest(ctx)
			if err != nil {
				t.Fatalf("GetManifest error: %v", err)
			}
			if m.Hash.String() != tt.wantConfigHash {
				t.Fatalf("manifest hash = %s, want %s", m.Hash.String(), tt.wantConfigHash)
			}
			if len(m.Layers) == 0 {
				t.Fatalf("no layers parsed")
			}
			found := false
			for _, l := range m.Layers {
				if l.Hash.String() == tt.wantLayerHash {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("expected layer hash %s not found in layers", tt.wantLayerHash)
			}
			// Ensure the expected layer blob was written into importDir.
			if _, err := os.Stat(filepath.Join(importDir, tt.wantLayerHash)); err != nil {
				t.Fatalf("expected layer file %s not found in importDir: %v", tt.wantLayerHash, err)
			}
		})
	}
}
