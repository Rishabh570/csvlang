// csvlang is a simple programming language that can be used to manipulate CSV files.
//
// The main.go file is the entry point of the application. It reads the file path from the command line arguments and evaluates the code in the file.
package main

import (
	"flag"
	"fmt"

	"github.com/Rishabh570/csvlang/repl"
)

func main() {
	// user, err := user.Current()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("Hello %s! This is the Monkey programming language!\n", user.Username)
	// fmt.Printf("Feel free to type in commands\n")
	// repl.Start(os.Stdin, os.Stdout)

	// Define a string flag called "path" with a default value of "" and a brief description.
	filePath := flag.String("path", "", "Path to the file")

	// Parse the command line flags.
	flag.Parse()

	// Use the file path after parsing. If it's empty, it means the flag was not provided.
	if *filePath == "" {
		fmt.Println("Please provide a file path using the -path flag.")
		return
	}

	// Output the provided file path.
	fmt.Printf("File path: %s\n", *filePath)

	// repl.StartFile(*filePath)
	repl.StartFileAllAtOnce(*filePath)
}
