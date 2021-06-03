package template

import (
	"regexp"
	"strings"

	"joao/wasm-manipulator/internal/wtemplate"
)

var variableRegex = regexp.MustCompile(`%[a-zA-Z][^%]*%`)
var operationRegex = regexp.MustCompile(`:!?[a-zA-Z][\w\d]*\([^)]*\)`)

// variableName returns the variable name.
func variableName(variable string) string {
	end := len(variable)
	if delIndex := strings.Index(variable, ":"); delIndex > -1 {
		end = delIndex
	}
	return strings.Trim(variable[:end], "%")
}

// operationName returns the operation name.
func operationName(operation string) string {
	end := len(operation) - 1
	if delIndex := strings.Index(operation, "("); delIndex > -1 {
		end = delIndex
	}
	return operation[1:end]
}

// operationArgs returns the operation arguments.
func operationArgs(operation string) []string {
	var res []string
	start := strings.Index(operation, "(")
	if start < 0 {
		return res
	}
	args := strings.Split(operation[start+1:len(operation)-1], ",")
	for _, arg := range args {
		res = append(res, strings.TrimSpace(arg))
	}
	return res
}

// Parse parses the template input.
func Parse(name, value string) (*wtemplate.Template, error) {
	res := wtemplate.NewTemplate(name, value)
	for _, v := range variableRegex.FindAllString(value, -1) {
		vName := variableName(v)
		res.AddVariable(vName)
		for _, o := range operationRegex.FindAllString(v, -1) {
			opName := operationName(o)
			if strings.Index(opName, wtemplate.NotOp) == 0 {
				res.AddOperation(vName, wtemplate.NotOp, []string{})
				res.AddOperation(vName, opName[len(wtemplate.NotOp):], operationArgs(o))
			} else {
				res.AddOperation(vName, opName, operationArgs(o))
			}
		}
	}
	return res, nil
}
