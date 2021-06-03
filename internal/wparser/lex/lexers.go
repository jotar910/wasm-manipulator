package lex

import (
	"strings"

	"joao/wasm-manipulator/pkg/wutils"
)

// Lexers is implemented by a lexer based structure.
// each lexer may expect a different state and act accordingly to the input.
type Lexers interface {
	upgrade(l *Lexer) stateFn
	execute(l *Lexer) stateFn
	downgrade(l *Lexer) stateFn
}

// LexText is a lexer for the text content.
type LexText struct {
}

// newLexText is a constructor for LexText.
func newLexText() stateFn {
	return LexText{}.upgrade
}

// upgrade executes the initial tasks for the lexer.
func (lt LexText) upgrade(*Lexer) stateFn {
	return lt.execute
}

// execute executes the main tasks for the lexer.
func (lt LexText) execute(l *Lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], LeftKeyword) {
			if l.pos > l.start {
				l.emit(ItemTypeText)
			}
			runLexers(newLexKeyword(), l)
			continue
		}
		if l.next() == eof {
			break
		}
	}
	if l.pos > l.start {
		l.emit(ItemTypeText)
	}
	l.emit(ItemTypeEOF)
	return lt.downgrade
}

// downgrade executes the ending tasks for the lexer.
func (lt LexText) downgrade(*Lexer) stateFn {
	return nil
}

// LexKeyword is a lexer for the keyword content.
type LexKeyword struct {
	endKeyword string
}

// newLexKeyword is a constructor for LexKeyword.
func newLexKeyword() stateFn {
	return LexKeyword{endKeyword: RightKeyword}.upgrade
}

// upgrade executes the initial tasks for the lexer.
func (lk LexKeyword) upgrade(l *Lexer) stateFn {
	l.pos += len(LeftKeyword)
	l.emit(ItemTypeLeftKeyword)
	return lk.execute
}

// execute executes the main tasks for the lexer.
func (lk LexKeyword) execute(l *Lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], lk.endKeyword) {
			if l.pos > l.start {
				l.emit(ItemTypeInsideAction)
			}
			return lk.downgrade
		}
		switch r := l.next(); {
		case r == eof:
			return l.errorf("unclosed action on keyword on index %d", l.pos-1)
		case isSpace(r) || r == '\n':
			l.ignore()
		case wutils.IsAlphaNumeric(r) || r == notKey:
			l.backup()
			runLexers(newLexIdentifier(), l)
		case r == '"', r == '\'', r == '`':
			runLexers(newLexQuote(r), l)
		case r == ':':
			l.backup()
			runLexers(newLexMethod(), l)
		case string(r) == leftIndex:
			l.backup()
			runLexers(newLexIndex(), l)
		case r == unionKey:
			l.backup()
			runLexers(newLexUnion(), l)
		case r == '-' || '0' <= r && r <= '9':
			l.backup()
			runLexers(newLexNumber(), l)
		case string(r) == leftKeywordGroup:
			runLexers(newLexKeywordGroup(), l)
		default:
			return l.errorf("unknown character inside keyword at index (%d) %q", l.pos, r)
		}
	}
}

// downgrade executes the ending tasks for the lexer.
func (lk LexKeyword) downgrade(l *Lexer) stateFn {
	l.pos += len(RightKeyword)
	l.emit(ItemTypeRightKeyword)
	return nil
}

// LexKeywordGroup is a lexer for the keyword group content.
type LexKeywordGroup struct {
}

// newLexKeywordGroup is a constructor for LexKeywordGroup.
func newLexKeywordGroup() stateFn {
	return LexKeywordGroup{}.upgrade
}

// upgrade executes the initial tasks for the lexer.
func (lg LexKeywordGroup) upgrade(l *Lexer) stateFn {
	l.ignore()
	l.emit(ItemTypeLeftKeywordGroup)
	return lg.execute
}

// execute executes the main tasks for the lexer.
func (lg LexKeywordGroup) execute(l *Lexer) stateFn {
	LexKeyword{endKeyword: rightKeywordGroup}.execute(l)
	l.next()
	return lg.downgrade
}

// downgrade executes the ending tasks for the lexer.
func (lg LexKeywordGroup) downgrade(l *Lexer) stateFn {
	l.ignore()
	l.emit(ItemTypeRightKeywordGroup)
	return nil
}

// LexMethodArgGroup is a lexer for the method argument content.
type LexMethodArgGroup struct {
}

// newLexMethodArgGroup is a constructor for LexMethodArgGroup.
func newLexMethodArgGroup() stateFn {
	return LexMethodArgGroup{}.upgrade
}

