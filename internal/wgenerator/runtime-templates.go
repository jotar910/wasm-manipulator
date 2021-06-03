package wgenerator

import (
	"text/template"

	_ "embed"
)

var (
	//go:embed resources/primitive-eval-template-start.go.txt
	primitiveEvalTemplateStartStr string

	//go:embed resources/composite-call-eval-template-start.go.txt
	compositeCallEvalTemplateStartStr string

	//go:embed resources/composite-call-eval-template-end.go.txt
	compositeCallEvalTemplateEndStr string

	//go:embed resources/composite-zone-eval-template-start.go.txt
	compositeZoneEvalTemplateStartStr string

	//go:embed resources/composite-return-eval-template.go.txt
	compositeReturnEvalTemplateStr string

	//go:embed resources/composite-return-eval-ref-template.go.txt
	compositeReturnEvalRefTemplateStr string

	//go:embed resources/set-starting-global-composite-template.go.txt
	setStartingGlobalCompositeTemplateStr string

	//go:embed resources/zone-local-composite-template.go.txt
	zoneLocalCompositeTemplateStr string

	//go:embed resources/zone-param-composite-template.go.txt
	zoneParamCompositeTemplateStr string

	//go:embed resources/zone-primitive-template.go.txt
	zonePrimitiveTemplateStr string
)

var (
	primitiveEvalTemplateStart         *template.Template
	compositeCallEvalTemplateStart     *template.Template
	compositeCallEvalTemplateEnd       *template.Template
	compositeZoneEvalTemplateStart     *template.Template
	compositeReturnEvalTemplate        *template.Template
	compositeReturnEvalRefTemplate     *template.Template
	setStartingGlobalCompositeTemplate *template.Template
	zoneLocalCompositeTemplate         *template.Template
	zoneParamCompositeTemplate         *template.Template
	zonePrimitiveTemplate              *template.Template
)
