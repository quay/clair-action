package main

import (
	"context"
	"errors"

	"github.com/quay/claircore"
	"github.com/quay/claircore/indexer"
)

var (
	_ indexer.FetchArena = (*LocalFetchArena)(nil)
	_ indexer.Realizer   = (*realizer)(nil)
)

type LocalFetchArena struct{}

type LocalFetcher struct{}

// Arena does coordination and global refcounting.
func (*LocalFetchArena) Realizer(context.Context) indexer.Realizer {
	return &realizer{}
}

func (*LocalFetchArena) Close(context.Context) error {
	return nil
}

type realizer struct {
	layers []*claircore.Layer
}

func (r *realizer) Realize(_ context.Context, ls []*claircore.Layer) error {
	r.layers = ls
	return nil
}

func (r *realizer) Close() error {
	errs := make([]error, len(r.layers))
	for i, l := range r.layers {
		errs[i] = l.Close()
	}
	return errors.Join(errs...)
}
