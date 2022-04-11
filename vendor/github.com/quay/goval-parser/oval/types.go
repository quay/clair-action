package oval

import (
	"encoding/xml"
	"fmt"
	"sync"
	"time"
)

// ErrNotFound is returned by Lookup methods when the specified identifier is
// not found.
type ErrNotFound string

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("oval: identifier %q not found", string(e))
}

// Root : root object
type Root struct {
	XMLName     xml.Name    `xml:"oval_definitions"`
	Generator   Generator   `xml:"generator"`
	Definitions Definitions `xml:"definitions"`
	Tests       Tests       `xml:"tests"`
	Objects     Objects     `xml:"objects"`
	States      States      `xml:"states"`
	Variables   Variables   `xml:"variables"`
}

// Generator : >generator
type Generator struct {
	XMLName        xml.Name `xml:"generator"`
	ProductName    string   `xml:"product_name"`
	ProductVersion string   `xml:"product_version"`
	SchemaVersion  string   `xml:"schema_version"`
	Timestamp      string   `xml:"timestamp"`
}

// Definitions : >definitions
type Definitions struct {
	XMLName     xml.Name     `xml:"definitions"`
	Definitions []Definition `xml:"definition"`
}

// Definition : >definitions>definition
type Definition struct {
	XMLName     xml.Name    `xml:"definition"`
	ID          string      `xml:"id,attr"`
	Class       string      `xml:"class,attr"`
	Title       string      `xml:"metadata>title"`
	Affecteds   []Affected  `xml:"metadata>affected"`
	References  []Reference `xml:"metadata>reference"`
	Description string      `xml:"metadata>description"`
	Advisory    Advisory    `xml:"metadata>advisory"` // RedHat, Oracle, Ubuntu
	Debian      Debian      `xml:"metadata>debian"`   // Debian
	Criteria    Criteria    `xml:"criteria"`
}

// Criteria : >definitions>definition>criteria
type Criteria struct {
	XMLName    xml.Name    `xml:"criteria"`
	Operator   string      `xml:"operator,attr"`
	Criterias  []Criteria  `xml:"criteria"`
	Criterions []Criterion `xml:"criterion"`
}

// Criterion : >definitions>definition>criteria>*>criterion
type Criterion struct {
	XMLName xml.Name `xml:"criterion"`
	Negate  bool     `xml:"negate,attr"`
	TestRef string   `xml:"test_ref,attr"`
	Comment string   `xml:"comment,attr"`
}

// Affected : >definitions>definition>metadata>affected
type Affected struct {
	XMLName   xml.Name `xml:"affected"`
	Family    string   `xml:"family,attr"`
	Platforms []string `xml:"platform"`
}

// Reference : >definitions>definition>metadata>reference
type Reference struct {
	XMLName xml.Name `xml:"reference"`
	Source  string   `xml:"source,attr"`
	RefID   string   `xml:"ref_id,attr"`
	RefURL  string   `xml:"ref_url,attr"`
}

// Advisory : >definitions>definition>metadata>advisory
// RedHat and Ubuntu OVAL
type Advisory struct {
	XMLName         xml.Name   `xml:"advisory"`
	Severity        string     `xml:"severity"`
	Cves            []Cve      `xml:"cve"`
	Bugzillas       []Bugzilla `xml:"bugzilla"`
	AffectedCPEList []string   `xml:"affected_cpe_list>cpe"`
	Refs            []Ref      `xml:"ref"` // Ubuntu Only
	Bugs            []Bug      `xml:"bug"` // Ubuntu Only
	PublicDate      Date       `xml:"public_date"`
	Issued          Date       `xml:"issued"`
	Updated         Date       `xml:"updated"`
}

// Date is a wrapper type for decoding a range of date, datestamp, and timestamp
// strings seen in the wild.
//
// Currently, it will only examine attributes with the key "date".
type Date struct {
	Date time.Time
}

var (
	_ xml.Unmarshaler     = (*Date)(nil)
	_ xml.UnmarshalerAttr = (*Date)(nil)

	emptyValue = (time.Time{}).Add(1)
)

