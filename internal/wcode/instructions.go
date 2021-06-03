package wcode

import (
	"errors"
	"strings"

	"joao/wasm-manipulator/internal/wlang"

	"github.com/sirupsen/logrus"
)

const returnKeyword = "return_"
const targetInstructionKeyword = "target"

var (
	instructionModule   = "module"
	instructionFunction = "func"
	instructionGlobal   = "global"
	instructionMutable  = "mut"
	instructionImport   = "import"
	instructionExport   = "export"
	instructionTable    = "table"
	instructionMemory   = "memory"
	instructionData     = "data"
	instructionElem     = "elem"
	instructionStart    = "start"

	instructionType             = "type"
	instructionResult           = "result"
	instructionLocal            = "local"
	instructionParam            = "param"
	instructionCodeCall         = "call"
	instructionCodeCallIndirect = "call_indirect"
	instructionCodeTeeLocal     = "local.tee"
	instructionCodeSetLocal     = "local.set"
	instructionCodeGetLocal     = "local.set"
	instructionCodeSetGlobal    = "global.set"
	instructionCodeConst        = "const"
	instructionCodeReturn       = "return"

	instructionCallZonePush = "(call $zone.push)"
	instructionCallZonePop  = "(call $zone.pop)"
)

var instructionsOrderValues = map[string]int{
	instructionType:     0,
	instructionImport:   1,
	instructionFunction: 2,
	instructionTable:    3,
	instructionMemory:   4,
	instructionGlobal:   5,
	instructionExport:   6,
	instructionStart:    7,
	instructionElem:     8,
	instructionData:     9,
}

// FuncInstrsString returns the instructions code for some block.
func FuncInstrsString(b Block) string {
	bInstr, ok := b.(*Instruction)
	if !ok || bInstr.name != instructionFunction {
		return b.String()
	}
	if _, instrIndex := findFirstInstruction(bInstr.values); instrIndex != -1 {
		var res []string
		for i := instrIndex; i < len(bInstr.values); i++ {
			res = append(res, bInstr.values[i].String())
		}
		return strings.Join(res, " ")
	}
	return ""
}

// AddCodeToControlFlow adds code blocks to some control flow instruction.
func AddCodeToControlFlow(block Block, code string, offset int) error {
	// Parse new code blocks.
	newBlockEl := NewCodeParser(code).parse()

	return AddBlocksToControlFlow(block, newBlockEl.blocks, offset)
}

// AddCodeToControlFlow adds code blocks to some control flow instruction.
func AddBlocksToControlFlow(block Block, newBlocks []Block, offset int) error {
	// Find the control flow instruction.
	parentInstr, index, err := findInstructionIndexOnControlFlowInstruction(block)
	if err != nil {
		return err
	}

	// Add the value to the parent instruction.
	parentInstr.values = addBlocks(parentInstr, parentInstr.values, newBlocks, index+offset)
	return nil
}

// AddCodeToControlFlowIndexFn adds code blocks to some control flow instruction
// at some index set with a function.
func AddCodeToControlFlowIndexFn(block Block, code string, offset func([]Block, int) int) error {
	// Parse new code blocks.
	newBlockEl := NewCodeParser(code).parse()

	return AddBlocksToControlFlowIndexFn(block, newBlockEl.blocks, offset)
}

// AddBlocksToControlFlowIndexFn adds code blocks to some control flow instruction
// at some index set with a function.
func AddBlocksToControlFlowIndexFn(block Block, newBlocks []Block, offset func([]Block, int) int) error {
	// Find the control flow instruction.
	parentInstr, index, err := findInstructionIndexOnControlFlowInstruction(block)
	if err != nil {
		return err
	}

	// Add the value to the parent instruction.
	parentInstr.values = addBlocks(parentInstr, parentInstr.values, newBlocks, offset(parentInstr.values, index))
	return nil
}

