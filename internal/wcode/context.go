package wcode

import (
	"errors"
	"fmt"
	"regexp"

	"joao/wasm-manipulator/internal/wgenerator"
	"joao/wasm-manipulator/internal/wparser/variable"
	"joao/wasm-manipulator/internal/wyaml"

	"github.com/sirupsen/logrus"
)

// ModuleContext contains the context data for some web assembly code module.
type ModuleContext struct {
	NeedJS          bool
	FunctionAlias   map[string]string
	GlobalAlias     map[string]string
	entryBlock      Block
	functions       map[string]*FunctionDefinition
	exportFunctions map[string]*FunctionDefinition
	importFunctions map[string]map[string]*FunctionDefinition
	startFunction   *FunctionDefinition
	types           map[string]*TypeDefinition
	globals         map[string]*GlobalDefinition
	exportGlobals   map[string]*GlobalDefinition
	importGlobals   map[string]map[string]*GlobalDefinition
	runtimeChanges  map[string]*runtimeChanges
	glueFunctions   *glueFunctionsState
	evalsToRemove   []Block
}

// NewModuleContext is the constructor for ModuleContext.
func NewModuleContext(entryBlock Block) *ModuleContext {
	moduleCtx := newModuleContext(entryBlock)

	// Fill module context for the entry block.
	moduleCtx.FillContext(entryBlock)

	return moduleCtx
}

// newModuleContext is the instance constructor for ModuleContext.
func newModuleContext(entryBlock Block) *ModuleContext {
	return &ModuleContext{
		FunctionAlias:   make(map[string]string),
		GlobalAlias:     make(map[string]string),
		entryBlock:      entryBlock,
		functions:       make(map[string]*FunctionDefinition),
		exportFunctions: make(map[string]*FunctionDefinition),
		importFunctions: make(map[string]map[string]*FunctionDefinition),
		types:           make(map[string]*TypeDefinition),
		globals:         make(map[string]*GlobalDefinition),
		exportGlobals:   make(map[string]*GlobalDefinition),
		importGlobals:   make(map[string]map[string]*GlobalDefinition),
		runtimeChanges:  make(map[string]*runtimeChanges),
		glueFunctions:   newGlueFunctionsState(),
	}
}

// String return the code of the module as a clean string.
func (ctx *ModuleContext) String() string {
	return ctx.entryBlock.String()
}

// StringIndent return the code of the module indented.
func (ctx *ModuleContext) StringIndent() string {
	return ctx.entryBlock.StringIndent("")
}

// OrderMap returns the element orders map.
func (ctx *ModuleContext) OrderMap() map[string]int {
	res := make(map[string]int)
	for _, f := range ctx.functions {
		res[f.Name] = f.order
	}
	for name, alias := range ctx.FunctionAlias {
		res[alias] = ctx.functions[name].order
	}
	return res
}

// FillContext fills the module context.
// the data filled will be obtained from the provided block.
func (ctx *ModuleContext) FillContext(block Block) {
	// Fill types.
	typeVisitor := newTypeContextVisitor(ctx)
	block.Traverse(typeVisitor)

	// Fill functions.
	fnVisitor := newFunctionContextVisitor(ctx)
	block.Traverse(fnVisitor)

	// Fill globals.
	globalVisitor := newGlobalContextVisitor(ctx)
	block.Traverse(globalVisitor)

	// Fill exported info.
	exportedVisitor := newExportedContextVisitor(ctx)
	block.Traverse(exportedVisitor)

	// Fill start function.
	startFunctionVisitor := newStartFunctionContextVisitor(ctx)
	block.Traverse(startFunctionVisitor)
}

// Function returns the function definition by its name.
func (ctx *ModuleContext) Function(name string) (*FunctionDefinition, bool) {
	res, ok := ctx.functions[name]
	return res, ok
}

// Functions returns the list of all function definition.
func (ctx *ModuleContext) Functions() []*FunctionDefinition {
	var res []*FunctionDefinition
	for _, fnDef := range ctx.functions {
		res = append(res, fnDef)
	}
	return res
}

// ExportFunction returns the exported function definition by its exported name.
func (ctx *ModuleContext) ExportFunction(exportName string) (*FunctionDefinition, bool) {
	res, ok := ctx.exportFunctions[exportName]
	return res, ok
}

