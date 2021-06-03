package variable

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer/stateful"
)

// Expr is the model for the variable expression.
type Expr struct {
	Type  []string `@(SimpleType | MapType | ArrayType)+`
	Value *Value   `("=" @@)?`
}

// GetValue returns the variable value as string.
func (expr *Expr) GetValue(def string) string {
	if expr.Value != nil {
		return expr.Value.GetValue(0, nil)
	}
	return def
}

// GetType returns the variable type.
func (expr *Expr) GetType() string {
	return strings.Join(expr.Type, "")
}

// Type represents the several type blocks that forms the variable type.
type Type []string

// Capture captures the input value to a variable type.
func (t *Type) Capture(values []string) error {
	newType := values[0]
	if len(*t) == 0 {
		*t = append(*t, newType)
		return nil
	}
	lastType := (*t)[len(*t)-1]
	if isSimpleType(lastType) {
		return errors.New("simple type cannot have subtypes")
	}
	return nil
}

// isSimpleType returns if a type is simple or not.
func isSimpleType(t string) bool {
	return regexp.MustCompile("^(i32|i64|f32|f64|string)$").MatchString(t)
}

// Value is the model for the variable value.
type Value struct {
	String      *string      `@String`
	Number      *float64     `| @Number`
	ArrayValues []ArrayValue `| "[" @@ ("," @@)* "]"`
	Empty       string       `| @ArrayType?`
}

// GetValue returns the variable value as string.
func (val *Value) GetValue(index int, parent *Value) string {
	if val.String != nil {
		if parent == nil {
			return strings.Trim(*val.String, `"`)
		}
		return *val.String
	}
	if val.Number != nil {
		value := strconv.FormatFloat(*val.Number, 'f', -1, 64)
		if value == "0" {
			return ""
		}
		return value
	}
	if len(val.ArrayValues) > 0 {
		var res []string
		for _, arrayValue := range val.ArrayValues {
			for _, v := range arrayValue.Values {
				res = append(res, strings.TrimSpace(v.GetValue(index, val)))
			}
		}
		return fmt.Sprintf("[%s]", strings.Join(res, ","))
	}
	return ""
}

// Value represents a key-value entry.
type ArrayValue struct {
	Values []Value `@@ ("," @@)*`
}

// KeyValue represents a key-value entry.
type KeyValue struct {
	Key   Value `"[" @@`
	Value Value `"," @@ "]"`
}

var (
	lexer = stateful.MustSimple([]stateful.Rule{
		{"String", `"(\\"|[^"])*"`, nil},
		{"SimpleType", `i32|i64|f32|f64|string`, nil},
		{"MapType", `map\[(i32|i64|f32|f64|string)\]`, nil},
		{"ArrayType", `\[\]`, nil},
		{"Number", `(?:\d*\.)?\d+`, nil},
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`, nil},
		{"Comment", `(?:#|//)[^\n]*\n?`, nil},
		{"Whitespace", `[ \t\n\r]+`, nil},
	})
	parser = participle.MustBuild(&Expr{},
		participle.Lexer(lexer),
		participle.Elide("Comment", "Whitespace"),
		participle.UseLookahead(2),
	)
)

// Parse parses the variable input.
func Parse(expr string) (*Expr, error) {
	ast := &Expr{}
	err := parser.ParseString("", expr, ast)
	if err != nil {
		return nil, fmt.Errorf("parsing input expression: %w", err)
	}
	return ast, nil
}
