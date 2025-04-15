package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Rishabh570/csvlang/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

/*
*
Identifier
*/
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

/*
*
Expression statement
eg.: `x+5;` most scripting langs allow it so we'll allow it too
*/
type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

/*
*
For re-assignment. Example:
let var = 5
var = 10 <-----
*/
type AssignmentStatement struct {
	Token token.Token // the identifier token
	Name  *Identifier
	Value Expression
}

func (as *AssignmentStatement) statementNode()       {}
func (as *AssignmentStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignmentStatement) String() string {
	var out bytes.Buffer
	out.WriteString(as.Name.String())
	out.WriteString(" = ")
	if as.Value != nil {
		out.WriteString(as.Value.String())
	}
	return out.String()
}

/*
*
Load statement
*/
type LoadStatement struct {
	Token    token.Token // the token.LOAD token
	Filename Expression
}

func (ls *LoadStatement) statementNode()       {}
func (ls *LoadStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LoadStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	if ls.Filename != nil {
		out.WriteString(ls.Filename.String())
	}

	return out.String()
}

// Create a new ReadExpression type that implements Expression interface
type ReadExpression struct {
	Token    token.Token
	Location LocationExpression
}

func (re *ReadExpression) expressionNode()      {}
func (re *ReadExpression) TokenLiteral() string { return re.Token.Literal }
func (re *ReadExpression) String() string {
	var out bytes.Buffer
	out.WriteString(re.TokenLiteral() + " ")
	if re.Location.String() != "" {
		out.WriteString(re.Location.String())
	}
	return out.String()
}

/*
*
Read statement
*/
type ReadStatement struct {
	*ReadExpression
	// Token    token.Token // the token.Read token
	// Location LocationExpression
}

func (rs *ReadStatement) statementNode()       {}
func (rs *ReadStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReadStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.Location.String() != "" {
		out.WriteString(rs.Location.String())
	}
	return out.String()
}

type ReadFilterExpression struct {
	Token      token.Token // the token.WHERE token
	ColumnName string
	Operator   string
	Value      Expression
}

func (le *ReadFilterExpression) expressionNode()      {}
func (le *ReadFilterExpression) TokenLiteral() string { return le.Token.Literal }
func (le *ReadFilterExpression) String() string {
	return fmt.Sprintf("Column: %s, Operator: %s, Value: %s", le.ColumnName, le.Operator, le.Value)
}

/*
*
LocationExpression
*/
type LocationExpression struct {
	Token    token.Token // the token.IDENT token
	RowIndex int
	// ColIndex int16
	ColIndex string
	Filter   *ReadFilterExpression
}

func (le *LocationExpression) expressionNode()      {}
func (le *LocationExpression) TokenLiteral() string { return le.Token.Literal }
func (le *LocationExpression) String() string {
	return fmt.Sprintf("Row: %d, Column: %s", le.RowIndex, le.ColIndex)
}

/*
*
Append statement
*/
type AppendStatement struct {
	Token  token.Token // the token.LOAD token
	Values []Expression
}

func (as *AppendStatement) statementNode()       {}
func (as *AppendStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AppendStatement) String() string {
	var out bytes.Buffer
	vals := []string{}
	for _, v := range as.Values {
		vals = append(vals, v.String())
	}
	out.WriteString("append " + strings.Join(vals, ","))
	return out.String()
}

/*
*
Let statement
*/
type LetStatement struct {
	Token token.Token // the token.LET token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")
	return out.String()
}

/*
*
Return statement
*/
type ReturnStatement struct {
	Token       token.Token // the 'return' token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

/*
*
Integer Expression
*/
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

/*
*
Prefix expression
*/
type PrefixExpression struct {
	Token token.Token // The prefix token, e.g. !

	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

/*
*
Infix expr
*/
type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (oe *InfixExpression) expressionNode()      {}
func (oe *InfixExpression) TokenLiteral() string { return oe.Token.Literal }
func (oe *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(oe.Left.String())
	out.WriteString(" " + oe.Operator + " ")
	out.WriteString(oe.Right.String())
	out.WriteString(")")
	return out.String()
}

/*
*
Boolean expr
*/
type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

/*
*
If Conditional expr
*/
type IfExpression struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if")
	out.WriteString(ie.Condition.String())

	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}
	return out.String()
}

