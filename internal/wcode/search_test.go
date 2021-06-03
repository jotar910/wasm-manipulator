package wcode

import (
	"fmt"
	"testing"

	"joao/wasm-manipulator/internal/wtemplate"
)

func TestJoinPointSearch_RemoveDuplicates(t *testing.T) {
	p := &CodeParser{
		code: wtemplate.ClearString(longSearchCode),
	}
	fn := p.parse().blocks[0].(*Instruction)
	s := &JoinPointSearch{
		found: []*JoinPointBlock{
			{
				block:    fn.values[3].(*Instruction),
				function: fn,
				depth:    1,
			},
			{
				block:    fn.values[3].(*Instruction).values[1].(*Instruction),
				function: fn,
				depth:    2,
			},
		},
	}
	res := s.RemoveDuplicates()
	fmt.Println(res, fn.values[3].(*Instruction).values[0].String())
}

var longSearchCode = `(func $e (type $t0)
    (local $l1 i64)
    (local.set $l1 (i64.const 100000))
    (local.set $l2 (i64.const 100000))
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
