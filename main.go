package main

import (
	"os"

	"github.com/peiblow/vvm/lexer"
)

func main() {
	content, _ := os.ReadFile("01.synx")

	tokens := lexer.Tokenize(string(content))
	for _, token := range tokens {
		token.Debug()
	}
}