/*
*
Block statement
*/
type BlockStatement struct {
	Token      token.Token // the { token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

/*
*
Function literal
*/
type FunctionLiteral struct {
	Token      token.Token // The 'fn' token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())
	return out.String()
}

/*
*
Call expr
*/
type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

/*
*
String
*/
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

// Array Literal Expression
type ArrayLiteral struct {
	Token    token.Token // the '[' token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// ArrayLiteralStatement is just like ArrayLiteral but is used as a statement
type ArrayLiteralStatement struct {
	*ArrayLiteral
}

func (rs *ArrayLiteralStatement) statementNode()       {}
func (rs *ArrayLiteralStatement) TokenLiteral() string { return rs.Token.Literal }

// Index Expression for accessing array elements
type IndexExpression struct {
	Token token.Token // The '[' token
	Left  Expression  // The array being indexed
	Index Expression  // The index value
}

func (al *IndexExpression) expressionNode()      {}
func (al *IndexExpression) TokenLiteral() string { return al.Token.Literal }
func (al *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(al.Left.String())
	out.WriteString("[")
	out.WriteString(al.Index.String())
	out.WriteString("])")
	return out.String()
}

type SaveStatement struct {
	Token    token.Token // the token.SAVE token
	Source   Expression  // Optional: identifier for custom rows
	Filename string
	Format   string // "csv" or "json"
}

func (al *SaveStatement) statementNode()       {}
func (al *SaveStatement) TokenLiteral() string { return al.Token.Literal }
func (ss *SaveStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ss.TokenLiteral() + " ")
	if ss.Source != nil {
		out.WriteString(ss.Source.String() + " as ")
	}
	out.WriteString(ss.Filename)
	return out.String()
}

// ForLoopExpression for iterating over arrays
type ForLoopExpression struct {
	Token       token.Token
	IndexName   *Identifier
	ElementName *Identifier
	Iterable    Expression
	Body        *BlockStatement
}

func (fl *ForLoopExpression) expressionNode()      {}
func (fl *ForLoopExpression) TokenLiteral() string { return fl.Token.Literal }
func (fl *ForLoopExpression) String() string {
	var out bytes.Buffer
	out.WriteString("for ")
	out.WriteString(fl.IndexName.String())
	out.WriteString(", ")
	out.WriteString(fl.ElementName.String())
	out.WriteString(" in ")
	out.WriteString(fl.Iterable.String())
	out.WriteString(" ")
	out.WriteString(fl.Body.String())
	return out.String()
}

// ForLoopStatement is just like ForLoopExpression but is used as a statement
type ForLoopStatement struct {
	*ForLoopExpression
}

func (fl *ForLoopStatement) statementNode()       {}
func (fl *ForLoopStatement) TokenLiteral() string { return fl.Token.Literal }
func (fl *ForLoopStatement) String() string {
	var out bytes.Buffer
	out.WriteString("for ")
	out.WriteString(fl.IndexName.String())
	out.WriteString(", ")
	out.WriteString(fl.ElementName.String())
	out.WriteString(" in ")
	out.WriteString(fl.Iterable.String())
	out.WriteString(" ")
	out.WriteString(fl.Body.String())
	return out.String()
}

// IndexAssignmentExpression for re-assigning values
type IndexAssignmentExpression struct {
	Token token.Token // The '=' token
	Left  *IndexExpression
	Value Expression
}

func (iae *IndexAssignmentExpression) expressionNode()      {}
func (iae *IndexAssignmentExpression) TokenLiteral() string { return iae.Token.Literal }
func (iae *IndexAssignmentExpression) String() string {
	var out bytes.Buffer
	out.WriteString(iae.Left.String())
	out.WriteString(" = ")
	out.WriteString(iae.Value.String())
	return out.String()
}
