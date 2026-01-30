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
	mode := os.Args[1]

	switch mode {
	case "deploy":
		deployContract()
	case "exec":
		if len(os.Args) < 3 {
			fmt.Println("Usage: run <function_name> [args...]")
			return
		}
		funcName := os.Args[2]
		runContract(&compiler.ContractArtifact{
			Bytecode:     []byte{0x01, 0x02, 0x03}, // Placeholder bytecode
			ConstPool:    []interface{}{"Hello, World!", 42},
			Functions:    map[string]compiler.FunctionMeta{},
			FunctionName: map[int]string{},
			Types:        map[string]compiler.TypeMeta{},
		}, funcName)
	default:
		fmt.Println("Unknown mode. Use 'deploy' or 'run'.")
	}
}

func deployContract() {
	content, _ := os.ReadFile("./contracts/00.snx")
	tokens := lexer.Tokenize(string(content))

	ast := parser.Parse(tokens)
	// litter.Dump(ast)

	cmpl := compiler.New()
	cmpl.CompileBlock(ast)

	// Debug Bytecode
	// cmpl.Debug()

	artifact := cmpl.Artifact()
	fmt.Println("Contract deployed successfully.", artifact)
}

func runContract(contractArtifact *compiler.ContractArtifact, funcName string, args ...interface{}) {
	virtualMachine := vm.NewFromArtifact(contractArtifact)
	initResult := virtualMachine.Run()
	if !initResult.Success {
		fmt.Println("Contract initialization failed:", initResult.Error.Error())
		return
	}

	contractArtifact.InitStorage = virtualMachine.GetStorage()

	result := virtualMachine.RunFunction(funcName, args...)

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
