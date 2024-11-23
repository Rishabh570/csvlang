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

func evalLoadStatement(ls *ast.LoadStatement, env *object.Environment) object.Object {
	fmt.Println("ls.filename: ", ls.Filename)
	fmt.Println("ls: ", ls.Token.Type, ls.Token.Literal)

	// filename := Eval(ls.Filename, env)
	// filename := ls.Filename
	// if isError(filename) {
	// 	return filename
	// }
	// fmt.Printf("filename: %v\n", filename)

	// Get the string value of the filename
	// filenameStr := ""
	// switch fn := ls.Filename.String().(type) {
	// case *object.String:
	// 	filenameStr = fn.Value
	// default:
	// 	return newError("filename must be a string, got %s", filename.Type())
	// }

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
	fmt.Printf("[evalLoadStatement] returning csvObj: %+v\n", csvObj)
	env.Set("csv", csvObj) // Store with key "csv"
	return csvObj
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

	if rs.Location.RowIndex == -1 {
		// TODO: return 2d array
		// return
	}

	val := csvObj.Rows[rs.Location.RowIndex]
	if val == nil {
		// fmt.Println("val is nil")
		return &object.Error{
			Message: fmt.Sprintf("no rows available with row index: %d", rs.Location.RowIndex),
		}
	}
	fmt.Printf("row val: %+v\n", val)

	if rs.Location.ColIndex == "" {
		// TODO: return 1d array
		// return
	}

	// fmt.Println("returning filtered row from evalRead")
	if rs.Location.ColIndex != "" {
		fmt.Printf("rs.Location.ColIndex: %s\n", rs.Location.ColIndex)
		colVal := val[rs.Location.ColIndex]
		fmt.Printf("colVal: %s\n", colVal)
		if colVal == "" {
			return &object.Error{
				Message: fmt.Sprintf("no available values with row index: %d and column name: %s\n", rs.Location.RowIndex, rs.Location.ColIndex),
			}
		}

		// Try to determine the type and convert accordingly
		if intValue, err := strconv.ParseInt(colVal, 10, 64); err == nil {
			return &object.Integer{Value: intValue}
		}

		// If not an integer, return as string
		return &object.String{Value: colVal}
	}

	return nil
	// return newError("no CSV file has been loaded")
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
		fmt.Printf("[evalIdentifier] returning val: %s\n", val.Inspect())
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
