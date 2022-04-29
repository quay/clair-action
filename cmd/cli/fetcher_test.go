package main

import (
	"context"
	"testing"

	"github.com/quay/claircore"
)

type fetcherTest struct {
	name      string
	digest    string
	layerPath string
}

func TestFetcher(t *testing.T) {
	tt := []*fetcherTest{
		{
			name:      "simple",
			digest:    "sha256:aaa35f7177b4a8621e5bf2058f04b5fa1105e4965fbd3e0bbc6a63039847caaa",
			layerPath: "/home/crozzy/.go/src/github.com/crozzy/local-clair/testdata/diff/",
		},
		{
			name:      "complex",
			digest:    "sha256:2d035f7177b4a8621e5bf2058f04b5fa1105e4965fbd3e0bbc6a63039847ccbe",
			layerPath: "/home/crozzy/.local/share/containers/storage/overlay/7699752e6ed63eef234d2736d4e37159a433e18e06cd617e254299f324f41797/diff/",
		},
	}

	ctx := context.TODO()
	lf := &LocalFetchArena{}

	for _, tc := range tt {
		d, err := claircore.ParseDigest(tc.digest)
		if err != nil {
			t.Fatalf("error parsing digest: %v", err)
		}
		l := &claircore.Layer{
			Hash: d,
			URI:  tc.layerPath,
		}

		err = lf.Realizer(ctx).Realize(ctx, []*claircore.Layer{l})
		if err != nil {
			t.Fatalf("error fetching layer: %v", err)
		}
	}
}
