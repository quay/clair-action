package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/quay/claircore/enricher/cvss"
	"github.com/quay/claircore/libindex"
	"github.com/quay/claircore/libvuln"
	"github.com/quay/claircore/libvuln/driver"
	_ "github.com/quay/claircore/matchers/defaults"
	"github.com/urfave/cli/v2"

	"github.com/crozzy/clair-action/datastore"
	"github.com/crozzy/clair-action/image"
	"github.com/crozzy/clair-action/output"
)

var defaultDBPath = filepath.Join(os.TempDir(), "matcher.db")

type EnumValue struct {
	Enum     []string
	Default  string
	selected string
}

func (e *EnumValue) Set(value string) error {
	for _, enum := range e.Enum {
		if enum == value {
			e.selected = value
			return nil
		}
	}

	return fmt.Errorf("allowed values are %s", strings.Join(e.Enum, ", "))
}

func (e EnumValue) String() string {
	if e.selected == "" {
		return e.Default
	}
	return e.selected
}

const (
	clairFmt = "clair"
	sarifFmt = "sarif"
)

var reportCmd = &cli.Command{
	Name:    "report",
	Aliases: []string{"r"},
	Usage:   "report on a manifest",
	Action:  report,
	Flags: []cli.Flag{
		&cli.PathFlag{
			Name:    "image-path",
			Value:   "",
			Usage:   "where to look for the saved image tar",
			EnvVars: []string{"IMAGE_PATH"},
		},
		&cli.PathFlag{
			Name:    "db-path",
			Value:   "",
			Usage:   "where to look for the matcher DB",
			EnvVars: []string{"DB_PATH"},
		},
		&cli.StringFlag{
			Name:    "image-ref",
			Value:   "",
			Usage:   "the remote location of the image",
			EnvVars: []string{"IMAGE_REF"},
		},
		&cli.StringFlag{
			Name:    "db-url",
			Value:   "",
			Usage:   "the remote location of the sqlite zstd DB",
			EnvVars: []string{"DB_URL"},
		},
		&cli.GenericFlag{
			Name:    "format",
			Aliases: []string{"f"},
			Value: &EnumValue{
				Enum:    []string{clairFmt, sarifFmt},
				Default: clairFmt,
			},
			Usage:   "what output format the results should be in",
			EnvVars: []string{"FORMAT"},
		},
		&cli.IntFlag{
			Name:    "return-code",
			Aliases: []string{"c"},
			Value:   0,
			Usage:   "what status code to return when vulnerabilites are found",
			EnvVars: []string{"RETURN_CODE"},
		},
	},
}

func report(c *cli.Context) error {
	ctx := c.Context

	var (
		// All arguments
		imgRef     = c.String("image-ref")
		imgPath    = c.String("image-path")
		dbPath     = c.String("db-path")
		dbURL      = c.String("db-url")
		format     = c.String("format")
		returnCode = c.Int("return-code")
	)

	var (
		img image.Image
		fa  libindex.Arena
	)
	switch {
	case imgPath != "":
		fa = &LocalFetchArena{}
		var err error
		img, err = image.NewDockerLocalImage(ctx, imgPath, os.TempDir())
		if err != nil {
			return fmt.Errorf("error getting image information %v", err)
		}
	case imgRef != "":
		cl := http.DefaultClient
		fa = libindex.NewRemoteFetchArena(cl, os.TempDir())
		img = image.NewDockerRemoteImage(ctx, imgRef)
	default:
		return fmt.Errorf("no $IMAGE_PATH / --image-path or $IMAGE_REF / --image-ref set")
	}

	switch {
	case dbPath != "":
	case dbURL != "":
		dbPath = defaultDBPath
		var err error
		err = datastore.DownloadDB(ctx, dbURL, defaultDBPath)
		if err != nil {
			return fmt.Errorf("could not download database: %v", err)
		}
	default:
		return fmt.Errorf("no $DB_PATH / --db-path or $DB_URL / --db-url set")
	}

	matcherStore, err := datastore.NewSQLiteMatcherStore(dbPath, true)
	if err != nil {
		return fmt.Errorf("error creating sqlite backend: %v", err)
	}

	matcherOpts := &libvuln.Options{
		Store:                    matcherStore,
		Locker:                   NewLocalLockSource(),
		DisableBackgroundUpdates: true,
		UpdateWorkers:            1,
		Enrichers: []driver.Enricher{
			&cvss.Enricher{},
		},
	}
	lv, err := libvuln.New(ctx, matcherOpts)
	if err != nil {
		return fmt.Errorf("error creating Libvuln: %v", err)
	}

	mf, err := img.GetManifest(ctx)
	if err != nil {
		return fmt.Errorf("error creating manifest %v", err)
	}

	indexerOpts := &libindex.Options{
		Store:      datastore.NewLocalIndexerStore(),
		Locker:     NewLocalLockSource(),
		FetchArena: fa,
	}
	li, err := libindex.New(ctx, indexerOpts, http.DefaultClient)
	if err != nil {
		return fmt.Errorf("error creating Libindex %v", err)
	}
	ir, err := li.Index(ctx, mf)
	if err != nil {
		return fmt.Errorf("error creating index report %v", err)
	}

	vr, err := lv.Scan(ctx, ir)
	if err != nil {
		return fmt.Errorf("error scanning index report %v", err)
	}

	switch format {
	case sarifFmt:
		tw, err := output.NewSarifWriter(os.Stdout)
		if err != nil {
			return fmt.Errorf("error creating sarif report writer %v", err)
		}
		err = tw.Write(vr)
		if err != nil {
			return fmt.Errorf("error writing sarif report %v", err)
		}
	default:
		b, err := json.MarshalIndent(vr, "", "  ")
		if err != nil {
			return fmt.Errorf("could not marshal vulnerability report: %v", err)
		}
		fmt.Println(string(b))
	}

	if len(vr.Vulnerabilities) > 0 {
		return cli.Exit("found vulnerabilies", returnCode)
	}
	return nil
}
