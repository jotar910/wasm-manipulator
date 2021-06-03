package lex

import (
	"log"
	"strconv"

	"github.com/sirupsen/logrus"
)

type TokenNodeType int

const (
	TokenTypeValue TokenNodeType = iota
	TokenTypeMethod
	TokenTypeUnknown
	TokenTypeNegation

	// Logical ops
	TokenTypeOpOr
	TokenTypeOpAnd

	// Comparators
	TokenTypeOpEqual
	TokenTypeOpNotEqual
	TokenTypeOpGreaterEqual
	TokenTypeOpGreater
	TokenTypeOpLessEqual
	TokenTypeOpLess

	// Modifiers
	TokenTypeOpBitwiseLeft
	TokenTypeOpBitwiseRight
	TokenTypeOpPlus
	TokenTypeOpMinus
	TokenTypeOpMultiplication
	TokenTypeOpDivision
	TokenTypeOpRemainder
)

// precedence returns the base weight for the precedence of a token node type.
// used to allow different token type to have equals weights.
func precedence(t TokenNodeType) TokenNodeType {
	// Check operators with same weight.
	switch t {
	case TokenTypeOpNotEqual:
		return TokenTypeOpEqual
	case TokenTypeOpGreater, TokenTypeOpLessEqual, TokenTypeOpLess:
		return TokenTypeOpGreaterEqual
	case TokenTypeOpBitwiseRight:
		return TokenTypeOpBitwiseLeft
	case TokenTypeOpMinus:
		return TokenTypeOpPlus
	case TokenTypeOpDivision, TokenTypeOpRemainder:
		return TokenTypeOpMultiplication
	default:
		return t
	}
}

// parseStacks contains all the stacks (one of each type) used for parsing expressions.
type parseStacks struct {
	blockStack   []Token
	wrapperStack []WrapperTokenNode
	opStack      []TokenNodeType
}

// TokenNode is a special token that can be used as a node while parsing expressions.
type TokenNode struct {
	Token
	t TokenNodeType
}

// newTokenNode is a constructor for TokenNode.
func newTokenNode(t TokenNodeType, token Token) *TokenNode {
	return &TokenNode{token, t}
}

// WrapperTokenNode is implemented by nodes that wraps another node.
type WrapperTokenNode interface {
	Set(t Token) Token
}

// MethodTokenNode is a token node for methods.
type MethodTokenNode struct {
	method Token
	child  Token
}

// newMethodTokenNode is a constructor for MethodTokenNode.
func newMethodTokenNode(method, child Token) *MethodTokenNode {
	return &MethodTokenNode{method, child}
}

// Execute executes the token functionality.
func (node *MethodTokenNode) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	childBridge := newEmitterReceiverBridge()
	methodBridge := newEmitterReceiverBridge()
	defer func() {
		if r := recover(); r != nil {
			if r != PanicAssertMethod {
				logrus.Fatalf("method token node panic: %v", r)
			}
			go methodBridge.Accept(newEmptierReceiver())
			visitor.VisitString("")
			childBridge.Close()
			panic(PanicAssertMethod)
		}
	}()
	go childBridge.Accept(methodBridge)
	node.child.Execute(r, childBridge, visited)
	node.method.Execute(r, visitor, methodBridge)
}

// NegationWrapperTokenNode is a token node for negation.
type NegationWrapperTokenNode struct {
	child Token
}

// newNegationWrapperTokenNode is a constructor for NegationWrapperTokenNode.
func newNegationWrapperTokenNode() *NegationWrapperTokenNode {
	return &NegationWrapperTokenNode{}
}

// Set sets the child token.
func (node *NegationWrapperTokenNode) Set(t Token) Token {
	node.child = t
	return node
}

func (node *NegationWrapperTokenNode) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	negation := newNegationEmitterReceiver()
	go negation.Accept(visitor)
	node.child.Execute(r, negation, visited)
}

// NegationWrapperTokenNode is a token node for binary operations.
type BinaryOperationTokenNode struct {
	nodeType TokenNodeType
	left     Token
	right    Token
}

