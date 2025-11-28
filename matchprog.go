package main

import (
	"combined/lexer"
	"combined/parser"
	"combined/tree"
)

func createMatchProgram(str string) (*tree.Node, error) {
	lxr := lexer.Lex(str)
	psr := parser.NewParser(lxr)

	return psr.Parse()
}

// Match exists to cover up the use of a recursive function
func Match(root *tree.Node, pe *parsedEntry) bool {
	return eval(root, pe)
}
