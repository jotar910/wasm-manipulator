package wcode

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"joao/wasm-manipulator/internal/wgenerator"
	"joao/wasm-manipulator/internal/wlang"
	"joao/wasm-manipulator/pkg/wutils"
)

var variableIndexReg = regexp.MustCompile(`[\w\d_]+\[`)

// evaluationType represents the element type of the evaluation.
type evaluationType string

const (
	evaluationTypeGlobal evaluationType = "global"
	evaluationTypeLocal  evaluationType = "local"
	evaluationTypeCall   evaluationType = "call"
	evaluationTypeReturn evaluationType = "return"
)

// evaluationZoneType represents the zone type of the evaluation.
type evaluationZoneType string

const (
	evaluationZoneTypeGlobal evaluationZoneType = "global"
	evaluationZoneTypeLocal  evaluationZoneType = "local"
	evaluationZoneTypeParam  evaluationZoneType = "param"
)

// AliasDefinition is implemented dictionaries which contains some key-value aliases.
type AliasDefinition interface {
	AliasKey(value string) (string, bool)
	AliasValue(key string) (string, bool)
}

// variableReferenceData contains the data for some variable reference.
type variableReferenceData struct {
	name  string
	key   string
	index string
}

// newVariableReferenceData is a constructor for variableReferenceData.
func newVariableReferenceData(name, key, index string) *variableReferenceData {
	return &variableReferenceData{
		name:  name,
		key:   key,
		index: index,
	}
}

// evaluationArgumentTarget represents the target for the call argument evaluation.
type evaluationArgumentTarget struct {
	callee           *FunctionDefinition
	typ              variableType
	callArgIndex     int
	calleeParamIndex int
}

// newEvaluationArgumentTarget is a constructor for evaluationArgumentTarget.
func newEvaluationArgumentTarget(callee *FunctionDefinition, typ variableType, callArgIndex int, calleeParamIndex int) *evaluationArgumentTarget {
	return &evaluationArgumentTarget{
		callee:           callee,
		typ:              typ,
		callArgIndex:     callArgIndex,
		calleeParamIndex: calleeParamIndex,
	}
}

// evaluationTarget represents the target for the variable set evaluation.
type evaluationTarget struct {
	alias string
	key   string
	index string
	typ   varType
}

// newEvaluationTarget is a constructor for evaluationTarget.
func newEvaluationTarget(alias, index string, typ varType) *evaluationTarget {
	return &evaluationTarget{alias, "", index, typ}
}

// newEvaluationTarget is a constructor for evaluationTarget.
func newEvaluationTargetMember(alias, key, index string, typ varType) *evaluationTarget {
	return &evaluationTarget{alias, key, index, typ}
}

// addedEvaluationTarget contains the information of the added evaluation variable target.
type addedEvaluationTarget struct {
	*evaluationTarget
	added bool
}

// newAddedEvaluationTarget is a constructor for addedEvaluationTarget.
func newAddedEvaluationTarget(target *evaluationTarget) *addedEvaluationTarget {
	return &addedEvaluationTarget{target, false}
}

// evaluationZoneTarget is implemented by the target zones that must suffer modifications.
// this type of modifications may be added at the beginning of the function and can be same for several evaluations.
// an example is the variable definition on the javascript scope because of its presence on the evaluation.
type evaluationZoneTarget interface {
	apply(ctx *ModuleContext, changes *runtimeChanges, fnDef *FunctionDefinition, fnInstr *Instruction, instr Block, index int) error
	code(opCount int) (string, int)
}

// evaluationZoneTarget represents a code target for primitive variable types.
type evalPrimitiveZoneTarget struct {
	key      string
	varIndex string
	varType  variableType
	zoneType evaluationZoneType
}

// newEvalPrimitiveZoneTarget is a constructor for evalPrimitiveZoneTarget.
func newEvalPrimitiveZoneTarget(key, varIndex string, varType variableType, zoneType evaluationZoneType) *evalPrimitiveZoneTarget {
	return &evalPrimitiveZoneTarget{
		key:      key,
		varIndex: varIndex,
		varType:  varType,
		zoneType: zoneType,
	}
}

// apply applies the modifications.
func (target *evalPrimitiveZoneTarget) apply(_ *ModuleContext, changes *runtimeChanges, _ *FunctionDefinition, _ *Instruction, instr Block, _ int) error {
	// Parse new code blocks.
	code, newOpCount := target.code(changes.operationsCount)
	newBlockEl := NewCodeParser(code).parse()

	// Add the value to the parent instruction.
	AddBlocksToControlFlow(instr, newBlockEl.blocks, 0)
	changes.operationsCount = newOpCount
	return nil
}

// code returns the code for the modifications.
func (target *evalPrimitiveZoneTarget) code(opCount int) (string, int) {
	typ := target.varType
	return wgenerator.GetZonePrimitiveCode(target.key, target.varIndex, string(typ.Type()), typ.Code(), target.zoneType != evaluationZoneTypeGlobal, opCount)
}

// evalCompositeZoneTarget represents a code target for composite variable types.
type evalCompositeZoneTarget struct {
	name     string
	key      string
	value    string
	typeCode int
	zoneType evaluationZoneType
}

// newEvalCompositeZoneTarget is a constructor for evalCompositeZoneTarget.
func newEvalCompositeZoneTarget(name, key, value string, typeCode int, zoneType evaluationZoneType) *evalCompositeZoneTarget {
	return &evalCompositeZoneTarget{
		name:     name,
		key:      key,
		value:    value,
		typeCode: typeCode,
		zoneType: zoneType,
	}
}

// apply applies the modifications.
func (target *evalCompositeZoneTarget) apply(ctx *ModuleContext, changes *runtimeChanges, fnDef *FunctionDefinition, _ *Instruction, _ Block, _ int) error {
	if _, ok := changes.addedCompositeZones[target.name]; ok {
		return nil
	}
	changes.addedCompositeZones[target.name] = struct{}{}
	code, newOpCount := target.code(changes.operationsCount)
	if code == "" {
		return nil
	}
	if err := ctx.addCodeAtFuncStart(code, fnDef); err != nil {
		logrus.Fatalf("adding zone code for string variable %v to zone: %v", target, err)
	}
	changes.operationsCount = newOpCount
	return nil
}

// code returns the code for the modifications.
func (target *evalCompositeZoneTarget) code(opCount int) (string, int) {
	switch target.zoneType {
	case evaluationZoneTypeLocal:
		return wgenerator.GetZoneCompositeLocalCode(target.name, target.key, target.value, target.typeCode, opCount)
	case evaluationZoneTypeParam:
		return wgenerator.GetZoneCompositeParamCode(target.name, target.value, opCount)
	case evaluationZoneTypeGlobal:
		// Empty by design. Global zone must already be defined.
	default:
		logrus.Errorf("cannot generate code for string zone operation: invalid zone type %s", target.zoneType)
	}
	return "", -1
}

