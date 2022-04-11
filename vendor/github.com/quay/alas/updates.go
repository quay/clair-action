// Copyright 2019 RedHat

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package alas

type Updates struct {
	Updates []Update `xml:"update"`
}

type Update struct {
	Author      string      `xml:"author,attr"`
	From        string      `xml:"from,attr"`
	Status      string      `xml:"status,attr"`
	Type        string      `xml:"type,attr"`
	Version     string      `xml:"version,attr"`
	ID          string      `xml:"id"`
	Title       string      `xml:"title"`
	Issued      Issued      `xml:"issued"`
	Updated     Updated     `xml:"updated"`
	Severity    string      `xml:"severity"`
	Description string      `xml:"description"`
	References  []Reference `xml:"references>reference"`
	Packages    []Package   `xml:"pkglist>collection>package"`
}

type Issued struct {
	Date string `xml:"date,attr"`
}

type Updated struct {
	Date string `xml:"date,attr"`
}

type Reference struct {
	Href  string `xml:"href,attr"`
	ID    string `xml:"id,attr"`
	Title string `xml:"title,attr"`
	Type  string `xml:"type,attr"`
}

type Package struct {
	Name     string `xml:"name,attr"`
	Epoch    string `xml:"epoch,attr"`
	Version  string `xml:"version,attr"`
	Release  string `xml:"release,attr"`
	Arch     string `xml:"arch,attr"`
	Filename string `xml:"filename"`
}
