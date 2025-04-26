package token

import "testing"

func TestLookupIdent(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{input: "fn", expected: FUNCTION},
		{input: "let", expected: LET},
		{input: "true", expected: TRUE},
		{input: "false", expected: FALSE},
		{input: "if", expected: IF},
		{input: "else", expected: ELSE},
		{input: "return", expected: RETURN},
		{input: "abc", expected: IDENT},
	}

	for _, tt := range tests {
		tok := LookupIdent(tt.input)
		if tok != tt.expected {
			t.Fatalf("LookupIdent(%s) returned %s, expected %s", tt.input, tok, tt.expected)
		}
	}
}
