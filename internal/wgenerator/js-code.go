package wgenerator

import (
	"bytes"
	"errors"
	"text/template"

	_ "embed"
)

var (
	ErrorUnnecessary = errors.New("unnecessary javascript creating")
)

//go:embed resources/template-js.go.txt
var jsTemplateStr string
var jsTemplate *template.Template

// jsTemplateIn contains the input data for the javascript template.
type jsTemplateIn struct {
	InternalFns []jsInternalFnIn
	ExternalFns []jsExternalFnIn
}

// newJsTemplateIn is a constructor for jsTemplateIn.
func newJsTemplateIn(internalFns []jsInternalFnIn, externalFns []jsExternalFnIn) jsTemplateIn {
	return jsTemplateIn{internalFns, externalFns}
}

// jsInternalFnIn contains the input data for the javascript internal function.
type jsInternalFnIn struct {
	Name            string
	OpName          string
	Args            []jsInternalParamIn
	NArgs           int
	NCompositeArgs  int
	CompositeReturn bool
}

// jsInternalParamIn constains the input data for the javascript internal function parameter.
type jsInternalParamIn struct {
	Index       int
	IsPrimitive bool
}

// jsExternalFnIn contains the input data for the javascript external function.
type jsExternalFnIn struct {
	Name            string
	OpName          string
	CompositeArgs   []jsExternalCompositeParamIn
	PrimitiveArgs   []jsExternalParamIn
	CompositeReturn bool
}

// jsExternalParamIn constains the input data for the javascript external function parameter.
type jsExternalParamIn struct {
	Index int
}

// jsExternalCompositeParamIn constains the input data for the javascript external function parameter of type composite.
type jsExternalCompositeParamIn struct {
	jsExternalParamIn
	Code int
}

// JsFunctionDefinition contains the definition data for some internal/external function.
type JsFunctionDefinition struct {
	Name            string
	OpName          string
	Scope           JsScopeType
	Args            []JsArgumentDefinition
	CompositeReturn bool
}

// NewJsFunctionDefinition is a constructor for JsFunctionDefinition.
func NewJsFunctionDefinition(opName, name string, scope JsScopeType, args []JsArgumentDefinition, compositeReturn bool) JsFunctionDefinition {
	return JsFunctionDefinition{OpName: opName, Name: name, Scope: scope, Args: args, CompositeReturn: compositeReturn}
}

// IsImported returns if the function is imported (otherwise is exported).
func (fnDef JsFunctionDefinition) IsImported() bool {
	return fnDef.Scope == JsScopeTypeImported
}

// JsArgumentDefinition contains the definition data for some function argument.
type JsArgumentDefinition struct {
	Code int
	Type JsArgumentType
}

// NewJsArgumentDefinition is a constructor for JsArgumentDefinition.
func NewJsArgumentDefinition(code int, isPrimitive bool) JsArgumentDefinition {
	if isPrimitive {
		return JsArgumentDefinition{Code: code, Type: JsArgumentTypePrimitive}
	}
	return JsArgumentDefinition{Code: code, Type: JsArgumentTypeComposite}
}

// IsPrimitive returns if the argument type is primitive.
func (argDef JsArgumentDefinition) IsPrimitive() bool {
	return argDef.Type == JsArgumentTypePrimitive
}

// JsArgumentType represents the argument type.
type JsArgumentType int

const (
	JsArgumentTypePrimitive JsArgumentType = iota
	JsArgumentTypeComposite
)

// JsArgumentType represents the function scope.
type JsScopeType int

const (
	JsScopeTypeImported JsScopeType = iota
	JsScopeTypeExported
)

// GetJsCode returns the javascript code for some definition.
func GetJsCode(functions []JsFunctionDefinition) (string, error) {
	var internalFns []jsInternalFnIn
	var externalFns []jsExternalFnIn
	for _, fn := range functions {
		if fn.IsImported() {
			if internalFn, ok := resolveJsInternal(fn); ok {
				internalFns = append(internalFns, internalFn)
			}
			continue
		}
		if externalFn, ok := resolveJsExternal(fn); ok {
			externalFns = append(externalFns, externalFn)
		}
	}
	buf := new(bytes.Buffer)
	if err := jsTemplate.Execute(buf, newJsTemplateIn(internalFns, externalFns)); err != nil {
		return "", err
	}
	if len(internalFns) == 0 && len(externalFns) == 0 {
		return buf.String(), ErrorUnnecessary
	}
	return buf.String(), nil
}

// resolveJsInternal resolves internal function.
func resolveJsInternal(fn JsFunctionDefinition) (jsInternalFnIn, bool) {
	res := jsInternalFnIn{
		Name:            fn.Name,
		OpName:          fn.OpName,
		CompositeReturn: fn.CompositeReturn,
		NArgs:           len(fn.Args),
	}
	var accumPrimitive int
	for i, arg := range fn.Args {
		isPrimitive := arg.IsPrimitive()
		if !isPrimitive {
			res.Args = append(res.Args, jsInternalParamIn{i, isPrimitive})
			res.NCompositeArgs++
		} else {
			res.Args = append(res.Args, jsInternalParamIn{accumPrimitive, isPrimitive})
			accumPrimitive++
		}
	}
	return res, res.NCompositeArgs > 0 || res.CompositeReturn
}

// resolveJsExternal resolves external function.
func resolveJsExternal(fn JsFunctionDefinition) (jsExternalFnIn, bool) {
	res := jsExternalFnIn{
		Name:            fn.Name,
		OpName:          fn.OpName,
		CompositeReturn: fn.CompositeReturn,
	}
	for i, arg := range fn.Args {
		isPrimitive := arg.IsPrimitive()
		index := jsExternalParamIn{i}
		if isPrimitive {
			res.PrimitiveArgs = append(res.PrimitiveArgs, index)
			continue
		}
		res.CompositeArgs = append(res.CompositeArgs,
			jsExternalCompositeParamIn{index, arg.Code})
	}
	return res, len(res.CompositeArgs) > 0 || res.CompositeReturn
}
