// lexer package is responsible for tokenizing the input string.
//
// The lexer reads the input string character by character and returns tokens based on the characters it reads.
// csvlang supports one and two character (eg. !=) tokens.
package lexer

import (
	"strings"

	"github.com/Rishabh570/csvlang/token"
)

// Lexer is the struct that represents the lexer.
// It contains the input string, the current position in the input string, the current character being read, and the next character to be read.
// It also contains the current line and column numbers for error reporting.
type Lexer struct {
	input        string // input string
	position     int    // current position in input (points to current char)
	readPosition int    // next position in input (points to next char)
	ch           byte   // current char under examination
	Line         int    // current line number
	Column       int    // current column number
}

// New creates a new Lexer with the given input string.
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		Line:   1,
		Column: 1,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1

	// Track line and column numbers
	if l.ch == '\n' {
		l.Line++
		l.Column = 1
	} else {
		l.Column++
	}
}

// readComment reads the comment until the end of the line
// not using l.readString() as we want to ignore double-quote when reading comments
func (l *Lexer) readComment() token.Token {
	position := l.position + 1
	for {
		l.readChar()

		// we don't want a single line comment to bleed into the next line
		// so we stop reading when we reach a newline or the end of the input
		if l.ch == 0 || l.ch == '\n' {
			break
		}
	}

	commentedText := l.input[position:l.position]
	// trim to remove leading and/or trailing spaces
	commentedText = strings.Trim(commentedText, " ")

	return token.Token{Type: token.SINGLE_LINE_COMMENT, Literal: commentedText}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()

		if l.ch == '"' || l.ch == 0 || l.ch == '\n' {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

// NextToken reads the next token from the input string.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	// fmt.Println("[l.NextToken] tok: ", tok.Literal, tok.Type)

	l.skipWhitespace()

	switch l.ch {
	case '#':
		// skip to next line
		tok = l.readComment()
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '"':
		tok = token.Token{Type: token.STRING, Literal: l.readString()}
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case '\n':
		tok = newToken(token.NEWLINE, l.ch)
		l.Line++     // Increment line number
		l.Column = 0 // Reset column counter
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		}
		if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		}

		tok = newToken(token.ILLEGAL, l.ch)
	}
	l.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '.' || ch == '/'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
