package tree

// Parse tree - a binary tree of objects of type Node,
// and associated functions and methods.

import (
	"fmt"
	"io"
	"regexp"

	"combined/lexer"
)

// Node has all elements exported, everything reaches inside instances
// of Node to find things out, or to change Left and Right. Private
// elements would cost me gross ol' getter and setter boilerplate.
type Node struct {
	Op         lexer.TokenType
	Lexeme     string
	FieldIndex int
	Pattern    *regexp.Regexp
	ExactValue string
	Left       *Node
	Right      *Node
}

// NewNode creates interior nodes of a parse tree
func NewNode(op lexer.TokenType, lexeme string) *Node {
	if op == lexer.MATCH_OP {
		switch lexeme {
		case "=":
			op = lexer.EXACT_MATCH
		case "~":
			op = lexer.REGEX_MATCH
		}
	}
	return &Node{
		Op:     op,
		Lexeme: lexeme,
	}
}

// NotNode handles "-something" situtations.
func NotNode(_ string, factor *Node) *Node {
	return &Node{
		Op:   lexer.NOT,
		Left: factor,
	}
}

// Print puts a human-readable, nicely formatted string representation
// of a parse tree onto the io.Writer, w.  Essentially just an in-order
// traversal of a binary tree, with accommodating a few oddities, like
// parenthesization, and the "-" (not) operator being a prefix.
func (p *Node) Print(w io.Writer) {
	if p == nil {
		return
	}

	if p.Op == lexer.NOT {
		w.Write([]byte{'-'})
		p.Left.Print(w)
		return
	}

	if p.Op == lexer.FIELD || p.Op == lexer.PATTERN {
		fmt.Fprintf(w, "%s", p.Lexeme)
		return
	}

	w.Write([]byte{'('})
	p.Left.Print(w)
	fmt.Fprintf(w, " %s ", p.Lexeme)
	p.Right.Print(w)
	w.Write([]byte{')'})
}

func (p *Node) String() string {
	return p.Lexeme
}

func (p *Node) graphNode(w io.Writer) {
	if p == nil {
		return
	}

	fmt.Fprintf(w, "N%p [label=%q];\n", p, p.Lexeme)

	p.Left.graphNode(w)
	p.Right.graphNode(w)

	if p.Left != nil {
		fmt.Fprintf(w, "N%p -> N%p;\n", p, p.Left)
	}
	if p.Right != nil {
		fmt.Fprintf(w, "N%p -> N%p;\n", p, p.Right)
	}
}

// GraphNode puts a dot-format text representation of
// a parse tree on w io.Writer.
func (p *Node) GraphNode(w io.Writer) {
	fmt.Fprintf(w, "digraph g {\n")
	p.graphNode(w)
	fmt.Fprintf(w, "}\n")
}
