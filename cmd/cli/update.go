package main

import (
	"fmt"
	"net/http"

	"github.com/quay/claircore/libvuln"
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
	},
}

func update(c *cli.Context) error {
	ctx := c.Context
	dbPath := c.String("db-path")
	matcherStore, err := datastore.NewSQLiteMatcherStore(dbPath, true)
	if err != nil {
		return fmt.Errorf("error creating sqlite backend: %v", err)
	}

	cl := &http.Client{}

	matcherOpts := &libvuln.Options{
		Client:                   cl,
		Store:                    matcherStore,
		Locker:                   NewLocalLockSource(),
		DisableBackgroundUpdates: true,
		UpdateWorkers:            1,
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
