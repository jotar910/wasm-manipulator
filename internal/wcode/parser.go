package wcode

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"joao/wasm-manipulator/internal/wtemplate"
	"joao/wasm-manipulator/pkg/wutils"

	"github.com/sirupsen/logrus"
)

// Block represents a code block.
type Block interface {
	Visited
	VisitedTree
	String() string
	StringIndent(string) string
	push(Block)
	setParent(Block)
	getParent() Block
	findEqual(Block) Block
	funcInstr() *Instruction
}

// blockImpl is the base implementation for the code block.
type blockImpl struct {
	parent Block
}

// getParent returns the code block parent.
func (bi *blockImpl) getParent() Block {
	return bi.parent
}

// setParent sets the code block parent.
func (bi *blockImpl) setParent(b Block) {
	bi.parent = b
}

// funcInstr returns the function instruction block where the instruction belongs.
func (bi *blockImpl) funcInstr() *Instruction {
	parent := bi.getParent()
	for parent != nil {
		if parentInstr, ok := parent.(*Instruction); ok && parentInstr.name == instructionFunction {
			return parentInstr
		}
		parent = parent.getParent()
	}
	return nil
}

// element is the code block that wraps a list of code blocks.
type element struct {
	*blockImpl
	sb     *strings.Builder
	blocks []Block
}

// newElement is the constructor for element.
func newElement() *element {
	return &element{blockImpl: new(blockImpl), sb: new(strings.Builder)}
}

// String returns the textual value for the code block.
func (el *element) String() string {
	sb := new(strings.Builder)
	for _, b := range el.blocks {
		sb.WriteString(b.String())
	}
	return sb.String()
}

// StringIndent returns the textual value indented for the code block.
func (el *element) StringIndent(ident string) string {
	sb := new(strings.Builder)
	for _, b := range el.blocks {
		sb.WriteString(ident + b.StringIndent(ident) + "\n")
	}
	return sb.String()
}