// upgrade executes the initial tasks for the lexer.
func (lg LexMethodArgGroup) upgrade(l *Lexer) stateFn {
	l.emit(ItemTypeLeftKeywordGroup)
	return lg.execute
}

// execute executes the main tasks for the lexer.
func (lg LexMethodArgGroup) execute(l *Lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], rightKeywordGroup) {
			return lg.downgrade
		}
		r := l.next()
		if r == eof {
			return l.errorf("unclosed method arg")
		}
		fn, ok := executeMethodArg(l, r, rightKeywordGroup)
		if ok {
			runLexers(fn, l)
			continue
		}
		l.backup()
		return l.errorf("unknown character inside method argument at index (%d) %q", l.pos, r)
	}
}

// downgrade executes the ending tasks for the lexer.
func (lg LexMethodArgGroup) downgrade(l *Lexer) stateFn {
	l.pos += len(rightKeywordGroup)
	l.emit(ItemTypeRightKeywordGroup)
	return nil
}

// LexIdentifier is a lexer for the identifier content.
type LexIdentifier struct {
}

// newLexIdentifier is a constructor for LexIdentifier.
func newLexIdentifier() stateFn {
	return LexIdentifier{}.upgrade
}

// upgrade executes the initial tasks for the lexer.
func (li LexIdentifier) upgrade(*Lexer) stateFn {
	return li.execute
}

// execute executes the main tasks for the lexer.
func (li LexIdentifier) execute(l *Lexer) stateFn {
	if l.peek() == notKey {
		runLexers(newLexNot(), l)
	}
	emitIdentifier := func(l *Lexer) {
		if l.pos > l.start {
			l.emit(ItemTypeIdentifier)
		}
	}
	for {
		switch peek := l.peek(); {
		case isSpace(peek), string(peek) == RightKeyword:
			emitIdentifier(l)
			return li.downgrade
		case string(peek) == leftIndex:
			emitIdentifier(l)
			runLexers(newLexIndex(), l)
		case string(peek) == objectPropSplit:
			emitIdentifier(l)
			runLexers(newLexObjectProperty(), l)
		case peek == ':':
			emitIdentifier(l)
			runLexers(newLexMethod(), l)
		default:
			if r := l.next(); !wutils.IsIdentifier(r) {
				l.backup()
				emitIdentifier(l)
				return li.downgrade
			}
		}
	}
}

// downgrade executes the ending tasks for the lexer.
func (li LexIdentifier) downgrade(*Lexer) stateFn {
	return nil
}

// LexQuote is a lexer for the quoted content.
type LexQuote struct {
	character rune
}

// newLexQuote is a constructor for LexQuote.
func newLexQuote(character rune) stateFn {
	return LexQuote{character: character}.upgrade
}

// upgrade executes the initial tasks for the lexer.
func (lq LexQuote) upgrade(l *Lexer) stateFn {
	l.ignore()
	l.emit(ItemTypeStringStart)
	return lq.execute
}

// execute executes the main tasks for the lexer.
func (lq LexQuote) execute(l *Lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], "\\") {
			if l.pos > l.start {
				l.emit(ItemTypeString)
			}
			l.pos += 1
			l.ignore()
			l.pos += 1
			continue
		}
		if strings.HasPrefix(l.input[l.pos:], LeftKeyword) {
			if l.pos > l.start {
				l.emit(ItemTypeString)
			}
			runLexers(newLexKeyword(), l)
			continue
		}
		switch r := l.next(); {
		// Raw character can have enters.
		case r == eof || lq.character != '`' && r == '\n':
			return l.errorf("unclosed quoted string: %q", l.input[l.start:l.pos])
		case lq.character == r:
			l.backup()
			l.emit(ItemTypeString)
			return lq.downgrade
		}
	}
}

// downgrade executes the ending tasks for the lexer.
func (lq LexQuote) downgrade(l *Lexer) stateFn {
	l.pos++
	l.ignore()
	l.emit(ItemTypeStringEnd)
	return nil
}

// LexNumber is a lexer for the number content.
type LexNumber struct {
}

// newLexNumber is a constructor for LexNumber.
func newLexNumber() stateFn {
	return LexNumber{}.upgrade
}

// upgrade executes the initial tasks for the lexer.
func (ln LexNumber) upgrade(*Lexer) stateFn {
	return ln.execute
}

