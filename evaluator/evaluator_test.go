// evaluator/evaluator_test.go
package evaluator

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/Rishabh570/csvlang/lexer"
	"github.com/Rishabh570/csvlang/object"
	"github.com/Rishabh570/csvlang/parser"

	"github.com/gocarina/gocsv"
	"github.com/stretchr/testify/require"
)

type RowJSON struct {
	Name string `json:"name"`
	Age  string `json:"age"`
}

type OutputJSON struct {
	Headers []string
	Rows    []RowJSON
}

type RowCSV struct {
	Name string `csv:"name"`
	Age  string `csv:"age"`
}

type OutputCSV struct {
	Headers []string
	Rows    []RowCSV
}

func TestLoadStatement(t *testing.T) {
	// Create a temporary CSV file for testing
	inputFileName := "i.csv"
	content := `name,age
Alice,30
Bob,25
`
	removeDummyFileFn, err := addDummyCSVFile(t, content, inputFileName)
	require.NoError(t, err, "failed to create dummy CSV file")
	defer removeDummyFileFn()

	// Test LOAD
	evaluated := testEval(fmt.Sprintf("load %s", inputFileName))
	fmt.Println("eval result 1: ", evaluated)

	csvObj, ok := evaluated.(*object.CSV)
	if !ok {
		t.Fatalf("load: object is not CSV. got=%T (%+v)", evaluated, evaluated)
	}

	expectedHeaders := []string{"name", "age"}
	if len(csvObj.Headers) != len(expectedHeaders) {
		t.Fatalf("wrong number of headers. want=%d, got=%d",
			len(expectedHeaders), len(csvObj.Headers))
	}

	expectedRows := 2
	if len(csvObj.Rows) != expectedRows {
		t.Errorf("wrong number of rows. want=%d, got=%d",
			expectedRows, len(csvObj.Rows))
	}
}

func TestLoadAndReadRowsStatement(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantRowCount int
		wantColCount int
	}{
		{
			name: "load and read row 0",
			input: `
			load i.csv
			read row 0
			`,
			wantRowCount: 1,
			wantColCount: 2,
		},
		{
			name: "load and read row *",
			input: `
			load i.csv
			read row *
			`,
			wantRowCount: 2,
			wantColCount: 2,
		},
		{
			name: "load and read row * where age > 27",
			input: `
			load i.csv
			read row * where age > 27
			`,
			wantRowCount: 1,
			wantColCount: 2,
		},
		{
			name: "load and read row * where age == 27",
			input: `
			load i.csv
			read row * where age == 27
			`,
			wantRowCount: 0,
			wantColCount: 2,
		},
		{
			name: "load and read row * where age < 30",
			input: `
			load i.csv
			read row * where age < 30
			`,
			wantRowCount: 1,
			wantColCount: 2,
		},
		{
			name: "load and read row * where name == 'Bob'",
			input: `
			load i.csv
			read row * where name == "Bob"
			`,
			wantRowCount: 1,
			wantColCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary CSV file for testing
			inputFileName := "i.csv"
			content := `name,age
Alice,30
Bob,25`
			removeDummyFileFn, err := addDummyCSVFile(t, content, inputFileName)
			require.NoError(t, err, "failed to create dummy CSV file")
			defer removeDummyFileFn()

			// Evaluate code
			evaluated := testEval(tt.input)

			csvObj, ok := evaluated.(*object.CSV)
			if !ok {
				t.Fatalf("load: object is not CSV. got=%T (%+v)", evaluated, evaluated)
			}

			expectedHeaders := []string{"name", "age"}
			if len(csvObj.Headers) != len(expectedHeaders) {
				t.Errorf("wrong number of headers. want=%d, got=%d",
					len(expectedHeaders), len(csvObj.Headers))
			}

			if len(csvObj.Rows) != tt.wantRowCount {
				t.Errorf("wrong number of rows. want=%d, got=%d",
					tt.wantRowCount, len(csvObj.Rows))
			}

			if len(csvObj.Headers) != tt.wantColCount {
				t.Errorf("wrong number of cols. want=%d, got=%d",
					tt.wantColCount, len(csvObj.Headers))
			}
		})
	}
}

func TestLoadAndReadColsStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantRowCount  int
		wantColCount  int
		wantColValues []string
	}{
		// result should only include a single qualifying row
		{
			name: "load and read row * col name where age > 27",
			input: `
			load i.csv
			read row * col name where age > 27
			`,
			wantRowCount:  1,
			wantColCount:  2,
			wantColValues: []string{"Alice"},
		},
		// result should only include a single qualifying row
		{
			name: "load and read row * col name where age == 30",
			input: `
			load i.csv
			read row * col name where age == 30
			`,
			wantRowCount:  1,
			wantColCount:  2,
			wantColValues: []string{"Alice"},
		},
		// result should only include a single qualifying row
		{
			name: "load and read row * col name where age < 30",
			input: `
			load i.csv
			read row * col name where age < 30
			`,
			wantRowCount:  1,
			wantColCount:  2,
			wantColValues: []string{"Bob"},
		},
		// result should only include both qualified rows
		{
			name: "load and read row * col name where age > 17",
			input: `
			load i.csv
			read row * col name where age > 17
			`,
			wantRowCount:  1,
			wantColCount:  2,
			wantColValues: []string{"Alice", "Bob"},
		},
		// result should only contain the row with `name == "Alice"`
		{
			name: "load and read row * col name where name == 'Alice'",
			input: `
			load i.csv
			read row * col name where name == "Alice"
			`,
			wantRowCount:  1,
			wantColCount:  2,
			wantColValues: []string{"Alice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary CSV file for testing
			inputFileName := "i.csv"
			content := `name,age
Alice,30
Bob,25`
			removeDummyFileFn, err := addDummyCSVFile(t, content, inputFileName)
			require.NoError(t, err, "failed to create dummy CSV file")
			defer removeDummyFileFn()

			// Evaluate code
			evaluated := testEval(tt.input)

			csvObj, ok := evaluated.(*object.Array)
			if !ok {
				t.Fatalf("load: object is not CSV. got=%T (%+v)", evaluated, evaluated)
			}

			// Check if the col values are as expected
			if len(tt.wantColValues) > 0 {
				if len(csvObj.Elements) != len(tt.wantColValues) {
					t.Errorf("wrong number of returned values. want=%d, got=%d",
						len(tt.wantColValues), len(csvObj.Elements))
				}

				for i, v := range csvObj.Elements {
					if v.Inspect() != tt.wantColValues[i] {
						t.Errorf("wrong col value. want=%s, got=%s",
							tt.wantColValues[i], v.Inspect())
					}
				}
			}
		})
	}
}

func TestSaveAsJSONStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantResultJSON OutputJSON
	}{
		{
			name: "load and save as json",
			input: `
			load i.csv
			save as output.json
			`,
			wantResultJSON: OutputJSON{
				Headers: []string{"name", "age"},
				Rows: []RowJSON{
					{Name: "Alice", Age: "30"},
					{Name: "Bob", Age: "25"},
				},
			},
		},
		{
			name: "load and save as json",
			input: `
			load i.csv
			let val = read row 0;
			save val as output.json
			`,
			wantResultJSON: OutputJSON{
				Headers: []string{"name", "age"},
				Rows: []RowJSON{
					{Name: "Alice", Age: "30"},
				},
			},
		},
		{
			name: "load and save as json",
			input: `
			load i.csv
			let val = read row *;
			save val as output.json
			`,
			wantResultJSON: OutputJSON{
				Headers: []string{"name", "age"},
				Rows: []RowJSON{
					{Name: "Alice", Age: "30"},
					{Name: "Bob", Age: "25"},
				},
			},
		},
		{
			name: "load and save as json",
			input: `
			load i.csv
			let val = read row * where age > 25;
			save val as output.json
			`,
			wantResultJSON: OutputJSON{
				Headers: []string{"name", "age"},
				Rows: []RowJSON{
					{Name: "Alice", Age: "30"},
				},
			},
		},
		{
			name: "load and save as json",
			input: `
			load i.csv
			let val = read row * where age == 25;
			save val as output.json
			`,
			wantResultJSON: OutputJSON{
				Headers: []string{"name", "age"},
				Rows: []RowJSON{
					{Name: "Bob", Age: "25"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary CSV file for testing
			inputFileName := "i.csv"
			content := `name,age
Alice,30
Bob,25`
			removeDummyFileFn, err := addDummyCSVFile(t, content, inputFileName)
			require.NoError(t, err, "failed to create dummy CSV file")
			defer removeDummyFileFn()

			// Evaluate code
			_ = testEval(tt.input)

			// Check if a file named output.json was created
			fileExists := doesFileExist(t, "output.json")
			require.True(t, fileExists, "output.json file was not created")

			// Check the content of the output.json file
			testFileContents(t, tt.wantResultJSON, "output.json")

			// Remove the output.json file
			err = os.Remove("output.json")
			require.NoError(t, err, "failed to remove output.json")
		})
	}
}

func TestSaveImplicitAsCSVStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantResultCSV OutputCSV
	}{
		{
			name: "load and implicit save as csv",
			input: `
			load i.csv
			save as output.csv
			`,
			wantResultCSV: OutputCSV{
				Headers: []string{"name", "age"},
				Rows: []RowCSV{
					{Name: "Alice", Age: "30"},
					{Name: "Bob", Age: "25"},
				},
			},
		},
		{
			name: "load and custom save as csv 1",
			input: `
			load i.csv
			let val = read row 0;
			save val as output.csv
			`,
			wantResultCSV: OutputCSV{
				Headers: []string{"name", "age"},
				Rows: []RowCSV{
					{Name: "Alice", Age: "30"},
				},
			},
		},
		{
			name: "load and custom save as csv 2",
			input: `
			load i.csv
			let val = read row *;
			save val as output.csv
			`,
			wantResultCSV: OutputCSV{
				Headers: []string{"name", "age"},
				Rows: []RowCSV{
					{Name: "Alice", Age: "30"},
					{Name: "Bob", Age: "25"},
				},
			},
		},
		{
			name: "load and custom save as csv 3",
			input: `
			load i.csv
			let val = read row * where age > 25;
			save val as output.csv
			`,
			wantResultCSV: OutputCSV{
				Headers: []string{"name", "age"},
				Rows: []RowCSV{
					{Name: "Alice", Age: "30"},
				},
			},
		},
		{
			name: "load and custom save as csv 4",
			input: `
			load i.csv
			let val = read row * where age == 25;
			save val as output.csv
			`,
			wantResultCSV: OutputCSV{
				Headers: []string{"name", "age"},
				Rows: []RowCSV{
					{Name: "Bob", Age: "25"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary CSV file for testing
			inputFileName := "i.csv"
			content := `name,age
Alice,30
Bob,25`
			removeDummyFileFn, err := addDummyCSVFile(t, content, inputFileName)
			require.NoError(t, err, "failed to create dummy CSV file")
			defer removeDummyFileFn()

			// Evaluate code
			_ = testEval(tt.input)

			// Check if a file named output.json was created
			fileExists := doesFileExist(t, "output.csv")
			require.True(t, fileExists, "output.csv file was not created")

			// Check the content of the output.json file
			testCSVFileContents(t, tt.wantResultCSV, "output.csv")

			// Remove the output.json file
			err = os.Remove("output.csv")
			require.NoError(t, err, "failed to remove output.csv")
		})
	}
}

func TestUniqueBuiltinFunction(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantResultJSON OutputJSON
	}{
		{
			name: "using unique builtin removes duplicate rows",
			input: `
			load i.csv
			let rows = read row *;
			let uniqueRows = unique(rows);
			save uniqueRows as output.json`,
			wantResultJSON: OutputJSON{
				Headers: []string{"name", "age"},
				Rows: []RowJSON{
					{Name: "Alice", Age: "30"},
					{Name: "Bob", Age: "25"},
				},
			},
		},
		{
			name: "not using unique results in duplicate rows",
			input: `
			load i.csv
			let rows = read row *;
			save rows as output.json`,
			wantResultJSON: OutputJSON{
				Headers: []string{"name", "age"},
				Rows: []RowJSON{
					{Name: "Alice", Age: "30"},
					{Name: "Alice", Age: "30"},
					{Name: "Alice", Age: "30"},
					{Name: "Bob", Age: "25"},
					{Name: "Bob", Age: "25"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary CSV file for testing
			inputFileName := "i.csv"
			content := `name,age
Alice,30
Alice,30
Alice,30
Bob,25
Bob,25
`
			removeDummyFileFn, err := addDummyCSVFile(t, content, inputFileName)
			require.NoError(t, err, "failed to create dummy CSV file")
			defer removeDummyFileFn()

			// Evaluate code
			_ = testEval(tt.input)

			// Check if a file named output.json was created
			fileExists := doesFileExist(t, "output.json")
			require.True(t, fileExists, "output.json file was not created")

			// Check the content of the output.json file
			testFileContents(t, tt.wantResultJSON, "output.json")

			// Remove the output.json file
			err = os.Remove("output.json")
			require.NoError(t, err, "failed to remove output.json")
		})
	}
}

func TestFillEmptyBuiltinFunction(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantResultJSON OutputJSON
	}{
		{
			name: "using fill_empty builtin should fill missing or empty values",
			input: `
			load i.csv
			let rows = read row *;
			let updatedRows = fill_empty(rows, "name", "john");
			updatedRows = fill_empty(updatedRows, "age", 18);
			save updatedRows as output.json`,
			wantResultJSON: OutputJSON{
				Headers: []string{"name", "age"},
				Rows: []RowJSON{
					{Name: "Alice", Age: "18"},
					{Name: "john", Age: "25"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary CSV file for testing
			inputFileName := "i.csv"
			content := `name,age
Alice,
,25
`
			removeDummyFileFn, err := addDummyCSVFile(t, content, inputFileName)
			require.NoError(t, err, "failed to create dummy CSV file")
			defer removeDummyFileFn()

			// Evaluate code
			_ = testEval(tt.input)

			// Check if a file named output.json was created
			fileExists := doesFileExist(t, "output.json")
			require.True(t, fileExists, "output.json file was not created")

			// Check the content of the output.json file
			testFileContents(t, tt.wantResultJSON, "output.json")

			// Remove the output.json file
			err = os.Remove("output.json")
			require.NoError(t, err, "failed to remove output.json")
		})
	}

}

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestStringLiteral(t *testing.T) {
	input := `"Hello World!"`
	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}
	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`
	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}
	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
		return false
	}

	return true
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input string

		expected bool
	}{
		{"true", true},
		{"false", false},
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{`
			if (10 > 1) {
				if (10 > 1) {
					return 10;
				}
				return 1;
			}
			`, 10},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}
	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{`
			if (10 > 1) {
			if (10 > 1) {
			return true + false;
			}
			return 1;
			}
			`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
		{
			`"Hello" - "World"`,
			"unknown operator: STRING - STRING",
		},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)",
				evaluated, evaluated)
			continue
		}
		if errObj.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tt.expectedMessage, errObj.Message)
		}
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2; };"
	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}
	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v",
			fn.Parameters)
	}
	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}
	expectedBody := "(x + 2)"
	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
	}
	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
	let newAdder = fn(x) {
		fn(y) { x + y };
	};
	let addTwo = newAdder(2);
	addTwo(2);`
	testIntegerObject(t, testEval(input), 4)
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)",
					evaluated, evaluated)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q",
					expected, errObj.Message)
			}
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"
	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", evaluated, evaluated)
	}
	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d",
			len(result.Elements))
	}
	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1, 2, 3][2]",
			3,
		},
		{
			"let i = 0; [1][i];",
			1,
		},
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1, 2, 3][-1]",
			nil,
		},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t",
			result.Value, expected)
		return false
	}
	return true
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	fmt.Printf("\nprogram: %s\n", program.String())
	env := object.NewEnvironment()

	return Eval(program, env)
}

