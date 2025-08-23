package main

import (
	"os"

	"github.com/peiblow/vvm/lexer"
	"github.com/peiblow/vvm/parser"
	"github.com/sanity-io/litter"
)

func main() {
	content, _ := os.ReadFile("00.synx")
	tokens := lexer.Tokenize(string(content))

	// for _, token := range tokens {
	// 	token.Debug()
	// }

	ast := parser.Parse(tokens)
	litter.Dump(ast)
}
