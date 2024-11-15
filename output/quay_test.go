package output

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/quay/claircore"
)

var tcs = []struct {
	name        string
	testFile    string
	numPackages int
	numVulns    int
}{
	{
		name:        "quay image",
		testFile:    "testdata/quay_v3_4.json",
		numPackages: 502,
		numVulns:    301,
	},
	{
		name:        "debian image",
		testFile:    "testdata/debian_stretch.json",
		numPackages: 83,
		numVulns:    212,
	},
}

// Currently there is no OpenAPI (or other) spec for the secscan format
// returned by Quay so these are just sanity checks and do not test for
// absolute correctness (if such a thing exists).
func TestQuayReport(t *testing.T) {
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			vrf, err := os.Open(tc.testFile)
			if err != nil {
				t.Fatalf("got an error opening vulnerability report: %v", err)
			}
			vrb, err := ioutil.ReadAll(vrf)
			if err != nil {
				t.Fatalf("got an error reading vuln report bytes: %v", err)
			}

			vr := &claircore.VulnerabilityReport{}
			err = json.Unmarshal(vrb, vr)
			if err != nil {
				t.Fatalf("error unmarshaling vr: %v", err)
			}

			quayReport, err := ReportToSecScan(vr)
			if err != nil {
				t.Fatalf("error creating quayReport: %v", err)
			}
			if len(quayReport.Data.Layer.Features) != tc.numPackages {
				t.Fatalf("found %d packages, wanted %d", len(quayReport.Data.Layer.Features), tc.numPackages)
			}
			totalVulns := 0
			for _, f := range quayReport.Data.Layer.Features {
				totalVulns = totalVulns + len(f.Vulnerabilities)
			}
			if totalVulns != tc.numVulns {
				t.Fatalf("found %d vulns, wanted %d", totalVulns, tc.numVulns)
			}
			rep, err := json.MarshalIndent(quayReport, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			t.Log(string(rep))
		})
	}
}
