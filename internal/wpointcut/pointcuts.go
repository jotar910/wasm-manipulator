package wpointcut

import (
	"fmt"
	"joao/wasm-manipulator/internal/wcode"
	"regexp"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"

	"joao/wasm-manipulator/internal/wkeyword"
	"joao/wasm-manipulator/internal/wparser/pointcut"
	"joao/wasm-manipulator/internal/wtemplate"
)

var getLocalReg = regexp.MustCompile(`\(local.get (?P<index>[^)\s]+)\)`)

// funcNode is a node for the function pointcut.
type funcNode struct {
	*NodeInstance
	props *funcPointcutProps
}

// newFuncNode is a constructor for funcNode.
func newFuncNode(name string, typ NodeType, definition *pointcut.FuncDefinition) *funcNode {
	return &funcNode{
		NodeInstance: newEmptyNodeInstance(name, typ),
		props:        &funcPointcutProps{newPointcutFuncProps(definition)},
	}
}

// Filter filters the pointcut context accordingly to the current node.
func (node *funcNode) Filter(in *PointcutContext) *PointcutContext {
	var joinPoints []*JoinPoint
	for _, jp := range in.joinPoints {
		var blocks []*wcode.JoinPointBlock
		for _, b := range jp.blocks {
			calls := in.context.FindFunctions(b, node.filterFuncFn)
			calls.RemoveDuplicates()
			found := calls.Found()
			for _, bf := range found {
				bf.Metadata.Join(b.Metadata)
				blocks = append(blocks, bf)
			}
		}
		if len(blocks) > 0 {
			joinPoints = append(joinPoints, newJoinPoint(blocks...))
		}
	}
	jp := in.clone()
	jp.joinPoints = joinPoints
	return jp
}

// filterFuncFn is the callback implementation for the function filter.
func (node *funcNode) filterFuncFn(ctx *wcode.ModuleContext, funcData *wcode.FuncData) (map[string]wkeyword.Object, bool) {
	return filterFuncBasedFn(&node.props.functionPointcutProps, ctx, funcData)
}

// callNode is a node for the call pointcut.
type callNode struct {
	*NodeInstance
	props *callPointcutProps
}

// newCallNode is a constructor for callNode.
func newCallNode(name string, typ NodeType, definition *pointcut.FuncDefinition) *callNode {
	return &callNode{
		NodeInstance: newEmptyNodeInstance(name, typ),
		props:        &callPointcutProps{newPointcutFuncProps(definition)},
	}
}

// Filter filters the pointcut context accordingly to the current node.
func (node *callNode) Filter(in *PointcutContext) *PointcutContext {
	var joinPoints []*JoinPoint
	for _, jp := range in.joinPoints {
		var blocks []*wcode.JoinPointBlock
		for _, b := range jp.blocks {
			calls := in.context.FindCalls(b, node.filterCallFn)
			calls.RemoveDuplicates()
			found := calls.Found()
			for _, bf := range found {
				bf.Metadata.Join(b.Metadata)
				blocks = append(blocks, bf)
			}
		}
		if len(blocks) > 0 {
			joinPoints = append(joinPoints, newJoinPoint(blocks...))
		}
	}
	jp := in.clone()
	jp.joinPoints = joinPoints
	return jp
}

// filterCallFn is the callback implementation for the call filter.
func (node *callNode) filterCallFn(ctx *wcode.ModuleContext, callData *wcode.CallData) (map[string]wkeyword.Object, bool) {
	return filterFuncBasedFn(&node.props.functionPointcutProps, ctx, callData.Callee)
}

// argValue is the model for a single argument value on the args zone.
type argValue struct {
	ElType string
	Index  string
	Type   string
}

// argsNode is a node for the call pointcut.
type argsNode struct {
	*NodeInstance
	props *argsPointcutProps
}

// newArgsNode is a constructor for argsNode.
func newArgsNode(name string, typ NodeType, definition *pointcut.ArgMethodInput, params map[string]ParsedParam) *argsNode {
	return &argsNode{
		NodeInstance: newEmptyNodeInstance(name, typ),
		props:        &argsPointcutProps{resolvePointcutArgsParams(definition, params)},
	}
}

