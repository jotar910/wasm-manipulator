package wgenerator

import (
	"bytes"
	"joao/wasm-manipulator/internal/wparser/variable"
	"joao/wasm-manipulator/internal/wyaml"

	"github.com/sirupsen/logrus"
)

const CodeIndexPrefix = "wmr_"

// functionTemplateIn contains the input data for the function template.
type functionTemplateIn struct {
	Name     string
	TypeName string
	Params   []*functionTemplateReferenceIn
	Result   string
	Code     string
}

// newFunctionTemplateIn is the constructor for functionTemplateIn.
func newFunctionTemplateIn(expr *wyaml.FunctionYAML, name, typeName string) *functionTemplateIn {
	var params []*functionTemplateReferenceIn
	for _, arg := range expr.Args {
		params = append(params, &functionTemplateReferenceIn{Name: "$" + arg.Name, Type: arg.Type})
	}
	return &functionTemplateIn{
		Name:     name,
		TypeName: typeName,
		Params:   params,
		Result:   expr.Result,
		Code:     expr.Code,
	}
}

// importFunctionTemplateIn contains the input data for the import-function template.
type importFunctionTemplateIn struct {
	Name       string
	TypeName   string
	ModuleName string
	ExportName string
}

// newImportFunctionTemplateIn is the constructor for importFunctionTemplateIn.
func newImportFunctionTemplateIn(name, typeName, moduleName, exportName string) *importFunctionTemplateIn {
	return &importFunctionTemplateIn{
		Name:       name,
		TypeName:   typeName,
		ModuleName: moduleName,
		ExportName: exportName,
	}
}

// exportFunctionTemplateIn contains the input data for the export-function template.
type exportFunctionTemplateIn struct {
	Name       string
	ExportName string
}

// newExportFunctionTemplateIn is the constructor for exportFunctionTemplateIn.
func newExportFunctionTemplateIn(name, exportName string) *exportFunctionTemplateIn {
	return &exportFunctionTemplateIn{
		Name:       name,
		ExportName: exportName,
	}
}

// functionTemplateReferenceIn contains the input data for the references in template template.
type functionTemplateReferenceIn struct {
	Name string
	Type string
}

// globalTemplateIn contains the input data for the global template.
type globalTemplateIn struct {
	Name  string
	Type  string
	Value string
}

// newGlobalTemplateIn is the constructor for globalTemplateIn.
func newGlobalTemplateIn(expr *variable.Expr, name string) *globalTemplateIn {
	return &globalTemplateIn{
		Name:  name,
		Type:  expr.GetType(),
		Value: expr.GetValue("0"),
	}
}

// localTemplateIn contains the input data for the local template.
type localTemplateIn struct {
	Name string
	Type string
}

// newLocalTemplateIn is the constructor for localTemplateIn.
func newLocalTemplateIn(name, typ string) *localTemplateIn {
	return &localTemplateIn{
		Name: name,
		Type: typ,
	}
}

// setLocalTemplateIn contains the input data for the set-local template.
type setLocalTemplateIn struct {
	Name  string
	Type  string
	Value string
}

// newSetLocalTemplateIn is the constructor for setLocalTemplateIn.
func newSetLocalTemplateIn(name, typ, value string) *setLocalTemplateIn {
	return &setLocalTemplateIn{
		Name:  name,
		Type:  typ,
		Value: value,
	}
}

// setLocalTemplateIn contains the input data for the set-local template.
type setLocalTemplateInstrIn struct {
	Name        string
	Instruction string
}

// newSetLocalTemplateInstrIn is the constructor for setLocalTemplateInstrIn.
func newSetLocalTemplateInstrIn(name, instruction string) *setLocalTemplateInstrIn {
	return &setLocalTemplateInstrIn{
		Name:        name,
		Instruction: instruction,
	}
}

// getVariableTemplateIn contains the input data for the get-local template.
type getVariableTemplateIn struct {
	Name    string
	IsLocal bool
}

