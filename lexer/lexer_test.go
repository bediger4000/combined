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
		wantType    TokenType
		wantLexeme  string
	}{
		{
			name:        "single left paren",
			singleToken: "(",
			wantType:    LPAREN,
			wantLexeme:  "(",
		},
		{
			name:        "single right paren",
			singleToken: ")",
			wantType:    RPAREN,
			wantLexeme:  ")",
		},
		{
			name:        "and token",
			singleToken: "&&",
			wantType:    AND,
			wantLexeme:  "&&",
		},
		{
			name:        "or token",
			singleToken: "||",
			wantType:    OR,
			wantLexeme:  "||",
		},
		{
			name:        "not token",
			singleToken: "-",
			wantType:    NOT,
			wantLexeme:  "-",
		},
		{
			name:        "exact match token",
			singleToken: "=",
			wantType:    MATCH_OP,
			wantLexeme:  "=",
		},
		{
			name:        "regexp match token",
			singleToken: "~",
			wantType:    MATCH_OP,
			wantLexeme:  "~",
		},
		{
			name:        "field token",
			singleToken: "timestamp",
			wantType:    FIELD,
			wantLexeme:  "timestamp",
		},
		{
			name:        "regexp pattern token",
			singleToken: "/abcdefg/",
			wantType:    PATTERN,
			wantLexeme:  "/abcdefg/",
		},
		{
			name:        "regexp pattern token, metachars",
			singleToken: "/a.b[cde]f\\//",
			wantType:    PATTERN,
			wantLexeme:  "/a.b[cde]f\\//",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Lex(tt.singleToken)
			gotToken, gotLexeme := l.NextToken()
			if gotToken != tt.wantType {
				t.Errorf("Lexer.NextToken() got = %v, want %v", gotToken, tt.wantType)
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

func TestLexer_NextTokenStream(t *testing.T) {
	type testItem struct {
		wantType   TokenType
		wantLexeme string
	}

	tests := []struct {
		name        string
		tokenString string
		wantItems   []testItem
	}{
		{
			name:        "parens, no whitespace",
			tokenString: "()()()",
			wantItems: []testItem{
				testItem{LPAREN, "("},
				testItem{RPAREN, ")"},
				testItem{LPAREN, "("},
				testItem{RPAREN, ")"},
				testItem{LPAREN, "("},
				testItem{RPAREN, ")"},
			},
		},
		{
			name:        "parens, whitespace",
			tokenString: "( ) (\t)\"(')",
			wantItems: []testItem{
				testItem{LPAREN, "("},
				testItem{RPAREN, ")"},
				testItem{LPAREN, "("},
				testItem{RPAREN, ")"},
				testItem{LPAREN, "("},
				testItem{RPAREN, ")"},
			},
		},
		{
			name:        "exact match",
			tokenString: "ipaddr = /10.0.40.70/",
			wantItems: []testItem{
				testItem{FIELD, "ipaddr"},
				testItem{MATCH_OP, "="},
				testItem{PATTERN, "/10.0.40.70/"},
			},
		},
		{
			name:        "logical expression",
			tokenString: "ipaddr = /10.0.40.70/ && -(method~/GET|HEAD/)",
			wantItems: []testItem{
				testItem{FIELD, "ipaddr"},
				testItem{MATCH_OP, "="},
				testItem{PATTERN, "/10.0.40.70/"},
				testItem{AND, "&&"},
				testItem{NOT, "-"},
				testItem{LPAREN, "("},
				testItem{FIELD, "method"},
				testItem{MATCH_OP, "~"},
				testItem{PATTERN, "/GET|HEAD/"},
				testItem{RPAREN, ")"},
			},
		},
		{
			name:        "difficult metacharacters",
			tokenString: `url=/http:\/\/bruceediger\.com\//`,
			wantItems: []testItem{
				testItem{FIELD, "url"},
				testItem{MATCH_OP, "="},
				testItem{PATTERN, `/http:\/\/bruceediger\.com\//`},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lxr := Lex(tt.tokenString)
			i := 0
			for gotType, gotToken := lxr.NextToken(); gotType != EOF; gotType, gotToken = lxr.NextToken() {
				if gotType != tt.wantItems[i].wantType {
					t.Errorf("Lexer.NextToken() got = %v, want %v", gotType, tt.wantItems[i].wantType)
				}
				if gotToken != tt.wantItems[i].wantLexeme {
					t.Errorf("Lexer.NextToken() got = %v, want %v", gotToken, tt.wantItems[i].wantLexeme)
				}
				lxr.Consume()
				i++
			}

		})
	}
}
