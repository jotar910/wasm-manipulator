package wtemplate

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

const (
	// Inbound
	DefinesOp    = "defines"
	IncludeOp    = "includes"
	IncludeOneOp = "includes_one"
	IncludeAllOp = "includes_all"
	NotOp        = "not_"
)

// GetValue returns the result value for the operations chain.
// the operations chain is defined over the template search result.
func GetValue(op OutboundOperation) (string, error) {
	result, err := op.Execute()
	if err != nil {
		return "", err
	}
	if len(result) == 0 || len(result[0].Iter) == 0 {
		return "", nil
	}
	return result[0].Iter[0].Found, nil
}

// InboundOperation is implemented by inbound template operations.
// these operations are meant to filter the template search result.
type InboundOperation interface {
	Execute(ctx *TemplateContext) (map[string]*TemplateInclude, error)
	Validate(s *SearchValue) error
	Error() error
}

// InboundOperationWrapperFn is a type for a wrapper function that accepts and returns inbound operations.
type InboundOperationWrapperFn func(o InboundOperation) InboundOperation

// InboundOperationWrapper is an inbound operation that wraps another inbound operation.
type InboundOperationWrapper struct {
	name string
	fn   InboundOperationWrapperFn
}

// newInboundOperationWrapper is a constructor for InboundOperationWrapper.
func newInboundOperationWrapper(name string, fn InboundOperationWrapperFn) *InboundOperationWrapper {
	return &InboundOperationWrapper{name: name, fn: fn}
}

// IncludeOperation implements the include operation.
// it is an inbound operation that validates if the template includes a variable.
type IncludeOperation struct {
	template    *Template
	variable    string
	key         string
	definitions []string
}

// NewIncludeOperation is a constructor for IncludeOperation.
func NewIncludeOperation(template *Template, variable string, keys, definitions []string) InboundOperation {
	if len(keys) != 1 {
		logrus.Fatal("include operation must have only one argument")
	}
	return &IncludeOperation{template: template, variable: variable, key: keys[0], definitions: definitions}
}

// Execute executes the operation.
func (o *IncludeOperation) Execute(ctx *TemplateContext) (map[string]*TemplateInclude, error) {
	return executeIncludeInboundOperation(ctx, []string{o.key}, o.definitions, o.variable)
}

// Validate validates if the operation restrictions are applied.
func (o *IncludeOperation) Validate(s *SearchValue) error {
	if includesVariable(s, o.variable, o.key) {
		return nil
	}
	return o.Error()
}

// Error returns the operation error.
func (o *IncludeOperation) Error() error {
	return OperationAbortedError{fmt.Sprintf("the template must be valid for the operation %q", IncludeOneOp)}
}

// IncludeOneOperation implements the include_one operation.
// it is an inbound operation that validates if the template at least includes one of the variables.
type IncludeOneOperation struct {
	template    *Template
	variable    string
	keys        []string
	definitions []string
}

// NewIncludeOneOperation is a constructor for IncludeOneOperation.
func NewIncludeOneOperation(template *Template, variable string, keys, definitions []string) InboundOperation {
	return &IncludeOneOperation{template: template, variable: variable, keys: keys, definitions: definitions}
}

// Execute executes the operation.
func (o *IncludeOneOperation) Execute(ctx *TemplateContext) (map[string]*TemplateInclude, error) {
	return executeIncludeInboundOperation(ctx, o.keys, o.definitions, o.variable)
}

// Validate validates if the operation restrictions are applied.
func (o *IncludeOneOperation) Validate(s *SearchValue) error {
	for _, templName := range o.keys {
		if includesVariable(s, o.variable, templName) {
			return nil
		}
	}
	return o.Error()
}

// Error returns the operation error.
func (o *IncludeOneOperation) Error() error {
	return OperationAbortedError{fmt.Sprintf("at least one template must be valid for the operation %q", IncludeOneOp)}
}

// IncludeAllOperation implements the include_all operation.
// it is an inbound operation that validates if the template includes all the variables.
type IncludeAllOperation struct {
	template    *Template
	variable    string
	keys        []string
	definitions []string
}

// NewIncludeAllOperation is a constructor for IncludeAllOperation.
func NewIncludeAllOperation(template *Template, variable string, keys, definitions []string) InboundOperation {
	return &IncludeAllOperation{template: template, variable: variable, keys: keys, definitions: definitions}
}

