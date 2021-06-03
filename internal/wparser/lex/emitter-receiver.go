package lex

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"joao/wasm-manipulator/internal/wkeyword"
	"joao/wasm-manipulator/internal/wtemplate"
	"joao/wasm-manipulator/pkg/wutils"
)

const PanicAssertMethod string = "method assert stopped execution chain"

// Emitter is implemented by those who want to emit its value to a receiver.
type Emitter interface {
	Accept(Receiver)
}

// Receiver is implemented by those who wants to receive values from an emitter..
type Receiver interface {
	VisitString(string)
	VisitStringSlice([]string)
	VisitSearch(wtemplate.OutboundOperation)
	VisitObject(wkeyword.Object)
}

// EmitterReceiver is implemented by emitters and receivers,
type EmitterReceiver interface {
	Receiver
	Emitter
}

// ReceiverChannels contains all the channels needed from a receiver.
type ReceiverChannels struct {
	chString      chan string
	chStringSlice chan []string
	chSearch      chan wtemplate.OutboundOperation
	chObject      chan wkeyword.Object
}

// newReceiverChannels is a constructor for ReceiverChannels.
func newReceiverChannels() *ReceiverChannels {
	return &ReceiverChannels{
		chString:      make(chan string),
		chStringSlice: make(chan []string),
		chSearch:      make(chan wtemplate.OutboundOperation),
		chObject:      make(chan wkeyword.Object),
	}
}

// Close closes all the channels.
func (rc *ReceiverChannels) Close() {
	close(rc.chString)
	close(rc.chStringSlice)
	close(rc.chSearch)
	close(rc.chObject)
}

// TextEmitter is the emitter for string values.
type TextEmitter struct {
	value string
}

// newTextEmitter is a constructor for TextEmitter.
func newTextEmitter(v string) *TextEmitter {
	return &TextEmitter{value: v}
}

// Accept accepts and visits a receiver.
func (tr *TextEmitter) Accept(e Receiver) {
	e.VisitString(tr.value)
}

// TextSliceEmitter is the emitter for string values.
type TextSliceEmitter struct {
	value []string
}

// newTextSliceEmitter is a constructor for TextSliceEmitter.
func newTextSliceEmitter(v []string) *TextSliceEmitter {
	return &TextSliceEmitter{value: v}
}

// Accept accepts and visits a receiver.
func (tr *TextSliceEmitter) Accept(e Receiver) {
	e.VisitStringSlice(tr.value)
}

// SearchEmitter is the emitter for string values.
type SearchEmitter struct {
	value wtemplate.OutboundOperation
}

// newSearchEmitter is a constructor for SearchEmitter.
func newSearchEmitter(v wtemplate.OutboundOperation) *SearchEmitter {
	return &SearchEmitter{value: v}
}

// Accept accepts and visits a receiver.
func (tr *SearchEmitter) Accept(e Receiver) {
	e.VisitSearch(tr.value)
}

// ObjectEmitter is the emitter for string values.
type ObjectEmitter struct {
	value wkeyword.Object
}

// newObjectEmitter is a constructor for ObjectEmitter.
func newObjectEmitter(v wkeyword.Object) *ObjectEmitter {
	return &ObjectEmitter{value: v}
}

// Accept accepts and visits a receiver.
func (tr *ObjectEmitter) Accept(e Receiver) {
	e.VisitObject(tr.value)
}

// TextOnlyReceiver is a receiver that transforms any received value into a string value.
type TextOnlyReceiver struct {
	ch chan string
}

// newTextOnlyReceiver is a constructor for TextOnlyReceiver.
func newTextOnlyReceiver() *TextOnlyReceiver {
	return &TextOnlyReceiver{ch: make(chan string)}
}

// Value returns the received value as string.
func (to *TextOnlyReceiver) Value() string {
	return <-to.ch
}

// VisitString receives a string value.
func (to *TextOnlyReceiver) VisitString(v string) {
	to.ch <- v
}

// VisitStringSlice receives a string slice value.
func (to *TextOnlyReceiver) VisitStringSlice(v []string) {
	to.ch <- strings.Join(v, "")
}

// VisitSearch receives a search value.
func (to *TextOnlyReceiver) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("getting template value on text only receiver: %v", err)
	}
	to.ch <- value
}

// VisitObject receives an object value.
func (to *TextOnlyReceiver) VisitObject(v wkeyword.Object) {
	to.ch <- v.String()
}

// SliceOnlyReceiver is a receiver that transforms any received value into a string slice value.
type SliceOnlyReceiver struct {
	ch chan []string
}

// newSliceOnlyReceiver is a constructor for SliceOnlyReceiver.
func newSliceOnlyReceiver() *SliceOnlyReceiver {
	return &SliceOnlyReceiver{ch: make(chan []string)}
}

// Value returns the received value as string slice.
func (so *SliceOnlyReceiver) Value() []string {
	return <-so.ch
}

// VisitString receives a string value.
func (so *SliceOnlyReceiver) VisitString(v string) {
	so.ch <- []string{v}
}

// VisitStringSlice receives a string slice value.
func (so *SliceOnlyReceiver) VisitStringSlice(v []string) {
	so.ch <- v
}

// VisitSearch receives a search value.
func (so *SliceOnlyReceiver) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("getting template value on text only receiver: %v", err)
	}
	so.VisitString(value)
}

// VisitObject receives an object value.
func (so *SliceOnlyReceiver) VisitObject(v wkeyword.Object) {
	if !wkeyword.IsPrimitive(v) {
		so.VisitStringSlice(v.StringSlice())
	} else {
		so.VisitString(v.String())
	}
}

// ObjectOnlyReceiver is a receiver that transforms any received value into any slice value.
type ObjectOnlyReceiver struct {
	ch chan wkeyword.Object
}

// newObjectOnlyReceiver is a constructor for SliceOnlyReceiver.
func newObjectOnlyReceiver() *ObjectOnlyReceiver {
	return &ObjectOnlyReceiver{ch: make(chan wkeyword.Object)}
}

// Value returns the received value as string slice.
func (so *ObjectOnlyReceiver) Value() wkeyword.Object {
	return <-so.ch
}

// VisitString receives a string value.
func (so *ObjectOnlyReceiver) VisitString(v string) {
	so.ch <- wkeyword.NewKwPrimitive(v)
}

// VisitStringSlice receives a string slice value.
func (so *ObjectOnlyReceiver) VisitStringSlice(v []string) {
	so.ch <- wkeyword.NewKwArray(v)
}

