// This is lifted from Clairctl

package image

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"

	"github.com/quay/claircore"
	"github.com/quay/claircore/pkg/tarfs"
	"github.com/quay/zlog"
)

const (
	userAgent = `clair-action/1`
)

func rt(ctx context.Context, ref string) (http.RoundTripper, error) {
	r, err := name.ParseReference(ref)
	if err != nil {
		return nil, err
	}
	repo := r.Context()

	auth, err := authn.DefaultKeychain.Resolve(repo)
	if err != nil {
		return nil, err
	}
	rt := http.DefaultTransport
	rt = transport.NewUserAgent(rt, userAgent)
	rt = transport.NewRetry(rt)
	rt, err = transport.NewWithContext(ctx, repo.Registry, auth, rt, []string{repo.Scope(transport.PullScope)})
	if err != nil {
		return nil, err
	}
	return rt, nil
}

func ManifestFromRemote(ctx context.Context, r string) (*claircore.Manifest, error) {
	rt, err := rt(ctx, r)
	if err != nil {
		return nil, err
	}

	ref, err := name.ParseReference(r)
	if err != nil {
		return nil, err
	}
	desc, err := remote.Get(ref, remote.WithTransport(rt))
	if err != nil {
		return nil, err
	}
	img, err := desc.Image()
	if err != nil {
		return nil, err
	}
	dig, err := img.Digest()
	if err != nil {
		return nil, err
	}
	ccd, err := claircore.ParseDigest(dig.String())
	if err != nil {
		return nil, err
	}
	out := claircore.Manifest{Hash: ccd}
	zlog.Debug(ctx).
		Str("ref", r).
		Stringer("digest", ccd).
		Msg("found manifest")

	ls, err := img.Layers()
	if err != nil {
		return nil, err
	}
	zlog.Debug(ctx).
		Str("ref", r).
		Int("count", len(ls)).
		Msg("found layers")

	repo := ref.Context()
	rURL := url.URL{
		Scheme: repo.Scheme(),
		Host:   repo.RegistryStr(),
	}
	c := http.Client{
		Transport: rt,
	}

	for _, l := range ls {
		d, err := l.Digest()
		if err != nil {
			return nil, err
		}
		ccd, err := claircore.ParseDigest(d.String())
		if err != nil {
			return nil, err
		}
		u, err := rURL.Parse(path.Join("/", "v2", strings.TrimPrefix(repo.RepositoryStr(), repo.RegistryStr()), "blobs", d.String()))
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Range", "bytes=0-0")
		res, err := c.Do(req)
		if err != nil {
			return nil, err
		}
		res.Body.Close()

		res.Request.Header.Del("User-Agent")
		res.Request.Header.Del("Range")
		out.Layers = append(out.Layers, &claircore.Layer{
			Hash:    ccd,
			URI:     res.Request.URL.String(),
			Headers: res.Request.Header,
		})
	}

	return &out, nil
}

type indexFile struct {
	Manifests []manifestInfo `json:"manifests"`
}

type manifestInfo struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

type manifestFile struct {
	Layers []layerInfo `json:"layers"`
}

type layerInfo struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

func ManifestFromLocal(ctx context.Context, exportDir string, importDir string) (*claircore.Manifest, error) {
	f, err := os.Open(exportDir)
	if err != nil {
		return nil, fmt.Errorf("unable to open tar: %w", err)
	}

	out := &claircore.Manifest{}
	m := &manifestFile{}
	i := &indexFile{}
	fs, err := tarfs.New(f)
	if err != nil {
		return nil, fmt.Errorf("unable to create tarfs: %w", err)
	}
	index, err := fs.Open("index.json")
	if err != nil {
		return nil, fmt.Errorf("unable to open index.json: %w", err)
	}
	defer index.Close()
	b, err := io.ReadAll(index)
	if err != nil {
		return nil, fmt.Errorf("unable to read index.json: %w", err)
	}
	err = json.Unmarshal(b, &i)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal index.json: %w", err)
	}
	manifestDigest := ""
	for _, m := range i.Manifests {
		if m.MediaType == "application/vnd.oci.image.manifest.v1+json" {
			manifestDigest = m.Digest
			break
		}
	}
	if manifestDigest == "" {
		return nil, fmt.Errorf("manifest digest not found")
	}
	md, err := claircore.ParseDigest(manifestDigest)
	if err != nil {
		return nil, fmt.Errorf("unable to parse manifest digest: %w", err)
	}
	out.Hash = md

	mdb := make([]byte, hex.EncodedLen(len(md.Checksum())))
	hex.Encode(mdb, md.Checksum())
	manifestPath := filepath.Join("blobs", md.Algorithm(), string(mdb))
	manifest, err := fs.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open manifest: %w", err)
	}
	defer manifest.Close()

	b, err = io.ReadAll(manifest)
	if err != nil {
		return nil, fmt.Errorf("unable to read manifest: %w", err)
	}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal manifest: %w", err)
	}

	// We have to revert to tar.NewReader() because tarfs.New() doesn't support
	// seeking.
	f.Seek(0, io.SeekStart)
	tr := tar.NewReader(f)
	out.Layers = make([]*claircore.Layer, len(m.Layers))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("unable to read layer: %w", err)
		}
		start, _ := f.Seek(0, io.SeekCurrent)
		for i, l := range m.Layers {
			ld, err := claircore.ParseDigest(l.Digest)
			if err != nil {
				return nil, fmt.Errorf("unable to parse layer digest: %w", err)
			}

			ldb := make([]byte, hex.EncodedLen(len(ld.Checksum())))
			hex.Encode(ldb, ld.Checksum())
			if hdr.Name == filepath.Join("blobs", ld.Algorithm(), string(ldb)) {
				ra := io.NewSectionReader(f, start, hdr.Size)
				var rAt io.ReaderAt
				switch l.MediaType {
				case "application/vnd.oci.image.layer.v1.tar+gzip", "application/vnd.docker.image.rootfs.diff.tar.gzip":
					gr, err := gzip.NewReader(ra)
					if err != nil {
						return nil, fmt.Errorf("unable to create gzip reader: %w", err)
					}
					tmp, err := os.CreateTemp("", "layer-*.tar")
					if err != nil {
						return nil, fmt.Errorf("unable to create temp file: %w", err)
					}
					if _, err := io.Copy(tmp, gr); err != nil {
						return nil, fmt.Errorf("unable to decompress layer: %w", err)
					}
					if err := gr.Close(); err != nil {
						return nil, fmt.Errorf("unable to close gzip reader: %w", err)
					}
					if _, err := tmp.Seek(0, io.SeekStart); err != nil {
						return nil, fmt.Errorf("unable to rewind temp file: %w", err)
					}
					rAt = tmp
				case "application/vnd.oci.image.layer.v1.tar", "application/vnd.docker.image.rootfs.diff.tar":
					// uncompressed tar, use section directly
					rAt = ra
				default:
					return nil, fmt.Errorf("unsupported layer media type: %s", l.MediaType)
				}
				layer := &claircore.Layer{Hash: ld}
				err = layer.Init(context.Background(), &claircore.LayerDescription{
					Digest:    ld.String(),
					MediaType: l.MediaType,
				}, rAt)
				if err != nil {
					return nil, fmt.Errorf("unable to initialize layer: %w", err)
				}
				out.Layers[i] = layer
			}
		}
	}
	return out, nil

}
