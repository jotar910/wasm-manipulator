package wcode

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"joao/wasm-manipulator/internal/wgenerator"
	"joao/wasm-manipulator/internal/wkeyword"
	"joao/wasm-manipulator/internal/wlang"
	"joao/wasm-manipulator/internal/wtemplate"
	"joao/wasm-manipulator/pkg/wutils"
)

// FuncFilterFn is the filter function prototype for func pointcuts.
type FuncFilterFn func(*ModuleContext, *FuncData) (map[string]wkeyword.Object, bool)

// emptyFuncFilterFn returns an empty implementation for FuncFilterFn.
func emptyFuncFilterFn() FuncFilterFn {
	return func(*ModuleContext, *FuncData) (map[string]wkeyword.Object, bool) {
		return make(map[string]wkeyword.Object), true
	}
}

// CallFilterFn is the filter function prototype for calls pointcuts.
type CallFilterFn func(*ModuleContext, *CallData) (map[string]wkeyword.Object, bool)

// ArgsFilterFn is the filter function prototype for args pointcuts.
type ArgsFilterFn func(*ModuleContext, *ArgsData) (map[string]wkeyword.Object, bool)

// ReturnsFilterFn is the filter function prototype for returns pointcuts.
type ReturnsFilterFn func(*ModuleContext, *ReturnsData) (map[string]wkeyword.Object, bool)

// PointcutVisitor is implemented by visitors used on pointcuts.
type PointcutVisitor interface {
	VisitFunc(*Instruction, *FuncData, map[string]wkeyword.Object)
	VisitCall(*Instruction, *CallData, map[string]wkeyword.Object)
	VisitArgs(*Instruction, *ArgsData, map[string]wkeyword.Object)
	VisitReturns(Block, *ReturnsData, map[string]wkeyword.Object)
}

// joinPointBlockSmartData contains the necessary data for the apply advice on smart mode..
type joinPointBlockSmartData struct {
	block      Block
	parent     *Instruction
	blockIndex int
	blockType  wlang.CodeBlockType
}

// newJoinPointBlockApplyData is a constructor for joinPointBlockSmartData.
func newJoinPointBlockApplyData(block Block, parent *Instruction, childIndex int, childType wlang.CodeBlockType) *joinPointBlockSmartData {
	return &joinPointBlockSmartData{block, parent, childIndex, childType}
}

// JoinPointBlock is the join-point block that adds metadata to the normal code blocks.
type JoinPointBlock struct {
	Metadata    wkeyword.Object
	Environment map[string]wkeyword.Object
	context     *ModuleContext
	block       Block
	function    *Instruction
	depth       int
}

// newJoinPointBlock is the implementation for JoinPointBlock.
func newJoinPointBlock(ctx *ModuleContext, b Block, f *Instruction, md wkeyword.Object, env map[string]wkeyword.Object) *JoinPointBlock {
	environment := make(map[string]wkeyword.Object, len(env))
	for k, v := range env {
		environment[wutils.CapitalizeFirstLetter(k)] = v
	}
	return &JoinPointBlock{
		Metadata:    md,
		Environment: environment,
		context:     ctx,
		function:    f,
		block:       b,
		depth:       0,
	}
}

// Is returns the type of keyword of some join point.
func (jpB *JoinPointBlock) Is(k string) wkeyword.KeywordType {
	_, typ, _ := jpB.Get(k)
	return typ
}

// Get returns the join-point value, for a given key, as a keyword value.
func (jpB *JoinPointBlock) Get(inK string) (interface{}, wkeyword.KeywordType, bool) {
	capsK := wutils.CapitalizeFirstLetter(inK)
	// Check metadata for Call, Args, Func, etc,
	if !wkeyword.IsNil(jpB.Metadata) {
		if v := jpB.Metadata.Prop(capsK); !wkeyword.IsNil(v) {
			return v, wkeyword.KeywordTypeObject, true
		}
	}
	// Check environment data for each instruction.
	if v, ok := jpB.Environment[capsK]; ok {
		return v, wkeyword.KeywordTypeObject, true
	}
	return nil, wkeyword.KeywordTypeUnknown, false
}

