package lex

import (
	"unicode"

	"joao/wasm-manipulator/internal/wkeyword"
)

// LEX

// isSpace returns if the character rune is space based.
func isSpace(s rune) bool {
	return unicode.IsSpace(s)
}

// PARSER

// OneStringArgBasicMethodFn is an utility type for token methods that receives an argument as string.
type OneStringArgBasicMethodFn func(*ParsingContext, string) EmitterReceiver

// OneIntArgBasicMethodFn is an utility type for token methods that receives an argument as int.
type OneIntArgBasicMethodFn func(*ParsingContext, int) EmitterReceiver

// InvokeOneStringArgBasicMethodWithToken invoke tokens that have a string as an argument.
// the string argument comes from emitted value of the argument token.
func InvokeOneStringArgBasicMethodWithToken(fn OneStringArgBasicMethodFn, argument Token, r *ParsingContext, visitor Receiver, visited Emitter) {
	receiver := newTextOnlyReceiver()
	go argument.Execute(r, receiver, nil)
	method := fn(r, <-receiver.ch)
	go method.Accept(visitor)
	visited.Accept(method)
}

// InvokeOneStringArgBasicMethodWithString invoke tokens that have a string as an argument.
func InvokeOneStringArgBasicMethodWithString(fn OneStringArgBasicMethodFn, argument string, r *ParsingContext, visitor Receiver, visited Emitter) {
	method := fn(r, argument)
	go method.Accept(visitor)
	visited.Accept(method)
}

// InvokeOneIntArgBasicMethodWithString invoke tokens that have an int as an argument.
func InvokeOneIntArgBasicMethodWithString(fn OneIntArgBasicMethodFn, argument int, r *ParsingContext, visitor Receiver, visited Emitter) {
	method := fn(r, argument)
	go method.Accept(visitor)
	visited.Accept(method)
}

// ReadString reads the string value from an emitter.
func ReadString(e Emitter) string {
	receiver := newTextOnlyReceiver()
	go e.Accept(receiver)
	return receiver.Value()
}

// ReadString reads the string slice value from an emitter.
func ReadSlice(e Emitter) []string {
	receiver := newSliceOnlyReceiver()
	go e.Accept(receiver)
	return receiver.Value()
}

// ReadObject reads any object value from an emitter.
func ReadObject(e Emitter) wkeyword.Object {
	receiver := newObjectOnlyReceiver()
	go e.Accept(receiver)
	return receiver.Value()
}

// StringToBool converts a string value to its adjacent boolean.
func StringToBool(v string) bool {
	return len(v) > 0 && v != False
}

// BoolToString converts a boolean value to its adjacent string.
func BoolToString(v bool) string {
	if v {
		return True
	}
	return False
}