// VisitSearch receives a search value.
func (so *ObjectOnlyReceiver) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("getting template value on text only receiver: %v", err)
	}
	so.VisitString(value)
}

// VisitObject receives an object value.
func (so *ObjectOnlyReceiver) VisitObject(v wkeyword.Object) {
	so.ch <- v
}

// NumberOnlyReceiver is a receiver that transforms any received value into a number value.
type NumberOnlyReceiver struct {
	ch chan float64
}

// newNumberOnlyReceiver is a constructor for NumberOnlyReceiver.
func newNumberOnlyReceiver() *NumberOnlyReceiver {
	return &NumberOnlyReceiver{ch: make(chan float64)}
}

func (tn *NumberOnlyReceiver) Value() float64 {
	return <-tn.ch
}

// VisitString receives a string value.
func (tn *NumberOnlyReceiver) VisitString(v string) {
	val, err := strconv.ParseFloat(v, 64)
	if err != nil {
		logrus.Fatalf("invalid expression: parsing float on number only receiver: %s must be a number", v)
	}
	tn.ch <- val

}

// VisitStringSlice receives a string slice value.
func (tn *NumberOnlyReceiver) VisitStringSlice(v []string) {
	logrus.Fatalf("invalid expression: %v must be a number", v)
}

// VisitSearch receives a search value.
func (tn *NumberOnlyReceiver) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("getting template value on number only receiver: %v", err)
	}
	tn.VisitString(value)
}

// VisitObject receives an object value.
func (tn *NumberOnlyReceiver) VisitObject(v wkeyword.Object) {
	res, err := strconv.ParseFloat(v.String(), 64)
	if err != nil {
		logrus.Fatal(err)
	}
	tn.ch <- res
}

// BooleanReceiver is a receiver that transforms any received value into a boolean value.
type BooleanReceiver struct {
	ch chan string
}

// newBooleanReceiver is a constructor for BooleanReceiver.
func newBooleanReceiver() *BooleanReceiver {
	return &BooleanReceiver{ch: make(chan string)}
}

func (br *BooleanReceiver) Value() string {
	return <-br.ch
}

// VisitString receives a string value.
func (br *BooleanReceiver) VisitString(v string) {
	br.emitBoolean(StringToBool(v))
}

// VisitStringSlice receives a string slice value.
func (br *BooleanReceiver) VisitStringSlice(vs []string) {
	br.emitBoolean(len(vs) > 0)
}

// VisitSearch receives a search value.
func (br *BooleanReceiver) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("error method call: boolean receiver failed while getting template value: %v", err)
	}
	br.VisitString(value)
}

// VisitObject receives an object value.
func (br *BooleanReceiver) VisitObject(v wkeyword.Object) {
	br.emitBoolean(v != nil && v.Len() > 0)
}

func (br *BooleanReceiver) emitBoolean(v bool) {
	if v {
		br.ch <- True
	} else {
		br.ch <- False
	}
}

// ContextBoolEmitterReceiver wraps an emitter/receiver while adding a boolean context.
type ContextBoolEmitterReceiver struct {
	wrapped EmitterReceiver
	value   bool
}

// newContextBoolEmitterReceiver is a constructor for ContextBoolEmitterReceiver.
func newContextBoolEmitterReceiver(w EmitterReceiver, v bool) *ContextBoolEmitterReceiver {
	return &ContextBoolEmitterReceiver{w, v}
}

// Accept accepts and visits a receiver.
func (cem *ContextBoolEmitterReceiver) Accept(e Receiver) {
	cem.wrapped.Accept(e)
}

// VisitString receives a string value.
func (cem *ContextBoolEmitterReceiver) VisitString(v string) {
	cem.wrapped.VisitString(v)
}

// VisitStringSlice receives a string slice value.
func (cem *ContextBoolEmitterReceiver) VisitStringSlice(v []string) {
	cem.wrapped.VisitStringSlice(v)
}

// VisitSearch receives a search value.
func (cem *ContextBoolEmitterReceiver) VisitSearch(v wtemplate.OutboundOperation) {
	cem.wrapped.VisitSearch(v)
}

// VisitObject receives an object value.
func (cem *ContextBoolEmitterReceiver) VisitObject(v wkeyword.Object) {
	cem.wrapped.VisitObject(v)
}

// EmptierReceiver is a receiver used to clear some emitter.
type EmptierReceiver struct {
	*ReceiverChannels
}

// newEmptierReceiver is a constructor for EmptierReceiver.
func newEmptierReceiver() *EmptierReceiver {
	return &EmptierReceiver{newReceiverChannels()}
}

// VisitString receives a string value.
func (er *EmptierReceiver) VisitString(v string) {
	// Empty by design.
}

// VisitStringSlice receives a string slice value.
func (er *EmptierReceiver) VisitStringSlice(v []string) {
	// Empty by design.
}

// VisitSearch receives a search value.
func (er *EmptierReceiver) VisitSearch(v wtemplate.OutboundOperation) {
	// Empty by design.
}

// VisitObject receives an object value.
func (er *EmptierReceiver) VisitObject(v wkeyword.Object) {
	// Empty by design.
}

// EmitterReceiverBridge is an emitter/receiver that allow bidirectional communication.
type EmitterReceiverBridge struct {
	*ReceiverChannels
}

// newEmitterReceiverBridge is a constructor for EmitterReceiverBridge.
func newEmitterReceiverBridge() *EmitterReceiverBridge {
	return &EmitterReceiverBridge{newReceiverChannels()}
}

// Accept accepts and visits a receiver.
func (erb *EmitterReceiverBridge) Accept(e Receiver) {
	select {
	case v := <-erb.chString:
		e.VisitString(v)
		return
	case v := <-erb.chStringSlice:
		e.VisitStringSlice(v)
		return
	case v := <-erb.chSearch:
		e.VisitSearch(v)
		return
	case v := <-erb.chObject:
		e.VisitObject(v)
		return
	}
}

// VisitString receives a string value.
func (erb *EmitterReceiverBridge) VisitString(v string) {
	erb.chString <- v
}

// VisitStringSlice receives a string slice value.
func (erb *EmitterReceiverBridge) VisitStringSlice(v []string) {
	erb.chStringSlice <- v
}

// VisitSearch receives a search value.
func (erb *EmitterReceiverBridge) VisitSearch(v wtemplate.OutboundOperation) {
	erb.chSearch <- v
}

// VisitObject receives an object value.
func (erb *EmitterReceiverBridge) VisitObject(v wkeyword.Object) {
	erb.chObject <- v
}

