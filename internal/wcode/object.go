package wcode

import "strings"

// FuncData contains the func pointcut data.
type FuncData struct {
	Index       string
	Order       int
	Name        string
	Params      []string
	ParamTypes  []string
	TotalParams int
	Locals      []string
	LocalTypes  []string
	TotalLocals int
	ResultType  string
	Code        string
	IsImported  bool
	IsExported  bool
	IsStart     bool
}

// newFuncData is the constructor for FuncData.
func newFuncData(ctx *ModuleContext, def *FunctionDefinition) *FuncData {
	return &FuncData{
		Index:       def.Name,
		Order:       def.Index(ctx),
		Name:        funcName(def),
		Params:      funcParams(def),
		ParamTypes:  funcParamTypes(def),
		TotalParams: len(def.Params),
		Locals:      funcLocals(def),
		LocalTypes:  funcLocalTypes(def),
		TotalLocals: len(def.Locals),
		ResultType:  def.Result,
		Code:        def.Code(),
		IsImported:  def.Imported != nil,
		IsExported:  def.Exported != nil,
		IsStart:     def.IsStart,
	}
}

// funcName returns the function name for some function definition
func funcName(def *FunctionDefinition) string {
	var name string
	switch {
	case def.Exported != nil:
		name = def.Exported.ExportName
	case def.Imported != nil:
		name = strings.Join([]string{def.Imported.ModuleName, def.Imported.ExportName}, ".")
	default:
		name = strings.Trim(def.Name, "$")
	}
	return name
}

// funcLocals returns the parameters for some function definition
func funcParams(def *FunctionDefinition) []string {
	var params []string
	for _, param := range def.Parameters() {
		params = append(params, param.Name)
	}
	return params
}

// funcParamTypes returns the parameter types for some function definition
func funcParamTypes(def *FunctionDefinition) []string {
	var paramTypes []string
	for _, param := range def.Parameters() {
		paramTypes = append(paramTypes, param.Type)
	}
	return paramTypes
}

// funcLocals returns the locals for some function definition
func funcLocals(def *FunctionDefinition) []string {
	var locals []string
	for _, local := range def.Locals {
		locals = append(locals, local.Name)
	}
	return locals
}

// funcLocalTypes returns the local types for some function definition
func funcLocalTypes(def *FunctionDefinition) []string {
	var localTypes []string
	for _, local := range def.Locals {
		localTypes = append(localTypes, local.Type)
	}
	return localTypes
}

// CallData contains the call pointcut data.
type CallData struct {
	Callee    *FuncData
	Caller    *FuncData
	Args      []*ArgData
	TotalArgs int
}

// newCallData is the constructor for CallData.
func newCallData(ctx *ModuleContext, def *CallDefinition) *CallData {
	// Fill arguments list.
	args := getArgsFromCallInstr(def)

	// Find function instruction.
	funcInstr := def.Instr
	for funcInstr.name != instructionFunction {
		if funcInstr.parent == nil {
			funcInstr = nil
			break
		}
		funcInstrAux, ok := funcInstr.parent.(*Instruction)
		if !ok {
			funcInstr = nil
			break
		}
		funcInstr = funcInstrAux
	}

	// Find caller definition.
	var caller *FuncData
	if funcInstr != nil && len(funcInstr.values) > 1 {
		if fnDef, ok := ctx.functions[funcInstr.values[0].String()]; ok {
			caller = newFuncData(ctx, fnDef)
		}
	}

	// Find callee definition.
	var callee *FuncData
	if fnDef, ok := ctx.functions[def.Instr.values[0].String()]; ok {
		callee = newFuncData(ctx, fnDef)
	}
	return &CallData{
		Caller:    caller,
		Callee:    callee,
		Args:      args,
		TotalArgs: len(args),
	}
}

// ArgData contains the argument data.
type ArgData struct {
	Type  string
	Order int
	Instr string
}

// ArgData contains the args pointcut data.
type ArgsData struct {
	Callee    *FuncData
	Caller    *FuncData
	Args      []*ArgData
	TotalArgs int
}

// newArgsData is the constructor for ArgsData.
func newArgsData(ctx *ModuleContext, def *CallDefinition) *ArgsData {
	// Fill arguments list.
	args := getArgsFromCallInstr(def)

	// Find function instruction.
	funcInstr := def.Instr
	for funcInstr.name != instructionFunction {
		if funcInstr.parent == nil {
			funcInstr = nil
			break
		}
		funcInstrAux, ok := funcInstr.parent.(*Instruction)
		if !ok {
			funcInstr = nil
			break
		}
		funcInstr = funcInstrAux
	}

	// Find caller definition.
	var caller *FuncData
	if funcInstr != nil && len(funcInstr.values) > 1 {
		if fnDef, ok := ctx.functions[funcInstr.values[0].String()]; ok {
			caller = newFuncData(ctx, fnDef)
		}
	}

	// Find callee definition.
	var callee *FuncData
	if fnDef, ok := ctx.functions[def.Instr.values[0].String()]; ok {
		callee = newFuncData(ctx, fnDef)
	}

	return &ArgsData{
		Callee:    callee,
		Caller:    caller,
		Args:      args,
		TotalArgs: len(args),
	}
}

// getArgsFromCallInstr returns the arguments data for some call instruction.
func getArgsFromCallInstr(def *CallDefinition) []*ArgData {
	var args []*ArgData
	fnParams := def.FunctionCalled.Parameters()
	fnParamsCount := len(fnParams)

	for i := 1; i < len(def.Instr.values); i++ {
		paramI := i - 1
		if paramI >= fnParamsCount {
			break
		}
		args = append(args, &ArgData{
			Type:  fnParams[paramI].Type,
			Order: paramI,
			Instr: def.Instr.values[i].String(),
		})
	}
	return args
}

// ReturnsData contains the returns pointcut data.
type ReturnsData struct {
	Func  *FuncData
	Instr string
	Type  string
}

// newReturns is the constructor for Returns.
func newReturnsData(ctx *ModuleContext, def *FunctionDefinition, instr Block) *ReturnsData {
	return &ReturnsData{
		Func:  newFuncData(ctx, def),
		Instr: instr.String(),
		Type:  def.Result,
	}
}