// runtimeChanges contains all the modifications data for some function change.
type runtimeChanges struct {
	fnDef               *FunctionDefinition
	hasNewZone          bool
	localsToAdd         map[string]*addedEvaluationTarget
	addedCompositeZones map[string]struct{} // It is only necessary a string zone variables definition per function. String values are outside module context.
	operationsCount     int
}

// newRuntimeChanges is a constructor for runtimeChanges.
func newRuntimeChanges(fnDef *FunctionDefinition) *runtimeChanges {
	return &runtimeChanges{
		fnDef:               fnDef,
		hasNewZone:          false,
		localsToAdd:         make(map[string]*addedEvaluationTarget),
		addedCompositeZones: make(map[string]struct{}),
		operationsCount:     0,
	}
}

// runtimeVisitor is used to visit all the evaluations on the module and proceed to the code modifications.
type runtimeVisitor struct {
	*visitorAdapter
	ctx             *ModuleContext
	glueFunctions   *glueFunctionsState
	functionChanges map[string]*runtimeChanges
}

// newRuntimeVisitor is a constructor for runtimeVisitor.
func newRuntimeVisitor(ctx *ModuleContext) *runtimeVisitor {
	return &runtimeVisitor{
		visitorAdapter:  new(visitorAdapter),
		ctx:             ctx,
		glueFunctions:   ctx.glueFunctions,
		functionChanges: ctx.runtimeChanges,
	}
}

// VisitEvaluation visits an evaluation block.
func (rv *runtimeVisitor) VisitEvaluation(eval *evaluation) bool {
	// Initialize evaluation visit.
	visitData, err := initVisit(newInitVisitInputData(rv, eval))
	if err != nil {
		logrus.Errorf("initiating evaluation visit: %v", err)
		return false
	}
	if visitData == nil {
		logrus.Trace("trying to evaluate deleted block")
		return false
	}

	// Check if function needs to have a zone.
	fnInstr, evalIndex := visitData.fnInstr, visitData.fnChildIndex
	fnDef, changes := visitData.fnDef, visitData.changes
	if vars, ok := rv.checkZoneVariables(fnDef, eval.blocks); ok {
		for _, v := range vars {
			err := v.apply(rv.ctx, changes, fnDef, fnInstr, eval, evalIndex)
			if err != nil {
				logrus.Fatalf("applying zone transformation: %v", err)
			}
		}
		if !changes.hasNewZone {
			if err := rv.ctx.addCodeAtFuncStart(instructionCallZonePush, fnDef); err != nil {
				logrus.Fatalf("adding entry zone code: %v", err)
			}
			if err := rv.ctx.addCodeAtFuncEnd(instructionCallZonePop, fnDef); err != nil {
				logrus.Fatalf("adding exit zone code: %v", err)
			}
			changes.hasNewZone = true
		}
	}

	// Process evaluation transformations accordingly to the operation type.
	switch visitData.operationType {
	case evaluationTypeCall:
		err = rv.visitCallEvaluation(fnDef, changes, eval)
	case evaluationTypeLocal:
		err = rv.visitLocalEvaluation(fnDef, changes, eval)
	case evaluationTypeGlobal:
		err = rv.visitGlobalEvaluation(changes, eval)
	case evaluationTypeReturn:
		err = rv.visitReturnEvaluation(fnDef, changes, eval)
	default:
		if err := rv.visitOtherEvaluation(fnDef, changes, eval); err != nil {
			logrus.Errorf("processing unknown runtime evaluation %s: %v", eval.String(), err)
			return false
		}
		err = nil
	}

	// Check if an error occurred in evaluation process.
	if err != nil {
		logrus.Errorf("processing runtime evaluation %s (type %s): %v", eval.String(), visitData.operationType, err)
		return false
	}

	// Add locals to function.
	for name, local := range changes.localsToAdd {
		if local.added {
			continue
		}
		if localType := local.typ; isVarTypePrimitive(localType) {
			_, err := rv.ctx.addLocal(name, string(localType), "", fnDef)
			if err != nil {
				logrus.Errorf("adding local for runtime evaluation: %v", err)
			}
		}
		local.added = true
	}

	return true
}

// VisitEvaluationRef visits an evaluation reference block.
func (rv *runtimeVisitor) VisitEvaluationRef(ref *evaluationRef) bool {
	// Initialize evaluation ref visit.
	visitData, err := initVisit(newInitVisitInputData(rv, ref))
	if err != nil {
		logrus.Errorf("initiating evaluation reference visit: %v", err)
		return false
	}
	if visitData == nil || visitData.operationType != evaluationTypeReturn {
		return false
	}

	// Process evaluation reference.
	fnDef, changes := visitData.fnDef, visitData.changes
	if err = rv.visitReturnEvaluationRef(fnDef, changes, ref); err != nil {
		logrus.Errorf("processing runtime evaluation reference %s (type %s): %v", ref.String(), visitData.operationType, err)
		return false
	}

	// Get zone variable to be added on JS.
	zoneVar, err := rv.getZoneVariable(fnDef, strings.TrimLeft(ref.String(), "#"))
	if err != nil {
		logrus.Errorf("getting zone variable ref: %s", err)
		return false
	}

	if zoneVar == nil {
		return true
	}

	// Add zone variable to JS.
	fnInstr, evalRefIndex := visitData.fnInstr, visitData.fnChildIndex
	err = zoneVar.apply(rv.ctx, changes, fnDef, fnInstr, ref, evalRefIndex)
	if err != nil {
		logrus.Fatalf("applying zone transformation: %v", err)
	}

	// Add the zone.push/zone.pop on function.
	if !changes.hasNewZone {
		if err := rv.ctx.addCodeAtFuncStart(instructionCallZonePush, fnDef); err != nil {
			logrus.Fatalf("adding entry zone code: %v", err)
		}
		if err := rv.ctx.addCodeAtFuncEnd(instructionCallZonePop, fnDef); err != nil {
			logrus.Fatalf("adding exit zone code: %v", err)
		}
		changes.hasNewZone = true
	}

	return true
}

// checkZoneVariables returns a list of zone targets for the evaluation.
func (rv *runtimeVisitor) checkZoneVariables(fnDef *FunctionDefinition, blocks []Block) ([]evaluationZoneTarget, bool) {
	var res []evaluationZoneTarget
	var hasVars bool
	for _, b := range blocks {
		switch block := b.(type) {
		case *evaluationId:
			value := block.value
			zoneVar, err := rv.getZoneVariable(fnDef, value)
			if err != nil {
				logrus.Errorf("getting zone reference id %s: %s", value, err)
				break
			}
			if zoneVar != nil {
				res = append(res, zoneVar)
			}
			hasVars = true
		case *evaluationIndex:
			var isIdentifier bool
			value := block.IndexString()
			for _, c := range value {
				if !wutils.IsIdentifier(c) && c != '$' {
					isIdentifier = false
					break
				}
				isIdentifier = true
			}
			if !isIdentifier {
				break
			}
			zoneVar, err := rv.getZoneVariable(fnDef, value)
			if err != nil {
				logrus.Errorf("getting zone reference id %s: %s", value, err)
				break
			}
			if zoneVar != nil {
				res = append(res, zoneVar)
			}
			hasVars = true
		case *evaluationRef:
			if indexVars, ok := rv.checkZoneVariables(fnDef, block.indexes); ok {
				res = append(res, indexVars...)
			}
		case *evaluationKeyword:
			if zoneVars, ok := rv.checkZoneVariables(fnDef, block.blocks); ok {
				res = append(res, zoneVars...)
			}
		}
	}
	return res, hasVars
}

