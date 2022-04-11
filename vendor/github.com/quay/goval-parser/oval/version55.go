package oval

import "encoding/xml"

// Version55Test : >tests>version55_test
type Version55Test struct {
	XMLName       xml.Name `xml:"version55_test"`
	ID            string   `xml:"id,attr"`
	StateOperator string   `xml:"state_operator,attr"`
	Comment       string   `xml:"comment,attr"`
	testRef
}

var _ Test = (*Version55Test)(nil)

// Version55Object : >objects>version55_object
type Version55Object struct {
	XMLName xml.Name `xml:"version55_object"`
	ID      string   `xml:"id,attr"`
}

// Version55State : >states>version55_state
type Version55State struct {
	XMLName       xml.Name `xml:"version55_state"`
	ID            string   `xml:"id,attr"`
	VersionString string   `xml:"version_string"`
}
