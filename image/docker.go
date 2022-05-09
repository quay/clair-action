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
	layerPaths  []string
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
		dir, fn := filepath.Split(hdr.Name)
		if fn == "layer.tar" {

			sha := filepath.Base(dir)

			layerFilePath := filepath.Join(importDir, "sha256:"+sha)
			layerFile, err := os.OpenFile(layerFilePath, os.O_CREATE|os.O_RDWR, os.FileMode(0600))
			if err != nil {
				return nil, err
			}

			_, err = io.Copy(layerFile, tr)
			if err != nil {
				return nil, err
			}

			di.layerPaths = append(di.layerPaths, layerFile.Name())
			layerFile.Close()
		}
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
			digest := strings.TrimSuffix(m.Config, filepath.Ext(m.Config))
			di.imageDigest = "sha256:" + digest
		}
	}

	var sortedPaths []string
	for _, p := range m.Layers {
		for _, l := range di.layerPaths {
			if filepath.Dir(p) == strings.TrimPrefix(filepath.Base(l), "sha256:") {
				sortedPaths = append(sortedPaths, l)
			}
		}
	}
	di.layerPaths = sortedPaths
	return di, nil
}

func (i *dockerLocalImage) getLayers() ([]*claircore.Layer, error) {
	if len(i.layerPaths) == 0 {
		return nil, nil
	}
	layers := []*claircore.Layer{}
	for _, layerStr := range i.layerPaths {

		_, d := filepath.Split(layerStr)
		hash, err := claircore.ParseDigest(d)
		if err != nil {
			return nil, err
		}
		l := &claircore.Layer{
			Hash: hash,
			URI:  layerStr,
		}
		layers = append(layers, l)
	}
	return layers, nil
}

func (i *dockerLocalImage) GetManifest(_ context.Context) (*claircore.Manifest, error) {
	digest, err := claircore.ParseDigest(i.imageDigest)
	if err != nil {
		return nil, err
	}

	layers, err := i.getLayers()
	if err != nil {
		return nil, err
	}

	return &claircore.Manifest{
		Hash:   digest,
		Layers: layers,
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
