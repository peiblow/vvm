package main

func main() {
	program := []byte{
		OP_PUSH, 10,
		OP_STORE, 1,
		OP_PUSH, 20,
		OP_STORE, 2,

		OP_SLOAD, 1,
		OP_SLOAD, 2,
		OP_ADD,
		OP_STORE, 3,

		OP_SLOAD, 3,
		OP_PUSH, 10,
		OP_SUB,
		OP_JMP_IF, 22,

		OP_PUSH, 1,
		OP_PRINT,
		OP_JMP, 30,

		OP_PUSH, 2,
		OP_PRINT,

		OP_HALT,
	}
	runProgram(program)
}
