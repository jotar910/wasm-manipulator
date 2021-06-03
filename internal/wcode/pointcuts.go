package wcode

import (
	"joao/wasm-manipulator/internal/wparser/lex"
	"strings"
)

// funcVisitor represents the module visitor for the func pointcut.
type funcVisitor struct {
	visitorAdapter
	context *ModuleContext
	visitor PointcutVisitor
	filter  FuncFilterFn
}

// newFuncVisitor is the constructor for funcVisitor.
func newFuncVisitor(context *ModuleContext, visitor PointcutVisitor, filter FuncFilterFn) *funcVisitor {
	return &funcVisitor{context: context, visitor: visitor, filter: filter}
}

// VisitInstruction handles some instructions block.
func (fv *funcVisitor) VisitInstruction(instr *Instruction) bool {
	if !isInternalInstruction(instr) || instr.name != instructionFunction || len(instr.values) == 0 {
		return false
	}
	fnDef, ok := fv.context.functions[instr.values[0].String()]
	if !ok {
		return false
	}
	fnData := newFuncData(fv.context, fnDef)
	if env, ok := fv.filter(fv.context, fnData); ok {
		fv.visitor.VisitFunc(instr, fnData, env)
	}
	return true
}

// callVisitor represents the module visitor for the call pointcut.
type callVisitor struct {
	visitorAdapter
	context *ModuleContext
	visitor PointcutVisitor
	filter  CallFilterFn
}

// newCallVisitor is the constructor for callVisitor.
func newCallVisitor(context *ModuleContext, visitor PointcutVisitor, filter CallFilterFn) *callVisitor {
	return &callVisitor{context: context, visitor: visitor, filter: filter}
}

// VisitInstruction handles some instructions block.
func (cv *callVisitor) VisitInstruction(instr *Instruction) bool {
	if instr.name != instructionCodeCall || len(instr.values) == 0 {
		return false
	}
	index := instr.values[0].String()
	fnDef, ok := cv.context.functions[index]
	if !ok {
		index, ok := cv.context.AliasKey(strings.TrimRight(strings.TrimLeft(index, lex.LeftKeyword), lex.RightKeyword))
		if !ok {
			return false
		}
		fnDef, ok = cv.context.functions[index]
		if !ok {
			return false
		}
	}
	callDef := newCallDefinition(instr, fnDef)
	callData := newCallData(cv.context, callDef)
	if env, ok := cv.filter(cv.context, callData); ok {
		cv.visitor.VisitCall(instr, callData, env)
	}
	return true
}

// argsVisitor represents the module visitor for the args pointcut.
type argsVisitor struct {
	visitorAdapter
	context *ModuleContext
	visitor PointcutVisitor
	filter  ArgsFilterFn
}

// newArgsVisitor is the constructor for argsVisitor.
func newArgsVisitor(context *ModuleContext, visitor PointcutVisitor, filter ArgsFilterFn) *argsVisitor {
	return &argsVisitor{context: context, visitor: visitor, filter: filter}
}

// VisitInstruction handles some instructions block.
func (cav *argsVisitor) VisitInstruction(instr *Instruction) bool {
	if instr.name != instructionCodeCall || len(instr.values) == 0 {
		return false
	}
	fnDef, ok := cav.context.functions[instr.values[0].String()]
	if !ok {
		return false
	}
	for i := 1; i < len(instr.values); i++ {
		argInstr, ok := instr.values[i].(*Instruction)
		// Only accepts calls with explicit reference to variables
		if !ok || len(argInstr.values) != 1 ||
			strings.Index(argInstr.name, instructionGlobal) != 0 && strings.Index(argInstr.name, instructionLocal) != 0 {
			return false
		}
	}
	callDef := newCallDefinition(instr, fnDef)
	argsData := newArgsData(cav.context, callDef)
	if env, ok := cav.filter(cav.context, argsData); ok {
		cav.visitor.VisitArgs(instr, argsData, env)
	}
	return true
}

// returnsVisitor represents the module visitor for the returns pointcut.
type returnsVisitor struct {
	visitorAdapter
	context *ModuleContext
	visitor PointcutVisitor
	filter  ReturnsFilterFn
}

// newReturnsVisitor is the constructor for returnsVisitor.
func newReturnsVisitor(context *ModuleContext, visitor PointcutVisitor, filter ReturnsFilterFn) *returnsVisitor {
	return &returnsVisitor{context: context, visitor: visitor, filter: filter}
}

// VisitInstruction handles some instructions block.
func (rv *returnsVisitor) VisitInstruction(instr *Instruction) bool {
	if instr.name != instructionFunction || !isInternalInstruction(instr) || len(instr.values) == 0 {
		return false
	}

	// Find function definition.
	fnName := instr.values[0].String()
	fnDef, ok := rv.context.functions[fnName]
	if !ok {
		// It is not a function declaration.
		return false
	}

	// Find returns on function.
	visitor := newReturnFuncVisitor()
	instr.Traverse(visitor)

	// Append the last instruction as return.
	returns := visitor.instrs
	if lastReturn := instr.values[len(instr.values)-1]; len(returns) == 0 || lastReturn != returns[len(returns)-1] {
		if fnDef.Result == "" {
			lastReturn = NewCodeParser("(return)").parse().blocks[0]
			fnDef.instr.values = append(fnDef.instr.values, lastReturn)
			lastReturn.setParent(fnDef.instr)
		}
		returns = append(returns, lastReturn)
	}

	// Call visitor for each return instruction.
	for _, ret := range returns {
		returnsData := newReturnsData(rv.context, fnDef, ret)
		if env, ok := rv.filter(rv.context, returnsData); ok {
			rv.visitor.VisitReturns(ret, returnsData, env)
		}
	}
	return true
}