// getZoneVariable returns the variable definition.
func (rv *runtimeVisitor) getZoneVariable(fnDef *FunctionDefinition, blockStr string) (evaluationZoneTarget, error) {
	var ok bool
	var index string
	name, key := rv.getDataFromVariableName(blockStr)
	if strings.HasPrefix(name, "$") {
		index, ok = name, true
		name = strings.TrimLeft(name, "$")
	} else {
		index, ok = fnDef.AliasKey(name)
	}

	if !ok { // Not found on the function. Must be a global!
		return rv.getZoneGlobalVariable(name, key)
	} // Found on the function the local index for the name.
	return rv.getZoneFunctionVariable(fnDef, name, key, index)
}

// getZoneGlobalVariable returns the global variable definition.
func (rv *runtimeVisitor) getZoneGlobalVariable(name, key string) (evaluationZoneTarget, error) {
	// Check if the variable is global.
	index, ok := rv.ctx.AliasKey(name)
	if !ok {
		return nil, fmt.Errorf("could not find the alias key for the name %s", name)
	}
	globalDef, ok := rv.ctx.globals[index]
	if !ok {
		return nil, fmt.Errorf("could not find the global definition for %s", index)
	}
	if !IsVarTypeStrPrimitive(globalDef.Type) {
		// Set global of complex types must be added to start function.
		return nil, nil
	}
	typ, err := newVariableType(globalDef.Type)
	if err != nil {
		return nil, fmt.Errorf("getting global definition type: %w", err)
	}
	return newEvalPrimitiveZoneTarget(name, index, typ, evaluationZoneTypeGlobal), nil
}

// getZoneFunctionVariable returns the variable definition inside a function.
func (rv *runtimeVisitor) getZoneFunctionVariable(fnDef *FunctionDefinition, name, key, index string) (evaluationZoneTarget, error) {
	// Check function parameters.
	if param, ok := fnDef.Params[index]; ok {
		if ok := isVarTypeStrValid(param.Type); !ok {
			return nil, fmt.Errorf("param identifier %s on evaluation has an invalid type (%s)", param.Name, param.Type)
		}
		typ, err := newVariableType(param.Type)
		if err != nil {
			return nil, fmt.Errorf("getting parameter on function definition type: %w", err)
		}
		if !IsVarTypeStrPrimitive(param.Type) {
			return newEvalCompositeZoneTarget(name, key, strconv.Itoa(param.order), typ.Code(), evaluationZoneTypeParam), nil
		}
		return newEvalPrimitiveZoneTarget(name, index, typ, evaluationZoneTypeParam), nil
	}

	// Check function locals.
	if local, ok := fnDef.Locals[index]; ok {
		if ok := isVarTypeStrValid(local.Type); !ok {
			return nil, fmt.Errorf("local identifier %s on evaluation has an invalid type (%s)", local.Name, local.Type)
		}
		typ, err := newVariableType(local.Type)
		if err != nil {
			return nil, fmt.Errorf("getting local on function definition type: %w", err)
		}
		if !IsVarTypeStrPrimitive(local.Type) {
			return newEvalCompositeZoneTarget(name, key, local.initialValue, typ.Code(), evaluationZoneTypeLocal), nil
		}
		return newEvalPrimitiveZoneTarget(name, index, typ, evaluationZoneTypeLocal), nil
	}

	return nil, nil
}

// getDataFromVariableName returns the variable data from a variable name.
func (rv *runtimeVisitor) getDataFromVariableName(input string) (string, string) {
	var key, name string
	if hasKey := variableIndexReg.MatchString(input); !hasKey {
		return input, ""
	}
	var i int
	for i < len(input) && input[i] != '[' {
		i++
	}
	j := len(input) - 1
	for j < len(input) && input[j] != ']' {
		j--
	}
	name = input[:i]
	key = input[i+1 : j]
	return name, key
}

// getDataFromReference returns the variable data from an evaluation reference.
func (rv *runtimeVisitor) getDataFromReference(refBlock Block, aliasMaps ...AliasDefinition) (*variableReferenceData, error) {
	if ref, ok := refBlock.(*evaluationRef); ok {
		var keys []string
		for _, k := range ref.indexes {
			keys = append(keys, strings.Trim(k.(*evaluationIndex).evaluation.String(), "/"))
		}

		name := ref.text.String()[1:]

		var index string
		for _, aliasMap := range aliasMaps {
			if i, ok := aliasMap.AliasKey(name); ok {
				index = i
				break
			}
		}
		if index == "" {
			return nil, fmt.Errorf("not found index for the reference name %s", name)
		}
		return newVariableReferenceData(name, strings.Join(keys, "]["), index), nil
	}
	reference := refBlock.String()
	if strings.HasPrefix(reference, "#") {
		reference = reference[1:]
		name, key := rv.getDataFromVariableName(reference)
		for _, aliasMap := range aliasMaps {
			if index, ok := aliasMap.AliasKey(name); ok {
				return newVariableReferenceData(name, key, index), nil
			}
		}
		return nil, fmt.Errorf("not found index for the variable name %s", name)
	}
	index := reference
	for _, aliasMap := range aliasMaps {
		if name, ok := aliasMap.AliasValue(index); ok {
			return newVariableReferenceData(name, "", index), nil
		}
	}
	return nil, fmt.Errorf("not found name for the variable index %s", index)
}

// visitCallEvaluation visits the evaluation block used on some call instruction.
func (rv *runtimeVisitor) visitCallEvaluation(fnDef *FunctionDefinition, changes *runtimeChanges, eval *evaluation) error {
	parent := eval.parent.(*Instruction)

	// Get argument type
	argTarget, err := rv.findCallArgumentTarget(fnDef, parent, eval)
	if err != nil {
		return fmt.Errorf("finding argument type: %w", err)
	}

	if isVarTypePrimitive(argTarget.typ.Type()) {
		// Handle primitive transformation.
		return rv.handleCallEvaluationPrimitive(changes, parent, eval, argTarget)
	}
	// Handle primitive composite.
	return rv.handleCallEvaluationComposite(changes, parent, eval, argTarget.typ, argTarget.calleeParamIndex)
}

