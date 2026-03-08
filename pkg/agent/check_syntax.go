//go:build ignore

package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
)

func main() {
	files := os.Args[1:]
	fset := token.NewFileSet()
	hasErrors := false
	
	for _, file := range files {
		_, err := parser.ParseFile(fset, file, nil, parser.AllErrors)
		if err != nil {
			fmt.Printf("ERROR in %s: %v\n", file, err)
			hasErrors = true
		} else {
			fmt.Printf("OK: %s\n", file)
		}
	}
	
	if hasErrors {
		os.Exit(1)
	}
}
