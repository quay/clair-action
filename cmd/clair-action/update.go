package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	clairconfig "github.com/quay/clair/config"
	"github.com/quay/claircore/libvuln"
	"github.com/quay/claircore/libvuln/driver"
	_ "github.com/quay/claircore/updater/defaults"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

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
		&cli.PathFlag{
			Name:    "config",
			Value:   "",
			Usage:   "path to a Clair v4 YAML config file (full format only, drop-in format not supported); the `updaters` section controls updater sets and per-updater configuration",
			EnvVars: []string{"CLAIR_CONFIG"},
		},
	},
}

func update(c *cli.Context) error {
	ctx := c.Context
	dbPath := c.String("db-path")
	configPath := c.String("config")

	matcherStore, err := datastore.NewSQLiteMatcherStore(dbPath, true)
	if err != nil {
		return fmt.Errorf("error creating sqlite backend: %v", err)
	}

	clientTimeout := 10 * time.Minute
	var updaterSets []string
	var updaterConfigs map[string]driver.ConfigUnmarshaler

	if configPath != "" {
		sets, cfgs, err := loadUpdaterOptions(configPath)
		if err != nil {
			return fmt.Errorf("error loading config: %v", err)
		}
		updaterSets = sets
		updaterConfigs = cfgs
	}

	matcherOpts := &libvuln.Options{
		Client:                   &http.Client{Timeout: clientTimeout},
		Store:                    matcherStore,
		Locker:                   NewLocalLockSource(),
		DisableBackgroundUpdates: true,
		UpdateWorkers:            1,
		UpdaterSets:              updaterSets,
		UpdaterConfigs:           updaterConfigs,
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

// loadUpdaterOptions parses a Clair v4 YAML config file and extracts the
// updater-relevant fields. Only the `updaters` section is used; all other
// Clair config fields (TLS, database connstrings, etc.) are parsed but
// ignored. Note that the Clair v4 drop-in config format is not supported;
// only the full config format (github.com/quay/clair/config.Config) is
// accepted here.
func loadUpdaterOptions(path string) (sets []string, configs map[string]driver.ConfigUnmarshaler, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("opening config file: %w", err)
	}
	defer f.Close()

	var cfg clairconfig.Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, nil, fmt.Errorf("decoding config file: %w", err)
	}

	configs = make(map[string]driver.ConfigUnmarshaler, len(cfg.Updaters.Config))
	for name, node := range cfg.Updaters.Config {
		configs[name] = func(v any) error {
			b, err := json.Marshal(node)
			if err != nil {
				return err
			}
			return json.Unmarshal(b, v)
		}
	}

	return cfg.Updaters.Sets, configs, nil
}
