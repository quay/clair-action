package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/quay/claircore"
	"github.com/quay/claircore/libvuln"
	"github.com/quay/claircore/libvuln/driver"
	"github.com/quay/claircore/rhel/vex"
	_ "github.com/quay/claircore/updater/defaults"
	"github.com/urfave/cli/v2"

	"github.com/quay/clair-action/datastore"
)

var updateCmd = &cli.Command{
	Name:    "update",
	Aliases: []string{"u"},
	Usage:   "update the database",
	Action:  update,
	Flags: []cli.Flag{
		&cli.PathFlag{
			Name:    "db-path",
			Value:   "",
			Usage:   "where to look for the matcher DB",
			EnvVars: []string{"DB_PATH"},
		},
		&cli.DurationFlag{
			Name:    "http-timeout",
			Value:   2 * time.Minute,
			Usage:   "the timeout for HTTP requests",
			EnvVars: []string{"HTTP_TIMEOUT"},
		},
	},
}

func update(c *cli.Context) error {
	ctx := c.Context
	dbPath := c.String("db-path")
	httpTimeout := c.Duration("http-timeout")
	matcherStore, err := datastore.NewSQLiteMatcherStore(dbPath, true)
	if err != nil {
		return fmt.Errorf("error creating sqlite backend: %v", err)
	}

	cl := &http.Client{
		Timeout: httpTimeout,
	}
	factoryConfigs := make(map[string]driver.ConfigUnmarshaler)
	factoryConfigs["rhel-vex"] = func(v interface{}) error {
		cfg := v.(*vex.FactoryConfig)
		cfg.CompressedFileTimeout = claircore.Duration(httpTimeout)
		return nil
	}

	matcherOpts := &libvuln.Options{
		Client:                   cl,
		Store:                    matcherStore,
		Locker:                   NewLocalLockSource(),
		DisableBackgroundUpdates: true,
		UpdateWorkers:            1,
		UpdaterConfigs:           factoryConfigs,
	}

	lv, err := libvuln.New(ctx, matcherOpts)
	if err != nil {
		return fmt.Errorf("error creating Libvuln: %v", err)
	}

	if err := lv.FetchUpdates(ctx); err != nil {
		return fmt.Errorf("error updating vulnerabilities: %v", err)
	}
	return nil
}
