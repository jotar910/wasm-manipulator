package wcode

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// VisitedTree is implemented by those who can be transversed.
// each node transversed may accept a visit from some code block visitor.
type VisitedTree interface {
	Traverse(Visitor)
	TraverseConditional(Visitor)
}

// Visited is implemented by those who accept a visit from some code block visitor.
type Visited interface {
	Accept(Visitor) bool
}

// Visitor is implemented by a visitor object that may visit a code block.
type Visitor interface {
	VisitElement(*element) bool
	VisitInstruction(*Instruction) bool
	VisitKeyword(*keyword) bool
	VisitQuoted(*quoted) bool
	VisitText(*text) bool
	VisitEvaluation(*evaluation) bool
	VisitEvaluationRef(*evaluationRef) bool
}

// visitorAdapter implements the Visitor interface with all actions as empty-
type visitorAdapter struct {
}

// VisitElement visits an element block.
func (*visitorAdapter) VisitElement(*element) bool {
	// Empty by design.
	return false
}

// VisitInstruction visits an instruction block.
func (*visitorAdapter) VisitInstruction(*Instruction) bool {
	// Empty by design.
	return false
}

// VisitKeyword visits a keyword block.
func (*visitorAdapter) VisitKeyword(*keyword) bool {
	// Empty by design.
	return false
}

// VisitQuoted visits a quoted block.
func (*visitorAdapter) VisitQuoted(*quoted) bool {
	// Empty by design.
	return false
}

// VisitText visits a text block.
func (*visitorAdapter) VisitText(*text) bool {
	// Empty by design.
	return false
}

// VisitEvaluation visits an evaluation block.
func (*visitorAdapter) VisitEvaluation(*evaluation) bool {
	// Empty by design.
	return false
}

// VisitEvaluation visits an evaluation reference block.
func (*visitorAdapter) VisitEvaluationRef(*evaluationRef) bool {
	// Empty by design.
	return false
}

// CallDefinition contains the definitions data for some call instruction.
type CallDefinition struct {
	Instr          *Instruction
	FunctionCalled *FunctionDefinition
}

// newCallDefinition is the constructor for CallDefinition.
func newCallDefinition(instr *Instruction, fn *FunctionDefinition) *CallDefinition {
	return &CallDefinition{Instr: instr, FunctionCalled: fn}
}

// ImportedDefinition contains the definitions data for some imported instruction.
type ImportedDefinition struct {
	ModuleName string
	ExportName string
}

// newImportedDefinition is the constructor for ImportedDefinition.
func newImportedDefinition(moduleName, exportName string) *ImportedDefinition {
	return &ImportedDefinition{moduleName, exportName}
}

// ExportedDefinition contains the definitions data for some exported instruction.
type ExportedDefinition struct {
	ExportName string
}

// newExportedDefinition is the constructor for ExportedDefinition.
func newExportedDefinition(exportName string) *ExportedDefinition {
	return &ExportedDefinition{exportName}
}

// FunctionDefinition contains the definitions data for some function instruction.
type FunctionDefinition struct {
	Name     string
	TypeName string
	Params   map[string]*FunctionParamDefinition
	Result   string
	Locals   map[string]*FunctionLocalDefinition
	Imported *ImportedDefinition
	Exported *ExportedDefinition
	IsStart  bool
	Alias    map[string]string
	order    int
	instr    *Instruction
}

// newFunctionDefinition is the constructor for FunctionDefinition.
func newFunctionDefinition(instr *Instruction) *FunctionDefinition {
	return &FunctionDefinition{
		Params: make(map[string]*FunctionParamDefinition),
		Locals: make(map[string]*FunctionLocalDefinition),
		Alias:  make(map[string]string),
		instr:  instr,
	}
}

// Index returns the function index value on the module context.
func (fn *FunctionDefinition) Index(ctx *ModuleContext) int {
	if fn.Imported != nil {
		return fn.order
	}
	var importsCount int
	for _, v := range ctx.importFunctions {
		importsCount += len(v)
	}
	return fn.order + importsCount
}

