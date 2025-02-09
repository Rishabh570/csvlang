package ast

import (
	"fmt"

	"github.com/Rishabh570/csvlang/token"
)

func ExampleIdentifier() {
	ident := &Identifier{
		Token: token.Token{Type: token.IDENT, Literal: "myVar"},
		Value: "myVar",
	}
	fmt.Println(ident.String())
	// Output: myVar
}

func ExampleExpressionStatement() {
	exprStmt := &ExpressionStatement{
		Token: token.Token{Type: token.INT, Literal: "5"},
		Expression: &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "5"},
			Value: 5,
		},
	}
	fmt.Println(exprStmt.String())
	// Output: 5
}

func ExampleAssignmentStatement() {
	assignStmt := &AssignmentStatement{
		Token: token.Token{Type: token.ASSIGN, Literal: "="},
		Name: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "myVar"},
			Value: "myVar",
		},
		Value: &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "5"},
			Value: 5,
		},
	}
	fmt.Println(assignStmt.String())
	// Output: myVar = 5
}

func ExampleLoadStatement() {
	loadStmt := &LoadStatement{
		Token: token.Token{
			Type:    token.LOAD,
			Literal: "load",
		},
		Filename: &StringLiteral{
			Token: token.Token{
				Type:    token.STRING,
				Literal: "\"config.csv\"",
			},
			Value: "config.csv",
		},
	}
	fmt.Println(loadStmt.String())
	// Output: load "config.csv"
}

func ExampleReadExpression() {
	readExpr := &ReadExpression{
		Token: token.Token{
			Type:    token.READ,
			Literal: "read",
		},
		Location: LocationExpression{
			Token: token.Token{
				Type:    token.READ_FILTER_LOCATION,
				Literal: "readFilterLocation",
			},
			RowIndex: 0,
			ColIndex: "name",
			Filter: &ReadFilterExpression{
				Token: token.Token{
					Type:    token.WHERE,
					Literal: "where",
				},
				ColumnName: "age",
				Operator:   "==",
				Value: &StringLiteral{
					Token: token.Token{
						Type:    token.STRING,
						Literal: "hello",
					},
					Value: "hello",
				},
			},
		},
	}
	fmt.Println(readExpr.String())
	// Output: read Row: 0, Column: name, Where: Column: age, Operator: ==, Value: hello
}

func ExampleReadFilterExpression() {
	readFilterExpr := &ReadFilterExpression{
		Token: token.Token{
			Type:    token.WHERE,
			Literal: "where",
		},
		ColumnName: "age",
		Operator:   "==",
		Value: &StringLiteral{
			Token: token.Token{
				Type:    token.STRING,
				Literal: "hello",
			},
			Value: "hello",
		},
	}
	fmt.Println(readFilterExpr.String())
	// Output: Column: age, Operator: ==, Value: hello
}

func ExampleLocationExpression() {
	locExpr := LocationExpression{
		Token: token.Token{
			Type:    token.READ_FILTER_LOCATION,
			Literal: "readFilterLocation",
		},
		RowIndex: 0,
		ColIndex: "name",
		Filter: &ReadFilterExpression{
			Token: token.Token{
				Type:    token.WHERE,
				Literal: "where",
			},
			ColumnName: "age",
			Operator:   "==",
			Value: &StringLiteral{
				Token: token.Token{
					Type:    token.STRING,
					Literal: "hello",
				},
				Value: "hello",
			},
		},
	}
	fmt.Println(locExpr.String())
	// Output: Row: 0, Column: name, Where: Column: age, Operator: ==, Value: hello
}

func ExampleLetStatement() {
	letStmt := &LetStatement{
		Token: token.Token{Type: token.LET, Literal: "let"},
		Name: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "myVar"},
			Value: "myVar",
		},
		Value: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
			Value: "anotherVar",
		},
	}
	fmt.Println(letStmt.String())
	// Output: let myVar = anotherVar
}

func ExampleReturnStatement() {
	returnStmt := &ReturnStatement{
		Token: token.Token{Type: token.RETURN, Literal: "return"},
		ReturnValue: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "myVar"},
			Value: "myVar",
		},
	}
	fmt.Println(returnStmt.String())
	// Output: return myVar
}

