package evaluator

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Rishabh570/csvlang/object"
)

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s",
					args[0].Type())
			}
		},
	},
	"first": &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY {
				return newError("argument to `first` must be ARRAY, got %s",
					args[0].Type())
			}
			arr := args[0].(*object.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}
			return NULL
		},
	},
	"last": &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {

				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY {
				return newError("argument to `last` must be ARRAY, got %s",
					args[0].Type())
			}
			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}
			return NULL
		},
	},
	"rest": &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY {
				return newError("argument to `rest` must be ARRAY, got %s",
					args[0].Type())
			}
			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]object.Object, length-1, length-1)
				copy(newElements, arr.Elements[1:length])

				return &object.Array{Elements: newElements}
			}
			return NULL
		},
	},
	"push": &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments")
			}

			// If first argument is CSV, convert second arg to CSV
			if args[0].Type() == object.CSV_OBJ {
				csv := args[0].(*object.CSV)
				toAdd, err := args[1].ToCSV(env)
				if err != nil {
					return newError(err.Error())
				}
				return mergeCSVs(csv, toAdd)
			}

			// If second argument is CSV, convert first arg to CSV
			if args[1].Type() == object.CSV_OBJ {
				arr, ok := args[0].(*object.Array)
				if !ok {
					return newError("first argument must be ARRAY or CSV when pushing CSV")
				}
				csv, err := arr.ToCSV(env)
				if err != nil {
					return newError(err.Error())
				}
				return mergeCSVs(csv, args[1].(*object.CSV))
			}

			// If neither is CSV, use regular array push
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("first argument must be ARRAY")
			}

			// If array is empty, treat as 1D array
			if len(arr.Elements) == 0 {
				newElements := make([]object.Object, 1)
				newElements[0] = args[1]
				return &object.Array{Elements: newElements}
			}

			// Check if it's a 2D array by looking at first element
			if _, ok := arr.Elements[0].(*object.Array); ok {
				// It's a 2D array
				// If pushing an array, add it directly
				if pushArr, ok := args[1].(*object.Array); ok {
					// Validate row length matches existing rows
					if len(pushArr.Elements) != len(arr.Elements[0].(*object.Array).Elements) {
						return newError("cannot push array of length %d to 2D array with row length %d",
							len(pushArr.Elements), len(arr.Elements[0].(*object.Array).Elements))
					}
					newElements := make([]object.Object, len(arr.Elements)+1)
					copy(newElements, arr.Elements)
					newElements[len(arr.Elements)] = pushArr
					return &object.Array{Elements: newElements}
				}
				// If pushing a single value, return error
				return newError("cannot push non-array value to 2D array")
			}

			// It's a 1D array
			newElements := make([]object.Object, len(arr.Elements)+1)
			copy(newElements, arr.Elements)
			newElements[len(arr.Elements)] = args[1]
			return &object.Array{Elements: newElements}
		},
	},
	"pop": &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments")
			}

			// Handle CSV pop
			if args[0].Type() == object.CSV_OBJ {
				csv := args[0].(*object.CSV)
				if len(csv.Rows) == 0 {
					return newError("cannot pop from empty CSV")
				}
				newRows := csv.Rows[:len(csv.Rows)-1]
				return &object.CSV{
					Headers:     csv.Headers,
					ColumnTypes: csv.ColumnTypes,
					Rows:        newRows,
				}
			}

			// Handle array pop
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument must be ARRAY or CSV")
			}
			if len(arr.Elements) == 0 {
				return newError("cannot pop from empty array")
			}
			length := len(arr.Elements)
			newElements := make([]object.Object, length-1)
			copy(newElements, arr.Elements[:length-1])
			return &object.Array{Elements: newElements}
		},
	},
	"unique": &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments: got=%d, want=1", len(args))
			}

			if args[0].Type() == object.CSV_OBJ {
				// Check if argument is CSV
				csv, ok := args[0].(*object.CSV)
				if !ok {
					return newError("argument must be CSV, got=%s", args[0].Type())
				}

				fmt.Printf("removing duplicates from the provided CSV object: %v\n", csv)

				return removeDuplicates(csv)
			}

			if args[0].Type() == object.ARRAY {
				// Check if argument is CSV
				csv, ok := args[0].(*object.Array)
				if !ok {
					return newError("argument must be CSV, got=%s", args[0].Type())
				}

				fmt.Printf("removing duplicates from the provided array: %v\n", csv)

				return removeDuplicatesFrom2dArray(csv, env)
			}

			return nil
		},
	},
	"sum": &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments: got=%d, want=1", len(args))
			}

			// Check if argument is array
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument must be ARRAY, got %s", args[0].Type())
			}

			// Calculate sum
			sum := int64(0)
			for _, elem := range arr.Elements {
				// Ensure each element is integer
				integer, ok := elem.(*object.Integer)
				if !ok {
					return newError("array elements must be INTEGER, got %s", elem.Type())
				}
				sum += integer.Value
			}

			return &object.Integer{Value: sum}
		},
	},
	"avg": &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments: got=%d, want=1", len(args))
			}

			// Check if argument is array
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument must be ARRAY, got %s", args[0].Type())
			}

			// Handle empty array
			if len(arr.Elements) == 0 {
				return newError("cannot calculate average of empty array")
			}

			// Calculate sum and validate elements
			sum := int64(0)
			for _, elem := range arr.Elements {
				// Handle both integer and float inputs
				switch num := elem.(type) {
				case *object.Integer:
					sum += int64(num.Value)
				default:
					return newError("array elements must be numeric, got %s", elem.Type())
				}
			}

			// Calculate average
			avg := sum / int64(len(arr.Elements))
			return &object.Integer{Value: avg}
		},
	},
	"count": &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments: got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.Array:
				// If empty array, return 0
				if len(arg.Elements) == 0 {
					return &object.Integer{Value: 0}
				}

				// If first element is array, it's 2D - return number of rows
				if _, ok := arg.Elements[0].(*object.Array); ok {
					return &object.Integer{Value: int64(len(arg.Elements))}
				}

				// For 1D array, return length
				return &object.Integer{Value: int64(len(arg.Elements))}

			case *object.CSV:
				// Return number of rows in CSV
				return &object.Integer{Value: int64(len(arg.Rows))}

			default:
				return newError("argument must be ARRAY or CSV, got %s", args[0].Type())
			}
		},
	},
	"fill_empty": &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments: got=%d, want=3", len(args))
			}

			csv, ok := args[0].(*object.CSV)
			if !ok {
				return newError("argument must be CSV, got %s", args[0].Type())
			}
			newRows := make([]map[string]string, len(csv.Rows))

			fieldValue, err := convertToString(args[2].Inspect())
			if err != nil {
				return newError(err.Error())
			}

			fieldName := args[1].Inspect()
			for i, row := range csv.Rows {
				newRow := make(map[string]string)
				for _, header := range csv.Headers {
					if header == fieldName && row[header] == "" {
						newRow[header] = fieldValue
					} else {
						newRow[header] = row[header]
					}
				}
				newRows[i] = newRow
			}

			modifiedCSV := &object.CSV{
				Headers:     csv.Headers,
				ColumnTypes: csv.ColumnTypes,
				Rows:        newRows,
			}
			// save to env
			env.Set("csv", modifiedCSV)
			return modifiedCSV
		},
	},
}

