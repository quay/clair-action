package image

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/quay/claircore"
)

var _ Image = (*dockerLocalImage)(nil)

type Image interface {
	GetManifest(context.Context) (*claircore.Manifest, error)
}

type imageInfo struct {
	Config string   `json:"Config"`
	Layers []string `json:"Layers"`
}

type dockerLocalImage struct {
	imageDigest string
	layers      []*claircore.Layer
}

func NewDockerLocalImage(ctx context.Context, exportDir string, importDir string) (*dockerLocalImage, error) {
	f, err := os.Open(exportDir)
	if err != nil {
		return nil, fmt.Errorf("unable to open tar: %w", err)
	}

	di := &dockerLocalImage{}
	m := &imageInfo{}

	tr := tar.NewReader(f)
	hdr, err := tr.Next()
	for ; err == nil; hdr, err = tr.Next() {
		_, fn := filepath.Split(hdr.Name)
		if fn == "manifest.json" {
			_m := []*imageInfo{}
			b, err := io.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(b, &_m)
			if err != nil {
				return nil, err
			}
			m = _m[0]
			digest := strings.TrimSuffix(filepath.Base(m.Config), filepath.Ext(m.Config))
			di.imageDigest = "sha256:" + digest
			continue
		}
	}
	// Rewind and find the layers defined in the manifest.json
	f.Seek(0, io.SeekStart)
	tr = tar.NewReader(f)
	// This is done to ensure the layers are in the order defined in the manifest.json
	di.layers = make([]*claircore.Layer, len(m.Layers))
	for {
		hdr, err = tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		start, _ := f.Seek(0, io.SeekCurrent)
		for i, l := range m.Layers {
			if hdr.Name == l {
				ra := io.NewSectionReader(f, start, hdr.Size)
				fName := strings.TrimSuffix(filepath.Base(l), filepath.Ext(l))
				hash, err := claircore.ParseDigest("sha256:" + fName)
				if err != nil {
					return nil, err
				}
				l := &claircore.Layer{
					Hash: hash,
				}
				l.Init(context.Background(), &claircore.LayerDescription{
					Digest:    "sha256:" + fName,
					MediaType: "application/vnd.oci.image.layer.v1.tar",
				}, ra)
				di.layers[i] = l
			}
		}
	}

	return di, nil
}

func (i *dockerLocalImage) GetManifest(_ context.Context) (*claircore.Manifest, error) {
	digest, err := claircore.ParseDigest(i.imageDigest)
	if err != nil {
		return nil, err
	}

	return &claircore.Manifest{
		Hash:   digest,
		Layers: i.layers,
	}, nil
}

type dockerRemoteImage struct {
	ref string
}

func NewDockerRemoteImage(ctx context.Context, imgRef string) *dockerRemoteImage {
	return &dockerRemoteImage{ref: imgRef}
}

func (i *dockerRemoteImage) GetManifest(ctx context.Context) (*claircore.Manifest, error) {
	return Inspect(ctx, i.ref)
}
