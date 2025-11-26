package lexer

import (
	"unicode"
)

// TokenType tells parser what the lexer thinks
// the category for this token is.
type TokenType int

// EOF and others: all the types of tokens
const (
	EOF         TokenType = iota
	PATTERN     TokenType = iota
	FIELD       TokenType = iota
	AND         TokenType = iota
	OR          TokenType = iota
	NOT         TokenType = iota
	LPAREN      TokenType = iota
	RPAREN      TokenType = iota
	MATCH_OP    TokenType = iota
	EXACT_MATCH TokenType = iota
	REGEX_MATCH TokenType = iota
	EOL         TokenType = iota
)

func (t TokenType) String() string {
	switch t {
	case PATTERN:
		return "PATTERN"
	case FIELD:
		return "FIELD"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	case MATCH_OP:
		return "MATCH_OP"
	case EXACT_MATCH:
		return "EXACT_MATCH"
	case REGEX_MATCH:
		return "REGEX_MATCH"
	case AND:
		return "AND"
	case OR:
		return "OR"
	case NOT:
		return "NOT"
	case EOL:
		return "EOL"
	case EOF:
		return "EOF"
	}
	return "unknown"
}

type item struct {
	kind   TokenType
	lexeme string
}

// Lexer instances hold information needed to break
// a string into arithmetic expression tokens.
type Lexer struct {
	input       []rune
	start       int
	pos         int
	items       chan item
	currentItem item
	consumed    bool
}

type stateFn func(*Lexer) stateFn

// Lex creates a new ready-to-go Lexer instance.
// Runs a goroutine in the background.
func Lex(input string) *Lexer {
	l := &Lexer{
		input:    []rune(input),
		items:    make(chan item),
		consumed: true,
	}
	go l.run()
	return l
}

// NextToken called by parser to retrieve whatever
// the lexer thinks is the next token.
func (l *Lexer) NextToken() (TokenType, string) {
	if l.consumed {
		l.currentItem = <-l.items
		l.consumed = false
	}
	return l.currentItem.kind, l.currentItem.lexeme
}

// Consume called by parser when it has found a place in the parse tree for the
// token. Parse can and does call NextToken() repeatedly to find out the
// token's type.
func (l *Lexer) Consume() {
	l.consumed = true
}

func (l *Lexer) run() {
	for state := lexWhiteSpace; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func lexWhiteSpace(l *Lexer) stateFn {
	for _, r := range l.input[l.start:] {
		switch r {
		case ' ', '"', '\'', '\t':
			l.pos++
			l.start++
		default:
			return l.nextStateFn()
		}
	}
	return nil
}

func (l *Lexer) nextStateFn() stateFn {
	if l.pos >= len(l.input) {
		return lexEOF
	}
	switch l.input[l.pos] {
	case '(':
		return lexLeftParen
	case ')':
		return lexRightParen
	case '/':
		return lexSlash
	case '|':
		return lexPipe
	case '&':
		return lexAmpersand
	case '-':
		return lexMinus
	case '=', '~':
		return lexMatchOp
	case '\n':
		return lexEOL
	default:
		if unicode.IsLetter(l.input[l.pos]) {
			return lexField
		}
		return lexWhiteSpace
	}
}

func (l *Lexer) emit(t TokenType) {
	l.items <- item{t, string(l.input[l.start:l.pos])}
	l.start = l.pos
}

func lexEOF(l *Lexer) stateFn {
	return nil
}

func identifierChar(r rune) bool {
	return unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_'
}

func lexField(l *Lexer) stateFn {
	for l.pos < len(l.input) && identifierChar(rune(l.input[l.pos])) {
		l.pos++
	}
	l.emit(FIELD)
	return l.nextStateFn()
}

func lexLeftParen(l *Lexer) stateFn {
	l.pos++
	l.emit(LPAREN)
	return l.nextStateFn()
}

func lexRightParen(l *Lexer) stateFn {
	l.pos++
	l.emit(RPAREN)
	return l.nextStateFn()
}

func lexMatchOp(l *Lexer) stateFn {
	l.pos++
	l.emit(MATCH_OP)
	return l.nextStateFn()
}

func lexMinus(l *Lexer) stateFn {
	l.pos++
	l.emit(NOT)
	return l.nextStateFn()
}

func lexAmpersand(l *Lexer) stateFn {
	l.pos += 2
	l.emit(AND)
	return l.nextStateFn()
}

func lexPipe(l *Lexer) stateFn {
	l.pos += 2
	l.emit(OR)
	return l.nextStateFn()
}

func lexSlash(l *Lexer) stateFn {
	var escaping bool
	l.pos++ // we know l.input[l.pos] == '/'
	for l.pos < len(l.input) {
		if !escaping && rune(l.input[l.pos]) == '/' {
			l.pos++
			break
		}

		if rune(l.input[l.pos]) == '\\' {
			escaping = true
			l.pos++
			continue
		}
		escaping = false
		l.pos++
	}
	l.emit(PATTERN)
	return l.nextStateFn()
}

func lexEOL(l *Lexer) stateFn {
	l.pos++
	l.emit(EOL)
	return l.nextStateFn()
}
