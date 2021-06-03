package wcode

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
)

var (
	arrayTypeRegex = regexp.MustCompile(`\[\]\w[\w\d]*`)
	mapTypeRegex   = regexp.MustCompile(`^map\[\w[\w\d]*\](.)+`)
)

const (
	varTypeIdentifier varType = "identifier"
	varTypeI32        varType = "i32"
	varTypeF32        varType = "f32"
	varTypeString     varType = "string"
	varTypeMap        varType = "map"
	varTypeArray      varType = "array"
	varTypeI64        varType = "i64"
	varTypeF64        varType = "f64"
)

var varTypeCodes map[string]int

// varTypeCode returns the code for some type.
func varTypeCode(typ string) int {
	if varTypeCodes == nil {
		fillVarTypeCodes()
	}
	if code, ok := varTypeCodes[typ]; ok {
		return code
	}
	return 0
}

// fillVarTypeCodes initializes the variable type codes data.
func fillVarTypeCodes() {
	types := []string{
		string(varTypeIdentifier),
		string(varTypeI32),
		string(varTypeF32),
		string(varTypeF64),
		string(varTypeString),
		fmt.Sprintf("%s_%s", varTypeMap, varTypeI32),
		fmt.Sprintf("%s_%s", varTypeMap, varTypeF32),
		fmt.Sprintf("%s_%s", varTypeMap, varTypeF64),
		fmt.Sprintf("%s_%s", varTypeMap, varTypeString),
		string(varTypeArray),
	}
	varTypeCodes = make(map[string]int)
	for i, t := range types {
		varTypeCodes[t] = i + 1
	}
}

// varType represents the type designation for variables.
type varType string

// variableType contains the type definition for variables.
type variableType interface {
	fmt.Stringer
	Type() varType
	Label() string
	Code() int
}

// newVariableType is a constructor for variableType.
func newVariableType(typeStr string) (variableType, error) {
	switch typeVar := varType(typeStr); {
	case typeVar == varTypeIdentifier, typeVar == varTypeString, typeVar == varTypeI32, typeVar == varTypeF32, typeVar == varTypeF64:
		return newSimpleType(typeVar), nil
	case mapTypeRegex.MatchString(typeStr):
		var keyStart, keyEnd int
		for pos, char := range typeStr {
			if char == ']' {
				keyEnd = pos
				break
			}
			if char == '[' {
				keyStart = pos
			}
		}
		keyType := typeStr[keyStart+1 : keyEnd]
		if !isVarSimpleTypeValid(keyType) {
			return nil, fmt.Errorf("key type %s is invalid for map", keyType)
		}
		valueType := typeStr[keyEnd+1:]
		subType, err := newVariableType(valueType)
		if err != nil {
			return nil, fmt.Errorf("value type %s is invalid for map: %w", valueType, err)
		}
		return newMapType(newSimpleType(varType(keyType)), subType), nil
	case arrayTypeRegex.MatchString(typeStr):
		valueType := typeStr[2:]
		subType, err := newVariableType(valueType)
		if err != nil {
			return nil, fmt.Errorf("value type %s is invalid for array: %w", valueType, err)
		}
		return newArrayType(subType), nil
	}
	return nil, fmt.Errorf("type %s is not valid", typeStr)
}

// simpleType is for variables with a simple type.
type simpleType struct {
	value varType
}

// newSimpleType is a constructor for simpleType.
func newSimpleType(t varType) *simpleType {
	return &simpleType{t}
}

// String returns the string value for the type.
func (t simpleType) String() string {
	return fmt.Sprint(t.value)
}

// Type returns the base type.
func (t simpleType) Type() varType {
	return t.value
}

// Label returns the type label.
func (t simpleType) Label() string {
	return t.String()
}

// Code returns the type code.
func (t simpleType) Code() int {
	return varTypeCode(t.Label())
}

// arrayType is for variables of type array.
type arrayType struct {
	value variableType
}

// newArrayType is a constructor for arrayType.
func newArrayType(v variableType) *arrayType {
	return &arrayType{v}
}

// String returns the string value for the type.
func (t arrayType) String() string {
	return fmt.Sprintf("[]%s", t.value)
}

// Type returns the base type.
func (t arrayType) Type() varType {
	return varTypeArray
}

// Label returns the type label.
func (t arrayType) Label() string {
	return string(t.Type())
}

// Code returns the type code.
func (t arrayType) Code() int {
	return getComplexCode(t.value.Code(), t.Label)
}

// mapType is for variables of type map.
type mapType struct {
	key   variableType
	value variableType
}

// newMapType is a constructor for mapType.
func newMapType(k, v variableType) *mapType {
	return &mapType{k, v}
}

// String returns the string value for the type.
func (t mapType) String() string {
	return fmt.Sprintf("map[%s]%s", t.key, t.value)
}

// Type returns the base type.
func (t mapType) Type() varType {
	return varTypeMap
}

// Label returns the type label.
func (t mapType) Label() string {
	return fmt.Sprintf("%s_%s", t.Type(), t.key.Label())
}

// Code returns the type code.
func (t mapType) Code() int {
	return getComplexCode(t.value.Code(), t.Label)
}

// IsVarTypeStrPrimitive returns if the variable in string format has a primitive type.
func IsVarTypeStrPrimitive(t string) bool {
	return isVarTypePrimitive(varType(t))
}

// isVarTypeStrValid returns if the variable type in string format is valid.
func isVarTypeStrValid(t string) bool {
	return isVarTypeValid(varType(t))
}

// isVarTypeValid returns if the variable type is valid.
func isVarTypeValid(vt varType) bool {
	t := string(vt)
	switch {
	case vt == varTypeI32, vt == varTypeF32, vt == varTypeF64, vt == varTypeString, vt == varTypeIdentifier,
		arrayTypeRegex.MatchString(t), mapTypeRegex.MatchString(t):
		return true
	default:
		return false
	}
}

// isVarSimpleTypeValid returns if the variable has a simple type.
func isVarSimpleTypeValid(t string) bool {
	switch vt := varType(t); {
	case IsVarTypeStrPrimitive(t), vt == varTypeString, vt == varTypeIdentifier:
		return true
	default:
		return false
	}
}

// isVarTypePrimitive returns if the variable has a primitive type.
func isVarTypePrimitive(t varType) bool {
	switch t {
	case varTypeI32, varTypeF32, varTypeI64, varTypeF64:
		return true
	default:
		return false
	}
}

// getComplexCode returns the complex type code.
func getComplexCode(subTypeCode int, labelFn func()string) int {
	minSubTypeShiftV := strconv.FormatInt(int64(len(varTypeCodes)), 2)
	minSubTypeShift := len(minSubTypeShiftV)

	subTypeShift := len(strconv.FormatInt(int64(subTypeCode), 2))

	if subTypeShift%minSubTypeShift != 0 {
		subTypeShift += minSubTypeShift - subTypeShift%minSubTypeShift
	}
	subTypeShift = int(math.Max(float64(subTypeShift), float64(minSubTypeShift)))
	thisLabel := labelFn()
	thisCode := varTypeCode(thisLabel)
	return (thisCode << subTypeShift) | subTypeCode
}