func ExampleIntegerLiteral() {
	intLit := &IntegerLiteral{
		Token: token.Token{Type: token.INT, Literal: "5"},
		Value: 5,
	}
	fmt.Println(intLit.String())
	// Output: 5
}

func ExamplePrefixExpression() {
	prefixExpr := &PrefixExpression{
		Token:    token.Token{Type: token.BANG, Literal: "!"},
		Operator: "!",
		Right: &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "5"},
			Value: 5,
		},
	}
	fmt.Println(prefixExpr.String())
	// Output: (!5)
}

func ExampleInfixExpression() {
	infixExpr := &InfixExpression{
		Token: token.Token{Type: token.PLUS, Literal: "+"},
		Left: &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "5"},
			Value: 5,
		},
		Operator: "+",
		Right: &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "10"},
			Value: 10,
		},
	}
	fmt.Println(infixExpr.String())
	// Output: (5 + 10)
}

func ExampleBoolean() {
	boolExpr := &Boolean{
		Token: token.Token{Type: token.TRUE, Literal: "true"},
		Value: true,
	}
	fmt.Println(boolExpr.String())
	// Output: true
}

func ExampleIfExpression() {
	ifExpr := &IfExpression{
		Token: token.Token{Type: token.IF, Literal: "if"},
		Condition: &Boolean{
			Token: token.Token{Type: token.TRUE, Literal: "true"},
			Value: true,
		},
		Consequence: &BlockStatement{
			Token: token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{
				&ExpressionStatement{
					Token: token.Token{Type: token.INT, Literal: "5"},
					Expression: &IntegerLiteral{
						Token: token.Token{Type: token.INT, Literal: "5"},
						Value: 5,
					},
				},
			},
		},
		Alternative: &BlockStatement{
			Token: token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{
				&ExpressionStatement{
					Token: token.Token{Type: token.INT, Literal: "10"},
					Expression: &IntegerLiteral{
						Token: token.Token{Type: token.INT, Literal: "10"},
						Value: 10,
					},
				},
			},
		},
	}
	fmt.Println(ifExpr.String())
	// Output: if true { 5 } else { 10 }
}

func ExampleBlockStatement() {
	blockStmt := &BlockStatement{
		Token: token.Token{Type: token.LBRACE, Literal: "{"},
		Statements: []Statement{
			&ExpressionStatement{
				Token: token.Token{Type: token.INT, Literal: "5"},
				Expression: &IntegerLiteral{
					Token: token.Token{Type: token.INT, Literal: "5"},
					Value: 5,
				},
			},
		},
	}
	fmt.Println(blockStmt.String())
	// Output: { 5 }
}

func ExampleFunctionLiteral() {
	funcLit := &FunctionLiteral{
		Token: token.Token{Type: token.FUNCTION, Literal: "fn"},
		Parameters: []*Identifier{
			{
				Token: token.Token{Type: token.IDENT, Literal: "x"},
				Value: "x",
			},
		},
		Body: &BlockStatement{
			Token: token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{
				&ExpressionStatement{
					Token: token.Token{Type: token.INT, Literal: "5"},
					Expression: &IntegerLiteral{
						Token: token.Token{Type: token.INT, Literal: "5"},
						Value: 5,
					},
				},
			},
		},
	}
	fmt.Println(funcLit.String())
	// Output: fn(x) { 5 }
}

func ExampleCallExpression() {
	callExpr := &CallExpression{
		Token: token.Token{Type: token.LPAREN, Literal: "("},
		Function: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "add"},
			Value: "add",
		},
		Arguments: []Expression{
			&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "5"},
				Value: 5,
			},
			&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "10"},
				Value: 10,
			},
		},
	}
	fmt.Println(callExpr.String())
	// Output: add(5, 10)
}

func ExampleStringLiteral() {
	strLit := &StringLiteral{
		Token: token.Token{Type: token.STRING, Literal: "\"hello\""},
		Value: "hello",
	}
	fmt.Println(strLit.String())
	// Output: "hello"
}

