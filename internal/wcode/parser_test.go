package wcode

import (
	"fmt"
	"testing"

	"joao/wasm-manipulator/internal/wtemplate"
)

func TestStructureParser_Easy(t *testing.T) {
	p := &CodeParser{
		code: "(local.get 0 (local.tee 2 %kkk%))",
	}
	fmt.Println(p.Parse())
}

func TestStructureParser_Offset(t *testing.T) {
	p := &CodeParser{
		code: `(func (local.set %width%
			(i32.shl
			 (i32.mul
			  (local.get %width%)
			   (local.get %height%))
			 (i32.const 2)))
		   (loop $L0
			(if $I1
			 (i32.gt_s
			  (local.get %width%)
			  (local.get %i%))
			 (then
			  (i32.store8
			   (local.get %i%)
			   (call %color_sin%
				 (i32.load8_u (local.get %i%))
			   ))
			  (i32.store8 offset=1
			   (local.get %i%)
			   (call %color_sin%
				 (i32.load8_u (i32.add (local.get %i%) (i32.const 1)))
			   ))
			  (i32.store8 offset=2
			   (local.get %i%)
			   (call %color_sin%
				 (i32.load8_u (i32.add (local.get %i%) (i32.const 2)))
			   ))
			  (local.set %i%
			   (i32.add
				(local.get %i%)
				(i32.const 4)))
			  (br $L0)))))`,
	}
	res := p.parse()
	fmt.Println(res.StringIndent(""))
}

func TestStructureParser_Hard(t *testing.T) {
	p := &CodeParser{
		code: wtemplate.ClearString(longCode),
	}
	fmt.Println(p.Parse())
}

func TestStructureParser_Evaluation(t *testing.T) {
	p := &CodeParser{
		code: wtemplate.ClearString(evaluationCode),
	}
	res := p.parse()
	fmt.Println(res.String())
}

var longCode = `(func $e (type $t0)
    (local $l1 i64)
    (local.set $l1 (i64.const 100000))
    (loop $L1
      (br_if $L1
        (i64.ge_s
          (local.tee $l1
            (i64.add
              (local.get $l1)
              (i64.const -1)
            )
          )
          (i64.const 0)
        )
      )
    )
  )`

var evaluationCode = "/p1 + `casa do ${povo} - ` + miranda + 'do corvo'/"