// AccumEmitterReceiverBridge is an emitter/receiver that aggregates a list of emitters.
// The result is the junction of the values emitted by the added emitters.
type AccumEmitterReceiverBridge struct {
	values []Emitter
}

// newAccumEmitterReceiverBridge is a constructor for AccumEmitterReceiverBridge.
func newAccumEmitterReceiverBridge() *AccumEmitterReceiverBridge {
	return &AccumEmitterReceiverBridge{}
}

func (ab *AccumEmitterReceiverBridge) Last() Emitter {
	if len(ab.values) == 0 {
		return nil
	}
	return ab.values[len(ab.values)-1]
}

// Accept accepts and visits a receiver.
func (ab *AccumEmitterReceiverBridge) Accept(e Receiver) {
	if len(ab.values) == 1 {
		ab.values[0].Accept(e)
		return
	}
	sb := new(strings.Builder)
	receiver := newTextOnlyReceiver()
	for _, v := range ab.values {
		go v.Accept(receiver)
		sb.WriteString(receiver.Value())

	}
	e.VisitString(sb.String())
}

// VisitString receives a string value.
func (ab *AccumEmitterReceiverBridge) VisitString(v string) {
	ab.values = append(ab.values, newTextEmitter(v))
}

// VisitStringSlice receives a string slice value.
func (ab *AccumEmitterReceiverBridge) VisitStringSlice(v []string) {
	ab.values = append(ab.values, newTextSliceEmitter(v))
}

// VisitSearch receives a search value.
func (ab *AccumEmitterReceiverBridge) VisitSearch(v wtemplate.OutboundOperation) {
	ab.values = append(ab.values, newSearchEmitter(v))
}

// VisitObject receives an object value.
func (ab *AccumEmitterReceiverBridge) VisitObject(v wkeyword.Object) {
	ab.values = append(ab.values, newObjectEmitter(v))
}

// NegationEmitterReceiver is an emitter/receiver that emits the negation of any received value.
type NegationEmitterReceiver struct {
	*EmitterReceiverBridge
}

// newNegationEmitterReceiver is a constructor for NegationEmitterReceiver.
func newNegationEmitterReceiver() *NegationEmitterReceiver {
	return &NegationEmitterReceiver{newEmitterReceiverBridge()}
}

// Accept accepts and visits a receiver.
func (ner *NegationEmitterReceiver) Accept(e Receiver) {
	e.VisitString(<-ner.chString)
}

// VisitString receives a string value.
func (ner *NegationEmitterReceiver) VisitString(v string) {
	ner.emitBoolean(len(v) == 0 || v == False)
}

// VisitStringSlice receives a string slice value.
func (ner *NegationEmitterReceiver) VisitStringSlice(vs []string) {
	ner.emitBoolean(len(vs) == 0)
}

// VisitSearch receives a search value.
func (ner *NegationEmitterReceiver) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("error method call: method 'negation' failed while getting template value: %v", err)
	}
	ner.VisitString(value)
}

// VisitObject receives an object value.
func (ner *NegationEmitterReceiver) VisitObject(v wkeyword.Object) {
	ner.emitBoolean(v == nil || v.Len() == 0)
}

// emitBoolean emits the boolean value as a string.
func (ner *NegationEmitterReceiver) emitBoolean(v bool) {
	ner.chString <- BoolToString(v)
}

// MethodTypeEmitter is the receiver/emitter for the method type.
// the method type emits the type of value received.
type MethodTypeEmitter struct {
	*EmitterReceiverBridge
}

// newMethodTypeEmitter is a constructor for MethodTypeEmitter.
func newMethodTypeEmitter() *MethodTypeEmitter {
	return &MethodTypeEmitter{newEmitterReceiverBridge()}
}

func (mm *MethodTypeEmitter) Accept(e Receiver) {
	e.VisitString(<-mm.chString)
}

// VisitString receives a string value.
func (mm *MethodTypeEmitter) VisitString(string) {
	mm.chString <- String
}

// VisitStringSlice receives a string slice value.
func (mm *MethodTypeEmitter) VisitStringSlice([]string) {
	mm.chString <- StringSlice
}

// VisitSearch receives a search value.
func (mm *MethodTypeEmitter) VisitSearch(wtemplate.OutboundOperation) {
	mm.chString <- Search
}

// VisitObject receives an object value.
func (mm *MethodTypeEmitter) VisitObject(wkeyword.Object) {
	mm.chString <- Object
}

// MethodIndexEmitter is the receiver/emitter for the index access.
// the index access emits the value that is in the provided index of the value received.
type MethodIndexEmitter struct {
	*EmitterReceiverBridge
	ctx   *ParsingContext
	index int
}

// newMethodIndexEmitter is a constructor for EmitterReceiver.
func newMethodIndexEmitter(ctx *ParsingContext, index int) EmitterReceiver {
	return &MethodIndexEmitter{newEmitterReceiverBridge(), ctx, index}
}

// Accept accepts and visits a receiver.
func (mi *MethodIndexEmitter) Accept(e Receiver) {
	select {
	case v := <-mi.chString:
		e.VisitString(v)
	case v := <-mi.chObject:
		e.VisitObject(v)
	}
}

// VisitString receives a string value.
func (mi *MethodIndexEmitter) VisitString(v string) {
	mi.chString <- string([]rune(v)[mi.index])
}

// VisitStringSlice receives a string slice value.
func (mi *MethodIndexEmitter) VisitStringSlice(v []string) {
	mi.chString <- v[mi.index]
}

// VisitSearch receives a search value.
func (mi *MethodIndexEmitter) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("error method call: access value by 'index' failed while getting template value: %v", err)
	}
	mi.VisitString(value)
}

// VisitObject receives an object value.
func (mi *MethodIndexEmitter) VisitObject(v wkeyword.Object) {
	mi.chObject <- v.Index(mi.index)
}

// MethodMapEmitter is the receiver/emitter for the method map.
// the method map maps the received value emitting it as a string slice.
type MethodMapEmitter struct {
	*EmitterReceiverBridge
	ctx    *ParsingContext
	lambda MethodLambda
	mutex  *sync.Mutex
}

// newMethodMapEmitterReceiver is a constructor for MethodMapEmitter.
func newMethodMapEmitterReceiver(ctx *ParsingContext, lambda MethodLambda) *MethodMapEmitter {
	return &MethodMapEmitter{newEmitterReceiverBridge(), ctx, lambda, new(sync.Mutex)}
}

