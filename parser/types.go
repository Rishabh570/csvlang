package parser

import (
	"fmt"
	"runtime"
)

type ParserError struct {
	Message string
	Stack   []uintptr
	Line    int
	Column  int
}

func (e *ParserError) Error() string {
	return fmt.Sprintf("%s at line %d, column %d", e.Message, e.Line, e.Column)
}

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
