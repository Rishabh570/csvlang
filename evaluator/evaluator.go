package evaluator

import (
	"csvlang/ast"
	"csvlang/object"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.ExpressionStatement:
		fmt.Printf("[Eval] expr stmt run for %s\n", node.String())
		return Eval(node.Expression, env)
	case *ast.LoadStatement:
		return evalLoadStatement(node, env)
	case *ast.ReadStatement:
		fmt.Printf("read stmt...")
		return evalReadStatement(node.ReadExpression, env)
	case *ast.ReadExpression:
		fmt.Printf("read expr...")
		return evalReadStatement(node, env)
	case *ast.AppendStatement:
		return evalAppendStatement(node, env)
	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		fmt.Printf("[Eval] Evaluating LET stmt, nodeVal: %v\n", node.Value)
		val := Eval(node.Value, env)
		fmt.Printf("[Eval] LET val: %v\n", val)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		fmt.Printf("[Eval] identifier run for %s %s\n", node.Value, node.String())
		return evalIdentifier(node, env)
	case *ast.AssignmentStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		// Check if variable exists
		if _, ok := env.Get(node.Name.Value); !ok {
			return newError("identifier not found: " + node.Name.Value)
		}
		env.Set(node.Name.Value, val)
		return val
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		fmt.Printf("[Eval] array literal: %v\n", elements)
		return &object.Array{Elements: elements}

	case *ast.IndexExpression:
		fmt.Printf("[Eval] index expr: %+v\n", node)
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		fmt.Printf("[Eval] index expr: left: %v, index: %v\n", left, index)
		return evalIndexExpression(left, index)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args)
	default:
		fmt.Printf("[Eval] no match: %+v\n", node.String())
	}
	return nil
}

func evalAppendStatement(node *ast.AppendStatement, env *object.Environment) object.Object {
	csvObj, ok := env.Get("csv")
	if !ok {
		return newError("no CSV file loaded")
	}

	csvData := csvObj.(*object.CSV)

	// Validate number of values matches headers
	if len(node.Values) != len(csvData.Headers) {
		return newError("wrong number of values: expected %d, got %d", len(csvData.Headers), len(node.Values))
	}

	// If column types not inferred yet, do it now
	// it should already be inferred when the CSV is loaded though
	if len(csvData.ColumnTypes) == 0 {
		csvData.InferColumnTypes()
	}

	// ========================
	// Validate types of new values and append to rows
	// ========================
	newRow := make(map[string]string)
	for i, value := range node.Values {
		evaluated := Eval(value, env)

		if evaluated.Type() != csvData.ColumnTypes[i].DataType {
			return newError("type mismatch for column %s: expected %s, got %s",
				csvData.Headers[i], csvData.ColumnTypes[i].DataType, evaluated.Type())
		}

		newRow[csvData.Headers[i]] = evaluated.Inspect()
	}

	// Append the new row
	csvData.Rows = append(csvData.Rows, newRow)

	// Store the updated CSV object in the environment
	env.Set("csv", csvObj)

	// ========================
	// After successful validation and row creation, write to CSV file
	// ========================

	filename, ok := env.Get("filename")
	if !ok {
		return newError("no filename found in environment")
	}
	filenameStr := filename.(*object.String).Value

	file, err := os.OpenFile(filenameStr, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return newError("could not open file for writing: %s", err)
	}
	defer file.Close()

	// First ensure we're on a new line by checking last byte
	fileInfo, err := file.Stat()
	if err != nil {
		return newError("could not get file info: %s", err)
	}

	// If file is not empty and doesn't end with newline, add one
	if fileInfo.Size() > 0 {
		lastByte := make([]byte, 1)
		if _, err := file.ReadAt(lastByte, fileInfo.Size()-1); err == nil {
			if lastByte[0] != '\n' {
				if _, err := file.WriteString("\n"); err != nil {
					return newError("could not write newline: %s", err)
				}
			}
		}
	}

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Convert map to slice in correct header order
	rowData := make([]string, len(csvData.Headers))
	for i, header := range csvData.Headers {
		rowData[i] = newRow[header]
	}

	// Write the new row
	if err := writer.Write(rowData); err != nil {
		return newError("could not write to file: %s", err)
	}

	return NULL
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)
	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}

func evalLoadStatement(ls *ast.LoadStatement, env *object.Environment) object.Object {
	// Store the filename in the environment
	env.Set("filename", &object.String{ls.Filename.String()})

	// Open and read the CSV file
	file, err := os.Open(ls.Filename.String())
	if err != nil {
		return newError("could not open file: %s", err)
	}
	defer file.Close()

	// Parse CSV
	reader := csv.NewReader(file)

	// Read headers
	headers, err := reader.Read()
	if err != nil {
		return newError("could not read CSV headers: %s", err)
	}

	// TODO: @dev ddos magnet 👀
	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return newError("could not read CSV records: %s", err)
	}

	// Convert records to rows of maps
	rows := make([]map[string]string, len(records))
	for i, record := range records {
		row := make(map[string]string)
		for j, header := range headers {
			row[header] = record[j]
		}
		rows[i] = row
	}

	// Store loaded CSV data in evaluator's in-app memory
	// TODO: is it a good idea that a language is storing computed results which can blow up in case of huge CSV data?
	csvObj := &object.CSV{
		Headers: headers,
		Rows:    rows,
	}

	// When the CSV is loaded successfully for the first time, infer column types and store the information for future use
	csvObj.InferColumnTypes()

	// Store the CSV object in the environment
	env.Set("csv", csvObj)
	return csvObj
}