// visitLocalEvaluation visits the evaluation block use on some set.local instruction.
func (rv *runtimeVisitor) visitLocalEvaluation(fnDef *FunctionDefinition, changes *runtimeChanges, eval *evaluation) error {
	parent := eval.parent.(*Instruction)

	// Get local type
	localTarget, err := rv.findLocalType(fnDef, parent)
	if err != nil {
		return fmt.Errorf("finding local type: %w", err)
	}

	if isVarTypePrimitive(localTarget.typ) {
		return rv.handleEvaluationPrimitive(fnDef, changes, parent, eval, localTarget.typ)
	}
	return rv.handleVariableEvaluationComposite(changes, parent, eval, localTarget, true)
}

// visitGlobalEvaluation visits the evaluation block used on some set.global instruction.
func (rv *runtimeVisitor) visitGlobalEvaluation(changes *runtimeChanges, eval *evaluation) error {
	parent := eval.parent.(*Instruction)

	//	Get global type
	globalTarget, err := rv.findGlobalType(parent)
	if err != nil {
		return fmt.Errorf("finding global type: %w", err)
	}

	if isVarTypePrimitive(globalTarget.typ) {
		return rv.handleEvaluationPrimitive(rv.ctx, changes, parent, eval, globalTarget.typ)
	}
	return rv.handleVariableEvaluationComposite(changes, parent, eval, globalTarget, false)
}

// visitReturnEvaluation visits the evaluation block used on some return instruction.
func (rv *runtimeVisitor) visitReturnEvaluation(fnDef *FunctionDefinition, changes *runtimeChanges, eval *evaluation) error {
	parent := eval.parent.(*Instruction)

	if IsVarTypeStrPrimitive(fnDef.Result) {
		return rv.handleEvaluationPrimitive(rv.ctx, changes, parent, eval, varType(fnDef.Result))
	}
	return rv.handleReturnEvaluationComposite(changes, parent, eval, varType(fnDef.Result))
}

// visitReturnEvaluationRef visits the evaluation reference block used on some return instruction.
func (rv *runtimeVisitor) visitReturnEvaluationRef(fnDef *FunctionDefinition, changes *runtimeChanges, ref *evaluationRef) error {
	parent := ref.parent.(*Instruction)

	// Add extra code for the composite evaluation.
	if err := rv.addEvaluationRefCompositeReturnCode(changes, parent, ref, varType(fnDef.Result)); err != nil {
		return fmt.Errorf("adding code for composite return evaluation reference: %w", err)
	}

	// Makes return instruction empty.
	err := parent.removeChild(ref)
	if err != nil {
		return fmt.Errorf("making return instruction empty: %w", err)
	}
	return nil
}

// visitOtherEvaluation visits the evaluation block used on an unknown instruction.
func (rv *runtimeVisitor) visitOtherEvaluation(fnDef *FunctionDefinition, changes *runtimeChanges, eval *evaluation) error {
	parent, ok := eval.parent.(*Instruction)

	if !ok {
		return fmt.Errorf("the evaluation parent must be an instruction")
	}

	//	Get variable type
	varType, err := rv.findOtherType(fnDef, parent, eval)
	if err != nil {
		return fmt.Errorf("finding global type: %w", err)
	}

	return rv.handleEvaluationPrimitive(fnDef, changes, parent, eval, varType.Type())
}

// addEvaluationPrimitiveCode adds the primitive code originated from the evaluation.
func (rv *runtimeVisitor) addEvaluationPrimitiveCode(changes *runtimeChanges, eval *evaluation, localName string, localType varType) error {
	// Mark module context to add glue functions.
	rv.glueFunctions.operations = true
	rv.glueFunctions.err = true

	// Mark function to add a new local (ex. local_i32_1)
	if _, ok := changes.localsToAdd[localName]; !ok {
		changes.localsToAdd[localName] = newAddedEvaluationTarget(newEvaluationTarget(localName, localName, localType))
	}

	// Add argument code for primitive
	newStartCode, newOperationsCount := wgenerator.GetPrimitiveEvalStartCode(eval.CleanString(), localName, string(localType), changes.operationsCount)
	if err := AddCodeToControlFlow(eval, newStartCode, 0); err != nil {
		return fmt.Errorf("adding evaluation start blocks of type primitive to function: %v", err)
	}
	changes.operationsCount = newOperationsCount
	return nil
}

// addEvaluationCompositeCallCode adds the composite code originated from the call evaluation.
func (rv *runtimeVisitor) addEvaluationCompositeCallCode(changes *runtimeChanges, parent *Instruction, eval *evaluation, typ variableType, argIndex int) error {
	var pushPopArgs bool
	// Mark module context to add glue functions.
	if _, ok := rv.glueFunctions.args[parent]; !ok {
		pushPopArgs = true
		rv.glueFunctions.args[parent] = struct{}{}
	}
	rv.glueFunctions.operations = true

	// Add composite argument code for the call.
	newStartCode, newOperationsCount := wgenerator.GetCompositeCallEvalStartCode(eval.CleanString(), argIndex, pushPopArgs, changes.operationsCount)
	if err := AddCodeToControlFlow(eval, newStartCode, 0); err != nil {
		return fmt.Errorf("adding evaluation start blocks of type %s to function: %v", typ.Type(), err)
	}
	addEndCode := func() (int, error) {
		newEndCode, newOperationsCount := wgenerator.GetCompositeCallEvalEndCode(pushPopArgs, newOperationsCount)
		if err := AddCodeToControlFlow(eval, newEndCode, 1); err != nil {
			return newOperationsCount, fmt.Errorf("adding evaluation end blocks of type %s to function: %v", typ.Type(), err)
		}
		return newOperationsCount, nil
	}
	if !pushPopArgs {
		changes.operationsCount = newOperationsCount
		return nil
	}
	returnParent, returnInstr, returnInstrIndex := findParentReturn(eval)
	if returnInstrIndex == -1 {
		newOperationsCount, err := addEndCode()
		if err != nil {
			return err
		}
		changes.operationsCount = newOperationsCount
		return nil
	}
	fnDef := changes.fnDef
	localName := rv.ctx.generateLocalIndexStr(fnDef.Result, "result") // result is an arbitrary name for the local that saves the result
	// Mark function to add a new local (ex. local_i32_1)
	if _, ok := changes.localsToAdd[localName]; !ok {
		changes.localsToAdd[localName] = newAddedEvaluationTarget(newEvaluationTarget(localName, localName, varType(fnDef.Result)))
	}
	returnInstr.name = instructionCodeSetLocal
	returnInstr.values = append([]Block{newText(localName)}, returnInstr.values...)
	returnParent.addChildrenByIndex(returnInstrIndex+1, NewCodeParser(fmt.Sprintf("(%s (%s %s))", instructionCodeReturn, instructionCodeGetLocal, localName)).parse().blocks...)
	newOperationsCount, err := addEndCode()
	if err != nil {
		return err
	}
	changes.operationsCount = newOperationsCount
	return nil
}