// Parameters returns all the parameters sorted.
func (fn *FunctionDefinition) Parameters() []*FunctionParamDefinition {
	var res []*FunctionParamDefinition
	for _, v := range fn.Params {
		res = append(res, v)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].order < res[j].order
	})
	return res
}

// LocalsArr returns all the locals sorted.
func (fn *FunctionDefinition) LocalsArr() []*FunctionLocalDefinition {
	var res []*FunctionLocalDefinition
	for _, v := range fn.Locals {
		res = append(res, v)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].order < res[j].order
	})
	return res
}

// AliasValue returns the alias value for some key.
func (fn *FunctionDefinition) AliasValue(key string) (string, bool) {
	res, ok := fn.Alias[key]
	return res, ok
}

// AliasKey returns the key alias for some value.
func (fn *FunctionDefinition) AliasKey(value string) (string, bool) {
	for k, v := range fn.Alias {
		if v == value {
			return k, true
		}
	}
	return "", false
}

// Code returns the function code,
// ignoring the non-expression instructions.
func (fn *FunctionDefinition) Code() string {
	return FuncInstrsString(fn.instr)
}

// AddCode adds new code to the end of the function.
func (fn *FunctionDefinition) AddCode(code string) {
	codeEl := NewCodeParser(code).parse()
	fn.addInstrsAtEnd(codeEl.blocks)
}

// addLocal adds some local to the function.
func (fn *FunctionDefinition) addLocal(local *FunctionLocalDefinition, value Block) {
	instrs := []Block{local.instr}
	if value != nil {
		instrs = append(instrs, value)
	}

	fn.Locals[local.Name] = local
	_, startIndex := fn.findFirstInstruction()
	if startIndex == -1 {
		fn.instr.values = append(fn.instr.values, instrs...)
	} else {
		var newValues []Block
		newValues = append(newValues, fn.instr.values[:startIndex]...)
		newValues = append(newValues, instrs...)
		newValues = append(newValues, fn.instr.values[startIndex:]...)
		fn.instr.values = newValues
	}
}

// addInstrsAtStart adds instructions at the beginning of the function.
func (fn *FunctionDefinition) addInstrsAtStart(instrs []Block) {
	_, startIndex := fn.findFirstInstruction()

	// Append values to the function.
	if startIndex == -1 {
		fn.instr.values = append(fn.instr.values, instrs...)
		return
	}

	// Check if there is static local.set
	for startIndex < len(fn.instr.values) {
		block := fn.instr.values[startIndex]
		blockInstr, ok := block.(*Instruction)
		if !ok || blockInstr.name != instructionCodeSetLocal || len(blockInstr.values) != 2 {
			break
		}
		blockArg := blockInstr.values[1]
		blockArgInstr, ok := blockArg.(*Instruction)
		if !ok || len(blockArgInstr.values) != 1 {
			break
		}
		if !strings.HasSuffix(blockArgInstr.name, "const") {
			break
		}
		startIndex++
	}

	// Append values to the function.
	if startIndex == len(fn.instr.values) {
		fn.instr.values = append(fn.instr.values, instrs...)
		return
	}
	// Add values in the middle of the function.
	var newValues []Block
	newValues = append(newValues, fn.instr.values[:startIndex]...)
	newValues = append(newValues, instrs...)
	newValues = append(newValues, fn.instr.values[startIndex:]...)
	fn.instr.values = newValues
}

