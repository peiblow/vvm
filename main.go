package main

import (
	"os"

	"github.com/peiblow/vvm/compiler"
	"github.com/peiblow/vvm/lexer"
	"github.com/peiblow/vvm/parser"
	"github.com/peiblow/vvm/vm"
)

func main() {
	content, _ := os.ReadFile("01.snx")
	tokens := lexer.Tokenize(string(content))

	ast := parser.Parse(tokens)

	// Compila o código
	cmpl := compiler.New()
	cmpl.CompileBlock(ast)

	// Debug (descomente para ver informações de debug)
	cmpl.Debug()

	// Executa o programa
	virtualMachine := vm.New(cmpl)
	virtualMachine.Run()
}
