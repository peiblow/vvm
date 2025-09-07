package main

import (
	"os"

	"github.com/peiblow/vvm/lexer"
	"github.com/peiblow/vvm/parser"
)

func main() {
	content, _ := os.ReadFile("01.snx")
	tokens := lexer.Tokenize(string(content))

	// for _, token := range tokens {
	// 	token.Debug()
	// }

	ast := parser.Parse(tokens)
	// litter.Dump(ast)

	cmpl := NewCompiler()
	cmpl.CompileBlock(ast)

	// opcodeNames := map[byte]string{
	// 	OP_PUSH:         "PUSH",
	// 	OP_ADD:          "ADD",
	// 	OP_SUB:          "SUB",
	// 	OP_MUL:          "MUL",
	// 	OP_DIV:          "DIV",
	// 	OP_GT:           "GT",
	// 	OP_LT:           "LT",
	// 	OP_EQ:           "EQ",
	// 	OP_POP:          "POP",
	// 	OP_DUP:          "DUP",
	// 	OP_SWAP:         "SWAP",
	// 	OP_PRINT:        "PRINT",
	// 	OP_NOP:          "NOP",
	// 	OP_JMP:          "JMP",
	// 	OP_JMP_IF:       "JMP_IF",
	// 	OP_CALL:         "CALL",
	// 	OP_RET:          "RET",
	// 	OP_STORE:        "STORE",
	// 	OP_SLOAD:        "SLOAD",
	// 	OP_MSTORE:       "MSTORE",
	// 	OP_MLOAD:        "MLOAD",
	// 	OP_DELETE:       "DELETE",
	// 	OP_HALT:         "HALT",
	// 	OP_CONST:        "CONST",
	// 	OP_GET_PROPERTY: "GET_PROPERTY",
	// }

	// for i := 0; i < len(cmpl.Code); i++ {
	// 	op := cmpl.Code[i]
	// 	name, ok := opcodeNames[op]
	// 	if !ok {
	// 		name = fmt.Sprintf("UNKNOWN_%02X", op)
	// 	}
	// 	fmt.Print(name)
	// 	switch op {
	// 	case OP_PUSH, OP_STORE, OP_SLOAD, OP_MSTORE, OP_MLOAD, OP_CALL, OP_JMP, OP_JMP_IF, OP_CONST:
	// 		i++
	// 		fmt.Printf(" %d", cmpl.Code[i])
	// 	}
	// 	fmt.Println()
	// }

	RunProgram(cmpl)
}