func ExampleArrayLiteral() {
	arrayLit := &ArrayLiteral{
		Token: token.Token{Type: token.LBRACKET, Literal: "["},
		Elements: []Expression{
			&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "1"},
				Value: 1,
			},
			&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "2"},
				Value: 2,
			},
		},
	}
	fmt.Println(arrayLit.String())
	// Output: [1, 2]
}

func ExampleArrayLiteralStatement() {
	arrayLitStmt := &ArrayLiteralStatement{
		&ArrayLiteral{
			Token: token.Token{Type: token.LBRACKET, Literal: "["},
			Elements: []Expression{
				&IntegerLiteral{
					Token: token.Token{Type: token.INT, Literal: "1"},
					Value: 1,
				},
				&IntegerLiteral{
					Token: token.Token{Type: token.INT, Literal: "2"},
					Value: 2,
				},
			},
		},
	}
	fmt.Println(arrayLitStmt.String())
	// Output: [1, 2]
}

func ExampleIndexExpression() {
	indexExpr := &IndexExpression{
		Token: token.Token{Type: token.LBRACKET, Literal: "["},
		Left: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "myArray"},
			Value: "myArray",
		},
		Index: &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "0"},
			Value: 0,
		},
	}
	fmt.Println(indexExpr.String())
	// Output: myArray[0]
}

func ExampleSaveStatement() {
	saveStmt := &SaveStatement{
		Token: token.Token{Type: token.SAVE, Literal: "save"},
		Source: &StringLiteral{
			Token: token.Token{Type: token.STRING, Literal: "\"config.csv\""},
			Value: "config.csv",
		},
		Filename: "output.csv",
		Format:   "csv",
	}
	fmt.Println(saveStmt.String())
	// Output: save "config.csv" as output.csv
}

func ExampleForLoopExpression() {
	forLoopExpr := &ForLoopExpression{
		Token: token.Token{Type: token.FOR, Literal: "for"},
		IndexName: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "i"},
			Value: "i",
		},
		ElementName: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "element"},
			Value: "element",
		},
		Iterable: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "myArray"},
			Value: "myArray",
		},
		Body: &BlockStatement{
			Token: token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{
				&ExpressionStatement{
					Token: token.Token{Type: token.INT, Literal: "5"},
					Expression: &IntegerLiteral{
						Token: token.Token{Type: token.INT, Literal: "5"},
						Value: 5,
					},
				},
			},
		},
	}
	fmt.Println(forLoopExpr.String())
	// Output: for i, element in myArray 5
}

func ExampleForLoopStatement() {
	forLoopStmt := &ForLoopStatement{
		&ForLoopExpression{
			Token: token.Token{Type: token.FOR, Literal: "for"},
			IndexName: &Identifier{
				Token: token.Token{Type: token.IDENT, Literal: "i"},
				Value: "i",
			},
			ElementName: &Identifier{
				Token: token.Token{Type: token.IDENT, Literal: "element"},
				Value: "element",
			},
			Iterable: &Identifier{
				Token: token.Token{Type: token.IDENT, Literal: "myArray"},
				Value: "myArray",
			},
			Body: &BlockStatement{
				Token: token.Token{Type: token.LBRACE, Literal: "{"},
				Statements: []Statement{
					&ExpressionStatement{
						Token: token.Token{Type: token.INT, Literal: "5"},
						Expression: &IntegerLiteral{
							Token: token.Token{Type: token.INT, Literal: "5"},
							Value: 5,
						},
					},
				},
			},
		},
	}
	fmt.Println(forLoopStmt.String())
	// Output: for i, element in myArray 5
}

func ExampleIndexAssignmentExpression() {
	indexAssignExpr := &IndexAssignmentExpression{
		Token: token.Token{Type: token.LBRACKET, Literal: "="},
		Left: &IndexExpression{
			Token: token.Token{Type: token.LBRACKET, Literal: "["},
			Left: &Identifier{
				Token: token.Token{Type: token.IDENT, Literal: "myArray"},
				Value: "myArray",
			},
			Index: &IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "0"},
				Value: 0,
			},
		},
		Value: &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "5"},
			Value: 5,
		},
	}
	fmt.Println(indexAssignExpr.String())
	// Output: (myArray[0]) = 5
}
