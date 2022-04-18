package main

import (
	"fmt"
	"html"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/quay/claircore"
)

type TemplateWriter struct {
	Output   io.Writer
	Template *template.Template
}

func (tw TemplateWriter) Write(vr *claircore.VulnerabilityReport) error {
	err := tw.Template.Execute(tw.Output, vr)
	if err != nil {
		return fmt.Errorf("failed to write with template: %w", err)
	}
	return nil
}

func NewTemplateWriter(output io.Writer, outputTemplate string) (*TemplateWriter, error) {
	templateFuncMap := template.FuncMap{}
	templateFuncMap["endWithPeriod"] = func(input string) string {
		if !strings.HasSuffix(input, ".") {
			input += "."
		}
		return input
	}
	templateFuncMap["escapeString"] = func(input string) string {
		return html.EscapeString(input)
	}
	tmpl, err := template.New("output template").Funcs(templateFuncMap).Parse(outputTemplate)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}
	return &TemplateWriter{Output: output, Template: tmpl}, nil
}

func NewSarifWriter(buf io.Writer) (*TemplateWriter, error) {
	tf, err := os.Open("templates/sarif.tpl")
	if err != nil {
		return nil, err
	}
	defer tf.Close()
	tfb, err := ioutil.ReadAll(tf)
	if err != nil {
		return nil, err
	}
	tw, err := NewTemplateWriter(buf, string(tfb))
	if err != nil {
		return nil, err
	}
	return tw, nil
}