// addInstrsAtEnd adds instructions at the end of the function.
func (fn *FunctionDefinition) addInstrsAtEnd(instrs []Block) {
	if len(fn.instr.values) == 0 {
		fn.instr.values = append(fn.instr.values, instrs...)
		return
	}

	// Find returns and add the values above it.
	returnVisitor := newReturnFuncVisitor()
	fn.instr.Traverse(returnVisitor)
	returnInstrs := returnVisitor.instrs

	// Add values above the return instruction.
	for i, returnInstr := range returnInstrs {
		parent := returnInstr.getParent()
		if parent == nil {
			continue
		}
		parentInstr, ok := parent.(*Instruction)
		if !ok {
			continue
		}
		err := parentInstr.addChildren(returnInstr, instrs...)
		if err != nil {
			logrus.Errorf("adding instruction at the end of the function: return index %d: %v", i, err)
		}
	}

	// Append values to the function if not ends on return.
	if len(returnInstrs) == 0 || fn.instr.values[len(fn.instr.values)-1] != returnInstrs[len(returnInstrs)-1] {
		fn.instr.values = append(fn.instr.values, instrs...)
	}
}

// findFirstInstruction returns the first code instruction on the function.
func (fn *FunctionDefinition) findFirstInstruction() (Block, int) {
	return findFirstInstruction(fn.instr.values)
}

// instrToImported returns the imported definition for the function.
func (fn *FunctionDefinition) instrToImported(instr *Instruction) *ImportedDefinition {
	if instr.parent == nil {
		return nil
	}
	parent, ok := instr.parent.(*Instruction)
	if !ok || parent.name != instructionImport {
		return nil
	}
	if err := validateImportFunction(parent); err != nil {
		logrus.Fatalf("checking if function is imported: invalid import instruction: %v", err)
	}
	return newImportedDefinition(strings.Trim(parent.values[0].String(), "\""),
		strings.Trim(parent.values[1].String(), "\""))
}

// instrToExported returns the exported definition for the function.
func (fn *FunctionDefinition) instrToExported(instr *Instruction) *ExportedDefinition {
	if instr.parent == nil {
		return nil
	}
	parent, ok := instr.parent.(*Instruction)
	if !ok || parent.name != instructionExport {
		return nil
	}
	if err := validateExportFunction(parent); err != nil {
		logrus.Fatalf("checking if function is imported: invalid import instruction: %v", err)
	}
	return newExportedDefinition(strings.Trim(parent.values[0].String(), "\""))
}

// FunctionParamDefinition contains the definitions data for some parameter instruction.
type FunctionParamDefinition struct {
	Name  string
	Type  string
	order int
}

// newFunctionParamDefinition is the constructor for FunctionParamDefinition.
func newFunctionParamDefinition(name, typ string, order int) *FunctionParamDefinition {
	return &FunctionParamDefinition{
		Name:  name,
		Type:  typ,
		order: order,
	}
}

// TypeCode returns the type code.
func (fp *FunctionParamDefinition) TypeCode() int {
	return varTypeCode(fp.Type)
}

// TypeCodeOnFn returns the type code accordingly to the function scope.
func (fp *FunctionParamDefinition) TypeCodeOnFn(isExported bool) int {
	if !isExported && fp.Type == string(varTypeString) {
		return varTypeCode(string(varTypeIdentifier))
	}
	return fp.TypeCode()
}

// IsPrimitive returns if the type is primitive.
func (fp *FunctionParamDefinition) IsPrimitive() bool {
	return IsVarTypeStrPrimitive(fp.Type)
}

// Index returns the parameter index.
func (fp *FunctionParamDefinition) Index() int {
	return fp.order
}

// FunctionLocalDefinition contains the definitions data for some local instruction.
type FunctionLocalDefinition struct {
	Name         string
	Type         string
	initialValue string
	instr        *Instruction
	order        int
}

// newFunctionLocalDefinition is the constructor for FunctionLocalDefinition.
func newFunctionLocalDefinition(instr *Instruction, name, typ, value string, order int) *FunctionLocalDefinition {
	return &FunctionLocalDefinition{
		Name:         name,
		Type:         typ,
		initialValue: value,
		instr:        instr,
		order:        order,
	}
}

// TypeDefinition contains the definitions data for some type instruction.
type TypeDefinition struct {
	Name   string
	Params []string
	Result string
}

// newTypeDefinition is a constructor for TypeDefinition.
func newTypeDefinition() *TypeDefinition {
	return &TypeDefinition{}
}