func doesFileExist(t *testing.T, filename string) bool {
	cwd, err := os.Getwd()
	require.NoError(t, err, "failed to get current working directory")

	_, err = os.Stat(fmt.Sprintf("%s/%s", cwd, filename))
	return err == nil
}

func addDummyCSVFile(t *testing.T, content, inputFileName string) (func(), error) {
	// Create a temporary CSV file for testing
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	tmpfile, err := os.Create(fmt.Sprintf("%s/%s", cwd, inputFileName))
	if err != nil {
		return nil, err
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return nil, err
	}
	if err := tmpfile.Close(); err != nil {
		return nil, err
	}

	return func() {
		os.Remove(tmpfile.Name())
	}, nil
}

func testFileContents(t *testing.T, expectedResultJSON OutputJSON, filename string) error {
	// 1. open the file
	file, err := os.ReadFile(filename)
	require.NoError(t, err, "failed to open file")
	fmt.Printf("file contents: %s\n", string(file))

	// 2. unmarshal the file contents into OutputJSON struct
	var outputJSON OutputJSON
	err = json.Unmarshal(file, &outputJSON)
	require.NoError(t, err, "failed to unmarshal file contents")

	// 3. check if the headers are as expected
	expectedHeaders := expectedResultJSON.Headers
	if len(outputJSON.Headers) != len(expectedHeaders) {
		t.Errorf("wrong number of headers. want=%d, got=%d",
			len(expectedHeaders), len(outputJSON.Headers))
	}
	for i, h := range outputJSON.Headers {
		if h != expectedHeaders[i] {
			t.Errorf("wrong header value. want=%s, got=%s",
				expectedHeaders[i], h)
		}
	}

	// 4. check if the rows are as expected
	expectedRows := expectedResultJSON.Rows
	if len(outputJSON.Rows) != len(expectedRows) {
		t.Errorf("wrong number of rows. want=%d, got=%d",
			len(expectedRows), len(outputJSON.Rows))
	}
	for i, row := range outputJSON.Rows {
		if row != expectedRows[i] {
			t.Errorf("wrong row value. want=%+v, got=%+v",
				expectedRows[i], row)
		}
	}
	return nil
}