// addEvaluationCompositeReturnCode adds the composite code originated from the return evaluation.
func (rv *runtimeVisitor) addEvaluationCompositeReturnCode(changes *runtimeChanges, parent *Instruction, eval *evaluation, typ varType) error {
	rv.glueFunctions.returns = true

	// Add composite code for the return.
	newCode, newOperationsCount := wgenerator.GetCompositeReturnEvalCode(eval.CleanString(), changes.operationsCount)
	if err := AddCodeToControlFlowIndexFn(eval, newCode, func(values []Block, index int) int {
		if len(values) == 0 || index == 0 {
			return index
		}
		prevInstr, ok := values[index-1].(*Instruction)
		if !ok || prevInstr.String() != instructionCallZonePop {
			return index
		}
		return index - 1
	}); err != nil {
		return fmt.Errorf("adding evaluation blocks of type %s to function: %v", typ, err)
	}
	changes.operationsCount = newOperationsCount
	return nil
}

// addEvaluationRefCompositeReturnCode adds the composite code originated from the return evaluation reference.
func (rv *runtimeVisitor) addEvaluationRefCompositeReturnCode(changes *runtimeChanges, parent *Instruction, ref *evaluationRef, typ varType) error {
	rv.glueFunctions.returns = true

	// Find the reference data.
	refData, err := rv.getDataFromReference(ref, changes.fnDef, rv.ctx)
	if err != nil {
		return err
	}

	// Add composite code for the return.
	newCode, newOperationsCount := wgenerator.GetCompositeReturnEvalRefCode(refData.name, refData.key, changes.operationsCount)
	if err := AddCodeToControlFlow(ref, newCode, 0); err != nil {
		return fmt.Errorf("adding evaluation reference blocks of type %s to function: %v", typ, err)
	}
	changes.operationsCount = newOperationsCount

	return nil
}

// handleCallEvaluationPrimitive handles the call evaluation for a primitive argument.
func (rv *runtimeVisitor) handleCallEvaluationPrimitive(changes *runtimeChanges, parent *Instruction, eval *evaluation, target *evaluationArgumentTarget) error {
	localName := rv.ctx.generateLocalIndex(target.typ.String(), target.calleeParamIndex)

	// Add extra code for the primitive evaluation.
	if err := rv.addEvaluationPrimitiveCode(changes, eval, localName, target.typ.Type()); err != nil {
		return fmt.Errorf("adding code for primitive evaluation: %w", err)
	}

	// Replace the argument with a local specific for the operation (ex. local_i32_1)
	getLocalCode := wgenerator.GetVariableToCode(localName, true)
	getLocalEl := NewCodeParser(getLocalCode).parse()
	if err := parent.replaceChildByIndex(target.callArgIndex, getLocalEl.blocks); err != nil {
		return fmt.Errorf("replacing evaluation with the local value: %w", err)
	}

	return nil
}

// handleCallEvaluationComposite handles the call evaluation for a composite argument.
func (rv *runtimeVisitor) handleCallEvaluationComposite(changes *runtimeChanges, parent *Instruction, eval *evaluation, typ variableType, paramIndex int) error {
	// Add extra code for the composite evaluation.
	if err := rv.addEvaluationCompositeCallCode(changes, parent, eval, typ, paramIndex); err != nil {
		return fmt.Errorf("adding code for composite evaluation: %w", err)
	}

	// The evaluation is registered to be removed
	rv.ctx.evalsToRemove = append(rv.ctx.evalsToRemove, eval)
	return nil
}

// handleEvaluationPrimitive handles the evaluation for primitive variables.
func (rv *runtimeVisitor) handleEvaluationPrimitive(aliasMap AliasDefinition, changes *runtimeChanges, parent *Instruction, eval *evaluation, primitiveType varType) error {
	localName := rv.ctx.generateLocalIndex(string(primitiveType), 0)

	// Add extra code for the primitive evaluation.
	if err := rv.addEvaluationPrimitiveCode(changes, eval, localName, primitiveType); err != nil {
		return fmt.Errorf("adding code for string evaluation: %w", err)
	}

	// Replace the set.local instruction with the code for primitive
	code := wgenerator.GetVariableToCode(localName, true)
	if err := parent.replaceChildWithCode(eval, code); err != nil {
		return fmt.Errorf("replacing evaluation code for the primitive local: %v", err)
	}

	// Replace reference if necessary
	reference := parent.values[0].String()
	if !strings.HasPrefix(reference, "#") {
		return nil
	}
	reference = reference[1:]
	name, _ := rv.getDataFromVariableName(reference)
	index, ok := aliasMap.AliasKey(name)
	if !ok {
		return fmt.Errorf("not found index for the reference name %s", name)
	}
	if err := parent.replaceChildWithCode(parent.values[0], index); err != nil {
		return fmt.Errorf("replacing evaluation reference for the index value: %v", err)
	}
	return nil
}

// handleVariableEvaluationComposite handles the evaluation for composite variables.
func (rv *runtimeVisitor) handleVariableEvaluationComposite(changes *runtimeChanges, parent *Instruction, eval *evaluation, target *evaluationTarget, isLocal bool) error {
	// Mark module context to add glue functions.
	rv.glueFunctions.operations = true

	// Add zone code for composite.
	newStartCode, newOperationsCount := wgenerator.GetCompositeZoneEvalStartCode(target.alias, target.key, eval.CleanString(), isLocal, changes.operationsCount)
	if err := AddCodeToControlFlow(eval, newStartCode, 0); err != nil {
		return fmt.Errorf("adding evaluation start blocks of type composite to function: %v", err)
	}
	changes.operationsCount = newOperationsCount

	// Get the parent of the full instruction.
	instructionParent := parent.getParent()
	if instructionParent == nil {
		return fmt.Errorf("instruction %s (string) has no parent block", parent.name)
	}
	instructionParentInstr, ok := instructionParent.(*Instruction)
	if !ok {
		return fmt.Errorf("parent of instruction %s (string) is not an instruction block", parent.name)
	}

	// Remove the instruction with the string instructions
	if err := instructionParentInstr.removeChild(parent); err != nil {
		return fmt.Errorf("removing set string local code: %v", err)
	}

	return nil
}

// handleReturnEvaluationComposite handles the evaluation for composite returns.
func (rv *runtimeVisitor) handleReturnEvaluationComposite(changes *runtimeChanges, parent *Instruction, eval *evaluation, typ varType) error {
	// Add extra code for the composite evaluation.
	if err := rv.addEvaluationCompositeReturnCode(changes, parent, eval, typ); err != nil {
		return fmt.Errorf("adding code for composite return evaluation: %w", err)
	}

	// Replace instruction for the empty return form.
	err := parent.replaceChildWithCode(eval, fmt.Sprintf("(%s)", instructionCodeReturn))
	if err != nil {
		return fmt.Errorf("replacing return composite with an empty return element: %w", err)
	}
	return nil
}

