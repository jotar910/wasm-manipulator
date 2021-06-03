package wgenerator

import (
	"bytes"
	"fmt"

	"github.com/sirupsen/logrus"
)

const bytesOn32bits = 4

var (
	operationFunctions = []*ImportFunctionDef{
		newOperationsFunction("clear", ""),
		newOperationsFunction("new_code", "", "i32"),
		newOperationsFunction("write", "", "i32"),
		newOperationsFunction("write_f32", "", "f32"),
		newOperationsFunction("write_f64", "", "f64"),
		newOperationsFunction("read", "i32", "i32"),
		newOperationsFunction("read_f32", "f32", "i32"),
		newOperationsFunction("read_f64", "f64", "i32"),
		newOperationsFunction("evaluate", "i32"),
	}
	argsFunctions = []*ImportFunctionDef{
		newArgsFunction("push", ""),
		newArgsFunction("pop", ""),
		newArgsFunction("new", "", "i32"),
		newArgsFunction("write", "", "i32"),
		newArgsFunction("write_f32", "", "f32"),
		newArgsFunction("write_f64", "", "f64"),
		newArgsFunction("new_copy", ""),
		newArgsFunction("copy_index", "", "i32"),
		newArgsFunction("copy_operation", "", "i32"),
	}
	zoneFunctions = []*ImportFunctionDef{
		newZoneFunction("push", ""),
		newZoneFunction("pop", ""),
		newZoneFunction("new", "", "i32"),
		newZoneFunction("write_name", "", "i32"),
		newZoneFunction("write_key", "", "i32"),
		newZoneFunction("write_value", "", "i32"),
		newZoneFunction("write_value_f32", "", "f32"),
		newZoneFunction("write_value_f64", "", "f64"),
		newZoneFunction("new_copy", ""),
		newZoneFunction("copy_name", "", "i32"),
		newZoneFunction("copy_key", "", "i32"),
		newZoneFunction("copy_arg", "", "i32"),
		newZoneFunction("copy_operation", "", "i32"),
		newZoneFunction("copy_operation_global", "", "i32"),
		newZoneFunction("set", ""),
		newZoneFunction("set_global", ""),
	}
	returnsFunctions = []*ImportFunctionDef{
		newReturnsFunction("new_copy", ""),
		newReturnsFunction("copy_name", "", "i32"),
		newReturnsFunction("copy_key", "", "i32"),
		newReturnsFunction("copy_var", ""),
		newReturnsFunction("copy_operation", "", "i32"),
	}
	errorFunctions = []*ImportFunctionDef{
		newErrorFunction("new", ""),
		newErrorFunction("set", "", "i32"),
		newErrorFunction("print", ""),
	}
)

// OperationFunctions returns the list of operation functions.
func OperationFunctions() []*ImportFunctionDef {
	return append([]*ImportFunctionDef{}, operationFunctions...)
}

// ArgsFunctions returns the list of args functions.
func ArgsFunctions() []*ImportFunctionDef {
	return append([]*ImportFunctionDef{}, argsFunctions...)
}

// ZoneFunctions returns the list of zone functions.
func ZoneFunctions() []*ImportFunctionDef {
	return append([]*ImportFunctionDef{}, zoneFunctions...)
}

// ErrorFunctions returns the list of returns functions.
func ReturnsFunctions() []*ImportFunctionDef {
	return append([]*ImportFunctionDef{}, returnsFunctions...)
}

// ErrorFunctions returns the list of error functions.
func ErrorFunctions() []*ImportFunctionDef {
	return append([]*ImportFunctionDef{}, errorFunctions...)
}

// ImportFunctionDef contains the definition data for imported functions.
type ImportFunctionDef struct {
	ModuleName   string
	ExportedName string
	Params       []string
	Result       string
}

// newImportFunction is a constructor for ImportFunctionDef.
func newImportFunction(moduleName, exportedName string, params []string, result string) *ImportFunctionDef {
	return &ImportFunctionDef{moduleName, exportedName, params, result}
}

// newOperationsFunction is a constructor for ImportFunctionDef.
// creates an import function of type operations.
func newOperationsFunction(exportedName string, result string, params ...string) *ImportFunctionDef {
	return newImportFunction("operations", exportedName, params, result)
}