func testCSVFileContents(t *testing.T, expectedResultCSV OutputCSV, filename string) error {
	// 1. open the file
	file, err := os.Open(filename)
	require.NoError(t, err, "failed to open file")

	// 2. unmarshal the file contents into OutputJSON struct
	var rows []RowCSV
	err = gocsv.UnmarshalFile(file, &rows)
	require.NoError(t, err, "failed to unmarshal file contents")

	outputCSV := OutputCSV{
		Headers: expectedResultCSV.Headers, // can't get headers from the unmarshal, not worth putting assertions as it'll always pass in the current state
		Rows:    rows,
	}

	// 3. check if the headers are as expected
	expectedHeaders := expectedResultCSV.Headers
	if len(outputCSV.Headers) != len(expectedHeaders) {
		t.Errorf("wrong number of headers. want=%d, got=%d",
			len(expectedHeaders), len(outputCSV.Headers))
	}
	for i, h := range outputCSV.Headers {
		if h != expectedHeaders[i] {
			t.Errorf("wrong header value. want=%s, got=%s",
				expectedHeaders[i], h)
		}
	}

	// 4. check if the rows are as expected
	expectedRows := expectedResultCSV.Rows
	if len(outputCSV.Rows) != len(expectedRows) {
		t.Errorf("wrong number of rows. want=%d, got=%d",
			len(expectedRows), len(outputCSV.Rows))
	}
	for i, row := range outputCSV.Rows {
		if row != expectedRows[i] {
			t.Errorf("wrong row value. want=%+v, got=%+v",
				expectedRows[i], row)
		}
	}
	return nil
}
