package lexer

import (
	"testing"

	"github.com/Rishabh570/csvlang/token"
)

func TestNextTokenOne(t *testing.T) {
	input := `
	load input.csv
	read row 0 col 0
	read row 0
	read row 0 where age > 12
	read row 0 col age
	read row *
	read row * where age > 12
	read row * col name
	read row * col name where age > 12
	#thisisacomment
	5 == 5
	let five = 5
	5 + 5
	5 - 5
	5 != 6
	5 / 5
	5 * 5
	5 < 5
	5 > 5
	true;
	let arr = [1,2]
	fn(x, y) { x + y; };
	!true
	"foobar"
	save as output.csv
	save as output.json
	save myRows as output.csv
	save myRows as output.json
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LOAD, "load"},
		{token.IDENT, "input.csv"},

		// read row 0 col 0
		{token.READ, "read"},
		{token.ROW, "row"},
		{token.INT, "0"},
		{token.COL, "col"},
		{token.INT, "0"},

		// read row 0
		{token.READ, "read"},
		{token.ROW, "row"},
		{token.INT, "0"},

		// read row 0 where age > 12
		{token.READ, "read"},
		{token.ROW, "row"},
		{token.INT, "0"},
		{token.WHERE, "where"},
		{token.IDENT, "age"},
		{token.GT, ">"},
		{token.INT, "12"},

		// read row 0 col age
		{token.READ, "read"},
		{token.ROW, "row"},
		{token.INT, "0"},
		{token.COL, "col"},
		{token.IDENT, "age"},

		// read row *
		{token.READ, "read"},
		{token.ROW, "row"},
		{token.ASTERISK, "*"},

		// read row * where age > 12
		{token.READ, "read"},
		{token.ROW, "row"},
		{token.ASTERISK, "*"},
		{token.WHERE, "where"},
		{token.IDENT, "age"},
		{token.GT, ">"},
		{token.INT, "12"},

		// read row * col name
		{token.READ, "read"},
		{token.ROW, "row"},
		{token.ASTERISK, "*"},
		{token.COL, "col"},
		{token.IDENT, "name"},

		// read row * col name where age > 12
		{token.READ, "read"},
		{token.ROW, "row"},
		{token.ASTERISK, "*"},
		{token.COL, "col"},
		{token.IDENT, "name"},
		{token.WHERE, "where"},
		{token.IDENT, "age"},
		{token.GT, ">"},
		{token.INT, "12"},

		// thisisacomment
		{token.SINGLE_LINE_COMMENT, "thisisacomment"},

		// 5 == 5
		{token.INT, "5"},
		{token.EQ, "=="},
		{token.INT, "5"},

		// let five = 5;
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},

		// 5 + 5
		{token.INT, "5"},
		{token.PLUS, "+"},
		{token.INT, "5"},

		// 5 - 5
		{token.INT, "5"},
		{token.MINUS, "-"},
		{token.INT, "5"},

		// 5 != 6
		{token.INT, "5"},
		{token.NOT_EQ, "!="},
		{token.INT, "6"},

		// 5 / 5
		{token.INT, "5"},
		{token.SLASH, "/"},
		{token.INT, "5"},

		// 5 * 5
		{token.INT, "5"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},

		// 5 < 5
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "5"},

		// 5 > 5
		{token.INT, "5"},
		{token.GT, ">"},
		{token.INT, "5"},

		// true;
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},

		// let arr = [1,2]
		{token.LET, "let"},
		{token.IDENT, "arr"},
		{token.ASSIGN, "="},
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},

		// fn(x, y) { x + y; };
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},

		// !true
		{token.BANG, "!"},
		{token.TRUE, "true"},

		// "foobar"
		{token.STRING, "foobar"},

		// save as output.csv
		{token.SAVE, "save"},
		{token.AS, "as"},
		{token.IDENT, "output.csv"},

		// save as output.json
		{token.SAVE, "save"},
		{token.AS, "as"},
		{token.IDENT, "output.json"},

		// save myRows as output.csv
		{token.SAVE, "save"},
		{token.IDENT, "myRows"},
		{token.AS, "as"},
		{token.IDENT, "output.csv"},

		// save myRows as output.json
		{token.SAVE, "save"},
		{token.IDENT, "myRows"},
		{token.AS, "as"},
		{token.IDENT, "output.json"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got =%q", i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got =%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextTokenTwo(t *testing.T) {
	input := `let five = 5;
	let ten = 10;
	let add = fn(x, y) {
		x + y;
	};

	let result = add(five, ten);
	!-/*5;
	5 < 10 > 5;

	if (5 < 10) {
		return true;
	} else {
		return false;
	}

	10 == 10;
	10 != 9;
	"foobar"
"foo bar"
	`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.STRING, "foobar"},
		{token.STRING, "foo bar"},
		{token.EOF, ""},
		{token.EOF, ""},
	}
	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