// ExportFunctionByRegex returns the exported function definition by regex.
func (ctx *ModuleContext) ExportFunctionByRegex(regexStr string) (*FunctionDefinition, bool) {
	reg, err := regexp.Compile(regexStr) // Must be checked before
	if err != nil {
		logrus.Fatalf("invalid regex: regex for function name")
		return nil, false
	}
	for k, v := range ctx.exportFunctions {
		if reg.MatchString(k) {
			return v, true
		}
	}
	return nil, false
}

// StartFunction returns the start function for the module on the context.
func (ctx *ModuleContext) StartFunction() (*FunctionDefinition, bool) {
	return ctx.startFunction, ctx.startFunction != nil
}

// AddGlobal adds a new global to the module.
func (ctx *ModuleContext) AddGlobal(value string) (*GlobalDefinition, error) {
	// Parses variable definition.
	expr, err := variable.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("parsing global variable value %q: %v", value, err)
	}

	// Transforms definition into code.
	globalIndex := fmt.Sprintf("$%sg%d", wgenerator.CodeIndexPrefix, len(ctx.globals))
	code := wgenerator.GlobalVariableToCode(expr, globalIndex)

	// Parses variable code.
	codeEl := NewCodeParser(code).parse()
	if len(codeEl.blocks) == 0 {
		return nil, errors.New("error parsing global code to add on module context")
	}

	// Adds variable element to module context.
	ctx.addBlocks(codeEl)
	return ctx.globals[globalIndex], nil
}

// AddType adds a new type to the module.
func (ctx *ModuleContext) AddType(function *wyaml.FunctionYAML) (*TypeDefinition, error) {
	// Parse function data.
	var params []string
	for _, arg := range function.Args {
		params = append(params, arg.Type)
	}
	result := function.Result

	// Add type.
	return ctx.addType(params, result)
}

// AddFunction adds a new function to the module.
func (ctx *ModuleContext) AddFunction(function *wyaml.FunctionYAML) (*FunctionDefinition, error) {
	// Prepare type data.
	typeDef, err := ctx.resolveNewFunctionType(function)
	if err != nil {
		return nil, fmt.Errorf("adding function: %w", err)
	}

	// Parses input and generates function code.
	functionIndex := ctx.generateFunctionIndex()
	code := wgenerator.FunctionToCode(function, functionIndex, typeDef.Name)

	// Adds function
	resDef, err := ctx.addFunction(functionIndex, code)
	if err != nil {
		return nil, fmt.Errorf("parsing function code to add on module context: %w", err)
	}
	return resDef, nil
}

// AddStartFunction adds start function.
func (ctx *ModuleContext) AddStartFunction(fnDef *FunctionDefinition) (*FunctionDefinition, error) {
	// Parses input and generates start function code.
	code := wgenerator.StartFunctionToCode(fnDef.Name)

	// Parses start function code.
	codeEl := NewCodeParser(code).parse()
	if len(codeEl.blocks) == 0 {
		return nil, errors.New("parsed code is empty")
	}

	// Adds start function element to module context.
	ctx.addBlocks(codeEl)
	return fnDef, nil
}

// AddImportFunction adds a new imported function to the module.
func (ctx *ModuleContext) AddImportFunction(function *wyaml.FunctionYAML) (*FunctionDefinition, error) {
	// Prepare type data.
	typeDef, err := ctx.resolveNewFunctionType(function)
	if err != nil {
		return nil, fmt.Errorf("adding import function: %w", err)
	}

	// Adds function
	return ctx.addImportFunction(typeDef, function.Imported.Module, function.Imported.Field)
}

// AddExportFunction adds a new exported function to the module.
func (ctx *ModuleContext) AddExportFunction(function *wyaml.FunctionYAML, fnDef *FunctionDefinition) (*FunctionDefinition, error) {
	// Parses input and generates function code.
	code := wgenerator.ExportFunctionToCode(fnDef.Name, *function.Exported)

	// Adds function
	resDef, err := ctx.addFunction(fnDef.Name, code)
	if err != nil {
		return nil, fmt.Errorf("parsing export function code to add on module context: %w", err)
	}
	return resDef, nil
}

