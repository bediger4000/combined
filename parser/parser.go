package parser

import (
	"combined/lexer"
	"combined/tree"
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

// Parse starts building a parse tree, and returns it
func (p *Parser) Parse() *tree.Node {
	return p.expr()
}

func (p *Parser) expr() *tree.Node {
	node := p.term()
	for kind, lexeme := p.lexer.NextToken(); kind == lexer.OR; kind, lexeme = p.lexer.NextToken() {
		tmp := tree.NewNode(kind, lexeme)
		p.lexer.Consume()
		tmp.Left = node
		node = tmp
		node.Right = p.term()
	}
	return node
}

func (p *Parser) term() *tree.Node {
	node := p.factor()
	for kind, lexeme := p.lexer.NextToken(); kind == lexer.AND; kind, lexeme = p.lexer.NextToken() {
		tmp := tree.NewNode(kind, lexeme)
		p.lexer.Consume()
		tmp.Left = node
		node = tmp
		node.Right = p.factor()
	}
	return node
}

func (p *Parser) factor() *tree.Node {
	kind, lexeme := p.lexer.NextToken()
	switch kind {
	case lexer.NOT:
		unaryOp := lexeme
		p.lexer.Consume()
		factor := p.factor()
		return tree.NotNode(unaryOp, factor)
	case lexer.FIELD:
		return p.boolean()
	case lexer.LPAREN:
		p.lexer.Consume() // left paren
		expr := p.expr()
		kind, lexeme = p.lexer.NextToken()
		if kind != lexer.RPAREN {
			fmt.Printf("Wanted an RPAREN, got %v: %q\n", kind, lexeme)
		}
		p.lexer.Consume() // right paren
		return expr
	default:
		fmt.Printf("Wanted a CONSTANT or LPAREN, got %v: %q\n", kind, lexeme)
	}
	return nil
}

func (p *Parser) boolean() *tree.Node {
	kind, lexeme := p.lexer.NextToken()
	field := lexeme
	p.lexer.Consume()
	fmt.Printf("boolean, field %q\n", field)

	kind, lexeme = p.lexer.NextToken()
	if kind != lexer.MATCH_OP {
		// what to do
		fmt.Printf("Wanted a MATCH_OP, got %v: %q\n", kind, lexeme)
		return nil
	}
	p.lexer.Consume()
	booleanNode := tree.NewNode(kind, lexeme)
	fmt.Printf("boolean, MATCH_OP %q\n", lexeme)

	kind, lexeme = p.lexer.NextToken()
	if kind != lexer.PATTERN {
		// what to do
		fmt.Printf("Wanted a PATTERN, got %v: %q\n", kind, lexeme)
		return nil
	}
	pattern := strings.TrimRight(strings.TrimLeft(lexeme, "/"), "/")
	p.lexer.Consume()
	fmt.Printf("boolean, PATTERN %q\n", pattern)

	var ok bool
	if booleanNode.FieldIndex, ok = FieldToIndex[field]; !ok {
		// what to do
		fmt.Printf("No field named %q available for matching\n", field)
		return nil
	}

	if booleanNode.Op == lexer.EXACT_MATCH {
		booleanNode.ExactValue = pattern
	} else if booleanNode.Op == lexer.REGEX_MATCH {
		var err error
		booleanNode.Pattern, err = regexp.Compile(pattern)
		if err != nil {
			// what to do
		}
	}

	return booleanNode
}

// NewParser creates a filled in Parser struct and returns it.
func NewParser(lxr *lexer.Lexer) *Parser {
	return &Parser{lexer: lxr}
}
