package token

import (
	"strings"
)

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT  = "IDENT" // add, foobar, x, y, ...
	INT    = "INT"   // 1343456
	STRING = "STRING"

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
	COMMA     = ","
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
	LOAD     = "LOAD"
	READ     = "READ"
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

	ROW   = "ROW"
	COL   = "COL"
	WHERE = "WHERE"
)

type Token struct {
	Type    TokenType
	Literal string
}

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

func LookupIdent(ident string) TokenType {
	// fmt.Println("[LookupIdent] ident: ", ident)
	// make keyword matching case-insensitive
	// i.e., load and LOAD will mean the same thing
	lowercaseIdent := strings.ToLower(ident)
	if tok, ok := keywords[lowercaseIdent]; ok {
		return tok
	}
	return IDENT
}
