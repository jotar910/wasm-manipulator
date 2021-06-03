package lex

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"joao/wasm-manipulator/internal/wkeyword"
	"joao/wasm-manipulator/internal/wtemplate"
)

const (
	String      = "string"
	StringSlice = "string_slice"
	Search      = "template_search"
	Object      = "object"
)

const (
	True  = "true"
	False = "false"
	NaN   = "NaN"
)

const (
	MethodTypeString   = "string"
	MethodTypeType     = "type"
	MethodTypeMap      = "map"
	MethodTypeRepeat   = "repeat"
	MethodTypeJoin     = "join"
	MethodTypeSplit    = "split"
	MethodTypeCount    = "count"
	MethodTypeContains = "contains"
	MethodTypeAssert   = "assert"
	MethodTypeReplace  = "replace"
	MethodTypeRemove   = "remove"
	MethodTypeFilter   = "filter"
	MethodTypeSlice    = "slice"
	MethodTypeSplice   = "splice"
	MethodTypeSelect   = "select"
	MethodTypeOrder    = "order"
	MethodTypeReverse  = "reverse"
)

// Token represents a token for the parsing flux.
type Token interface {
	Execute(*ParsingContext, Receiver, Emitter)
}

// Parse parses a code input.
func Parse(input string, orderMap map[string]int, keywordMaps ...wkeyword.KeywordsMap) *ParseResult {
	return parse(input, newParsingContext(orderMap, keywordMaps))
}

// parse parses a code input.
func parse(input string, ctx *ParsingContext) *ParseResult {
	_, ch := Lex("Parser", input)
	parsedTokens := parseText(ctx, ch)
	receiver := newTextOnlyReceiver()
	go parsedTokens.Execute(ctx, receiver, nil)
	return newParseResult(ctx, input, receiver.Value())
}

// ParseResult contains the parse process result.
type ParseResult struct {
	Output string
	input  string
	ctx    *ParsingContext
}

// newParseResult is a constructor for ParseResult.
func newParseResult(ctx *ParsingContext, input string, output string) *ParseResult {
	return &ParseResult{
		Output: output,
		input:  input,
		ctx:    ctx,
	}
}

// ParsingContext contains all the context for the parsing process.
type ParsingContext struct {
	orderMap     map[string]int
	keywordsMaps []wkeyword.KeywordsMap
	mutex        *sync.Mutex
}

// newParsingContext is a constructor for ParsingContext.
func newParsingContext(orderMap map[string]int, keywordsMaps []wkeyword.KeywordsMap) *ParsingContext {
	return &ParsingContext{orderMap: orderMap, keywordsMaps: keywordsMaps, mutex: new(sync.Mutex)}
}

// shift removes the first keyword map from the parsing context and returns it.
func (r *ParsingContext) shift(v wkeyword.KeywordsMap) wkeyword.KeywordsMap {
	defer r.mutex.Unlock()
	if len(r.keywordsMaps) == 0 {
		return nil
	}
	r.mutex.Lock()
	for i, m := range r.keywordsMaps {
		if &m == &v {
			cur := r.keywordsMaps
			r.keywordsMaps = cur[:i]
			if i < len(r.keywordsMaps)-1 {
				r.keywordsMaps = append(r.keywordsMaps, cur[i+1:]...)
			}
			return cur[i]
		}
	}
	return nil
}

// unshift adds in the beginning a new keyword map on the parsing context.
func (r *ParsingContext) unshift(v wkeyword.KeywordsMap) {
	r.mutex.Lock()
	r.keywordsMaps = append([]wkeyword.KeywordsMap{v}, r.keywordsMaps...)
	r.mutex.Unlock()
}

// Tokens

// TextToken represents the token for text.
type TextToken struct {
	value string
}

// Execute executes the token functionality.
func (t TextToken) Execute(_ *ParsingContext, visitor Receiver, _ Emitter) {
	newTextEmitter(t.value).Accept(visitor)
}

// NumberToken represents the token for numbers.
type NumberToken struct {
	value float64
}

// newNumberToken is a constructor for NumberToken.
func newNumberToken(value string) (NumberToken, error) {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return NumberToken{}, err
	}
	return NumberToken{v}, nil
}

// Execute executes the token functionality.
func (t NumberToken) Execute(_ *ParsingContext, visitor Receiver, _ Emitter) {
	newTextEmitter(strconv.FormatFloat(t.value, 'f', -1, 64)).Accept(visitor)
}

// StringToken represents the token for strings.
type StringToken struct {
	blocks []Token
}