// Filter filters the pointcut context accordingly to the current node.
func (node *argsNode) Filter(in *PointcutContext) *PointcutContext {
	var joinPoints []*JoinPoint
	for _, jp := range in.joinPoints {
		var blocks []*wcode.JoinPointBlock
		for _, b := range jp.blocks {
			calls := in.context.FindArgs(b, node.filterArgsFn)
			calls.RemoveDuplicates()
			found := calls.Found()
			for _, bf := range found {
				bf.Metadata.Join(b.Metadata)
				blocks = append(blocks, bf)
			}
		}
		if len(blocks) > 0 {
			joinPoints = append(joinPoints, newJoinPoint(blocks...))
		}
	}
	jp := in.clone()
	jp.joinPoints = joinPoints
	return jp
}

// filterArgsFn is the callback implementation for the args filter.
func (node *argsNode) filterArgsFn(_ *wcode.ModuleContext, argsData *wcode.ArgsData) (map[string]wkeyword.Object, bool) {
	environment := make(map[string]wkeyword.Object)
	for i, param := range *node.props.params {
		var newArg argValue

		// Check local.get instruction.
		matches := getLocalReg.FindStringSubmatch(argsData.Args[i].Instr)
		if len(matches) < 2 && matches[1] != param.Index {
			return nil, false
		}

		if param.Variable == "param" {
			if ok := node.checkPointcutParameter(argsData.Caller, param); !ok {
				return nil, false
			}
			newArg.Index = param.Index // fill param index.
		} else {
			if ok := node.checkPointcutLocal(argsData.Caller, param); !ok {
				return nil, false
			}
			newArg.Index = param.Index // fill param index.
		}

		newArg.Type = param.Variable // fill param type.
		newArg.ElType = param.Type   // fill param local.

		environment[param.Name] = wkeyword.NewKwObject(newArg)
	}
	return environment, true
}