// ReplaceBlocks replaces a set of blocks for some code.
func ReplaceBlocks(jpBlocks []*JoinPointBlock, code string) {
	if len(jpBlocks) == 0 {
		return
	}
	block := jpBlocks[0].block
	parentBlock := block.getParent()
	if parentBlock == nil {
		return
	}
	// Parse new code blocks.
	newBlockEl := NewCodeParser(code).parse()

	// Replace the parent child values.
	switch parent := parentBlock.(type) {
	case *element:
		parent.blocks = replaceBlocks(parentBlock, jpBlocks, parent.blocks, newBlockEl.blocks)
	case *Instruction:
		blockInstr, ok := block.(*Instruction)
		if parent.name != instructionModule || !ok || blockInstr.name != instructionFunction {
			parent.values = replaceBlocks(parentBlock, jpBlocks, parent.values, newBlockEl.blocks)
			return
		}
		_, firstInstrIndex := findFirstInstruction(blockInstr.values)
		if firstInstrIndex != -1 {
			blockInstr.values = blockInstr.values[:firstInstrIndex]
		}
		// Update new blocks parent.
		for _, newBlock := range newBlockEl.blocks {
			newBlock.setParent(blockInstr)
		}
		blockInstr.values = append(blockInstr.values, newBlockEl.blocks...)
	default:
		return
	}
}

// RearrangeBlocks compresses a list of blocks by joining closed blocks.
func RearrangeBlocks(jpBlocks []*JoinPointBlock, maxInstrInARow int, countFn func(string) int) []*JoinPointBlock {
	if len(jpBlocks) == 0 {
		return jpBlocks
	}
	var res []*JoinPointBlock
	jpBlocksLen, jpBlocksAccum := len(jpBlocks), []*JoinPointBlock{jpBlocks[0]}
	for i := 1; i < jpBlocksLen; i++ {
		jpBlocksAccumLen := len(jpBlocksAccum)
		if jpBlocks[i].depth == jpBlocksAccum[jpBlocksAccumLen-1].depth && jpBlocksAccumLen < maxInstrInARow {
			jpBlocksAccum = append(jpBlocksAccum, jpBlocks[i])
			continue
		}
		res = append(res, resolveRearrangedBlock(jpBlocksAccum, jpBlocksAccumLen))
		jpBlocksAccum = []*JoinPointBlock{jpBlocks[i]}
	}
	res = append(res, resolveRearrangedBlock(jpBlocksAccum, len(jpBlocksAccum)))
	return res
}

// resolveRearrangedBlock returns the starting block for the blocks group,
// and removes the other blocks.
func resolveRearrangedBlock(jpBlocks []*JoinPointBlock, blocksLen int) *JoinPointBlock {
	res := jpBlocks[0]
	if blocksLen == 1 {
		return res
	}
	parentBlock := res.block.getParent()
	if parentBlock == nil {
		return res
	}
	parent, ok := parentBlock.(*Instruction)
	if !ok {
		return res
	}
	err := parent.removeChildren(jpBlocks[1].block, len(jpBlocks)-1)
	if err != nil {
		logrus.Errorf("rearranging block group: removing instructions from the group: %v", err)
	}
	return res
}

// replaceBlockWithCode replaces a block for some code.
func replaceBlockWithCode(block Block, code string) error {
	// Parse new code blocks.
	newBlockEl := NewCodeParser(code).parse()
	return replaceBlock(block, newBlockEl.blocks)
}

// replaceBlock replaces a block for some code blocks.
func replaceBlock(block Block, newCodeBlocks []Block) error {
	parentBlock := block.getParent()
	if parentBlock == nil {
		return errors.New("could not replace block: parent is null")
	}

	switch parent := parentBlock.(type) {
	case *Instruction:
		break
	case *element:
		// Replace the parent child values.
		if err := parent.replaceChild(block, newCodeBlocks); err != nil {
			logrus.Errorf("replacing block: %v", err)
		}
		return nil
	default:
		return errors.New("could not replace block: parent is not an instruction neither an element")
	}

	// Replace the parent child values.
	parent := parentBlock.(*Instruction)
	blockInstr, ok := block.(*Instruction)
	if parent.name != instructionModule || !ok || blockInstr.name != instructionFunction {
		if err := parent.replaceChild(block, newCodeBlocks); err != nil {
			logrus.Errorf("replacing block: %v", err)
		}
		return nil
	}
	_, firstInstrIndex := findFirstInstruction(blockInstr.values)
	if firstInstrIndex != -1 {
		blockInstr.values = blockInstr.values[:firstInstrIndex]
	}
	// Update new blocks parent.
	for _, newBlock := range newCodeBlocks {
		newBlock.setParent(blockInstr)
	}
	blockInstr.values = append(blockInstr.values, newCodeBlocks...)
	return nil
}

