package datastore

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"sync"

	"github.com/quay/claircore"
	"github.com/quay/claircore/indexer"

	"github.com/google/uuid"
)

type localStore struct {
	pkgMap    map[string][]*claircore.Package
	distroMap map[string][]*claircore.Distribution
	repoMap   map[string][]*claircore.Repository
	fileMap   map[string][]claircore.File
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
			fileMap:   make(map[string][]claircore.File),
		},
	}
}

// DistributionsByLayer implements base method.
func (m *LocalIndexerStore) DistributionsByLayer(_ context.Context, d claircore.Digest, vss indexer.VersionedScanners) ([]*claircore.Distribution, error) {
	m.ls.lock.RLock()
	defer m.ls.lock.RUnlock()
	distros := make([]*claircore.Distribution, 0)
	for _, vs := range vss {
		distros = append(distros, m.ls.distroMap[d.String()+vs.Name()]...)
	}
	return distros, nil
}

// RepositoriesByLayer implements base method.
func (m *LocalIndexerStore) RepositoriesByLayer(_ context.Context, d claircore.Digest, vss indexer.VersionedScanners) ([]*claircore.Repository, error) {
	m.ls.lock.RLock()
	defer m.ls.lock.RUnlock()
	repos := make([]*claircore.Repository, 0)
	for _, vs := range vss {
		repos = append(repos, m.ls.repoMap[d.String()+vs.Name()]...)
	}
	return repos, nil
}

// PackagesByLayer implements base method.
func (m *LocalIndexerStore) PackagesByLayer(_ context.Context, d claircore.Digest, vss indexer.VersionedScanners) ([]*claircore.Package, error) {
	m.ls.lock.RLock()
	defer m.ls.lock.RUnlock()
	pkgs := make([]*claircore.Package, 0)
	for _, vs := range vss {
		pkgs = append(pkgs, m.ls.pkgMap[d.String()+vs.Name()]...)
	}
	return pkgs, nil
}

// FilesByLayer implements base method.
func (m *LocalIndexerStore) FilesByLayer(_ context.Context, d claircore.Digest, vss indexer.VersionedScanners) ([]claircore.File, error) {
	m.ls.lock.RLock()
	defer m.ls.lock.RUnlock()
	files := make([]claircore.File, 0)
	for _, vs := range vss {
		files = append(files, m.ls.fileMap[d.String()+vs.Name()]...)
	}
	return files, nil
}

// IndexDistributions implements base method.
func (m *LocalIndexerStore) IndexDistributions(_ context.Context, distros []*claircore.Distribution, l *claircore.Layer, vs indexer.VersionedScanner) error {
	m.ls.lock.Lock()
	defer m.ls.lock.Unlock()
	for _, d := range distros {
		d.ID = uuid.New().String()
		m.ls.distroMap[l.Hash.String()+vs.Name()] = append(m.ls.distroMap[l.Hash.String()+vs.Name()], d)
	}
	return nil
}

// IndexRepositories implements base method.
func (m *LocalIndexerStore) IndexRepositories(_ context.Context, repos []*claircore.Repository, l *claircore.Layer, vs indexer.VersionedScanner) error {
	m.ls.lock.Lock()
	defer m.ls.lock.Unlock()
	for _, r := range repos {
		r.ID = uuid.New().String()
		m.ls.repoMap[l.Hash.String()+vs.Name()] = append(m.ls.repoMap[l.Hash.String()+vs.Name()], r)
	}
	return nil
}

// IndexPackages implements base method.
func (m *LocalIndexerStore) IndexPackages(_ context.Context, pkgs []*claircore.Package, l *claircore.Layer, vs indexer.VersionedScanner) error {
	m.ls.lock.Lock()
	defer m.ls.lock.Unlock()
	for _, p := range pkgs {
		if p.Source == nil {
			p.Source = &claircore.Package{}
		}

		_, hashBin := md5Package(p)
		p.ID = base64.StdEncoding.EncodeToString([]byte(hashBin))
		m.ls.pkgMap[l.Hash.String()+vs.Name()] = append(m.ls.pkgMap[l.Hash.String()+vs.Name()], p)
	}
	return nil
}

// IndexFiles implements base method.
func (m *LocalIndexerStore) IndexFiles(ctx context.Context, files []claircore.File, l *claircore.Layer, vs indexer.VersionedScanner) error {
	m.ls.lock.Lock()
	defer m.ls.lock.Unlock()
	for _, f := range files {
		m.ls.fileMap[l.Hash.String()+vs.Name()] = append(m.ls.fileMap[l.Hash.String()+vs.Name()], f)
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

func md5Package(p *claircore.Package) (string, []byte) {
	var b bytes.Buffer
	b.WriteString(p.Name)
	b.WriteString(p.Version)
	b.WriteString(p.Kind)
	b.WriteString(p.Module)
	b.WriteString(p.Arch)
	if p.Source != nil {
		b.WriteString(p.Source.Name)
		b.WriteString(p.Source.Version)
		b.WriteString(p.Source.Kind)
		b.WriteString(p.Source.Module)
		b.WriteString(p.Source.Arch)
	}
	s := md5.Sum(b.Bytes())
	return "md5", s[:]
}
