package main

func main() {
	program := []byte{OP_PUSH, 2, OP_PUSH, 3, OP_CALL, 7, OP_HALT, OP_ADD, OP_RET}
	runProgram(program)
}