// Accept accepts and visits a receiver.
func (mm *MethodMapEmitter) Accept(e Receiver) {
	select {
	case v := <-mm.chString:
		e.VisitString(v)
	case v := <-mm.chStringSlice:
		e.VisitStringSlice(v)
	case v := <-mm.chObject:
		e.VisitObject(v)
	}
}

// VisitString receives a string value.
func (mm *MethodMapEmitter) VisitString(v string) {
	receiver := newTextOnlyReceiver()
	clonedCtx := newParsingContext(mm.ctx.orderMap, append([]wkeyword.KeywordsMap{}, mm.ctx.keywordsMaps...))
	mm.lambda.Execute(clonedCtx, nil, newObjectEmitter(wkeyword.NewKwArray([]interface{}{v, "0"})))
	go mm.lambda.argument.Execute(clonedCtx, receiver, nil)
	res := receiver.Value()
	mm.lambda.Clear(clonedCtx)
	mm.chString <- res
}

// VisitStringSlice receives a string slice value.
func (mm *MethodMapEmitter) VisitStringSlice(vs []string) {
	var res []string
	receiver := newTextOnlyReceiver()
	for i, v := range vs {
		clonedCtx := newParsingContext(mm.ctx.orderMap, append([]wkeyword.KeywordsMap{}, mm.ctx.keywordsMaps...))
		mm.lambda.Execute(clonedCtx, nil, newObjectEmitter(wkeyword.NewKwArray([]interface{}{v, strconv.Itoa(i)})))
		go mm.lambda.argument.Execute(clonedCtx, receiver, nil)
		res = append(res, receiver.Value())
		mm.lambda.Clear(clonedCtx)
	}
	mm.chStringSlice <- res
}

// VisitSearch receives a search value.
func (mm *MethodMapEmitter) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("error method call: method 'map' failed while getting template value: %v", err)
	}
	mm.VisitStringSlice([]string{value})
}

// VisitObject receives an object value.
func (mm *MethodMapEmitter) VisitObject(v wkeyword.Object) {
	if wkeyword.IsPrimitive(v) {
		mm.VisitString(v.String())
		return
	}
	var res []wkeyword.Object
	receiver := newObjectOnlyReceiver()
	objectSlice := v.Slice()
	for i, v := range objectSlice {
		clonedCtx := newParsingContext(mm.ctx.orderMap, append([]wkeyword.KeywordsMap{}, mm.ctx.keywordsMaps...))
		mm.lambda.Execute(clonedCtx, nil, newObjectEmitter(wkeyword.NewKwArray([]interface{}{v, strconv.Itoa(i)})))
		go mm.lambda.argument.Execute(clonedCtx, receiver, nil)
		res = append(res, receiver.Value())
		mm.lambda.Clear(clonedCtx)
	}
	mm.chObject <- wkeyword.NewKwArray(res)
}

// MethodFilterEmitter is the receiver/emitter for the method filter.
// the method filter filters the received value emitting it as a string slice.
type MethodFilterEmitter struct {
	*EmitterReceiverBridge
	ctx    *ParsingContext
	lambda MethodLambda
}

// newMethodFilterEmitter is a constructor for MethodFilterEmitter.
func newMethodFilterEmitter(ctx *ParsingContext, lambda MethodLambda) *MethodFilterEmitter {
	return &MethodFilterEmitter{newEmitterReceiverBridge(), ctx, lambda}
}

// Accept accepts and visits a receiver.
func (mf *MethodFilterEmitter) Accept(e Receiver) {
	select {
	case v := <-mf.chString:
		e.VisitString(v)
	case v := <-mf.chStringSlice:
		e.VisitStringSlice(v)
	}
}

// VisitString receives a string value.
func (mf *MethodFilterEmitter) VisitString(v string) {
	sb := new(strings.Builder)
	booleanReceiver := newBooleanReceiver()
	for i, c := range v {
		clonedCtx := newParsingContext(mf.ctx.orderMap, append([]wkeyword.KeywordsMap{}, mf.ctx.keywordsMaps...))
		mf.lambda.Execute(mf.ctx, nil, newObjectEmitter(wkeyword.NewKwArray([]interface{}{v, strconv.Itoa(i)})))
		go mf.lambda.argument.Execute(clonedCtx, booleanReceiver, nil)
		if booleanReceiver.Value() == True {
			sb.WriteRune(c)
		}
		mf.lambda.Clear(clonedCtx)
	}
	mf.chString <- sb.String()
}

// VisitStringSlice receives a string slice value.
func (mf *MethodFilterEmitter) VisitStringSlice(vs []string) {
	var res []string
	booleanReceiver := newBooleanReceiver()
	for i, v := range vs {
		clonedCtx := newParsingContext(mf.ctx.orderMap, append([]wkeyword.KeywordsMap{}, mf.ctx.keywordsMaps...))
		mf.lambda.Execute(clonedCtx, nil, newObjectEmitter(wkeyword.NewKwArray([]interface{}{v, strconv.Itoa(i)})))
		go mf.lambda.argument.Execute(clonedCtx, booleanReceiver, nil)
		if booleanReceiver.Value() == True {
			res = append(res, v)
		}
		mf.lambda.Clear(clonedCtx)
	}
	mf.chStringSlice <- res
}

// VisitSearch receives a search value.
func (mf *MethodFilterEmitter) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("error method call: method 'filter' failed while getting template value: %v", err)
	}
	mf.VisitString(value)
}

// VisitObject receives an object value.
func (mf *MethodFilterEmitter) VisitObject(v wkeyword.Object) {
	if !wkeyword.IsPrimitive(v) {
		mf.VisitStringSlice(v.StringSlice())
	} else {
		mf.VisitString(v.String())
	}
}

// MethodAssertEmitter is the receiver/emitter for the method assert.
// the method asserts a condition and breaks the method chain if the condition is not met.
// panics PanicAssertMethod when the condition is not met,
type MethodAssertEmitter struct {
	*EmitterReceiverBridge
	ctx    *ParsingContext
	lambda MethodLambda
	failed chan bool
}

// newMethodAssertEmitter is a constructor for MethodAssertEmitter.
func newMethodAssertEmitter(ctx *ParsingContext, lambda MethodLambda) EmitterReceiver {
	return &MethodAssertEmitter{newEmitterReceiverBridge(), ctx, lambda, make(chan bool)}
}

// Accept accepts and visits a receiver.
func (ma *MethodAssertEmitter) Accept(e Receiver) {
	select {
	case v := <-ma.chString:
		ma.checkResult()
		e.VisitString(v)
	case v := <-ma.chStringSlice:
		ma.checkResult()
		e.VisitStringSlice(v)
	case v := <-ma.chSearch:
		ma.checkResult()
		e.VisitSearch(v)
	case v := <-ma.chObject:
		ma.checkResult()
		e.VisitObject(v)
	}
}

