package waspect

import (
	"fmt"
	"runtime"
	"sort"
	"sync"

	"github.com/sirupsen/logrus"

	"joao/wasm-manipulator/internal/wcode"
	"joao/wasm-manipulator/internal/wconfigs"
	"joao/wasm-manipulator/internal/wgenerator"
	"joao/wasm-manipulator/internal/wkeyword"
	"joao/wasm-manipulator/internal/wparser/lex"
	"joao/wasm-manipulator/internal/wparser/pointcut"
	"joao/wasm-manipulator/internal/wpointcut"
	"joao/wasm-manipulator/internal/wyaml"
)

// advice is the model that contains the input and the pointcut already parsed.
type advice struct {
	name     string
	input    wyaml.AdviceYAML
	pointcut *wpointcut.ParsedPointcut
	order    *int
	all      bool
	smart    bool
}

// Transformation is responsible for managing the transformation process.
type Transformation struct {
	code          string
	input         *wyaml.BaseYAML
	context       *wcode.ModuleContext
	globalZone    *contextVariablesZone
	functionsZone map[string]*contextVariablesZone
}

// NewTransformation is the constructor for Transformation.
func NewTransformation(code string, input *wyaml.BaseYAML) *Transformation {
	logrus.Infoln("Parsing input module")
	parser := wcode.NewCodeParser(code)
	entryBlock := parser.Parse()
	context := wcode.NewModuleContext(entryBlock)
	return &Transformation{
		code:          code,
		input:         input,
		context:       context,
		globalZone:    newContextVariablesZone(nil),
		functionsZone: make(map[string]*contextVariablesZone),
	}
}

// fillJoinPoints fills up the join-points list for each advice.
func (tf *Transformation) fillJoinPoints() []advice {
	logrus.Infoln("Parsing and filling up the advices")

	var advicesList []advice
	advices := tf.input.Aspects.Advices

	// Fill the join-points for each advice
	for adviceName, adviceValue := range filterAdvices(advices) {
		logFields := logrus.Fields{"name": adviceName}

		logrus.WithFields(logFields).Traceln("Parsing advice")

		if adviceValue.Pointcut == "" {
			logrus.WithFields(logFields).Warn("Pointcut not defined... Skipping advice")
			continue
		}

		// Parse pointcut input.
		pc, err := pointcut.ParseWithContext(adviceValue.Pointcut)
		if err != nil {
			logrus.Fatalf("parsing pointcut expression %q: %v", adviceValue.Pointcut, err)
		}

		// Transform pointcut into parsed expression.
		parsedPointcut := wpointcut.NewParsedPointcut(pc, tf.input)

		if !adviceValue.All {
			// Pointcut must be initiated before adding the global context (added functions and globals).
			parsedPointcut = parsedPointcut.Init(tf.context, tf.input)
		}

		advicesList = append(advicesList, advice{
			name:     adviceName,
			input:    adviceValue,
			pointcut: parsedPointcut,
			order:    adviceValue.Order,
			all:      adviceValue.All,
			smart:    adviceValue.Smart,
		})
	}

	if !wconfigs.Get().ConfigIgnoreOrder {
		sort.Slice(advicesList, func(i, j int) bool {
			if advicesList[j].order == nil {
				return true
			}
			if advicesList[i].order == nil {
				return false
			}
			return *advicesList[i].order < *advicesList[j].order
		})
	}

	return advicesList
}

