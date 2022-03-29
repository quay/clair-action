package main

import (
	"context"
	"sync"

	"github.com/quay/claircore"
	"github.com/quay/claircore/indexer"

	"github.com/google/uuid"
)

type localStore struct {
	pkgMap    map[string][]*claircore.Package
	distroMap map[string][]*claircore.Distribution
	repoMap   map[string][]*claircore.Repository
	lock      *sync.RWMutex
}

var _ indexer.Indexer = (*LocalIndexerStore)(nil)

// LocalIndexerStore is an implementation of indexer.Indexer.
type LocalIndexerStore struct {
	ls *localStore
}

func NewLocalIndexerStore() *LocalIndexerStore {
	return &LocalIndexerStore{
		ls: &localStore{
			lock:      &sync.RWMutex{},
			pkgMap:    make(map[string][]*claircore.Package),
			distroMap: make(map[string][]*claircore.Distribution),
			repoMap:   make(map[string][]*claircore.Repository),
		},
	}
}

// DistributionsByLayer implements base method.
func (m *LocalIndexerStore) DistributionsByLayer(_ context.Context, d claircore.Digest, _ indexer.VersionedScanners) ([]*claircore.Distribution, error) {
	m.ls.lock.RLock()
	defer m.ls.lock.RUnlock()
	distros := m.ls.distroMap[d.String()]
	return distros, nil
}

// RepositoriesByLayer implements base method.
func (m *LocalIndexerStore) RepositoriesByLayer(_ context.Context, d claircore.Digest, _ indexer.VersionedScanners) ([]*claircore.Repository, error) {
	m.ls.lock.RLock()
	defer m.ls.lock.RUnlock()
	repos := m.ls.repoMap[d.String()]
	return repos, nil
}

// PackagesByLayer implements base method.
func (m *LocalIndexerStore) PackagesByLayer(_ context.Context, d claircore.Digest, _ indexer.VersionedScanners) ([]*claircore.Package, error) {
	m.ls.lock.RLock()
	defer m.ls.lock.RUnlock()
	pkgs := m.ls.pkgMap[d.String()]
	return pkgs, nil
}

// IndexDistributions implements base method.
func (m *LocalIndexerStore) IndexDistributions(_ context.Context, distros []*claircore.Distribution, l *claircore.Layer, _ indexer.VersionedScanner) error {
	m.ls.lock.Lock()
	defer m.ls.lock.Unlock()
	for _, d := range distros {
		d.ID = uuid.New().String()
		m.ls.distroMap[l.Hash.String()] = append(m.ls.distroMap[l.Hash.String()], d)
	}
	return nil
}

// IndexRepositories implements base method.
func (m *LocalIndexerStore) IndexRepositories(_ context.Context, repos []*claircore.Repository, l *claircore.Layer, _ indexer.VersionedScanner) error {
	m.ls.lock.Lock()
	defer m.ls.lock.Unlock()
	for _, r := range repos {
		r.ID = uuid.New().String()
		m.ls.repoMap[l.Hash.String()] = append(m.ls.repoMap[l.Hash.String()], r)
	}
	return nil
}

// IndexPackages implements base method.
func (m *LocalIndexerStore) IndexPackages(_ context.Context, pkgs []*claircore.Package, l *claircore.Layer, _ indexer.VersionedScanner) error {
	m.ls.lock.Lock()
	defer m.ls.lock.Unlock()
	for _, p := range pkgs {
		p.ID = uuid.New().String()
		m.ls.pkgMap[l.Hash.String()] = append(m.ls.pkgMap[l.Hash.String()], p)
	}
	return nil
}

// AffectedManifests implements base method.
func (m *LocalIndexerStore) AffectedManifests(_ context.Context, _ claircore.Vulnerability, _ claircore.CheckVulnernableFunc) ([]claircore.Digest, error) {
	return nil, nil
}

// Close implements base method.
func (m *LocalIndexerStore) Close(_ context.Context) error {
	return nil
}

// DeleteManifests implements base method.
func (m *LocalIndexerStore) DeleteManifests(_ context.Context, _ ...claircore.Digest) ([]claircore.Digest, error) {
	return nil, nil
}

// IndexManifest implements base method.
func (m *LocalIndexerStore) IndexManifest(_ context.Context, _ *claircore.IndexReport) error {
	return nil
}

// IndexReport implements base method.
func (m *LocalIndexerStore) IndexReport(_ context.Context, _ claircore.Digest) (*claircore.IndexReport, bool, error) {
	return nil, false, nil
}

// LayerScanned implements base method.
func (m *LocalIndexerStore) LayerScanned(_ context.Context, _ claircore.Digest, _ indexer.VersionedScanner) (bool, error) {
	return false, nil
}

// ManifestScanned implements base method.
func (m *LocalIndexerStore) ManifestScanned(_ context.Context, _ claircore.Digest, _ indexer.VersionedScanners) (bool, error) {
	return false, nil
}

// PersistManifest implements base method.
func (m *LocalIndexerStore) PersistManifest(_ context.Context, _ claircore.Manifest) error {
	return nil
}

// RegisterScanners implements base method.
func (m *LocalIndexerStore) RegisterScanners(_ context.Context, _ indexer.VersionedScanners) error {
	return nil
}

// SetIndexFinished implements base method.
func (m *LocalIndexerStore) SetIndexFinished(_ context.Context, _ *claircore.IndexReport, _ indexer.VersionedScanners) error {
	return nil
}

// SetIndexReport implements base method.
func (m *LocalIndexerStore) SetIndexReport(_ context.Context, _ *claircore.IndexReport) error {
	return nil
}

// SetLayerScanned implements base method.
func (m *LocalIndexerStore) SetLayerScanned(_ context.Context, _ claircore.Digest, _ indexer.VersionedScanner) error {
	return nil
}
