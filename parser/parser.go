// Parser package is responsible for parsing the tokens from the lexer and constructing the AST.
//
// It uses the lexer to process tokens one at a time and records any parser errors in Errors slice.
// Registered prefix and infix parsing functions for different token types allow the parser to parse different expressions and statements.
package parser

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/Rishabh570/csvlang/ast"
	"github.com/Rishabh570/csvlang/lexer"
	"github.com/Rishabh570/csvlang/token"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
	ASSIGN      // =
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
	token.ASSIGN:   ASSIGN,
}

type (
	prefixParseFn     func() ast.Expression               // prefixParseFn is a function that holds the custom parsing logic for a prefix token
	infixParseFn      func(ast.Expression) ast.Expression // infixParseFn is a function that holds the custom parsing logic for an infix token
	prefixParseReadFn func() ast.LocationExpression
)

// Parser is responsible for parsing the tokens from the lexer and constructing the AST
// It uses the lexer to process tokens one at a time and records any parser errors in Errors slice
// Registered prefix and infix parsing functions for different token types allows the parser to parse different expressions and statements
type Parser struct {
	l         *lexer.Lexer
	prevToken token.Token
	curToken  token.Token
	peekToken token.Token

	// stores all the parsing errors
	Errors []*ParserError

	prefixParseFns     map[token.TokenType]prefixParseFn
	infixParseFns      map[token.TokenType]infixParseFn
	prefixParseReadFns map[token.TokenType]prefixParseReadFn
}

// TODO: rename to NewParser
// New creates a new Parser instance with the given lexer
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:                  l,
		Errors:             []*ParserError{},
		prefixParseFns:     make(map[token.TokenType]prefixParseFn),
		infixParseFns:      make(map[token.TokenType]infixParseFn),
		prefixParseReadFns: make(map[token.TokenType]prefixParseReadFn),
	}

	// Register only the parse functions we need for now
	// p.registerReadPrefix(token.ROW, p.parseLocationExpression)

	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.READ, p.parseReadAsExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteralAsExpression)
	p.registerPrefix(token.FOR, p.parseForLoopAsExpression)

	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.ASSIGN, p.parseIndexAssignment)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()
	return p
}

// nextToken advances the parser by one token
func (p *Parser) nextToken() {
	p.prevToken = p.curToken
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// addError creates a new ParserError with the given message, line, column, and stack trace
func (p *Parser) addError(message string) {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:]) // Skip first two frames

	er := &ParserError{
		Message: message,
		Stack:   stack[:length],
		Line:    p.l.Line,
		Column:  p.l.Column,
	}

	p.Errors = append(p.Errors, er)
}

