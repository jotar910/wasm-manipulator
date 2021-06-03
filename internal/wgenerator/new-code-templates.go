package wgenerator

import (
	"text/template"

	"github.com/sirupsen/logrus"
)

var (
	globalTemplateStr   = `(global {{ .Name }} (mut {{ .Type }}) ({{ .Type }}.const {{ if (ne .Value "") }}{{ .Value }}{{ else }}0{{ end }}))`
	typeTemplateStr     = `(type {{ .Name }} (func {{ with .Params }}(param{{ range . }} {{ . }}{{ end }}){{ end }} {{ if (ne .Result "") }}(result {{ .Result }}){{ end }}))`
	functionTemplateStr = `(func {{ .Name }} (type {{ .TypeName }})
	{{ with .Params }}{{ range . }}(param {{ .Name }} {{ .Type }}) {{ end }}{{ end }} {{ if (ne .Result "") }}(result {{ .Result }}){{ end }}
	{{ printf "%s" .Code }}
)`
	startFunctionTemplateStr  = `(start {{ .Name }})`
	importFunctionTemplateStr = `(import "{{ .ModuleName }}" "{{ .ExportName }}" (func {{ .Name }} (type {{ .TypeName }})))`
	exportFunctionTemplateStr = `(export "{{ .ExportName }}" (func {{ .Name }}))`
	localTemplateStr          = `(local {{ .Name }} {{ .Type }})`
	setLocalTemplateStr       = `(local.set {{ .Name }} ({{ .Type }}.const {{ .Value }}))`
	setLocalInstrTemplateStr  = `(local.set {{ .Name }} {{ .Instruction }})`
	getVariableTemplateStr    = `{{ if .IsLocal }}(local.get {{ .Name }}){{ else }}(global.get {{ .Name }}){{ end }}`
)

var (
	globalTemplate         *template.Template
	typeTemplate           *template.Template
	functionTemplate       *template.Template
	startFunctionTemplate  *template.Template
	importFunctionTemplate *template.Template
	exportFunctionTemplate *template.Template
	localTemplate          *template.Template
	setLocalTemplate       *template.Template
	setLocalInstrTemplate  *template.Template
	getVariableTemplate    *template.Template
)

// init initializes the templates.
func init() {
	var err error
	// Parse global template.
	globalTemplate, err = template.New("global-template").Parse(globalTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
	// Parse type template.
	typeTemplate, err = template.New("type-template").Parse(typeTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
	// Parse function template.
	functionTemplate, err = template.New("function-template").Parse(functionTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
	// Parse function template.
	startFunctionTemplate, err = template.New("start-function-template").Parse(startFunctionTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
	// Parse import function template.
	importFunctionTemplate, err = template.New("import-function-template").Parse(importFunctionTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
	// Parse export function template.
	exportFunctionTemplate, err = template.New("export-function-template").Parse(exportFunctionTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
	// Parse local template.
	localTemplate, err = template.New("local-template").Parse(localTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
	// Parse set local template.
	setLocalTemplate, err = template.New("set-local-template").Parse(setLocalTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
	// Parse set local instruction template.
	setLocalInstrTemplate, err = template.New("set-local-instruction-template").Parse(setLocalInstrTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
	// Parse get local template.
	getVariableTemplate, err = template.New("get-local-template").Parse(getVariableTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
}
