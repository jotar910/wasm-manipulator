package wgenerator

import (
	"text/template"

	"github.com/sirupsen/logrus"
)

// init initializes the templates.
func init() {
	var err error
	primitiveEvalTemplateStart, err = template.New("primitive-eval-start-template").Parse(primitiveEvalTemplateStartStr)
	if err != nil {
		logrus.Fatal(err)
	}
	compositeCallEvalTemplateStart, err = template.New("composite-call-eval-start-template").Parse(compositeCallEvalTemplateStartStr)
	if err != nil {
		logrus.Fatal(err)
	}
	compositeCallEvalTemplateEnd, err = template.New("composite-call-eval-end-template").Parse(compositeCallEvalTemplateEndStr)
	if err != nil {
		logrus.Fatal(err)
	}
	compositeZoneEvalTemplateStart, err = template.New("composite-zone-eval-start-template").Parse(compositeZoneEvalTemplateStartStr)
	if err != nil {
		logrus.Fatal(err)
	}
	compositeReturnEvalTemplate, err = template.New("composite-return-eval-template").Parse(compositeReturnEvalTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
	compositeReturnEvalRefTemplate, err = template.New("composite-return-eval-ref-template").Parse(compositeReturnEvalRefTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
	setStartingGlobalCompositeTemplate, err = template.New("set-starting-global-composite-template").Parse(setStartingGlobalCompositeTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
	zoneLocalCompositeTemplate, err = template.New("zone-local-composite-template").Parse(zoneLocalCompositeTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
	zoneParamCompositeTemplate, err = template.New("zone-param-composite-template").Parse(zoneParamCompositeTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
	zonePrimitiveTemplate, err = template.New("zone-primitive-template").Parse(zonePrimitiveTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
	jsTemplate, err = template.New("js-template").Parse(jsTemplateStr)
	if err != nil {
		logrus.Fatal(err)
	}
}