// ParseProgram parses the program and returns the AST
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	for p.curToken.Type != token.EOF {
		if p.curTokenIs(token.SINGLE_LINE_COMMENT) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

// parseStatement parses a statement and returns the AST node
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.LOAD:
		return p.parseLoadStatement()
	case token.READ:
		return p.parseReadStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.SAVE:
		return p.parseSaveStatement()
	case token.FOR:
		return p.parseForLoopStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseIndexAssignment(left ast.Expression) ast.Expression {
	// Check if left side is an index expression
	indexExp, ok := left.(*ast.IndexExpression)
	if !ok {
		return nil
	}

	exp := &ast.IndexAssignmentExpression{
		Token: p.curToken,
		Left:  indexExp,
	}

	p.nextToken() // move past '='
	exp.Value = p.parseExpression(LOWEST)

	return exp
}

// Two options:
// 1. save as filtered.csv/filtered.json
// 2. save myCustomRows as filtered.csv/filtered.json
func (p *Parser) parseSaveStatement() *ast.SaveStatement {
	stmt := &ast.SaveStatement{Token: p.curToken}

	if p.peekTokenIs(token.IDENT) {
		stmt.Source = &ast.Identifier{Token: p.peekToken, Value: p.peekToken.Literal}

		p.nextToken()
	}

	if !p.expectPeek(token.AS) {
		return nil
	}
	p.nextToken() // move past AS

	// Parse filename
	if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.STRING) {
		p.addError("expected filename")
		return nil
	}

	stmt.Filename = p.curToken.Literal

	// Determine format from filename extension
	if strings.HasSuffix(stmt.Filename, ".json") {
		stmt.Format = "json"
	} else if strings.HasSuffix(stmt.Filename, ".csv") {
		stmt.Format = "csv"
	} else {
		p.addError("unsupported file format")
		return nil
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseArrayLiteralStatement() ast.Statement {
	array := p.parseArrayLiteral()
	return &ast.ArrayLiteralStatement{ArrayLiteral: array}
}

func (p *Parser) parseArrayLiteral() *ast.ArrayLiteral {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseArrayLiteralAsExpression() ast.Expression {
	return p.parseArrayLiteral()
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()

	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	// If current token is identifier and next token is =, it's an assignment
	// user can choose to reassign values to a variable
	if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.ASSIGN) {
		return p.parseAssignmentStatement()
	}

	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseAssignmentStatement() *ast.AssignmentStatement {
	stmt := &ast.AssignmentStatement{Token: p.curToken}

	// Current token is the identifier
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Move past identifier
	p.nextToken()
	// Move past '='
	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	lit.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	lit.Body = p.parseBlockStatement()
	return lit
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	// move past opening LPAREN "("
	p.nextToken()

	// create an ident for the first param and append to identifiers slice
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		// add the following (second, third, and so on) params to identifiers slice
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return identifiers
}

// func (p *Parser) parseReadPrefixInLetStatement() ast.Expression {
// 	return p.parseReadStatement()
// }

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.addError(msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseLoadStatement() *ast.LoadStatement {
	stmt := &ast.LoadStatement{Token: p.curToken}

	p.nextToken()

	// Parse the filename as an expression instead of identifier
	filename := p.parseExpression(LOWEST)
	if filename == nil {
		return nil
	}
	stmt.Filename = filename

	return stmt
}

func (p *Parser) parseReadStatement() *ast.ReadStatement {
	readExp := p.parseReadExpression()
	return &ast.ReadStatement{ReadExpression: readExp}
}

func (p *Parser) parseReadExpression() *ast.ReadExpression {
	expr := &ast.ReadExpression{Token: p.curToken}

	p.nextToken()

	// Parse location
	location := p.parseLocationExpression()
	expr.Location = location

	return expr
}

// This is for expression usage - implements prefixParseFn
func (p *Parser) parseReadAsExpression() ast.Expression {
	return p.parseReadExpression()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) parseForLoopAsExpression() ast.Expression {
	return p.parseForLoopExpression()
}

func (p *Parser) parseForLoopExpression() *ast.ForLoopExpression {
	stmt := &ast.ForLoopExpression{Token: p.curToken}

	// Parse index identifier
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.IndexName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Parse comma
	if !p.expectPeek(token.COMMA) {
		return nil
	}

	// Parse element identifier
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.ElementName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Parse 'in' keyword
	if !p.expectPeek(token.IN) {
		return nil
	}

	// Parse iterable expression
	p.nextToken()
	stmt.Iterable = p.parseExpression(LOWEST)

	// Parse body
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()

	if !p.curTokenIs(token.RBRACE) {
		return nil
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseForLoopStatement() ast.Statement {
	forLoopExpr := p.parseForLoopExpression()
	return &ast.ForLoopStatement{ForLoopExpression: forLoopExpr}
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.addError(msg)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p

	}
	return LOWEST
}

/*
*
Prefix handlers
*/

// 1. prefix as identifier
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

// 2. prefix as read row(eg. read row 0 OR read row)
// eg.1 => read row 0
// eg. 2 => read row 0 col 0
func (p *Parser) parseLocationExpression() ast.LocationExpression {
	// 1. 🏁🏁🏁 Parse row
	if p.curToken.Type != token.ROW {
		errMsg := fmt.Sprintf("READ: expected first modifier key to be ROW, got %s", p.curToken.Type)
		p.addError(errMsg)
		return ast.LocationExpression{
			RowIndex: -1,
			ColIndex: "",
		}
	}

	p.nextToken()

	if p.curToken.Type != token.INT && p.curToken.Type != token.ASTERISK {
		errMsg := fmt.Sprintf("READ: expected first modifier value to be INT or ASTERISK, got %s", p.curToken.Type)
		p.addError(errMsg)
		return ast.LocationExpression{
			RowIndex: -1,
			ColIndex: "",
		}
	}

	locExpr := ast.LocationExpression{}

	// cur token can be either of the two: INT or ASTERISK
	if p.curToken.Type == token.INT {
		num, err := strconv.Atoi(p.curToken.Literal)
		if err != nil {
			errMsg := fmt.Sprintf("Error converting string to int: %v", err)
			p.addError(errMsg)
			return ast.LocationExpression{
				RowIndex: -1,
				ColIndex: "",
			}
		}
		locExpr.RowIndex = num
	} else {
		locExpr.RowIndex = -2 // -2 is a special value to denote asterisk
	}

	if p.peekTokenIs(token.SEMICOLON) || p.peekTokenIs(token.EOF) {
		p.nextToken()
		return locExpr
	}

	p.nextToken()

	// 2. 🏁🏁🏁 Parse column

	// return error if not "col"
	if !p.curTokenIs(token.COL) && !p.curTokenIs(token.WHERE) {
		errMsg := fmt.Sprintf("READ: expected second modifier to be COL or WHERE, got %s", p.peekToken.Type)
		p.addError(errMsg)
		// reject complete statement, don't process rowIndex even when passed correctly
		// csvlang mandates complete statement to be syntactically correct
		return ast.LocationExpression{
			RowIndex: -1,
			ColIndex: "",
		}
	}

	// If cur token is COL
	if p.curTokenIs(token.COL) {
		p.nextToken()

		locExpr.ColIndex = p.curToken.Literal

		if p.peekTokenIs(token.SEMICOLON) || p.peekTokenIs(token.EOF) {
			p.nextToken()
			return locExpr
		}

		if !p.peekTokenIs(token.WHERE) {
			errMsg := fmt.Sprintf("READ: expected WHERE token to follow COL, got %s", p.peekToken.Type)
			p.addError(errMsg)
			return ast.LocationExpression{
				RowIndex: -1,
				ColIndex: "",
			}
		}

		p.nextToken()
	}

	// 🏁🏁🏁 3. the cur token is WHERE, start parsing the filter expression
	filterExpr := ast.ReadFilterExpression{Token: p.curToken}

	p.nextToken()

	if p.curToken.Type != token.IDENT {
		errMsg := fmt.Sprintf("READ: expected column name to be IDENT, got %s", p.curToken.Type)
		p.addError(errMsg)
		return ast.LocationExpression{
			RowIndex: -1,
			ColIndex: "",
		}
	}
	filterExpr.ColumnName = p.curToken.Literal

	p.nextToken()

	if p.curToken.Type != token.EQ &&
		p.curToken.Type != token.NOT_EQ &&
		p.curToken.Type != token.LT &&
		p.curToken.Type != token.GT {
		errMsg := fmt.Sprintf("READ: expected operator to be one of [EQ, NOT_EQ, LT, GT] got %s", p.curToken.Type)
		p.addError(errMsg)
		return ast.LocationExpression{
			RowIndex: -1,
			ColIndex: "",
		}
	}
	filterExpr.Operator = p.curToken.Literal

	p.nextToken()

	if p.curToken.Type != token.STRING && p.curToken.Type != token.INT {
		errMsg := fmt.Sprintf("READ: expected value to be either STRING or INT, got %s", p.curToken.Type)
		p.addError(errMsg)
		return ast.LocationExpression{
			RowIndex: -1,
			ColIndex: "",
		}
	}
	// filterExpr.Value =  p.curToken.Literal
	filterExpr.Value = p.parseExpression(LOWEST)

	locExpr.Filter = &filterExpr

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return locExpr
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerReadPrefix(tokenType token.TokenType, fn prefixParseReadFn) {
	p.prefixParseReadFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.addError(msg)
}