// signature returns the type signature.
func (typ *TypeDefinition) signature() string {
	return typeSignature(typ.Result, typ.Params)
}

// GlobalDefinition contains the definitions data for some global instruction.
type GlobalDefinition struct {
	Name         string
	Type         string
	Mutable      bool
	Imported     *ImportedDefinition
	Exported     *ExportedDefinition
	initialValue string
}

// newGlobalDefinition is the constructor for GlobalDefinition.
func newGlobalDefinition() *GlobalDefinition {
	return &GlobalDefinition{}
}

// instrToImported returns the imported definition for the global.
func (global *GlobalDefinition) instrToImported(instr *Instruction) *ImportedDefinition {
	if instr.parent == nil {
		return nil
	}
	parent, ok := instr.parent.(*Instruction)
	if !ok || parent.name != instructionImport {
		return nil
	}
	if err := validateImportGlobal(parent); err != nil {
		logrus.Fatalf("checking if global is imported: invalid import instruction: %v", err)
	}
	return newImportedDefinition(strings.Trim(parent.values[0].String(), "\""),
		strings.Trim(parent.values[1].String(), "\""))
}

// startFunctionContextVisitor is responsible to fill the start function data for the module context.
type startFunctionContextVisitor struct {
	visitorAdapter
	ctx *ModuleContext
}

// newExportedContextVisitor is the constructor for startFunctionContextVisitor.
func newStartFunctionContextVisitor(ctx *ModuleContext) *startFunctionContextVisitor {
	return &startFunctionContextVisitor{ctx: ctx}
}

// VisitInstruction visits an instruction block.
func (sc *startFunctionContextVisitor) VisitInstruction(instr *Instruction) bool {
	if !isStartFunction(instr) {
		return false
	}
	if len(instr.values) != 1 {
		logrus.Fatalf("invalid start function instruction")
	}
	fnName := instr.values[0].String()
	fn, ok := sc.ctx.functions[fnName]
	if !ok {
		logrus.Fatalf("invalid start function: function with name %s not found in module", fnName)
	}
	fn.IsStart = true
	sc.ctx.startFunction = fn
	return true
}

// exportedContextVisitor is responsible to fill the exported data for the module context.
type exportedContextVisitor struct {
	visitorAdapter
	ctx *ModuleContext
}

// newExportedContextVisitor is the constructor for exportedContextVisitor.
func newExportedContextVisitor(ctx *ModuleContext) *exportedContextVisitor {
	return &exportedContextVisitor{ctx: ctx}
}

// VisitInstruction visits an instruction block.
func (ec *exportedContextVisitor) VisitInstruction(instr *Instruction) bool {
	if !isExportInstruction(instr) {
		return false
	}
	switch instr.name {
	case instructionFunction:
		ec.inspectFuncExportInstruction(instr)
	case instructionGlobal:
		ec.inspectGlobalExportInstruction(instr)
	default:
		return false
	}
	return true
}

// inspectFuncExportInstruction inspects export function instruction.
func (ec *exportedContextVisitor) inspectFuncExportInstruction(instr *Instruction) {
	if len(instr.values) != 1 {
		logrus.Fatalf("invalid export function instruction")
	}
	fnName := instr.values[0].String()
	fn, ok := ec.ctx.functions[fnName]
	if !ok {
		logrus.Fatalf("invalid exported function: function with name %s not found in module", fnName)
	}
	fn.Exported = fn.instrToExported(instr)
	ec.ctx.setExportFunction(fn.Exported.ExportName, fn)
}

// inspectGlobalExportInstruction inspects export global instruction.
func (ec *exportedContextVisitor) inspectGlobalExportInstruction(instr *Instruction) {
	if len(instr.values) != 1 {
		logrus.Fatalf("invalid export global instruction")
	}
	parentInstr, ok := instr.parent.(*Instruction)
	if !ok {
		panic("exported global parent must always be of type instruction")
	}
	exportedName := strings.Trim(parentInstr.values[0].String(), "\"")
	globalName := "$" + exportedName
	global, ok := ec.ctx.globals[globalName]
	if !ok {
		logrus.Fatalf("invalid exported global: global with name %s not found in module", globalName)
	}
	global.Exported = newExportedDefinition(exportedName)
	ec.ctx.setExportGlobal(exportedName, global)
}