// execute executes the main tasks for the lexer.
func (ln LexNumber) execute(l *Lexer) stateFn {
	// Optional leading sign.
	l.accept("+-")
	// Is it hex?
	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}
	// Is it imaginary?
	l.accept("i")
	// Next thing mustn't be alphanumeric.
	if wutils.IsAlphaNumeric(l.peek()) {
		l.next()
		return l.errorf("bad number syntax: %q",
			l.input[l.start:l.pos])
	}
	l.emit(ItemTypeNumber)
	return ln.downgrade
}

// downgrade executes the ending tasks for the lexer.
func (ln LexNumber) downgrade(*Lexer) stateFn {
	return nil
}

// LexUnion is a lexer for the union content.
type LexUnion struct {
}

// newLexUnion is a constructor for LexUnion.
func newLexUnion() stateFn {
	return LexUnion{}.upgrade
}

// upgrade executes the initial tasks for the lexer.
func (lu LexUnion) upgrade(l *Lexer) stateFn {
	l.pos += len(string(unionKey))
	l.emit(ItemTypeUnionKeyword)
	return lu.execute
}

// execute executes the main tasks for the lexer.
func (lu LexUnion) execute(*Lexer) stateFn {
	return lu.downgrade
}

// downgrade executes the ending tasks for the lexer.
func (lu LexUnion) downgrade(*Lexer) stateFn {
	return nil
}

// LexOperator is a lexer for the operator content.
type LexOperator struct {
	operator     rune
	nextOperator rune
}

// newLexOperator is a constructor for LexOperator.
func newLexOperator(op rune) stateFn {
	return LexOperator{op, '\r'}.upgrade
}

// newLexEqOperator is a constructor for LexOperator (equal operator next).
func newLexEqOperator(op rune) stateFn {
	return LexOperator{op, equalKey}.upgrade
}

// newLexDupOperator is a constructor for LexOperator (duplicated operator next).
func newLexDupOperator(op rune) stateFn {
	return LexOperator{op, op}.upgrade
}

// itemType returns the operator type.
func (lu LexOperator) itemType() ItemType {
	switch lu.operator {
	case andKey:
		return ItemTypeAndKeyword
	case orKey:
		return ItemTypeOrKeyword
	case equalKey:
		return ItemTypeEqualKeyword
	case notKey:
		return ItemTypeNotEqualKeyword
	case greaterKey:
		switch {
		case lu.operator == lu.nextOperator:
			return ItemTypeBitwiseRightKeyword
		case lu.nextOperator == equalKey:
			return ItemTypeGreaterEqualKeyword
		default:
			return ItemTypeGreaterKeyword
		}
	case lessKey:
		switch {
		case lu.operator == lu.nextOperator:
			return ItemTypeBitwiseLeftKeyword
		case lu.nextOperator == equalKey:
			return ItemTypeLessEqualKeyword
		default:
			return ItemTypeLessKeyword
		}
	case plusKey:
		return ItemTypePlusKeyword
	case minusKey:
		return ItemTypeMinusKeyword
	case multiplyKey:
		return ItemTypeMultiplyKeyword
	case divisorKey:
		return ItemTypeDivisorKeyword
	case remainderKey:
		return ItemTypeRemainderKeyword
	default:
		return ItemTypeUnknown
	}
}

// upgrade executes the initial tasks for the lexer.
func (lu LexOperator) upgrade(l *Lexer) stateFn {
	l.emit(lu.itemType())
	return lu.execute
}

// execute executes the main tasks for the lexer.
func (lu LexOperator) execute(*Lexer) stateFn {
	return lu.downgrade
}

// downgrade executes the ending tasks for the lexer.
func (lu LexOperator) downgrade(*Lexer) stateFn {
	return nil
}

// LexMethod is a lexer for the method content.
type LexMethod struct {
}

// newLexMethod is a constructor for LexMethod.
func newLexMethod() stateFn {
	return LexMethod{}.upgrade
}

// upgrade executes the initial tasks for the lexer.
func (lm LexMethod) upgrade(l *Lexer) stateFn {
	l.emit(ItemTypeMethodStart)
	l.pos += len(":")
	l.ignore()
	for {
		if strings.HasPrefix(l.input[l.pos:], leftMethodCall) {
			if l.pos <= l.start {
				return l.errorf("bad method invoked")
			}
			break
		}
		if r := l.next(); !wutils.IsAlphaNumeric(r) {
			return l.errorf("invalid method name")
		}
	}
	l.emit(ItemTypeMethodName)
	l.pos += len(leftMethodCall)
	l.ignore()
	return lm.execute
}