// findFuncDefinition finds the function definition.
func (rv *runtimeVisitor) findFuncDefinition(findFuncInstr func() *Instruction) (*FunctionDefinition, error) {
	funcInstr := findFuncInstr()
	if funcInstr == nil || len(funcInstr.values) == 0 {
		return nil, errors.New("block is outside function code")
	}
	fnDef, ok := rv.ctx.functions[funcInstr.values[0].String()]
	if !ok {
		return nil, errors.New("function definition not found for evaluation block")
	}
	return fnDef, nil
}

// findOperationType finds the operation type.
func (rv *runtimeVisitor) findOperationType(block Block) (evaluationType, error) {
	parent := block.getParent()
	if parent == nil {
		return "", errors.New("evaluation block does not have a parent")
	}
	parentInstr, ok := parent.(*Instruction)
	if !ok {
		return "", errors.New("evaluation block parent is not an instruction block")
	}
	switch name := parentInstr.name; name {
	case instructionCodeCall:
		return evaluationTypeCall, nil
	case instructionCodeSetLocal, instructionCodeTeeLocal:
		return evaluationTypeLocal, nil
	case instructionCodeSetGlobal:
		return evaluationTypeGlobal, nil
	case instructionCodeReturn:
		return evaluationTypeReturn, nil
	default:
		return "", nil
	}
}

// findCallArgumentTarget finds the call argument target.
func (rv *runtimeVisitor) findCallArgumentTarget(fnDef *FunctionDefinition, call *Instruction, eval *evaluation) (*evaluationArgumentTarget, error) {
	if len(call.values) == 0 {
		return nil, errors.New("invalid call instruction: expected call to have more than one argument")
	}
	calleeName := call.values[0].String()
	callee, ok := rv.ctx.functions[calleeName]
	if !ok {
		return nil, fmt.Errorf("function called does not exist (%s)", calleeName)
	}
	if len(callee.Params) != len(call.values)-1 {
		return nil, fmt.Errorf("arguments passed on function call do not match the function parameters (%s): expected %d but got %d", rv.ctx.FunctionAlias[calleeName], len(callee.Params), len(call.values)-1)
	}
	for i, block := range call.values {
		if block == eval {
			paramType := callee.Parameters()[i-1].Type
			// Validate argument type
			if ok := isVarTypeStrValid(paramType); !ok {
				return nil, fmt.Errorf("parameter type %s is not valid", paramType)
			}
			typ, err := newVariableType(paramType)
			if err != nil {
				return nil, fmt.Errorf("getting parameter type: %w", err)
			}
			return newEvaluationArgumentTarget(fnDef, typ, i, i-1), nil
		}
	}
	return nil, errors.New("evaluation block not found on call instruction")
}

// findLocalType finds the local type.
func (rv *runtimeVisitor) findLocalType(fnDef *FunctionDefinition, instr *Instruction) (*evaluationTarget, error) {
	if len(instr.values) != 2 {
		return nil, errors.New("invalid local instruction: expected instruction to have two arguments")
	}
	ref, err := rv.getDataFromReference(instr.values[0], fnDef)
	if err != nil {
		return nil, err
	}
	if param, ok := fnDef.Params[ref.index]; ok {
		// Validate argument type
		typ, err := newVariableType(param.Type)
		if err != nil {
			return nil, fmt.Errorf("getting parameter type: %w", err)
		}
		return newEvaluationTargetMember(ref.name, ref.key, ref.index, typ.Type()), nil
	}
	if local, ok := fnDef.Locals[ref.index]; ok {
		// Validate local type
		typ, err := newVariableType(local.Type)
		if err != nil {
			return nil, fmt.Errorf("getting local type: %w", err)
		}
		return newEvaluationTargetMember(ref.name, ref.key, ref.index, typ.Type()), nil
	}
	return nil, fmt.Errorf("local named %s not found", ref.name)
}

// findGlobalType finds the global type.
func (rv *runtimeVisitor) findGlobalType(instr *Instruction) (*evaluationTarget, error) {
	if len(instr.values) != 2 {
		return nil, errors.New("invalid global instruction: expected instruction to have two arguments")
	}
	ref, err := rv.getDataFromReference(instr.values[0], rv.ctx)
	if err != nil {
		return nil, err
	}
	if global, ok := rv.ctx.globals[ref.index]; ok {
		// Validate argument type
		typ, err := newVariableType(global.Type)
		if err != nil {
			return nil, fmt.Errorf("getting global type: %w", err)
		}
		return newEvaluationTargetMember(ref.name, ref.key, ref.index, typ.Type()), nil
	}
	return nil, fmt.Errorf("global named %s not found", ref.name)
}

// findOtherType finds another unknown instruction type.
func (rv *runtimeVisitor) findOtherType(fnDef *FunctionDefinition, parentInstr *Instruction, eval *evaluation) (variableType, error) {
	if parentInstr.name == instructionFunction {
		if fnDef.Result == "" {
			return nil, fmt.Errorf("could not get the instruction type %s because the function has no result type", parentInstr.name)
		}
		if !IsVarTypeStrPrimitive(fnDef.Result) {
			return nil, fmt.Errorf("invalid result type %s on function", fnDef.Result)
		}
		return newVariableType(fnDef.Result)
	}
	parentDef, ok := wlang.GetInstrDefinition(parentInstr.name)
	if !ok {
		return nil, fmt.Errorf("could not find the definition for the parent instruction %s of the evaluation", parentInstr.name)
	}
	evalIndex := parentInstr.childIndex(eval)
	if evalIndex == -1 {
		return nil, errors.New("could not find the evaluation index on parent values")
	}
	if evalIndex >= len(parentDef.Args) {
		return nil, errors.New("evaluation argument out of range")
	}
	arg := parentDef.Args[evalIndex]
	if !IsVarTypeStrPrimitive(string(arg)) {
		return nil, fmt.Errorf("invalid type %s on instruction %s", arg, parentInstr.name)
	}
	return newVariableType(string(arg))
}

// initVisitInputData contains the data for initiating a runtime block visit.
type initVisitInputData struct {
	rv    *runtimeVisitor
	block Block
}

// newInitVisitInputData is a constructor for newInitVisitInputData.
func newInitVisitInputData(rv *runtimeVisitor, block Block) *initVisitInputData {
	return &initVisitInputData{
		rv:    rv,
		block: block,
	}
}

// initVisitOutputData contains the data resultant from the initiation of a runtime block visit.
type initVisitOutputData struct {
	fnDef         *FunctionDefinition
	fnInstr       *Instruction
	fnChildIndex  int
	operationType evaluationType
	changes       *runtimeChanges
}

