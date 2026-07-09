package image

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/quay/claircore/libindex"
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
			defer func() {
				for _, l := range m.Layers {
					if l != nil {
						l.Close()
					}
				}
			}()
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

func TestRemoteManifestStableURIs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	srv := httptest.NewServer(registry.New())
	defer srv.Close()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	img, err := random.Image(1024, 3)
	if err != nil {
		t.Fatal(err)
	}
	refStr := u.Host + "/test/repo:latest"
	ref, err := name.ParseReference(refStr)
	if err != nil {
		t.Fatal(err)
	}
	if err := remote.Write(ref, img); err != nil {
		t.Fatal(err)
	}

	cl, err := NewRegistryClient(ctx, refStr)
	if err != nil {
		t.Fatal(err)
	}
	mf, err := ManifestFromRemote(ctx, cl, refStr)
	if err != nil {
		t.Fatal(err)
	}
	if len(mf.Layers) != 3 {
		t.Fatalf("got %d layers, want 3", len(mf.Layers))
	}
	for _, l := range mf.Layers {
		lu, err := url.Parse(l.URI)
		if err != nil {
			t.Fatal(err)
		}
		if lu.Host != u.Host {
			t.Errorf("layer URI host = %q, want %q", lu.Host, u.Host)
		}
		wantPath := "/v2/test/repo/blobs/" + l.Hash.String()
		if lu.Path != wantPath {
			t.Errorf("layer URI path = %q, want %q", lu.Path, wantPath)
		}
		if lu.RawQuery != "" {
			t.Errorf("layer URI has query %q, want none", lu.RawQuery)
		}
		if len(l.Headers) != 0 {
			t.Errorf("layer has captured headers: %v", l.Headers)
		}
	}
}

// TestRealizeThroughRedirect exercises the registry 302s blob GETs to a
// separate "object storage" server, mimicking registries backed by S3-style
// storage. Layer URLs must be resolved at fetch time, and registry
// credentials must not follow the redirect.
// See also: quay/clair-action#275
func TestRealizeThroughRedirect(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	reg := registry.New()

	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			http.Error(w, "unexpected Authorization header", http.StatusBadRequest)
			return
		}
		reg.ServeHTTP(w, r)
	}))
	defer storage.Close()

	regSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/blobs/sha256:") {
			http.Redirect(w, r, storage.URL+r.URL.Path, http.StatusTemporaryRedirect)
			return
		}
		reg.ServeHTTP(w, r)
	}))
	defer regSrv.Close()

	u, err := url.Parse(regSrv.URL)
	if err != nil {
		t.Fatal(err)
	}
	img, err := random.Image(1024, 2)
	if err != nil {
		t.Fatal(err)
	}
	refStr := u.Host + "/test/redirect:latest"
	ref, err := name.ParseReference(refStr)
	if err != nil {
		t.Fatal(err)
	}
	if err := remote.Write(ref, img); err != nil {
		t.Fatal(err)
	}

	cl, err := NewRegistryClient(ctx, refStr)
	if err != nil {
		t.Fatal(err)
	}
	mf, err := ManifestFromRemote(ctx, cl, refStr)
	if err != nil {
		t.Fatal(err)
	}

	fa := libindex.NewRemoteFetchArena(cl, t.TempDir())
	defer fa.Close(ctx)
	rz := fa.Realizer(ctx)
	defer rz.Close()
	if err := rz.Realize(ctx, mf.Layers); err != nil {
		t.Fatalf("realize: %v", err)
	}
	for _, l := range mf.Layers {
		if _, err := l.Reader(); err != nil {
			t.Errorf("layer %v not realized: %v", l.Hash, err)
		}
	}
}