// execute executes the main tasks for the lexer.
func (lm LexMethod) execute(l *Lexer) stateFn {
	for {
		isEnd := strings.HasPrefix(l.input[l.pos:], rightMethodCall)
		if isEnd {
			return lm.downgrade
		}
		isComma := strings.HasPrefix(l.input[l.pos:], ",")
		if isComma {
			l.pos++
			l.ignore()
			continue
		}
		if strings.HasPrefix(l.input[l.pos:], " ") {
			l.pos++
			l.ignore()
			continue
		}
		switch r := l.next(); {
		case r == '(':
			l.backup()
		case r == '"', r == '`', r == '\'', r == '!':
			l.backup()
		case !wutils.IsIdentifier(r):
			l.backup()
			return l.errorf("unknown character inside method at index (%d) %q", l.pos, r)
		default:
			l.backup()
		}
		l.emit(ItemTypeMethodArgStart)
		runLexers(newLexMethodArg(), l)
		l.emit(ItemTypeMethodArgEnd)
	}
}

// downgrade executes the ending tasks for the lexer.
func (lm LexMethod) downgrade(l *Lexer) stateFn {
	l.emit(ItemTypeMethodEnd)
	l.pos += len(rightMethodCall)
	l.ignore()
	return nil
}

// LexIndex is a lexer for the index content.
type LexIndex struct {
}

// newLexIndex is a constructor for LexIndex.
func newLexIndex() stateFn {
	return LexIndex{}.upgrade
}

// upgrade executes the initial tasks for the lexer.
func (lm LexIndex) upgrade(l *Lexer) stateFn {
	l.pos += len(leftIndex)
	l.emit(ItemTypeIndexStart)
	return lm.execute
}

// execute executes the main tasks for the lexer.
func (lm LexIndex) execute(l *Lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], rightIndex) {
			return lm.downgrade
		}
		switch r := l.next(); {
		case r == '+' || r == '-' || '0' <= r && r <= '9':
			l.backup()
			runLexers(newLexNumber(), l)
		default:
			return l.errorf("unknown character inside index brackets at index (%d) %q", l.pos, r)
		}
	}
}

// downgrade executes the ending tasks for the lexer.
func (lm LexIndex) downgrade(l *Lexer) stateFn {
	l.pos += len(rightIndex)
	l.emit(ItemTypeIndexEnd)
	return nil
}

// LexObjectProperty is a lexer for the object property content.
type LexObjectProperty struct {
}

// newLexObjectProperty is a constructor for LexObjectProperty.
func newLexObjectProperty() stateFn {
	return LexObjectProperty{}.upgrade
}

// upgrade executes the initial tasks for the lexer.
func (lop LexObjectProperty) upgrade(l *Lexer) stateFn {
	l.pos += len(objectPropSplit)
	l.ignore()
	return lop.execute
}

// execute executes the main tasks for the lexer.
func (lop LexObjectProperty) execute(l *Lexer) stateFn {
	for {
		if !wutils.IsIdentifier(l.peek()) {
			l.emit(ItemTypeObjectProperty)
			return lop.downgrade
		}
		l.next()
	}
}

// downgrade executes the ending tasks for the lexer.
func (lop LexObjectProperty) downgrade(*Lexer) stateFn {
	return nil
}

// LexMethodArg is a lexer for the method argument content.
type LexMethodArg struct {
}

// newLexMethodArg is a constructor for LexMethodArg.
func newLexMethodArg() stateFn {
	return LexMethodArg{}.upgrade
}

// upgrade executes the initial tasks for the lexer.
func (lma LexMethodArg) upgrade(*Lexer) stateFn {
	return lma.execute
}

// execute executes the main tasks for the lexer.
func (lma LexMethodArg) execute(l *Lexer) stateFn {
	if r := l.next(); r == '(' {
		runLexers(newLexLambda(), l)
	} else {
		l.backup()
	}
	for {
		r := l.next()
		if r == ',' || r == ')' {
			l.backup()
			return lma.downgrade
		}
 		if r == eof {
			return l.errorf("unclosed method arg")
		}
		if r == '\n' {
			l.ignore()
			continue
		}
		fn, ok := executeMethodArg(l, r, ",", ")")
		if ok {
			runLexers(fn, l)
			continue
		}
		l.backup()
		return l.errorf("unknown character inside method argument at index (%d) %q", l.pos, r)
	}
}

// downgrade executes the ending tasks for the lexer.
func (lma LexMethodArg) downgrade(*Lexer) stateFn {
	return nil
}

// LexLambda is a lexer for the lambda content.
type LexLambda struct {
}

// newLexLambda is a constructor for LexLambda.
func newLexLambda() stateFn {
	return LexLambda{}.upgrade
}

