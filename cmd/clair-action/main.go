package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/quay/zlog"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

var (
	logout = zerolog.New(&zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}).Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Logger()
)

func main() {
	app := &cli.App{
		Name:                 "clair-action",
		Version:              Version,
		Usage:                "clair-action --help",
		Description:          "A CLI application for running Clair v4 locally",
		EnableBashCompletion: true,
		Before: func(c *cli.Context) error {
			level, err := zerolog.ParseLevel(c.String("level"))
			if err != nil {
				return fmt.Errorf("unknown log level %s", c.String("level"))
			}
			logout = logout.Level(level)
			if c.String("level") == "" {
				logout = logout.Level(zerolog.InfoLevel)
			}
			zlog.Set(&logout)
			return nil
		},

		Commands: []*cli.Command{
			reportCmd,
			updateCmd,
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "level",
				Aliases: []string{"l"},
				Usage:   "specify level log for output (debug, warn, info, disabled, default: info)",
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
