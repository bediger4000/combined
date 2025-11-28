package lexer

import (
	"testing"
)

func TestTokenType_String(t *testing.T) {
	tests := []struct {
		name string
		tr   TokenType
		want string
	}{
		{
			name: "EOF token type",
			tr:   EOF,
			want: "EOF",
		},
		{
			name: "PATTERN token type", tr: PATTERN, want: "PATTERN"},
		{
			name: "FIELD token type", tr: FIELD, want: "FIELD"},
		{
			name: "AND token type", tr: AND, want: "AND"},
		{
			name: "OR token type", tr: OR, want: "OR"},
		{
			name: "NOT token type", tr: NOT, want: "NOT"},
		{
			name: "LPAREN token type", tr: LPAREN, want: "LPAREN"},
		{
			name: "RPAREN token type", tr: RPAREN, want: "RPAREN"},
		{
			name: "MATCH_OP token type", tr: MATCH_OP, want: "MATCH_OP"},
		{
			name: "EXACT_MATCH token type", tr: EXACT_MATCH, want: "EXACT_MATCH"},
		{
			name: "REGEX_MATCH token type", tr: REGEX_MATCH, want: "REGEX_MATCH"},
		{
			name: "EOL token type", tr: EOL, want: "EOL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.String(); got != tt.want {
				t.Errorf("TokenType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLexer_NextToken(t *testing.T) {
	tests := []struct {
		name        string
		singleToken string
		wantToken   TokenType
		wantLexeme  string
	}{
		{
			name:        "single left paren",
			singleToken: "(",
			wantToken:   LPAREN,
			wantLexeme:  "(",
		},
		{
			name:        "single right paren",
			singleToken: ")",
			wantToken:   RPAREN,
			wantLexeme:  ")",
		},
		{
			name:        "and token",
			singleToken: "&&",
			wantToken:   AND,
			wantLexeme:  "&&",
		},
		{
			name:        "or token",
			singleToken: "||",
			wantToken:   OR,
			wantLexeme:  "||",
		},
		{
			name:        "not token",
			singleToken: "-",
			wantToken:   NOT,
			wantLexeme:  "-",
		},
		{
			name:        "exact match token",
			singleToken: "=",
			wantToken:   MATCH_OP,
			wantLexeme:  "=",
		},
		{
			name:        "regexp match token",
			singleToken: "~",
			wantToken:   MATCH_OP,
			wantLexeme:  "~",
		},
		{
			name:        "field token",
			singleToken: "timestamp",
			wantToken:   FIELD,
			wantLexeme:  "timestamp",
		},
		{
			name:        "regexp pattern token",
			singleToken: "/abcdefg/",
			wantToken:   PATTERN,
			wantLexeme:  "/abcdefg/",
		},
		{
			name:        "regexp pattern token, metachars",
			singleToken: "/a.b[cde]f\\//",
			wantToken:   PATTERN,
			wantLexeme:  "/a.b[cde]f\\//",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Lex(tt.singleToken)
			gotToken, gotLexeme := l.NextToken()
			if gotToken != tt.wantToken {
				t.Errorf("Lexer.NextToken() got = %v, want %v", gotToken, tt.wantToken)
			}
			if gotLexeme != tt.wantLexeme {
				t.Errorf("Lexer.NextToken() got lexeme = %v, want %v", gotLexeme, tt.wantLexeme)
			}
			if l.consumed {
				t.Errorf("Lexer consumed an item  without being asked")
			}
			l.Consume()
			if !l.consumed {
				t.Errorf("Lexer did not consume an item when asked")
			}
		})
	}
}