// newArgsFunction is a constructor for ImportFunctionDef.
// creates an import function of type args.
func newArgsFunction(exportedName string, result string, params ...string) *ImportFunctionDef {
	return newImportFunction("args", exportedName, params, result)
}

// newZoneFunction is a constructor for ImportFunctionDef.
// creates an import function of type zone.
func newZoneFunction(exportedName string, result string, params ...string) *ImportFunctionDef {
	return newImportFunction("zone", exportedName, params, result)
}

// newReturnsFunction is a constructor for ImportFunctionDef.
// creates an import function of type returns.
func newReturnsFunction(exportedName string, result string, params ...string) *ImportFunctionDef {
	return newImportFunction("returns", exportedName, params, result)
}

// newErrorFunction is a constructor for ImportFunctionDef.
// creates an import function of type error.
func newErrorFunction(exportedName string, result string, params ...string) *ImportFunctionDef {
	return newImportFunction("error", exportedName, params, result)
}

// primitiveEvalTemplateIn contains the input data for the evaluation of primitives.
type primitiveEvalTemplateIn struct {
	EvalExpr  []int32
	Type      string
	LocalName string
	BlockName string
}

// newPrimitiveEvalTemplateIn is a constructor for primitiveEvalTemplateIn.
func newPrimitiveEvalTemplateIn(expr, local, typ, block string) *primitiveEvalTemplateIn {
	return &primitiveEvalTemplateIn{
		EvalExpr:  stringToInt32(expr),
		Type:      typ,
		LocalName: local,
		BlockName: block,
	}
}

// compositeCallEvalTemplateIn contains the input data for the evaluation of composite types on a call.
type compositeCallEvalTemplateIn struct {
	EvalExpr []int32
	ArgIndex int
	PushArgs bool
}

// newCompositeCallEvalTemplateIn is a constructor for compositeEvalTemplateIn.
func newCompositeCallEvalTemplateIn(expr string, index int, pushArgs bool) *compositeCallEvalTemplateIn {
	return &compositeCallEvalTemplateIn{
		EvalExpr: stringToInt32(expr),
		ArgIndex: index,
		PushArgs: pushArgs,
	}
}

// compositeReturnEvalTemplateIn contains the input data for the evaluation of composite types on a return.
type compositeReturnEvalTemplateIn struct {
	EvalExpr []int32
}

// newCompositeReturnEvalTemplateIn is a constructor for compositeReturnEvalTemplateIn.
func newCompositeReturnEvalTemplateIn(expr string) *compositeReturnEvalTemplateIn {
	return &compositeReturnEvalTemplateIn{
		EvalExpr: stringToInt32(expr),
	}
}

// compositeReturnEvalTemplateIn contains the input data for the evaluation reference of composite types on a return.
type compositeReturnEvalRefTemplateIn struct {
	Name []int32
	Key  []int32
}

// newCompositeReturnEvalRefTemplateIn is a constructor for compositeReturnEvalRefTemplateIn.
func newCompositeReturnEvalRefTemplateIn(name, key string) *compositeReturnEvalRefTemplateIn {
	return &compositeReturnEvalRefTemplateIn{
		Name: stringToInt32(name),
		Key:  stringToInt32(key),
	}
}

// compositeZoneEvalTemplateIn contains the input data for the evaluation of composite zones.
type compositeZoneEvalTemplateIn struct {
	Name      []int32
	Key       []int32
	EvalExpr  []int32
	BlockName string
	LoopName  string
	IsLocal   bool
}

// compositeZoneEvalTemplateIn is a constructor for compositeZoneEvalTemplateIn.
func newCompositeZoneEvalTemplateIn(name, key, expr, block, loop string, isLocal bool) *compositeZoneEvalTemplateIn {
	return &compositeZoneEvalTemplateIn{
		Name:      stringToInt32(name),
		Key:       stringToInt32(key),
		EvalExpr:  stringToInt32(expr),
		BlockName: block,
		LoopName:  loop,
		IsLocal:   isLocal,
	}
}

// setStartingGlobalCompositeTemplateIn contains the input data for the composite global initialization.
type setStartingGlobalCompositeTemplateIn struct {
	Name     []int32
	Value    []int32
	TypeCode int
}

