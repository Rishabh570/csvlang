package evaluator

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Rishabh570/csvlang/ast"
	"github.com/Rishabh570/csvlang/object"
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
	case *ast.SaveStatement:
		return evalSaveStatement(node, env)
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
		return applyFunction(function, args, env)
	case *ast.ForLoopStatement:
		return evalForLoopStatement(node.ForLoopExpression, env)
	case *ast.ForLoopExpression:
		return evalForLoopStatement(node, env)
	case *ast.IndexAssignmentExpression:
		return evalIndexAssignmentExpression(node, env)
	default:
		fmt.Printf("[Eval] no match: %+v\n", node.String())
	}
	return nil
}

func evalIndexAssignmentExpression(node *ast.IndexAssignmentExpression, env *object.Environment) object.Object {
	// Evaluate the array
	array := Eval(node.Left.Left, env)
	if isError(array) {
		return array
	}

	// Evaluate the index
	index := Eval(node.Left.Index, env)
	if isError(index) {
		return index
	}

	// Evaluate the value to assign
	value := Eval(node.Value, env)
	if isError(value) {
		return value
	}

	// Check if we're working with an array
	arr, ok := array.(*object.Array)
	if !ok {
		return newError("index assignment not supported for type: %s", array.Type())
	}

	// Check if index is integer
	idx, ok := index.(*object.Integer)
	if !ok {
		return newError("array index must be INTEGER, got %s", index.Type())
	}

	// Check bounds
	if idx.Value < 0 || idx.Value >= int64(len(arr.Elements)) {
		return newError("array index out of bounds: %d", idx.Value)
	}

	// Update array element
	arr.Elements[idx.Value] = value
	return value
}

func evalSaveStatement(node *ast.SaveStatement, env *object.Environment) object.Object {
	var dataToSave *object.CSV

	if node.Source != nil {
		// Get custom source data
		value := Eval(node.Source, env)
		if isError(value) {
			return value
		}

		// Validate it's a CSV object
		csv, ok := value.(*object.CSV)
		if !ok {
			return newError("cannot save non-CSV data")
		}
		dataToSave = csv
	} else {
		// Get latest CSV from environment
		value, ok := env.Get("csv")
		if !ok {
			return newError("no CSV data to save")
		}
		dataToSave = value.(*object.CSV)
	}

	// Save based on format
	switch node.Format {
	case "csv":
		return saveAsCSV(dataToSave, node.Filename)
	case "json":
		return saveAsJSON(dataToSave, node.Filename)
	default:
		return newError("unsupported format: %s", node.Format)
	}
}

func saveAsCSV(csvData *object.CSV, filename string) object.Object {
	file, err := os.Create(filename)
	if err != nil {
		return newError("could not create file: %s", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	if err := writer.Write(csvData.Headers); err != nil {
		return newError("error writing headers: %s", err)
	}

	// Write rows
	for _, row := range csvData.Rows {
		record := make([]string, len(csvData.Headers))
		for i, header := range csvData.Headers {
			record[i] = row[header]
		}
		if err := writer.Write(record); err != nil {
			return newError("error writing row: %s", err)
		}
	}

	return NULL
}

func saveAsJSON(csv *object.CSV, filename string) object.Object {
	data := map[string]interface{}{
		"headers": csv.Headers,
		"rows":    csv.Rows,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return newError("error converting to JSON: %s", err)
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return newError("error writing file: %s", err)
	}

	return NULL
}

func evalForLoopStatement(fl *ast.ForLoopExpression, env *object.Environment) object.Object {
	iterableObj := Eval(fl.Iterable, env)
	if isError(iterableObj) {
		return iterableObj
	}

	var elements []object.Object
	switch iterable := iterableObj.(type) {
	case *object.Array:
		elements = iterable.Elements
	default:
		return newError("for loop iterable must be ARRAY, got %s", iterableObj.Type())
	}

	for i, element := range elements {
		// Create new scope for each iteration
		loopEnv := object.NewEnclosedEnvironment(env)

		// Bind index and element
		loopEnv.Set(fl.IndexName.Value, &object.Integer{Value: int64(i)})
		loopEnv.Set(fl.ElementName.Value, element)

		fmt.Printf("[evalForLoopStatement] evaluating block statment: %+v\n", fl.Body)
		// Evaluate body
		Eval(fl.Body, loopEnv)

		// Update original array if modified
		if arr, ok := iterableObj.(*object.Array); ok {
			if val, ok := loopEnv.Get(fl.ElementName.Value); ok {
				fmt.Printf("[evalForLoopStatement] updating element: %+v\n", val.Inspect())
				arr.Elements[i] = val
			}
		}

		// Copy any modified outer scope variables back to parent environment
		for name, val := range loopEnv.GetStore() {
			if _, exists := env.Get(name); exists {
				fmt.Printf("[evalForLoopStatement] setting %s to %s\n", name, val.Inspect())
				env.Set(name, val)
			}
		}

		// unset values from the loopEnv set in this iteration of the for loop
		// loopEnv.Unset(fl.IndexName.Value)
		// loopEnv.Unset(fl.ElementName.Value)
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

func applyFunction(fn object.Object, args []object.Object, env *object.Environment) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return fn.Fn(env, args...)
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