// typeContextVisitor is responsible to fill the types data for the module context.
type typeContextVisitor struct {
	visitorAdapter
	ctx *ModuleContext
}

// newTypeContextVisitor is the constructor for typeContextVisitor.
func newTypeContextVisitor(ctx *ModuleContext) *typeContextVisitor {
	return &typeContextVisitor{ctx: ctx}
}

// VisitInstruction visits an instruction block.
func (tc *typeContextVisitor) VisitInstruction(instr *Instruction) bool {
	if !isFunctionTypeInstruction(instr) {
		return false
	}
	tc.inspectTypeInstruction(instr)
	return true
}

// inspectTypeInstruction inspects type instruction.
func (tc *typeContextVisitor) inspectTypeInstruction(instr *Instruction) {
	typ := newTypeDefinition()
	typeInstr, ok := instr.parent.(*Instruction)
	if !ok || len(typeInstr.values) != 2 {
		return
	}
	typ.Name = typeInstr.values[0].String()

	canRun := true
	for i := 0; i < len(instr.values) && canRun; i++ {
		arg, ok := instr.values[i].(*Instruction)
		if !ok {
			break
		}
		switch arg.name {
		case "param":
			for _, paramType := range arg.values {
				typ.Params = append(typ.Params, paramType.String())
			}
		case "result":
			if len(arg.values) != 1 {
				break
			}
			typ.Result = arg.values[0].String()
		default:
			canRun = false
		}
	}

	tc.ctx.setType(typ.Name, typ)
}

// functionContextVisitor is responsible to fill the functions data for the module context.
type functionContextVisitor struct {
	visitorAdapter
	ctx *ModuleContext
}

// newFunctionContextVisitor is the constructor for functionContextVisitor.
func newFunctionContextVisitor(ctx *ModuleContext) *functionContextVisitor {
	return &functionContextVisitor{ctx: ctx}
}

// VisitInstruction visits an instruction block.
func (fc *functionContextVisitor) VisitInstruction(instr *Instruction) bool {
	if instr.name != instructionFunction {
		return false
	}
	switch {
	case isImportInstruction(instr):
		fc.inspectFuncImportedInstruction(instr)
	case isInternalInstruction(instr):
		fc.inspectFuncInstruction(instr)
	default:
		return false
	}
	return true
}

// inspectFuncImportedInstruction inspects import function instruction.
func (fc *functionContextVisitor) inspectFuncImportedInstruction(instr *Instruction) {
	if len(instr.values) != 2 {
		logrus.Fatalf("invalid import function instruction")
	}
	fn := newFunctionDefinition(instr)
	fn.order = len(fc.ctx.importFunctions)
	fn.Name = instr.values[0].String()
	fn.Imported = fn.instrToImported(instr)
	typeInstr, ok := instr.values[1].(*Instruction)
	if !ok || len(typeInstr.values) != 1 {
		return
	}
	fn.TypeName = typeInstr.values[0].String()
	typeDef, ok := fc.ctx.types[fn.TypeName]
	if !ok {
		logrus.Fatalf("filling function import definition: type name %s not found in module", fn.TypeName)
	}
	for i, param := range typeDef.Params {
		name := strconv.Itoa(i)
		fn.Params[name] = newFunctionParamDefinition(name, param, i)
	}
	fn.Result = typeDef.Result
	fc.ctx.setFunction(fn.Name, fn)
	fc.ctx.setImportFunction(fn.Imported.ModuleName, fn.Imported.ExportName, fn)
}