// object.CSV is our primary data type; it's best to implicitly convert the data type
func removeDuplicatesFrom2dArray(arr *object.Array, env *object.Environment) *object.CSV {
	// Handle empty array
	if len(arr.Elements) == 0 {
		return &object.CSV{
			Headers:     []string{},
			ColumnTypes: []object.ColumnType{},
			Rows:        []map[string]string{},
		}
	}

	// Validate it's a 2D array and all rows have same length
	firstRow, ok := arr.Elements[0].(*object.Array)
	if !ok {
		return nil
	}
	rowLength := len(firstRow.Elements)

	// Get headers from environment if present, otherwise generate
	var headers []string
	var columnTypes []object.ColumnType

	if csvObj, ok := env.Get("csv"); ok {
		currentCSV := csvObj.(*object.CSV)
		headers = currentCSV.Headers
		columnTypes = currentCSV.ColumnTypes
		// Validate row length matches headers
		if len(headers) != rowLength {
			return nil
		}
	} else {
		// Generate headers
		headers = make([]string, rowLength)
		columnTypes = make([]object.ColumnType, rowLength)
		for i := 0; i < rowLength; i++ {
			headers[i] = fmt.Sprintf("col%d", i+1)
			// Infer type from first row
			columnTypes[i] = object.InferType(firstRow.Elements[i])
		}
	}

	seen := make(map[string]bool)
	uniqueRows := []map[string]string{}

	// Create key slice for deduplication
	key := make([]string, len(arr.Elements))
	for _, oneDArr := range arr.Elements {
		row, ok := oneDArr.(*object.Array)
		if !ok || len(row.Elements) != rowLength {
			return nil
		}

		for i, ele := range row.Elements {
			key[i] = ele.Inspect()
		}
		rowKey := strings.Join(key, "|")
		fmt.Printf("rowKey:: %s\n", rowKey)

		if !seen[rowKey] {
			fmt.Printf("setting rowKey: %s to true\n", rowKey)
			seen[rowKey] = true

			// Create row map for CSV
			rowMap := make(map[string]string)
			for i, elem := range row.Elements {
				rowMap[headers[i]] = elem.Inspect()
			}
			uniqueRows = append(uniqueRows, rowMap)

		}
	}

	fmt.Printf("uniqueRows: %+v\n", uniqueRows)

	// Return new CSV object with unique rows
	return &object.CSV{
		Headers:     headers,
		ColumnTypes: columnTypes,
		Rows:        uniqueRows,
	}
}

