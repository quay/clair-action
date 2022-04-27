package output

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/quay/claircore"
)

func TestTemplateRender(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	tw, err := NewSarifWriter(buf)
	if err != nil {
		t.Fatalf("got an error creating template writer: %v", err)
	}

	// Vulnerability Report processing
	vrf, err := os.Open("testdata/quay-rhel8.json")
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
	err = tw.Write(vr)
	if err != nil {
		t.Fatalf("error writing template: %v", err)
	}
}
