package lex

import (
	"fmt"
	"testing"
)

func TestLexer(t *testing.T) {
	vs := []string{
		"%abc.asd:met()[1].kkk[0].ok%",
		"%'abc'[1]%",
		"%print[1]%",
		"%print:method()[1]%",
		"%(print:method())[1]%",
		"%print:method1()[1]:method2()%",
		"%print:method('arg'[0])%",
		"%print:method((x) => x[0])%",
		"%'casa':map((x) => x == 'casa')%",
		"%print:map((print) => !print && (!(print:count()) || !!print); !print && print)%",
		"%!(print:count())%",
		"%var:count():map((x) => var+': '+x+': '+!(x:map((x) => !x)))%",
		"%a:replace(b, \"(call %x% (i32.const 32))\" + 1):repeat(200)%",
		"%a:replace(b, '(call %x% (i32.const 32))' + 1):repeat(200)%",
		"casa %do:remove(`\npovo\n`):repeat(200, '123%print+\" o que?\"% l' +   '234' + var, var,var)% passa %bola% %bola2:x()%",
		"casa %do:remove(`\npovo\n`):repeat(200, '123%print+\" o que?\"% l' +   '234' + var, var,var) + ' ' + print% passa %bola% %bola2:x()%",
		"%do:remove(123) + 'ola'%",
		"%t1:replace(a1, a2)%",
		`(block %root:select(block_label)%
          (local.set %rest%
              (i64.rem_s
                  (local.get %root:select(index)%)
                  (i64.const 2)
              )
          )
          (local.set %root:select(index)%
              (i64.sub
                  (local.get %root:select(index)%)
                  (local.get %rest%)
              )
          )
          %root:select(initial_condition)%
          (loop %root:select(label)%
              %root:select(body)%
              %root:select(body)%
              (;call %debug% (local.get %p0%);)
              (br_if %root:select(label)%
                (i32.eqz
                    (i64.eqz
                        (local.tee %root:select(index)%
                            (i64.add
                                (local.get %root:select(index)%)
                                (i64.const -2)
                            )
                        )
                    )
                )
              )
          )
          (block $extra_b
              (br_if $extra_b
                  (i64.le_s
                      (local.get %rest%)
                      (i64.const 0)
                  )
              )
              (loop $extra_l
                  %root:select(body)%
                  (br_if $extra_l
                    (i32.eqz
                        (i64.eqz
                            (local.tee %rest%
                                (i64.add
                                    (local.get %rest%)
                                    (i64.const -1)
                                )
                            )
                        )
                    )
                  )
              )
          )
          %root:select(after_loop)%
        )`,
	}
	for i, v := range vs {
		_, ch := Lex("lexer", v)
		fmt.Printf("Test %d\n", i+1)
		for item := range ch {
			fmt.Printf("{%s\t%q}\n", item.t, item.v)
		}
		fmt.Println("-----------")
		fmt.Println()
	}
}