// AddLocal adds a new local to the module.
func (ctx *ModuleContext) AddLocal(value string, fnDef *FunctionDefinition) (*FunctionLocalDefinition, error) {
	// Parses variable definition.
	expr, err := variable.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("parsing local variable value %q: %v", value, err)
	}

	// Generates an unique local index.
	localIndex := fmt.Sprintf("$%sl%d", wgenerator.CodeIndexPrefix, len(fnDef.Locals))
	return ctx.addLocal(localIndex, expr.GetType(), expr.GetValue(""), fnDef)
}

// SetStartFunction sets a new start function for the module.
func (ctx *ModuleContext) SetStartFunction(fn *FunctionDefinition) {
	fn.IsStart = true
	ctx.startFunction = fn
}

// ApplyRuntimeTransformations applies runtime transformations to module.
func (ctx *ModuleContext) ApplyRuntimeTransformations() {
	logrus.Infoln("Applying runtime modifications")

	// Find start function.
	startFnDef := ctx.startFunction
	hasStartFnDef := startFnDef != nil
	if !hasStartFnDef {
		fnDef, err := ctx.AddFunction(new(wyaml.FunctionYAML))
		if err != nil {
			logrus.Fatalf("creating new function to be the starting function: %v", err)
		}
		startFnDef = fnDef
		if _, err := ctx.AddStartFunction(startFnDef); err != nil {
			logrus.Fatalf("adding start function instruction: %v", err)
		}
	}

	// Add composite globals to start function.
	logrus.Traceln("Handling runtime composite globals")
	globalRuntimeVisitor := newGlobalRuntimeVisitor(ctx, startFnDef)
	ctx.entryBlock.Traverse(globalRuntimeVisitor)

	// Removes the start function if not needed.
	if !hasStartFnDef {
		if globalRuntimeVisitor.foundCount == 0 {
			module := startFnDef.instr.getParent().(*Instruction)
			err := module.removeChild(startFnDef.instr)
			if err != nil {
				logrus.Errorf("removing unnecessary start function: %v", err)
			}
			for i, b := range module.values {
				if bInstr, ok := b.(*Instruction); ok && bInstr.name == instructionStart {
					err := module.removeChildByIndex(i)
					if err != nil {
						logrus.Errorf("removing unnecessary start function instruction: %v", err)
					}
				}
			}
		} else {
			ctx.SetStartFunction(startFnDef)
		}
	}

	// Make sure every composite return is inside a return instruction.
	logrus.Traceln("Handling runtime composite returns")
	returnFuncVisitor := newCompositeReturnFuncVisitor(ctx)
	ctx.entryBlock.Traverse(returnFuncVisitor)

	// Make sure every instruction that is a call for a function that returns a composite value is suportted.
	logrus.Traceln("Handling runtime calls to functions with composite returns")
	callFuncReturnCompositeVisitor := newCallCompositeReturnFuncVisitor(returnFuncVisitor.fnDefs)
	ctx.entryBlock.Traverse(callFuncReturnCompositeVisitor)

	// Remove unknown web assembly types.
	logrus.Traceln("Removing runtime composite types from module")
	compositeVisitor := newCompositePropsVisitor()
	ctx.entryBlock.Traverse(compositeVisitor)

	// Apply runtime transformations.
	logrus.Traceln("Executing runtime expressions/references")
	runtimeVisitor := newRuntimeVisitor(ctx)
	ctx.entryBlock.Traverse(runtimeVisitor)

	// Change composite imports to operations module.
	logrus.Traceln("Changing composite imports to internal operations module")
	importCompositeFuncVisitor := newImportCompositeFuncVisitor(ctx)
	ctx.entryBlock.Traverse(importCompositeFuncVisitor)

	// Fixes bug caused by WABT when parsing wasm to wat generating names.
	// Elem name is generated wrongly.
	elemTableFixVisitor := newElemTableFixVisitor()
	ctx.entryBlock.Traverse(elemTableFixVisitor)

	// Remove the composite evaluations.
	logrus.Traceln("Removing all the runtime syntax elements")
	for _, eval := range ctx.evalsToRemove {
		parentBlock := eval.getParent()
		if parentBlock == nil {
			logrus.Errorf("instruction %s (composite) has no parent block", parentBlock)
			continue
		}
		parentInstr, ok := parentBlock.(*Instruction)
		if !ok {
			logrus.Errorf("parent of instruction %s (composite) is not an instruction block", parentInstr.name)
			continue
		}
		err := parentInstr.removeChild(eval)
		if err != nil {
			logrus.Errorf("removing composite evaluation instruction: %v", err)
		}
	}

	// Add glue code for the runtime changes work.
	logrus.Traceln("Adding glue functions to module")
	if err := ctx.addGlueFunctions(); err != nil {
		logrus.Fatalf("adding glue functions code: %v", err)
	}
}