// findFirstInstruction finds the first instruction type block.
func findFirstInstruction(values []Block) (Block, int) {
	for i, child := range values {
		switch child.(type) {
		case *text, *Instruction:
			// Empty by design.
		default:
			return child, i
		}
		childInstr, ok := child.(*Instruction)
		if !ok {
			continue
		}
		switch childInstr.name {
		case instructionType, instructionParam, instructionResult, instructionLocal:
		// Empty by design.
		default:
			return child, i
		}
	}
	return nil, -1
}

// instructionOrder return the order weight for some code block.
// uses the instruction name.
func instructionOrder(block Block) int {
	if instr, ok := block.(*Instruction); ok {
		return instructionNameOrder(instr.name)
	}
	return len(instructionsOrderValues)
}

// instructionNameOrder return the order weight for some code block.
func instructionNameOrder(instr string) int {
	if i, ok := instructionsOrderValues[instr]; ok {
		return i
	}
	return len(instructionsOrderValues)
}

// isCodeInstruction returns if the code block is of type code instruction.
func isCodeInstruction(block Block) bool {
	blockInstr, ok := block.(*Instruction)
	if !ok {
		return false
	}
	switch blockInstr.name {
	case instructionType, instructionParam, instructionResult, instructionLocal:
		return false
	default:
		return true
	}
}

// addBlocks adds new blocks to the current blocks list.
func addBlocks(parent Block, curBlocks []Block, newBlocks []Block, index int) []Block {
	// Update new blocks parent.
	for _, newBlock := range newBlocks {
		newBlock.setParent(parent)
	}

	// Clone current blocks.
	newCurBlocks := append([]Block{}, curBlocks...)

	// Add the new blocks to the filtered blocks.
	if index >= len(curBlocks) {
		return append(newCurBlocks, newBlocks...)
	}
	var blocksWithAdded []Block
	blocksWithAdded = append(blocksWithAdded, newCurBlocks[:index]...)
	blocksWithAdded = append(blocksWithAdded, newBlocks...)
	blocksWithAdded = append(blocksWithAdded, newCurBlocks[index:]...)
	return blocksWithAdded
}

// replaceBlocks replaces a set of blocks for some new code.
func replaceBlocks(parent Block, jpBlocks []*JoinPointBlock, curBlocks []Block, newBlocks []Block) []Block {
	// Find the indexes to be removed.
	var indexesToRemove []int
	var jpIndex int
	for i, b := range curBlocks {
		if b != jpBlocks[jpIndex].block {
			continue
		}
		indexesToRemove = append(indexesToRemove, i)
		jpIndex++
		if jpIndex >= len(jpBlocks) {
			break
		}
	}

	// If no indexes to remove, return the actual blocks.
	if len(indexesToRemove) == 0 {
		return curBlocks
	}

	// Filter the blocks by removing the found ones.
	var filteredBlocks []Block
	filteredBlocks = append(filteredBlocks, curBlocks[:indexesToRemove[0]]...)
	for i := 1; i < len(indexesToRemove)-1; i++ {
		if indexesToRemove[i] > indexesToRemove[i+1] {
			filteredBlocks = append(filteredBlocks, curBlocks[indexesToRemove[i+1]:indexesToRemove[i]]...)
		}
	}
	if lastIndex := indexesToRemove[len(indexesToRemove)-1]; lastIndex+1 < len(curBlocks) {
		filteredBlocks = append(filteredBlocks, curBlocks[lastIndex+1:]...)
	}

	// If no new blocks to add, return the filtered blocks.
	if len(newBlocks) == 0 {
		return filteredBlocks
	}

	// Update new blocks parent.
	for _, newBlock := range newBlocks {
		newBlock.setParent(parent)
	}

	// Add the new blocks to the filtered blocks.
	if indexesToRemove[0] >= len(filteredBlocks) {
		return append(filteredBlocks, newBlocks...)
	}
	var blocksWithAdded []Block
	blocksWithAdded = append(blocksWithAdded, filteredBlocks[:indexesToRemove[0]]...)
	blocksWithAdded = append(blocksWithAdded, newBlocks...)
	blocksWithAdded = append(blocksWithAdded, filteredBlocks[indexesToRemove[0]:]...)
	return blocksWithAdded
}

