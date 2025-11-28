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
	type fields struct {
		input       []rune
		start       int
		pos         int
		items       chan item
		currentItem item
		consumed    bool
	}
	tests := []struct {
		name   string
		fields fields
		want   TokenType
		want1  string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lexer{
				input:       tt.fields.input,
				start:       tt.fields.start,
				pos:         tt.fields.pos,
				items:       tt.fields.items,
				currentItem: tt.fields.currentItem,
				consumed:    tt.fields.consumed,
			}
			got, got1 := l.NextToken()
			if got != tt.want {
				t.Errorf("Lexer.NextToken() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Lexer.NextToken() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