// NewSearch initiates an empty join-point blocks search.
func (ctx *ModuleContext) NewSearch() *JoinPointSearch {
	search := newJoinPointSearch(ctx)
	return search
}

// InitSearch initiates a new join-point blocks search.
func (ctx *ModuleContext) InitSearch() *JoinPointSearch {
	search := newJoinPointSearch(ctx)
	ctx.entryBlock.Traverse(newFuncVisitor(ctx, search, emptyFuncFilterFn()))
	return search
}

// FindCalls searches the join-point blocks for some call definition.
func (ctx *ModuleContext) FindCalls(jpBlock *JoinPointBlock, callback CallFilterFn) *JoinPointSearch {
	search := newJoinPointSearch(ctx)
	jpBlock.block.Traverse(newCallVisitor(ctx, search, callback))
	return search
}

// FindFunctions searches the join-point blocks for some function definition.
func (ctx *ModuleContext) FindFunctions(jpBlock *JoinPointBlock, callback FuncFilterFn) *JoinPointSearch {
	search := newJoinPointSearch(ctx)
	jpBlock.block.Traverse(newFuncVisitor(ctx, search, callback))
	return search
}

// FindArgs searches the join-point blocks for some args definition.
func (ctx *ModuleContext) FindArgs(jpBlock *JoinPointBlock, callback ArgsFilterFn) *JoinPointSearch {
	search := newJoinPointSearch(ctx)
	jpBlock.block.Traverse(newArgsVisitor(ctx, search, callback))
	return search
}

// FindReturns searches the join-point blocks for some returns definition.
func (ctx *ModuleContext) FindReturns(jpBlock *JoinPointBlock, callback ReturnsFilterFn) *JoinPointSearch {
	search := newJoinPointSearch(ctx)
	jpBlock.block.Traverse(newReturnsVisitor(ctx, search, callback))
	return search
}

// FindInstructions searches for the join-point blocks that are equal to the provided code.
func (ctx *ModuleContext) FindInstructions(block Block, code string) *JoinPointSearch {
	resultSearch := newJoinPointSearch(ctx)

	// Parse new code to transform into instructions.
	newBlocks := NewCodeParser(code).parse().blocks

	// Use the found instructions to find the equivalent instruction on the existent code.
	parentBlock := block
	for _, block := range newBlocks {
		if found := parentBlock.findEqual(block); found != nil {
			if len(resultSearch.found) == 0 {
				parentBlock = found.getParent()
			}
			resultSearch.VisitBlock(found)
		}
	}

	return resultSearch
}

// Union returns the list of join-point blocks resultant from the union of two lists.
func (ctx *ModuleContext) Union(a, b []*JoinPointBlock) []*JoinPointBlock {
	var blocks []*JoinPointBlock
	blocks = append(blocks, a...)
	blocks = append(blocks, b...)
	newFound, _ := removeDuplicates(blocks, func(old, new *JoinPointBlock) *JoinPointBlock {
		if new.depth > old.depth {
			new.Metadata.Join(old.Metadata)
			return new
		}
		old.Metadata.Join(new.Metadata)
		return old
	})
	return newFound
}

// AliasValue returns the alias value for some key.
func (ctx *ModuleContext) AliasValue(key string) (string, bool) {
	res, ok := ctx.GlobalAlias[key]
	if ok {
		return res, ok
	}
	res, ok = ctx.FunctionAlias[key]
	return res, ok
}

// AliasKey returns the key alias for some value.
func (ctx *ModuleContext) AliasKey(value string) (string, bool) {
	for k, v := range ctx.GlobalAlias {
		if v == value {
			return k, true
		}
	}
	for k, v := range ctx.FunctionAlias {
		if v == value {
			return k, true
		}
	}
	return "", false
}