func removeDuplicates(csv *object.CSV) *object.CSV {
	seen := make(map[string]bool)
	uniqueRows := []map[string]string{}

	for _, row := range csv.Rows {
		// Create a unique key for the row
		key := make([]string, len(csv.Headers))
		for i, header := range csv.Headers {
			key[i] = row[header]
		}
		rowKey := strings.Join(key, "|")

		if !seen[rowKey] {
			seen[rowKey] = true
			uniqueRows = append(uniqueRows, row)
		}
	}

	// Return new CSV object with unique rows
	return &object.CSV{
		Headers:     csv.Headers,
		ColumnTypes: csv.ColumnTypes,
		Rows:        uniqueRows,
	}
}

func mergeCSVs(target, source *object.CSV) object.Object {
	// Validate column compatibility
	if len(source.Headers) != len(target.Headers) {
		return newError("column count mismatch: expected %d, got %d",
			len(target.Headers), len(source.Headers))
	}

	// Validate column types
	for i, targetType := range target.ColumnTypes {
		if !isCompatibleColumnType(targetType, source.ColumnTypes[i]) {
			return newError("incompatible column types for column %s", target.Headers[i])
		}
	}

	// Merge rows
	newRows := make([]map[string]string, len(target.Rows)+len(source.Rows))
	copy(newRows, target.Rows)
	copy(newRows[len(target.Rows):], source.Rows)

	return &object.CSV{
		Headers:     target.Headers,
		ColumnTypes: target.ColumnTypes,
		Rows:        newRows,
	}
}

func isCompatibleColumnType(target, source object.ColumnType) bool {
	// If target is string, accept any type
	if target.DataType == object.STRING_OBJ {
		return true
	}
	return target.DataType == source.DataType
}

func convertToString(value any) (string, error) {
	switch v := value.(type) {
	case int:
		return strconv.Itoa(v), nil
	case string:
		return v, nil
	default:
		return "", errors.New("unsupported type: only integers are supported")
	}
}