// String returns the string description for the join-point block.
func (jpB *JoinPointBlock) String() string {
	return fmt.Sprintf("fn: %s, depth: %d", jpB.function.values[0].String(), jpB.depth)
}

// FuncDefinition returns the join-point function definition.
func (jpB *JoinPointBlock) FuncDefinition() *FunctionDefinition {
	if jpB.function == nil {
		return nil
	}
	if def, ok := jpB.context.functions[jpB.function.values[0].String()]; ok {
		return def
	}
	return nil
}

// Instr returns the instruction block of the join-point block.
func (jpB *JoinPointBlock) Instr() Block {
	return jpB.block
}

// applyNonSmart applies non smart modification
func (jpB *JoinPointBlock) applyNonSmart(fnDef *FunctionDefinition, code string) error {
	resultType, err := resolveResultType(jpB.context, fnDef, jpB.block)
	if err != nil {
		logrus.Warn(err)
	}
	if resultType != "" && resultType != wlang.Any {
		logrus.Warnf("Applying advice on instruction that requires a result type %s (Code: %s)", resultType, wtemplate.ClearString(code))
	}
	return replaceBlockWithCode(jpB.block, code)
}

// generateLocal generates local variable
func (jpB *JoinPointBlock) generateLocal(fnDef *FunctionDefinition, smartData *joinPointBlockSmartData) (*FunctionLocalDefinition, error) {
	blockIndex, blockType := smartData.blockIndex, smartData.blockType
	// Generates a local index to handle the result.
	localIndex := fmt.Sprintf("$%sl%d_%s", wgenerator.CodeIndexPrefix, blockIndex, blockType)
	if fnLocal, ok := fnDef.Locals[localIndex]; ok {
		return fnLocal, nil
	}
	fnLocal, err := jpB.context.addLocal(localIndex, string(blockType), "", fnDef)
	if err != nil {
		return nil, fmt.Errorf("adding local %s to function to handle the result value: %w", localIndex, err)
	}
	if smartData.parent.name == instructionFunction {
		smartData.blockIndex++
	}
	return fnLocal, nil
}

