package main

import (
	"combined/lexer"
	"combined/parser"
)

func createMatchProgram(str string) (*matchSentence, error) {
	lxr := lexer.Lex(str)
	psr := parser.NewParser(lxr)

	root := psr.Parse()

	return &matchSentence{something: root}, nil
}

func (ms *matchSentence) Match(pe *parsedEntry) bool {
	return eval(ms.something, pe)
}