// Execute executes the token functionality.
func (t StringToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	totalBlocks := len(t.blocks)
	res := make([]string, totalBlocks)
	wg := new(sync.WaitGroup)
	wg.Add(totalBlocks)
	for i, block := range t.blocks {
		go func(i int, b Token) {
			receiver := newTextOnlyReceiver()
			go b.Execute(r, receiver, visited)
			res[i] = receiver.Value()
			wg.Done()
		}(i, block)
	}
	wg.Wait()

	newTextEmitter(strings.Join(res, "")).Accept(visitor)
}

// KeywordToken represents the token for keywords.
type KeywordToken struct {
	blocks []Token
}

// Execute executes the token functionality.
func (t *KeywordToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	var bridge EmitterReceiver = newEmitterReceiverBridge()
	accum := newAccumEmitterReceiverBridge()

	ch := make(chan bool)
	for _, block := range t.blocks {
		go executeMethodBlock(block, r, bridge, visited, ch)
		bridge.Accept(accum)
		if !<-ch {
			newTextEmitter("").Accept(visitor)
			return
		}
	}
	close(ch)
	accum.Accept(visitor)
}

// executeMethodBlock executes the method token.
func executeMethodBlock(b Token, r *ParsingContext, visitor Receiver, visited Emitter, ch chan<- bool) {
	defer func() {
		if r := recover(); r != nil {
			if r != PanicAssertMethod {
				logrus.Fatalf("method block panic: %v", r)
			}
			ch <- false
		}
	}()
	b.Execute(r, visitor, visited)
	ch <- true
}

// IdentifierToken represents the token for identifiers.
type IdentifierToken struct {
	name string
}

// newIdentifierToken is a constructor for IdentifierToken.
func newIdentifierToken(_ *ParsingContext, name string) *IdentifierToken {
	return &IdentifierToken{name: name}
}

// Execute executes the token functionality.
func (t IdentifierToken) Execute(r *ParsingContext, visitor Receiver, _ Emitter) {
	for _, keywordsMap := range r.keywordsMaps {
		if val, typ, ok := keywordsMap.Get(t.name); ok {
			// Resolve keyword value.
			switch typ {
			case wkeyword.KeywordTypeTemplate:
				templateVal, ok := val.(*wkeyword.TemplateKeyword)
				if !ok {
					break
				}
				ctx := templateVal.Context()
				visitor.VisitSearch(wtemplate.NewOutboundOp(templateVal.Result(), ctx.KnownVariables))
				return
			case wkeyword.KeywordTypeString:
				visitor.VisitString(val.(string))
				return
			case wkeyword.KeywordTypeObject:
				visitor.VisitObject(val.(wkeyword.Object))
				return
			default:
				logrus.Fatalf("unknown variable %q", t.name)
			}
		}
	}
	logrus.Fatalf("parsing code keyword variables: %s not found in scope", t.name)
}

// IdentifierPropertyToken represents the token for identifier properties.
type IdentifierPropertyToken struct {
	property string
}

// newObjectPropertyToken is a constructor for IdentifierPropertyToken.
func newObjectPropertyToken(property string) *IdentifierPropertyToken {
	return &IdentifierPropertyToken{property: property}
}

// Execute executes the token functionality.
func (t IdentifierPropertyToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	method := newObjectPropertyEmitter(r, t.property)
	go method.Accept(visitor)
	visited.Accept(method)
}

// ReferenceToken represents the token for references.
type ReferenceToken struct {
	name string
}

// newReferenceToken is a constructor for ReferenceToken.
func newReferenceToken(name string) *ReferenceToken {
	return &ReferenceToken{name: name}
}

// Execute executes the token functionality.
func (t ReferenceToken) Execute(_ *ParsingContext, visitor Receiver, _ Emitter) {
	newTextEmitter(t.name).Accept(visitor)
}

// MethodIndexToken represents the token for indexes.
type MethodIndexToken struct {
	index int
}

// Execute executes the token functionality.
func (t *MethodIndexToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	InvokeOneIntArgBasicMethodWithString(newMethodIndexEmitter, t.index, r, visitor, visited)
}

// MethodStringToken represents the token for method string cast.
type MethodStringToken struct {
}

// Execute executes the token functionality.
func (t *MethodStringToken) Execute(_ *ParsingContext, visitor Receiver, visited Emitter) {
	visitor.VisitString(ReadString(visited))
}

// MethodTypeToken represents the token for method type.
type MethodTypeToken struct {
}