// VisitString receives a string value.
func (ma *MethodAssertEmitter) VisitString(v string) {
	ma.chString <- v
	ma.Assert(newTextEmitter(v))
}

// VisitStringSlice receives a string slice value.
func (ma *MethodAssertEmitter) VisitStringSlice(vs []string) {
	ma.chStringSlice <- vs
	ma.Assert(newTextSliceEmitter(vs))
}

// VisitSearch receives a search value.
func (ma *MethodAssertEmitter) VisitSearch(v wtemplate.OutboundOperation) {
	ma.chSearch <- v
	ma.Assert(newSearchEmitter(v))
}

// VisitObject receives an object value.
func (ma *MethodAssertEmitter) VisitObject(v wkeyword.Object) {
	ma.chObject <- v
	ma.Assert(newObjectEmitter(v))
}

// Assert tests the condition for the received value.
// the value is received using a custom emitter.
func (ma *MethodAssertEmitter) Assert(e Emitter) {
	receiver := newMethodAssertValidator(ma.ctx)
	clonedCtx := newParsingContext(ma.ctx.orderMap, append([]wkeyword.KeywordsMap{}, ma.ctx.keywordsMaps...))
	ma.lambda.Execute(clonedCtx, nil, e)
	defer ma.lambda.Clear(clonedCtx)
	go ma.lambda.argument.Execute(clonedCtx, receiver, nil)
	ma.failed <- !StringToBool(ReadString(receiver))
}

func (ma *MethodAssertEmitter) checkResult() {
	if <-ma.failed {
		panic(PanicAssertMethod)
	}
}

// MethodAssertValidator is responsible for testing the method assert condition.
type MethodAssertValidator struct {
	ctx *ParsingContext
	ch  chan string
}

// newMethodAssertValidator is a constructor for MethodAssertValidator.
func newMethodAssertValidator(ctx *ParsingContext) EmitterReceiver {
	return &MethodAssertValidator{ctx, make(chan string)}
}

// Accept accepts and visits a receiver.
func (mav *MethodAssertValidator) Accept(e Receiver) {
	e.VisitString(<-mav.ch)
}

// VisitString receives a string value.
func (mav *MethodAssertValidator) VisitString(v string) {
	mav.ch <- BoolToString(StringToBool(v))
}

// VisitStringSlice receives a string slice value.
func (mav *MethodAssertValidator) VisitStringSlice(vs []string) {
	for _, v := range vs {
		if !StringToBool(v) {
			mav.ch <- BoolToString(false)
			return
		}
	}
	mav.ch <- BoolToString(true)
}

// VisitSearch receives a search value.
func (mav *MethodAssertValidator) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("error method call: validating expression: method 'assert'' failed while getting template value: %v", err)
	}
	mav.VisitString(value)
}

// VisitObject receives an object value.
func (mav *MethodAssertValidator) VisitObject(v wkeyword.Object) {
	if !wkeyword.IsPrimitive(v) {
		mav.VisitStringSlice(v.StringSlice())
		return
	}
	mav.VisitString(v.String())
}

// MethodRepeatEmitter is the receiver/emitter for the method repeat.
// the method repeat repeats the received value emitting it as a string slice.
type MethodRepeatEmitter struct {
	*EmitterReceiverBridge
	times int
	ctx   *ParsingContext
}

// newMethodRepeatEmitter is a constructor for MethodRepeatEmitter.
func newMethodRepeatEmitter(ctx *ParsingContext, times int) *MethodRepeatEmitter {
	return &MethodRepeatEmitter{newEmitterReceiverBridge(), times, ctx}
}

// Accept accepts and visits a receiver.
func (mm *MethodRepeatEmitter) Accept(e Receiver) {
	e.VisitStringSlice(<-mm.chStringSlice)
}

// VisitString receives a string value.
func (mm *MethodRepeatEmitter) VisitString(v string) {
	mm.VisitStringSlice([]string{v})
}

// VisitStringSlice receives a string slice value.
func (mm *MethodRepeatEmitter) VisitStringSlice(v []string) {
	var res []string
	for i := 0; i < mm.times; i++ {
		res = append(res, v...)
	}
	mm.chStringSlice <- res
}

// VisitSearch receives a search value.
func (mm *MethodRepeatEmitter) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("error method call: method 'map' failed while getting template value: %v", err)
	}
	mm.VisitString(value)
}

// VisitObject receives an object value.
func (mm *MethodRepeatEmitter) VisitObject(v wkeyword.Object) {
	if wkeyword.IsArray(v) {
		mm.VisitStringSlice(v.StringSlice())
	} else {
		mm.VisitString(v.String())
	}
}

type MethodJoinEmitter struct {
	*EmitterReceiverBridge
	separator string
}

// newMethodJoinEmitter is a constructor for EmitterReceiver.
func newMethodJoinEmitter(_ *ParsingContext, separator string) EmitterReceiver {
	return &MethodJoinEmitter{newEmitterReceiverBridge(), separator}
}

// Accept accepts and visits a receiver.
func (mj *MethodJoinEmitter) Accept(e Receiver) {
	e.VisitString(<-mj.chString)
}

// VisitString receives a string value.
func (mj *MethodJoinEmitter) VisitString(v string) {
	mj.chString <- v
}

// VisitStringSlice receives a string slice value.
func (mj *MethodJoinEmitter) VisitStringSlice(v []string) {
	mj.chString <- strings.Join(v, mj.separator)
}

// VisitSearch receives a search value.
func (mj *MethodJoinEmitter) VisitSearch(wtemplate.OutboundOperation) {
	logrus.Fatalf("invalid method call: method 'join' cannot be called in template type values")
}

// VisitObject receives an object value.
func (mj *MethodJoinEmitter) VisitObject(v wkeyword.Object) {
	mj.VisitStringSlice(v.StringSlice())
}

// MethodSplitEmitter is the receiver/emitter for the method split.
// the method split splits the received value by some separator, emitting it as a string slice.
type MethodSplitEmitter struct {
	*EmitterReceiverBridge
	ctx       *ParsingContext
	separator string
}

// newMethodSplitEmitter is a constructor for EmitterReceiver.
func newMethodSplitEmitter(ctx *ParsingContext, separator string) EmitterReceiver {
	return &MethodSplitEmitter{newEmitterReceiverBridge(), ctx, separator}
}

