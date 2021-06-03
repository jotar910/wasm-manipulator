package wyaml

// BaseYAML is the base model for the transformation language.
type BaseYAML struct {
	Templates map[string]string
	Pointcuts map[string]string
	Aspects   AspectYAML
}

// AspectYAML contains the aspect data.
type AspectYAML struct {
	Start   string
	Context ContextYAML
	Advices map[string]AdviceYAML
}

// AdviceYAML is defined in the aspect and contains the advice data.
type AdviceYAML struct {
	Variables map[string]string
	Pointcut  string
	Advice    string
	Order     *int
	All       bool
	Smart     bool
}

// ContextYAML contains the context data.
// allows the interaction with new variables and functions
// can be defined as global or local to the aspect.
type ContextYAML struct {
	Variables map[string]string
	Functions map[string]FunctionYAML
}

// CountImportedFunctions counts the number of imported function in context.
func (c *ContextYAML) CountImportedFunctions() int {
	var count int
	for _, function := range c.Functions {
		if function.Imported != nil {
			count++
		}
	}
	return count
}

// FunctionYAML contains the function definition data.
type FunctionYAML struct {
	Variables map[string]string
	Args      []FunctionArgYAML
	Result    string
	Code      string
	Imported  *FunctionImportYAML `yaml:",omitempty"`
	Exported  *string             `yaml:",omitempty"`
}

// FunctionArgYAML represents a function argument.
type FunctionArgYAML struct {
	Name string
	Type string
}

// FunctionImportYAML represents an imported function.
type FunctionImportYAML struct {
	Module string
	Field  string
}