// Execute executes the token functionality.
func (t *MethodTypeToken) Execute(_ *ParsingContext, visitor Receiver, visited Emitter) {
	method := newMethodTypeEmitter()
	go method.Accept(visitor)
	visited.Accept(method)
}

// MethodLambda represents the token for lambdas.
type MethodLambda struct {
	keys          []string
	argument      Token
	newKeywordMap wkeyword.KeywordsMap
}

// Execute executes the token functionality.
func (t *MethodLambda) Execute(r *ParsingContext, _ Receiver, visited Emitter) {
	if len(t.keys) < 1 && len(t.keys) > 2 {
		logrus.Fatalf("lambda has one or two argument")
	}
	readValue := ReadObject(visited)
	var values *wkeyword.KwArray
	switch v := readValue.(type) {
	case *wkeyword.KwArray:
		values = v
	default:
		values = wkeyword.NewKwArray([]interface{}{v})
	}
	if vLen := values.Len(); len(t.keys) > vLen {
		logrus.Fatalf("lambda expects %d arguments but only got %d", len(t.keys), vLen)
	}
	var valuesToAdd []wkeyword.KeyValueObject
	for i, k := range t.keys {
		valuesToAdd = append(valuesToAdd, wkeyword.NewKeyValueObject(k, values.Index(i)))
	}
	t.newKeywordMap = wkeyword.NewObjectValuesMap(valuesToAdd...)
	r.unshift(t.newKeywordMap)
}

func (t *MethodLambda) Clear(r *ParsingContext) {
	r.shift(t.newKeywordMap)
}

// MethodMapToken represents the token for method map.
type MethodMapToken struct {
	lambda MethodLambda
}

// Execute executes the token functionality.
func (t *MethodMapToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	method := newMethodMapEmitterReceiver(r, t.lambda)
	go method.Accept(visitor)
	visited.Accept(method)
}

// MethodFilterToken represents the token for method filter.
type MethodFilterToken struct {
	lambda MethodLambda
}

// Execute executes the token functionality.
func (t *MethodFilterToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	method := newMethodFilterEmitter(r, t.lambda)
	go method.Accept(visitor)
	visited.Accept(method)
}

// MethodAssertToken represents the token for method assert.
type MethodAssertToken struct {
	lambda MethodLambda
}

// Execute executes the token functionality.
func (t *MethodAssertToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	method := newMethodAssertEmitter(r, t.lambda)
	go visited.Accept(method)
	method.Accept(visitor)
}

// MethodRepeatToken represents the token for method repeat.
type MethodRepeatToken struct {
	times int
}

// Execute executes the token functionality.
func (t *MethodRepeatToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	method := newMethodRepeatEmitter(r, t.times)
	go method.Accept(visitor)
	visited.Accept(method)
}

// MethodJoinToken represents the token for method join.
type MethodJoinToken struct {
	argument Token
}

// Execute executes the token functionality.
func (t *MethodJoinToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	InvokeOneStringArgBasicMethodWithToken(newMethodJoinEmitter, t.argument, r, visitor, visited)
}

// MethodSplitToken represents the token for method split.
type MethodSplitToken struct {
	sep Token
}

// Execute executes the token functionality.
func (t *MethodSplitToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	InvokeOneStringArgBasicMethodWithToken(newMethodSplitEmitter, t.sep, r, visitor, visited)
}

// MethodSliceToken represents the token for method slice.
type MethodSliceToken struct {
	start *int
	end   *int
}

// Execute executes the token functionality.
func (t *MethodSliceToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	method := newMethodSliceEmitter(r, t.start, t.end)
	go method.Accept(visitor)
	visited.Accept(method)
}

// MethodSpliceToken represents the token for method splice.
type MethodSpliceToken struct {
	start *int
	end   *int
}

// Execute executes the token functionality.
func (t *MethodSpliceToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	method := newMethodSpliceEmitter(r, t.start, t.end)
	go method.Accept(visitor)
	visited.Accept(method)
}

// MethodCountToken represents the token for method count.
type MethodCountToken struct {
}

// Execute executes the token functionality.
func (t *MethodCountToken) Execute(_ *ParsingContext, visitor Receiver, visited Emitter) {
	method := newMethodCountEmitter()
	go method.Accept(visitor)
	visited.Accept(method)
}

// MethodContainsToken represents the token for method contains.
type MethodContainsToken struct {
	argument Token
}

// Execute executes the token functionality.
func (t *MethodContainsToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	InvokeOneStringArgBasicMethodWithToken(newMethodContainsEmitter, t.argument, r, visitor, visited)
}

