// object/object.go
package object

import (
	"bytes"
	"csvlang/ast"
	"fmt"
	"strconv"
	"strings"
)

type ObjectType string

const (
	NULL_OBJ         = "NULL"
	ERROR_OBJ        = "ERROR"
	CSV_OBJ          = "CSV"
	CSV_ROW          = "CSV_ROW"
	CSV_VAL          = "CSV_VAL"
	STRING_OBJ       = "STRING"
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	FUNCTION_OBJ     = "FUNCTION"

	BUILTIN_OBJ = "BUILTIN"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type BuiltinFunction func(args ...Object) Object

// Integer
type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

// String
type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

// Bool
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

// Null object
type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

// Return
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

// Error object
type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

// Built-in functionality to our lang which the host lang (Go) doesn't provide
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }

// Function object
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")
	return out.String()
}

// Stores data type info about columns in a CSV
type ColumnType struct {
	Name     string
	DataType ObjectType // STRING_OBJ or INTEGER_OBJ
}

// CSV object
type CSV struct {
	Headers     []string
	ColumnTypes []ColumnType
	Rows        []map[string]string
}

func (c *CSV) Type() ObjectType { return CSV_OBJ }
func (c *CSV) Inspect() string {
	// Determine the width of each column
	colWidths := make(map[string]int)
	for _, header := range c.Headers {
		colWidths[header] = len(header)
	}

	for _, row := range c.Rows {
		for _, header := range c.Headers {
			if len(row[header]) > colWidths[header] {
				colWidths[header] = len(row[header])
			}
		}
	}

	// Create a builder to efficiently build the string
	var builder strings.Builder

	// Build the header row
	for _, header := range c.Headers {
		builder.WriteString(fmt.Sprintf("%-*s ", colWidths[header], header))
	}
	builder.WriteString("\n")

	// Build a separator row
	for _, header := range c.Headers {
		builder.WriteString(strings.Repeat("-", colWidths[header]) + " ")
	}
	builder.WriteString("\n")

	// Build each row of data
	for _, row := range c.Rows {
		for _, header := range c.Headers {
			builder.WriteString(fmt.Sprintf("%-*s ", colWidths[header], row[header]))
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
func (c *CSV) InferColumnTypes() {
	if len(c.Rows) == 0 {
		return
	}

	firstRow := c.Rows[0]
	c.ColumnTypes = make([]ColumnType, len(c.Headers))

	for i, header := range c.Headers {
		value := firstRow[header]
		if _, err := strconv.Atoi(value); err == nil {
			c.ColumnTypes[i] = ColumnType{Name: header, DataType: INTEGER_OBJ}
		} else {
			c.ColumnTypes[i] = ColumnType{Name: header, DataType: STRING_OBJ}
		}
	}
}

type CSVRow struct {
	Row map[string]string
}

func (c *CSVRow) Type() ObjectType { return CSV_ROW }
func (c *CSVRow) Inspect() string {
	// Create a builder to efficiently build the string
	var builder strings.Builder

	// Build each row of data
	for key, value := range c.Row {
		builder.WriteString(key)
		builder.WriteString(": ")
		builder.WriteString(value)
		builder.WriteString("\n") // Adds a newline for each pair, you can change it to any separator you want
	}
	// builder.WriteString("\n")

	return builder.String()
}

type CSVVal struct {
	Value string
}

func (c *CSVVal) Type() ObjectType { return CSV_VAL }
func (c *CSVVal) Inspect() string {
	return c.Value
}