// applyGlobalContextTransformations applies the global input context to the module.
func (tf *Transformation) applyGlobalContextTransformations() []string {
	aspects, zone := tf.input.Aspects, tf.globalZone

	logrus.WithFields(logrus.Fields{
		"globals":   len(aspects.Context.Variables),
		"functions": len(aspects.Context.Functions),
		"total":     len(aspects.Context.Variables) + len(aspects.Context.Functions),
	}).Infoln("Applying module modifications from global context")

	// Handle global variables
	for name, value := range aspects.Context.Variables {
		logFields := logrus.Fields{"name": name, "value": value}
		logrus.WithFields(logFields).Traceln("Adding global variable")

		globalDef, err := tf.context.AddGlobal(value)
		if err != nil {
			logrus.WithFields(logFields).Fatalf("adding global variable: %v", err)
		}
		tf.context.GlobalAlias[globalDef.Name] = name
		zone.AddVariable(name, globalDef.Name)
	}

	var fns []string

	// Handle global functions
	for name, function := range aspects.Context.Functions {
		isImported := function.Imported != nil
		isExported := function.Exported != nil

		logFields := logrus.Fields{"name": name, "imported": isImported, "exported": isExported}
		logrus.WithFields(logFields).Traceln("Adding new function")

		if isImported {
			// Create function element and add to module
			fnDef, err := tf.context.AddImportFunction(&function)
			if err != nil {
				logrus.WithFields(logFields).Fatalf("adding an imported input function: %v", err)
			}
			fns = append(fns, fnDef.Name)
			zone.AddFunction(name, newFunctionZone(fnDef.Index(tf.context), fnDef.Name))
			continue
		}

		// Create function element and add to module
		fnDef, err := tf.context.AddFunction(&function)
		if err != nil {
			logrus.Fatalf("adding a regular input function: %v", err)
		}
		fns = append(fns, fnDef.Name)

		// Set function as exported if that's the case
		if function.Exported != nil {
			fnDef, err = tf.context.AddExportFunction(&function, fnDef)
			if err != nil {
				logrus.Fatalf("adding an exported input function: %v", err)
			}
		}

		functionZone := newContextVariablesZone(tf.globalZone)

		// Add local variables to function
		for varName, varValue := range function.Variables {
			logFields := logrus.Fields{"function": name, "name": varName, "value": varValue}
			logrus.WithFields(logFields).Traceln("Adding local variable")

			// Adds local using the variable value.
			localDef, err := tf.context.AddLocal(varValue, fnDef)
			if err != nil {
				logrus.WithFields(logFields).Fatalf("adding local to function %s: %v", varName, err)
			}
			fnDef.Alias[localDef.Name] = varName
			functionZone.AddVariable(varName, localDef.Name)
		}

		// Save function parameters to zone.
		for i, fnParam := range fnDef.Parameters() {
			internalName := function.Args[i].Name
			functionZone.AddVariable(internalName, fnParam.Name)
			fnDef.Alias[fnParam.Name] = internalName
		}

		tf.functionsZone[name] = functionZone
		tf.context.FunctionAlias[fnDef.Name] = name

		// Save function on global zone.
		zone.AddFunction(name, newFunctionZone(fnDef.Index(tf.context), fnDef.Name))
	}
	return fns
}

// addStartFunctionCode adds the start function code.
func (tf *Transformation) addStartFunctionCode(code string) string {
	logrus.Infoln("Adding starting code")

	ctx := tf.context

	// Find start function.
	startFnDef, ok := ctx.StartFunction()
	if !ok {
		fnDef, err := ctx.AddFunction(new(wyaml.FunctionYAML))
		if err != nil {
			logrus.Fatalf("creating new function to be the starting function: %v", err)
		}
		startFnDef = fnDef
		if _, err := ctx.AddStartFunction(startFnDef); err != nil {
			logrus.Fatalf("adding start function instruction: %v", err)
		}
		ctx.SetStartFunction(startFnDef)
		logrus.WithFields(logrus.Fields{"function": startFnDef.Name}).
			Trace("Added new starting function")
	}

	// Adds code to start function.
	startFnDef.AddCode(code)
	return startFnDef.Name
}

// applyLocalContextTransformations applies the specific input context from an advice to the module.
func (tf *Transformation) applyLocalContextTransformations(advice advice, fns []string) {
	pointcut := advice.pointcut
	if !pointcut.Initiated {
		// It may be initialized when the "apply all" flag on the advice is set to false
		pointcut = pointcut.Init(tf.context, tf.input)
	}
	parsedContext := pointcut.Execute()
	joinPoints := parsedContext.All()

	logrus.WithFields(logrus.Fields{
		"advice": advice.name,
		"total":  len(joinPoints),
	}).Infof("Applying static transformations to join-points")

	wg := new(sync.WaitGroup)

	// Execute each joinpoint. Each one is referred to a function.
	// The advices are filtered to remove function added (just a backup condition)
	for _, joinPoint := range filterAddedFnsOnJoinPoints(advice.all, joinPoints, fns) {
		wg.Add(1)

		go func(joinPoint *wpointcut.JoinPoint) {
			var functionZone *contextVariablesZone
			fnDef := joinPoint.FuncDefinition()
			exportedName, ok := tf.context.AliasValue(fnDef.Name)
			if ok {
				functionZone = tf.functionsZone[exportedName]
			}
			if functionZone == nil {
				functionZone = newContextVariablesZone(tf.globalZone)
			}

			// Add local variables to function
			for name, value := range advice.input.Variables {
				logFields := logrus.Fields{"advice": advice.name, "function": fnDef.Name, "name": name, "value": value}
				logrus.WithFields(logFields).Traceln("Adding local variable")

				// Adds local using the variable value.
				localDef, err := tf.context.AddLocal(value, fnDef)
				if err != nil {
					logrus.WithFields(logFields).Fatalf("adding local to function %s: %v", name, err)
				}
				fnDef.Alias[localDef.Name] = name
				functionZone.AddVariable(name, localDef.Name)
			}

			// Apply transformations on join-points
			tf.applyJoinPointTransformations(joinPoint, parsedContext, fnDef,
				advice.input.Advice, advice.name, advice.smart,
				newPointcutParameters(fnDef, advice.pointcut.Params),
				newContextVariables(functionZone))

			wg.Done()
		}(joinPoint)
	}

	wg.Wait()
}

