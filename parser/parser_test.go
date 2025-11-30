package parser

import (
	"combined/lexer"
	"combined/tree"
	"reflect"
	"regexp"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name      string
		stringrep string
		want      *tree.Node
		wantErr   bool
	}{
		{
			name:      "lexical match",
			stringrep: "ipaddr = /abc/",
			want: &tree.Node{
				Op:         lexer.EXACT_MATCH,
				Lexeme:     "=",
				FieldIndex: 0,
				ExactValue: "abc",
			},
			wantErr: false,
		},
		{
			name:      "regular expression match",
			stringrep: "timestamp~/a.b.c/",
			want: &tree.Node{
				Op:         lexer.REGEX_MATCH,
				Lexeme:     "~",
				FieldIndex: 2,
				Pattern:    regexp.MustCompile(`a.b.c`),
			},
			wantErr: false,
		},
		{
			name:      "regular expression match",
			stringrep: "timestamp~/a.b.c/",
			want: &tree.Node{
				Op:         lexer.REGEX_MATCH,
				Lexeme:     "~",
				FieldIndex: 2,
				Pattern:    regexp.MustCompile(`a.b.c`),
			},
			wantErr: false,
		},
		{
			name:      "unknown match operator",
			stringrep: "ipaddr X /a.b.c/",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "unknown field, regex match",
			stringrep: "ipcraddr ~ /a.b.c/",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "bad regular expression",
			stringrep: "ipaddr ~ /[/",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "missing right paren",
			stringrep: "-(ipaddr ~ /abcdefg/",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "missing field",
			stringrep: "~ /abcdefg/",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "regular expression match, no pattern",
			stringrep: "url~a.b.c",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "NOT lexical match",
			stringrep: "-(url=/abc/)",
			want: &tree.Node{
				Op: lexer.NOT,
				Left: &tree.Node{
					Op:         lexer.EXACT_MATCH,
					Lexeme:     "=",
					FieldIndex: 4,
					ExactValue: "abc",
				},
			},
			wantErr: false,
		},
		{
			name:      "lexical match AND lexical match",
			stringrep: "url=/abc/ && method=/GET/",
			want: &tree.Node{
				Op:     lexer.AND,
				Lexeme: "&&",
				Left: &tree.Node{
					Op:         lexer.EXACT_MATCH,
					Lexeme:     "=",
					FieldIndex: 4,
					ExactValue: "abc",
				},
				Right: &tree.Node{
					Op:         lexer.EXACT_MATCH,
					Lexeme:     "=",
					FieldIndex: 3,
					ExactValue: "GET",
				},
			},
			wantErr: false,
		},
		{
			name:      "regex match OR regex match",
			stringrep: "url~/abc|def/ || method~/[Gg][Ee][Tt]/",
			want: &tree.Node{
				Op:     lexer.OR,
				Lexeme: "||",
				Left: &tree.Node{
					Op:         lexer.REGEX_MATCH,
					Lexeme:     "~",
					FieldIndex: 4,
					Pattern:    regexp.MustCompile(`abc|def`),
				},
				Right: &tree.Node{
					Op:         lexer.REGEX_MATCH,
					Lexeme:     "~",
					FieldIndex: 3,
					Pattern:    regexp.MustCompile("[Gg][Ee][Tt]"),
				},
			},
			wantErr: false,
		},
		{
			name:      "exact match AND regex match OR regex match",
			stringrep: "ipaddr=/zork/ && url~/abc|def/ || method~/[Gg][Ee][Tt]/",
			want: &tree.Node{
				Op:     lexer.OR,
				Lexeme: "||",
				Left: &tree.Node{
					Op:     lexer.AND,
					Lexeme: "&&",
					Left: &tree.Node{
						Op:         lexer.EXACT_MATCH,
						Lexeme:     "=",
						FieldIndex: 0,
						ExactValue: "zork",
					},
					Right: &tree.Node{
						Op:         lexer.REGEX_MATCH,
						Lexeme:     "~",
						FieldIndex: 4,
						Pattern:    regexp.MustCompile(`abc|def`),
					},
				},
				Right: &tree.Node{
					Op:         lexer.REGEX_MATCH,
					Lexeme:     "~",
					FieldIndex: 3,
					Pattern:    regexp.MustCompile("[Gg][Ee][Tt]"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(lexer.Lex(tt.stringrep))
			got, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parser.Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