func selectRows(rows []map[string]string, rowIndex int) []map[string]string {
	// rowIndex -2 means select all rows
	if rowIndex == -2 {
		return rows
	}

	// do not return anything if invalid rowIndex
	if rowIndex >= len(rows) || rowIndex < 0 {
		return nil
	}

	return []map[string]string{rows[rowIndex]}
}

func evaluateNumericCondition(columnValue string, operator string, compareValue int64) bool {
	// Convert column value to number
	rowVal, err := strconv.ParseInt(columnValue, 10, 64)
	if err != nil {
		return false
	}

	switch operator {
	case ">":
		return rowVal > compareValue
	case "<":
		return rowVal < compareValue
	case ">=":
		return rowVal >= compareValue
	case "<=":
		return rowVal <= compareValue
	case "==":
		return rowVal == compareValue
	case "!=":
		return rowVal != compareValue
	default:
		return false
	}
}

func evaluateStringCondition(columnValue string, operator string, compareValue string) bool {
	switch operator {
	case "==":
		return columnValue == compareValue
	case "!=":
		return columnValue != compareValue
	case ">":
		return columnValue > compareValue
	case "<":
		return columnValue < compareValue
	case ">=":
		return columnValue >= compareValue
	case "<=":
		return columnValue <= compareValue
	default:
		return false
	}
}

func evaluateBooleanCondition(columnValue string, operator string, compareValue bool) bool {
	rowVal, err := strconv.ParseBool(columnValue)
	if err != nil {
		return false
	}

	switch operator {
	case "==":
		return rowVal == compareValue
	case "!=":
		return rowVal != compareValue
	default:
		return false
	}
}

func evaluateCondition(row map[string]string, where *ast.ReadFilterExpression, env *object.Environment) bool {
	columnValue := row[where.ColumnName]

	// First evaluate the condition's value
	compareValue := Eval(where.Value, env)
	if isError(compareValue) {
		return false
	}

	switch compareValue.Type() {
	case object.INTEGER_OBJ:
		return evaluateNumericCondition(columnValue, where.Operator, compareValue.(*object.Integer).Value)

	case object.STRING_OBJ:
		return evaluateStringCondition(columnValue, where.Operator, compareValue.(*object.String).Value)

	case object.BOOLEAN_OBJ:
		return evaluateBooleanCondition(columnValue, where.Operator, compareValue.(*object.Boolean).Value)
	default:
		return false
	}
}

func filterRows(rows []map[string]string, where *ast.ReadFilterExpression, env *object.Environment) []map[string]string {
	var filtered []map[string]string

	for _, row := range rows {
		if evaluateCondition(row, where, env) {
			filtered = append(filtered, row)
		}
	}

	return filtered
}

func extractColumns(rows []map[string]string, column string) *object.Array {
	var values object.Array

	for _, row := range rows {
		if val, ok := row[column]; ok {
			if intValue, err := strconv.ParseInt(val, 10, 64); err == nil {
				values.Elements = append(values.Elements, &object.Integer{Value: intValue})
			} else {
				values.Elements = append(values.Elements, &object.String{Value: val})
			}
		}
	}

	return &values
}

func evalReadStatement(rs *ast.ReadExpression, env *object.Environment) object.Object {
	fmt.Printf("[evalReadStatement] type: %s, lit: %s, row: %d, col: %s\n", rs.Token.Type, rs.Token.Literal, rs.Location.RowIndex, rs.Location.ColIndex)

	// Retrieve stored CSV object
	csv, ok := env.Get("csv")
	if !ok {
		return nil
	}

	csvObj, ok := csv.(*object.CSV)
	if !ok {
		fmt.Println("csv object does not follow the intended structure")
		return nil
	}

	rows := selectRows(csvObj.Rows, rs.Location.RowIndex)

	if rs.Location.Filter != nil {
		rows = filterRows(rows, rs.Location.Filter, env)
	}

	if rs.Location.ColIndex != "" {
		return extractColumns(rows, rs.Location.ColIndex)
	}

	return &object.CSV{Rows: rows, Headers: csvObj.Headers, ColumnTypes: csvObj.ColumnTypes}
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

func evalExpressions(
	exps []ast.Expression,
	env *object.Environment,
) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)
	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}
	return env
}

func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range stmts {
		fmt.Println("evaluating stmt: ", statement.String())
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

// func applyFunction(fn object.Object, args []object.Object) object.Object {
// 	function, ok := fn.(*object.Function)
// 	if !ok {
// 		return newError("not a function: %s", fn.Type())
// 	}
// 	extendedEnv := extendFunctionEnv(function, args)
// 	evaluated := Eval(function.Body, extendedEnv)
// 	return unwrapReturnValue(evaluated)
// }

// func extendFunctionEnv(
// 	fn *object.Function,
// 	args []object.Object,
// ) *object.Environment {
// 	env := object.NewEnclosedEnvironment(fn.Env)
// 	for paramIdx, param := range fn.Parameters {
// 		env.Set(param.Value, args[paramIdx])
// 	}
// 	return env
// }

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func evalIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	fmt.Printf("[evalIdentifier] starting, node.Value: %s, node.String(): %s\n", node.Value, node.String())
	if val, ok := env.Get(node.Value); ok {
		fmt.Printf("[evalIdentifier] returning val: %s\n", val)
		return val
	}

	// Check in built-ins if the identifier is not present in the env object
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func evalInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

func evalIntegerInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value
	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

/*
Error
*/
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
