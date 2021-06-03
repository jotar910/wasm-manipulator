package wpointcut

import (
	"fmt"
	"regexp"

	"github.com/shivamMg/ppds/tree"
	"github.com/sirupsen/logrus"

	"joao/wasm-manipulator/internal/wcode"
	"joao/wasm-manipulator/internal/wparser/pointcut"
	"joao/wasm-manipulator/internal/wyaml"
)

// Node is implemented by the elements present in the pointcut expression.
type Node interface {
	tree.Node
	Filter(in *PointcutContext) *PointcutContext
}

// newPointcutNode is a constructor for Node.
// the pointcut node is selected accordingly to the block type.
func newPointcutNode(pp *ParsedPointcut, blockType NodeType, block *pointcut.Method, joinPointParams map[string]ParsedParam) Node {
	switch {
	case block.Func != nil:
		return newFuncNode(block.Func.Name, blockType, block.Func.Input)
	case block.Call != nil:
		return newCallNode(block.Call.Name, blockType, block.Call.Input)
	case block.Args != nil:
		return newArgsNode(block.Args.Name, blockType, block.Args.Input, joinPointParams)
	case block.Returns != nil:
		var returnType *string
		if !block.Returns.Input.Any {
			returnType = &block.Returns.Input.Value
		}
		return newReturnsNode(block.Returns.Name, blockType, returnType)
	case block.Templ != nil:
		return newTemplateNode(block.Templ.Name, blockType, block.Templ.Input.Template, bool(block.Templ.Input.JustCheck))
	case block.Other != nil:
		for name, p := range pp.Transformation.Pointcuts {
			if name == block.Other.Name {
				return newOtherMethodNode(pp, name, blockType, p, block.Other.Arguments)
			}
		}
		logrus.Fatalf("unable to build block node: unknown pointcut %s", block.Other)
	default:
		logrus.Fatalf("unable to build block node: unknown block type")
	}
	return nil
}

// NodeType is the type for the enumeration.
type NodeType int

const (
	NodeTypeOr NodeType = iota
	NodeTypeAnd
	NodeTypeLeaf
)

// ParsedPointcut is responsible for parsing a pointcut.
// generates a pointcut expression.
type ParsedPointcut struct {
	Initiated      bool
	Params         map[string]ParsedParam
	Instrs         []pointcut.Instruction
	Transformation *wyaml.BaseYAML
	context        *PointcutContext
	expr           Node
}

// NewParsedPointcut is a constructor for ParsedPointcut.
// fills the pointcut parameters.
// parses the pointcut into an expression.
func NewParsedPointcut(expr *pointcut.ExprWithContext, transformation *wyaml.BaseYAML) *ParsedPointcut {
	params := make(map[string]ParsedParam)
	fillPointcutParams(params, expr.Args)
	return &ParsedPointcut{Params: params, Instrs: expr.Instrs, Transformation: transformation}
}

// String returns the string format for the pointcut expression.
func (pp ParsedPointcut) String() string {
	return fmt.Sprintf("{params: %+v, instructions: %+v}", pp.Params, pp.Instrs)
}

// Init initiates the parsed pointcut, parsing the pointcut expression.
func (pp *ParsedPointcut) Init(context *wcode.ModuleContext, transformation *wyaml.BaseYAML) *ParsedPointcut {
	defer func() {
		pp.Initiated = true
	}()
	// Prepare initial context to be executed
	pp.context = newPointcutContext(context, transformation)
	pp.expr = pp.parseExpression(pp.Instrs, &parseStacks{})
	return pp
}

// Execute executes the pointcut, returning the context data obtained during the process.
func (pp *ParsedPointcut) Execute() *PointcutContext {
	return pp.expr.Filter(pp.context)
}

// ParseExpression parses the pointcut expression.
// returns the root node of the resultant pointcut tree.
func (pp *ParsedPointcut) ParseExpression() Node {
	return pp.parseExpression(pp.Instrs, &parseStacks{})
}

