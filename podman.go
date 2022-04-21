package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/quay/claircore"
)

type Image interface {
	GetManifest() (claircore.Manifest, error)
}

var _ Image = (*podmanLocalImage)(nil)

type index struct {
	Manifests []*manifest `json:"manifests"`
}

type manifest struct {
	Digest string `json:"digest"`
}

type podmanLocalImage struct {
	imageDigest string
	layerPaths  []string
}

func NewPodmanLocalImage(ctx context.Context, exportDir string) (*podmanLocalImage, error) {
	f, err := os.Open(exportDir)
	if err != nil {
		return nil, fmt.Errorf("claircore: unable to open tar: %w", err)
	}

	pi := &podmanLocalImage{}

	tr := tar.NewReader(f)
	hdr, err := tr.Next()
	for ; err == nil; hdr, err = tr.Next() {
		_, fn := filepath.Split(hdr.Name)
		if strings.HasPrefix(hdr.Name, "blobs/sha256/") && hdr.Typeflag == tar.TypeReg {
			peekBytes := make([]byte, 1)
			_, err := tr.Read(peekBytes)
			if err != nil {
				return nil, err
			}
			if string(peekBytes) == "{" {
				continue
			}

			b := bytes.NewBuffer(peekBytes)
			_, err = io.Copy(b, tr)
			if err != nil {
				return nil, err
			}

			gr, err := gzip.NewReader(b)
			if err != nil {
				return nil, err
			}

			_, fn := filepath.Split(hdr.Name)
			layerFile, err := os.OpenFile("/tmp/sha256:"+fn, os.O_CREATE|os.O_RDWR, os.FileMode(0600))
			if err != nil {
				return nil, err
			}

			_, err = io.Copy(layerFile, gr)
			if err != nil {
				return nil, err
			}

			pi.layerPaths = append(pi.layerPaths, layerFile.Name())
			layerFile.Close()
		}
		if fn == "index.json" {
			i := &index{}
			b, err := io.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(b, i)
			if err != nil {
				return nil, err
			}
			pi.imageDigest = i.Manifests[0].Digest
		}
	}
	return pi, nil
}

func (i *podmanLocalImage) getLayers() ([]*claircore.Layer, error) {
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

func (i *podmanLocalImage) GetManifest() (claircore.Manifest, error) {
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