// Instr returns the instruction block of the join-point block.
func (jpB *JoinPointBlock) Apply(code string, smart bool) error {
	codeEl := NewCodeParser(code).parse()

	// No need for smart changes.
	if len(codeEl.blocks) == 1 {
		return replaceBlock(jpB.block, codeEl.blocks)
	}

	// Gets the function definition.
	fnDef, ok := jpB.context.Function(jpB.function.values[0].String())
	if !ok {
		return errors.New("finding block function definition")
	}

	// Apply code on non-smart mode.
	if !smart {
		return jpB.applyNonSmart(fnDef, code)
	}

	// Checks expected type for the block result.
	smartData, err := resolveApplySmartData(jpB.context, fnDef, jpB.block)
	if err != nil {
		return fmt.Errorf("finding result type for the block: %w", err)
	}

	// Generate the blocks and find the target elements.
	targetVisitor := newTargetInstrsVisitor()
	codeEl.Traverse(targetVisitor)

	// Replace whole code if it is already on a control flow instruction.
	if smartData == nil || wlang.IsControlFlow(smartData.parent.name) {
		for _, target := range targetVisitor.targets {
			if err := replaceBlock(target, target.values); err != nil {
				return fmt.Errorf("replacing target instruction with no result value by the child values: %w", err)
			}
		}
		return replaceBlock(jpB.block, codeEl.blocks)
	}

	var fnLocal *FunctionLocalDefinition

	if len(targetVisitor.targets) == 0 { // No targets found. All the code will be set to the local variable.
		logrus.Warn("Target not found on smart mode. Searching for the original instruction...")
		target := codeEl.findEqual(jpB.block)
		if target == nil {
			logrus.Warn("Original instruction not found on the new code...")
			return jpB.applyNonSmart(fnDef, code)
		}

		// Generates a local index to handle the result.
		fnLocal, err = jpB.generateLocal(fnDef, smartData)
		if err != nil {
			return fmt.Errorf("generating local when no targets exist: %w", err)
		}

		// Generate local.set instruction.
		setLocalEl := NewCodeParser(wgenerator.SetLocalInstructionToCode(fnLocal.Name, target.String())).parse()
		if err := replaceBlock(target, setLocalEl.blocks); err != nil {
			return fmt.Errorf("saving the original instruction in local variable: %w", err)
		}
	} else {
		// Generates a local index to handle the result.
		fnLocal, err = jpB.generateLocal(fnDef, smartData)
		if err != nil {
			return fmt.Errorf("generating local when a target is defined: %w", err)
		}

		// Remove first targets.
		for i := 0; i < len(targetVisitor.targets)-1; i++ {
			target := targetVisitor.targets[i]
			if err := replaceBlock(target, target.values); err != nil {
				return fmt.Errorf("replacing target instruction with no result value by the child values: %w", err)
			}
		}
		// Update last target to be set on the set.local instruction.
		lastTarget := targetVisitor.targets[len(targetVisitor.targets)-1]
		lastTarget.name = instructionCodeSetLocal
		lastTargetName := newText(fnLocal.Name)
		lastTargetName.setParent(lastTarget)
		lastTarget.values = append([]Block{lastTargetName}, lastTarget.values...)
	}

	// AddBlocksToControlFlow adds the code to the control flow instruction.
	if err = AddBlocksToControlFlow(jpB.block, codeEl.blocks, 0); err != nil {
		return fmt.Errorf("adding new code to the control flow instruction: %w", err)
	}

	// NewCodeParser generates local.get instruction and replace block with it.
	getLocalEl := NewCodeParser(wgenerator.GetVariableToCode(fnLocal.Name, true)).parse()
	replaceIndex := smartData.blockIndex
	if smartData.parent.name == instructionFunction {
		// After the new blocks
		replaceIndex += len(codeEl.blocks)
	}
	err = smartData.parent.replaceChildByIndex(replaceIndex, getLocalEl.blocks)
	if err != nil {
		return fmt.Errorf("replacing block with the respective local: %w", err)
	}

	// Resolve new instruction start point.
	startPointBlock := jpB.block
	if len(codeEl.blocks) > 0 {
		startPointBlock = codeEl.blocks[0]
	}

	// Reorder the whole instruction above the block.
	if smartData.blockIndex != 0 {
		if _, ok := smartData.parent.values[smartData.blockIndex-1].(*text); ok {
			return nil
		}
		return reorderOnSmartApply(jpB.context, fnDef, smartData.parent.values[smartData.blockIndex-1], startPointBlock, smartData.block, "", false)
	}
	return reorderOnSmartApply(jpB.context, fnDef, smartData.parent, startPointBlock, smartData.block, "p", true)
}

