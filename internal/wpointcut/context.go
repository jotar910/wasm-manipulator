package wpointcut

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"joao/wasm-manipulator/internal/wcode"
	"joao/wasm-manipulator/internal/wkeyword"
	"joao/wasm-manipulator/internal/wparser/template"
	"joao/wasm-manipulator/internal/wtemplate"
	"joao/wasm-manipulator/internal/wyaml"
)

// PointcutContext is the context model for a pointcut execution.
// it contains the information needed to find join-points, as well as apply modifications.
type PointcutContext struct {
	context    *wcode.ModuleContext
	joinPoints []*JoinPoint
	templ      *templateManager
}

// newPointcutContext is the constructor for PointcutContext.
func newPointcutContext(context *wcode.ModuleContext, transformation *wyaml.BaseYAML) *PointcutContext {
	var joinPoints []*JoinPoint
	for _, found := range context.InitSearch().Found() {
		joinPoints = append(joinPoints, newJoinPoint(found))
	}
	return &PointcutContext{
		context:    context,
		joinPoints: joinPoints,
		templ:      newTemplateManager(transformation),
	}
}

// Templates returns the templates keyword maps from the context.
func (ctx *PointcutContext) Templates(fnIndex string, instrIndex int, keywordMaps ...wkeyword.KeywordsMap) (wkeyword.KeywordsMap, bool) {
	if res, ok := ctx.templ.GetResults(fnIndex, instrIndex, keywordMaps); ok {
		if res == nil {
			return nil, true
		}
		return res, true
	}
	return nil, false
}

// All returns the list of join-points.
func (ctx *PointcutContext) All() []*JoinPoint {
	return ctx.joinPoints
}

// append appends a pointcut-context to the current.
func (ctx *PointcutContext) append(o *PointcutContext) *PointcutContext {
	if len(o.joinPoints) == 0 {
		return ctx.clone()
	}
	if len(ctx.joinPoints) == 0 {
		return o.clone()
	}

	res := ctx.clone()
	res.joinPoints = []*JoinPoint{}

	jpMap := make(map[string]*JoinPoint)

	for _, jp := range ctx.joinPoints {
		jpMap[jp.FuncDefinition().Name] = jp.clone()
	}

	for _, jp := range o.joinPoints {
		jpEntry, ok := jpMap[jp.FuncDefinition().Name]
		if !ok {
			jpMap[jp.FuncDefinition().Name] = jp.clone()
			continue
		}
		jpEntry.blocks = res.context.Union(jpEntry.blocks, jp.blocks)
	}

	for _, jp := range jpMap {
		res.joinPoints = append(res.joinPoints, jp)
	}

	return res
}

// clone clones the pointcut-context.
func (ctx *PointcutContext) clone() *PointcutContext {
	var joinPoints []*JoinPoint
	for _, jp := range ctx.joinPoints {
		joinPoints = append(joinPoints, jp.clone())
	}
	return &PointcutContext{
		context:    ctx.context,
		joinPoints: joinPoints,
		templ:      ctx.templ.clone(),
	}
}

// JoinPoint encapsulates the data for a join-point.
type JoinPoint struct {
	blocks []*wcode.JoinPointBlock
}

// newJoinPoint is a constructor for JoinPoint.
func newJoinPoint(blocks ...*wcode.JoinPointBlock) *JoinPoint {
	return &JoinPoint{blocks: blocks}
}

// FuncDefinition returns the function definition for the join-point.
func (jp *JoinPoint) FuncDefinition() *wcode.FunctionDefinition {
	if len(jp.blocks) == 0 {
		return nil
	}
	return jp.blocks[0].FuncDefinition()
}

// InstrString returns some instruction as string.
func (jp *JoinPoint) InstrString(i int) string {
	return wcode.FuncInstrsString(jp.blocks[i].Instr())
}

// Blocks returns the block instructions.
func (jp *JoinPoint) Blocks() []*wcode.JoinPointBlock {
	return jp.blocks
}

// InstrString returns the join-point instruction value as string.
func (jp *JoinPoint) InstrsString() string {
	var res []string
	for i := range jp.blocks {
		if str := jp.InstrString(i); str != "" {
			res = append(res, str)
		}
	}
	return strings.Join(res, " ")
}

// String returns the textual description for the join-point.
func (jp *JoinPoint) String() string {
	var res []string
	for _, block := range jp.blocks {
		res = append(res, block.String())
	}
	return fmt.Sprintf("{%s}", strings.Join(res, ", "))
}

// clone clones the join-pont.
func (jp *JoinPoint) clone() *JoinPoint {
	return &JoinPoint{blocks: append([]*wcode.JoinPointBlock{}, jp.blocks...)}
}

// templateResultsMap is a map that keeps track of the template results while executing the pointcut.
// the map contains as key the function index.
// the map contains as value a map of results per template name.
// each result is present in a function that can be matched by multiple templates.
type templateResultsMap map[string]map[string][]*wtemplate.SearchValue

// newTemplateResultsMap is a constructor for templateResultsMap.
func newTemplateResultsMap() templateResultsMap {
	return make(templateResultsMap)
}

