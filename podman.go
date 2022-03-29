package main

import (
	"context"
	"os"
	"path"
	"strings"

	"github.com/quay/claircore"

	"github.com/containers/podman/v4/pkg/bindings"
	"github.com/containers/podman/v4/pkg/bindings/images"
	"github.com/containers/podman/v4/pkg/domain/entities"
)

type Image interface {
	GetLayers() ([]*claircore.Layer, error)
	GetImageDigest() (claircore.Digest, error)
}

var _ Image = (*podmanImage)(nil)

type podmanImage struct {
	info *entities.ImageInspectReport
}

func NewPodmanImage(ctx context.Context, imageTag string) (*podmanImage, error) {
	sockDir := os.Getenv("XDG_RUNTIME_DIR")
	socket := "unix:" + sockDir + "/podman/podman.sock"
	connText, err := bindings.NewConnection(ctx, socket)
	if err != nil {
		return nil, err
	}
	info, err := images.GetImage(connText, imageTag, &images.GetOptions{})
	if err != nil {
		return nil, err
	}
	return &podmanImage{info}, nil
}

func (i *podmanImage) GetLayers() ([]*claircore.Layer, error) {
	if i.info == nil {
		return nil, nil
	}
	layers := []*claircore.Layer{}
	upperDir := i.info.GraphDriver.Data["UpperDir"]
	ud := path.Dir(path.Dir(upperDir))
	for _, layerStr := range i.info.RootFS.Layers {
		hash, err := claircore.ParseDigest(layerStr.String())
		if err != nil {
			return nil, err
		}
		check := strings.Split(layerStr.String(), ":")[1]
		l := &claircore.Layer{
			Hash: hash,
			URI:  path.Join(ud, check, "diff"),
		}
		layers = append(layers, l)
	}
	return layers, nil
}

func (i *podmanImage) GetImageDigest() (claircore.Digest, error) {
	return claircore.ParseDigest(string(i.info.Digest))
}

func (i *podmanImage) GetManifest() (claircore.Manifest, error) {
	digest, err := i.GetImageDigest()
	if err != nil {
		return claircore.Manifest{}, err
	}

	layers, err := i.GetLayers()
	if err != nil {
		return claircore.Manifest{}, err
	}

	return claircore.Manifest{
		Hash:   digest,
		Layers: layers,
	}, nil
}
