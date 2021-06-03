package wtemplate

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var variableRegex = regexp.MustCompile(`%[^%]+%`)
var unknownValueRegex = regexp.MustCompile(`[^\\]\?`)
var fixQuestionMarkRegex = regexp.MustCompile(`\\\?`)

var ErrNotFound error = errors.New("search not found")

var ErrNotMatch error = errors.New("template not match")

const (
	FormatterVariable = ":[%s]"
)

// createTemplateContext creates a new template context for some template.
// the template context may be cached in memory.
func createTemplateContext(ctx *TemplateContext, key string) (*TemplateContext, error) {
	t, ok := ctx.TemplatesMap[key]
	if !ok {
		return nil, fmt.Errorf("template %q not found", key)
	}
	ctx, err := newTemplateContext(t, ctx)
	if err != nil {
		return nil, fmt.Errorf("creating template context for template %q: %v", t.Key, err)
	}
	return ctx, nil
}

// TemplateInclude provides more info to the template, dictating if should or not be included in some variable.
type TemplateInclude struct {
	*Template
	included bool
}

// Clone clones the template include.
func (ti *TemplateInclude) Clone() *TemplateInclude {
	return &TemplateInclude{ti.Template.Clone(), ti.included}
}

// Template is the model used for aggregate the template data.
type Template struct {
	Key       string
	Value     string
	Variables map[string]*TemplateVariable
	Children  map[string]map[string]*TemplateInclude
}

// NewTemplate is a constructor for Template.
func NewTemplate(key, value string) *Template {
	return &Template{
		Key:       key,
		Value:     value,
		Variables: make(map[string]*TemplateVariable),
		Children:  make(map[string]map[string]*TemplateInclude),
	}
}

// AddChild adds a template to the template children list.
func (t *Template) AddChild(env string, ct *Template, include bool) *TemplateInclude {
	ti := &TemplateInclude{ct, include}
	if _, ok := t.Children[env]; !ok {
		t.Children[env] = make(map[string]*TemplateInclude)
	}
	t.Children[env][ct.Key] = ti
	return ti
}

// AddVariable adds a new variable to the template.
func (t *Template) AddVariable(varName string) {
	if _, ok := t.Variables[varName]; !ok {
		t.Variables[varName] = &TemplateVariable{Name: varName}
	}
}

// AddOperation adds a new variable operation to the template.
func (t *Template) AddOperation(varName, argName string, args []string) {
	if _, ok := t.Variables[varName]; !ok {
		t.Variables[varName] = &TemplateVariable{Name: varName}
	}
	t.Variables[varName].Operations = append(t.Variables[varName].Operations,
		&VariableOperation{
			Name: argName,
			Args: args,
		})
}

// Comby returns the template value in a format valid for the comby search.
func (t *Template) Comby() string {
	return ClearString(
		fixQuestionMarkRegex.ReplaceAllStringFunc(
			unknownValueRegex.ReplaceAllStringFunc(
				variableRegex.ReplaceAllStringFunc(t.Value, func(variable string) string {
					end := len(variable) - 1
					if delIndex := strings.Index(variable, ":"); delIndex > -1 {
						end = delIndex
					}
					return fmt.Sprintf(FormatterVariable, variable[1:end])
				}),
				func(variable string) string {
					return fmt.Sprintf(variable[:1]+FormatterVariable, "_")
				}),
			func(variable string) string {
				return variable[1:]
			}),
	)
}