// Accept accepts and visits a receiver.
func (ms *MethodSplitEmitter) Accept(e Receiver) {
	e.VisitStringSlice(<-ms.chStringSlice)
}

// VisitString receives a string value.
func (ms *MethodSplitEmitter) VisitString(v string) {
	ms.chStringSlice <- strings.Split(v, ms.separator)
}

// VisitStringSlice receives a string slice value.
func (ms *MethodSplitEmitter) VisitStringSlice([]string) {
	logrus.Fatalf("invalid method call: method 'split' cannot be called in string slice values")
}

// VisitSearch receives a search value.
func (ms *MethodSplitEmitter) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("error method call: method 'split' failed while getting template value: %v", err)
	}
	ms.VisitString(value)
}

// VisitObject receives an object value.
func (ms *MethodSplitEmitter) VisitObject(v wkeyword.Object) {
	if !wkeyword.IsPrimitive(v) {
		logrus.Fatalf("invalid method call: method 'split' cannot be called in arrays/maps")
	}
	ms.VisitString(v.String())
}

// MethodSliceEmitter is the receiver/emitter for the method slice.
// the method slice gets a limited sub-value of the received value.
// accepts either string and string slice.
type MethodSliceEmitter struct {
	*EmitterReceiverBridge
	ctx   *ParsingContext
	start *int
	end   *int
}

// newMethodSliceEmitter is a constructor for EmitterReceiver.
func newMethodSliceEmitter(ctx *ParsingContext, start, end *int) EmitterReceiver {
	return &MethodSliceEmitter{newEmitterReceiverBridge(), ctx, start, end}
}

// Accept accepts and visits a receiver.
func (ms *MethodSliceEmitter) Accept(e Receiver) {
	select {
	case v := <-ms.chString:
		e.VisitString(v)
	case v := <-ms.chStringSlice:
		e.VisitStringSlice(v)
	}
}

// VisitString receives a string value.
func (ms *MethodSliceEmitter) VisitString(v string) {
	if ms.end == nil {
		ms.chString <- v[*ms.start:]
		return
	}
	ms.chString <- v[*ms.start:*ms.end]
}

// VisitStringSlice receives a string slice value.
func (ms *MethodSliceEmitter) VisitStringSlice(v []string) {
	if ms.end == nil {
		ms.chStringSlice <- v[*ms.start:]
		return
	}
	ms.chStringSlice <- v[*ms.start:*ms.end]
}

// VisitSearch receives a search value.
func (ms *MethodSliceEmitter) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("error method call: method 'slice' failed while getting template value: %v", err)
	}
	ms.VisitString(value)
}

// VisitObject receives an object value.
func (ms *MethodSliceEmitter) VisitObject(v wkeyword.Object) {
	if !wkeyword.IsPrimitive(v) {
		ms.VisitStringSlice(v.StringSlice())
	} else {
		ms.VisitString(v.String())
	}
}

// MethodSpliceEmitter is the receiver/emitter for the method splice.
// the method splice gets the external sub-value of the received value for the limit defined.
// accepts either string and string slice.
type MethodSpliceEmitter struct {
	*EmitterReceiverBridge
	ctx   *ParsingContext
	start *int
	end   *int
}

// newMethodSpliceEmitter is a constructor for EmitterReceiver.
func newMethodSpliceEmitter(ctx *ParsingContext, start, end *int) EmitterReceiver {
	return &MethodSpliceEmitter{newEmitterReceiverBridge(), ctx, start, end}
}

// Accept accepts and visits a receiver.
func (ms *MethodSpliceEmitter) Accept(e Receiver) {
	select {
	case v := <-ms.chString:
		e.VisitString(v)
	case v := <-ms.chStringSlice:
		e.VisitStringSlice(v)
	}
}

// VisitString receives a string value.
func (ms *MethodSpliceEmitter) VisitString(v string) {
	if ms.end == nil {
		ms.chString <- v[:*ms.start]
		return
	}
	ms.chString <- v[:*ms.start] + v[*ms.end:]
}

// VisitStringSlice receives a string slice value.
func (ms *MethodSpliceEmitter) VisitStringSlice(v []string) {
	if ms.end == nil {
		ms.chStringSlice <- v[:*ms.start]
		return
	}
	ms.chStringSlice <- append(v[:*ms.start], v[*ms.end:]...)
}

// VisitSearch receives a search value.
func (ms *MethodSpliceEmitter) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("error method call: method 'splice' failed while getting template value: %v", err)
	}
	ms.VisitString(value)
}

// VisitObject receives an object value.
func (ms *MethodSpliceEmitter) VisitObject(v wkeyword.Object) {
	if !wkeyword.IsPrimitive(v) {
		ms.VisitStringSlice(v.StringSlice())
	} else {
		ms.VisitString(v.String())
	}
}

// MethodCountEmitter is the receiver/emitter for the method count.
// the method count counts characters/words of the received value.
type MethodCountEmitter struct {
	*EmitterReceiverBridge
}

// newMethodCountEmitter is a constructor for MethodCountEmitter.
func newMethodCountEmitter() *MethodCountEmitter {
	return &MethodCountEmitter{newEmitterReceiverBridge()}
}

// Accept accepts and visits a receiver.
func (mc *MethodCountEmitter) Accept(e Receiver) {
	e.VisitString(<-mc.chString)
}

// VisitString receives a string value.
func (mc *MethodCountEmitter) VisitString(v string) {
	mc.chString <- fmt.Sprintf("%d", len(v))
}

// VisitStringSlice receives a string slice value.
func (mc *MethodCountEmitter) VisitStringSlice(v []string) {
	mc.chString <- fmt.Sprintf("%d", len(v))
}

// VisitSearch receives a search value.
func (mc *MethodCountEmitter) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("error method call: method 'count' failed while getting template value: %v", err)
	}
	mc.VisitString(value)
}

// VisitObject receives an object value.
func (mc *MethodCountEmitter) VisitObject(v wkeyword.Object) {
	if !wkeyword.IsPrimitive(v) {
		mc.VisitStringSlice(v.StringSlice())
	} else {
		mc.VisitString(v.String())
	}
}

type MethodContainsEmitter struct {
	*EmitterReceiverBridge
	substr string
	ctx    *ParsingContext
}

// newMethodContainsEmitter is a constructor for EmitterReceiver.
func newMethodContainsEmitter(ctx *ParsingContext, substr string) EmitterReceiver {
	return &MethodContainsEmitter{newEmitterReceiverBridge(), substr, ctx}
}

