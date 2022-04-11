package libvuln

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/quay/zlog"
	"github.com/remind101/migrate"

	"github.com/quay/claircore/datastore"
	"github.com/quay/claircore/datastore/postgres/migrations"
	"github.com/quay/claircore/libvuln/driver"
	"github.com/quay/claircore/pkg/poolstats"
)

const (
	DefaultUpdateInterval  = 30 * time.Minute
	DefaultUpdateWorkers   = 10
	DefaultMaxConnPool     = 50
	DefaultUpdateRetention = 2
)

type Options struct {
	Store  datastore.MatcherStore
	Locker LockSource
	// The maximum number of database connections in the
	// connection pool.
	MaxConnPool int32
	// A connection string to the database Libvuln will use.
	//
	// TODO(hank) This should be a factory function so the data store can be
	// a clean abstraction.
	ConnString string
	// An interval on which Libvuln will check for new security database
	// updates.
	//
	// This duration will have jitter added to it, to help with smearing load on
	// installations.
	UpdateInterval time.Duration
	// Determines if Libvuln will manage database migrations
	Migrations bool
	// A slice of strings representing which updaters libvuln will create.
	//
	// If nil all default UpdaterSets will be used.
	//
	// The following sets are supported:
	// "alpine"
	// "aws"
	// "debian"
	// "oracle"
	// "photon"
	// "pyupio"
	// "rhel"
	// "suse"
	// "ubuntu"
	UpdaterSets []string
	// A list of out-of-tree updaters to run.
	//
	// This list will be merged with any defined UpdaterSets.
	//
	// If you desire no updaters to run do not add an updater
	// into this slice.
	Updaters []driver.Updater
	// A slice of strings representing which
	// matchers will be used.
	//
	// If nil all default Matchers will be used
	//
	// The following names are supported by default:
	// "alpine"
	// "aws"
	// "debian"
	// "oracle"
	// "photon"
	// "python"
	// "rhel"
	// "suse"
	// "ubuntu"
	// "crda" - remotematcher calls hosted api via RPC.
	MatcherNames []string

	// Config holds configuration blocks for MatcherFactories and Matchers,
	// keyed by name.
	MatcherConfigs map[string]driver.MatcherConfigUnmarshaler

	// A list of out-of-tree matchers you'd like libvuln to
	// use.
	//
	// This list will me merged with the default matchers.
	Matchers []driver.Matcher

	// Enrichers is a slice of enrichers to use with all VulnerabilityReport
	// requests.
	Enrichers []driver.Enricher

	// UpdateWorkers controls the number of update workers running concurrently.
	// If less than or equal to zero, a sensible default will be used.
	UpdateWorkers int

	// UpdateRetention controls the number of updates to retain between
	// garbage collection periods.
	//
	// The lowest possible value is 2 in order to compare updates for notification
	// purposes.
	UpdateRetention int

	// If set to true, there will not be a goroutine launched to periodically
	// run updaters.
	DisableBackgroundUpdates bool

	// UpdaterConfigs is a map of functions for configuration of Updaters.
	UpdaterConfigs map[string]driver.ConfigUnmarshaler

	// Client is an http.Client for use by all updaters. If unset,
	// http.DefaultClient will be used.
	Client *http.Client
}

// Pool creates and returns a configured pxgpool.Pool.
func (o *Options) pool(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(o.ConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ConnString: %v", err)
	}
	cfg.MaxConns = o.MaxConnPool
	const appnameKey = `application_name`
	params := cfg.ConnConfig.RuntimeParams
	if _, ok := params[appnameKey]; !ok {
		params[appnameKey] = `libvuln`
	}

	pool, err := pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Pool: %v", err)
	}

	if err := prometheus.Register(poolstats.NewCollector(pool, "libvuln")); err != nil {
		zlog.Info(ctx).Msg("pool metrics already registered")
	}

	return pool, nil
}

// Migrations performs migrations if the configuration asks for it.
func (o *Options) migrations(_ context.Context) error {
	// The migrate package doesn't use the context, which is... disconcerting.
	if !o.Migrations {
		return nil
	}
	cfg, err := pgx.ParseConfig(o.ConnString)
	if err != nil {
		return err
	}
	db, err := sql.Open("pgx", stdlib.RegisterConnConfig(cfg))
	if err != nil {
		return err
	}
	defer db.Close()

	migrator := migrate.NewPostgresMigrator(db)
	migrator.Table = migrations.MatcherMigrationTable
	if err := migrator.Exec(migrate.Up, migrations.MatcherMigrations...); err != nil {
		return err
	}

	return nil
}
