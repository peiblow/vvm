package main

import (
	"fmt"
	"os"

	"github.com/peiblow/vvm/commiter"
	"github.com/peiblow/vvm/compiler"
	"github.com/peiblow/vvm/lexer"
	"github.com/peiblow/vvm/parser"
	"github.com/peiblow/vvm/vm"
)

func main() {
	content, _ := os.ReadFile("00.snx")
	tokens := lexer.Tokenize(string(content))

	ast := parser.Parse(tokens)
	// litter.Dump(ast)

	// Code Compilation
	cmpl := compiler.New()
	cmpl.CompileBlock(ast)

	// Debug Bytecode
	// cmpl.Debug()

	// Execute VM
	virtualMachine := vm.New(cmpl)
	result := virtualMachine.Run()
	if result.Success {
		println("Program executed successfully.")

		committer := &commiter.MockCommitter{}
		if err := committer.Commit(result.Journal); err != nil {
			fmt.Println("❌ Commit failed:", err)
			return
		}
	} else {
		println("Program execution failed:", result.Error.Error())
	}
}
