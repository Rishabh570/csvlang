// object/object.go
package object

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/Rishabh570/csvlang/ast"
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
	ARRAY            = "ARRAY"

	BUILTIN_OBJ = "BUILTIN"
)

type Object interface {
	Type() ObjectType
	Inspect() string
	// Add method to attempt conversion to CSV
	ToCSV(env *Environment) (*CSV, error)
}

type BuiltinFunction func(env *Environment, args ...Object) Object

// Integer
type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) ToCSV(env *Environment) (*CSV, error) {
	// Get headers from environment if present
	var header string
	var columnType ColumnType
	if csvObj, ok := env.Get("csv"); ok {
		currentCSV := csvObj.(*CSV)
		if len(currentCSV.Headers) > 0 {
			header = currentCSV.Headers[0]
			columnType = currentCSV.ColumnTypes[0]
			// Validate type compatibility
			if columnType.DataType != INTEGER_OBJ && columnType.DataType != STRING_OBJ {
				return nil, fmt.Errorf("type mismatch: cannot convert INTEGER to %s", columnType.DataType)
			}
		}
	}

	if header == "" {
		header = "col1"
		columnType = ColumnType{DataType: INTEGER_OBJ}
	}

	return &CSV{
		Headers:     []string{header},
		ColumnTypes: []ColumnType{columnType},
		Rows:        []map[string]string{{header: i.Inspect()}},
	}, nil
}

// String
type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
func (s *String) ToCSV(env *Environment) (*CSV, error) {
	var header string
	var columnType ColumnType
	if csvObj, ok := env.Get("csv"); ok {
		currentCSV := csvObj.(*CSV)
		if len(currentCSV.Headers) > 0 {
			header = currentCSV.Headers[0]
			columnType = currentCSV.ColumnTypes[0]
		}
	}

	if header == "" {
		header = "col1"
		columnType = ColumnType{DataType: STRING_OBJ}
	}

	return &CSV{
		Headers:     []string{header},
		ColumnTypes: []ColumnType{columnType},
		Rows:        []map[string]string{{header: s.Value}},
	}, nil
}

// Bool
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) ToCSV(env *Environment) (*CSV, error) {
	var header string
	var columnType ColumnType
	if csvObj, ok := env.Get("csv"); ok {
		currentCSV := csvObj.(*CSV)
		if len(currentCSV.Headers) > 0 {
			header = currentCSV.Headers[0]
			columnType = currentCSV.ColumnTypes[0]
			if columnType.DataType != BOOLEAN_OBJ && columnType.DataType != STRING_OBJ {
				return nil, fmt.Errorf("type mismatch: cannot convert BOOLEAN to %s", columnType.DataType)
			}
		}
	}

	if header == "" {
		header = "col1"
		columnType = ColumnType{DataType: BOOLEAN_OBJ}
	}

	return &CSV{
		Headers:     []string{header},
		ColumnTypes: []ColumnType{columnType},
		Rows:        []map[string]string{{header: b.Inspect()}},
	}, nil
}

// Null object
type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }
func (n *Null) ToCSV(env *Environment) (*CSV, error) {
	return nil, fmt.Errorf("cannot convert null to CSV")
}

// Return
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }
func (rv *ReturnValue) ToCSV(env *Environment) (*CSV, error) {
	return nil, fmt.Errorf("cannot convert return value to CSV")
}

// Error object
type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
func (e *Error) ToCSV(env *Environment) (*CSV, error) {
	return nil, fmt.Errorf("cannot convert error to CSV: %s", e.Message)
}

