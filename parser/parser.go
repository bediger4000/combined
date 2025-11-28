package parser

import (
	"combined/lexer"
	"combined/tree"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

/*
expr     -> term { OR term }
term     -> factor { AND factor }
factor   -> '(' expr ')' | NOT factor | boolean
boolean  -> FIELD match-op PATTERN
match-op -> '='|'~'
*/

/*
* One parse method per non-terminal symbol
* A non-terminal symbol on the RHS of a rewrite rule
  leads to a call to the parse method for that non-terminal
* Terminal symbol on the RHS of a rewrite rule leads to "consuming"
  that token from the input token string
* | in the grammar leads to "if-else" in the parser
* {...} in the grammar leads to "while" in the parser
*/

// Parser carries around what it needs to create a parse tree from a stream of
// lexer tokens. Type Parser exists independently of type lexer.Lexer to allow
// separation of parsing functions, the methods of *Parser, from lexing
// functions.
type Parser struct {
	lexer *lexer.Lexer
}

// Parse starts building a parse tree. Covers up the
// use of an un-exported non-terminal function.
func (p *Parser) Parse() (*tree.Node, error) {
	return p.expr()
}

func (p *Parser) expr() (*tree.Node, error) {
	node, err := p.term()
	for kind, lexeme := p.lexer.NextToken(); kind == lexer.OR; kind, lexeme = p.lexer.NextToken() {
		tmp := tree.NewNode(kind, lexeme)
		p.lexer.Consume()
		tmp.Left = node
		node = tmp
		var e error
		node.Right, e = p.term()
		err = errors.Join(err, e)
	}
	return node, err
}

func (p *Parser) term() (*tree.Node, error) {
	node, err := p.factor()
	for kind, lexeme := p.lexer.NextToken(); kind == lexer.AND; kind, lexeme = p.lexer.NextToken() {
		tmp := tree.NewNode(kind, lexeme)
		p.lexer.Consume()
		tmp.Left = node
		node = tmp
		var e error
		node.Right, e = p.factor()
		err = errors.Join(err, e)
	}
	return node, err
}

func (p *Parser) factor() (*tree.Node, error) {
	kind, lexeme := p.lexer.NextToken()
	switch kind {
	case lexer.NOT:
		unaryOp := lexeme
		p.lexer.Consume()
		factor, err := p.factor()
		return tree.NotNode(unaryOp, factor), err
	case lexer.FIELD:
		return p.boolean()
	case lexer.LPAREN:
		p.lexer.Consume() // left paren
		expr, err := p.expr()
		kind, lexeme = p.lexer.NextToken()
		if kind != lexer.RPAREN {
			err = errors.Join(err, fmt.Errorf("wanted an RPAREN, got %v: %q\n", kind, lexeme))
		}
		p.lexer.Consume() // right paren
		return expr, err
	default:
		return nil, fmt.Errorf("wanted NOT, FIELD or LPAREN, got %v: %q\n", kind, lexeme)
	}
	return nil, nil
}

func (p *Parser) boolean() (*tree.Node, error) {
	kind, lexeme := p.lexer.NextToken()
	field := lexeme
	p.lexer.Consume()

	kind, lexeme = p.lexer.NextToken()
	if kind != lexer.MATCH_OP {
		return nil, fmt.Errorf("wanted a MATCH_OP, got %v: %q\n", kind, lexeme)
	}
	p.lexer.Consume()
	booleanNode := tree.NewNode(kind, lexeme)

	kind, lexeme = p.lexer.NextToken()
	if kind != lexer.PATTERN {
		return nil, fmt.Errorf("wanted a PATTERN, got %v: %q\n", kind, lexeme)
	}
	pattern := strings.TrimSuffix(strings.TrimPrefix(lexeme, "/"), "/")
	p.lexer.Consume()

	var ok bool
	if booleanNode.FieldIndex, ok = FieldToIndex[field]; !ok {
		return nil, fmt.Errorf("no field named %q available for matching\n", field)
	}

	if booleanNode.Op == lexer.EXACT_MATCH {
		booleanNode.ExactValue = pattern
	} else if booleanNode.Op == lexer.REGEX_MATCH {
		var err error
		booleanNode.Pattern, err = regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
	}

	return booleanNode, nil
}

// NewParser creates a filled in Parser struct and returns it.
func NewParser(lxr *lexer.Lexer) *Parser {
	return &Parser{lexer: lxr}
}
