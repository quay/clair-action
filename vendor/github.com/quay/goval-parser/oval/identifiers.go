package oval

import (
	"bytes"
	"encoding"
	"fmt"
	"strconv"
)

func ParseID(ref string) (*ID, error) {
	id := new(ID)
	return id, id.UnmarshalText([]byte(ref))
}

// ID is an OVAL identifier.
type ID struct {
	Namespace string
	Type      IDType
	Value     int64
}

func (id ID) String() string {
	return fmt.Sprintf(`oval:%s:%s:%d`, id.Namespace, id.Type, id.Value)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (id *ID) UnmarshalText(data []byte) error {
	fs := bytes.FieldsFunc(data, func(r rune) bool { return r == ':' })
	if got := len(fs); got != 4 {
		return fmt.Errorf("oval: malformed ID: found %d segments", got)
	}
	for i, f := range fs {
		switch i {
		case 0:
			if !bytes.Equal(f, []byte("oval")) {
				return fmt.Errorf(`oval: malformed ID: prefix is %q, not "oval"`, string(f))
			}
		case 1:
			id.Namespace = string(f)
		case 2:
			if err := id.Type.UnmarshalText(f); err != nil {
				return fmt.Errorf("oval: malformed ID: %w", err)
			}
		case 3:
			var err error
			id.Value, err = strconv.ParseInt(string(f), 10, 64)
			if err != nil {
				return fmt.Errorf("oval: malformed ID: %w", err)
			}
		}
	}
	return nil
}

var _ encoding.TextUnmarshaler = (*ID)(nil)

// IDType enumerates the valid OVAL identifier types.
type IDType int

// This block declares the valid OVAL identifier types.
const (
	_              IDType = iota // invalid
	OvalDefinition               // def
	OvalTest                     // tst
	OvalObject                   // obj
	OvalState                    // ste
	OvalVariable                 // var
)

//go:generate stringer -type=IDType -linecomment

// UnmarshalText implements encoding.TextUnmarshaler.
func (i *IDType) UnmarshalText(data []byte) error {
	switch {
	case bytes.Equal(data, []byte("def")):
		*i = OvalDefinition
	case bytes.Equal(data, []byte("tst")):
		*i = OvalTest
	case bytes.Equal(data, []byte("obj")):
		*i = OvalObject
	case bytes.Equal(data, []byte("ste")):
		*i = OvalState
	case bytes.Equal(data, []byte("var")):
		*i = OvalVariable
	default:
		return fmt.Errorf("unknown id type %q", string(data))
	}
	return nil
}

var _ encoding.TextUnmarshaler = (*IDType)(nil)
