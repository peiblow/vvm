package main

import (
	"os"
)

func main() {
	content, _ := os.ReadFile("01.synx")

	tokens := tokenize(string(content))
	for _, token := range tokens {
		token.Debug()
	}
}