// UnmarshalXML implements xml.Unmarshaler.
func (d *Date) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	// Attempt attrs first:
	for _, a := range start.Attr {
		if err := d.UnmarshalXMLAttr(a); err == nil {
			break
		}
	}
	// Attempt the inner node:
	var s string
	if err := dec.DecodeElement(&s, &start); err != nil {
		return err
	}
	switch {
	case d.Date.Equal(emptyValue):
		// If we set the date to this sentinel value, then this element is
		// pointless but we need to not return an error.
		d.Date = time.Time{}
		return nil
	case s == "" && !d.Date.IsZero():
		// If the date is set but an empty string is the inner element, then the
		// date was set by an attr.
		return nil
	default:
	}
	var err error
	// Try a variety of formats, because everything is terrible.
	for _, f := range []string{
		"2006-01-02",              // Debian style `YYYY-MM-DD`
		"2006-01-02 15:04:05 MST", // Ubuntu style `YYYY-MM-DD time zone`
		time.RFC1123,              // The rest of these seem like someone might use them.
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05", // Ubuntu style `YYYY-MM-DD time`, for when you want to seem precise.
	} {
		d.Date, err = time.Parse(f, s)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("unable to decode string as datestamp: %q", s)
}

// UnmarshalXMLAttr implements xml.UnmarshalerAttr.
func (d *Date) UnmarshalXMLAttr(attr xml.Attr) error {
	const dsfmt = `2006-01-02`
	var name = xml.Name{Local: `date`}
	if attr.Name.Local != name.Local {
		return xml.UnmarshalError(fmt.Sprintf("unexpected attr : %v", attr))
	}
	// We want to allow for an empty value, because some vendors can't be
	// bothered to remove empty entities from their database.
	if attr.Value == "" {
		d.Date = emptyValue
		return nil
	}
	var err error
	d.Date, err = time.Parse(dsfmt, attr.Value)
	if err != nil {
		return err
	}
	return nil
}

// Ref : >definitions>definition>metadata>advisory>ref
// Ubuntu OVAL
type Ref struct {
	XMLName xml.Name `xml:"ref"`
	URL     string   `xml:",chardata"`
}

// Bug : >definitions>definition>metadata>advisory>bug
// Ubuntu OVAL
type Bug struct {
	XMLName xml.Name `xml:"bug"`
	URL     string   `xml:",chardata"`
}

// Cve : >definitions>definition>metadata>advisory>cve
// RedHat OVAL
type Cve struct {
	XMLName xml.Name `xml:"cve"`
	CveID   string   `xml:",chardata"`
	Cvss2   string   `xml:"cvss2,attr"`
	Cvss3   string   `xml:"cvss3,attr"`
	Cwe     string   `xml:"cwe,attr"`
	Impact  string   `xml:"impact,attr"`
	Href    string   `xml:"href,attr"`
	Public  string   `xml:"public,attr"`
}

// Bugzilla : >definitions>definition>metadata>advisory>bugzilla
// RedHat OVAL
type Bugzilla struct {
	XMLName xml.Name `xml:"bugzilla"`
	ID      string   `xml:"id,attr"`
	URL     string   `xml:"href,attr"`
	Title   string   `xml:",chardata"`
}

// Debian : >definitions>definition>metadata>debian
type Debian struct {
	XMLName  xml.Name `xml:"debian"`
	MoreInfo string   `xml:"moreinfo"`
	Date     Date     `xml:"date"`
}

// Tests : >tests
type Tests struct {
	once                   sync.Once
	XMLName                xml.Name                `xml:"tests"`
	LineTests              []LineTest              `xml:"line_test"`
	Version55Tests         []Version55Test         `xml:"version55_test"`
	RPMInfoTests           []RPMInfoTest           `xml:"rpminfo_test"`
	DpkgInfoTests          []DpkgInfoTest          `xml:"dpkginfo_test"`
	RPMVerifyFileTests     []RPMVerifyFileTest     `xml:"rpmverifyfile_test"`
	UnameTests             []UnameTest             `xml:"uname_test"`
	TextfileContent54Tests []TextfileContent54Test `xml:"textfilecontent54_test"`
	lineMemo               map[string]int
	version55Memo          map[string]int
	rpminfoMemo            map[string]int
	dpkginfoMemo           map[string]int
	rpmverifyfileMemo      map[string]int
	unameMemo              map[string]int
	textfilecontent54Memo  map[string]int
}

