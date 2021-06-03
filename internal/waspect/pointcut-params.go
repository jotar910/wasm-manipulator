package waspect

import (
	"joao/wasm-manipulator/internal/wcode"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"joao/wasm-manipulator/internal/wkeyword"
	"joao/wasm-manipulator/internal/wpointcut"
)

// pointcutParameters contains the pointcut parameters that belong to some function.
type pointcutParameters struct {
	fnDef  *wcode.FunctionDefinition
	params map[string]wpointcut.ParsedParam
}

// newPointcutParameters is a constructor for pointcutParameters.
func newPointcutParameters(fnDef *wcode.FunctionDefinition, params map[string]wpointcut.ParsedParam) *pointcutParameters {
	for _, val := range params {
		var name string

		// Find the immediate value.
		if val.Variable == "param" {
			name = parameterIndex(fnDef, val)
		} else {
			name = localIndex(fnDef, val)
		}

		fnDef.Alias[name] = val.Name
	}
	return &pointcutParameters{
		fnDef:  fnDef,
		params: params,
	}
}

// Is returns the type of keyword of some pointcut parameter.
func (er *pointcutParameters) Is(k string) wkeyword.KeywordType {
	if _, ok := er.params[k]; ok {
		return wkeyword.KeywordTypeString
	}
	return wkeyword.KeywordTypeUnknown
}

// Get returns the pointcut parameter, for a given key, as a keyword value.
func (er *pointcutParameters) Get(k string) (interface{}, wkeyword.KeywordType, bool) {
	if val, ok := er.params[k]; ok {
		var name string

		// Find the immediate value.
		if val.Variable == "param" {
			name = parameterIndex(er.fnDef, val)
		} else {
			name = localIndex(er.fnDef, val)
		}

		return name, wkeyword.KeywordTypeString, true
	}
	return nil, wkeyword.KeywordTypeUnknown, false
}

// parameterIndex returns the index value for some parameter.
func parameterIndex(fnDef *wcode.FunctionDefinition, val wpointcut.ParsedParam) string {
	if index, err := strconv.Atoi(val.Index); err == nil {
		params := fnDef.Parameters()
		if index >= len(params) {
			logrus.Fatalf("finding parameter for function %s: parameter index out of range (index: %d, maximum: %d)",
				fnDef.Name, index, len(params)-1)
		}
		return params[index].Name
	}
	index := val.Index
	if !strings.HasPrefix(index, "$") {
		indexAux, ok := fnDef.AliasKey(index)
		if !ok {
			logrus.Fatalf("finding parameter for function %s: parameter named %s not found",
				fnDef.Name, index)
		}
		index = indexAux
	}
	if param, ok := fnDef.Params[index]; ok {
		return param.Name
	}
	logrus.Fatalf("finding parameter for function %s: parameter index %s not found",
		fnDef.Name, index)
	return ""
}

// localIndex returns the index value for some local.
func localIndex(fnDef *wcode.FunctionDefinition, val wpointcut.ParsedParam) string {
	if index, err := strconv.Atoi(val.Index); err == nil {
		locals := fnDef.LocalsArr()
		if index >= len(locals) {
			logrus.Fatalf("finding local for function %s: local index out of range (index: %d, maximum: %d)",
				fnDef.Name, index, len(locals)-1)
		}
		return locals[index].Name
	}
	index := val.Index
	if !strings.HasPrefix(index, "$") {
		indexAux, ok := fnDef.AliasKey(index)
		if !ok {
			logrus.Fatalf("finding local for function %s: local named %s not found",
				fnDef.Name, index)
		}
		index = indexAux
	}
	if local, ok := fnDef.Locals[index]; ok {
		return local.Name
	}
	logrus.Fatalf("finding local for function %s: local index %s not found",
		fnDef.Name, index)
	return ""
}