// newInitVisitOutputData is a constructor for initVisitOutputData.
func newInitVisitOutputData(fnDef *FunctionDefinition, fnInstr *Instruction, fnChildIndex int, operationType evaluationType, changes *runtimeChanges) *initVisitOutputData {
	return &initVisitOutputData{
		fnDef:         fnDef,
		fnInstr:       fnInstr,
		fnChildIndex:  fnChildIndex,
		operationType: operationType,
		changes:       changes,
	}
}

// initVisit is responsible for initiating a runtime block visit.
func initVisit(in *initVisitInputData) (*initVisitOutputData, error) {
	// Register on visitor the new runtime change (if needed)
	fnDef, err := in.rv.findFuncDefinition(in.block.funcInstr)
	if err != nil {
		return nil, fmt.Errorf("finding function definition for evaluation %s: %v", in.block.String(), err)
	}
	changes, ok := in.rv.functionChanges[fnDef.Name]
	if !ok {
		changes = newRuntimeChanges(fnDef)
		in.rv.functionChanges[fnDef.Name] = changes
	}

	// Get operation type
	operationType, err := in.rv.findOperationType(in.block)
	if err != nil {
		return nil, fmt.Errorf("finding operation type for evaluation %s: %v", in.block.String(), err)
	}

	// Get the evaluation index on function.
	fnInstr, evalIndex, err := findInstructionIndexOnFunction(in.block)
	if err != nil {
		// The value might already been deleted from the module.
		return nil, nil
	}
	return newInitVisitOutputData(fnDef, fnInstr, evalIndex, operationType, changes), nil
}

// compositePropsVisitor is the visitor struct used to replace composite properties on the module code.
type compositePropsVisitor struct {
	*visitorAdapter
}

// newCompositePropsVisitor is a constructor for compositePropsVisitor.
func newCompositePropsVisitor() *compositePropsVisitor {
	return &compositePropsVisitor{new(visitorAdapter)}
}

// VisitInstruction handles some instructions block.
func (cv *compositePropsVisitor) VisitInstruction(instr *Instruction) bool {
	switch instr.name {
	case instructionParam:
		return cv.visitParamInstruction(instr)
	case instructionLocal, instructionResult:
		return cv.visitLocalOrResultInstruction(instr)
	default:
		return false
	}
}

// visitParamInstruction visits some parameter instruction block.
func (cv *compositePropsVisitor) visitParamInstruction(instr *Instruction) bool {
	var changed bool
	for _, value := range instr.values {
		valueStr := value.String()
		if strings.HasPrefix(valueStr, "$") {
			continue
		}
		if !IsVarTypeStrPrimitive(valueStr) {
			instr.values = deleteBlock(instr.values, value)
			changed = true
		}
	}
	if len(instr.values) == 0 || len(instr.values) == 1 && strings.HasPrefix(instr.values[0].String(), "$") {
		parentBlock := instr.getParent()
		if parentBlock == nil {
			return false
		}
		switch parent := parentBlock.(type) {
		case *Instruction:
			parent.values = deleteBlock(parent.values, instr)
		case *element:
			parent.blocks = deleteBlock(parent.blocks, instr)
		default:
			// Empty by design.
			return changed
		}
	}
	return changed
}

// visitLocalInstruction visits some local instruction block.
func (cv *compositePropsVisitor) visitLocalOrResultInstruction(instr *Instruction) bool {
	if len(instr.values) != 0 && IsVarTypeStrPrimitive(instr.values[len(instr.values)-1].String()) {
		return false
	}
	parentBlock := instr.getParent()
	if parentBlock == nil {
		return false
	}
	switch parent := parentBlock.(type) {
	case *Instruction:
		parent.values = deleteBlock(parent.values, instr)
	case *element:
		parent.blocks = deleteBlock(parent.blocks, instr)
	default:
		// Empty by design.
		return false
	}
	return true
}

// globalCompositeVisitor is a visitor used to initialize the global composite variables on the javascript scope.
type globalCompositeVisitor struct {
	*visitorAdapter
	ctx             *ModuleContext
	startFn         *FunctionDefinition
	operationsCount int
	foundCount      int
}

// newGlobalRuntimeVisitor is a constructor for globalCompositeVisitor.
func newGlobalRuntimeVisitor(ctx *ModuleContext, startFn *FunctionDefinition) *globalCompositeVisitor {
	return &globalCompositeVisitor{
		visitorAdapter: new(visitorAdapter),
		ctx:            ctx,
		startFn:        startFn,
	}
}

// VisitInstruction handles some instruction block.
func (gv *globalCompositeVisitor) VisitInstruction(instr *Instruction) bool {
	if instr.name != instructionGlobal || len(instr.values) == 0 {
		return false
	}

	// Find global definition.
	globalName := instr.values[0].String()
	globalDef, ok := gv.ctx.globals[globalName]
	if !ok || IsVarTypeStrPrimitive(globalDef.Type) {
		return false
	}

	// Find instruction parent.
	parentBlock := instr.getParent()
	if parentBlock == nil {
		logrus.Fatalf("global instruction %s has no parent block", globalName)
	}

	// Remove the instruction from the parent children.
	switch parent := parentBlock.(type) {
	case *Instruction:
		if err := parent.removeChild(instr); err != nil {
			logrus.Fatalf("removing global instruction %s: %v", globalName, err)
		}
	case *element:
		parent.blocks = deleteBlock(parent.blocks, instr)
	default:
		logrus.Fatalf("global instruction %s has an invalid parent block type", globalName)
	}

	// Parse set global instruction.
	variableType, err := newVariableType(globalDef.Type)
	if err != nil {
		logrus.Fatalf("global instruction %s has an invalid type %s", globalName, globalDef.Type)
	}
	code, newOpCount := wgenerator.GetSetStartingGlobalCompositeCode(
		gv.ctx.GlobalAlias[globalDef.Name],
		globalDef.initialValue,
		variableType.Code(),
		gv.operationsCount,
	)
	setGlobalEl := NewCodeParser(code).parse()

	// Add code blocks to start function.
	gv.startFn.addInstrsAtStart(setGlobalEl.blocks)
	gv.operationsCount = newOpCount

	// Sets that the start function needs to have a new zone inition.
	if _, ok := gv.ctx.runtimeChanges[gv.startFn.Name]; !ok {
		gv.ctx.runtimeChanges[gv.startFn.Name] = newRuntimeChanges(gv.startFn)
		gv.ctx.runtimeChanges[gv.startFn.Name].hasNewZone = true
	}

	// Increments found count.
	gv.foundCount++
	return true
}

// compositeReturnFuncVisitor is a visitor used to find the functions with returns of composite type.
type compositeReturnFuncVisitor struct {
	*visitorAdapter
	ctx    *ModuleContext
	fnDefs map[string]struct{}
}