// MethodReplaceToken represents the token for method replace.
type MethodReplaceToken struct {
	old Token
	new Token
}

// Execute executes the token functionality.
func (t *MethodReplaceToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	_, oldIsReference := t.old.(*ReferenceToken)
	_, newIsReference := t.new.(*ReferenceToken)
	oldEmitter := newContextBoolEmitterReceiver(newEmitterReceiverBridge(), oldIsReference)
	newEmitter := newContextBoolEmitterReceiver(newEmitterReceiverBridge(), newIsReference)
	go t.old.Execute(r, oldEmitter, visited)
	go t.new.Execute(r, newEmitter, visited)
	method := newMethodReplaceEmitter(r, oldEmitter, newEmitter)
	go method.Accept(visitor)
	visited.Accept(method)
}

// MethodRemoveToken represents the token for method remove.
type MethodRemoveToken struct {
	argument Token
}

// Execute executes the token functionality.
func (t *MethodRemoveToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	InvokeOneStringArgBasicMethodWithToken(newMethodRemoveEmitter, t.argument, r, visitor, visited)
}

// MethodSelectToken represents the token for method select.
type MethodSelectToken struct {
	variable string
}

// Execute executes the token functionality.
func (t *MethodSelectToken) Execute(r *ParsingContext, visitor Receiver, visited Emitter) {
	InvokeOneStringArgBasicMethodWithString(newMethodSelectEmitter, t.variable, r, visitor, visited)
}

// MethodOrderToken represents the token for method order.
type MethodOrderToken struct {
}

// Execute executes the token functionality.
func (t *MethodOrderToken) Execute(ctx *ParsingContext, visitor Receiver, visited Emitter) {
	method := newMethodOrderEmitter(ctx)
	go method.Accept(visitor)
	visited.Accept(method)
}

// MethodReverseToken represents the token for method reverse.
type MethodReverseToken struct {
}

// Execute executes the token functionality.
func (t *MethodReverseToken) Execute(ctx *ParsingContext, visitor Receiver, visited Emitter) {
	method := newMethodReverseEmitter(ctx)
	go method.Accept(visitor)
	visited.Accept(method)
}

// Parsers

// parseText parses text content.
func parseText(ctx *ParsingContext, ch <-chan Item) Token {
	res := StringToken{}
	for item := range ch {
		switch item.t {
		case ItemTypeText:
			res.blocks = append(res.blocks, TextToken{item.v})
		case ItemTypeLeftKeyword:
			res.blocks = append(res.blocks, parseKeyword(ctx, ch))
		case ItemTypeEOF:
			return res
		default:
			logrus.Fatalf("error parsing text: unable to parse item {%s   %s}", item.t, item.v)
		}
	}
	logrus.Fatalf("unexpected end while parsing text")
	return nil
}

// parseString parses string content.
func parseString(ctx *ParsingContext, ch <-chan Item) Token {
	res := StringToken{}
	for item := range ch {
		switch item.t {
		case ItemTypeString:
			res.blocks = append(res.blocks, TextToken{item.v})
		case ItemTypeLeftKeyword:
			res.blocks = append(res.blocks, parseKeyword(ctx, ch))
		case ItemTypeStringEnd:
			return res
		default:
			logrus.Fatalf("error parsing string: unable to parse item {%s   %s}", item.t, item.v)
		}
	}
	logrus.Fatalf("unexpected end while parsing string")
	return nil
}