// Execute executes the operation.
func (o *IncludeAllOperation) Execute(ctx *TemplateContext) (map[string]*TemplateInclude, error) {
	return executeIncludeInboundOperation(ctx, o.keys, o.definitions, o.variable)
}

// Validate validates if the operation restrictions are applied.
func (o *IncludeAllOperation) Validate(s *SearchValue) error {
	for _, templName := range o.keys {
		if !includesVariable(s, o.variable, templName) {
			return o.Error()
		}
	}
	return nil
}

// Error returns the operation error.
func (o *IncludeAllOperation) Error() error {
	return OperationAbortedError{fmt.Sprintf("all the templates must be valid for the operation %q", IncludeAllOp)}
}

// NotOperation implements the negation operation.
// it is an inbound operation that wraps another operation.
// it negates the underlying operation result.
type NotOperation struct {
	op InboundOperation
}

// NewNotOperation is a constructor for NotOperation.
func NewNotOperation(op InboundOperation) InboundOperation {
	return &NotOperation{op: op}
}

// Execute executes the operation.
func (o *NotOperation) Execute(ctx *TemplateContext) (map[string]*TemplateInclude, error) {
	res, err := o.op.Execute(ctx)
	if err != nil {
		return res, err
	}
	for _, v := range res {
		v.included = false
	}
	return res, nil
}

// Validate validates if the operation restrictions are applied.
func (o *NotOperation) Validate(s *SearchValue) error {
	res := o.op.Validate(s)
	if res == nil {
		return o.Error()
	}
	return nil
}

// Error returns the operation error.
func (o *NotOperation) Error() error {
	return OperationAbortedError{fmt.Sprintf("not(%q)", o.op.Error())}
}

// OutboundOperation is implemented by outbound template operations.
// these operations are meant to customize the output of a template search result.
type OutboundOperation interface {
	Execute() ([]*SearchValue, error)
	Valid(k string) bool
}

// OutboundOp consists on the base structure for outbound operations.
// it contains the unchanged search result and some common context data.
type OutboundOp struct {
	search         []*SearchValue
	knownVariables map[string]struct{}
}

// NewOutboundOp is a constructor for OutboundOp.
func NewOutboundOp(search []*SearchValue, knownVariables map[string]struct{}) OutboundOperation {
	return &OutboundOp{search: search, knownVariables: knownVariables}
}

// Execute executes the operation.
// returns a new value accordingly to the type of operation.
func (o *OutboundOp) Execute() ([]*SearchValue, error) {
	return o.search, nil
}

// Valid returns if the selector k is a valid selector.
func (o *OutboundOp) Valid(k string) bool {
	_, ok := o.knownVariables[k]
	return ok
}

// SelectOp represents the operation select,
// it is an outbound operation responsible to find a variable in the search result.
// the variable name must match the selector.
type SelectOp struct {
	prev     OutboundOperation
	selector string
}

// NewSelectOp is a constructor for SelectOp.
func NewSelectOp(prev OutboundOperation, selector string) OutboundOperation {
	return &SelectOp{
		prev:     prev,
		selector: selector,
	}
}

// Execute executes the operation.
// returns a new value accordingly to the type of operation.
func (o *SelectOp) Execute() ([]*SearchValue, error) {
	if !o.Valid(o.selector) {
		return nil, fmt.Errorf("invalid selector executing select operation: %q is not registered as a known variable", o.selector)
	}
	search, err := o.prev.Execute()
	if err != nil {
		return nil, fmt.Errorf("executing previous operation on select operation: %w", err)
	}
	var res []*SearchValue
	for _, s := range search {
		if s, ok := s.Get(o.selector); ok {
			res = append(res, s)
		}
	}
	return res, nil
}

// Valid returns if the selector k is a valid selector.
func (o *SelectOp) Valid(k string) bool {
	if o.prev == nil {
		return false
	}
	return o.prev.Valid(k)
}

// RemoveOp represents the operation remove,
// it is an outbound operation responsible to remove a variable in the search result.
// the variable name must match the selector.
// it returns the new search result without the value of the removed variable.
type RemoveOp struct {
	prev     OutboundOperation
	selector string
}

// NewRemoveOp is a constructor for RemoveOp.
func NewRemoveOp(prev OutboundOperation, selector string) OutboundOperation {
	return &RemoveOp{
		prev:     prev,
		selector: selector,
	}
}

