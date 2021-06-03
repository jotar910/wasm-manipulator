package pointcut

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer/stateful"
)

// ExprWithContext is the model for the pointcut expression with context in arguments.
type ExprWithContext struct {
	Args   []ArgumentWithContext `"(" ( @@ ("," @@)* )? ")"`
	Instrs []Instruction         ` "="">" @@+`
}

// ExprWithoutContext is the model for the pointcut expression with no context in arguments..
type ExprWithoutContext struct {
	Args   []ArgumentWithoutContext `"(" ( @@ ("," @@)* )? ")"`
	Instrs []Instruction            ` "="">" @@ ( @@ )*`
}

// argumentVarIndex is the argument index type.
type argumentVarIndex string

// Capture captures the input value to an argument index type.
func (a *argumentVarIndex) Capture(values []string) error {
	value := values[0]
	if value == "?" {
		*a = ""
		return nil
	}
	*a = argumentVarIndex(value)
	return nil
}

// ArgumentWithContext is the pointcut argument model with execution context.
type ArgumentWithContext struct {
	Type     string           `( @( WasmType ) "." )?`
	VarType  string           `@WasmLocalType`
	VarIndex argumentVarIndex `"[" @(Index | Number | Identifier) "]"` // Todo: include "?"
	Name     string           ` @Identifier`
}

// ArgumentWithoutContext is the pointcut argument model with no execution context.
type ArgumentWithoutContext struct {
	Type string `( @( WasmType ) )?`
	Name string `@Identifier`
}

// Instruction is the pointcut instruction model.
type Instruction struct {
	GroupBlock []Instruction `( "(" @@ (@@)* ")" )?`
	AndBlock   *string       `( @("&""&") )?`
	OrBlock    *string       `( @("|""|") )?`
	Block      *Method       `( @@ )?`
}

// Method represents a pointcut method.
type Method struct {
	Func    *funcMethod     `( @@ )`
	Call    *callMethod     `| ( @@ )`
	Args    *argMethod      `| ( @@ )`
	Returns *returnsMethod  `| ( @@ )`
	Templ   *templateMethod `| ( @@ )`
	Other   *OtherMethod    `| ( @@ )`
}

// funcMethod is the pointcut method func.
type funcMethod struct {
	Name  string          `@( "func" )`
	Input *FuncDefinition `( "(" @@ ")" )`
}

// callMethod is the pointcut method call.
type callMethod struct {
	Name  string          `@( "call" )`
	Input *FuncDefinition `( "(" @@ ")" )`
}

// argMethod is the pointcut method arg.
type argMethod struct {
	Name  string          `@( "args" )`
	Input *ArgMethodInput `( "(" @@ ")" )`
}

// returnsMethod is the pointcut method returns.
type returnsMethod struct {
	Name  string              `@( "returns" )`
	Input *returnsMethodInput `( "(" @@ ")" )`
}

// templateMethod is the pointcut method template.
type templateMethod struct {
	Name  string               `@( "template" )`
	Input *templateMethodInput `( "(" @@ ")" )`
}

// OtherMethod represents any other method that is not known.
// it can be a valid pointcut, declared in the pointcuts or it can be some invalid one.
// in case of an invalid pointcut, an error must be dispatched.
type OtherMethod struct {
	Name      string   `@Identifier`
	Arguments []string `( "(" ( @Identifier ( "," @Identifier )* )? ")" )`
}

// Boolean consists on a boolean that implements the Capture for the parser on participle.
type Boolean bool

// Capture captures the input value to a boolean.
func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true"
	return nil
}

// FuncDefinition is the definition for func method.
type FuncDefinition struct {
	ReturnType *FuncDefinitionReturn `@@`
	Name       *FuncDefinitionName   ` @@`
	Params     *FuncDefinitionParams `"(" @@? ")"`
	Scope      FunctionScope         `("," @("imported" | "exported" | "internal" | "start"))?`
}

// FunctionScope consists on a function scope type that implements the Capture for the parser on participle.
type FunctionScope int

const (
	FunctionScopeAny FunctionScope = iota
	FunctionScopeInternal
	FunctionScopeImported
	FunctionScopeExported
	FunctionScopeStart
)

// Capture captures the input value to a FunctionScope.
func (s *FunctionScope) Capture(values []string) error {
	switch {
	case len(values) == 0:
		*s = FunctionScopeAny
	case values[0] == "internal":
		*s = FunctionScopeInternal
	case values[0] == "imported":
		*s = FunctionScopeImported
	case values[0] == "exported":
		*s = FunctionScopeExported
	case values[0] == "start":
		*s = FunctionScopeStart
	default:
		*s = FunctionScopeAny
	}
	return nil
}

// AnyTermBoolean is the boolean type for any term types.
type AnyTermBoolean bool

// Capture captures the input value into any term type.
func (b *AnyTermBoolean) Capture(values []string) error {
	*b = len(values) == 1 && values[0] == "*"
	return nil
}

// AnyArrayBoolean is the boolean type for any array types.
type AnyArrayBoolean bool

// Capture captures the input value into any array type.
func (b *AnyArrayBoolean) Capture(values []string) error {
	*b = len(values) == 2 && values[0] == "." && values[1] == "."
	return nil
}