// newBinaryOperationTokenNode is a constructor for Binary Token Operations.
func newBinaryOperationTokenNode(nodeType TokenNodeType, left, right Token) Token {
	b := BinaryOperationTokenNode{nodeType, left, right}
	switch nodeType {
	case TokenTypeOpAnd, TokenTypeOpOr:
		return &OperationLogicalTokenNode{b}
	case TokenTypeOpEqual, TokenTypeOpNotEqual:
		return &OperationEqualTokenNode{b}
	case TokenTypeOpGreaterEqual, TokenTypeOpGreater, TokenTypeOpLess, TokenTypeOpLessEqual:
		return &OperationCompareNumbersTokenNode{b}
	case TokenTypeOpBitwiseLeft, TokenTypeOpBitwiseRight, TokenTypeOpPlus, TokenTypeOpMinus, TokenTypeOpMultiplication, TokenTypeOpDivision, TokenTypeOpRemainder:
		return &OperationModifyNumbersTokenNode{b}
	default:
		return nil
	}
}

// NegationWrapperTokenNode is a token node for logical operations.
type OperationLogicalTokenNode struct {
	BinaryOperationTokenNode
}

// Execute executes the token functionality.
func (node *OperationLogicalTokenNode) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	receiver := newBooleanReceiver()
	go node.left.Execute(r, receiver, visited)
	leftValue := receiver.Value()
	if node.nodeType == TokenTypeOpAnd {
		if leftValue == False {
			visitor.VisitString(False)
			return
		}
	} else {
		if leftValue == True {
			visitor.VisitString(True)
			return
		}
	}
	go node.right.Execute(r, receiver, visited)
	visitor.VisitString(receiver.Value())
}

// OperationEqualTokenNode is a token node for comparison operations.
type OperationEqualTokenNode struct {
	BinaryOperationTokenNode
}

// Execute executes the token functionality.
func (node *OperationEqualTokenNode) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	leftReceiver := newTextOnlyReceiver()
	rightReceiver := newTextOnlyReceiver()
	go node.left.Execute(r, leftReceiver, visited)
	go node.right.Execute(r, rightReceiver, visited)
	leftValue := leftReceiver.Value()
	rightValue := rightReceiver.Value()
	if node.nodeType == TokenTypeOpEqual {
		visitor.VisitString(BoolToString(leftValue == rightValue))
		return
	}
	visitor.VisitString(BoolToString(leftValue != rightValue))
}

// OperationEqualTokenNode is a token node for comparison operations of numbers.
type OperationCompareNumbersTokenNode struct {
	BinaryOperationTokenNode
}

// Execute executes the token functionality.
func (node *OperationCompareNumbersTokenNode) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	leftReceiver := newNumberOnlyReceiver()
	rightReceiver := newNumberOnlyReceiver()
	go node.left.Execute(r, leftReceiver, visited)
	go node.right.Execute(r, rightReceiver, visited)
	leftValue := leftReceiver.Value()
	rightValue := rightReceiver.Value()
	switch node.nodeType {
	case TokenTypeOpGreaterEqual:
		visitor.VisitString(BoolToString(leftValue >= rightValue))
	case TokenTypeOpGreater:
		visitor.VisitString(BoolToString(leftValue > rightValue))
	case TokenTypeOpLessEqual:
		visitor.VisitString(BoolToString(leftValue <= rightValue))
	case TokenTypeOpLess:
		visitor.VisitString(BoolToString(leftValue < rightValue))
	default:
		visitor.VisitString(NaN)
	}
}

// OperationEqualTokenNode is a token node for number modifiers.
type OperationModifyNumbersTokenNode struct {
	BinaryOperationTokenNode
}

