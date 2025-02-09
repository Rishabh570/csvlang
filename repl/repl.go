// repl package is responsible for handling the file mode (repl mode is not supported) which reads the entire file content and evaluates the entire program at once.
//
// This is useful when you want to use variables defined on one line in another line.
// StartFileAllAtOnce reads the entire file content and evaluates the entire program at once.
package repl

import (
	"fmt"
	"io"
	"os"

	"github.com/Rishabh570/csvlang/evaluator"
	"github.com/Rishabh570/csvlang/lexer"
	"github.com/Rishabh570/csvlang/object"
	"github.com/Rishabh570/csvlang/parser"
)

const PROMPT = ">> "

// StartFileAllAtOnce reads the entire file content and evaluates the entire program at once.
// This helps when, for instance, you want to use variables defined on one line in another line.
func StartFileAllAtOnce(path string) {
	// Read the entire file content
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		os.Exit(1)
	}

	// Create environment
	env := object.NewEnvironment()

	// Parse and evaluate the entire program
	l := lexer.New(string(content))
	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors) != 0 {
		printParserErrors(os.Stdout, p.Errors)
		return
	}

	// Evaluate each statement in the program
	for _, statement := range program.Statements {
		// fmt.Printf("🚧 evaluating program statement: %s\n", statement.String())
		evaluated := evaluator.Eval(statement, env)
		if evaluated != nil {
			io.WriteString(os.Stdout, evaluated.Inspect())
			io.WriteString(os.Stdout, "\n")

			// Stop further execution if an error is encountered
			if evaluated.Type() == object.ERROR_OBJ {
				return
			}
		}
	}
}

func printParserErrors(out io.Writer, errors []*parser.ParserError) {
	io.WriteString(out, "Woops! We ran into some error here\n")
	io.WriteString(out, " parser errors:\n")
	for _, err := range errors {
		fmt.Fprintf(out, "%+v\n", err)
	}
}
