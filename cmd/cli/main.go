package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "local-clair",
		Version:              "0.0.1",
		Usage:                "local-clair --help",
		Description:          "A CLI application for running Clair v4 locally",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			reportCmd,
			updateCmd,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
