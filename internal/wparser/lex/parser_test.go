package lex

import (
	"fmt"
	"log"
	"os"
	"testing"

	"joao/wasm-manipulator/internal/wkeyword"
	"joao/wasm-manipulator/internal/wparser/template"
	"joao/wasm-manipulator/internal/wtemplate"
)

const (
	t1   = "(i32.add %a1:includes_one(t2, t3):defines(b1, b2)% %a2%)"
	t2   = "(local.tee %b1% (i32.add %b2%))"
	t3   = "(local.set %b1% (local.get %b2%) (i32.const ?))"
	code = `(func
    (i32.add
	  (local.tee 1 (i32.add (local.set 1 (i32.add (local.get 0))) (i32.const 1)))
	  (local.get 0)
	)
)`
	moduleCode = `(module
  (type $t0 (func (param f32) (result f32)))
  (func $calculate (type $t0) (param $p0 f32) (result f32)
    (f32.mul
      (f32.const 2)
      (local.get $p0)))
  (export "calculate" (func $calculate)))`
)

type ObjectMap map[string]interface{}

type ObjectMapInit struct {
	k string
	v interface{}
}

func newObjectMap(startValues ...ObjectMapInit) ObjectMap {
	res := make(ObjectMap)
	for _, startValue := range startValues {
		res[startValue.k] = startValue.v
	}
	return res
}

func (om ObjectMap) Is(k string) wkeyword.KeywordType {
	if _, ok := om[k]; ok {
		return wkeyword.KeywordTypeObject
	}
	return wkeyword.KeywordTypeUnknown
}

func (om ObjectMap) Get(k string) (interface{}, wkeyword.KeywordType, bool) {
	if v, ok := om[k]; ok {
		return wkeyword.NewKwObject(v), wkeyword.KeywordTypeObject, true
	}
	return nil, wkeyword.KeywordTypeUnknown, false
}

func setupTemplateTest(t *testing.T) (map[string]*wtemplate.Template, map[string]*wtemplate.TemplateContext, map[string][]*wtemplate.SearchValue) {
	templatesMap := make(map[string]*wtemplate.Template)
	for _, vals := range [][]string{
		{"t1", t1},
		{"t2", t2},
		{"t3", t3},
	} {
		t, err := template.Parse(vals[0], vals[1])
		if err != nil {
			log.Fatal(err)
		}
		templatesMap[vals[0]] = t
	}
	templatesContextMap := make(map[string]*wtemplate.TemplateContext)
	searchResultsMap := make(map[string][]*wtemplate.SearchValue)

	// Get templates context.
	for _, name := range []string{"t2", "t3", "t1"} {
		template := templatesMap[name]
		templateCtx, err := wtemplate.NewTemplateContext(template, templatesMap)
		if err != nil {
			t.Errorf("creating template context for template %q: %v", template.Key, err)
		}
		templatesContextMap[name] = templateCtx
	}

	// Using the template context, build comby result.
	search, err := templatesContextMap["t1"].Search(code)
	if err != nil {
		t.Error(search)
	}
	searchResultsMap["t1"] = search

	return templatesMap, templatesContextMap, searchResultsMap
}

