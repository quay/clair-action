package main

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
		return nil, fmt.Errorf("claircore: unable to open tar: %w", err)
	}

	di := &dockerLocalImage{}

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
			m := []*imageInfo{}
			b, err := io.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(b, &m)
			if err != nil {
				return nil, err
			}
			digest := strings.TrimSuffix(m[0].Config, filepath.Ext(m[0].Config))
			di.imageDigest = "sha256:" + digest
		}
	}
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

func (i *dockerLocalImage) GetManifest() (claircore.Manifest, error) {
	digest, err := claircore.ParseDigest(i.imageDigest)
	if err != nil {
		return claircore.Manifest{}, err
	}

	layers, err := i.getLayers()
	if err != nil {
		return claircore.Manifest{}, err
	}

	return claircore.Manifest{
		Hash:   digest,
		Layers: layers,
	}, nil
}