// Accept accepts a visitor to visit the code block.
func (el *element) Accept(v Visitor) bool {
	return v.VisitElement(el)
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (el *element) Traverse(v Visitor) {
	el.Accept(v)
	for _, b := range el.blocks {
		b.Traverse(v)
	}
}

// TraverseConditional transverses the element until a visit for some code block is accepted.
func (el *element) TraverseConditional(v Visitor) {
	if el.Accept(v) {
		return
	}
	for _, b := range el.blocks {
		b.Traverse(v)
	}
}

// add adds a new character byte to the block buffer.
func (el *element) add(b byte) {
	el.sb.WriteByte(b)
}

// flush clears the block buffer and handle its value.
func (el *element) flush() {
	val := el.sb.String()
	defer el.sb.Reset()
	if len(strings.TrimSpace(val)) > 0 {
		el.blocks = append(el.blocks, newText(val))
	}
}

// push pushes a child block.
func (el *element) push(b Block) {
	el.flush()
	el.blocks = append(el.blocks, b)
}

// findEqual returns a block equal to the provided one.
func (el *element) findEqual(b Block) Block {
	if b.String() == el.String() {
		return el
	}
	for _, child := range el.blocks {
		if res := child.findEqual(b); res != nil {
			return res
		}
	}
	return nil
}

// childIndex returns the index of the instruction children.
func (el *element) childIndex(block Block) int {
	for i, b := range el.blocks {
		if b == block {
			return i
		}
	}
	return -1
}

// replaceChild replaces some child block.
func (el *element) replaceChild(child Block, blocks []Block) error {
	index := el.childIndex(child)
	if index == -1 {
		return errors.New("child block not found")
	}
	return el.replaceChildByIndex(index, blocks)
}

// replaceChildByIndex replaces some child block by index.
func (el *element) replaceChildByIndex(index int, blocks []Block) error {
	length := len(el.blocks)
	if index < 0 || index >= length {
		return fmt.Errorf("index out of range (index %d, length: %d)", index, length)
	}
	children := append([]Block{}, el.blocks[:index]...)
	if len(blocks) != 0 {
		for _, b := range blocks {
			b.setParent(el)
		}
		children = append(children, blocks...)
	}
	if index < length-1 {
		children = append(children, el.blocks[index+1:]...)
	}
	el.blocks = children
	return nil
}

// Instruction is the code block that represents a web assembly instruction.
type Instruction struct {
	*blockImpl
	sb     *strings.Builder
	name   string
	values []Block
}

// newInstruction is the constructor for Instruction.
func newInstruction() *Instruction {
	return &Instruction{blockImpl: new(blockImpl), sb: new(strings.Builder)}
}

// Child returns the child instruction at some index.
func (iv *Instruction) Child(index int) Block {
	if index < 0 || index >= len(iv.values) {
		logrus.Fatalf("getting child value from instr: index out of range (index: %d, length: %d)", index, len(iv.values))
	}
	return iv.values[index]
}

// String returns the textual value for the code block.
func (iv *Instruction) String() string {
	if len(iv.values) == 0 {
		return fmt.Sprintf("(%s)", iv.name)
	}
	var valuesStr []string
	for _, v := range iv.values {
		valuesStr = append(valuesStr, v.String())
	}
	return fmt.Sprintf("(%s %s)", iv.name, strings.Join(valuesStr, " "))
}

// StringIndent returns the textual value indented for the code block.
func (iv *Instruction) StringIndent(ident string) string {
	if len(iv.values) == 0 {
		return fmt.Sprintf("(%s)", iv.name)
	}
	var index int
	sb := new(strings.Builder)
	sb.WriteString("(" + iv.name)
	newIdent := ident + "  "
	for index < len(iv.values) {
		if _, ok := iv.values[index].(*Instruction); ok {
			break
		}
		sb.WriteString(" " + iv.values[index].StringIndent(newIdent))
		index++
	}
	if index < len(iv.values) {
		var isInsideFunction bool
		if iv.name != instructionModule {
			parent := iv
			for parent.name != instructionFunction {
				parentAux, ok := parent.getParent().(*Instruction)
				if !ok {
					parent = nil
					break
				}
				parent = parentAux
			}
			isInsideFunction = parent != nil && parent.name == instructionFunction
		}
		for ; index < len(iv.values); index++ {
			v := iv.values[index]
			if iv.name == instructionModule || isInsideFunction && isCodeInstruction(v) {
				sb.WriteString("\n" + newIdent + v.StringIndent(newIdent))
			} else {
				sb.WriteString(" " + v.StringIndent(newIdent))
			}
		}
	}
	if iv.name == instructionModule {
		sb.WriteString("\n")
	}
	sb.WriteString(")")
	return sb.String()
}

// Accept accepts a visitor to visit the code block.
func (iv *Instruction) Accept(v Visitor) bool {
	return v.VisitInstruction(iv)
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (iv *Instruction) Traverse(v Visitor) {
	iv.Accept(v)
	for _, b := range iv.values {
		b.Traverse(v)
	}
}

// TraverseConditional transverses the element until a visit for some code block is accepted.
func (iv *Instruction) TraverseConditional(v Visitor) {
	if iv.Accept(v) {
		return
	}
	for _, b := range iv.values {
		b.Traverse(v)
	}
}

// add adds a new character byte to the block buffer.
func (iv *Instruction) add(b byte) {
	iv.sb.WriteByte(b)
}

// flush clears the block buffer and handle its value.
func (iv *Instruction) flush() {
	val := iv.sb.String()
	defer iv.sb.Reset()
	if len(strings.TrimSpace(val)) == 0 {
		return
	}
	if iv.name == "" {
		iv.name = val
	} else {
		iv.values = append(iv.values, newText(val))
	}
}

// push pushes a child block.
func (iv *Instruction) push(b Block) {
	iv.flush()
	iv.values = append(iv.values, b)
}

// findEqual returns a block equal to the provided one.
func (iv *Instruction) findEqual(b Block) Block {
	if iv.String() == b.String() {
		return iv
	}
	for _, v := range iv.values {
		if res := v.findEqual(b); res != nil {
			return res
		}
	}
	return nil
}

// childIndex returns the index of the instruction children.
func (iv *Instruction) childIndex(block Block) int {
	for i, b := range iv.values {
		if b == block {
			return i
		}
	}
	return -1
}

// addChildren adds children blocks at the same index as the provided block.
func (iv *Instruction) addChildren(child Block, blocks ...Block) error {
	index := iv.childIndex(child)
	if index == -1 {
		return errors.New("child block not found")
	}
	return iv.addChildrenByIndex(index, blocks...)
}

// addChildrenByIndex adds children blocks at some index.
func (iv *Instruction) addChildrenByIndex(index int, blocks ...Block) error {
	length := len(iv.values)
	if index < 0 || index >= length {
		return fmt.Errorf("index out of range (index %d, length: %d)", index, length)
	}
	children := append([]Block{}, iv.values[:index]...)
	children = append(children, blocks...)
	if index < length {
		children = append(children, iv.values[index:]...)
	}
	iv.values = children
	return nil
}

// removeChild removes a child block.
func (iv *Instruction) removeChild(child Block) error {
	return iv.removeChildren(child, 1)
}

// removeChildByIndex removes a child block by index.
func (iv *Instruction) removeChildByIndex(index int) error {
	return iv.removeChildrenByIndex(index, 1)
}

// removeChild removes a set of children blocks by the given start block.
func (iv *Instruction) removeChildren(child Block, count int) error {
	index := iv.childIndex(child)
	if index == -1 {
		return errors.New("child block not found")
	}
	return iv.removeChildrenByIndex(index, count)
}

// removeChildrenByIndex removes a set of children blocks by the given start index.
func (iv *Instruction) removeChildrenByIndex(index int, count int) error {
	length := len(iv.values)
	if index < 0 || index >= length {
		return fmt.Errorf("index out of range (index %d, length: %d)", index, length)
	}
	children := append([]Block{}, iv.values[:index]...)
	if index < length-count {
		children = append(children, iv.values[index+count:]...)
	}
	iv.values = children
	return nil
}

// replaceChildWithCode replaces some child block with code.
func (iv *Instruction) replaceChildWithCode(child Block, code string) error {
	index := iv.childIndex(child)
	if index == -1 {
		return errors.New("child block not found")
	}
	// Parse new code blocks.
	newBlockEl := NewCodeParser(code).parse()
	return iv.replaceChildByIndex(index, newBlockEl.blocks)
}

// replaceChild replaces some child block.
func (iv *Instruction) replaceChild(child Block, blocks []Block) error {
	index := iv.childIndex(child)
	if index == -1 {
		return errors.New("child block not found")
	}
	return iv.replaceChildByIndex(index, blocks)
}

// replaceChildByIndex replaces some child block by index.
func (iv *Instruction) replaceChildByIndex(index int, blocks []Block) error {
	length := len(iv.values)
	if index < 0 || index >= length {
		return fmt.Errorf("index out of range (index %d, length: %d)", index, length)
	}
	children := append([]Block{}, iv.values[:index]...)
	if len(blocks) != 0 {
		for _, b := range blocks {
			b.setParent(iv)
		}
		children = append(children, blocks...)
	}
	if index < length-1 {
		children = append(children, iv.values[index+1:]...)
	}
	iv.values = children
	return nil
}

// keyword is the code block that represents the keyword input code.
type keyword struct {
	*element
}

// newKeyword is the constructor for keyword.
func newKeyword() *keyword {
	return &keyword{newElement()}
}

// String returns the textual value for the code block.
func (kw *keyword) String() string {
	return fmt.Sprintf(`%%%s%%`, kw.element.String())
}

// StringIndent returns the textual value indented for the code block.
func (kw *keyword) StringIndent(string) string {
	return kw.String()
}

// Accept accepts a visitor to visit the code block.
func (kw *keyword) Accept(v Visitor) bool {
	return v.VisitKeyword(kw)
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (kw *keyword) Traverse(v Visitor) {
	kw.Accept(v)
	kw.element.Traverse(v)
}

// TraverseConditional transverses the element until a visit for some code block is accepted.
func (kw *keyword) TraverseConditional(v Visitor) {
	if !kw.Accept(v) {
		kw.element.Traverse(v)
	}
}

// quoted is the code block that represents the string quoted data.
type quoted struct {
	*element
	c byte
}

// newQuoted is the constructor for quoted.
func newQuoted(c byte) *quoted {
	return &quoted{newElement(), c}
}

// String returns the textual value for the code block.
func (qt *quoted) String() string {
	return fmt.Sprintf(`%c%s%c`, qt.c, qt.element.String(), qt.c)
}

// StringIndent returns the textual value indented for the code block.
func (qt *quoted) StringIndent(string) string {
	return qt.String()
}

// Accept accepts a visitor to visit the code block.
func (qt *quoted) Accept(v Visitor) bool {
	return v.VisitQuoted(qt)
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (qt *quoted) Traverse(v Visitor) {
	qt.Accept(v)
	qt.element.Traverse(v)
}

// TraverseConditional transverses the element until a visit for some code block is accepted.
func (qt *quoted) TraverseConditional(v Visitor) {
	if !qt.Accept(v) {
		qt.element.Traverse(v)
	}
}

// text is the code block for the string data.
type text struct {
	*blockImpl
	sb *strings.Builder
}

// newText is the constructor for text.
func newText(v string) *text {
	sb := new(strings.Builder)
	sb.WriteString(v)
	return &text{new(blockImpl), sb}
}

// String returns the textual value for the code block.
func (t *text) String() string {
	return t.sb.String()
}

// StringIndent returns the textual value indented for the code block.
func (t *text) StringIndent(string) string {
	return t.String()
}

// Accept accepts a visitor to visit the code block.
func (t *text) Accept(v Visitor) bool {
	return v.VisitText(t)
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (t *text) Traverse(v Visitor) {
	t.Accept(v)
}

// TraverseConditional transverses the element until a visit for some code block is accepted.
func (t *text) TraverseConditional(v Visitor) {
	t.Accept(v)
}

// push pushes a child block.
func (t *text) push(_ Block) {
	panic("push must not be called in textual elements")
}

// findEqual returns a block equal to the provided one.
func (t *text) findEqual(b Block) Block {
	if t.String() == b.String() {
		return t
	}
	return nil
}

// evaluation is the code block that represents the evaluation input code.
type evaluation struct {
	*blockImpl
	sb     *strings.Builder
	blocks []Block
}

// newEvaluation is a constructor for evaluation.
func newEvaluation() *evaluation {
	return &evaluation{blockImpl: new(blockImpl), sb: new(strings.Builder)}
}

// String returns the textual value for the code block.
func (eval *evaluation) String() string {
	return fmt.Sprintf("/%s/", eval.CleanString())
}

// CleanString returns the textual clean value for the code block.
func (eval *evaluation) CleanString() string {
	var res []string
	for _, b := range eval.blocks {
		res = append(res, b.String())
	}
	return strings.Join(res, "")
}

// StringIndent returns the textual value indented for the code block.
func (eval *evaluation) StringIndent(string) string {
	return eval.String()
}

// Accept accepts a visitor to visit the code block.
func (eval *evaluation) Accept(v Visitor) bool {
	return v.VisitEvaluation(eval)
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (eval *evaluation) Traverse(v Visitor) {
	eval.Accept(v)
	for _, b := range eval.blocks {
		b.Traverse(v)
	}
}

// TraverseConditional transverses the element until a visit for some code block is accepted.
func (eval *evaluation) TraverseConditional(v Visitor) {
	eval.Accept(v)
}

// add adds a new character byte to the block buffer.
func (eval *evaluation) add(b byte) {
	eval.sb.WriteByte(b)
}

// identifier returns the identifier name being accumulated and clears the block buffer.
func (eval *evaluation) identifier() Block {
	val := eval.sb.String()
	defer eval.sb.Reset()
	if len(strings.TrimSpace(val)) == 0 {
		blocksLen := len(eval.blocks)
		if blocksLen == 0 {
			return nil
		}
		last := eval.blocks[blocksLen-1]
		eval.blocks = eval.blocks[:blocksLen-1]
		return last
	}
	values := strings.Split(val, "+")
	for i := 0; i < len(values)-1; i++ {
		eval.blocks = append(eval.blocks, newEvaluationText(values[i]))
	}
	return newEvaluationText(strings.TrimSpace(values[len(values)-1]))
}

// flush clears the block buffer and handle its value.
func (eval *evaluation) flush() {
	val := eval.sb.String()
	defer eval.sb.Reset()
	if len(strings.TrimSpace(val)) == 0 {
		return
	}
	eval.blocks = append(eval.blocks, newEvaluationText(val))
}

// push pushes a child block.
func (eval *evaluation) push(b Block) {
	eval.flush()
	eval.blocks = append(eval.blocks, b)
}

// findEqual returns a block equal to the provided one.
func (eval *evaluation) findEqual(b Block) Block {
	if eval.String() == b.String() {
		return eval
	}
	return nil
}

// evaluationId is the code block for the evaluation string data.
type evaluationId struct {
	*blockImpl
	value string
}

// newEvaluationId is a constructor for evaluationId.
func newEvaluationId(v string) *evaluationId {
	return &evaluationId{new(blockImpl), v}
}

// String returns the textual value for the code block.
func (ei *evaluationId) String() string {
	return fmt.Sprintf("#%s", ei.value)
}

// StringIndent returns the textual value indented for the code block.
func (ei *evaluationId) StringIndent(ident string) string {
	return ei.String()
}

// Accept accepts a visitor to visit the code block.
func (ei *evaluationId) Accept(v Visitor) bool {
	// Empty by desing.
	return false
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (ei *evaluationId) Traverse(v Visitor) {
	// Empty by desing.
}

// TraverseConditional transverses the element until a visit for some code block is accepted.
func (ei *evaluationId) TraverseConditional(v Visitor) {
	// Empty by desing.
}

// findEqual returns a block equal to the provided one.
func (ei *evaluationId) findEqual(b Block) Block {
	if ei.String() == b.String() {
		return ei
	}
	return nil
}

// push pushes a child block.
func (ei *evaluationId) push(_ Block) {
	// Empty by design.
}

// evaluationText is the code block for the evaluation string data.
type evaluationText struct {
	*text
}

// newEvaluationText is a constructor for evaluationText.
func newEvaluationText(v string) *evaluationText {
	return &evaluationText{newText(v)}
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (et *evaluationText) Traverse(v Visitor) {
	// Empty by desing.
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (et *evaluationText) TraverseConditional(v Visitor) {
	// Empty by desing.
}

// evaluationKeyword is the code block for the evaluation keyword data.
type evaluationKeyword struct {
	*evaluation
}

// newEvaluationKeyword is a constructor for evaluationKeyword.
func newEvaluationKeyword() *evaluationKeyword {
	return &evaluationKeyword{newEvaluation()}
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (eKw *evaluationKeyword) Traverse(v Visitor) {
	// Empty by design.
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (eKw *evaluationKeyword) TraverseConditional(v Visitor) {
	// Empty by design.
}

// String returns the textual value for the code block.
func (eKw *evaluationKeyword) String() string {
	var res []string
	for _, b := range eKw.blocks {
		res = append(res, b.String())
	}
	return strings.Join(res, "+")
}

// String returns the textual value for the code block.
func (eKw *evaluationKeyword) StringIndent(string) string {
	return eKw.String()
}

// evaluationIndex represents some evaluation that contains an access to a member by index.
type evaluationIndex struct {
	*evaluation
	identifier Block
}

// newEvaluationIndex is a constructor for evaluationIndex.
func newEvaluationIndex(identifier Block) *evaluationIndex {
	return &evaluationIndex{newEvaluation(), identifier}
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (evalIndex *evaluationIndex) Traverse(v Visitor) {
	// Empty by design.
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (evalIndex *evaluationIndex) TraverseConditional(v Visitor) {
	// Empty by design.
}

// String returns the textual value for the code block.
func (evalIndex *evaluationIndex) String() string {
	var res []string
	for _, b := range evalIndex.blocks {
		res = append(res, b.String())
	}
	return fmt.Sprintf("%s[%s]", evalIndex.identifier.String(), strings.Join(res, ""))
}

// StringIndent returns the textual value indented for the code block.
func (evalIndex *evaluationIndex) StringIndent(string) string {
	return evalIndex.String()
}

// IndexString returns the textual value for the index.
func (evalIndex *evaluationIndex) IndexString() string {
	return evalIndex.evaluation.String()
}

// evaluationQuoted is the code block that represents the string quoted data for some evalation.
type evaluationQuoted struct {
	*quoted
}

// newEvaluationQuoted is a constructor for evaluationQuoted.
func newEvaluationQuoted(c byte) *evaluationQuoted {
	return &evaluationQuoted{newQuoted(c)}
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (evalQt *evaluationQuoted) Traverse(v Visitor) {
	// Empty by design.
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (evalQt *evaluationQuoted) TraverseConditional(v Visitor) {
	// Empty by design.
}

// String returns the textual value for the code block.
func (evalQt *evaluationQuoted) String() string {
	var res []string
	sb := new(strings.Builder)
	for _, b := range evalQt.blocks {
		if textB, ok := b.(*text); ok {
			txt := textB.String()
			for i := 0; i < len(txt); i++ {
				if txt[i] == '"' && (i == 0 || txt[i-1] != '\\') {
					sb.WriteByte('\\')
				}
				sb.WriteByte(txt[i])
			}
			res = append(res, fmt.Sprintf(`"%s"`, sb.String()))
			sb.Reset()
		} else {
			res = append(res, b.String())
		}
	}
	return strings.Join(res, "+")
}

// StringIndent returns the textual value indented for the code block.
func (evalQt *evaluationQuoted) StringIndent(string) string {
	return evalQt.String()
}

// flush clears the block buffer and handle its value.
func (evalQt *evaluationQuoted) flush() {
	val := evalQt.sb.String()
	defer evalQt.sb.Reset()
	if val != "" {
		evalQt.blocks = append(evalQt.blocks, newText(val))
	}
}

// evaluationRef is the code block that represents a variable reference to evaluate.
type evaluationRef struct {
	*text
	indexes []Block
}

// newEvaluationRef is a constructor for evaluationRef.
func newEvaluationRef(value string) *evaluationRef {
	return &evaluationRef{text: newText(value)}
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (erv *evaluationRef) Traverse(v Visitor) {
	erv.Accept(v)
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (erv *evaluationRef) TraverseConditional(v Visitor) {
	erv.Accept(v)
}

// Accept accepts a visitor to visit the code block.
func (erv *evaluationRef) Accept(v Visitor) bool {
	return v.VisitEvaluationRef(erv)
}

// String returns the textual value for the code block.
func (erv *evaluationRef) String() string {
	sb := new(strings.Builder)
	for _, b := range erv.indexes {
		sb.WriteRune('[')
		sb.WriteString(b.(*evaluationIndex).evaluation.String())
		sb.WriteRune(']')
	}
	return fmt.Sprintf("%s%s", erv.text.String(), sb.String())
}

// Identifier returns the textual value for the identifier block.
func (erv *evaluationRef) Identifier() string {
	sb := new(strings.Builder)
	for _, b := range erv.indexes {
		sb.WriteRune('[')
		sb.WriteString(b.(*evaluationIndex).evaluation.String())
		sb.WriteRune(']')
	}
	return erv.text.String()
}

// push pushes a child block.
func (erv *evaluationRef) push(index Block) {
	erv.indexes = append(erv.indexes, index)
}

// comment is the code block the web assembly comments.
type comment struct {
	*blockImpl
}

// newComment is a constructor for comment.
func newComment() *comment {
	return &comment{new(blockImpl)}
}

// String returns the textual value for the code block.
func (t *comment) String() string {
	return ""
}

// StringIndent returns the textual value indented for the code block.
func (t *comment) StringIndent(string) string {
	return t.String()
}

// Accept accepts a visitor to visit the code block.
func (t *comment) Accept(Visitor) bool {
	// Empty by design.
	return false
}

// Traverse transverses the element, accepting a visit for all code blocks.
func (t *comment) Traverse(Visitor) {
	// Empty by design.
}

// TraverseConditional transverses the element until a visit for some code block is accepted.
func (t *comment) TraverseConditional(Visitor) {
	// Empty by design.
}

// push pushes a child block.
func (t *comment) push(_ Block) {
	panic("push must not be called in comment elements")
}

// findEqual returns a block equal to the provided one.
func (t *comment) findEqual(Block) Block {
	return nil
}

// CodeParser is responsible to parse code.
type CodeParser struct {
	code string
}

// NewCodeParser is the constructor for CodeParser.
func NewCodeParser(code string) *CodeParser {
	return &CodeParser{
		code: wtemplate.ClearString(code),
	}
}

// Len returns the code length.
func (parser *CodeParser) Len() int {
	return len(parser.code)
}

// Index returns the code byte on some position.
func (parser *CodeParser) Index(i int) byte {
	return parser.code[i]
}

// Parse parses the code to blocks.
func (parser *CodeParser) Parse() Block {
	return parser.parse()
}

// parse parses element blocks.
func (parser *CodeParser) parse() *element {
	el := newElement()
	defer el.flush()
	for i := 0; i < parser.Len(); {
		switch c := parser.Index(i); {
		case c == '(':
			i = addBlock(el, i, parser.parseInstruction)
		case c == '%':
			i = addBlock(el, i, parser.parseKeyword)
		case c == '/':
			i = addBlock(el, i, parser.parseEvaluation)
		case c == '"', c == '\'', c == '`':
			i = addBlock(el, i, parser.parseString)
		case c == ';':
			i = addBlock(el, i, parser.parseInlineComment)
		default:
			el.add(c)
			i++
		}
	}
	return el
}

// parseInstruction parses Instruction blocks.
func (parser *CodeParser) parseInstruction(start int) (Block, int) {
	instr := newInstruction()
	defer instr.flush()
	for i := start + 1; i < parser.Len(); {
		switch c := parser.Index(i); {
		case c == '"', c == '\'', c == '`':
			i = addBlock(instr, i, parser.parseString)
		case c == '#':
			i = addBlock(instr, i, parser.parseEvaluationRef)
		case c == '(':
			i = addBlock(instr, i, parser.parseInstruction)
		case c == ')':
			return instr, i
		case c == '%':
			i = addBlock(instr, i, parser.parseKeyword)
		case c == '/':
			i = addBlock(instr, i, parser.parseEvaluation)
		case c == ';':
			if i == start+1 {
				return parser.parseMultilineComment(start)
			}
			return parser.parseInlineComment(start)
		case unicode.IsSpace(rune(c)):
			instr.flush()
			i++
		default:
			instr.add(c)
			i++
		}
	}
	return instr, parser.Len()
}

// parseKeyword parses keyword blocks.
func (parser *CodeParser) parseKeyword(start int) (Block, int) {
	kw := newKeyword()
	defer kw.flush()
	for i := start + 1; i < parser.Len(); {
		switch c := parser.Index(i); {
		case c == '"', c == '\'', c == '`':
			i = addBlock(kw, i, parser.parseQuote)
		case c == '%':
			return kw, i
		default:
			kw.add(c)
			i++
		}
	}
	return kw, parser.Len()
}

// parseString parses basic string quoted blocks.
func (parser *CodeParser) parseString(start int) (Block, int) {
	qtC := parser.Index(start)
	qt := newQuoted(qtC)
	defer qt.flush()
	for i := start + 1; i < parser.Len(); {
		switch c := parser.Index(i); {
		case c == qtC && i > start+1 && parser.Index(i-1) != '\\':
			return qt, i
		default:
			qt.add(c)
			i++
		}
	}
	return qt, parser.Len()
}

// parseQuote parses string quoted blocks.
func (parser *CodeParser) parseQuote(start int) (Block, int) {
	qtC := parser.Index(start)
	qt := newQuoted(qtC)
	defer qt.flush()
	for i := start + 1; i < parser.Len(); {
		switch c := parser.Index(i); {
		case c == qtC && i > start+1 && parser.Index(i-1) != '\\':
			return qt, i
		case c == '%':
			i = addBlock(qt, i, parser.parseKeyword)
		default:
			qt.add(c)
			i++
		}
	}
	return qt, parser.Len()
}

// parseEvaluation parses evaluation blocks.
func (parser *CodeParser) parseEvaluation(start int) (Block, int) {
	eval := newEvaluation()
	defer eval.flush()
	return parser.parseEvaluationLike(eval, start, '/')
}

// parseEvaluationId parses evaluation id blocks.
func (parser *CodeParser) parseEvaluationId(start int) (Block, int) {
	sb := new(strings.Builder)
	i := start + 1
	for ; i < parser.Len(); i++ {
		c := parser.Index(i)
		if !wutils.IsIdentifier(rune(c)) && c != '$' {
			break
		}
		sb.WriteByte(c)
	}
	return newEvaluationId(sb.String()), i - 1
}

// parseEvaluationQuote parses evaluation quoted blocks.
func (parser *CodeParser) parseEvaluationQuote(start int) (Block, int) {
	qtC := parser.Index(start)
	qt := newEvaluationQuoted(qtC)
	defer qt.flush()
	for i := start + 1; i < parser.Len(); {
		switch c := parser.Index(i); {
		case c == qtC && i > start+1 && parser.Index(i-1) != '\\':
			return qt, i
		case c == '%':
			i = addBlock(qt, i, parser.parseKeyword)
		case c == ' ':
			qt.add(c)
			i++
		default:
			if c == '$' && qtC == '`' && i < parser.Len()-1 && parser.Index(i+1) == '{' {
				qt.flush()
				i = addBlock(qt, i, parser.parseEvaluationKeyword)
				break
			}
			qt.add(c)
			i++
		}
	}
	return qt, parser.Len()
}

// parseEvaluationKeyword parses evaluation keyword blocks.
func (parser *CodeParser) parseEvaluationKeyword(start int) (Block, int) {
	evalKw := newEvaluationKeyword()
	defer evalKw.flush()
	_, i := parser.parseEvaluationLike(evalKw.evaluation, start+1, '}')
	return evalKw, i
}

// parseEvaluationIndex parses evaluation index blocks.
func (parser *CodeParser) parseEvaluationIndex(identifier Block) func(int) (Block, int) {
	return func(start int) (Block, int) {
		evalIndex := newEvaluationIndex(identifier)
		defer evalIndex.flush()
		_, i := parser.parseEvaluationLike(evalIndex.evaluation, start, ']')
		return evalIndex, i
	}
}

// parseEvaluationIndex parses evaluation references blocks.
func (parser *CodeParser) parseEvaluationRef(start int) (Block, int) {
	i := start + 1
	for i < parser.Len() {
		r := rune(parser.Index(i))
		if unicode.IsSpace(r) || r == '(' || r == ')' || r == '[' {
			break
		}
		i++
	}
	ref := newEvaluationRef(parser.code[start:i])
	for i < parser.Len() {
		r := rune(parser.Index(i))
		if r != '[' {
			break
		}
		i = addBlock(ref, i, parser.parseEvaluationIndex(ref))
	}
	return ref, i
}

// parseMultilineComment parses multi-line comment blocks.
func (parser *CodeParser) parseMultilineComment(start int) (Block, int) {
	cmt := newComment()
	for i := start + 2; i < parser.Len(); {
		if strings.HasPrefix(parser.code[i:], ";)") {
			return cmt, i + 1
		}
		i++
	}
	return cmt, parser.Len()
}

// parseInlineComment parses simple inline comment blocks.
func (parser *CodeParser) parseInlineComment(start int) (Block, int) {
	cmt := newComment()
	for i := start + 1; i < parser.Len(); {
		if parser.Index(i) == '\n' {
			return cmt, i
		}
		i++
	}
	return cmt, parser.Len()
}

// parseEvaluation parses blocks similar to evaluation blocks.
func (parser *CodeParser) parseEvaluationLike(eval *evaluation, start int, endC byte) (Block, int) {
	for i := start + 1; i < parser.Len(); {
		switch c := parser.Index(i); {
		case c == '%':
			eval.flush()
			i = addBlock(eval, i, parser.parseKeyword)
		case c == '"', c == '\'', c == '`':
			eval.flush()
			i = addBlock(eval, i, parser.parseEvaluationQuote)
		case c == '#':
			eval.flush()
			i = addBlock(eval, i, parser.parseEvaluationId)
		case c == '[' && endC == ']':
			eval.flush()
			i = addBlock(eval, i, parser.parseEvaluationIndex(eval.identifier()))
		case c == endC:
			eval.flush()
			return eval, i
		default:
			eval.add(c)
			i++
		}
	}
	return eval, parser.Len()
}

// addBlock adds a block to some target.
// parses the code using the callback function.
// returns the next original position.
func addBlock(target Block, i int, fn func(int) (Block, int)) int {
	b, end := fn(i)
	if _, ok := b.(*comment); !ok {
		// comments must not be added as value.
		target.push(b)
		b.setParent(target)
	}
	return end + 1
}