// newSetStartingGlobalCompositeTemplateIn is a constructor for setStartingGlobalCompositeTemplateIn.
func newSetStartingGlobalCompositeTemplateIn(name, value string, typeCode int) *setStartingGlobalCompositeTemplateIn {
	return &setStartingGlobalCompositeTemplateIn{
		Name:     stringToInt32(name),
		Value:    stringToInt32(value),
		TypeCode: typeCode,
	}
}

// zonePrimitiveTemplateIn contains the input data for the evaluation of primitive zones.
type zonePrimitiveTemplateIn struct {
	Name     []int32
	Index    string
	Type     string
	TypeCode int
	IsLocal  bool
}

// newZonePrimitiveTemplateIn is a constructor for zonePrimitiveTemplateIn.
func newZonePrimitiveTemplateIn(name, index, typeName string, typeCode int, local bool) *zonePrimitiveTemplateIn {
	return &zonePrimitiveTemplateIn{
		Name:     stringToInt32(name),
		Index:    index,
		Type:     typeName,
		TypeCode: typeCode,
		IsLocal:  local,
	}
}

// zoneLocalCompositeTemplateIn contains the input data for the composite local initialization.
type zoneLocalCompositeTemplateIn struct {
	LocalName  []int32
	LocalKey   []int32
	LocalValue []int32
	Type       int
}

// newZoneLocalCompositeTemplateIn is a constructor for zoneLocalCompositeTemplateIn.
func newZoneLocalCompositeTemplateIn(name, key, value string, typeCode int) *zoneLocalCompositeTemplateIn {
	return &zoneLocalCompositeTemplateIn{
		LocalName:  stringToInt32(name),
		LocalKey:   stringToInt32(key),
		LocalValue: stringToInt32(value),
		Type:       typeCode,
	}
}

// zoneParamCompositeTemplateIn contains the input data for the composite parameter template.
type zoneParamCompositeTemplateIn struct {
	ParamName  []int32
	ParamIndex string
	BlockName  string
	LoopName   string
}

// newZoneParamCompositeTemplateIn is a constructor for zoneParamCompositeTemplateIn.
func newZoneParamCompositeTemplateIn(name, index, blockName, loopName string) *zoneParamCompositeTemplateIn {
	return &zoneParamCompositeTemplateIn{
		ParamName:  stringToInt32(name),
		ParamIndex: index,
		BlockName:  blockName,
		LoopName:   loopName,
	}
}

// GetPrimitiveEvalStartCode returns the code to add before a primitive variable operation on an evaluation.
func GetPrimitiveEvalStartCode(expression, localName, localType string, opCount int) (string, int) {
	buf := new(bytes.Buffer)
	if err := primitiveEvalTemplateStart.Execute(buf,
		newPrimitiveEvalTemplateIn(expression, localName, localType, uniqueBlockName(opCount)),
	); err != nil {
		logrus.Fatal(err)
	}
	return buf.String(), opCount + 1
}

// GetCompositeCallEvalStartCode returns the code to add before a call evaluation.
func GetCompositeCallEvalStartCode(expression string, index int, shouldPushArgs bool, opCount int) (string, int) {
	buf := new(bytes.Buffer)
	if err := compositeCallEvalTemplateStart.Execute(buf,
		newCompositeCallEvalTemplateIn(expression, index, shouldPushArgs),
	); err != nil {
		logrus.Fatal(err)
	}
	return buf.String(), opCount
}

// GetCompositeReturnEvalCode returns the code to add on a composite return evaluation.
func GetCompositeReturnEvalCode(expression string, opCount int) (string, int) {
	buf := new(bytes.Buffer)
	if err := compositeReturnEvalTemplate.Execute(buf,
		newCompositeReturnEvalTemplateIn(expression),
	); err != nil {
		logrus.Fatal(err)
	}
	return buf.String(), opCount
}

// GetCompositeReturnEvalRefCode returns the code to add on a composite return evaluation reference.
func GetCompositeReturnEvalRefCode(name, key string, opCount int) (string, int) {
	buf := new(bytes.Buffer)
	if err := compositeReturnEvalRefTemplate.Execute(buf,
		newCompositeReturnEvalRefTemplateIn(name, key),
	); err != nil {
		logrus.Fatal(err)
	}
	return buf.String(), opCount
}