// applyJoinPointTransformations applies the join-point transformations to the module.
func (tf *Transformation) applyJoinPointTransformations(joinPoint *wpointcut.JoinPoint, ctx *wpointcut.PointcutContext,
	fnDef *wcode.FunctionDefinition, code, adviceName string, smartAdvice bool, keywordMappers ...wkeyword.KeywordsMap) {
	logFields := logrus.Fields{
		"advice":     adviceName,
		"join-point": joinPoint.String(),
		"index":      fnDef.Name,
	}
	logrus.WithFields(logFields).Traceln("Applying transformation to join-point")

	keywords := wkeyword.NewStringValuesMap()
	var staticMappers []wkeyword.KeywordsMap
	staticMappers = append(staticMappers, keywords)
	staticMappers = append(staticMappers, keywordMappers...)

	blocks := joinPoint.Blocks()
	for i, b := range blocks {
		mappers := append(staticMappers, b)
		keywords["this"] = joinPoint.InstrString(i)

		if templatesMapper, ok := ctx.Templates(fnDef.Name, i, mappers...); !ok {
			// Eager template filter fails
			logrus.WithFields(logFields).Infof("Join-point aborted due to unmatched template after filtering with context variables")
			return
		} else if templatesMapper != nil {
			mappers = append(mappers, templatesMapper)
		}

		parsedOutput := lex.Parse(code, tf.context.OrderMap(), mappers...)
		err := b.Apply(parsedOutput.Output, smartAdvice)
		if err != nil {
			logrus.Fatalf("applying join-point code: %v", err)
		}
	}
}

// applyTransformationsToAddedFunctions applies the transformations to the added functions
func (tf *Transformation) applyTransformationsToAddedFunctions(fns []string) {
	logrus.WithFields(logrus.Fields{
		"total": len(fns),
	}).Infof("Applying static transformations to added function")

	globalZone := newContextVariables(tf.globalZone)
	jpSearch := tf.context.InitSearch()
	fnsMap := make(map[string]struct{}, len(fns))

	for _, fn := range fns {
		fnsMap[fn] = struct{}{}
	}

	wg := new(sync.WaitGroup)
	semaphore := make(chan struct{}, runtime.NumCPU())
	for _, found := range jpSearch.Found() {
		wg.Add(1)
		go func(found *wcode.JoinPointBlock) {
			zone := globalZone
			name := found.Instr().(*wcode.Instruction).Child(0).String()
			if _, ok := fnsMap[name]; !ok {
				// Function was not added be the user.
				<-semaphore
				wg.Done()
				return
			}
			fnName, ok := tf.context.FunctionAlias[name]
			if !ok {
				// Function is not defined in input transformations.
				fnName = name
			}
			if localZone, ok := tf.functionsZone[fnName]; ok {
				zone = newContextVariables(localZone)
			}
			logrus.WithFields(logrus.Fields{"function": name, "alias": fnName}).Traceln("Executing static expressions on global function")
			code := wcode.FuncInstrsString(found.Instr())
			parsedOutput := lex.Parse(code, tf.context.OrderMap(), zone)
			if parsedOutput.Output != code {
				logrus.WithFields(logrus.Fields{"function": name, "alias": fnName}).Traceln("Applying modifications to global function")
				wcode.ReplaceBlocks([]*wcode.JoinPointBlock{found}, parsedOutput.Output)
			}
			<-semaphore
			wg.Done()
		}(found)
		semaphore <- struct{}{}
	}
	wg.Wait()
	close(semaphore)
}

