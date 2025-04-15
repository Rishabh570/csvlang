package lexer

import (
	"github.com/Rishabh570/csvlang/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	Line         int // current line number
	Column       int // current column number
}

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

func (l *Lexer) readComment() token.Token {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}

	return token.Token{Type: token.SINGLE_LINE_COMMENT, Literal: token.SINGLE_LINE_COMMENT}
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

		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

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
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		// TODO: do we need to keep track of the previous token to achieve load functionality?
		// eg. Load input.csv => when we're on "input.csv" token, if prev token was "load", "input.csv" will have tok.Type = FILEPATH (might help later to identify path when reading file)
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			// fmt.Println("[l.NextToken] isLetter: ", tok.Literal, tok.Type)
			tok.Type = token.LookupIdent(tok.Literal)
			// fmt.Println("[NextToken] type", tok.Type)
			// fmt.Println("[NextToken] literal", tok.Literal)
			return tok
		}
		if isDigit(l.ch) {
			// fmt.Println("[l.NextToken] isDigit: ", tok.Literal, tok.Type)
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		}
		// fmt.Println("char not matching anything in lexer!!!", l.ch)

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
