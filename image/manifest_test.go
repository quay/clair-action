package image

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

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
		name := fe.Name
		data := fe.Data
		if strings.HasSuffix(name, ".b64") {
			decoded, err := base64.StdEncoding.DecodeString(string(data))
			if err != nil {
				t.Fatalf("base64 decode %s: %v", name, err)
			}
			data = decoded
			name = strings.TrimSuffix(name, ".b64")
		}

		h := &tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(data)),
		}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatalf("write header %s: %v", name, err)
		}
		if _, err := io.Copy(tw, bytes.NewReader(data)); err != nil {
			t.Fatalf("write contents %s: %v", name, err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	return tmpTar
}

func TestLocalManifest(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		txtarRelPath     string
		wantManifestHash string
		wantLayerHash    string
	}{
		{
			name:             "docker save",
			txtarRelPath:     "testdata/docker_save.txtar",
			wantManifestHash: "sha256:c9b978d8d0fa53a27117f46b2e17ce906a9de863df82d7709e73868a4932f750",
			wantLayerHash:    "sha256:e7328e803158cca63d8efdbe1caefb1b51654de77e5fa8691079ad06db1abf75",
		},
		{
			name:             "podman save",
			txtarRelPath:     "testdata/podman_save.txtar",
			wantManifestHash: "sha256:869d3637f2f9b10c265e4bab4b0eccfe8770520e6e903e7dd8acf33b4987bfc1",
			wantLayerHash:    "sha256:3c6d585e6a72780f0632d16bb8bfd98dfc35b403a11f5cd61925ec31643a76d3",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			exportTar := writeTarFromTxtar(t, tt.txtarRelPath)
			importDir := t.TempDir()
			ctx := context.Background()
			m, err := ManifestFromLocal(ctx, exportTar, importDir)
			if err != nil {
				t.Fatalf("InspectLocal error: %v", err)
			}
			if m.Hash.String() != tt.wantManifestHash {
				t.Fatalf("manifest hash = %s, want %s", m.Hash.String(), tt.wantManifestHash)
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
		})
	}
}
