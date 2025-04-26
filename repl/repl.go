// repl package is responsible for handling the file mode (repl mode is not supported) which reads the entire file content and evaluates the entire program at once.
//
// This is useful when you want to use variables defined on one line in another line.
// StartFileAllAtOnce reads the entire file content and evaluates the entire program at once.
package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Rishabh570/csvlang/evaluator"
	"github.com/Rishabh570/csvlang/lexer"
	"github.com/Rishabh570/csvlang/object"
	"github.com/Rishabh570/csvlang/parser"
	"github.com/Rishabh570/csvlang/token"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors) != 0 {
			printParserErrors(out, p.Errors)
			continue
		}
		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

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
	// print all tokens from lexer
	// for {
	// 	tok := l.NextToken()
	// 	if tok.Type == token.EOF {
	// 		break
	// 	}
	// 	fmt.Printf("🚧 reading program token: %s\n", tok)
	// }

	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors) != 0 {
		printParserErrors(os.Stdout, p.Errors)
		return
	}

	// Evaluate each statement in the program
	for _, statement := range program.Statements {
		fmt.Printf("🚧 evaluating program statement: %s\n", statement.String())
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

func StartLexer(path string) {
	// Read the entire file content
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		os.Exit(1)
	}

	// Parse and evaluate the entire program
	l := lexer.New(string(content))
	// print all tokens from lexer
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			break
		}
		fmt.Printf("🚧 reading program token: %s\n", tok)
	}

}

func StartFile(path string) {
	// Read all lines from the file
	lines, err := readLines(path)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		os.Exit(1)
	}

	// Create a single environment to be used across all evaluations
	env := object.NewEnvironment()

	// Evaluate each line sequentially
	for _, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}
		fmt.Printf("🚧 reading program line: %s\n", line)

		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		// TODO: support stack traces?
		if len(p.Errors) != 0 {
			printParserErrors(os.Stdout, p.Errors)
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			io.WriteString(os.Stdout, evaluated.Inspect())
			io.WriteString(os.Stdout, "\n")
		}
	}
}

func StartRPPL(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors) != 0 {
			printParserErrors(out, p.Errors)
			continue
		}
		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
	}
}

// Helper function to read lines from file
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

const MONKEY_FACE = `   			__,__
			 .--. .-" "-. .--.
	  / .. \/ .-. .-. \/ .. \
		  | | '| / Y \ |' | |
		 | \ \ \ 0 | 0 / / / |
		\ '- ,\.-"""""""-./, -' /
	  	''-' /_ ^ ^ _\ '-''
					| \._ _./ |
					\ \ '~' / /
				 '._ '-=-' _.'
				    '-----'
`

func printParserErrors(out io.Writer, errors []*parser.ParserError) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, err := range errors {
		fmt.Fprintf(out, "%+v\n", err) // Using %+v to get detailed error output
	}
}
