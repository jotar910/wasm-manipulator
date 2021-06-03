package lex

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	LeftKeyword       = "%"
	RightKeyword      = "%"
	leftKeywordGroup  = "("
	rightKeywordGroup = ")"
	leftIndex         = "["
	rightIndex        = "]"
	leftMethodCall    = "("
	rightMethodCall   = ")"
	objectPropSplit   = "."
	unionKey          = ';'
	andKey            = '&'
	orKey             = '|'
	notKey            = '!'
	equalKey          = '='
	greaterKey        = '>'
	lessKey           = '<'
	plusKey           = '+'
	minusKey          = '-'
	multiplyKey       = '*'
	divisorKey        = '/'
	remainderKey      = '%'
)

const (
	eof rune = iota
)

// stateFn represents the state of the scanner
// as a function that returns the next state.
type stateFn func(*Lexer) stateFn

// Lexer holds the state of the scanner.
type Lexer struct {
	name  string
	input string
	start int
	pos   int
	width int
	items chan Item
}

// Lex is a constructor for Lexer.
// it starts running the lexer.
func Lex(name, input string) (*Lexer, <-chan Item) {
	l := &Lexer{
		name:  name,
		input: input,
		items: make(chan Item),
	}
	go l.run()
	return l, l.items
}

// run lexes the input by executing state functions until the state is nil.
func (l *Lexer) run() {
	runLexers(newLexText(), l)
	close(l.items)
}

// emit passes an item back to the client.
func (l *Lexer) emit(t ItemType) {
	l.items <- Item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

// current returns the current rune in the input.
func (l *Lexer) current() (rune rune) {
	if l.pos >= len(l.input) {
		return eof
	}
	rune, _ = utf8.DecodeRuneInString(l.input[l.pos-1:])
	return rune
}

// next returns the next rune in the input.
func (l *Lexer) next() (rune rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	rune, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return rune
}

// ignore skips over the pending input before this point.
func (l *Lexer) ignore() {
	l.start = l.pos
}

// skipSpaces skips over the spaces on the pe.ding input after this point.
func (l *Lexer) skipSpaces() rune {
	r := l.current()
	for r == ' ' {
		r = l.next()
	}
	return r
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *Lexer) backup() {
	l.pos -= l.width
}

// peek returns but does not consume the next rune in the input.
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// accept consumes the next rune
// if it's from the valid set.
func (l *Lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *Lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// errorf returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Item{ItemTypeError, fmt.Sprintf(format, args...)}
	return nil
}
