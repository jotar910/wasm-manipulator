package lex

type ItemType int

const (
	ItemTypeText ItemType = iota
	ItemTypeString
	ItemTypeStringStart
	ItemTypeStringEnd
	ItemTypeLeftKeyword
	ItemTypeInsideAction
	ItemTypeRightKeyword
	ItemTypeLeftKeywordGroup
	ItemTypeRightKeywordGroup
	ItemTypeUnionKeyword
	ItemTypeAndKeyword
	ItemTypeOrKeyword
	ItemTypeEqualKeyword
	ItemTypeNotEqualKeyword
	ItemTypeBitwiseLeftKeyword
	ItemTypeBitwiseRightKeyword
	ItemTypeGreaterKeyword
	ItemTypeGreaterEqualKeyword
	ItemTypeLessKeyword
	ItemTypeLessEqualKeyword
	ItemTypePlusKeyword
	ItemTypeMinusKeyword
	ItemTypeMultiplyKeyword
	ItemTypeDivisorKeyword
	ItemTypeRemainderKeyword
	ItemTypeIdentifier
	ItemTypeNegation
	ItemTypeIndexStart
	ItemTypeIndexEnd
	ItemTypeObjectProperty
	ItemTypeMethodStart
	ItemTypeMethodEnd
	ItemTypeMethodName
	ItemTypeMethodArgStart
	ItemTypeMethodArgEnd
	ItemTypeError
	ItemTypeNumber
	ItemTypeLambdaStart
	ItemTypeLambdaEnd
	ItemTypeLambdaArg
	ItemTypeEOF
	ItemTypeUnknown
)

func (t ItemType) String() string {
	return []string{
		"ItemTypeText",
		"ItemTypeString",
		"ItemTypeStringStart",
		"ItemTypeStringEnd",
		"ItemTypeLeftKeyword",
		"ItemTypeInsideAction",
		"ItemTypeRightKeyword",
		"ItemTypeLeftKeywordGroup",
		"ItemTypeRightKeywordGroup",
		"ItemTypeUnionKeyword",
		"ItemTypeAndKeyword",
		"ItemTypeOrKeyword",
		"ItemTypeEqualKeyword",
		"ItemTypeNotEqualKeyword",
		"ItemTypeBitwiseLeftKeyword",
		"ItemTypeBitwiseRightKeyword",
		"ItemTypeGreaterKeyword",
		"ItemTypeGreaterEqualKeyword",
		"ItemTypeLessKeyword",
		"ItemTypeLessEqualKeyword",
		"ItemTypePlusKeyword",
		"ItemTypeMinusKeyword",
		"ItemTypeMultiplyKeyword",
		"ItemTypeDivisorKeyword",
		"ItemTypeRemainderKeyword",
		"ItemTypeIdentifier",
		"ItemTypeNegation",
		"ItemTypeIndexStart",
		"ItemTypeIndexEnd",
		"ItemTypeObjectProperty",
		"ItemTypeMethodStart",
		"ItemTypeMethodEnd",
		"ItemTypeMethodName",
		"ItemTypeMethodArgStart",
		"ItemTypeMethodArgEnd",
		"ItemTypeError",
		"ItemTypeNumber",
		"ItemTypeLambdaStart",
		"ItemTypeLambdaEnd",
		"ItemTypeLambdaArg",
		"ItemTypeEOF",
		"ItemTypeUnknown",
	}[t]
}

// Item represents the values dispatched from the lexer.
type Item struct {
	t ItemType
	v string
}
