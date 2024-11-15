package output

// There is no official documentation for this format but it's basically a port of
// https://github.com/quay/quay/blob/master/data/secscan_model/secscan_v4_model.py#L447.
// This is what is returned by Quay's /security?vulnerabilities=true endpoint.

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/quay/claircore"
)

const enrichmentMap = "message/vnd.clair.map.vulnerability; enricher=clair.cvss schema=https://csrc.nist.gov/schema/nvd/feed/1.1/cvss-v3.x.json"

type QuayReport struct {
	Status string    `json:"status"`
	Data   *ScanData `json:"data"`
}

type ScanData struct {
	Layer *LayerData `json:"Layer"`
}

type LayerData struct {
	Name             string     `json:"Name"`
	ParentName       string     `json:"ParentName"`
	NamespaceName    string     `json:"NamespaceName"`
	IndexedByVersion int        `json:"IndexedByVersion"`
	Features         []*Feature `json:"Features"`
}

type Feature struct {
	Name            string  `json:"Name"`
	VersionFormat   string  `json:"VersionFormat"`
	NamespaceName   string  `json:"NamespaceName"`
	AddedBy         string  `json:"AddedBy"`
	Version         string  `json:"Version"`
	Vulnerabilities []*Vuln `json:"Vulnerabilities"`
}

type Vuln struct {
	Severity      string    `json:"Severity"`
	NamespaceName string    `json:"NamespaceName"`
	Link          string    `json:"Link"`
	FixedBy       string    `json:"FixedBy"`
	Description   string    `json:"Description"`
	Name          string    `json:"Name"`
	Metadata      *Metadata `json:"Metadata"`
	// This is a departure from the "spec" but needed for Konflux
	Issued time.Time `json:"Issued"`
}

type Metadata struct {
	UpdatedBy     string `json:"UpdatedBy"`
	RepoName      string `json:"RepoName"`
	RepoLink      string `json:"RepoLink"`
	DistroName    string `json:"DistroName"`
	DistroVersion string `json:"DistroVersion"`
	NVD           *NVD   `json:"NVD"`
}

type NVD struct {
	CVSSv3 struct {
		Vectors string  `json:"Vectors"`
		Score   float64 `json:"Score"`
	}
}

type Enrichment struct {
	Version               string  `json:"version"`
	VectorString          string  `json:"vectorString"`
	AttackVector          string  `json:"attackVector"`
	AttackComplexity      string  `json:"attackComplexity"`
	PrivilegesRequired    string  `json:"privilegesRequired"`
	UserInteraction       string  `json:"userInteraction"`
	Scope                 string  `json:"scope"`
	ConfidentialityImpact string  `json:"confidentialityImpact"`
	IntegrityImpact       string  `json:"integrityImpact"`
	AvailabilityImpact    string  `json:"availabilityImpact"`
	BaseSeverity          string  `json:"baseSeverity"`
	BaseScore             float64 `json:"baseScore"`
}

// ReportToSecScan will take a claircore.VulnerabilityReport and return a
// QuayReport which is a format that mimics how Quay surfaces Clair results
// at an API level.
func ReportToSecScan(vr *claircore.VulnerabilityReport) (*QuayReport, error) {
	report := &QuayReport{Status: "scanned", Data: &ScanData{}}
	report.Data.Layer = &LayerData{Name: vr.Hash.String(), IndexedByVersion: 4}
	report.Data.Layer.Features = make([]*Feature, len(vr.Packages))
	packageIndex := 0
	for id, p := range vr.Packages {
		vulns := vr.PackageVulnerabilities[id]
		report.Data.Layer.Features[packageIndex] = &Feature{
			Name: p.Name,
			// What environment are we supposed to choose here when
			// there are multiple? Quay chooses the first like this
			AddedBy:         vr.Environments[id][0].IntroducedIn.String(),
			Version:         p.Version,
			Vulnerabilities: make([]*Vuln, len(vulns)),
		}
		quayVulns := []*Vuln{}
		for _, vulnID := range vulns {
			vuln := vr.Vulnerabilities[vulnID]
			v := &Vuln{
				Severity:      vuln.NormalizedSeverity.String(),
				NamespaceName: vuln.Updater,
				Link:          vuln.Links,
				FixedBy:       vuln.FixedInVersion,
				Description:   vuln.Description,
				Name:          vuln.Name,
				Issued:        vuln.Issued,
				Metadata: &Metadata{
					UpdatedBy:     vuln.Updater,
					RepoName:      vuln.Repo.Name,
					RepoLink:      vuln.Repo.URI,
					DistroName:    vuln.Dist.Name,
					DistroVersion: vuln.Dist.Version,
				},
			}
			enrichmentsObj, ok := vr.Enrichments[enrichmentMap]
			if ok {
				ens := map[string][]*Enrichment{}
				err := json.Unmarshal(enrichmentsObj[0], &ens)
				if err != nil {
					return nil, err
				}
				enrichments, ok := ens[vulnID]
				if ok {
					enrichment := getMostSevereEnrichment(enrichments)
					// Reclassify severity using the CVSS severity if Unknown
					if v.Severity == "Unknown" {
						v.Severity = strings.Title(strings.ToLower(enrichment.BaseSeverity))
					}
					v.Metadata.NVD = &NVD{}
					v.Metadata.NVD.CVSSv3.Score = enrichment.BaseScore
					v.Metadata.NVD.CVSSv3.Vectors = enrichment.VectorString
				}
			}

			quayVulns = append(quayVulns, v)
		}
		report.Data.Layer.Features[packageIndex].Vulnerabilities = quayVulns
		packageIndex++
	}
	return report, nil
}

// GetMostSevereEnrichment will take a slice of Enrichments and return
// the one with the highest baseScore.
//
// This is done to mimic how quay displays this data, future enhancements
// should present all enrichment data keyed by CVE.
func getMostSevereEnrichment(enrichments []*Enrichment) *Enrichment {
	if len(enrichments) == 1 {
		return enrichments[0]
	}
	severeEnrichment := enrichments[0]
	for i := 1; i < len(enrichments)-1; i++ {
		if enrichments[i].BaseScore > severeEnrichment.BaseScore {
			severeEnrichment = enrichments[i]
		}
	}
	return severeEnrichment
}