// upgrade executes the initial tasks for the lexer.
func (lb LexLambda) upgrade(l *Lexer) stateFn {
	l.ignore()
	l.emit(ItemTypeLambdaStart)
	return lb.execute
}

// execute executes the main tasks for the lexer.
func (lb LexLambda) execute(l *Lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == ' ', r == '\n':
			l.ignore()
		case r == ',':
			l.backup()
			l.emit(ItemTypeLambdaArg)
			l.acceptRun(", ")
			l.ignore()
		case r == ')':
			l.backup()
			l.emit(ItemTypeLambdaArg)
			l.pos++
			return lb.downgrade
		case r == eof:
			l.backup()
			return l.errorf("unclosed lambda at position %d", l.pos)
		case !wutils.IsIdentifier(r):
			l.backup()
			return l.errorf("unknown character inside lambda at index (%d) %q", l.pos, r)
		}
	}
}

// downgrade executes the ending tasks for the lexer.
func (lb LexLambda) downgrade(l *Lexer) stateFn {
	l.next() // character ')'
	r := l.skipSpaces()
	if r != '=' {
		l.backup()
		return l.errorf("wrong lambda syntax at position %d (expected '=' but got '%c')", l.pos, r)
	}
	r = l.next()
	if r != '>' {
		l.backup()
		return l.errorf("wrong lambda syntax at position %d (expected '>' but got '%c')", l.pos, r)
	}
	l.skipSpaces()
	l.ignore()
	l.emit(ItemTypeMethodArgStart)
	runLexers(newLexMethodArg(), l)
	l.emit(ItemTypeMethodArgEnd)
	l.emit(ItemTypeLambdaEnd)
	return nil
}

// LexNot is a lexer for the negation content.
type LexNot struct {
}

// newLexNot is a constructor for LexNot.
func newLexNot() stateFn {
	return LexNot{}.upgrade
}

// upgrade executes the initial tasks for the lexer.
func (ln LexNot) upgrade(l *Lexer) stateFn {
	l.pos += len(string(notKey))
	l.emit(ItemTypeNegation)
	return ln.execute
}

// execute executes the main tasks for the lexer.
func (ln LexNot) execute(l *Lexer) stateFn {
	if next := l.peek(); next == eof || !wutils.IsAlphaNumeric(next) && next != '!' && string(next) != leftKeywordGroup {
		return l.errorf("bad negation syntax: %q",
			l.input[l.start:l.pos])
	}
	return ln.downgrade
}

// downgrade executes the ending tasks for the lexer.
func (ln LexNot) downgrade(*Lexer) stateFn {
	return nil
}

// runLexers reads and executes a sequence of lexers.
func runLexers(state stateFn, l *Lexer) {
	for s := state; s != nil; {
		s = s(l)
	}
}

// executeMethodArg reads the method arguments content.
func executeMethodArg(l *Lexer, r rune, end ...string) (stateFn, bool) {
	var res stateFn
	switch {
	case isSpace(r):
		l.ignore()
	case (r == notKey || r == equalKey || r == greaterKey || r == lessKey) && l.peek() == equalKey:
		l.next()
		res = newLexEqOperator(r)
	case (r == greaterKey || r == lessKey) && l.peek() == r:
		l.next()
		res = newLexDupOperator(r)
	case (r == greaterKey || r == lessKey) && l.peek() == equalKey:
		l.next()
		res = newLexOperator(r)
	case r == greaterKey, r == lessKey, r == minusKey, r == multiplyKey, r == divisorKey, r == remainderKey:
		res = newLexOperator(r)
	case r == plusKey:
		res = newLexOperator(r)
	case r == andKey, r == orKey:
		nextR := l.next()
		if nextR != r {
			res = l.errorf("invalid operator format (%c%c)", r, nextR)
			break
		}
		res = newLexOperator(r)
	case r == unionKey:
		l.backup()
		runLexers(newLexUnion(), l)
	case r == '+' || r == '-' || '0' <= r && r <= '9':
		l.backup()
		res = newLexNumber()
	case r == notKey:
		l.backup()
		res = newLexNot()
	case string(r) == leftIndex:
		l.backup()
		res = newLexIndex()
	case r == '"', r == '\'', r == '`':
		res = newLexQuote(r)
	case wutils.IsIdentifier(r) || r == notKey:
		l.backup()
		res = newLexIdentifier()
	case string(r) == leftKeywordGroup:
		res = newLexMethodArgGroup()
	default:
		l.backup()
		for _, keyword := range end {
			if keyword == string(r) {
				return nil, true
			}
		}
		return nil, false
	}
	return res, true
}