// filterAdvices filters the advices map accordingly to the input configurations.
func filterAdvices(advices map[string]wyaml.AdviceYAML) map[string]wyaml.AdviceYAML {
	config := wconfigs.Get()
	includeLen, excludeLen := len(config.Include), len(config.Exclude)
	res := make(map[string]wyaml.AdviceYAML)

	if includeLen > 0 {
		for _, k := range config.Include {
			if advice, ok := advices[k]; ok {
				res[k] = advice
			}
		}
	} else {
		for k, v := range advices {
			res[k] = v
		}
	}

	if excludeLen > 0 {
		for _, k := range config.Exclude {
			delete(res, k)
		}
	}

	return res
}

// filterAddedFnsOnJoinPoints filters the joinpoints with the added functions.
// it depends on the 'All' state defined on the advice.
func filterAddedFnsOnJoinPoints(all bool, jps []*wpointcut.JoinPoint, fns []string) []*wpointcut.JoinPoint {
	if all {
		return jps
	}
	fnsMap := make(map[string]struct{})
	for _, fn := range fns {
		fnsMap[fn] = struct{}{}
	}
	var res []*wpointcut.JoinPoint
	for _, jp := range jps {
		if _, ok := fnsMap[jp.FuncDefinition().Name]; !ok {
			res = append(res, jp)
		}
	}
	return res
}

// TransformationResult is responsible to manage the response data from the transformation.
type TransformationResult struct {
	*wcode.ModuleContext
	tf *Transformation
}

// newTranformationResult is a constructor for TransformationResult.
func newTranformationResult(tf *Transformation) *TransformationResult {
	return &TransformationResult{tf.context, tf}
}

// GenerateJsData generates the javascript code.
func (tr *TransformationResult) GenerateJsData() (string, error) {
	var fns []wgenerator.JsFunctionDefinition
	for _, fn := range tr.Functions() {
		var fnName, opName string
		var fnScope wgenerator.JsScopeType
		var isExported bool
		if fn.Imported != nil {
			opName = fn.Imported.ExportName
			fnName = fmt.Sprintf("%s.%s", fn.Imported.ModuleName, fn.Imported.ExportName)
			fnScope = wgenerator.JsScopeTypeImported
		} else if fn.Exported != nil {
			opName = fn.Exported.ExportName
			fnName = fn.Exported.ExportName
			fnScope = wgenerator.JsScopeTypeExported
			isExported = true
		} else {
			continue
		}
		var fnArgs []wgenerator.JsArgumentDefinition
		for _, arg := range fn.Parameters() {
			fnArgs = append(fnArgs, wgenerator.NewJsArgumentDefinition(arg.TypeCodeOnFn(isExported), arg.IsPrimitive()))
		}
		returnsComposite := fn.Result != "" && !wcode.IsVarTypeStrPrimitive(fn.Result)
		fns = append(fns, wgenerator.NewJsFunctionDefinition(opName, fnName, fnScope, fnArgs, returnsComposite))
	}
	return wgenerator.GetJsCode(fns)
}

// Run executes the module transformation.
func Run(code string, input *wyaml.BaseYAML) (*TransformationResult, bool) {
	// Creating transformation manager.
	transformation := NewTransformation(code, input)

	// Fill the join-points for each advice
	advicesList := transformation.fillJoinPoints()

	if len(advicesList) == 0 && !wconfigs.Get().AllowEmpty {
		logrus.WithFields(logrus.Fields{"original": len(input.Aspects.Advices), "filtered": len(advicesList)}).
			Infoln("Aborted transformations because no advices were defined")
		return newTranformationResult(transformation), false
	}

	// Modify module accordingly to global context
	fns := transformation.applyGlobalContextTransformations()

	// Modify start function accordingly to definition
	if transformation.input.Aspects.Start != "" {
		fns = append(fns, transformation.addStartFunctionCode(transformation.input.Aspects.Start))
	}

	// Apply transformations on each advice.
	for _, v := range advicesList {
		transformation.applyLocalContextTransformations(v, fns)
	}

	// Resolve static expressions for the global functions.
	transformation.applyTransformationsToAddedFunctions(fns)
	// Apply runtime transformations.
	transformation.context.ApplyRuntimeTransformations()

	return newTranformationResult(transformation), true
}