// reorderOnSmartApply reorders all the instructions when applying advice on smart mode.
func reorderOnSmartApply(context *ModuleContext, fnDef *FunctionDefinition, block Block, startPointBlock Block, endPointBlock Block, levelId string, useEndPointBlock bool) error {
	// Checks expected type for the block result.
	applyData, err := resolveApplySmartData(context, fnDef, block)
	if err != nil {
		return fmt.Errorf("finding result type for the block %s: %w", block, err)
	}

	// Stops when the parent is a base instruction or there is no relevant information about it.
	if applyData == nil || applyData.parent.name == instructionFunction || wlang.IsControlFlow(applyData.parent.name) {
		return nil
	}

	// Generates a local index to handle the result.
	localIndex := fmt.Sprintf("$%sl%d%s_%s", wgenerator.CodeIndexPrefix, applyData.blockIndex, levelId, applyData.blockType)
	fnLocal, ok := fnDef.Locals[localIndex]
	if !ok {
		fnLocal, err = context.addLocal(localIndex, string(applyData.blockType), "", fnDef)
		if err != nil {
			return fmt.Errorf("adding local %s to function to handle the result value: %w", localIndex, err)
		}
	}

	// Resolve the point block where the code will be moved.
	pointBlock := startPointBlock
	if useEndPointBlock {
		pointBlock = endPointBlock
	}

	// Generate local.set instruction and
	// add the code to the control flow instruction.
	codeEl := NewCodeParser(wgenerator.SetLocalInstructionToCode(fnLocal.Name, "")).parse()
	if err = AddBlocksToControlFlow(pointBlock, codeEl.blocks, 0); err != nil {
		return fmt.Errorf("moving the block to a control flow instruction: %w", err)
	}
	block.setParent(codeEl.blocks[0])
	codeEl.blocks[0].(*Instruction).values = append(codeEl.blocks[0].(*Instruction).values, block)

	// Generate local.get instruction and replace block with it.
	getLocalEl := NewCodeParser(wgenerator.GetVariableToCode(fnLocal.Name, true)).parse()
	err = applyData.parent.replaceChildByIndex(applyData.blockIndex, getLocalEl.blocks)
	if err != nil {
		return fmt.Errorf("replacing the moved block with the respective local: %w", err)
	}

	// If it is not using the end point, it must update the start point.
	if !useEndPointBlock && len(codeEl.blocks) != 0 {
		startPointBlock = codeEl.blocks[0]
	}

	// Reorder the block above the current one.
	if applyData.blockIndex != 0 {
		if _, ok := applyData.parent.values[applyData.blockIndex-1].(*text); ok {
			return nil
		}
		return reorderOnSmartApply(context, fnDef, applyData.parent.values[applyData.blockIndex-1], startPointBlock, endPointBlock, levelId, false)
	}
	// Reorder the parent block when no block is above this.
	return reorderOnSmartApply(context, fnDef, applyData.parent, startPointBlock, endPointBlock, levelId+"p", true)
}

// resolveResultType returns some instruction result type.
func resolveResultType(context *ModuleContext, fnDef *FunctionDefinition, block Block) (wlang.CodeBlockType, error) {
	if block == nil {
		return "", errors.New("resolving result type from a null block")
	}

	// Check block instruction to get the result type.
	blockInstr, ok := block.(*Instruction)
	if !ok {
		return "", nil
	}
	switch n := blockInstr.name; {
	case n == instructionFunction, n == instructionCodeReturn:
		return wlang.CodeBlockType(fnDef.Result), nil
	case n == instructionCodeCall, n == instructionCodeCallIndirect:
		return resolveResultType(context, fnDef, block.getParent())
	case n == instructionCodeTeeLocal, n == instructionCodeSetLocal, n == instructionCodeGetLocal:
		localName := blockInstr.values[0].String()
		if local, ok := fnDef.Locals[localName]; ok {
			return wlang.CodeBlockType(local.Type), nil
		}
		if param, ok := fnDef.Params[localName]; ok {
			return wlang.CodeBlockType(param.Type), nil
		}
		return "", fmt.Errorf("could not find local with name %s", localName)
	case n == instructionCodeGetLocal, n == instructionCodeSetGlobal:
		globalName := blockInstr.values[0].String()
		if global, ok := context.globals[globalName]; ok {
			return wlang.CodeBlockType(global.Type), nil
		}
		return "", fmt.Errorf("could not find global with name %s", globalName)
	}
	blockDef, ok := wlang.GetInstrDefinition(blockInstr.name)
	if !ok || blockDef.NReturns == 0 {
		return "", nil
	}
	if blockDef.Returns[0] != wlang.Any {
		return wlang.CodeBlockType(fnDef.Result), nil
	}

	// Check parent instruction to get the argument type.
	parent := block.getParent()
	if parent == nil {
		return "", errors.New("block does not have a parent block")
	}
	parentInstr, ok := parent.(*Instruction)
	if !ok || wlang.IsControlFlow(parentInstr.name) {
		return wlang.Any, nil
	}
	if parentInstr.name == instructionFunction || parentInstr.name == instructionCodeReturn {
		return wlang.CodeBlockType(fnDef.Result), nil
	}
	parentDef, ok := wlang.GetInstrDefinition(parentInstr.name)
	if !ok {
		return "", nil
	}
	index := parentInstr.childIndex(block)
	if index == -1 {
		return "", errors.New("not found child block on parent values")
	}
	if index < parentDef.NArgs && parentDef.Args[index] != wlang.Any {
		return parentDef.Args[index], nil
	}

	// Keep searching on its parent.
	return resolveResultType(context, fnDef, parent)
}

