package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/quay/claircore/enricher/cvss"
	"github.com/quay/claircore/libindex"
	"github.com/quay/claircore/libvuln"
	"github.com/quay/claircore/libvuln/driver"
	_ "github.com/quay/claircore/matchers/defaults"
	"github.com/urfave/cli/v2"
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
	},
}

func report(c *cli.Context) error {
	ctx := c.Context

	imgRef := c.String("image-ref")
	imgPath := c.String("image-path")
	fmt.Println(imgRef)

	var (
		img Image
		fa  libindex.Arena
		err error
	)

	switch {
	case imgPath != "":
		fa = &LocalFetchArena{}
		img, err = NewDockerLocalImage(ctx, imgPath, os.TempDir())
	case imgRef != "":
		cl := http.DefaultClient
		fa = libindex.NewRemoteFetchArena(cl, os.TempDir())
		img, err = NewDockerRemoteImage(ctx, imgRef)
	default:
		return fmt.Errorf("no $IMAGE_PATH / --image-path or $IMAGE_REF / --image-ref set")
	}
	if err != nil {
		return fmt.Errorf("error getting image information %v", err)
	}

	dbPath := c.String("db-path")
	if dbPath == "" {
		return fmt.Errorf("no $DB_PATH or --db-path set")
	}

	matcherStore, err := NewSQLiteMatcherStore(dbPath, true)
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
		Store:      NewLocalIndexerStore(),
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

	// b, err := json.MarshalIndent(vr, "", "  ")
	// if err != nil {
	// 	return fmt.Errorf("could not marshal vulnerability report: %v", err)
	// }
	// fmt.Println(string(b))

	tw, err := NewSarifWriter(os.Stdout)
	if err != nil {
		return fmt.Errorf("error creating sarif report writer %v", err)
	}
	err = tw.Write(vr)
	if err != nil {
		return fmt.Errorf("error writing sarif report %v", err)
	}
	return nil
}