// parseKeywordGroup parses a keyword group content.
func parseKeywordGroup(ctx *ParsingContext, ch <-chan Item, stop ItemType) Token {
	var res KeywordToken
	var expr []*TokenNode

	parseKeywordStart := func() {
		for item := range ch {
			switch item.t {
			case ItemTypeNegation:
				expr = append(expr, newTokenNode(TokenTypeNegation, nil))
			case ItemTypeLeftKeywordGroup:
				expr = append(expr, newTokenNode(TokenTypeValue, parseKeywordGroup(ctx, ch, ItemTypeRightKeywordGroup)))
				return
			case ItemTypeIdentifier:
				expr = append(expr, newTokenNode(TokenTypeValue, newIdentifierToken(ctx, item.v)))
				return
			case ItemTypeStringStart:
				expr = append(expr, newTokenNode(TokenTypeValue, parseString(ctx, ch)))
				return
			case ItemTypeNumber:
				v, err := newNumberToken(item.v)
				if err != nil {
					logrus.Fatalf("error parsing method args: creating number token: %v", err)
				}
				expr = append(expr, newTokenNode(TokenTypeValue, v))
				return
			default:
				logrus.Fatalf("invalid keyword start (error %s - %s)", item.t, item.v)
			}
		}
	}

	parseKeywordStart() // Parse keyword start

	for item := range ch {
		switch item.t {
		case stop:
			res.blocks = append(res.blocks, ParseExpr(expr))
			return &res
		case ItemTypeUnionKeyword:
			res.blocks = append(res.blocks, ParseExpr(expr))
			expr = []*TokenNode{}
			parseKeywordStart() // Parse next keyword start
		case ItemTypeObjectProperty:
			expr = append(expr, newTokenNode(TokenTypeMethod, newObjectPropertyToken(item.v)))
		case ItemTypeIndexStart:
			expr = append(expr, newTokenNode(TokenTypeMethod, parseIndex(ctx, ch)))
		case ItemTypeMethodStart:
			expr = append(expr, newTokenNode(TokenTypeMethod, parseMethod(ctx, ch)))
		default:
			if tokenType, ok := opItemTypeToTokenType(item.t); ok {
				expr = append(expr, newTokenNode(tokenType, nil))
				parseKeywordStart() // Parse next keyword start
			}
		}
	}
	return nil
}

// parseKeyword parses a keyword content.
func parseKeyword(ctx *ParsingContext, ch <-chan Item) Token {
	return parseKeywordGroup(ctx, ch, ItemTypeRightKeyword)
}

// parseIndex parses indexes content.
func parseIndex(_ *ParsingContext, ch <-chan Item) Token {
	res := &MethodIndexToken{}
	item := assertParse(ch, ItemTypeNumber, "value access by index: empty index")
	index, err := strconv.ParseInt(item.v, 10, 32)
	if err != nil {
		logrus.Fatalf("value access by index: first argument must be of type int")
	}
	res.index = int(index)
	assertParse(ch, ItemTypeIndexEnd, "value access by index: unclosed access")
	return res
}

// parseMethod parses a method content.
func parseMethod(ctx *ParsingContext, ch <-chan Item) Token {
	for item := range ch {
		switch item.t {
		case ItemTypeMethodName:
			return selectMethod(ctx, item.v, ch)
		default:
			logrus.Fatalf("error parsing method: unable to parse item {%s   %s}", item.t, item.v)
		}
	}
	logrus.Fatalf("unexpected end while parsing method")
	return nil
}

// selectMethod selects and parses the correct method.
func selectMethod(ctx *ParsingContext, name string, ch <-chan Item) Token {
	switch name {
	case MethodTypeString:
		return parseMethodString(ctx, ch)
	case MethodTypeType:
		return parseMethodType(ctx, ch)
	case MethodTypeMap:
		return parseMethodMap(ctx, ch)
	case MethodTypeRepeat:
		return parseMethodRepeat(ctx, ch)
	case MethodTypeJoin:
		return parseMethodJoin(ctx, ch)
	case MethodTypeSplit:
		return parseMethodSplit(ctx, ch)
	case MethodTypeCount:
		return parseMethodCount(ctx, ch)
	case MethodTypeContains:
		return parseMethodContains(ctx, ch)
	case MethodTypeAssert:
		return parseMethodAssert(ctx, ch)
	case MethodTypeReplace:
		return parseMethodReplace(ctx, ch)
	case MethodTypeRemove:
		return parseMethodRemove(ctx, ch)
	case MethodTypeFilter:
		return parseMethodFilter(ctx, ch)
	case MethodTypeSlice:
		return parseMethodSlice(ctx, ch)
	case MethodTypeSplice:
		return parseMethodSplice(ctx, ch)
	case MethodTypeSelect:
		return parseMethodSelect(ctx, ch)
	case MethodTypeOrder:
		return parseMethodOrder(ctx, ch)
	case MethodTypeReverse:
		return parseMethodReverse(ctx, ch)
	default:
		logrus.Fatalf("unknown method name %q", name)
		return nil
	}
}

// parseMethodString parses the method string cast content.
func parseMethodString(_ *ParsingContext, ch <-chan Item) Token {
	res := &MethodStringToken{}
	assertParse(ch, ItemTypeMethodEnd, "method string: must have no arguments")
	return res
}

// parseMethodType parses the method type content.
func parseMethodType(_ *ParsingContext, ch <-chan Item) Token {
	res := &MethodTypeToken{}
	assertParse(ch, ItemTypeMethodEnd, "method type: must have no arguments")
	return res
}