// Accept accepts and visits a receiver.
func (mc *MethodContainsEmitter) Accept(e Receiver) {
	e.VisitString(<-mc.chString)
}

// VisitString receives a string value.
func (mc *MethodContainsEmitter) VisitString(v string) {
	mc.emitBoolean(strings.Contains(v, mc.substr))
}

// VisitStringSlice receives a string slice value.
func (mc *MethodContainsEmitter) VisitStringSlice(vs []string) {
	for _, v := range vs {
		if v == mc.substr {
			mc.emitBoolean(true)
			return
		}
	}
	mc.emitBoolean(false)
}

// VisitSearch receives a search value.
func (mc *MethodContainsEmitter) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("error method call: method 'contains' failed while getting template value: %v", err)
	}
	mc.VisitString(value)
}

// VisitObject receives an object value.
func (mc *MethodContainsEmitter) VisitObject(v wkeyword.Object) {
	if wkeyword.IsArray(v) {
		mc.VisitStringSlice(v.StringSlice())
	} else if wkeyword.IsObject(v) {
		mc.VisitStringSlice(v.(*wkeyword.KwObject).KeysSlice())
	} else {
		mc.VisitString(v.String())
	}
}

// emitBoolean emits the boolean value as a string.
func (mc *MethodContainsEmitter) emitBoolean(v bool) {
	if v {
		mc.chString <- True
	} else {
		mc.chString <- False
	}
}

// MethodRemoveEmitter is the receiver/emitter for the method remove.
// the method remove removes the sub-values from the received value that matches the defined value.
type MethodRemoveEmitter struct {
	*EmitterReceiverBridge
	value string
	ctx   *ParsingContext
}

// newMethodRemoveEmitter is a constructor for EmitterReceiver.
func newMethodRemoveEmitter(ctx *ParsingContext, value string) EmitterReceiver {
	return &MethodRemoveEmitter{newEmitterReceiverBridge(), value, ctx}
}

// Accept accepts and visits a receiver.
func (mr *MethodRemoveEmitter) Accept(e Receiver) {
	select {
	case v := <-mr.chString:
		e.VisitString(v)
	case v := <-mr.chStringSlice:
		e.VisitStringSlice(v)
	case v := <-mr.chSearch:
		e.VisitSearch(v)
	case v := <-mr.chObject:
		e.VisitObject(v)
	}
}

// VisitString receives a string value.
func (mr *MethodRemoveEmitter) VisitString(v string) {
	mr.chString <- strings.ReplaceAll(v, mr.value, "")
}

// VisitStringSlice receives a string slice value.
func (mr *MethodRemoveEmitter) VisitStringSlice(vs []string) {
	var res []string
	for _, v := range vs {
		if v != mr.value {
			res = append(res, v)
		}
	}
	mr.chStringSlice <- res
}

// VisitSearch receives a search value.
func (mr *MethodRemoveEmitter) VisitSearch(v wtemplate.OutboundOperation) {
	mr.chSearch <- wtemplate.NewRemoveOp(v, mr.value)
}

// VisitObject receives an object value.
func (mr *MethodRemoveEmitter) VisitObject(v wkeyword.Object) {
	if wkeyword.IsPrimitive(v) || wkeyword.IsNil(v) {
		mr.VisitString(v.String())
	} else if wkeyword.IsObject(v) {
		mr.chObject <- v.(*wkeyword.KwObject).RemoveProp(wutils.CapitalizeFirstLetter(mr.value))
	} else if wkeyword.IsArray(v) {
		mr.chObject <- v.(*wkeyword.KwArray).RemoveValue(mr.value)
	} else {
		mr.VisitStringSlice(v.StringSlice())
	}
}

// MethodReplaceEmitter is the receiver/emitter for the method replace.
// the method replace replaces the received value to another value.
type MethodReplaceEmitter struct {
	*EmitterReceiverBridge
	old Emitter
	new Emitter
	ctx *ParsingContext
}

// newMethodReplaceEmitter is a constructor for EmitterReceiver.
func newMethodReplaceEmitter(ctx *ParsingContext, old, new Emitter) EmitterReceiver {
	return &MethodReplaceEmitter{newEmitterReceiverBridge(), old, new, ctx}
}

// Accept accepts and visits a receiver.
func (mr *MethodReplaceEmitter) Accept(e Receiver) {
	select {
	case v := <-mr.chString:
		e.VisitString(v)
		return
	case v := <-mr.chStringSlice:
		e.VisitStringSlice(v)
		return
	case v := <-mr.chSearch:
		e.VisitSearch(v)
		return
	case v := <-mr.chObject:
		e.VisitObject(v)
		return
	}
}

// VisitString receives a string value.
func (mr *MethodReplaceEmitter) VisitString(v string) {
	oldValue := ReadString(mr.old)
	newValue := ReadString(mr.new)
	mr.chString <- strings.ReplaceAll(v, oldValue, newValue)
}

// VisitStringSlice receives a string slice value.
func (mr *MethodReplaceEmitter) VisitStringSlice(vs []string) {
	oldValue := ReadString(mr.old)
	newValue := ReadString(mr.new)
	mr.visitStringSlice(vs, oldValue, newValue)
}

// VisitSearch receives a search value.
func (mr *MethodReplaceEmitter) VisitSearch(v wtemplate.OutboundOperation) {
	oldContext := mr.old.(*ContextBoolEmitterReceiver)
	newContext := mr.new.(*ContextBoolEmitterReceiver)
	oldValue := ReadString(mr.old)
	newValue := ReadString(mr.new)
	mr.chSearch <- wtemplate.NewReplaceOp(v, wtemplate.NewReplaceOpArg(oldValue, oldContext.value), wtemplate.NewReplaceOpArg(newValue, newContext.value))
}

// VisitObject receives an object value.
func (mr *MethodReplaceEmitter) VisitObject(v wkeyword.Object) {
	oldValue := ReadString(mr.old)
	newValue := ReadString(mr.new)
	if wkeyword.IsPrimitive(v) || wkeyword.IsNil(v) {
		mr.chString <- strings.ReplaceAll(v.String(), oldValue, newValue)
	} else if wkeyword.IsObject(v) {
		mr.chObject <- v.(*wkeyword.KwObject).ReplacePropValue(wutils.CapitalizeFirstLetter(oldValue), newValue)
	} else if wkeyword.IsArray(v) {
		mr.chObject <- v.(*wkeyword.KwArray).ReplaceValue(oldValue, newValue)
	} else {
		mr.visitStringSlice(v.StringSlice(), oldValue, newValue)
	}
}