// newGetVariableTemplateIn is the constructor for getVariableTemplateIn.
func newGetVariableTemplateIn(name string, isLocal bool) *getVariableTemplateIn {
	return &getVariableTemplateIn{Name: name, IsLocal: isLocal}
}

// typeTemplateIn contains the input data for the type template.
type typeTemplateIn struct {
	Name   string
	Result string
	Params []string
}

// newTypeTemplateIn is a constructor for typeTemplateIn.
func newTypeTemplateIn(params []string, result, name string) *typeTemplateIn {
	return &typeTemplateIn{
		Name:   name,
		Result: result,
		Params: params,
	}
}

// FunctionToCode returns the wat code containing the function.
func FunctionToCode(expr *wyaml.FunctionYAML, name, typeName string) string {
	buf := new(bytes.Buffer)
	if err := functionTemplate.Execute(buf, newFunctionTemplateIn(expr, name, typeName)); err != nil {
		logrus.Fatal(err)
	}
	return buf.String()
}

func StartFunctionToCode(name string) string {
	buf := new(bytes.Buffer)
	if err := startFunctionTemplate.Execute(buf, struct{ Name string }{name}); err != nil {
		logrus.Fatal(err)
	}
	return buf.String()
}

// ExportFunctionToCode returns the wat code containing the export function.
func ExportFunctionToCode(name, exportName string) string {
	buf := new(bytes.Buffer)
	if err := exportFunctionTemplate.Execute(buf, newExportFunctionTemplateIn(name, exportName)); err != nil {
		logrus.Fatal(err)
	}
	return buf.String()
}

// ImportFunctionToCode returns the wat code containing the import function.
func ImportFunctionToCode(name, typeName, moduleName, exportName string) string {
	buf := new(bytes.Buffer)
	if err := importFunctionTemplate.Execute(buf, newImportFunctionTemplateIn(name, typeName, moduleName, exportName)); err != nil {
		logrus.Fatal(err)
	}
	return buf.String()
}

// FunctionTypeToCode returns the wat code containing the function type.
func FunctionTypeToCode(params []string, result string, name string) string {
	buf := new(bytes.Buffer)
	if err := typeTemplate.Execute(buf, newTypeTemplateIn(params, result, name)); err != nil {
		logrus.Fatal(err)
	}
	return buf.String()
}

// GlobalVariableToCode returns the wat code containing the global variable.
func GlobalVariableToCode(expr *variable.Expr, name string) string {
	buf := new(bytes.Buffer)
	if err := globalTemplate.Execute(buf, newGlobalTemplateIn(expr, name)); err != nil {
		logrus.Fatal(err)
	}
	return buf.String()
}

// LocalVariableToCode returns the wat code containing the local variable.
func LocalVariableToCode(name string, typ string) string {
	buf := new(bytes.Buffer)
	if err := localTemplate.Execute(buf, newLocalTemplateIn(name, typ)); err != nil {
		logrus.Fatal(err)
	}
	return buf.String()
}

// SetLocalToCode returns the wat code containing the set local instruction.
func SetLocalToCode(name, typ, value string) string {
	buf := new(bytes.Buffer)
	if err := setLocalTemplate.Execute(buf, newSetLocalTemplateIn(name, typ, value)); err != nil {
		logrus.Fatal(err)
	}
	return buf.String()
}

// SetLocalToCode returns the wat code containing the set local instruction that receives instructions as values.
func SetLocalInstructionToCode(name, instruction string) string {
	buf := new(bytes.Buffer)
	if err := setLocalInstrTemplate.Execute(buf, newSetLocalTemplateInstrIn(name, instruction)); err != nil {
		logrus.Fatal(err)
	}
	return buf.String()
}

// GetVariableToCode returns the wat code containing the get variable instruction.
func GetVariableToCode(name string, isLocal bool) string {
	buf := new(bytes.Buffer)
	if err := getVariableTemplate.Execute(buf, newGetVariableTemplateIn(name, isLocal)); err != nil {
		logrus.Fatal(err)
	}
	return buf.String()
}