// newCompositeReturnFuncVisitor is a constructor for compositeReturnFuncVisitor.
func newCompositeReturnFuncVisitor(ctx *ModuleContext) *compositeReturnFuncVisitor {
	return &compositeReturnFuncVisitor{
		visitorAdapter: new(visitorAdapter),
		ctx:            ctx,
		fnDefs:         make(map[string]struct{}),
	}
}

// VisitInstruction handles some instructions block.
func (cr *compositeReturnFuncVisitor) VisitInstruction(instr *Instruction) bool {
	if instr.name != instructionFunction || // Type function
		len(instr.values) < 2 { // Must contain at least the name and an instruction.
		return false
	}

	// Find function definition.
	fnName := instr.values[0].String()
	fnDef, ok := cr.ctx.functions[fnName]
	if !ok {
		// It is not a function declaration.
		return false
	}

	// Check if result type is primitive.
	if fnDef.Result == "" || IsVarTypeStrPrimitive(fnDef.Result) {
		return false
	} // If not is composite.

	// Mark function as complex return.
	cr.fnDefs[fnDef.Name] = struct{}{}

	// If it is not internal, it don't need to suffer changes.
	if !isInternalInstruction(instr) {
		return false
	}

	var returnVal string

	// Find if the last return value is valid.
	switch lastReturn := instr.values[len(instr.values)-1].(type) {
	case *evaluationRef, *evaluation:
		returnVal = lastReturn.String()
	default:
		// The last value must be an evaluation or a runtime reference.
		return false
	}

	// If is valid, must be inserted into a return instruction.
	returnCode := fmt.Sprintf("(%s %s)", instructionCodeReturn, returnVal)
	returnEl := NewCodeParser(returnCode).parse()
	err := instr.replaceChildByIndex(len(instr.values)-1, returnEl.blocks)
	if err != nil {
		logrus.Errorf("replacing last composite instructions to be inside a return element: %v", err)
		return false
	}

	return true
}

// callCompositeReturnFuncVisitor is the visitor for the call instructions that call functions with composite returns.
// changes the instructions on the module while it executes.
type callCompositeReturnFuncVisitor struct {
	*visitorAdapter
	fns map[string]struct{}
}

// newCallCompositeReturnFuncVisitor is a constructor for callCompositeReturnFuncVisitor.
func newCallCompositeReturnFuncVisitor(fns map[string]struct{}) *callCompositeReturnFuncVisitor {
	return &callCompositeReturnFuncVisitor{
		visitorAdapter: new(visitorAdapter),
		fns:            fns,
	}
}

// VisitInstruction handles some instructions block.
func (crf *callCompositeReturnFuncVisitor) VisitInstruction(instr *Instruction) bool {
	if instr.name != instructionCodeCall || len(instr.values) == 0 {
		return false
	}
	if _, ok := crf.fns[instr.values[0].String()]; !ok {
		return false
	}
	parent, child := instr.getParent(), Block(instr)
	for {
		if parent == nil {
			return false
		}
		parentInstr, ok := parent.(*Instruction)
		if !ok {
			return false
		}
		if wlang.IsControlFlow(parentInstr.name) || parentInstr.name == instructionFunction {
			break
		}
		child = parent
		parent = parent.getParent()
	}
	callingParent := instr.getParent().(*Instruction)
	if err := callingParent.replaceChildWithCode(instr, fmt.Sprintf("/%s/", returnKeyword)); err != nil {
		logrus.Errorf("could not replace composite call for the adjacent return evaluation: %v", err)
		return false
	}
	parentInstr, childInstr := parent.(*Instruction), child.(*Instruction)
	if err := parentInstr.addChildren(childInstr, instr); err != nil {
		logrus.Errorf("could not move composite call to the outside instruction: %v", err)
		return false
	}
	return true
}

// importCompositeFuncVisitor is a visitor responsible to visit the import functions
// 	that have some kind of composite value in their declaration,
// 	and change it to the operations import module.
type importCompositeFuncVisitor struct {
	*visitorAdapter
	ctx *ModuleContext
}

// newImportCompositeFuncVisitor is a constructor for importCompositeFuncVisitor.
func newImportCompositeFuncVisitor(ctx *ModuleContext) *importCompositeFuncVisitor {
	return &importCompositeFuncVisitor{
		visitorAdapter: new(visitorAdapter),
		ctx:            ctx,
	}
}

// VisitInstruction handles some instructions block.
func (iv *importCompositeFuncVisitor) VisitInstruction(instr *Instruction) bool {
	if instr.name != instructionFunction || !isImportInstruction(instr) || len(instr.values) == 0 {
		return false
	}

	// Find function definition.
	fnName := instr.values[0].String()
	fnDef, ok := iv.ctx.functions[fnName]
	if !ok {
		logrus.Errorf("finding import composite functions: function %s do not exist on context module", fnName)
		return false
	}

	var isComposite bool
	for _, p := range fnDef.Parameters() {
		if !IsVarTypeStrPrimitive(p.Type) {
			isComposite = true
			break
		}
	}
	isComposite = isComposite || fnDef.Result != "" && !IsVarTypeStrPrimitive(fnDef.Result)
	if !isComposite {
		return false
	}

	// Replace module name for operation
	err := instr.parent.(*Instruction).replaceChildByIndex(0, []Block{newText(`"operations"`)}) // Module name
	if err != nil {
		logrus.Errorf("replacing module name for composite import function: %v", err)
		return false
	}
	return true
}

// elemTableFixVisitor is a visitor responsible to fix the table element index bug
// 	originated from the WABT tool when using the flag "--generate-names".
type elemTableFixVisitor struct {
	*visitorAdapter
}

// newElemTableFixVisitor is a constructor for elemTableFixVisitor.
func newElemTableFixVisitor() *elemTableFixVisitor {
	return &elemTableFixVisitor{
		visitorAdapter: new(visitorAdapter),
	}
}

// VisitInstruction handles some instructions block.
func (ev *elemTableFixVisitor) VisitInstruction(instr *Instruction) bool {
	if instr.name != instructionElem || !isInternalInstruction(instr) {
		return false
	}
	err := instr.replaceChildByIndex(0, []Block{newText("0")}) // Only one table must exist on module.
	if err != nil {
		logrus.Errorf("changing elem element index: replacing child by index: %v", err)
		return false
	}
	return true
}

// returnFuncVisitor is the visitor responsible to record all the return instructions.
type returnFuncVisitor struct {
	*visitorAdapter
	instrs []Block
}

// newReturnFuncVisitor is a constructor for returnFuncVisitor.
func newReturnFuncVisitor() *returnFuncVisitor {
	return &returnFuncVisitor{
		visitorAdapter: new(visitorAdapter),
	}
}

// VisitInstruction handles some instructions block.
func (rv *returnFuncVisitor) VisitInstruction(instr *Instruction) bool {
	if instr.name != instructionCodeReturn {
		return false
	}
	rv.instrs = append(rv.instrs, instr)
	return true
}
