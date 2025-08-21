package main

import (
	"os"
)

func main() {
	content, _ := os.ReadFile("00.synx")

	tokens := tokenize(string(content))
	for _, token := range tokens {
		token.Debug()
	}
}