// funcVariableType is the model for function variables.
type funcVariableType struct {
	Variable string `@Identifier ":"`
	Type     string `@( WasmType )`
}

// regex represents a regex input value.
type regex string

// String prints regex indirect value.
func (r regex) String() string {
	return strings.Trim(string(r), "/")
}

// identifierRegex is the model for an identifier that matches a regex.
type identifierRegex struct {
	Variable string `@Identifier`
	Regex    *regex `( ":" @Regex )`
}

// identifierName is the model for an identifier that matches a name.
type identifierName struct {
	Variable string  `@Identifier`
	Name     *string `( ":" @Identifier )`
}

// identifierIndexName is the model for an identifier that matches an index name.
type identifierIndexName struct {
	Variable  string  `@Identifier`
	IndexName *string `( ":" @Index )`
}

// identifierIndex is the model for an identifier that matches an index value.
type identifierIndex struct {
	Variable string `@Identifier`
	Index    *int   `(":" "[" @Number "]")`
}

// FuncDefinitionReturn is the return definition for the func method.
type FuncDefinitionReturn struct {
	Any          AnyTermBoolean    `@( "*" )`
	Type         string            `| ( "void" | @( WasmType ) )`
	Variable     string            `| ( "%" @Identifier "%" )`
	VariableType *funcVariableType `| ( "%" @@ "%" )`
}

// FuncDefinitionName is the name definition for the func method.
type FuncDefinitionName struct {
	Any               AnyTermBoolean       `@( "*" )`
	Name              string               `| ( @Identifier )`
	IndexName         string               `| ( @Index )`
	Index             *int                 `| ( "[" @Number "]" )`
	Regex             regex                `| ( @Regex )`
	Variable          string               `| ( "%" @Identifier "%" )`
	VariableRegex     *identifierRegex     `| ( "%" @@ "%" )`
	VariableName      *identifierName      `| ( "%" @@ "%" )`
	VariableIndexName *identifierIndexName `| ( "%" @@ "%" )`
	VariableIndex     *identifierIndex     `| ( "%" @@ "%" )`
}

// funcDefinitionParamType is the parameter type definition for the func method.
type funcDefinitionParamType struct {
	Value string         `( @( WasmType ) )`
	Any   AnyTermBoolean `| @( "*" )`
}

// funcDefinitionParamName is the parameter name definition for the func method.
type funcDefinitionParamName struct {
	Any      AnyTermBoolean `@( "*" )`
	Variable string         `| ( "%" @Identifier "%" )`
}

// funcDefinitionParam is the parameter definition for the func method.
type funcDefinitionParam struct {
	Type *funcDefinitionParamType `@@`
	Name *funcDefinitionParamName `@@?`
}

// FuncDefinitionParams is the parameters definition for the func method.
type FuncDefinitionParams struct {
	Any    AnyArrayBoolean       `@( ".""." )`
	Params []funcDefinitionParam `| ( @@ ( "," @@ )* )`
}

// ArgMethodInput is the definition used on arg method.
type ArgMethodInput struct {
	Params []string `( @Identifier ( "," @Identifier )* )?`
}

// templateMethodInput represents the input value passed to the template method
type templateMethodInput struct {
	Template  string  `@Identifier`
	JustCheck Boolean `( "," @("true" | "false") )?`
}

// returnsMethodInput represents the input value passed to the returns method
type returnsMethodInput struct {
	Any   AnyTermBoolean `@( "*" )`
	Value string         `| ( "void" | @( WasmType ) )`
}

var (
	lexer = stateful.MustSimple([]stateful.Rule{
		{"String", `"(\\"|[^"])*"`, nil},
		{"Number", `(?:\d*\.)?\d+`, nil},
		{"Index", `\$[a-zA-Z][\w\d_]*`, nil},
		{"Regex", `/(\\/|[^/])*/`, nil},
		{"WasmType", `(i32|i64|f32|f64)`, nil},
		{"WasmLocalType", `(param|local)`, nil},

		{"Identifier", `[a-zA-Z][\w\d_]*`, nil},
		{"Punctuation", `[-[!@#$%^&*()+_={}\|:;"'<,>.?]|]`, nil},
		{"Whitespace", `[ \t\n\r]+`, nil},
	})
	parserWithContext = participle.MustBuild(&ExprWithContext{},
		participle.Lexer(lexer),
		participle.Elide("Whitespace"),
		participle.UseLookahead(2),
	)
	parserWithoutContext = participle.MustBuild(&ExprWithoutContext{},
		participle.Lexer(lexer),
		participle.Elide("Whitespace"),
		participle.UseLookahead(2),
	)
)

// ParseWithContext parses a pointcut input that can contain the context of the around environment,.
func ParseWithContext(value string) (*ExprWithContext, error) {
	ast := &ExprWithContext{}
	err := parserWithContext.ParseString("", value, ast)
	if err != nil {
		return nil, fmt.Errorf("parsing input expression (with context): %w", err)
	}
	return ast, nil
}

// ParseWithoutContext parses a pointcut input that has no knowledge of the around environment,.
func ParseWithoutContext(value string) (*ExprWithoutContext, error) {
	ast := &ExprWithoutContext{}
	err := parserWithoutContext.ParseString("", value, ast)
	if err != nil {
		return nil, fmt.Errorf("parsing input expression (without context): %w", err)
	}
	return ast, nil
}
