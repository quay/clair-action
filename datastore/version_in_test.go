package datastore

import (
	"database/sql/driver"
	"errors"
	"testing"
)

func TestMatcherIntegration(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		vulnRange  string
		vulnerable bool
		err        error
	}{
		{
			name:       "happy path vulnerable",
			version:    "0.1.1.0.0.0.0.0.0.0",
			vulnRange:  "{0,0,0,0,0,0,0,0,0,0}__{1,0,0,0,0,0,0,0,0,0}",
			vulnerable: true,
			err:        nil,
		},
		{
			name:       "happy path vulnerable exact",
			version:    "0.0.0.0.0.0.0.0.0.0",
			vulnRange:  "{0,0,0,0,0,0,0,0,0,0}__{1,0,0,0,0,0,0,0,0,0}",
			vulnerable: true,
			err:        nil,
		},
		{
			name:       "happy path not vulnerable",
			version:    "2.0.0.0.0.0.0.0.0.0",
			vulnRange:  "{0,0,0,0,0,0,0,0,0,0}__{1,1,0,0,0,0,0,0,0,0}",
			vulnerable: false,
			err:        nil,
		},
		{
			name:       "happy path not vulnerable exact",
			version:    "1.1.0.0.0.0.0.0.0.0",
			vulnRange:  "{0,0,0,0,0,0,0,0,0,0}__{1,1,0,0,0,0,0,0,0,0}",
			vulnerable: false,
			err:        nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vulnerable, err := _sqliteVersionIn(nil, []driver.Value{tt.version, tt.vulnRange})
			if !errors.Is(err, tt.err) {
				t.Fatalf("mismatched errors %v %v", err, tt.err)
			}
			vulnerable = vulnerable.(bool)
			if vulnerable != tt.vulnerable {
				t.Fatalf("expected vulnerable == %v but got vulnerable %v", tt.vulnerable, vulnerable)
			}
		})
	}

}
