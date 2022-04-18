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
        }
      },
      "results": [
    {{- $t_first := true }}
        {{- range $index, $vulnerability := .Vulnerabilities -}}
          {{- if $t_first -}}
            {{- $t_first = false -}}
          {{ else -}}
            ,
          {{- end }}
        {
          "ruleId": "{{ $vulnerability.ID }}",
          "ruleIndex": {{ $index }},
          "level": "error",
          "message": {
            "text": {0} - {1} - {2} {{ endWithPeriod $vulnerability.Description | printf "%q" }},
            "arguments": [
                {{ $vulnerability.Package.Name }},
                {{ $vulnerability.Package.Version }},
                {{ $vulnerability.NormalizedSeverity }},
            ]
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
        {{- end -}}
      ],
      "columnKind": "utf16CodeUnits"
    }
  ]
}