// parseMethodMap parses the method map content.
func parseMethodMap(ctx *ParsingContext, ch <-chan Item) Token {
	res := &MethodMapToken{}
	assertParse(ch, ItemTypeMethodArgStart, "method map: must have one argument")
	assertParse(ch, ItemTypeLambdaStart, "method map: argument must be a lambda")
	item := assertParse(ch, ItemTypeLambdaArg, "method map: lambda must have an argument a lambda")
	res.lambda.keys = []string{item.v}
	item = <-ch
	for item.t == ItemTypeLambdaArg {
		res.lambda.keys = append(res.lambda.keys, item.v)
		item = <-ch
	}
	if item.t != ItemTypeMethodArgStart {
		logrus.Fatalf(fmt.Sprintf("method map: empty lambda body: expected %s got %s", ItemTypeMethodArgStart, item.t))
	}
	res.lambda.argument = parseMethodArg(ctx, ch)
	assertParse(ch, ItemTypeLambdaEnd, "method map: unclosed lambda")
	assertParse(ch, ItemTypeMethodArgEnd, "method map: unclosed argument")
	assertParse(ch, ItemTypeMethodEnd, "method map: unclosed method")
	return res
}

// parseMethodFilter parses the method filter content.
func parseMethodFilter(ctx *ParsingContext, ch <-chan Item) Token {
	res := &MethodFilterToken{}
	assertParse(ch, ItemTypeMethodArgStart, "method filter: must have one argument")
	assertParse(ch, ItemTypeLambdaStart, "method filter: argument must be a lambda")
	item := assertParse(ch, ItemTypeLambdaArg, "method filter: lambda must have an argument a lambda")
	res.lambda.keys = []string{item.v}
	item = <-ch
	for item.t == ItemTypeLambdaArg {
		res.lambda.keys = append(res.lambda.keys, item.v)
		item = <-ch
	}
	if item.t != ItemTypeMethodArgStart {
		logrus.Fatalf(fmt.Sprintf("method filter: empty lambda body: expected %s got %s", ItemTypeMethodArgStart, item.t))
	}
	res.lambda.argument = parseMethodArg(ctx, ch)
	assertParse(ch, ItemTypeLambdaEnd, "method filter: unclosed lambda")
	assertParse(ch, ItemTypeMethodArgEnd, "method filter: unclosed argument")
	assertParse(ch, ItemTypeMethodEnd, "method filter: unclosed method")
	return res
}

// parseMethodSlice parses the method slice content.
func parseMethodSlice(_ *ParsingContext, ch <-chan Item) Token {
	res := &MethodSliceToken{}
	assertParse(ch, ItemTypeMethodArgStart, "method slice: must have at least one argument")
	item := assertParse(ch, ItemTypeNumber, "method slice: empty first argument")
	start, err := strconv.Atoi(item.v)
	if err != nil {
		logrus.Fatalf("method slice: first argument must be of type int")
	}
	res.start = &start
	assertParse(ch, ItemTypeMethodArgEnd, "method slice: unclosed first argument")
	item = <-ch
	if item.t == ItemTypeMethodEnd {
		return res
	}
	if item.t != ItemTypeMethodArgStart {
		logrus.Fatalf(fmt.Sprintf("method slice: unclosed method: expected %s got %s", ItemTypeMethodEnd, item.t))
	}
	item = assertParse(ch, ItemTypeNumber, "method slice: empty second argument")
	end, err := strconv.Atoi(item.v)
	if err != nil {
		logrus.Fatalf("method slice: second argument must be of type int")
	}
	res.end = &end
	assertParse(ch, ItemTypeMethodArgEnd, "method slice: unclosed first argument")
	assertParse(ch, ItemTypeMethodEnd, "method slice: unclosed method")
	return res
}

// parseMethodSplice parses the method splice content.
func parseMethodSplice(_ *ParsingContext, ch <-chan Item) Token {
	res := &MethodSpliceToken{}
	assertParse(ch, ItemTypeMethodArgStart, "method splice: must have at least one argument")
	item := assertParse(ch, ItemTypeNumber, "method splice: empty first argument")
	start, err := strconv.Atoi(item.v)
	if err != nil {
		logrus.Fatalf("method splice: first argument must be of type int")
	}
	res.start = &start
	assertParse(ch, ItemTypeMethodArgEnd, "method splice: unclosed first argument")
	item = <-ch
	if item.t == ItemTypeMethodEnd {
		return res
	}
	if item.t != ItemTypeMethodArgStart {
		logrus.Fatalf(fmt.Sprintf("method splice: unclosed method: expected %s got %s", ItemTypeMethodEnd, item.t))
	}
	item = assertParse(ch, ItemTypeNumber, "method splice: empty second argument")
	end, err := strconv.Atoi(item.v)
	if err != nil {
		logrus.Fatalf("method splice: second argument must be of type int")
	}
	res.end = &end
	assertParse(ch, ItemTypeMethodArgEnd, "method splice: unclosed first argument")
	assertParse(ch, ItemTypeMethodEnd, "method splice: unclosed method")
	return res
}