// Search executes the comby search in a target.
func (t *Template) Search(parent *Template, id, tmplKey, target string) ([]*SearchValue, error) {
	res, err := t.search(parent, id, tmplKey, target)
	if err == ErrNotFound || err == ErrNotMatch {
		return []*SearchValue{}, nil
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (t *Template) search(parent *Template, id, tmplKey, target string) ([]*SearchValue, error) {
	iterations, err := resolveSearch(t, parent, id, tmplKey, t.Comby(), target)
	if err != nil {
		return nil, err
	}
	if len(iterations) == 0 {
		return nil, ErrNotFound
	}
	var res []*SearchValue
	for _, it := range iterations {
		res = append(res, NewSearchValue(id, t.Key, []*SearchIteration{it}))
	}
	return res, nil
}

func resolveSearch(t, parent *Template, id, tmplKey, input, target string) ([]*SearchIteration, error) {
	cr, err := Execute(input, target)
	if err != nil {
		return nil, fmt.Errorf("executing comby command for the template %q: %w", t.Key, err)
	}

	// Check template include state.
	ti, ok := parent.Children[id][tmplKey]
	if cr == nil {
		if ok && ti.included {
			return nil, ErrNotMatch
		}
		return nil, ErrNotFound
	}
	if ok && !ti.included {
		return nil, ErrNotMatch
	}

	return getMatchesIterations(t, parent, id, tmplKey, input, cr.Matches)
}

func getMatchesIterations(t, parent *Template, id, tmplKey, input string, matches []Match) ([]*SearchIteration, error) {
	var foundIteration []*SearchIteration
	for _, match := range matches {
		// Fill the variables.
		variables, err := getEnvironmentValues(t, match.Environment)
		if err == ErrNotFound {
			if len(match.Matched) == 0 {
				continue
			}
			insideIterations, err := resolveSearch(t, parent, id, tmplKey, input, match.Matched[1:])
			if err == ErrNotFound || err == ErrNotMatch {
				continue
			}
			if err != nil {
				return nil, err
			}
			foundIteration = append(foundIteration, insideIterations...)
			continue
		}
		if err != nil {
			return nil, err
		}
		foundIteration = append(foundIteration, createMatchIteration(variables, match))
	}
	return foundIteration, nil
}

func getEnvironmentValues(t *Template, environment []MatchEnvironment) ([]*SearchValue, error) {
	var foundVariables []*SearchValue
	for _, env := range environment {
		// Check if the variable must match some template.
		children, ok := t.Children[env.Variable]

		// Create new variable.
		variable := createVariableValue(t.Key, env)
		foundVariables = append(foundVariables, variable)

		if !ok || len(children) == 0 {
			continue
		}

		// Add the children values to the variable.
		var invalidCount int
		for _, child := range children {
			values, err := child.search(t, env.Variable, child.Key, env.Value)
			switch {
			case err == ErrNotMatch:
				invalidCount++
				continue // The validation is made by the operations.
			case err == ErrNotFound:
				return nil, ErrNotFound
			case err != nil:
				return nil, fmt.Errorf("executing comby command for the template %q: %w", t.Key, err)
			}
			appendVariableValues(variable, values)
		}
		if invalidCount > 0 && invalidCount == len(children) {
			return nil, ErrNotFound
		}
	}
	return foundVariables, nil
}

func createMatchIteration(variables []*SearchValue, match Match) *SearchIteration {
	return &SearchIteration{
		Found:  ClearString(match.Matched),
		Values: variables,
	}
}

func createVariableValue(templKey string, env MatchEnvironment) *SearchValue {
	return NewSearchValue(env.Variable, templKey, []*SearchIteration{{Found: ClearString(env.Value)}})
}

func appendVariableValues(variable *SearchValue, values []*SearchValue) {
	// The variable must always have at least one iteration.
	variable.Iter[0].Values = append(variable.Iter[0].Values, values...)
}

// Clone clones the template.
func (t *Template) Clone() *Template {
	variables := make(map[string]*TemplateVariable)
	for k, v := range t.Variables {
		variables[k] = v.Clone()
	}
	children := make(map[string]map[string]*TemplateInclude)
	for k, templates := range t.Children {
		for _, template := range templates {
			if _, ok := children[k]; !ok {
				children[k] = make(map[string]*TemplateInclude)
			}
			children[k][template.Key] = template.Clone()
		}
	}
	return &Template{
		Key:       t.Key,
		Value:     t.Value,
		Variables: variables,
		Children:  children,
	}
}

// TemplateVariable represents a template variable and operations associated,
type TemplateVariable struct {
	Name       string
	Operations []*VariableOperation
}

// Clone clones the template variable.
func (tv *TemplateVariable) Clone() *TemplateVariable {
	var operations []*VariableOperation
	for _, o := range tv.Operations {
		operations = append(operations, o.Clone())
	}
	return &TemplateVariable{
		Name:       tv.Name,
		Operations: operations,
	}
}

// VariableOperation represents a template variable operation,
type VariableOperation struct {
	Name string
	Args []string
}

// Clone clones the variable operation.
func (vo *VariableOperation) Clone() *VariableOperation {
	var args []string
	args = append(args, vo.Args...)
	return &VariableOperation{
		Name: vo.Name,
		Args: args,
	}
}

// TemplateContext contains all the context data related to the template search.
type TemplateContext struct {
	Template           *Template
	TemplatesMap       map[string]*Template
	TemplatesCtxMap    map[string]*TemplateContext
	VariableOperations map[string][]InboundOperation
	KnownVariables     map[string]struct{}
}

// newTemplateContext is a constructor for TemplateContext.
// it uses a template context previously created.
func newTemplateContext(template *Template, ctx *TemplateContext) (*TemplateContext, error) {
	res := &TemplateContext{
		Template:           template,
		TemplatesMap:       ctx.TemplatesMap,
		TemplatesCtxMap:    ctx.TemplatesCtxMap,
		VariableOperations: make(map[string][]InboundOperation),
		KnownVariables:     make(map[string]struct{}),
	}
	return res.buildTemplateContext()
}

// NewTemplateContext is a constructor for TemplateContext.
func NewTemplateContext(template *Template, templatesMap map[string]*Template) (*TemplateContext, error) {
	res := &TemplateContext{
		Template:           template,
		TemplatesMap:       templatesMap,
		TemplatesCtxMap:    make(map[string]*TemplateContext),
		VariableOperations: make(map[string][]InboundOperation),
		KnownVariables:     make(map[string]struct{}),
	}
	return res.buildTemplateContext()
}

// Search uses the template value to execute a search in code.
func (ctx *TemplateContext) Search(code string) ([]*SearchValue, error) {
	return ctx.Template.Search(ctx.Template, ctx.Template.Key, ctx.Template.Key, code)
}

// ValidateChildren validates the variable operations for a set of search results.
func (ctx *TemplateContext) ValidateChildren(searches []*SearchValue) error {
	for _, search := range searches {
		for varName, operations := range ctx.VariableOperations {
			for _, op := range operations {
				if err := op.Validate(search); err != nil {
					return fmt.Errorf("validating children variable %q on template %q: %w", varName, ctx.Template.Key, err)
				}
			}
		}
	}
	return nil
}

// ValidateValues validates the variable values for a set of search results.
func (ctx *TemplateContext) ValidateValues(searches []*SearchValue, keyValue map[string]string) error {
	for _, search := range searches {
		if len(search.Iter) == 0 {
			continue
		}
		if value, ok := keyValue[search.Key]; ok {
			if value != search.Iter[0].Found {
				return fmt.Errorf("validating value on variable %q on template %q: values for the same variable do not match (expected %s but got %s)", search.Key, ctx.Template.Key, value, search.Iter[0].Found)
			}
			continue
		}
		keyValue[search.Key] = search.Iter[0].Found
		for _, it := range search.Iter {
			if err := ctx.ValidateValues(it.Values, keyValue); err != nil {
				return err
			}
		}
	}
	return nil
}

// buildTemplateContext builds a template context.
// parses the inbound operations.
// fills context data for the search.
func (ctx *TemplateContext) buildTemplateContext() (*TemplateContext, error) {
	for vName, vValue := range ctx.Template.Variables {
		var definitions []string
		for i := len(vValue.Operations) - 1; i > -1; i-- {
			var wrapper *InboundOperationWrapper
			if i > 0 && vValue.Operations[i-1].Name == NotOp {
				wrapper = newInboundOperationWrapper(vValue.Operations[i-1].Name, NewNotOperation)
			}

			op := vValue.Operations[i]
			switch op.Name {
			case DefinesOp:
				if wrapper != nil {
					return nil, fmt.Errorf("operation %q must not be wrapped with %q", op.Name, wrapper.name)
				}
				if i == 0 || !IsTemplateOperation(vValue.Operations[i-1]) {
					return nil, fmt.Errorf("operation %q must be after some template operation (e.g. %q, %q, ...)", op.Name, IncludeOp, IncludeOneOp)
				}
				for _, arg := range op.Args {
					ctx.KnownVariables[arg] = struct{}{}
					definitions = append(definitions, arg)
				}
			case IncludeOp, IncludeOneOp, IncludeAllOp:
				var inOp InboundOperation
				switch op.Name {
				case IncludeOp:
					inOp = NewIncludeOperation(ctx.Template, vName, op.Args, definitions)
				case IncludeOneOp:
					inOp = NewIncludeOneOperation(ctx.Template, vName, op.Args, definitions)
				case IncludeAllOp:
					inOp = NewIncludeAllOperation(ctx.Template, vName, op.Args, definitions)
				}
				if wrapper != nil {
					inOp = wrapper.fn(inOp)
					i--
				}

				if _, err := inOp.Execute(ctx); err != nil {
					return nil, err
				}
				ctx.VariableOperations[vName] = append(ctx.VariableOperations[vName], inOp)
				definitions = []string{}
			case NotOp:
				return nil, fmt.Errorf("command %q cannot be used with %q", op.Name, vValue.Operations[i+1].Name)
			default:
				return nil, fmt.Errorf("unknown command %q", op.Name)
			}
		}
		ctx.KnownVariables[vName] = struct{}{}
	}
	return ctx, nil
}

func IsTemplateOperation(v *VariableOperation) bool {
	for _, op := range []string{IncludeOp, IncludeOneOp, IncludeAllOp} {
		if v.Name == op {
			return true
		}
	}
	return false
}