// ParseExpression parses the pointcut expression.
func (pp *ParsedPointcut) parseExpression(expr []pointcut.Instruction, stacks *parseStacks) Node {
	for _, e := range expr {
		if len(e.GroupBlock) > 0 {
			newStacks := parseStacks{}
			stacks.blockStack = append(stacks.blockStack, pp.parseExpression(e.GroupBlock, &newStacks))
		}
		if e.Block != nil {
			blockNode := newPointcutNode(pp, NodeTypeLeaf, e.Block, pp.Params)
			stacks.blockStack = append(stacks.blockStack, blockNode)
		}
		if e.AndBlock != nil {
			handleParseOpNode(NodeTypeAnd, stacks, true)
		}
		if e.OrBlock != nil {
			handleParseOpNode(NodeTypeOr, stacks, true)
		}
	}
	for i := len(stacks.opStack) - 1; i > -1; i-- {
		doOperation(stacks)
	}
	if len(stacks.blockStack) != 1 {
		logrus.Fatalf("group ended with %d nodes left on blocks stack... expected 1", len(stacks.blockStack))
	}
	if len(stacks.opStack) != 0 {
		logrus.Fatalf("group ended with %d nodes left on operations stack... expected 0", len(stacks.blockStack))
	}
	return popBlock(&stacks.blockStack)
}

// ParsedParam contains the data for a pointcut parameter already parsed.
type ParsedParam struct {
	Name     string
	Type     string
	Variable string // local or param
	Index    string
}

// NodeInstance is the base model for any node instance.
type NodeInstance struct {
	name     string
	nodeType NodeType
	left     Node
	right    Node
}

// newEmptyNodeInstance is a constructor for NodeInstance with no children.
func newEmptyNodeInstance(name string, nodeType NodeType) *NodeInstance {
	return &NodeInstance{name: name, nodeType: nodeType}
}

// newNodeInstance is a constructor for NodeInstance.
func newNodeInstance(name string, nodeType NodeType, left, right Node) *NodeInstance {
	return &NodeInstance{name: name, nodeType: nodeType, left: left, right: right}
}

// String returns the string format for the node instance.
func (n NodeInstance) String() string {
	return fmt.Sprintf("{Type: %q, Name: %+v}", n.nodeType, n.name)
}

// Data returns the node itself.
func (n NodeInstance) Data() interface{} {
	return n
}

// Children returns the node children.
func (n NodeInstance) Children() []tree.Node {
	var res []tree.Node
	if n.left != nil {
		res = append(res, n.left)
	}
	if n.right != nil {
		res = append(res, n.right)
	}
	return res
}

// OperationNode is a node of type operation.
type OperationNode struct {
	*NodeInstance
}

// newOperationNode is a constructor for OperationNode.
func newOperationNode(nodeType NodeType, left, right Node) *OperationNode {
	var name string
	switch nodeType {
	case NodeTypeOr:
		name = "or"
	case NodeTypeAnd:
		name = "and"
	default:
		name = "unknown"
	}
	return &OperationNode{newNodeInstance(name, nodeType, left, right)}
}

// Filter filters the pointcut context accordingly to the current node.
func (p *OperationNode) Filter(ctx *PointcutContext) *PointcutContext {
	switch p.nodeType {
	case NodeTypeAnd:
		return p.executeAndOperation(ctx)
	case NodeTypeOr:
		return p.executeOrOperation(ctx)
	default:
		logrus.Fatalf("unknown pointcut operation node with type %v", p.nodeType)
		return ctx
	}
}

// executeAndOperation executes the logical operation AND return the resultant context.
func (p *OperationNode) executeAndOperation(ctx *PointcutContext) *PointcutContext {
	return newJoinPointMapMerge(ctx, p.left.Filter, p.right.Filter).and()
}

// executeAndOperation executes the logical operation OR return the resultant context.
func (p *OperationNode) executeOrOperation(ctx *PointcutContext) *PointcutContext {
	return newJoinPointMapMerge(ctx, p.left.Filter, p.right.Filter).or()
}

// parseStacks contains all the stacks (one of each type) used for parsing the pointcut expression.
type parseStacks struct {
	blockStack []Node
	opStack    []NodeType
}

// handleParseOpNode handles the operation node.
func handleParseOpNode(opType NodeType, stacks *parseStacks, add bool) {
	for i := len(stacks.opStack) - 1; i > -1 && stacks.opStack[i] >= opType; i-- {
		doOperation(stacks)
	}
	if add {
		stacks.opStack = append(stacks.opStack, opType)
	}
}

// doOperation runs some operation operation, manipulating the stacks.
func doOperation(stacks *parseStacks) {
	op := popOp(&stacks.opStack)
	valueB := popBlock(&stacks.blockStack)
	valueA := popBlock(&stacks.blockStack)
	stacks.blockStack = append(stacks.blockStack, newOperationNode(op, valueA, valueB))
}

