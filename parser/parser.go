package parser

import (
	"csvlang/ast"
	"csvlang/lexer"
	"csvlang/token"
	"fmt"
	"strconv"
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
}

type Parser struct {
	l         *lexer.Lexer
	prevToken token.Token
	curToken  token.Token
	peekToken token.Token

	// stores all the parsing errors
	Errors []string

	prefixParseFns     map[token.TokenType]prefixParseFn
	infixParseFns      map[token.TokenType]infixParseFn
	prefixParseReadFns map[token.TokenType]prefixParseReadFn
}

type (
	prefixParseFn     func() ast.Expression
	infixParseFn      func(ast.Expression) ast.Expression
	prefixParseReadFn func() ast.LocationExpression
)

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:                  l,
		Errors:             []string{},
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
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)

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

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.prevToken = p.curToken
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	for p.curToken.Type != token.EOF {
		if p.curTokenIs(token.SINGLE_LINE_COMMENT) {
			p.nextToken()
			continue
		}
		fmt.Printf("[ParseProgram] starting... p.curToken.Type: %s, p.curToken.Literal: %s\n", p.curToken.Type, p.curToken.Literal)

		stmt := p.parseStatement()
		fmt.Printf("[ParseProgram] parsed stmt: %s\n", stmt)
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		fmt.Println("[parseStatement] parsing LET stmt...")
		return p.parseLetStatement()
	case token.LOAD:
		fmt.Println("[parseStatement] parsing LOAD stmt...")
		return p.parseLoadStatement()
	case token.READ:
		fmt.Println("[parseStatement] parsing READ stmt...")
		return p.parseReadStatement()
	case token.APPEND:
		return p.parseAppendStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
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

func (p *Parser) parseAppendStatement() *ast.AppendStatement {
	stmt := &ast.AppendStatement{Token: p.curToken}
	stmt.Values = []ast.Expression{}

	p.nextToken() // move past 'append'

	// Check if we've reached EOF or no values provided
	if p.curTokenIs(token.EOF) {
		p.Errors = append(p.Errors, "incomplete APPEND statement: no values provided")
		return nil
	}

	// Parse comma-separated values
	for !p.curTokenIs(token.EOF) && !p.curTokenIs(token.SEMICOLON) {
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
			continue
		}

		value := p.parseExpression(LOWEST)
		if value != nil {
			stmt.Values = append(stmt.Values, value)
		}

		p.nextToken()
	}

	if len(stmt.Values) == 0 {
		p.Errors = append(p.Errors, "WARN: no values to append")
		return nil
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	fmt.Printf("[parseExpression] tok: %s, %s\n", p.curToken.Type, p.curToken.Literal)
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		fmt.Println("[parseExpression] no prefix parse fn...\n")
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()
	fmt.Printf("[parseExpression] leftExp: %s\n", leftExp)

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		fmt.Printf("parsing infix token, type: %s, lit: %s\n", p.peekToken.Type, p.peekToken.Literal)
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	fmt.Printf("[parseExpression] returning leftExpr: %s\n", leftExp)
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
	fmt.Printf("[parseIfExpression] starting...\n")
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

	fmt.Printf("[parseIfExpression] returning expr: %s, consequence: %s, stmts: %+v\n", expression.Condition.String(), expression.Consequence.String(), expression.Consequence.Statements)
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
		p.Errors = append(p.Errors, msg)
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
	fmt.Printf("[parseLetStatement] type: %s, lit: %s\n", p.curToken.Type, p.curToken.Literal)
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
	fmt.Printf("[parseLetStatement] stmt.Value: %s\n", stmt.Value.String())
	return stmt
}

func (p *Parser) parseLoadStatement() *ast.LoadStatement {
	fmt.Printf("\n[ParseLoad] cur token: %s, %s\n", p.curToken.Type, p.curToken.Literal)
	stmt := &ast.LoadStatement{Token: p.curToken}

	p.nextToken()

	fmt.Printf("\n[ParseLoad] cur token: %s, %s\n", p.curToken.Type, p.curToken.Literal)

	// Parse the filename as an expression instead of identifier
	filename := p.parseExpression(LOWEST)
	fmt.Printf("filenameee: %s\n", filename.TokenLiteral())
	if filename == nil {
		return nil
	}
	stmt.Filename = filename

	fmt.Printf("returning load stmt: type: %s, lit: %s, filename: %s, stmt: %s\n", stmt.Token.Type, stmt.Token.Literal, stmt.Filename.String(), stmt.String())
	return stmt
}

func (p *Parser) parseReadStatement() *ast.ReadStatement {
	fmt.Printf("[parseReadStatement] starting...")
	readExp := p.parseReadExpression()
	return &ast.ReadStatement{ReadExpression: readExp}
}

func (p *Parser) parseReadExpression() *ast.ReadExpression {
	fmt.Printf("[parseReadExpression] tok.type: %s, tok.lit: %s\n", p.curToken.Type, p.curToken.Literal)
	expr := &ast.ReadExpression{Token: p.curToken}

	p.nextToken()

	// Parse location
	location := p.parseLocationExpression()
	fmt.Printf("[parseReadExpression] location expr: %s\n", location.String())
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

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.Errors = append(p.Errors, msg)
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

// func (p *Parser) Errors() []string {
// return p.errors
// }

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
	fmt.Println("parsing location expr...\n", p.curToken.Type, p.curToken.Literal)

	// 1. 🏁🏁🏁 Parse row
	if p.curToken.Type != token.ROW {
		errMsg := fmt.Sprintf("READ: expected first modifier key to be ROW, got %s", p.curToken.Type)
		p.Errors = append(p.Errors, errMsg)
		return ast.LocationExpression{
			RowIndex: -1,
			ColIndex: "",
		}
	}

	p.nextToken()

	if p.curToken.Type != token.INT && p.curToken.Type != token.ASTERISK {
		errMsg := fmt.Sprintf("READ: expected first modifier value to be INT or ASTERISK, got %s", p.curToken.Type)
		p.Errors = append(p.Errors, errMsg)
		return ast.LocationExpression{
			RowIndex: -1,
			ColIndex: "",
		}
	}

	// fmt.Printf("[parseLocationExpression] curTOkennnn: %s, %s\n", p.curToken.Type, p.curToken.Literal)

	locExpr := ast.LocationExpression{}

	// cur token can be either of the two: INT or ASTERISK
	if p.curToken.Type == token.INT {
		num, err := strconv.Atoi(p.curToken.Literal)
		if err != nil {
			errMsg := fmt.Sprintf("Error converting string to int: %v", err)
			p.Errors = append(p.Errors, errMsg)
			return ast.LocationExpression{
				RowIndex: -1,
				ColIndex: "",
			}
		}
		locExpr.RowIndex = num
	} else {
		locExpr.RowIndex = -2 // -2 is a special value to denote asterisk
	}

	// fmt.Println("[parseLocationExpression] locExpr: ", locExpr.RowIndex, locExpr.ColIndex)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
		return locExpr
	}

	p.nextToken()

	// 2. 🏁🏁🏁 Parse column
	// fmt.Printf("[parseLocationExpression] curTOken: %s, %s\n", p.curToken.Type, p.curToken.Literal)

	// return error if not "col"
	if !p.curTokenIs(token.COL) && !p.curTokenIs(token.WHERE) {
		errMsg := fmt.Sprintf("READ: expected second modifier to be COL or WHERE, got %s", p.peekToken.Type)
		p.Errors = append(p.Errors, errMsg)
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

		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
			return locExpr
		}

		if !p.peekTokenIs(token.EOF) && !p.peekTokenIs(token.WHERE) {
			errMsg := fmt.Sprintf("READ: expected WHERE token to follow COL, got %s", p.peekToken.Type)
			p.Errors = append(p.Errors, errMsg)
			return ast.LocationExpression{
				RowIndex: -1,
				ColIndex: "",
			}
		}

		p.nextToken()
	}

	// 🏁🏁🏁 3. the cur token is WHERE, start parsing the filter expression
	filterExpr := ast.ReadFilterExpression{Token: p.curToken}
	// fmt.Printf("[parseLocationExpression] curTken: %s, %s\n", p.curToken.Type, p.curToken.Literal)

	p.nextToken()

	// fmt.Printf("[parseLocationExpression] curTken: %s, %s\n", p.curToken.Type, p.curToken.Literal)
	if p.curToken.Type != token.IDENT {
		errMsg := fmt.Sprintf("READ: expected column name to be IDENT, got %s", p.curToken.Type)
		p.Errors = append(p.Errors, errMsg)
		return ast.LocationExpression{
			RowIndex: -1,
			ColIndex: "",
		}
	}
	filterExpr.ColumnName = p.curToken.Literal

	p.nextToken()

	// fmt.Printf("[parseLocationExpression] curTken: %s, %s\n", p.curToken.Type, p.curToken.Literal)
	if p.curToken.Type != token.EQ &&
		p.curToken.Type != token.NOT_EQ &&
		p.curToken.Type != token.LT &&
		p.curToken.Type != token.GT {
		errMsg := fmt.Sprintf("READ: expected operator to be one of [EQ, NOT_EQ, LT, GT] got %s", p.curToken.Type)
		p.Errors = append(p.Errors, errMsg)
		return ast.LocationExpression{
			RowIndex: -1,
			ColIndex: "",
		}
	}
	filterExpr.Operator = p.curToken.Literal

	p.nextToken()

	// fmt.Printf("[parseLocationExpression] curTken: %s, %s\n", p.curToken.Type, p.curToken.Literal)
	if p.curToken.Type != token.STRING && p.curToken.Type != token.INT {
		errMsg := fmt.Sprintf("READ: expected value to be either STRING or INT, got %s", p.curToken.Type)
		p.Errors = append(p.Errors, errMsg)
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
	p.Errors = append(p.Errors, msg)
}
