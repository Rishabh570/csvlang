// token package is used to define the token types and the Token struct.
//
// The TokenType type is a string that represents the type of a token.
// The Token struct has two fields: Type and Literal.
//
// 1. Type is the type of the token, and
//
// 2. Literal is the actual value of the token.
//
// csvlang has some reserved keywords and the keywords map is used to store them.
package token

import (
	"strings"
)

type TokenType string

const (
	ILLEGAL = "ILLEGAL" // unknown token
	EOF     = "EOF"     // end of file

	// Identifiers + literals
	IDENT  = "IDENT"  // add, foobar, x, y, ...
	INT    = "INT"    // 1343456
	STRING = "STRING" // "foobar"

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	LT       = "<"
	GT       = ">"
	EQ       = "=="
	NOT_EQ   = "!="

	// Delimiters
	COMMA     = "," // acts as a delimiter in arrays
	SEMICOLON = ";"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Comment
	SINGLE_LINE_COMMENT = "#"

	// Keywords
	LOAD     = "LOAD" // load csv file
	READ     = "READ" // read data from the loaded csv file
	UPDATE   = "UPDATE"
	DELETE   = "DELETE"
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	FOR      = "FOR"
	IN       = "IN"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	SAVE     = "SAVE"
	AS       = "AS" // used in "save rows as filtered.csv" statements

	ROW   = "ROW"   // read particular rows from the loaded csv file
	COL   = "COL"   // read particular columns from the loaded csv rows
	WHERE = "WHERE" // filter rows based on a condition
)

type Token struct {
	Type    TokenType
	Literal string
}

// keywords is a map of reserved keywords in csvlang
var keywords = map[string]TokenType{
	"load":   LOAD,
	"read":   READ,
	"update": UPDATE,
	"delete": DELETE,
	"row":    ROW,
	"col":    COL,
	"where":  "WHERE",
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"save":   SAVE,
	"as":     AS,
	"for":    FOR,
	"in":     IN,
}

// LookupIdent checks if the given identifier is a keyword
// defaults to IDENT if not a keyword
//
// Example:
//
//	LookupIdent("fn") // returns FUNCTION
//	LookupIdent("abc") // returns IDENT
func LookupIdent(ident string) TokenType {
	// make keyword matching case-insensitive
	// i.e., load and LOAD will mean the same thing
	lowercaseIdent := strings.ToLower(ident)
	if tok, ok := keywords[lowercaseIdent]; ok {
		return tok
	}
	return IDENT
}