// resolveApplySmartData returns the necessary data for the apply advice operation when it is using the smart mode.
func resolveApplySmartData(context *ModuleContext, fnDef *FunctionDefinition, block Block) (*joinPointBlockSmartData, error) {
	if false {
		return nil, nil
	}

	parent := block.getParent()
	if parent == nil {
		return nil, errors.New("block does not have a parent block")
	}
	parentInstr, ok := parent.(*Instruction)
	if !ok || wlang.IsControlFlow(parentInstr.name) {
		return nil, nil
	}
	index := parentInstr.childIndex(block)
	if index == -1 {
		return nil, errors.New("not found child block on parent values")
	}

	blockType, err := resolveResultType(context, fnDef, block)
	if err != nil {
		return nil, fmt.Errorf("resolving block type: %w", err)
	}
	if blockType == "" {
		return nil, nil
	}
	return newJoinPointBlockApplyData(block, parentInstr, index, blockType), nil
}

// JoinPointSearch is responsible to control the join-points search.
type JoinPointSearch struct {
	context *ModuleContext
	found   []*JoinPointBlock
}

// newJoinPointSearch is a constructor for JoinPointSearch.
func newJoinPointSearch(ctx *ModuleContext) *JoinPointSearch {
	return &JoinPointSearch{context: ctx}
}

// VisitBlock handles some code block.
func (js *JoinPointSearch) VisitBlock(b Block) {
	bInstr, ok := b.(*Instruction)
	if !ok {
		js.found = append(js.found, newJoinPointBlock(js.context, b, nil, wkeyword.NewKwNil(), make(map[string]wkeyword.Object)))
		return
	}
	jpBlock, err := js.createJoinPointBlock(js.context, bInstr, bInstr, wkeyword.NewKwNil(), make(map[string]wkeyword.Object))
	if err != nil {
		logrus.Fatalf("general instruction block found: %v", err)
	}
	js.found = append(js.found, jpBlock)
}

// VisitFunc handles some function instruction.
func (js *JoinPointSearch) VisitFunc(b *Instruction, data *FuncData, env map[string]wkeyword.Object) {
	jpBlock, err := js.createJoinPointBlock(js.context, b, b, newFuncMetadataObject(data), env)
	if err != nil {
		logrus.Fatalf("function block found: %v", err)
	}
	js.found = append(js.found, jpBlock)
}

// VisitCall handles some call instruction.
func (js *JoinPointSearch) VisitCall(b *Instruction, data *CallData, env map[string]wkeyword.Object) {
	jpBlock, err := js.createJoinPointBlock(js.context, b, b, newCallMetadataObject(data), env)
	if err != nil {
		logrus.Fatalf("call block found: %v", err)
	}
	js.found = append(js.found, jpBlock)
}

// VisitArgs handles some arg instruction.
func (js *JoinPointSearch) VisitArgs(b *Instruction, data *ArgsData, env map[string]wkeyword.Object) {
	jpBlock, err := js.createJoinPointBlock(js.context, b, b, newArgsMetadataObject(data), env)
	if err != nil {
		logrus.Fatalf("args block found: %v", err)
	}
	js.found = append(js.found, jpBlock)
}

// VisitArgs handles some arg instruction.
func (js *JoinPointSearch) VisitReturns(b Block, data *ReturnsData, env map[string]wkeyword.Object) {
	jpBlock, err := js.createJoinPointBlock(js.context, b, b.(*Instruction), newReturnsMetadataObject(data), env)
	if err != nil {
		logrus.Fatalf("returns block found: %v", err)
	}
	js.found = append(js.found, jpBlock)
}

// Found returns the found join-point blocks on the search.
func (js *JoinPointSearch) Found() []*JoinPointBlock {
	return js.found
}