// Execute executes the token functionality.
func (node *OperationModifyNumbersTokenNode) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	leftReceiver := newNumberOnlyReceiver()
	rightReceiver := newNumberOnlyReceiver()
	go node.left.Execute(r, leftReceiver, visited)
	go node.right.Execute(r, rightReceiver, visited)
	leftValue := leftReceiver.Value()
	rightValue := rightReceiver.Value()
	formatInt := func(n int) string {
		return strconv.Itoa(n)
	}
	formatFloat := func(n float64) string {
		return strconv.FormatFloat(n, 'f', -1, 64)
	}
	switch node.nodeType {
	case TokenTypeOpBitwiseLeft:
		visitor.VisitString(formatInt(int(leftValue) << int(rightValue)))
	case TokenTypeOpBitwiseRight:
		visitor.VisitString(formatInt(int(leftValue) >> int(rightValue)))
	case TokenTypeOpPlus:
		visitor.VisitString(formatFloat(leftValue + rightValue))
	case TokenTypeOpMinus:
		visitor.VisitString(formatFloat(leftValue - rightValue))
	case TokenTypeOpMultiplication:
		visitor.VisitString(formatFloat(leftValue * rightValue))
	case TokenTypeOpDivision:
		visitor.VisitString(formatFloat(leftValue / rightValue))
	case TokenTypeOpRemainder:
		visitor.VisitString(formatInt(int(leftValue) % int(rightValue)))
	default:
		visitor.VisitString("0")
	}
}

// ParseExpr parses the expression returning the root node for the parsed tree.
func ParseExpr(expr []*TokenNode) Token {
	return parseExpr(expr, &parseStacks{})
}

// parseExpr parses the expression returning the root node for the parsed tree.
func parseExpr(expr []*TokenNode, stacks *parseStacks) Token {
	for i := 0; i < len(expr); i++ {
		switch e := expr[i]; e.t {
		case TokenTypeValue:
			var value Token = e
			for _, w := range stacks.wrapperStack {
				value = w.Set(value)
			}
			stacks.wrapperStack = nil
			stacks.blockStack = append(stacks.blockStack, value)
		case TokenTypeMethod:
			stacks.blockStack = append(stacks.blockStack, newMethodTokenNode(e, popBlock(&stacks.blockStack)))
		case TokenTypeNegation:
			stacks.wrapperStack = append(stacks.wrapperStack, newNegationWrapperTokenNode())
		case TokenTypeOpOr, TokenTypeOpAnd, TokenTypeOpEqual, TokenTypeOpNotEqual,
			TokenTypeOpGreaterEqual, TokenTypeOpGreater, TokenTypeOpLessEqual, TokenTypeOpLess,
			TokenTypeOpBitwiseLeft, TokenTypeOpBitwiseRight, TokenTypeOpPlus, TokenTypeOpMinus, TokenTypeOpMultiplication, TokenTypeOpDivision, TokenTypeOpRemainder:
			handleParseOpNode(expr[i].t, stacks)
		default:
			logrus.Fatalf("unknown token expression node %v", e.t)
		}
	}
	return returnParseExpr(stacks)
}

// returnParseExpr ends the parse expression processes, returning the final node.
// clears all the stacks and checks for errors.
func returnParseExpr(stacks *parseStacks) Token {
	for i := len(stacks.opStack) - 1; i > -1; i-- {
		doOperation(stacks)
	}
	if len(stacks.blockStack) != 1 {
		log.Fatalf("group ended with %d nodes left on blocks stack... expected 1", len(stacks.blockStack))
	}
	if len(stacks.opStack) != 0 {
		log.Fatalf("group ended with %d nodes left on operations stack... expected 0", len(stacks.blockStack))
	}
	return popBlock(&stacks.blockStack)
}

// handleParseOpNode handles the parsed operation node.
func handleParseOpNode(opType TokenNodeType, stacks *parseStacks) {
	for i := len(stacks.opStack) - 1; i > -1 && precedence(stacks.opStack[i]) >= precedence(opType); i-- {
		doOperation(stacks)
	}
	stacks.opStack = append(stacks.opStack, opType)
}

// doOperation runs some operation operation, manipulating the stacks.
func doOperation(stacks *parseStacks) {
	opType := popOp(&stacks.opStack)
	valueB := popBlock(&stacks.blockStack)
	valueA := popBlock(&stacks.blockStack)
	stacks.blockStack = append(stacks.blockStack, newBinaryOperationTokenNode(opType, valueA, valueB))
}

// popBlock pops a value from the block stack.
func popBlock(s *[]Token) Token {
	stack := *s
	n := len(stack) - 1
	val := stack[n]
	*s = stack[:n]
	return val
}

// popOp pops a value from the operations stack.
func popOp(s *[]TokenNodeType) TokenNodeType {
	stack := *s
	n := len(stack) - 1
	val := stack[n]
	*s = stack[:n]
	return val
}
