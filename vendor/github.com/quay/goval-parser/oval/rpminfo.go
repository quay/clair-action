package oval

import (
	"encoding/xml"
)

// RPMInfoTest : >tests>rpminfo_test
type RPMInfoTest struct {
	XMLName xml.Name `xml:"rpminfo_test"`
	ID      string   `xml:"id,attr"`
	Comment string   `xml:"comment,attr"`
	Check   string   `xml:"check,attr"`
	Version int      `xml:"version,attr"`
	testRef
}

var _ Test = (*RPMInfoTest)(nil)

// RPMVerifyFileTest : >tests>rpmverifyfile_test
type RPMVerifyFileTest struct {
	XMLName xml.Name `xml:"rpmverifyfile_test"`
	ID      string   `xml:"id,attr"`
	Comment string   `xml:"comment,attr"`
	Check   string   `xml:"check,attr"`
	Version int      `xml:"version,attr"`
	testRef
}

var _ Test = (*RPMVerifyFileTest)(nil)

// RPMInfoObject : >objects>RPMInfo_object
type RPMInfoObject struct {
	XMLName xml.Name `xml:"rpminfo_object"`
	ID      string   `xml:"id,attr"`
	Version int      `xml:"version,attr"`
	Name    string   `xml:"name"`
}

type RPMVerifyFileObject struct {
	XMLName    xml.Name            `xml:"rpmverifyfile_object"`
	ID         string              `xml:"id,attr"`
	Version    int                 `xml:"version,attr"`
	Behaviors  VerifyFileBehaviors `xml:"behaviors"`
	Name       VerifyFileOp        `xml:"name"`
	Epoch      VerifyFileOp        `xml:"epoch"`
	Release    VerifyFileOp        `xml:"release"`
	RPMVersion VerifyFileOp        `xml:"version"`
	Arch       VerifyFileOp        `xml:"arch"`
	Filepath   string              `xml:"filepath"`
}

type VerifyFileBehaviors struct {
	XMLName       xml.Name `xml:"behaviors"`
	NoConfigFiles bool     `xml:"noconfigfiles"`
	NoGhostFiles  bool     `xml:"noghostfiles"`
	NoGroup       bool     `xml:"nogroup"`
	NoLinkTo      bool     `xml:"nolinkto"`
	NoMD5         bool     `xml:"nomd5"`
	NoMode        bool     `xml:"nomode"`
	NoMTime       bool     `xml:"nomtime"`
	NoRDev        bool     `xml:"nordev"`
	NoSize        bool     `xml:"nosize"`
	NoUser        bool     `xml:"nouser"`
}

type VerifyFileOp struct {
	XMLName xml.Name
	Op      string `xml:"operation,attr"`
}

// RPMInfoState : >states>rpminfo_state
type RPMInfoState struct {
	XMLName        xml.Name           `xml:"rpminfo_state"`
	ID             string             `xml:"id,attr"`
	Version        int                `xml:"version,attr"`
	Arch           *Arch              `xml:"arch"`
	Epoch          *Epoch             `xml:"epoch"`
	Release        *Release           `xml:"release"`
	RPMVersion     *Version           `xml:"version"`
	EVR            *EVR               `xml:"evr"`
	SignatureKeyID *RPMSignatureKeyID `xml:"signature_keyid"`
}

type RPMSignatureKeyID struct {
	XMLName   xml.Name  `xml:"signature_keyid"`
	Operation Operation `xml:"operation,attr"`
	Body      string    `xml:",chardata"`
}