// checkPointcutParameter checks if the pointcut parameter is valid for the caller data.
func (node *argsNode) checkPointcutParameter(caller *wcode.FuncData, param ParsedParam) bool {
	if index, err := strconv.Atoi(param.Index); err == nil {
		if index >= caller.TotalParams || caller.ParamTypes[index] != param.Type {
			return false
		}
	} else {
		var found bool
		for index, paramName := range caller.Params {
			if paramName == param.Index {
				if caller.ParamTypes[index] == param.Type {
					found = true
				}
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// checkPointcutParameter checks if the pointcut local is valid for the caller data.
func (node *argsNode) checkPointcutLocal(caller *wcode.FuncData, param ParsedParam) bool {
	if index, err := strconv.Atoi(param.Index); err == nil {
		if index >= caller.TotalLocals || caller.LocalTypes[index] != param.Type {
			return false
		}
	} else {
		var found bool
		for index, localName := range caller.Locals {
			if localName == param.Index {
				if caller.LocalTypes[index] == param.Type {
					found = true
				}
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// returnsNode is a node for the call pointcut.
type returnsNode struct {
	*NodeInstance
	typ *string
}

// newReturnsNode is a constructor for returnsNode.
func newReturnsNode(name string, typ NodeType, returnsTyp *string) *returnsNode {
	return &returnsNode{
		NodeInstance: newEmptyNodeInstance(name, typ),
		typ:          returnsTyp,
	}
}

// Filter filters the pointcut context accordingly to the current node.
func (node *returnsNode) Filter(in *PointcutContext) *PointcutContext {
	var joinPoints []*JoinPoint
	for _, jp := range in.joinPoints {
		var blocks []*wcode.JoinPointBlock
		for _, b := range jp.blocks {
			calls := in.context.FindReturns(b, node.filterReturnsFn)
			calls.RemoveDuplicates()
			found := calls.Found()
			for _, bf := range found {
				bf.Metadata.Join(b.Metadata)
				blocks = append(blocks, bf)
			}
		}
		if len(blocks) > 0 {
			joinPoints = append(joinPoints, newJoinPoint(blocks...))
		}
	}
	jp := in.clone()
	jp.joinPoints = joinPoints
	return jp
}

// filterReturnsFn is the callback implementation for the returns filter.
func (node *returnsNode) filterReturnsFn(_ *wcode.ModuleContext, returnsData *wcode.ReturnsData) (map[string]wkeyword.Object, bool) {
	if node.typ != nil && returnsData.Type != *node.typ {
		return nil, false
	}
	return make(map[string]wkeyword.Object), true
}

// templateNode is a node for the call pointcut.
type templateNode struct {
	*NodeInstance
	props     string
	justCheck bool
}

// newTemplateNode is a constructor for templateNode.
func newTemplateNode(name string, typ NodeType, definition string, justCheck bool) *templateNode {
	return &templateNode{
		NodeInstance: newEmptyNodeInstance(name, typ),
		props:        definition,
		justCheck:    justCheck,
	}
}

// templateNodeFilterAux is an auxiliary structure to filter and find templates results.
type templateNodeFilterAux struct {
	index   int
	results []*wcode.JoinPointBlock
}

// Filter filters the pointcut context accordingly to the current node.
func (node *templateNode) Filter(in *PointcutContext) *PointcutContext {
	var joinPoints []*JoinPoint
	for _, jp := range in.joinPoints {
		ch := make(chan *templateNodeFilterAux)
		blocksTable := make([][]*wcode.JoinPointBlock, len(jp.blocks))
		go func(jp *JoinPoint) {
			wg := new(sync.WaitGroup)
			for i, b := range jp.blocks {
				wg.Add(1)
				go func(i int, b *wcode.JoinPointBlock) {
					if res, ok := node.findResults(in, b); ok {
						ch <- &templateNodeFilterAux{i, res}
					} else {
						ch <- &templateNodeFilterAux{i, []*wcode.JoinPointBlock{}}
					}
					wg.Done()
				}(i, b)
			}
			wg.Wait()
			close(ch)
		}(jp)
		for bRes := range ch {
			var results []*wcode.JoinPointBlock
			for _, br := range bRes.results {
				br.Metadata = jp.blocks[bRes.index].Metadata
				results = append(results, bRes.results...)
			}
			blocksTable[bRes.index] = results
		}
		var blocks []*wcode.JoinPointBlock
		for _, bs := range blocksTable {
			blocks = append(blocks, bs...)
		}
		if len(blocks) > 0 {
			if node.justCheck { // Just for checking if there is a match.
				joinPoints = append(joinPoints, newJoinPoint(jp.blocks...))
			} else {
				joinPoints = append(joinPoints, newJoinPoint(blocks...))
			}
		}
	}
	jp := in.clone()
	jp.joinPoints = joinPoints
	return jp
}

// findResults searches for the template results on the join-point block.
func (node *templateNode) findResults(jp *PointcutContext, jpBlock *wcode.JoinPointBlock) ([]*wcode.JoinPointBlock, bool) {
	jpInstr := jpBlock.Instr()

	// Applies template to function instructions (filling the context)
	result, tc, err := applyTemplateToFunction(jp.templ, node.props, jpInstr.String())
	if err != nil {
		logrus.Fatalf("applying template to function: %v", err)
	}

	// Add result code to response
	if len(result) == 0 || len(result[0].Iter) == 0 {
		return nil, false
	}

	funcDef := jpBlock.FuncDefinition()

	newTemplResults := newTemplateResultsMap().
		addResult(funcDef.Name, node.props, result).
		merge(jp.templ.results)
	jp.templ.results = newTemplResults

	var res []*wcode.JoinPointBlock
	for _, result := range result {
		resultsSearch := jp.context.NewSearch()
		for _, iter := range result.Iter {
			resultsSearch.Merge(jp.context.FindInstructions(jpInstr, iter.Found))
		}
		res = append(res, resultsSearch.Found()...)
	}

	return wcode.RearrangeBlocks(res, countInstructions(tc), countInstructions), true
}

// otherMethodNode is a method node for different pointcut methods than the default ones.
// these methods are included in the global pointcuts definition.
type otherMethodNode struct {
	*NodeInstance
	Expr Node
}

// newOtherMethodNode is a constructor for otherMethodNode.
func newOtherMethodNode(pp *ParsedPointcut, name string, typ NodeType, pointcutValue string, arguments []string) *otherMethodNode {
	// Parse pointcut input.
	pc, err := pointcut.ParseWithoutContext(pointcutValue)
	if err != nil {
		logrus.Fatalf("creating other method node %s: parsing pointcut expression %q: %v", name, pointcutValue, err)
	}
	if expectedLen, gotLen := len(pc.Args), len(arguments); expectedLen != gotLen {
		logrus.Fatalf("creating other method node %s: expects %d arguments but got %d", name, expectedLen, gotLen)
	}
	for argIndex, callArg := range arguments {
		joinPointParam, ok := pp.Params[callArg]
		if !ok {
			logrus.Fatalf("creating other method node %s: unknown argument at position %d", name, argIndex)
		}
		if joinPointParam.Type != pc.Args[argIndex].Type {
			logrus.Fatalf("creating other method node %s: argument type at position %d does not match the expected: expected %s but got %s", name, argIndex, pc.Args[argIndex].Type, joinPointParam.Type)
		}
	}
	return &otherMethodNode{
		NodeInstance: newEmptyNodeInstance(name, typ),
		Expr:         pp.parseExpression(pc.Instrs, &parseStacks{}),
	}
}

// Filter filters the pointcut context accordingly to the current node.
func (o *otherMethodNode) Filter(in *PointcutContext) *PointcutContext {
	return o.Expr.Filter(in)
}

// functionPointcutProps is a pointcut definition for the function properties.
type functionPointcutProps struct {
	returnType *functionPointcutPropsReturn
	name       *functionPointcutPropsName
	params     *[]functionPointcutPropsParams
	scope      pointcut.FunctionScope
}

// newPointcutFuncProps is a constructor for functionPointcutProps.
func newPointcutFuncProps(props *pointcut.FuncDefinition) functionPointcutProps {
	return functionPointcutProps{
		name:       resolvePointcutFunctionName(props.Name),
		returnType: resolvePointcutFunctionReturn(props.ReturnType),
		params:     resolvePointcutFunctionParams(props.Params),
		scope:      props.Scope,
	}
}

// functionPointcutPropsReturn is a pointcut definition for the return value on function.
type functionPointcutPropsReturn struct {
	Type     *string
	Variable *string
}

// functionPointcutPropsName is a pointcut definition for the name value on function.
type functionPointcutPropsName struct {
	Variable  *string
	Regex     *string
	Name      *string
	IndexName *string
	Index     *int
}

// functionPointcutPropsParams is a pointcut definition for the parameters values on function.
type functionPointcutPropsParams struct {
	Type     *string
	Variable *string
}

// funcPointcutProps contains the func pointcut properties.
type funcPointcutProps struct {
	functionPointcutProps
}

// callPointcutProps contains the call pointcut properties.
type callPointcutProps struct {
	functionPointcutProps
}

// argsPointcutProps contains the args pointcut properties.
type argsPointcutProps struct {
	params *[]ParsedParam
}

// filterFuncBasedFn validates if the input function properties match with some function based pointcut.
func filterFuncBasedFn(props *functionPointcutProps, ctx *wcode.ModuleContext, data *wcode.FuncData) (map[string]wkeyword.Object, bool) {
	if props.name != nil {
		switch n := props.name; {
		case n.Index != nil:
			if data.Order != *n.Index {
				return nil, false
			}
		case n.Name != nil:
			if data.Name != *n.Name {
				return nil, false
			}
		case n.IndexName != nil:
			if data.Index != *n.IndexName {
				return nil, false
			}
		case n.Regex != nil:
			if _, ok := ctx.ExportFunctionByRegex(*n.Regex); !ok {
				return nil, false
			}
		}
	}
	if props.returnType != nil && props.returnType.Type != nil &&
		*props.returnType.Type != data.ResultType {
		return nil, false
	}
	if props.params != nil {
		params := *props.params
		if len(params) != data.TotalParams {
			return nil, false
		}
		for i, param := range params {
			if *param.Type != data.ParamTypes[i] {
				return nil, false
			}
		}
	}
	switch props.scope {
	case pointcut.FunctionScopeInternal:
		if data.IsImported || data.IsExported {
			return nil, false
		}
	case pointcut.FunctionScopeExported:
		if !data.IsExported {
			return nil, false
		}
	case pointcut.FunctionScopeImported:
		if !data.IsImported {
			return nil, false
		}
	case pointcut.FunctionScopeStart:
		if !data.IsStart {
			return nil, false
		}
	default:
		// Empty by design.
	}
	return getFuncBasedEnvironment(props, data), true
}

// getFuncBasedEnvironment fills the environment zone with the function definition data.
func getFuncBasedEnvironment(props *functionPointcutProps, callData *wcode.FuncData) map[string]wkeyword.Object {
	environment := make(map[string]wkeyword.Object)
	if props.name != nil && props.name.Variable != nil {
		environment[*props.name.Variable] = wkeyword.NewKwPrimitive(callData.Name)
	}
	if props.returnType != nil && props.returnType.Variable != nil {
		environment[*props.returnType.Variable] = wkeyword.NewKwPrimitive(callData.ResultType)
	}
	if props.params != nil {
		for i, params := range *props.params {
			if i >= callData.TotalParams {
				break
			}
			if params.Variable == nil {
				continue
			}
			environment[*params.Variable] = wkeyword.NewKwObject(map[string]interface{}{
				"Name": callData.Params[i],
				"Type": callData.ParamTypes[i],
			})
		}
	}
	return environment
}

// applyTemplateToFunction applies a template search to the function.
func applyTemplateToFunction(templManager *templateManager, name string, code string) ([]*wtemplate.SearchValue, string, error) {
	if len(code) == 0 {
		return nil, "", nil
	}
	ctxTempl, ok := templManager.context[name]
	if !ok {
		template, ok := templManager.templates[name]
		if !ok {
			return nil, "", fmt.Errorf("template not found %q", name)
		}
		templateCtx, err := wtemplate.NewTemplateContext(template, templManager.templates)
		if err != nil {
			return nil, "", fmt.Errorf("creating template context for template %q: %v", name, err)
		}
		templManager.context[name] = templateCtx
		ctxTempl = templateCtx
	}

	search, err := ctxTempl.Search(code)
	if err != nil {
		return nil, "", fmt.Errorf("searching for template %q", name)
	}
	if len(search) == 0 {
		return search, ctxTempl.Template.Comby(), nil
	}
	// Validate search with the template context.
	err = ctxTempl.ValidateChildren(search)
	if err != nil {
		logrus.Debugf("validating template: %v", err)
		return []*wtemplate.SearchValue{}, ctxTempl.Template.Comby(), nil
	}
	// Validate search result values.
	// This validation is made only after the first level
	for _, result := range search {
		if len(result.Iter) == 0 {
			continue
		}
		for _, iter := range result.Iter {
			err = ctxTempl.ValidateValues(iter.Values, make(map[string]string))
			if err != nil {
				return []*wtemplate.SearchValue{}, ctxTempl.Template.Comby(), nil
			}
		}
	}
	return search, ctxTempl.Template.Comby(), nil
}

// countInstructions counts the number of instructions.
func countInstructions(value string) int {
	var count int
	var openCount int
	for _, c := range value {
		if c == '(' {
			if openCount == 0 {
				count++
			}
			openCount++
			continue
		}
		if c != ')' || openCount == 0 {
			continue
		}
		openCount--
	}
	return count
}