// inspectFuncInstruction inspects function instruction.
func (fc *functionContextVisitor) inspectFuncInstruction(instr *Instruction) {
	if len(instr.values) == 0 {
		logrus.Fatalf("inspecting function instruction: illegal empty function")
	}

	fn := newFunctionDefinition(instr)
	fn.order = len(fc.ctx.functions) - len(fc.ctx.importFunctions)
	fn.Name = instr.values[0].String()
	fn.Imported = fn.instrToImported(instr)

	canRun := true
	for i := 1; i < len(instr.values) && canRun; i++ {
		arg, ok := instr.values[i].(*Instruction)
		if !ok {
			break
		}
		switch arg.name {
		case instructionType:
			if len(arg.values) != 1 {
				break
			}
			fn.TypeName = arg.values[0].String()
		case instructionParam:
			if len(arg.values) != 2 {
				break
			}
			name := arg.values[0].String()
			fn.Params[name] = newFunctionParamDefinition(name, arg.values[1].String(), len(fn.Params))
		case instructionResult:
			if len(arg.values) != 1 {
				break
			}
			fn.Result = arg.values[0].String()
		case instructionLocal:
			if len(arg.values) != 2 {
				break
			}
			name := arg.values[0].String()
			fn.Locals[name] = newFunctionLocalDefinition(arg, name, arg.values[1].String(), "", len(fn.Locals))
		default:
			canRun = false
		}
	}
	fc.ctx.setFunction(fn.Name, fn)
}

// globalContextVisitor is responsible to fill the globals data for the module context.
type globalContextVisitor struct {
	visitorAdapter
	ctx *ModuleContext
}

// newGlobalContextVisitor is the constructor for globalContextVisitor.
func newGlobalContextVisitor(ctx *ModuleContext) *globalContextVisitor {
	return &globalContextVisitor{ctx: ctx}
}

// VisitInstruction visits an instruction block.
func (gc *globalContextVisitor) VisitInstruction(instr *Instruction) bool {
	if instr.name != instructionGlobal {
		return false
	}
	switch {
	case isImportInstruction(instr):
		gc.inspectGlobalImportedInstruction(instr)
	case isInternalInstruction(instr):
		gc.inspectGlobalInstruction(instr)
	default:
		return false
	}
	return true
}

// inspectGlobalImportedInstruction inspects import global instruction.
func (gc *globalContextVisitor) inspectGlobalImportedInstruction(instr *Instruction) {
	if len(instr.values) != 2 {
		logrus.Fatalf("invalid import global instruction")
	}
	global := newGlobalDefinition()
	global.Name = instr.values[0].String()
	global.Imported = global.instrToImported(instr)
	gc.fillGlobalType(instr, global)
	gc.fillGlobalInitialValue(instr, global)
	gc.ctx.setGlobal(global.Name, global)
	gc.ctx.setImportGlobal(global.Imported.ModuleName, global.Imported.ExportName, global)
}

// inspectGlobalInstruction inspects global instruction.
func (gc *globalContextVisitor) inspectGlobalInstruction(instr *Instruction) {
	if len(instr.values) < 2 {
		logrus.Fatalf("invalid global instruction")
	}
	global := newGlobalDefinition()
	global.Name = instr.values[0].String()
	gc.fillGlobalType(instr, global)
	gc.fillGlobalInitialValue(instr, global)
	gc.ctx.setGlobal(global.Name, global)

}

func (gc *globalContextVisitor) fillGlobalType(instr *Instruction, global *GlobalDefinition) {
	mutInstr, ok := instr.values[1].(*Instruction)
	if !ok || mutInstr.name != instructionMutable {
		global.Type = instr.values[1].String()
		return
	}
	global.Mutable = true
	global.Type = mutInstr.values[0].String()
}

// fillGlobalInitialValue fills the global initial value.
func (gc *globalContextVisitor) fillGlobalInitialValue(instr *Instruction, global *GlobalDefinition) {
	if len(instr.values) < 3 {
		return
	}
	valueInstr, ok := instr.values[2].(*Instruction)
	if !ok || !strings.HasSuffix(valueInstr.name, "."+instructionCodeConst) || len(valueInstr.values) < 1 {
		return
	}
	var valueArr []string
	for _, v := range valueInstr.values {
		valueArr = append(valueArr, v.String())
	}
	global.initialValue = strings.Join(valueArr, "")
}

