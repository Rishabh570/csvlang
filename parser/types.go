package parser

import (
	"fmt"
	"runtime"
)

// ParserError is an error type that is returned when a parsing error occurs.
type ParserError struct {
	Message string    // Parsing error message
	Stack   []uintptr // Stack trace
	Line    int       // Line number where the error occurred
	Column  int       // Column number where the error occurred
}

// Error creates a new ParserError with the given message, line, column, and stack trace.
func (e *ParserError) Error() string {
	return fmt.Sprintf("%s at line %d, column %d", e.Message, e.Line, e.Column)
}

// Format formats the ParserError for printing.
func (e *ParserError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			// Print full stack trace
			frames := runtime.CallersFrames(e.Stack)
			fmt.Fprintf(s, "%s at line %d, column %d\n", e.Message, e.Line, e.Column)
			for {
				frame, more := frames.Next()
				fmt.Fprintf(s, "\t%s\n\t\t%s:%d\n", frame.Function, frame.File, frame.Line)
				if !more {
					break
				}
			}
			return
		}
		fallthrough
	case 's':
		fmt.Fprintf(s, "%s at line %d, column %d", e.Message, e.Line, e.Column)
	}
}