// findInstructionIndexOnFunction finds the entry level instruction index on some function.
func findInstructionIndexOnFunction(block Block) (*Instruction, int, error) {
	// Find the function instruction.
	parentInstr, childBlock, err := findInstructionOnFunction(block)
	if err != nil {
		return nil, -1, err
	}

	// Find the instruction index to add.
	index := parentInstr.childIndex(childBlock)
	if index == -1 {
		return nil, -1, errors.New("code index to add the blocks not found on function values")
	}
	return parentInstr, index, nil
}

// findParentReturn returns the parent intruction of type return.
func findParentReturn(block Block) (*Instruction, *Instruction, int) {
	var parentInstr *Instruction
	childBlock := block
	for {
		parent := childBlock.getParent()
		if parent == nil {
			return nil, nil, -1
		}
		parentInstrAux, ok := parent.(*Instruction)
		if !ok {
			childBlock = parent
			continue
		}
		if parentInstrAux.name == instructionFunction {
			return nil, nil, -1
		}
		if parentInstrAux.name == instructionCodeReturn {
			parentInstr = parentInstrAux
			break
		}
		childBlock = parent
	}
	parentParent := parentInstr.getParent()
	if parentParent == nil {
		return nil, nil, -1
	}
	parentParentInstr, ok := parentParent.(*Instruction)
	if !ok {
		return nil, nil, -1
	}
	return parentParentInstr, parentInstr, parentParentInstr.childIndex(parentInstr)
}

// findInstructionIndexOnFunction finds the entry level instruction index on some function.
func findInstructionIndexOnControlFlowInstruction(block Block) (*Instruction, int, error) {
	// Find the function instruction.
	parentInstr, childBlock, err := findInstructionOnControlFlowInstruction(block)
	if err != nil {
		return nil, -1, err
	}

	// Find the instruction index to add.
	index := parentInstr.childIndex(childBlock)
	if index == -1 {
		return nil, -1, errors.New("code index to add the blocks not found on control flow values")
	}
	return parentInstr, index, nil
}

// findInstructionOnFunction finds the entry level instruction on some function.
func findInstructionOnFunction(block Block) (*Instruction, Block, error) {
	var parentInstr *Instruction
	childBlock := block
	for {
		parent := childBlock.getParent()
		if parent == nil {
			return nil, nil, errors.New("parent of type function not found")
		}
		parentInstrAux, ok := parent.(*Instruction)
		if !ok || parentInstrAux.name != instructionFunction {
			childBlock = parent
			continue
		}
		parentInstr = parentInstrAux
		break
	}
	return parentInstr, childBlock, nil
}

// findInstructionOnFunction finds the entry level instruction on some function.
func findInstructionOnControlFlowInstruction(block Block) (*Instruction, Block, error) {
	var parentInstr *Instruction
	childBlock := block
	for {
		parent := childBlock.getParent()
		if parent == nil {
			return nil, nil, errors.New("parent of type function not found")
		}
		parentInstrAux, ok := parent.(*Instruction)
		if !ok || !wlang.IsControlFlow(parentInstrAux.name) && parentInstrAux.name != instructionFunction {
			childBlock = parent
			continue
		}
		parentInstr = parentInstrAux
		break
	}
	return parentInstr, childBlock, nil
}

// deleteBlock deletes a block from a list of blocks.
func deleteBlock(blocks []Block, block Block) []Block {
	for i, b := range blocks {
		if b == block {
			if i == len(blocks)-1 {
				return blocks[:i]
			}
			return append(append([]Block{}, blocks[:i]...), blocks[i+1:]...)
		}
	}
	return blocks
}