// ObjectRef : >tests>line_test>object-object_ref
//           : >tests>version55_test>object-object_ref
type ObjectRef struct {
	XMLName   xml.Name `xml:"object"`
	ObjectRef string   `xml:"object_ref,attr"`
}

// StateRef : >tests>line_test>state-state_ref
//          : >tests>version55_test>state-state_ref
type StateRef struct {
	XMLName  xml.Name `xml:"state"`
	StateRef string   `xml:"state_ref,attr"`
}

// Objects : >objects
type Objects struct {
	once                     sync.Once
	XMLName                  xml.Name                  `xml:"objects"`
	LineObjects              []LineObject              `xml:"line_object"`
	Version55Objects         []Version55Object         `xml:"version55_object"`
	TextfileContent54Objects []TextfileContent54Object `xml:"textfilecontent54_object"`
	RPMInfoObjects           []RPMInfoObject           `xml:"rpminfo_object"`
	RPMVerifyFileObjects     []RPMVerifyFileObject     `xml:"rpmverifyfile_object"`
	DpkgInfoObjects          []DpkgInfoObject          `xml:"dpkginfo_object"`
	lineMemo                 map[string]int
	version55Memo            map[string]int
	textfilecontent54Memo    map[string]int
	rpminfoMemo              map[string]int
	rpmverifyfileMemo        map[string]int
	dpkginfoMemo             map[string]int
}

// States : >states
type States struct {
	once            sync.Once
	XMLName         xml.Name         `xml:"states"`
	LineStates      []LineState      `xml:"line_state"`
	Version55States []Version55State `xml:"version55_state"`
	RPMInfoStates   []RPMInfoState   `xml:"rpminfo_state"`
	DpkgInfoStates  []DpkgInfoState  `xml:"dpkginfo_state"`
	lineMemo        map[string]int
	version55Memo   map[string]int
	rpminfoMemo     map[string]int
	dpkginfoMemo    map[string]int
}

// Value
type Value struct {
	XMLName xml.Name `xml:"value"`
	Body    string   `xml:",chardata"`
}

// ConstantVariable
type ConstantVariable struct {
	XMLName  xml.Name `xml:"constant_variable"`
	ID       string   `xml:"id,attr"`
	Version  string   `xml:"version,attr"`
	Datatype string   `xml:"datatype,attr"`
	Comment  string   `xml:"comment,attr"`
	Values   []Value  `xml:"value"`
}

// Variables : >variables
type Variables struct {
	once              sync.Once
	XMLName           xml.Name           `xml:"variables"`
	ConstantVariables []ConstantVariable `xml:"constant_variable"`
	dpkginfoMemo      map[string]int
}

// Arch
type Arch struct {
	XMLName   xml.Name  `xml:"arch"`
	Operation Operation `xml:"operation,attr"`
	Body      string    `xml:",chardata"`
}

// Epoch
type Epoch struct {
	XMLName   xml.Name  `xml:"epoch"`
	Operation Operation `xml:"operation,attr"`
	Body      string    `xml:",chardata"`
}

// Release
type Release struct {
	XMLName   xml.Name  `xml:"release"`
	Operation Operation `xml:"operation,attr"`
	Body      string    `xml:",chardata"`
}

// Version
type Version struct {
	XMLName   xml.Name  `xml:"version"`
	Operation Operation `xml:"operation,attr"`
	Body      string    `xml:",chardata"`
}

// EVR
type EVR struct {
	XMLName   xml.Name  `xml:"evr"`
	Operation Operation `xml:"operation,attr"`
	Body      string    `xml:",chardata"`
}
