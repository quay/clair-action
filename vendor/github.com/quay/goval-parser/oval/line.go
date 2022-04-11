package oval

import "encoding/xml"

// LineTest : >tests>line_test
type LineTest struct {
	XMLName       xml.Name `xml:"line_test"`
	ID            string   `xml:"id,attr"`
	StateOperator string   `xml:"state_operator,attr"`
	Comment       string   `xml:"comment,attr"`
	testRef
}

var _ Test = (*LineTest)(nil)

// LineObject : >objects>line_object
type LineObject struct {
	XMLName         xml.Name `xml:"line_object"`
	ID              string   `xml:"id,attr"`
	ShowSubcommands []string `xml:"show_subcommand"`
}

// LineState : >states>line_state
type LineState struct {
	XMLName        xml.Name   `xml:"line_state"`
	ID             string     `xml:"id,attr"`
	ShowSubcommand string     `xml:"show_subcommand"`
	ConfigLine     ConfigLine `xml:"config_line"`
}

// ConfigLine : >states>line_state>config_line
type ConfigLine struct {
	XMLName   xml.Name `xml:"config_line"`
	Body      string   `xml:",chardata"`
	Operation string   `xml:"operation,attr"`
}