// setFunction sets a new function definition.
func (ctx *ModuleContext) setFunction(name string, fn *FunctionDefinition) {
	ctx.functions[name] = fn
}

// setImportFunction sets a new imported function definition.
func (ctx *ModuleContext) setImportFunction(moduleName, exportName string, fn *FunctionDefinition) {
	if moduleMap, ok := ctx.importFunctions[moduleName]; ok {
		moduleMap[exportName] = fn
		return
	}
	ctx.importFunctions[moduleName] = make(map[string]*FunctionDefinition)
	ctx.importFunctions[moduleName][exportName] = fn
}

// setExportFunction sets a new exported function definition.
func (ctx *ModuleContext) setExportFunction(exportName string, fn *FunctionDefinition) {
	ctx.exportFunctions[exportName] = fn
}

// setType sets a new type definition.
func (ctx *ModuleContext) setType(name string, typ *TypeDefinition) {
	ctx.types[name] = typ
}

// setGlobal sets a new global definition.
func (ctx *ModuleContext) setGlobal(name string, global *GlobalDefinition) {
	ctx.globals[name] = global
}

// setImportGlobal sets a new import global definition.
func (ctx *ModuleContext) setImportGlobal(moduleName, exportName string, global *GlobalDefinition) {
	if moduleMap, ok := ctx.importGlobals[moduleName]; ok {
		moduleMap[exportName] = global
		return
	}
	ctx.importGlobals[moduleName] = make(map[string]*GlobalDefinition)
	ctx.importGlobals[moduleName][exportName] = global
}

// setExportGlobal sets a new export global definition.
func (ctx *ModuleContext) setExportGlobal(exportName string, global *GlobalDefinition) {
	ctx.exportGlobals[exportName] = global
}

// generateFunctionIndex generates a unique function index.
func (ctx *ModuleContext) generateFunctionIndex() string {
	return fmt.Sprintf("$%sf%d", wgenerator.CodeIndexPrefix, len(ctx.functions))
}

// generateLocalIndex generates a unique local index.
func (ctx *ModuleContext) generateLocalIndex(typeName string, num int) string {
	return fmt.Sprintf("$%s%s_%s_%d", wgenerator.CodeIndexPrefix, instructionLocal, typeName, num)
}

// generateLocalIndex generates a unique local index.
func (ctx *ModuleContext) generateLocalIndexStr(typeName string, val string) string {
	return fmt.Sprintf("$%s%s_%s_%s", wgenerator.CodeIndexPrefix, instructionLocal, typeName, val)
}

// addCodeAtFuncStart adds code at the beginning of a function.
func (ctx *ModuleContext) addCodeAtFuncStart(code string, fnDef *FunctionDefinition) error {
	// Parses code.
	codeEl := NewCodeParser(code).parse()
	if len(codeEl.blocks) == 0 {
		return errors.New(" parsing code to add on function start")
	}

	// Adds new instructions to function.
	fnDef.addInstrsAtStart(codeEl.blocks)
	return nil
}

// addCodeAtFuncEnd adds code at the ending of a function.
func (ctx *ModuleContext) addCodeAtFuncEnd(code string, fnDef *FunctionDefinition) error {
	// Parses code.
	codeEl := NewCodeParser(code).parse()
	if len(codeEl.blocks) == 0 {
		return errors.New(" parsing code to add on function end")
	}

	// Adds new instructions to function.
	fnDef.addInstrsAtEnd(codeEl.blocks)
	return nil
}

// addFunction adds the code blocks of some function to the module tree.
func (ctx *ModuleContext) addFunction(index, code string) (*FunctionDefinition, error) {
	// Parses variable code.
	codeEl := NewCodeParser(code).parse()

	// Adds variable element to module context.
	ctx.addBlocks(codeEl)
	return ctx.functions[index], nil
}

// addImportFunction adds a new imported function to the module.
func (ctx *ModuleContext) addImportFunction(typeDef *TypeDefinition, moduleName, exportedName string) (*FunctionDefinition, error) {
	return ctx.addCustomImportFunction(ctx.generateFunctionIndex(), typeDef, moduleName, exportedName)
}

