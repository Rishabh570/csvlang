// evaluator package is responsible for evaluating the AST nodes and returning the result of the evaluation.
//
// It uses the object package to create and return the evaluated objects.
// The evaluator package is the heart of the interpreter, where the actual evaluation of the AST nodes happens.
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

// Eval function is the entry point to the evaluator package.
// It takes an AST node and an environment object as input and returns the evaluated object.
// The environment object is used to store and retrieve variables and their values.
// The Eval function is a recursive function that evaluates the AST nodes and returns the evaluated object.
func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	// ================ Statements ================
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.LoadStatement:
		return evalLoadStatement(node, env)
	case *ast.ReadStatement:
		return evalReadStatement(node.ReadExpression, env)
	case *ast.ReadExpression:
		return evalReadStatement(node, env)
	case *ast.SaveStatement:
		return evalSaveStatement(node, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
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
	case *ast.ForLoopStatement:
		return evalForLoopStatement(node.ForLoopExpression, env)

	// ================ Expressions ================
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
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
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
	case *ast.ForLoopExpression:
		return evalForLoopStatement(node, env)
	case *ast.IndexAssignmentExpression:
		return evalIndexAssignmentExpression(node, env)
	default:
		fmt.Printf("[Eval] no match: %+v\n", node.String())
	}
	return nil
}

// evalIndexAssignmentExpression evaluates an index assignment expression.
// Example: `array[index] = value`.
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

// evalSaveStatement evaluates a save statement.
// It saves the CSV data to a file in the specified format (CSV or JSON).
// Example: `save csv as "output.csv"` or `save json as "output.json"`.
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

// saveAsCSV saves the CSV data to a file in CSV format.
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

// saveAsJSON saves the CSV data to a file in JSON format.
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

// evalForLoopStatement evaluates a for loop statement.
// Example: `for i in array { ... }`.
// It iterates over the elements of the array and executes the body of the loop for each element.
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

// evalIndexExpression evaluates an index expression by calling evalArrayIndexExpression.
// Example: `array[index]`.
// It retrieves the element at the specified index from the array.
func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

// evalArrayIndexExpression evaluates an array index expression.
// It retrieves the element at the specified index from the array.
// Example: `array[index]`.
func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)
	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}

// evalLoadStatement evaluates a load statement.
// It loads a CSV file and stores its data in the environment.
// Example: `load "data.csv"`.
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

// selectRows selects rows based on the rowIndex.
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

// evaluateNumericCondition evaluates a numeric condition based on the operator and value.
// It compares the column value with the compare value using the specified operator.
// Example: `column > 5`, `column < 10`, etc.
// It returns true if the condition is satisfied, otherwise false.
// The column value is expected to be a string that can be converted to an integer.
// The compare value is an integer.
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

// evaluateStringCondition evaluates a string condition based on the operator and value.
// Example: `column == "value"`, `column != "value"`, etc.
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

// evaluateBooleanCondition evaluates a boolean condition based on the operator and value.
// Example: `column == true`, `column != false`, etc.
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

// evaluateCondition evaluates a condition based on the column value, operator, and compare value.
// It checks if the column value satisfies the condition specified in the where clause.
// Example: `column > 5`, `column == "value"`, etc.
// It returns true if the condition is satisfied, otherwise false.
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

// filterRows filters the rows based on the where clause.
// It checks if each row satisfies the condition specified in the where clause.
func filterRows(rows []map[string]string, where *ast.ReadFilterExpression, env *object.Environment) []map[string]string {
	var filtered []map[string]string

	for _, row := range rows {
		if evaluateCondition(row, where, env) {
			filtered = append(filtered, row)
		}
	}

	return filtered
}

// extractColumns extracts the specified columns from the rows.
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

// evalReadStatement evaluates a read statement.
// It retrieves the CSV data from the environment and filters it based on the specified conditions.
func evalReadStatement(rs *ast.ReadExpression, env *object.Environment) object.Object {
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

// evalBlockStatement evaluates a block statement.
// It executes each statement in the block and returns the result of the last statement.
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

// evalExpressions evaluates a list of expressions and returns the evaluated objects.
// It returns an array of evaluated objects.
func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
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

// applyFunction applies a function to the given arguments.
// It evaluates the function and its arguments, and returns the result of the function call.
// It handles both user-defined functions and built-in functions.
// Example: `fn(arg1, arg2)` or `builtin(arg1, arg2)`.
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

// extendFunctionEnv extends the function environment with the given arguments.
// It creates a new environment for the function call and sets the parameters to the corresponding arguments.
// This allows the function to access its arguments using the parameter names.
// Example: `fn(param1, param2)`.
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)
	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}
	return env
}

// evalProgram evaluates a program by executing each statement in the program.
// It returns the result of the last statement in the program.
// Example: `let x = 5; let y = 10; x + y;`.
// The result of the last statement (x + y) is returned.
func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range stmts {
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

// nativeBoolToBooleanObject converts a native boolean to a Boolean object.
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

// evalPrefixExpression evaluates a prefix expression.
// It applies the operator to the right operand and returns the result.
// Example: `!true`, `-5`, etc.
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

// unwrapReturnValue unwraps the return value from a function call.
func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

// evalIdentifier evaluates an identifier.
// It retrieves the value of the identifier from the environment.
// Example: `x`, `y`, etc.
func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
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

// evalIfExpression evaluates an if expression.
// It checks the condition and executes the consequence or alternative block based on the condition's truthiness.
// Example: `if (condition) { ... } else { ... }`.
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

// isTruthy checks if an object is truthy.
// It returns true if the object is not NULL, TRUE, or FALSE.
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

// evalInfixExpression evaluates an infix expression.
// It applies the operator to the left and right operands and returns the result.
// Example: `5 + 10`, `x > y`, etc.
func evalInfixExpression(operator string, left, right object.Object) object.Object {
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

// evalStringInfixExpression evaluates a string infix expression.
// It applies the operator to the left and right string operands and returns the result.
// Example: `"hello" + "world"`.
// It only supports the "+" operator for string concatenation.
func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

// evalIntegerInfixExpression evaluates an integer infix expression.
// It applies the operator to the left and right integer operands and returns the result.
// Example: `5 + 10`, `x > y`, etc.
func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
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

// evalMinusPrefixOperatorExpression evaluates a prefix minus operator.
// It negates the value of the right operand.
// Example: `-5`, `-x`, etc.
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

// evalBangOperatorExpression evaluates a prefix bang operator.
// It negates the truthiness of the right operand.
// Example: `!true`, `!false`, etc.
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

// newError creates a new error object with the specified format and arguments.
// It formats the error message using fmt.Sprintf and returns an Error object.
// Example: `newError("error: %s", "something went wrong")`.
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// isError checks if an object is an error.
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
