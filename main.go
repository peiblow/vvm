package main

import (
	"fmt"
	"os"
)

func main() {
	content, _err := os.ReadFile("test.synx")

	if _err != nil {
		fmt.Println(_err)
	}

	contract := tokenize(string(content))
	fmt.Println("-> ", contract)
}
