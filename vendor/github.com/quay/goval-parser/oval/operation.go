package oval

import (
	"encoding"
	"fmt"
	"strings"
)

type Operation int

const (
	_ Operation = iota // Invalid

	OpEquals                   // equals
	OpNotEquals                // not equals
	OpCaseInsensitiveEquals    // case insensitive equals
	OpCaseInsensitiveNotEquals // case insensitive not equals
	OpGreaterThan              // greater than
	OpLessThan                 // less than
	OpGreaterThanOrEqual       // greater than or equal
	OpLessThanOrEqual          // less than or equal
	OpBitwiseAnd               // bitwise and
	OpBitwiseOr                // bitwise or
	OpPatternMatch             // pattern match
	OpSubset                   // subset of
	OpSuperset                 // superset of
)

//go:generate stringer -type=Operation -linecomment

// UnmarshalText implements encoding.TextUnmarshaler.
func (o *Operation) UnmarshalText(text []byte) error {
	off := strings.Index(_Operation_name, string(text))
	for i, idx := range _Operation_index {
		if int(idx) == off {
			*o = Operation(i + 1)
			return nil
		}
	}
	o = new(Operation)
	return fmt.Errorf("oval: unknown operation %q", string(text))
}

var _ encoding.TextUnmarshaler = (*Operation)(nil)
