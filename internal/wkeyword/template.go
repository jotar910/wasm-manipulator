package wkeyword

import (
	"github.com/sirupsen/logrus"
	"joao/wasm-manipulator/internal/wtemplate"
)

// TemplateKeyword is the keyword type for a template search.
type TemplateKeyword struct {
	Key     string
	Value   string
	results *TemplateResults
}

// NewTemplateKeyword is a constructor for TemplateKeyword.
func NewTemplateKeyword(k, v string, t *TemplateResults) *TemplateKeyword {
	return &TemplateKeyword{Key: k, Value: v, results: t}
}

// Context returns the context template of the search.
func (tk *TemplateKeyword) Context() *wtemplate.TemplateContext {
	ctxMap := *tk.results.context
	ctx, ok := ctxMap[tk.Key]
	if !ok {
		logrus.Fatalf("template context not found %q", tk.Key)
	}
	return ctx
}

// Result returns the result of the search.
func (tk *TemplateKeyword) Result() []*wtemplate.SearchValue {
	resultsMap := *tk.results.results
	result, ok := resultsMap[tk.Key]
	if !ok {
		logrus.Fatalf("template result not found %q", tk.Key)
	}
	return result
}

// TemplateResults contains the data resultant from a template search.
// it represents the keyword map for template values.
type TemplateResults struct {
	context   *map[string]*wtemplate.TemplateContext
	templates *map[string]*wtemplate.Template
	results   *map[string][]*wtemplate.SearchValue
}

// NewTemplateResults is a constructor for TemplateResults.
func NewTemplateResults(context *map[string]*wtemplate.TemplateContext, templates *map[string]*wtemplate.Template,
	results *map[string][]*wtemplate.SearchValue) *TemplateResults {
	return &TemplateResults{
		context:   context,
		templates: templates,
		results:   results,
	}
}

// Is returns the keyword type.
// KeywordTypeUnknown if not found.
func (t *TemplateResults) Is(k string) KeywordType {
	templatesMap := *t.templates
	if _, ok := templatesMap[k]; !ok {
		return KeywordTypeUnknown
	}

	templatesSearchMap := *t.results
	result, ok := templatesSearchMap[k]
	if !ok {
		return KeywordTypeUnknown
	}

	if len(result) == 0 || len(result[0].Iter) == 0 {
		return KeywordTypeUnknown
	}
	return KeywordTypeTemplate
}

// Get return the keyword if presented in the map.
func (t *TemplateResults) Get(k string) (interface{}, KeywordType, bool) {
	templatesMap := *t.templates
	if _, ok := templatesMap[k]; !ok {
		return nil, KeywordTypeUnknown, false
	}

	templatesSearchMap := *t.results
	result, ok := templatesSearchMap[k]
	if !ok {
		return nil, KeywordTypeUnknown, false
	}

	if len(result) == 0 || len(result[0].Iter) == 0 {
		return nil, KeywordTypeUnknown, false
	}
	return NewTemplateKeyword(k, result[0].Iter[0].Found, t), KeywordTypeTemplate, true
}