// RemoveDuplicates deletes duplicated found blocks.
func (js *JoinPointSearch) RemoveDuplicates() bool {
	newFound, hasDuplicates := removeDuplicates(js.found, func(old, new *JoinPointBlock) *JoinPointBlock { return old })
	js.found = newFound
	return hasDuplicates
}

// Merge merges a join-point search to this one.
func (js *JoinPointSearch) Merge(o *JoinPointSearch) {
	js.found = append(js.found, o.found...)
	js.RemoveDuplicates()
}

// createJoinPointBlock creates a join-point block.
func (js *JoinPointSearch) createJoinPointBlock(ctx *ModuleContext, block Block, parent *Instruction,
	data wkeyword.Object, env map[string]wkeyword.Object) (*JoinPointBlock, error) {
	jpBlock := newJoinPointBlock(ctx, block, parent, data, env)
	for jpBlock.function.name != instructionFunction {
		jpBlock.depth++
		if jpBlock.function.parent == nil {
			return nil, errors.New("instruction is outside module")
		}
		funcInstrAux, ok := jpBlock.function.parent.(*Instruction)
		if !ok {
			return nil, errors.New("instruction is inside of an invalid block")
		}
		jpBlock.function = funcInstrAux
	}
	return jpBlock, nil
}

// callMetadata is the metadata model for func.
type funcMetadata struct {
	Func *FuncData
}

// newFuncMetadataObject is a constructor for funcMetadata.
func newFuncMetadataObject(d *FuncData) wkeyword.Object {
	return wkeyword.NewKwObject(&funcMetadata{d})
}

// callMetadata is the metadata model for call.
type callMetadata struct {
	Call *CallData
}

// newCallMetadataObject is a constructor for callMetadata.
func newCallMetadataObject(d *CallData) wkeyword.Object {
	return wkeyword.NewKwObject(&callMetadata{d})
}

// argsMetadata is the metadata model for args.
type argsMetadata struct {
	Args *ArgsData
}

// newArgsMetadataObject is a constructor for argsMetadata.
func newArgsMetadataObject(d *ArgsData) wkeyword.Object {
	return wkeyword.NewKwObject(&argsMetadata{d})
}

// returnsMetadata is the metadata model for returns.
type returnsMetadata struct {
	Returns *ReturnsData
}

// newReturnsMetadataObject is a constructor for returnsMetadata.
func newReturnsMetadataObject(d *ReturnsData) wkeyword.Object {
	return wkeyword.NewKwObject(&returnsMetadata{d})
}

// CollisionResolverFn is the function prototype for the collision resolver callback.
type CollisionResolverFn func(old, new *JoinPointBlock) *JoinPointBlock

// removeDuplicates removes duplicated join-point blocks.
// it uses a callback function to select which block must result from a collision.
func removeDuplicates(inBlocks []*JoinPointBlock, collisionFn CollisionResolverFn) ([]*JoinPointBlock, bool) {
	var hasDuplicates bool
	var newFound []*JoinPointBlock
	m := make(map[string][]*JoinPointBlock)
	for _, jpBlock := range inBlocks {
		fName := jpBlock.function.name
		m[fName] = append(m[fName], jpBlock)
	}
	for _, jpBlocks := range m {
		minDepth := jpBlocks[0].depth
		for i := 1; i < len(jpBlocks); i++ {
			if jpBlocks[i].depth < minDepth {
				minDepth = jpBlocks[i].depth
			}
		}
		hlBlocks := make(map[Block]*JoinPointBlock)
		for _, jpBlock := range jpBlocks {
			b := jpBlock.block
			for d := jpBlock.depth; d > minDepth; d-- {
				b = b.getParent()
			}
			if cur, ok := hlBlocks[b]; ok {
				hlBlocks[b] = collisionFn(cur, jpBlock)
				hasDuplicates = true
				continue
			}
			hlBlocks[b] = jpBlock
		}
		for _, jpBlock := range hlBlocks {
			newFound = append(newFound, jpBlock)
		}
	}
	return newFound, hasDuplicates
}
