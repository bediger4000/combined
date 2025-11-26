package main

import (
	"combined/lexer"
	"combined/tree"
	"fmt"
	"os"
)

func eval(node *tree.Node, pe *parsedEntry) bool {
	if node == nil {
		fmt.Fprintf(os.Stderr, "reached nil node in error\n")
	}
	switch node.Op {
	case lexer.OR:
		if leftVal := eval(node.Left, pe); leftVal {
			// short circuit evaluation
			return true
		}
		return eval(node.Right, pe)
	case lexer.AND:
		if leftVal := eval(node.Left, pe); !leftVal {
			// short circuit evaluation
			return false
		}
		return eval(node.Right, pe)
	case lexer.NOT:
		return !eval(node.Left, pe)
	case lexer.EXACT_MATCH:
		return node.ExactValue == pe.fields[node.FieldIndex]
	case lexer.REGEX_MATCH:
		matched := node.Pattern.MatchString(pe.fields[node.FieldIndex])
		return matched
	default:
		fmt.Fprintf(os.Stderr, "reached node with Type %s in error\n", node.Op)
		return false
	}
}