// addCustomImportFunction adds a new imported function with custom index to the module.
func (ctx *ModuleContext) addCustomImportFunction(functionIndex string, typeDef *TypeDefinition, moduleName, exportedName string) (*FunctionDefinition, error) {
	// Parses input and generates function code.
	code := wgenerator.ImportFunctionToCode(functionIndex, typeDef.Name, moduleName, exportedName)

	// Adds function
	resDef, err := ctx.addFunction(functionIndex, code)
	if err != nil {
		return nil, fmt.Errorf("parsing import function code to add on module context: %w", err)
	}
	return resDef, nil
}

// addType adds a new type to the module.
func (ctx *ModuleContext) addType(params []string, result string) (*TypeDefinition, error) {
	// Transforms definition into code.
	typeIndex := fmt.Sprintf("$%st%d", wgenerator.CodeIndexPrefix, len(ctx.types))
	code := wgenerator.FunctionTypeToCode(params, result, typeIndex)

	// Parses variable code.
	codeEl := NewCodeParser(code).parse()
	if len(codeEl.blocks) == 0 {
		return nil, errors.New("parsing type code to add on module context")
	}

	// Adds type element to module context.
	ctx.addBlocks(codeEl)
	return ctx.types[typeIndex], nil
}

// addLocal adds a new local to the module.
func (ctx *ModuleContext) addLocal(localIndex, localType, localValue string, fnDef *FunctionDefinition) (*FunctionLocalDefinition, error) {
	// Transforms definition into code.
	code := wgenerator.LocalVariableToCode(localIndex, localType)

	// Parses variable code.
	localCodeEl := NewCodeParser(code).parse()
	if len(localCodeEl.blocks) == 0 {
		return nil, errors.New("error parsing local code to add on function context")
	}
	localBlockInstr, ok := localCodeEl.blocks[0].(*Instruction)
	if !ok {
		return nil, errors.New("local code to add on function context is not an instruction")
	}
	localBlockInstr.setParent(fnDef.instr)

	// Parses variable value code.
	var valueBlock Block
	if IsVarTypeStrPrimitive(localType) && localValue != "" {
		// Transforms value set into code.
		valueCode := wgenerator.SetLocalToCode(localIndex, localType, localValue)

		// Parses set variable value code.
		valueCodeEl := NewCodeParser(valueCode).parse()
		if len(valueCodeEl.blocks) == 0 {
			return nil, errors.New("error parsing local value code to add on function context")
		}
		valueBlockInstr, ok := valueCodeEl.blocks[0].(*Instruction)
		if !ok {
			return nil, errors.New("local value code to add on function context is not an instruction")
		}
		valueBlock = valueBlockInstr
		valueBlock.setParent(fnDef.instr)
	}

	// Adds variable element to function context.
	local := newFunctionLocalDefinition(localBlockInstr, localIndex, localType, localValue, len(fnDef.Locals))
	fnDef.addLocal(local, valueBlock)
	return local, nil
}

// addBlocks adds the code block to the module tree.
// completes the context accordingly to the added blocks.
func (ctx *ModuleContext) addBlocks(entry *element) {
	// Finds module.
	moduleVisitor := newModuleInstrsVisitor()
	ctx.entryBlock.TraverseConditional(moduleVisitor)
	module, ok := moduleVisitor.Module()
	if !ok {
		return
	}

	// Adds blocks to the target context.
	for _, codeBlock := range entry.blocks {
		codeInstr, ok := codeBlock.(*Instruction)
		if !ok {
			module.values = append(module.values, codeBlock)
			continue
		}
		codeInstrWeight := instructionOrder(codeInstr)
		var index int
		for _, block := range module.values {
			if instructionOrder(block) > codeInstrWeight {
				break
			}
			index++
		}
		codeInstr.setParent(module)
		if index == len(module.values) {
			module.values = append(module.values, codeInstr)
		} else {
			var values []Block
			values = append(values, module.values[:index]...)
			values = append(values, codeInstr)
			values = append(values, module.values[index:]...)
			module.values = values
		}
		codeInstr.setParent(module)

		// Updates module context for the added block.
		ctx.FillContext(codeBlock)
	}
}

