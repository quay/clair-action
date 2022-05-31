{
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Clair V4",
          "informationUri": "https://github.com/quay/clair",
          "fullName": "Clair V4 Vulnerability Scanner",
          "version": "v0.0.1",
          "rules": [
          {{- $t_first := true }}
            {{- range $package_id, $vulnerability_list := .PackageVulnerabilities -}}
              {{- range $vuln_id := $vulnerability_list -}}
                {{- if $t_first -}}
                  {{- $t_first = false -}}
                {{ else -}}
                  ,
                {{- end }}
                  {
                  "id": "{{ $vuln_id }}-{{ $package_id }}",
                  "name": "scan_results",
                  "shortDescription": {
                    "text": "{{ with ( index $.Packages $package_id ) }}{{ .Name }}{{ end }} - {{ with ( index $.Packages $package_id ) }}{{ .Version }}{{ end }} - {{ with ( index $.Vulnerabilities $vuln_id ) }}{{ .Name }}{{ end }}"
                  },
                  "fullDescription": {
                    "text": {{ with ( index $.Vulnerabilities $vuln_id ) }}"{{ (.Description | escapeString) | js }}"{{ end }}
                  },
                  "help": {
                    "text": {{ with ( index $.Vulnerabilities $vuln_id ) }}"{{ (.Description | escapeString) | js }}{{- if ne .FixedInVersion "" }}\n\nFixed In: {{- end }}{{ .FixedInVersion }}"{{ end }}
                  },
                  "properties": {
                    "tags": [
                      "vulnerability",
                      {{ with ( index $.Vulnerabilities $vuln_id ) }}"{{ .Repo.Name }}"{{ end }},
                      {{ with ( index $.Packages $package_id ) }}"{{ .Name }}"{{ end }},
                      {{ with ( index $.Packages $package_id ) }}"{{ .Version }}"{{ end }}
                    ],
                    "precision": "very-high"
                  }
                }
               {{- end -}}
             {{- end -}}
            ]
        }
      },
      "results": [
      {{ $i := 0 }}
      {{- $t_first := true }}
        {{- range $package_id, $vulnerability_list := .PackageVulnerabilities -}}
          {{- range $vuln_id := $vulnerability_list -}}
            {{- if $t_first -}}
              {{- $t_first = false -}}
            {{ else -}}
              ,
            {{- end }}
            {
              "ruleId": "{{ $vuln_id }}-{{ $package_id }}",
              "ruleIndex": {{ $i }},
              "level": "warning",
              "message": {
                "text": "{{ with ( index $.Packages $package_id ) }}{{ .Name }}{{ end }} - {{ with ( index $.Packages $package_id ) }}{{ .Version }}{{ end }} - {{ with ( index $.Vulnerabilities $vuln_id ) }}{{ .Name }}{{ end }}"
              },
              "locations": [{
                "physicalLocation": {
                  "artifactLocation": {
                    "uri": "Dockerfile"
                  },
                  "region": {
                    "startLine": 1,
                    "startColumn": 1,
                    "endColumn": 1
                  }
                }
              }]
            }
            {{ $i = inc $i }}
          {{- end }}
        {{- end }}
      ],
      "columnKind": "utf16CodeUnits"
    }
  ]
}