// parseMethodRepeat parses the method repeat content.
func parseMethodRepeat(_ *ParsingContext, ch <-chan Item) Token {
	res := &MethodRepeatToken{}
	assertParse(ch, ItemTypeMethodArgStart, "method repeat: must have one argument")
	item := assertParse(ch, ItemTypeNumber, "method repeat: first argument must be a number")
	times, err := strconv.ParseFloat(item.v, 64)
	if err != nil {
		logrus.Fatalf("method repeat: first argument is invalid")
	}
	res.times = int(times)
	assertParse(ch, ItemTypeMethodArgEnd, "method repeat: unclosed first argument")
	assertParse(ch, ItemTypeMethodEnd, "method repeat: unclosed method")
	return res
}

// parseMethodJoin parses the method join content.
func parseMethodJoin(ctx *ParsingContext, ch <-chan Item) Token {
	res := &MethodJoinToken{}
	assertParse(ch, ItemTypeMethodArgStart, "method join: must have one argument")
	res.argument = parseMethodArg(ctx, ch)
	assertParse(ch, ItemTypeMethodEnd, "method join: unclosed method")
	return res
}

// parseMethodSplit parses the method split content.
func parseMethodSplit(ctx *ParsingContext, ch <-chan Item) Token {
	res := &MethodSplitToken{}
	assertParse(ch, ItemTypeMethodArgStart, "method split: must have one argument")
	res.sep = parseMethodArg(ctx, ch)
	assertParse(ch, ItemTypeMethodEnd, "method split: unclosed method")
	return res
}

// parseMethodCount parses the method count content.
func parseMethodCount(_ *ParsingContext, ch <-chan Item) Token {
	res := &MethodCountToken{}
	assertParse(ch, ItemTypeMethodEnd, "method count: must have no arguments")
	return res
}

// parseMethodContains parses the method contains content.
func parseMethodContains(ctx *ParsingContext, ch <-chan Item) Token {
	res := &MethodContainsToken{}
	assertParse(ch, ItemTypeMethodArgStart, "method contains: must have one argument")
	res.argument = parseMethodArg(ctx, ch)
	assertParse(ch, ItemTypeMethodEnd, "method contains: unclosed method")
	return res
}

// parseMethodAssert parses the method assert content.
func parseMethodAssert(ctx *ParsingContext, ch <-chan Item) Token {
	res := &MethodAssertToken{}
	assertParse(ch, ItemTypeMethodArgStart, "method assert: must have one argument")
	assertParse(ch, ItemTypeLambdaStart, "method assert: argument must be a lambda")
	item := assertParse(ch, ItemTypeLambdaArg, "method assert: lambda must have an argument a lambda")
	res.lambda.keys = []string{item.v}
	item = <-ch
	for item.t == ItemTypeLambdaArg {
		res.lambda.keys = append(res.lambda.keys, item.v)
		item = <-ch
	}
	if item.t != ItemTypeMethodArgStart {
		logrus.Fatalf(fmt.Sprintf("method assert: empty lambda body: expected %s got %s", ItemTypeMethodArgStart, item.t))
	}
	res.lambda.argument = parseMethodArg(ctx, ch)
	assertParse(ch, ItemTypeLambdaEnd, "method assert: unclosed lambda")
	assertParse(ch, ItemTypeMethodArgEnd, "method assert: unclosed argument")
	assertParse(ch, ItemTypeMethodEnd, "method assert: unclosed method")
	return res
}

// parseMethodReplace parses the method replace content.
func parseMethodReplace(ctx *ParsingContext, ch <-chan Item) Token {
	res := &MethodReplaceToken{}
	assertParse(ch, ItemTypeMethodArgStart, "method replace: must have two arguments")
	switch item := <-ch; item.t {
	case ItemTypeIdentifier:
		res.old = newReferenceToken(item.v)
	case ItemTypeStringStart:
		res.old = parseString(ctx, ch)
	default:
		logrus.Fatalf(fmt.Sprintf("method replace: first argument must be a variable reference of template or a string but got %s", item.t))
	}
	assertParse(ch, ItemTypeMethodArgEnd, "method replace: unclosed first argument")
	assertParse(ch, ItemTypeMethodArgStart, "method replace: must declare the second argument")
	switch item := <-ch; item.t {
	case ItemTypeIdentifier:
		res.new = newReferenceToken(item.v)
	case ItemTypeStringStart:
		res.new = parseString(ctx, ch)
	default:
		logrus.Fatalf(fmt.Sprintf("method replace: second argument must be a variable reference of template or a string but got %s", item.t))
	}
	assertParse(ch, ItemTypeMethodArgEnd, "method replace: unclosed second argument")
	assertParse(ch, ItemTypeMethodEnd, "method replace: unclosed method")
	return res
}

