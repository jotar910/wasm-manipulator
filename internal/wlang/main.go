package wlang

var codeBlockDefinition map[string]CodeBlockDefinition

// CodeBlockDefinition represents an instructions signature.
type CodeBlockDefinition struct {
	Name     string
	Args     []CodeBlockType
	NArgs    int
	Returns  []CodeBlockType
	NReturns int
}

// GetInstrDefinition returns the instruction definition for the provided name.
func GetInstrDefinition(name string) (CodeBlockDefinition, bool) {
	res, ok := codeBlockDefinition[name]
	return res, ok
}

// IsControlFlow returns if the instructions name is of type control flow.
func IsControlFlow(name string) bool {
	m := map[string]struct{}{
		CodeBlockNameBlock:       {},
		CodeBlockNameLoop:        {},
		CodeBlockNameThen:        {},
		CodeBlockNameElse:        {},
		CodeBlockNameUnreachable: {},
		CodeBlockNameDrop:        {},
	}
	_, ok := m[name]
	return ok
}

// init is the initial function for this file.
// initiates all the code block definitions.
func init() {
	codeBlockDefinition = make(map[string]CodeBlockDefinition)
	// Control Flow Instructions
	addCodeBlockDefinition(CodeBlockNameBlock, list(), list())
	addCodeBlockDefinition(CodeBlockNameLoop, list(), list())
	addCodeBlockDefinition(CodeBlockNameBr, list(Any), list())
	addCodeBlockDefinition(CodeBlockNameBrIf, list(Any), list())
	addCodeBlockDefinition(CodeBlockNameBrTable, list(Any, I32), list())
	addCodeBlockDefinition(CodeBlockNameIf, list(I32, Any, Any), list())
	addCodeBlockDefinition(CodeBlockNameThen, list(Any), list())
	addCodeBlockDefinition(CodeBlockNameElse, list(Any), list())
	addCodeBlockDefinition(CodeBlockNameEnd, list(Any), list())
	addCodeBlockDefinition(CodeBlockNameReturn, list(Any), list(Any))
	addCodeBlockDefinition(CodeBlockNameUnreachable, list(), list())

	// Basic Instructions
	addCodeBlockDefinition(CodeBlockNameNop, list(), list())
	addCodeBlockDefinition(CodeBlockNameDrop, list(Any), list())
	addCodeBlockDefinition(CodeBlockNameI32Const, list(), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Const, list(), list(I64))
	addCodeBlockDefinition(CodeBlockNameF32Const, list(), list(F32))
	addCodeBlockDefinition(CodeBlockNameF64Const, list(), list(F64))
	addCodeBlockDefinition(CodeBlockNameGetLocal, list(), list(Any))
	addCodeBlockDefinition(CodeBlockNameSetLocal, list(Any), list())
	addCodeBlockDefinition(CodeBlockNameTeeLocal, list(Any), list(Any))
	addCodeBlockDefinition(CodeBlockNameGetGlobal, list(), list(Any))
	addCodeBlockDefinition(CodeBlockNameSetGlobal, list(Any), list())
	addCodeBlockDefinition(CodeBlockNameSelect, list(Any, Any, I32), list(Any))
	addCodeBlockDefinition(CodeBlockNameCall, list(), list())         // Depends on the function called (nArgs: N, nReturns: M)
	addCodeBlockDefinition(CodeBlockNameCallIndirect, list(), list()) // Depends on the function called (nArgs: N + 1, nReturns: M)

	// Integer Arithmetic Instructions
	addCodeBlockDefinition(CodeBlockNameI32Add, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Add, list(I64, I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32Sub, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Sub, list(I64, I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32Mul, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Mul, list(I64, I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32DivS, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64DivS, list(I64, I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32DivU, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64DivU, list(I64, I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32RemS, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64RemS, list(I64, I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32RemU, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64RemU, list(I64, I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32And, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64And, list(I64, I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32Or, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Or, list(I64, I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32Xor, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Xor, list(I64, I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32Shl, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Shl, list(I64, I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32ShrS, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64ShrS, list(I64, I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32ShrU, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64ShrU, list(I64, I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32Rotl, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Rotl, list(I64, I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32Rotr, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Rotr, list(I64, I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32Clz, list(I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Clz, list(I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32Ctz, list(I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Ctz, list(I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32Popcnt, list(I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Popcnt, list(I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameI32Eqz, list(I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Eqz, list(I64), list(I64))

	// Floating-Point Arithmetic Instructions
	addCodeBlockDefinition(CodeBlockNameF32Div, list(F32, F32), list(F32))
	addCodeBlockDefinition(CodeBlockNameF64Div, list(F64, F64), list(F64))
	addCodeBlockDefinition(CodeBlockNameF32Sqrt, list(F32), list(F32))
	addCodeBlockDefinition(CodeBlockNameF64Sqrt, list(F64), list(F64))
	addCodeBlockDefinition(CodeBlockNameF32Min, list(F32, F32), list(F32))
	addCodeBlockDefinition(CodeBlockNameF64Min, list(F64, F64), list(F64))
	addCodeBlockDefinition(CodeBlockNameF32Max, list(F32, F32), list(F32))
	addCodeBlockDefinition(CodeBlockNameF64Max, list(F64, F64), list(F64))
	addCodeBlockDefinition(CodeBlockNameF32Ceil, list(F32), list(F32))
	addCodeBlockDefinition(CodeBlockNameF64Ceil, list(F64), list(F64))
	addCodeBlockDefinition(CodeBlockNameF32Floor, list(F32), list(F32))
	addCodeBlockDefinition(CodeBlockNameF64Floor, list(F64), list(F64))
	addCodeBlockDefinition(CodeBlockNameF32Trunc, list(F32), list(F32))
	addCodeBlockDefinition(CodeBlockNameF64Trunc, list(F64), list(F64))
	addCodeBlockDefinition(CodeBlockNameF32Nearest, list(F32), list(F32))
	addCodeBlockDefinition(CodeBlockNameF64Nearest, list(F64), list(F64))
	addCodeBlockDefinition(CodeBlockNameF32Abs, list(F32), list(F32))
	addCodeBlockDefinition(CodeBlockNameF64Abs, list(F64), list(F64))
	addCodeBlockDefinition(CodeBlockNameF32Neg, list(F32), list(F32))
	addCodeBlockDefinition(CodeBlockNameF64Neg, list(F64), list(F64))
	addCodeBlockDefinition(CodeBlockNameF32Copysign, list(F32, F32), list(F32))
	addCodeBlockDefinition(CodeBlockNameF64Copysign, list(F64, F64), list(F64))

	// Integer Comparison Instructions
	addCodeBlockDefinition(CodeBlockNameI32Eq, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Eq, list(I64, I64), list(I32))
	addCodeBlockDefinition(CodeBlockNameI32Ne, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Ne, list(I64, I64), list(I32))
	addCodeBlockDefinition(CodeBlockNameI32LtS, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64LtS, list(I64, I64), list(I32))
	addCodeBlockDefinition(CodeBlockNameI32LtU, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64LtU, list(I64, I64), list(I32))
	addCodeBlockDefinition(CodeBlockNameI32LeS, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64LeS, list(I64, I64), list(I32))
	addCodeBlockDefinition(CodeBlockNameI32LeU, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64LeU, list(I64, I64), list(I32))
	addCodeBlockDefinition(CodeBlockNameI32GtS, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64GtS, list(I64, I64), list(I32))
	addCodeBlockDefinition(CodeBlockNameI32GtU, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64GtU, list(I64, I64), list(I32))
	addCodeBlockDefinition(CodeBlockNameI32GeS, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64GeS, list(I64, I64), list(I32))
	addCodeBlockDefinition(CodeBlockNameI32GeU, list(I32, I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64GeU, list(I64, I64), list(I32))

	// Floating-Point Comparison Instructions
	addCodeBlockDefinition(CodeBlockNameF32Lt, list(F32, F32), list(I32))
	addCodeBlockDefinition(CodeBlockNameF64Lt, list(F64, F64), list(I32))
	addCodeBlockDefinition(CodeBlockNameF32Le, list(F32, F32), list(I32))
	addCodeBlockDefinition(CodeBlockNameF64Le, list(F64, F64), list(I32))
	addCodeBlockDefinition(CodeBlockNameF32Gt, list(F32, F32), list(I32))
	addCodeBlockDefinition(CodeBlockNameF64Gt, list(F64, F64), list(I32))
	addCodeBlockDefinition(CodeBlockNameF32Ge, list(F32, F32), list(I32))
	addCodeBlockDefinition(CodeBlockNameF64Ge, list(F64, F64), list(I32))

	// Conversion Instructions
	addCodeBlockDefinition(CodeBlockNameWrapI64, list(I64), list(I32))
	addCodeBlockDefinition(CodeBlockNameExtendI32S, list(I32), list(I64))
	addCodeBlockDefinition(CodeBlockNameExtendI32U, list(I32), list(I64))
	addCodeBlockDefinition(CodeBlockNameTruncI32F32S, list(F32), list(I32))
	addCodeBlockDefinition(CodeBlockNameTruncI32F64S, list(F64), list(I32))
	addCodeBlockDefinition(CodeBlockNameTruncI64F32S, list(F32), list(I64))
	addCodeBlockDefinition(CodeBlockNameTruncI64F64S, list(F64), list(I64))
	addCodeBlockDefinition(CodeBlockNameTruncI32F32U, list(F32), list(I32))
	addCodeBlockDefinition(CodeBlockNameTruncI32F64U, list(F64), list(I32))
	addCodeBlockDefinition(CodeBlockNameTruncI64F32U, list(F32), list(I64))
	addCodeBlockDefinition(CodeBlockNameTruncI64F64U, list(F64), list(I64))
	addCodeBlockDefinition(CodeBlockNameDemoteF64, list(F64), list(F32))
	addCodeBlockDefinition(CodeBlockNamePromoteF32, list(F32), list(F64))
	addCodeBlockDefinition(CodeBlockNameConvertF32I32S, list(I32), list(F32))
	addCodeBlockDefinition(CodeBlockNameConvertF32I64S, list(I64), list(F32))
	addCodeBlockDefinition(CodeBlockNameConvertF64I32S, list(I32), list(F64))
	addCodeBlockDefinition(CodeBlockNameConvertF64I64S, list(I64), list(F64))
	addCodeBlockDefinition(CodeBlockNameConvertF32I32U, list(I32), list(F32))
	addCodeBlockDefinition(CodeBlockNameConvertF32I64U, list(I64), list(F32))
	addCodeBlockDefinition(CodeBlockNameConvertF64I32U, list(I32), list(F64))
	addCodeBlockDefinition(CodeBlockNameConvertF64I64U, list(I64), list(F64))
	addCodeBlockDefinition(CodeBlockNameReinterpretF32, list(F32), list(I32))
	addCodeBlockDefinition(CodeBlockNameReinterpretF64, list(F64), list(I64))
	addCodeBlockDefinition(CodeBlockNameReinterpretI32, list(I32), list(F32))
	addCodeBlockDefinition(CodeBlockNameReinterpretI64, list(I64), list(F64))
	addCodeBlockDefinition(CodeBlockNameExtendI32I8S, list(I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameExtendI32I16S, list(I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameExtendI64I8S, list(I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameExtendI64I16S, list(I64), list(I64))
	addCodeBlockDefinition(CodeBlockNameExtendI64I32S, list(I64), list(I64))

	// Load And Store Instructions
	addCodeBlockDefinition(CodeBlockNameI32Load, list(Any), list(I32))
	addCodeBlockDefinition(CodeBlockNameI64Load, list(Any), list(I64))
	addCodeBlockDefinition(CodeBlockNameF32Load, list(Any), list(F32))
	addCodeBlockDefinition(CodeBlockNameF64Load, list(Any), list(F64))
	addCodeBlockDefinition(CodeBlockNameLoadI32I8S, list(Any), list(I32))
	addCodeBlockDefinition(CodeBlockNameLoadI32I16S, list(Any), list(I32))
	addCodeBlockDefinition(CodeBlockNameLoadI64I8S, list(Any), list(I64))
	addCodeBlockDefinition(CodeBlockNameLoadI64I16S, list(Any), list(I64))
	addCodeBlockDefinition(CodeBlockNameLoadI64I32S, list(Any), list(I64))
	addCodeBlockDefinition(CodeBlockNameLoadI32I8U, list(Any), list(I32))
	addCodeBlockDefinition(CodeBlockNameLoadI32I16U, list(Any), list(I32))
	addCodeBlockDefinition(CodeBlockNameLoadI64I8U, list(Any), list(I64))
	addCodeBlockDefinition(CodeBlockNameLoadI64I16U, list(Any), list(I64))
	addCodeBlockDefinition(CodeBlockNameLoadI64I32U, list(Any), list(I64))
	addCodeBlockDefinition(CodeBlockNameStoreI32I8, list(Any, I32), list())
	addCodeBlockDefinition(CodeBlockNameStoreI32I16, list(Any, I32), list())
	addCodeBlockDefinition(CodeBlockNameStoreI64I8, list(Any, I64), list())
	addCodeBlockDefinition(CodeBlockNameStoreI64I16, list(Any, I64), list())
	addCodeBlockDefinition(CodeBlockNameStoreI64I32, list(Any, I64), list())

	// Additional Memory-Related Instructions
	addCodeBlockDefinition(CodeBlockNameCurrentMemory, list(I32), list(I32))
	addCodeBlockDefinition(CodeBlockNameGrowMemory, list(), list(I32))
}

// addCodeBlockDefinition adds a code definitions to the package level map.
func addCodeBlockDefinition(name string, args, returns []CodeBlockType) {
	codeBlockDefinition[name] = CodeBlockDefinition{
		Name:     name,
		Args:     args,
		NArgs:    len(args),
		Returns:  returns,
		NReturns: len(returns),
	}
}

// list returns a list of types equivalent to the provided one.
func list(t ...CodeBlockType) []CodeBlockType {
	return t
}