// moduleInstrsVisitor is responsible to group all the module instructions.
type moduleInstrsVisitor struct {
	visitorAdapter
	modules []*Instruction
}

// newModuleInstrsVisitor is the constructor for moduleInstrsVisitor.
func newModuleInstrsVisitor() *moduleInstrsVisitor {
	return &moduleInstrsVisitor{}
}

// VisitInstruction visits an instruction block.
func (mc *moduleInstrsVisitor) VisitInstruction(instr *Instruction) bool {
	if instr.name != instructionModule {
		return false
	}
	mc.modules = append(mc.modules, instr)
	return true
}

func (mc *moduleInstrsVisitor) Module() (*Instruction, bool) {
	if len(mc.modules) > 0 {
		return mc.modules[0], true
	}
	return nil, false
}

// targetInstrsVisitor is responsible to find all the target instructions.
type targetInstrsVisitor struct {
	visitorAdapter
	targets []*Instruction
}

// newTargetInstrsVisitor is the constructor for targetInstrsVisitor.
func newTargetInstrsVisitor() *targetInstrsVisitor {
	return &targetInstrsVisitor{}
}

// VisitInstruction visits an instruction block.
func (tv *targetInstrsVisitor) VisitInstruction(instr *Instruction) bool {
	if instr.name != targetInstructionKeyword || len(instr.values) == 0 {
		return false
	}
	tv.targets = append(tv.targets, instr)
	return true
}

// validateImportFunction returns if some instruction is a valid import function.
func validateImportFunction(instr *Instruction) error {
	if nValues := len(instr.values); nValues != 3 {
		return fmt.Errorf("import function instructions expects 3 arguments but got %d", nValues)
	}
	return nil
}

// validateImportGlobal returns if some instruction is a valid import global.
func validateImportGlobal(instr *Instruction) error {
	if nValues := len(instr.values); nValues != 3 {
		return fmt.Errorf("import global instructions expects 3 arguments but got %d", nValues)
	}
	return nil
}

// validateExportFunction returns if some instruction is a valid export function.
func validateExportFunction(instr *Instruction) error {
	if nValues := len(instr.values); nValues != 2 {
		return fmt.Errorf("export function instructions expects 2 arguments but got %d", nValues)
	}
	return nil
}

// isImportInstruction returns if instruction is an import instruction.
func isImportInstruction(instr *Instruction) bool {
	if instr.parent == nil {
		return false
	}
	parent, ok := instr.parent.(*Instruction)
	return ok && parent.name == instructionImport
}

// isStartFunction returns if instruction is a start function instruction.
func isStartFunction(instr *Instruction) bool {
	if instr.name != instructionFunction || instr.parent == nil {
		return false
	}
	parent, ok := instr.parent.(*Instruction)
	return ok && parent.name == instructionStart
}

// isExportInstruction returns if instruction is an export instruction.
func isExportInstruction(instr *Instruction) bool {
	if instr.parent == nil {
		return false
	}
	parent, ok := instr.parent.(*Instruction)
	return ok && parent.name == instructionExport
}

// isInternalInstruction returns if instruction is an private instruction.
func isInternalInstruction(instr *Instruction) bool {
	if instr.parent == nil {
		return true
	}
	parent, ok := instr.parent.(*Instruction)
	if !ok {
		return false
	}
	return parent.name == instructionModule
}

// isFunctionTypeInstruction returns if instruction is a function type instruction.
func isFunctionTypeInstruction(instr *Instruction) bool {
	if instr.name != instructionFunction || instr.parent == nil {
		return false
	}
	parent, ok := instr.parent.(*Instruction)
	return ok && parent.name == instructionType
}

// typeSignature returns the type signature.
func typeSignature(result string, params []string) string {
	return fmt.Sprintf("%s_%s", result, strings.Join(params, "_"))
}