// fillPointcutParams fills the pointcut parameters from the arguments to the pointcut.
func fillPointcutParams(params map[string]ParsedParam, args []pointcut.ArgumentWithContext) {
	for _, arg := range args {
		params[arg.Name] = ParsedParam{
			Name:     arg.Name,
			Type:     arg.Type,
			Variable: arg.VarType,
			Index:    string(arg.VarIndex),
		}
	}
}

// resolvePointcutFunctionName resolves the function name on a pointcut.
func resolvePointcutFunctionName(inName *pointcut.FuncDefinitionName) *functionPointcutPropsName {
	if inName == nil || inName.Any {
		return nil
	}
	res := &functionPointcutPropsName{}
	switch n := inName; {
	case n.Name != "":
		res.Name = &n.Name
	case n.IndexName != "":
		res.IndexName = &n.IndexName
	case n.Index != nil:
		res.Index = n.Index
	case n.Regex != "":
		reg := n.Regex.String()
		_, err := regexp.Compile(reg)
		if err != nil {
			logrus.Fatalf("invalid regex %q", n.Regex)
		}
		res.Regex = &reg
	case n.Variable != "":
		res.Variable = &n.Variable
	default:
		switch {
		case n.VariableName != nil:
			res.Variable = &n.VariableName.Variable
			if n.VariableName.Name != nil {
				res.Name = n.VariableName.Name
			}
		case n.VariableIndexName != nil:
			res.Variable = &n.VariableIndexName.Variable
			if n.VariableIndexName.IndexName != nil {
				res.IndexName = n.VariableIndexName.IndexName
			}
		case n.VariableIndex != nil:
			res.Variable = &n.VariableIndex.Variable
			if n.VariableIndex.Index != nil {
				res.Index = n.VariableIndex.Index
			}
		case n.VariableRegex != nil:
			res.Variable = &n.VariableRegex.Variable
			if n.VariableRegex.Regex != nil {
				reg := n.VariableRegex.Regex.String()
				res.Regex = &reg
			}
		}
	}
	return res
}

// resolvePointcutFunctionReturn resolves the function return value on a pointcut.
func resolvePointcutFunctionReturn(inReturn *pointcut.FuncDefinitionReturn) *functionPointcutPropsReturn {
	if inReturn == nil || inReturn.Any {
		return nil
	}
	ret := inReturn
	res := &functionPointcutPropsReturn{}
	switch {
	case ret.Type != "":
		res.Type = &ret.Type
	case ret.Variable != "":
		res.Variable = &ret.Variable
	case ret.VariableType != nil:
		res.Variable = &ret.VariableType.Variable
		res.Type = &ret.VariableType.Type
	}
	return res
}

// resolvePointcutFunctionReturn resolves the function parameters values on a pointcut.
func resolvePointcutFunctionParams(inParams *pointcut.FuncDefinitionParams) *[]functionPointcutPropsParams {
	if inParams == nil {
		return new([]functionPointcutPropsParams)
	}
	if inParams.Any {
		return nil
	}
	var res []functionPointcutPropsParams
	for _, param := range inParams.Params {
		p := functionPointcutPropsParams{}
		if param.Type != nil && !param.Type.Any {
			p.Type = &param.Type.Value
		}
		if param.Name != nil && !param.Name.Any {
			p.Variable = &param.Name.Variable
		}
		res = append(res, p)
	}
	return &res
}

// resolvePointcutArgsParams resolves the arguments parameters values on a pointcut.
func resolvePointcutArgsParams(inProps *pointcut.ArgMethodInput, joinPointParams map[string]ParsedParam) *[]ParsedParam {
	var res []ParsedParam

	if inProps != nil {
		for _, inParam := range inProps.Params {
			param, ok := joinPointParams[inParam]
			if !ok {
				logrus.Fatalf("resolving args parameters for %q: parameter not found on inputs", inParam)
			}
			res = append(res, param)
		}
	}

	return &res
}

// popBlock pops a value from the block stack.
func popBlock(s *[]Node) Node {
	stack := *s
	n := len(stack) - 1
	val := stack[n]
	*s = stack[:n]
	return val
}

// popOp pops a value from the operations stack.
func popOp(s *[]NodeType) NodeType {
	stack := *s
	n := len(stack) - 1
	val := stack[n]
	*s = stack[:n]
	return val
}
