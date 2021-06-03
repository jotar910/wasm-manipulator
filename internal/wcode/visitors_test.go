package wcode

import (
	"fmt"
	"testing"

	"joao/wasm-manipulator/internal/wtemplate"
)

func TestVisitors_FunctionContext(t *testing.T) {
	parser := &CodeParser{
		code: wtemplate.ClearString(longFunctionVisitorCode),
	}
	entryBlock := parser.parse()
	moduleCtx := newModuleContext(entryBlock)
	typeVisitor := newTypeContextVisitor(moduleCtx)
	fnVisitor := newFunctionContextVisitor(moduleCtx)
	exportedFnVisitor := newExportedContextVisitor(moduleCtx)
	entryBlock.Traverse(typeVisitor)
	entryBlock.Traverse(fnVisitor)
	entryBlock.Traverse(exportedFnVisitor)
	fmt.Printf("%+v\n%+v\n%+v\n", moduleCtx, moduleCtx.functions, moduleCtx.types)
}

func TestVisitors_GlobalContext(t *testing.T) {
	parser := &CodeParser{
		code: wtemplate.ClearString(longGlobalVisitorCode),
	}
	entryBlock := parser.parse()
	moduleCtx := newModuleContext(entryBlock)
	globalVisitor := newGlobalContextVisitor(moduleCtx)
	exportedFnVisitor := newExportedContextVisitor(moduleCtx)
	entryBlock.Traverse(globalVisitor)
	entryBlock.Traverse(exportedFnVisitor)
	fmt.Printf("%+v\n%+v\n", moduleCtx, moduleCtx.globals)
}

var longFunctionVisitorCode = `
(module
	(type $t1 (func (param i32) (result i32)))
	(import "env" "f2" (func $f2 (type $t1)))
	(export "f2" (func $f1))
	(func $f1 (type $t1) (param $p0 i32) (param $p1 i64) (result i32) (local $l2 i64) (local $l3 i64) (local.get 0 (local.tee 2 %kkk%)))
)
`

var longGlobalVisitorCode = `
(module
	(import "js" "g1" (global $js.g i32))
	(import "js" "g2" (global $js.g_1 (mut i32)))
	(global $jsg (mut i32) (i32.const 0))
	(global $jsg2 (mut i32) (i32.const 0))
	(global $g4 (mut i32) (i32.const 0))
	(global $g5 i32 (i32.const 0))
	(global $g6 (mut i32) (global.get $js.g))
	(global $g7 i32)
	(export "jsg" (global 2))
	(export "jsg2" (global 3))
)
`
