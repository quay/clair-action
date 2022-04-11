package oval

import (
	"encoding/xml"
)

// DpkgInfoTest : >tests>dpkginfo_test
type DpkgInfoTest struct {
	XMLName xml.Name `xml:"dpkginfo_test"`
	ID      string   `xml:"id,attr"`
	Comment string   `xml:"comment,attr"`
	Check   string   `xml:"check,attr"`
	Version int      `xml:"version,attr"`
	testRef
}

var _ Test = (*DpkgInfoTest)(nil)

// DpkgName : >objects>dpkginfo_object>name
//
// when parsing ubuntu var_ref is a reference
// to the <variables> section of the document
//
// when parsing debian var_ref is empty and
// the Body field is used directly
type DpkgName struct {
	XMLName xml.Name `xml:"name"`
	Ref     string   `xml:"var_ref,attr"`
	Body    string   `xml:",chardata"`
}

// DpkgInfoObject : >objects>dpkginfo_object
type DpkgInfoObject struct {
	XMLName xml.Name  `xml:"dpkginfo_object"`
	ID      string    `xml:"id,attr"`
	Version int       `xml:"version,attr"`
	Name    *DpkgName `xml:"name"`
}

// DpkgInfoState : >states>dpkginfo_state
type DpkgInfoState struct {
	XMLName     xml.Name `xml:"dpkginfo_state"`
	ID          string   `xml:"id,attr"`
	Version     int      `xml:"version,attr"`
	Arch        *Arch    `xml:"arch"`
	Epoch       *Epoch   `xml:"epoch"`
	Release     *Release `xml:"release"`
	DpkgVersion *Version `xml:"version"`
	EVR         *EVR     `xml:"evr"`
}