// Built-in functionality to our lang which the host lang (Go) doesn't provide
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }
func (b *Builtin) ToCSV(env *Environment) (*CSV, error) {
	return nil, fmt.Errorf("cannot convert builtin function to CSV")
}

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
func (f *Function) ToCSV(env *Environment) (*CSV, error) {
	return nil, fmt.Errorf("cannot convert function to CSV")
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
func (csv *CSV) ToCSV(env *Environment) (*CSV, error) {
	return csv, nil // Already a CSV
}

// Array object
type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY }
func (a *Array) Inspect() string {
	var out bytes.Buffer
	elements := []string{}
	for _, e := range a.Elements {
		elements = append(elements, e.Inspect())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}
func (arr *Array) ToCSV(env *Environment) (*CSV, error) {
	return ArrayToCSV(arr, env)
}

// Array to CSV utils
func ArrayToCSV(arr *Array, env *Environment) (*CSV, error) {
	// Get current CSV headers if present in environment
	var headers []string
	var columnTypes []ColumnType

	if csvObj, ok := env.Get("csv"); ok {
		currentCSV := csvObj.(*CSV)
		headers = currentCSV.Headers
		columnTypes = currentCSV.ColumnTypes
	}

	// Handle empty array
	if len(arr.Elements) == 0 {
		if headers == nil {
			return &CSV{
				Headers:     []string{},
				ColumnTypes: []ColumnType{},
				Rows:        []map[string]string{},
			}, nil
		}
		return &CSV{
			Headers:     headers,
			ColumnTypes: columnTypes,
			Rows:        []map[string]string{},
		}, nil
	}

	// Determine if it's 1D or 2D array
	isOneD := true
	if firstElem, ok := arr.Elements[0].(*Array); ok {
		isOneD = false
		// Validate all elements are arrays of same length
		length := len(firstElem.Elements)
		for _, elem := range arr.Elements {
			if row, ok := elem.(*Array); !ok || len(row.Elements) != length {
				return nil, fmt.Errorf("inconsistent row lengths in 2D array")
			}
		}
	}

	if isOneD {
		return oneDArrayToCSV(arr, headers, columnTypes)
	}
	return twoDArrayToCSV(arr, headers, columnTypes)
}

func oneDArrayToCSV(arr *Array, existingHeaders []string, existingTypes []ColumnType) (*CSV, error) {
	var headers []string
	var columnTypes []ColumnType

	// Use existing headers if available, otherwise generate
	if existingHeaders != nil {
		// Validate array length matches header count
		if len(arr.Elements) != len(existingHeaders) {
			return nil, fmt.Errorf("array length %d does not match expected column count %d",
				len(arr.Elements), len(existingHeaders))
		}
		headers = existingHeaders
		columnTypes = existingTypes
	} else {
		// Generate headers (col1, col2, ...)
		headers = make([]string, len(arr.Elements))
		columnTypes = make([]ColumnType, len(arr.Elements))
		for i := range arr.Elements {
			headers[i] = fmt.Sprintf("col%d", i+1)
			columnTypes[i] = InferType(arr.Elements[i])
		}
	}

	// Create single row
	row := make(map[string]string)
	for i, elem := range arr.Elements {
		// Validate type if using existing column types
		if existingTypes != nil {
			if !isCompatibleType(elem, columnTypes[i]) {
				return nil, fmt.Errorf("type mismatch in column %s: expected %s, got %s",
					headers[i], columnTypes[i].DataType, elem.Type())
			}
		}
		row[headers[i]] = elem.Inspect()
	}

	return &CSV{
		Headers:     headers,
		ColumnTypes: columnTypes,
		Rows:        []map[string]string{row},
	}, nil
}

func twoDArrayToCSV(arr *Array, existingHeaders []string, existingTypes []ColumnType) (*CSV, error) {
	firstRow := arr.Elements[0].(*Array)

	var headers []string
	var columnTypes []ColumnType

	// Use existing headers if available, otherwise generate
	if existingHeaders != nil {
		// Validate row length matches header count
		if len(firstRow.Elements) != len(existingHeaders) {
			return nil, fmt.Errorf("row length %d does not match expected column count %d",
				len(firstRow.Elements), len(existingHeaders))
		}
		headers = existingHeaders
		columnTypes = existingTypes
	} else {
		// Generate headers from first row length
		headers = make([]string, len(firstRow.Elements))
		columnTypes = make([]ColumnType, len(firstRow.Elements))
		for i := range firstRow.Elements {
			headers[i] = fmt.Sprintf("col%d", i+1)
		}
	}

	// Create rows and validate/infer types
	rows := make([]map[string]string, len(arr.Elements))

	for i, rowObj := range arr.Elements {
		row := rowObj.(*Array)
		rowMap := make(map[string]string)

		for j, elem := range row.Elements {
			if existingTypes != nil {
				// Validate against existing types
				if !isCompatibleType(elem, columnTypes[j]) {
					return nil, fmt.Errorf("type mismatch in row %d, column %s: expected %s, got %s",
						i, headers[j], columnTypes[j].DataType, elem.Type())
				}
			} else if i == 0 {
				// Infer types from first row if no existing types
				columnTypes[j] = InferType(elem)
			} else {
				// Validate against inferred types
				if !isCompatibleType(elem, columnTypes[j]) {
					return nil, fmt.Errorf("type mismatch in row %d, column %s: expected %s, got %s",
						i, headers[j], columnTypes[j].DataType, elem.Type())
				}
			}
			rowMap[headers[j]] = elem.Inspect()
		}
		rows[i] = rowMap
	}

	return &CSV{
		Headers:     headers,
		ColumnTypes: columnTypes,
		Rows:        rows,
	}, nil
}

func InferType(obj Object) ColumnType {
	switch obj.(type) {
	case *Integer:
		return ColumnType{DataType: INTEGER_OBJ}
	case *String:
		return ColumnType{DataType: STRING_OBJ}
	case *Boolean:
		return ColumnType{DataType: BOOLEAN_OBJ}
	default:
		return ColumnType{DataType: STRING_OBJ} // default to string
	}
}

func isCompatibleType(value Object, colType ColumnType) bool {
	switch colType.DataType {
	case INTEGER_OBJ:
		_, ok := value.(*Integer)
		return ok
	case STRING_OBJ:
		_, ok := value.(*String)
		return ok
	case BOOLEAN_OBJ:
		_, ok := value.(*Boolean)
		return ok
	default:
		return true // string can hold anything
	}
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}