// GetCompositeCallEvalEndCode returns the code to add after a call evaluation.
func GetCompositeCallEvalEndCode(shouldPopArgs bool, opCount int) (string, int) {
	buf := new(bytes.Buffer)
	if err := compositeCallEvalTemplateEnd.Execute(buf, struct{ PopArgs bool }{shouldPopArgs}); err != nil {
		logrus.Fatal(err)
	}
	return buf.String(), opCount
}

// GetCompositeZoneEvalStartCode returns the code to add before using a variable in evaluation.
func GetCompositeZoneEvalStartCode(name, key, expression string, isLocal bool, opCount int) (string, int) {
	buf := new(bytes.Buffer)
	if err := compositeZoneEvalTemplateStart.Execute(buf,
		newCompositeZoneEvalTemplateIn(name, key, expression, uniqueBlockName(opCount), uniqueLoopName(opCount+1), isLocal),
	); err != nil {
		logrus.Fatal(err)
	}
	return buf.String(), opCount + 2
}

// GetSetStartingGlobalCompositeCode returns the initialization code for globals of type composite.
func GetSetStartingGlobalCompositeCode(name, value string, typeCode int, opCount int) (string, int) {
	buf := new(bytes.Buffer)
	if err := setStartingGlobalCompositeTemplate.Execute(buf, newSetStartingGlobalCompositeTemplateIn(name, value, typeCode)); err != nil {
		logrus.Fatal(err)
	}
	return buf.String(), opCount
}

// GetZonePrimitiveCode returns the initialization code for primitive variables.
func GetZonePrimitiveCode(name, index, typeName string, typeCode int, local bool, opCount int) (string, int) {
	buf := new(bytes.Buffer)
	if err := zonePrimitiveTemplate.Execute(buf, newZonePrimitiveTemplateIn(name, index, typeName, typeCode, local)); err != nil {
		logrus.Fatal(err)
	}
	return buf.String(), opCount
}

// GetZoneCompositeLocalCode returns the initialization code for locals of type composite.
func GetZoneCompositeLocalCode(name, key, value string, typeCode, opCount int) (string, int) {
	buf := new(bytes.Buffer)
	if err := zoneLocalCompositeTemplate.Execute(buf, newZoneLocalCompositeTemplateIn(name, key, value, typeCode)); err != nil {
		logrus.Fatal(err)
	}
	return buf.String(), opCount
}

// GetZoneCompositeParamCode returns the initialization code for parameters of type composite.
func GetZoneCompositeParamCode(name, index string, opCount int) (string, int) {
	buf := new(bytes.Buffer)
	if err := zoneParamCompositeTemplate.Execute(buf,
		newZoneParamCompositeTemplateIn(name, index, uniqueBlockName(opCount), uniqueLoopName(opCount+1)),
	); err != nil {
		logrus.Fatal(err)
	}
	return buf.String(), opCount + 2
}

// uniqueBlockName returns an unique block name for an operation.
func uniqueBlockName(opCount int) string {
	return fmt.Sprintf("$%sB%d", CodeIndexPrefix, opCount)
}

// uniqueLoopName returns an unique loop name for an operation.
func uniqueLoopName(opCount int) string {
	return fmt.Sprintf("$%sL%d", CodeIndexPrefix, opCount)
}

// stringToInt32 converts a string into an array of 32-bit integers.
func stringToInt32(expr string) []int32 {
	if len(expr) == 0 {
		return []int32{}
	}
	b := []byte(expr)
	lenBytes := len(b)
	for i := 0; i < (bytesOn32bits-(lenBytes%bytesOn32bits))%bytesOn32bits; i++ {
		b = append(b, 0)
	}
	lenBytes = len(b)
	var exprInt32 []int32
	for i := 0; i < lenBytes; i += 4 {
		lb, ls, rb, rs := int32(b[i]), int32(b[i+1]), int32(b[i+2]), int32(b[i+3])
		intValue := lb<<24 | (ls&0xff)<<16 | (rb&0xff)<<8 | (rs & 0xff)
		exprInt32 = append(exprInt32, intValue)
	}
	return exprInt32
}