// visitStringSlice replaces an old value for a new value in a string slice.
func (mr *MethodReplaceEmitter) visitStringSlice(vs []string, oldValue, newValue string) {
	var res []string
	for _, v := range vs {
		if v == oldValue {
			res = append(res, newValue)
		} else {
			res = append(res, v)
		}
	}
	mr.chStringSlice <- res
}

// MethodSelectEmitter is the receiver/emitter for the method select.
// the method select selects the value of some sub-variable for the received search value.
type MethodSelectEmitter struct {
	*EmitterReceiverBridge
	reference string
}

// newMethodSelectEmitter is a constructor for EmitterReceiver.
func newMethodSelectEmitter(_ *ParsingContext, reference string) EmitterReceiver {
	return &MethodSelectEmitter{newEmitterReceiverBridge(), reference}
}

// Accept accepts and visits a receiver.
func (mm *MethodSelectEmitter) Accept(e Receiver) {
	e.VisitSearch(<-mm.chSearch)
}

// VisitString receives a string value.
func (mm *MethodSelectEmitter) VisitString(string) {
	logrus.Fatalf("invalid method call: method 'select' cannot be called in string type values")
}

// VisitStringSlice receives a string slice value.
func (mm *MethodSelectEmitter) VisitStringSlice([]string) {
	logrus.Fatalf("invalid method call: method 'select' cannot be called in string slice type values")
}

// VisitSearch receives a search value.
func (mm *MethodSelectEmitter) VisitSearch(v wtemplate.OutboundOperation) {
	mm.chSearch <- wtemplate.NewSelectOp(v, mm.reference)
}

// VisitObject receives an object value.
func (mm *MethodSelectEmitter) VisitObject(wkeyword.Object) {
	logrus.Fatalf("invalid method call: method 'select' cannot be called in object type values")
}

// ObjectPropertyEmitter is the receiver/emitter for ana object property access.
type ObjectPropertyEmitter struct {
	*EmitterReceiverBridge
	property string
}

// newObjectPropertyEmitter is a constructor for EmitterReceiver.
func newObjectPropertyEmitter(_ *ParsingContext, property string) EmitterReceiver {
	return &ObjectPropertyEmitter{newEmitterReceiverBridge(), property}
}

// Accept accepts and visits a receiver.
func (mop *ObjectPropertyEmitter) Accept(e Receiver) {
	e.VisitObject(<-mop.chObject)
}

// VisitString receives a string value.
func (mop *ObjectPropertyEmitter) VisitString(string) {
	logrus.Fatalf("invalid method call: cannot access properties in string type values")
}

// VisitStringSlice receives a string slice value.
func (mop *ObjectPropertyEmitter) VisitStringSlice([]string) {
	logrus.Fatalf("invalid method call: cannot access properties in string slice type values")
}

// VisitSearch receives a search value.
func (mop *ObjectPropertyEmitter) VisitSearch(wtemplate.OutboundOperation) {
	logrus.Fatalf("invalid method call: cannot access properties in search type values")
}

// VisitObject receives an object value.
func (mop *ObjectPropertyEmitter) VisitObject(v wkeyword.Object) {
	mop.chObject <- v.Prop(wutils.CapitalizeFirstLetter(mop.property))
}

// MethodOrderEmitter returns the function order for some function name.
type MethodOrderEmitter struct {
	*EmitterReceiverBridge
	ctx *ParsingContext
}

// newMethodOrderEmitter is a constructor for MethodOrderEmitter.
func newMethodOrderEmitter(ctx *ParsingContext) *MethodOrderEmitter {
	return &MethodOrderEmitter{newEmitterReceiverBridge(), ctx}
}

// Accept accepts and visits a receiver.
func (mo *MethodOrderEmitter) Accept(e Receiver) {
	e.VisitString(<-mo.chString)
}

// VisitString receives a string value.
func (mo *MethodOrderEmitter) VisitString(v string) {
	if order, ok := mo.ctx.orderMap[v]; ok {
		mo.chString <- strconv.Itoa(order + 1)
	} else {
		mo.chString <- ""
	}
}

// VisitStringSlice receives a string slice value.
func (mo *MethodOrderEmitter) VisitStringSlice(v []string) {
	mo.VisitString(strings.Join(v, ""))
}

// VisitSearch receives a search value.
func (mo *MethodOrderEmitter) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("getting template value on method order: %v", err)
	}
	mo.VisitString(value)
}

// VisitObject receives an object value.
func (mo *MethodOrderEmitter) VisitObject(v wkeyword.Object) {
	mo.VisitString(v.String())
}

// MethodReverseEmitter returns the reverse for some function name.
type MethodReverseEmitter struct {
	*EmitterReceiverBridge
	ctx *ParsingContext
}

// newMethodOrderEmitter is a constructor for MethodReverseEmitter.
func newMethodReverseEmitter(ctx *ParsingContext) *MethodReverseEmitter {
	return &MethodReverseEmitter{newEmitterReceiverBridge(), ctx}
}

// Accept accepts and visits a receiver.
func (mr *MethodReverseEmitter) Accept(e Receiver) {
	select {
	case v := <-mr.chString:
		e.VisitString(v)
	case v := <-mr.chStringSlice:
		e.VisitStringSlice(v)
	case v := <-mr.chSearch:
		e.VisitSearch(v)
	case v := <-mr.chObject:
		e.VisitObject(v)
	}
}

// VisitString receives a string value.
func (mr *MethodReverseEmitter) VisitString(v string) {
	runes := []rune(v)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	mr.chString <- string(runes)
}

// VisitStringSlice receives a string slice value.
func (mr *MethodReverseEmitter) VisitStringSlice(v []string) {
	res := make([]string, len(v))
	for i, j := 0, len(v)-1; j > -1; i, j = i+1, j-1 {
		res[j] = v[i]
	}
	mr.chStringSlice <- res
}

// VisitSearch receives a search value.
func (mr *MethodReverseEmitter) VisitSearch(v wtemplate.OutboundOperation) {
	value, err := wtemplate.GetValue(v)
	if err != nil {
		logrus.Fatalf("error method call: method 'reverse' failed while getting template value: %v", err)
	}
	mr.VisitString(value)
}

// VisitObject receives an object value.
func (mr *MethodReverseEmitter) VisitObject(v wkeyword.Object) {
	if !wkeyword.IsPrimitive(v) {
		mr.VisitStringSlice(v.StringSlice())
	} else {
		mr.VisitString(v.String())
	}
}