func TestEmitterReceiver(t *testing.T) {
	t.Parallel()
	//logrus.SetLevel(logrus.InfoLevel)

	err := os.Setenv("PATH", fmt.Sprintf("/home/joao/go/src/wasm-manipulator/dependencies/wabt:/home/joao/go/src/wasm-manipulator/dependencies/minifyjs/bin:/home/joao/go/src/wasm-manipulator/dependencies/comby:%s", os.Getenv("PATH")))
	if err != nil {
		t.Error(err)
	}

	values := []string{
		`(local.set $wmr_l5 (i32.mul (i32.mul (local.get $width) (local.get $height)) (i32.const 4))) (call $wmr_f8 /!!cachePtr["sepia"]/) (block $B0 (br_if $B0 (i32.ne /!!cachePtr["sepia"]/ (i32.const 0))) (global.set #cachePtr /last+cachePtrAccum/)) (local.set #cachePtrVal /cachePtr["sepia"]/) (call $wmr_f9 (i32.const 0)) (local.set width (i32.shl (i32.mul (local.get width) (local.get height)) (i32.const 2))) (loop $L0 (if $I1 (i32.gt_s (local.get width) (local.get i)) (then (local.set r (i32.load8_u (local.get i))) (local.set g (i32.load8_u offset=1 (local.get i))) (local.set b (i32.load8_u offset=2 (local.get i))) (i32.store8 (local.get i) (i32.trunc_f32_u (f32.min (f32.add (f32.add (f32.mul (f32.convert_i32_s (local.get r)) (f32.const 0.393)) (f32.mul (f32.convert_i32_s (local.get g)) (f32.const 0.769))) (f32.mul (f32.convert_i32_s (local.get b)) (f32.const 0.189))) (f32.const 255)))) (local.set $wmr_l6 (i32.load8_u (i32.add (local.get i) (i32.const 1)))) (local.set $wmr_l7 (i32.load8_u (i32.add (local.get $wmr_l4) (local.get $wmr_l6)))) (call $wmr_f8 /tmp/) (if (i32.eqz (local.tee $wmr_l8 (i32.load8_u (i32.add (local.get $wmr_l4) (local.get $wmr_l6))))) (then (local.set $wmr_l8 (i32.trunc_f32_u (f32.min (f32.add (f32.add (f32.mul (f32.convert_i32_s (local.get r)) (f32.const 0.349)) (f32.mul (f32.convert_i32_s (local.get g)) (f32.const 0.686))) (f32.mul (f32.convert_i32_s (local.get b)) (f32.const 0.168))) (f32.const 255)))) (call $wmr_f8 /"no: "+res/) (i32.store8 (i32.add (local.get $wmr_l4) (local.get $wmr_l6)) (local.get $wmr_l8))) (else (call $wmr_f8 /"yes: "+res/))) (i32.store8 offset=1 (local.get i) (local.get $wmr_l8)) (local.set $wmr_l6 (i32.load8_u (i32.add (local.get i) (i32.const 2)))) (local.set $wmr_l7 (i32.load8_u (i32.add (local.get $wmr_l4) (local.get $wmr_l6)))) (call $wmr_f8 /tmp/) (if (i32.eqz (local.tee $wmr_l8 (i32.load8_u (i32.add (local.get $wmr_l4) (local.get $wmr_l6))))) (then (local.set $wmr_l8 (i32.trunc_f32_u (f32.min (f32.add (f32.add (f32.mul (f32.convert_i32_s (local.get r)) (f32.const 0.272)) (f32.mul (f32.convert_i32_s (local.get g)) (f32.const 0.534))) (f32.mul (f32.convert_i32_s (local.get b)) (f32.const 0.131))) (f32.const 255)))) (call $wmr_f8 /"no: "+res/) (i32.store8 (i32.add (local.get $wmr_l4) (local.get $wmr_l6)) (local.get $wmr_l8))) (else (call $wmr_f8 /"yes: "+res/))) (i32.store8 offset=2 (local.get i) (local.get $wmr_l8)) (local.set i (i32.add (local.get i) (i32.const 4))) (br $L0)))) (call $wmr_f9 (i32.const 1)) (return)`,
		"casa %'abc':map((x) => '123'; '-456: '; x)% povo",
		"%print:map((x) => x;' %bola%')%",
		"casa %'abc':map((x) => '123';'-456: ';x):repeat(2)% povo",
		"casa %'abc':map((x) => '123';'-456: ';x):repeat(2):join(' - ')% povo",
		"casa %'abc':map((x) => '123';'-456: ';x):repeat(2):join(' - '):count()% povo",
		"casa %'abc':map((x) => '123';'-456: ';x):repeat(2):count()% povo",
		"casa %'abc':map((x) => '123';'-456: ';x):repeat(2):join(' - '):contains('456')% povo",
		"casa %'abc':map((x) => '123';'-456: ';x):repeat(2):contains('123-456: abc')% povo",
		"casa %'abc':map((x) => '123';'-456: ';x):repeat(2):contains('456')% povo",
		"casa %'abc':map((x) => '123';'-456: ';x):repeat(2):contains('123';'-456: abc')% povo",
		"casa %'abc':map((x) => '123';'-456: ';x):repeat(2):contains('123-456: %var%')% povo",
		"%code%",
		"casa \"%t1:remove(a1);'\" ';print% passa %bola% %bola2:count()%",
		"casa \"%t1:remove(a2)%\" passa %bola% %bola2:count()%",
		"casa \"%t1:select(a2)%\" passa %bola% %bola2:count()%",
		"%bola:remove('cr')% - %'bola':remove('cr')% - %'cr 79':remove('cr 7')% - %'cr79':remove(cr7)%",
		"%t1:replace(a1, a2)% - %t1:replace(a1, '123')% - %t1:replace('0', '123')% - %t1:replace('0', a2)%",
		"%t1:replace(a1, '123'):remove(a2):select(a1):map((x) => '`';x;'`'):string():count():type():map((x) => ''):type()%",
		"%print% %!print% %!!print% %!!!print%",
		"%!!(var:count())% %!(var:count())% %!var:count()% %var:count()% %'':map((_) => !var && var)% %'':map((_) => !var || var)% %'':map((_) => !var || var && !var)%",
		"%'':map((_) => !var;var && !var;!!var;var || !var)%",
		"%'casa':map((x) => x == 'casa')% %var:map((x) => x == 'variable')% %'variable':map((x) => x == var)% %'variable':map((x) => x != var)% %var:map((x) => x != x && !!x || x == x && !x)%",
		"%var:count():map((x) => var;': ';x;': ';!(x:map((x) => !x):string()))%",
		"%var:split('a'):join('|')% %var:split(''):filter((l) => l != 'a'):join('|')%",
		"%'123'[1]% %'123':split('3')[0]% %var:split('3')[0]%",
		"%'joao paulo':slice(1, 6)% %'joao paulo rodrigues':split(' '):slice(1)% %'joao paulo rodrigues':split(' ')[1]% %var:slice(2)%",
		"%'joao paulo':splice(1, 6)% %'joao paulo':split(' '):splice(1)% %var:splice(2)%",
		"%'':map((_) => 24.5 * 2);'; ';'':map((_) => 2 << 2)%",
		"%call.kkk:join(' ')%",
		"%call:contains('kkk')%",
		"%var:assert((x) => x == 'variable')%",
		"%'123':assert((v) => v - 1 == 122):count()%",
		"%'123':assert((v) => v - 1 != 122):count()%",
		"%call.kkk:remove('777'):remove('123'):remove('567'):remove('987')%",
		"%call:remove('kkk'):remove('123'):remove('yyy'):remove('987')%",
		"%call.kkk:replace('123', '456'):replace('456','777')%",
		"%call:replace('kkk','568')%",
		"%'147':filter((v)=>v%2!=0)%",
		"%'147':split(''):filter((v)=>v%2==0)%",
		"%'147':splice(1)%",
		"%'147':split(''):splice(0,1)%",
		"%'147':slice(1)%",
		"%'147':split(''):slice(0,1)%",
	}
	for _, v := range values {
		func(code string, v string) {
			t.Run(fmt.Sprintf("Test Emitter Receiver %q", v), func(t *testing.T) {
				t.Parallel()

				var templatesMap map[string]*wtemplate.Template
				var templatesContextMap map[string]*wtemplate.TemplateContext
				var searchResultsMap map[string][]*wtemplate.SearchValue

				templatesMap, templatesContextMap, searchResultsMap = setupTemplateTest(t)

				res := parse(v, newParsingContext(nil, []wkeyword.KeywordsMap{
					wkeyword.NewTemplateResults(
						&templatesContextMap,
						&templatesMap,
						&searchResultsMap,
					),
					wkeyword.NewStringValuesMap(
						[]string{"print", "hello"},
						[]string{"var", "variable"},
						[]string{"bola", "cr7"},
						[]string{"bola2", "cr9"},
						[]string{"code", code},
					),
					newObjectMap(
						ObjectMapInit{
							k: "call",
							v: struct {
								Kkk []string
								Yyy int
							}{
								Kkk: []string{"123", "987"},
								Yyy: 456,
							},
						},
					),
				}))
				fmt.Printf("\n Input: %q\nOutput: %q\n\n", v, res.Output)
			})
		}(moduleCode, v)
	}
}