// resolveNewFunctionType resolves the type definition for some new function.
func (ctx *ModuleContext) resolveNewFunctionType(function *wyaml.FunctionYAML) (*TypeDefinition, error) {
	var typeDef *TypeDefinition
	for _, def := range ctx.types {
		if ok := compareArguments(function.Args, def.Params); !ok {
			continue
		}
		if function.Result != def.Result {
			continue
		}
		typeDef = def
		break
	}

	if typeDef == nil {
		typeDefAux, err := ctx.AddType(function)
		if err != nil {
			return nil, fmt.Errorf("creating new type for the adding function: %w", err)
		}
		typeDef = typeDefAux
	}
	return typeDef, nil
}

// typesMapSign creates a map with the type definitions.
// the map key is the type signature value.
func (ctx *ModuleContext) typesMapSign() map[string]*TypeDefinition {
	res := make(map[string]*TypeDefinition)
	for _, typ := range ctx.types {
		res[typ.signature()] = typ
	}
	return res
}

// addGlueFunctions adds the glue code to the module functions.
func (ctx *ModuleContext) addGlueFunctions() error {
	var types map[string]*TypeDefinition
	var needJS bool
	fnState := ctx.glueFunctions

	// Check operations functions.
	if fnState.operations {
		needJS = true
		types = ctx.typesMapSign()
		err := ctx.addGlueFunction(wgenerator.OperationFunctions(), types)
		if err != nil {
			return fmt.Errorf("operation glue functions: %w", err)
		}
	}

	// Check args functions.
	if len(fnState.args) > 0 {
		needJS = true
		if types == nil {
			types = ctx.typesMapSign()
		}
		err := ctx.addGlueFunction(wgenerator.ArgsFunctions(), types)
		if err != nil {
			return fmt.Errorf("args glue functions: %w", err)
		}
	}

	// Check zone functions.
	for _, changes := range ctx.runtimeChanges {
		if changes.hasNewZone {
			needJS = true
			if types == nil {
				types = ctx.typesMapSign()
			}
			err := ctx.addGlueFunction(wgenerator.ZoneFunctions(), types)
			if err != nil {
				return fmt.Errorf("zone glue functions: %w", err)
			}
			break
		}
	}

	// Check returns functions.
	if fnState.returns {
		needJS = true
		if types == nil {
			types = ctx.typesMapSign()
		}
		err := ctx.addGlueFunction(wgenerator.ReturnsFunctions(), types)
		if err != nil {
			return fmt.Errorf("returns glue functions: %w", err)
		}
	}

	// Check error functions.
	if fnState.err {
		needJS = true
		if types == nil {
			types = ctx.typesMapSign()
		}
		err := ctx.addGlueFunction(wgenerator.ErrorFunctions(), types)
		if err != nil {
			return fmt.Errorf("error glue functions: %w", err)
		}
	}
	ctx.NeedJS = needJS
	return nil
}

// addGlueFunction adds the glue code to a specific function.
func (ctx *ModuleContext) addGlueFunction(fns []*wgenerator.ImportFunctionDef, types map[string]*TypeDefinition) error {
	for _, fn := range fns {
		typeSignature := typeSignature(fn.Result, fn.Params)
		typeDef, ok := types[typeSignature]
		if !ok {
			typeDefAux, err := ctx.addType(fn.Params, fn.Result)
			if err != nil {
				return fmt.Errorf("adding new type: %w", err)
			}
			types[typeSignature] = typeDefAux
			typeDef = typeDefAux
		}
		_, err := ctx.addCustomImportFunction(fmt.Sprintf("$%s.%s", fn.ModuleName, fn.ExportedName), typeDef, fn.ModuleName, fn.ExportedName)
		if err != nil {
			return fmt.Errorf("adding import function: %w", err)
		}
	}
	return nil
}

// glueFunctionsState contains the presence state of the glue functions.
type glueFunctionsState struct {
	args       map[Block]struct{}
	operations bool
	returns    bool
	err        bool
}

// newGlueFunctionsState is a constructor for glueFunctionsState.
func newGlueFunctionsState() *glueFunctionsState {
	return &glueFunctionsState{args: make(map[Block]struct{})}
}

// compareArguments compares input arguments with the module parameters.
func compareArguments(args []wyaml.FunctionArgYAML, params []string) bool {
	if len(args) != len(params) {
		return false
	}
	for i, arg := range args {
		if arg.Type != params[i] {
			return false
		}
	}
	return true
}
