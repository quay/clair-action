package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/quay/clair-action/output"
	"github.com/quay/claircore"
	"github.com/urfave/cli/v2"
)

var convertCmd = &cli.Command{
	Name:    "convert",
	Aliases: []string{"c"},
	Usage:   "Convert a Clair Vulnerability report to the Quay secscan or Sarif format",
	Action:  convert,
	Flags: []cli.Flag{
		&cli.PathFlag{
			Name:    "file-path",
			Value:   "",
			Usage:   "where to look for the Clair Vulnerability report to convert",
			EnvVars: []string{"FILE_PATH"},
		},
		&cli.GenericFlag{
			Name:    "format",
			Aliases: []string{"f"},
			Value: &EnumValue{
				Enum:    []string{sarifFmt, quayFmt},
				Default: quayFmt,
			},
			Usage:   "what output format the results should be in",
			EnvVars: []string{"FORMAT"},
		},
	},
}

func convert(c *cli.Context) error {
	var (
		// All arguments
		filePath = c.String("file-path")
		format   = c.String("format")
	)
	vulnReportFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}

	defer vulnReportFile.Close()

	bs, err := io.ReadAll(vulnReportFile)
	if err != nil {
		return fmt.Errorf("error reading file data: %v", err)
	}

	var vr claircore.VulnerabilityReport
	err = json.Unmarshal(bs, &vr)
	if err != nil {
		return fmt.Errorf("error unmarshaling Vulnerability report: %v", err)
	}
	switch format {
	case sarifFmt:
		tw, err := output.NewSarifWriter(os.Stdout)
		if err != nil {
			return fmt.Errorf("error creating sarif report writer: %v", err)
		}
		err = tw.Write(&vr)
		if err != nil {
			return fmt.Errorf("error writing sarif report: %v", err)
		}
	case quayFmt:
		quayReport, err := output.ReportToSecScan(&vr)
		if err != nil {
			return fmt.Errorf("error creating quay format report: %v", err)
		}
		b, err := json.MarshalIndent(quayReport, "", "  ")
		if err != nil {
			return fmt.Errorf("could not marshal quay report: %v", err)
		}
		fmt.Println(string(b))
	default:
		return fmt.Errorf("unrecognized format: %q", format)
	}
	return nil
}