// Execute executes the operation.
// returns a new value accordingly to the type of operation.
func (o *RemoveOp) Execute() ([]*SearchValue, error) {
	if !o.Valid(o.selector) {
		return nil, fmt.Errorf("invalid selector executing remove operation: %q is not registered as a known variable", o.selector)
	}
	search, err := o.prev.Execute()
	if err != nil {
		return nil, fmt.Errorf("executing previous operation on select operation: %w", err)
	}
	var res []*SearchValue
	for _, s := range search {
		if newS := s.Clone().Remove(o.selector); newS != nil {
			res = append(res, newS)
		}
	}
	return res, err
}

// Valid returns if the selector k is a valid selector.
func (o *RemoveOp) Valid(k string) bool {
	if o.prev == nil {
		return false
	}
	return o.prev.Valid(k)
}

// ReplaceOpArg is the argument model for the replace operation argument.
type ReplaceOpArg struct {
	value       string
	isReference bool
}

// NewReplaceOpArg is a constructor for ReplaceOpArg.
func NewReplaceOpArg(value string, isReference bool) *ReplaceOpArg {
	return &ReplaceOpArg{value: value, isReference: isReference}
}

// ReplaceOp represents the operation replace,
// it is an outbound operation responsible to replace a variable in the search result.
// the variable name must match the selector.
// it returns the new search result with the new variable value replacing the old one.
type ReplaceOp struct {
	prev OutboundOperation
	old  *ReplaceOpArg
	new  *ReplaceOpArg
}

// NewReplaceOp is a constructor for ReplaceOp.
func NewReplaceOp(prev OutboundOperation, old, new *ReplaceOpArg) OutboundOperation {
	return &ReplaceOp{
		prev: prev,
		old:  old,
		new:  new,
	}
}

// Execute executes the operation.
// returns a new value accordingly to the type of operation.
func (o *ReplaceOp) Execute() ([]*SearchValue, error) {
	if o.old.isReference && !o.Valid(o.old.value) {
		return nil, fmt.Errorf("invalid old selector executing replace operation: %q is not registered as a known variable", o.old.value)
	}
	if o.new.isReference && !o.Valid(o.new.value) {
		return nil, fmt.Errorf("invalid new selector executing replace operation: %q is not registered as a known variable", o.new.value)
	}
	search, err := o.prev.Execute()
	if err != nil {
		return nil, fmt.Errorf("executing previous operation on select operation: %w", err)
	}
	var res []*SearchValue
	for _, s := range search {
		if newS := s.Clone().Replace(o.old, o.new); newS != nil {
			res = append(res, newS)
		}
	}
	return res, err
}

// Valid returns if the selector k is a valid selector.
func (o *ReplaceOp) Valid(k string) bool {
	if o.prev == nil {
		return false
	}
	return o.prev.Valid(k)
}

// executeIncludeInboundOperation executes inbound operations of type include.
func executeIncludeInboundOperation(ctx *TemplateContext, keys, definitions []string, variable string) (map[string]*TemplateInclude, error) {
	res := make(map[string]*TemplateInclude)
	for _, key := range keys {
		otherTemplCtx, ok := ctx.TemplatesCtxMap[key]
		if !ok {
			// Creating the context for the template that belongs to this key.
			ctxAux, err := createTemplateContext(ctx, key)
			if err != nil {
				return nil, err
			}
			ctx.TemplatesCtxMap[ctxAux.Template.Key] = ctxAux
			otherTemplCtx = ctxAux
		}
		// Check if all the definitions on the referent template are defined,
		for _, def := range definitions {
			if _, ok := otherTemplCtx.KnownVariables[def]; !ok {
				return nil, fmt.Errorf("variable %s on template %s was not defined on dependent templates", def, key)
			}
		}

		// Append the template to this on attributing to its variable
		res[otherTemplCtx.Template.Key] = ctx.Template.AddChild(variable, otherTemplCtx.Template, true)
	}
	return res, nil
}

// includesVariable returns if some search value includes a variable.
func includesVariable(s *SearchValue, varName, templName string) bool {
	if len(s.Iter) == 0 {
		return false
	}
	for _, iter := range s.Iter {
		if iter.Values == nil {
			continue
		}
		for _, val := range iter.Values {
			if templName == val.Templ {
				return true
			}
			if varName == val.Key {
				return includesVariable(val, varName, templName)
			}
		}
	}
	return false
}