// parseMethodRemove parses the method remove content.
func parseMethodRemove(ctx *ParsingContext, ch <-chan Item) Token {
	res := &MethodRemoveToken{}
	assertParse(ch, ItemTypeMethodArgStart, "method remove: must have one argument")
	switch item := <-ch; item.t {
	case ItemTypeIdentifier:
		res.argument = newReferenceToken(item.v)
	case ItemTypeStringStart:
		res.argument = parseString(ctx, ch)
	default:
		logrus.Fatalf(fmt.Sprintf("method remove: argument must be a variable reference of template or a string but got %s", item.t))
	}
	assertParse(ch, ItemTypeMethodArgEnd, "method remove: unclosed argument")
	assertParse(ch, ItemTypeMethodEnd, "method remove: unclosed method")
	return res
}

// parseMethodSelect parses the method select content.
func parseMethodSelect(_ *ParsingContext, ch <-chan Item) Token {
	res := &MethodSelectToken{}
	assertParse(ch, ItemTypeMethodArgStart, "method select: must have one argument")
	item := assertParse(ch, ItemTypeIdentifier, "method select: argument must be a variable reference of template")
	res.variable = item.v
	assertParse(ch, ItemTypeMethodArgEnd, "method select: unclosed argument")
	assertParse(ch, ItemTypeMethodEnd, "method select: unclosed method")
	return res
}

// parseMethodOrder parses the method order content.
func parseMethodOrder(_ *ParsingContext, ch <-chan Item) Token {
	res := &MethodOrderToken{}
	assertParse(ch, ItemTypeMethodEnd, "method order: must have no arguments")
	return res
}

// parseMethodReverse parses the method reverse content.
func parseMethodReverse(_ *ParsingContext, ch <-chan Item) Token {
	res := &MethodReverseToken{}
	assertParse(ch, ItemTypeMethodEnd, "method reverse: must have no arguments")
	return res
}

// parseMethodArg parses a method argument.
func parseMethodArg(ctx *ParsingContext, ch <-chan Item) Token {
	return parseKeywordGroup(ctx, ch, ItemTypeMethodArgEnd)
}

func assertParse(ch <-chan Item, t ItemType, m string) Item {
	v := <-ch
	if v.t == t {
		return v
	}
	logrus.Fatalf(fmt.Sprintf("%s: expected %s got %s (%s)", m, t, v.t, v.v))
	return Item{}
}

func opItemTypeToTokenType(t ItemType) (TokenNodeType, bool) {
	switch t {
	case ItemTypeAndKeyword:
		return TokenTypeOpAnd, true
	case ItemTypeOrKeyword:
		return TokenTypeOpOr, true
	case ItemTypeEqualKeyword:
		return TokenTypeOpEqual, true
	case ItemTypeNotEqualKeyword:
		return TokenTypeOpNotEqual, true
	case ItemTypeBitwiseRightKeyword:
		return TokenTypeOpBitwiseRight, true
	case ItemTypeBitwiseLeftKeyword:
		return TokenTypeOpBitwiseLeft, true
	case ItemTypeGreaterKeyword:
		return TokenTypeOpGreater, true
	case ItemTypeGreaterEqualKeyword:
		return TokenTypeOpGreaterEqual, true
	case ItemTypeLessKeyword:
		return TokenTypeOpLess, true
	case ItemTypeLessEqualKeyword:
		return TokenTypeOpLessEqual, true
	case ItemTypePlusKeyword:
		return TokenTypeOpPlus, true
	case ItemTypeMinusKeyword:
		return TokenTypeOpMinus, true
	case ItemTypeMultiplyKeyword:
		return TokenTypeOpMultiplication, true
	case ItemTypeDivisorKeyword:
		return TokenTypeOpDivision, true
	case ItemTypeRemainderKeyword:
		return TokenTypeOpRemainder, true
	default:
		return TokenTypeUnknown, false
	}
}