// addResult adds a new result to the map.
func (tr templateResultsMap) addResult(i, k string, r []*wtemplate.SearchValue) templateResultsMap {
	if _, ok := tr[i]; !ok {
		tr[i] = make(map[string][]*wtemplate.SearchValue)
	}
	tr[i][k] = r
	return tr
}

// merge merges a template results map into the current map.
func (tr templateResultsMap) merge(o templateResultsMap) templateResultsMap {
	if o == nil {
		return tr
	}
	for k, v := range o {
		if _, ok := tr[k]; !ok {
			tr[k] = v
			continue
		}
		for rK, r := range v {
			if _, ok := tr[k][rK]; !ok {
				tr[k][rK] = r
			}
		}
	}
	return tr
}

// templateManager is responsible to manage the template execution state while executing the pointcut.
// keeps track of the template context map, the existent templates data, and the results map.
type templateManager struct {
	context   map[string]*wtemplate.TemplateContext
	templates map[string]*wtemplate.Template
	results   templateResultsMap
}

// newTemplateManager is a constructor for templateManager.
func newTemplateManager(transformation *wyaml.BaseYAML) *templateManager {
	return &templateManager{
		templates: fillTemplatesMap(transformation),
		context:   make(map[string]*wtemplate.TemplateContext),
		results:   make(map[string]map[string][]*wtemplate.SearchValue),
	}
}

// GetResults returns the template results for a given function.
// it filters the template results by the current context state.
func (t *templateManager) GetResults(fnIndex string, instrIndex int, keywordMaps []wkeyword.KeywordsMap) (*wkeyword.TemplateResults, bool) {
	currentRes := t.results[fnIndex]
	if len(currentRes) == 0 {
		return nil, true
	}
	filteredRes := make(map[string][]*wtemplate.SearchValue)
	for key, search := range currentRes {
		if instrIndex >= len(search) {
			if _, ok := filteredRes[key]; !ok {
				filteredRes[key] = search
			}
			continue
		}
		filteredSearch := filterSearchWithContextVars([]*wtemplate.SearchValue{search[instrIndex]}, keywordMaps)
		if len(filteredSearch) > 0 {
			filteredRes[key] = filteredSearch
		}
	}
	if len(filteredRes) == 0 {
		return nil, false
	}
	return wkeyword.NewTemplateResults(&t.context, &t.templates, &filteredRes), true
}

// clone clones the template manager instance.
func (t *templateManager) clone() *templateManager {
	clonedResults := make(map[string]map[string][]*wtemplate.SearchValue)
	for k, vPsMap := range t.results {
		cVPsMap := make(map[string][]*wtemplate.SearchValue)
		for mapK, vPs := range vPsMap {
			var cVPs []*wtemplate.SearchValue
			for _, vP := range vPs {
				v := *vP
				cVPs = append(cVPs, &v)
			}
			cVPsMap[mapK] = cVPs
		}
		clonedResults[k] = cVPsMap
	}
	return &templateManager{
		context:   t.context,
		templates: t.templates,
		results:   clonedResults,
	}
}

// filterSearchWithContextVars filters search results with the provided context variables.
func filterSearchWithContextVars(cur []*wtemplate.SearchValue, contextVarsMap []wkeyword.KeywordsMap) []*wtemplate.SearchValue {
	var res []*wtemplate.SearchValue
	for _, search := range cur {
		addSearch := &wtemplate.SearchValue{Key: search.Key, Templ: search.Templ, Iter: []*wtemplate.SearchIteration{}}
		contextValue, ok := getContextVar(search.Key, contextVarsMap)
		for _, iter := range search.Iter {
			if ok && contextValue != iter.Found {
				continue
			}
			if len(iter.Values) == 0 {
				addSearch.Iter = append(addSearch.Iter, &wtemplate.SearchIteration{Found: iter.Found, Values: iter.Values})
				continue
			}
			addChildSearch := filterSearchWithContextVars(iter.Values, contextVarsMap)
			if len(addChildSearch) == 0 {
				return nil
			}
			addSearch.Iter = append(addSearch.Iter, &wtemplate.SearchIteration{Found: iter.Found, Values: addChildSearch})
		}
		if len(addSearch.Iter) > 0 {
			res = append(res, addSearch)
		}
	}
	return res
}

// getContextVar returns the value for some context variable.
func getContextVar(k string, contextVarsMaps []wkeyword.KeywordsMap) (string, bool) {
	for _, contextVarsMap := range contextVarsMaps {
		if val, typ, ok := contextVarsMap.Get(k); ok {
			// Resolve keyword value.
			switch typ {
			case wkeyword.KeywordTypeString:
				return val.(string), true
			}
		}
	}
	return "", false
}

// fillTemplatesMap fills the templates map with the transformation input.
func fillTemplatesMap(transformation *wyaml.BaseYAML) map[string]*wtemplate.Template {
	templatesMap := make(map[string]*wtemplate.Template)
	for templateName, templateStr := range transformation.Templates {
		t, err := template.Parse(templateName, templateStr)
		if err != nil {
			logrus.Fatalf("parsing template %q", templateName)
		}
		templatesMap[templateName] = t
	}
	return templatesMap
}
