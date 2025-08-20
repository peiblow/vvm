package main

import "fmt"

func runProgram(code []byte) {
	stack := []int{}
	callStack := []int{}
	retStack := []int{}

	storage := make(map[int]int)
	mstorage := make(map[int]int)

	ip := 0
	for {
		op := code[ip]
		ip++

		switch op {
		case OP_PUSH:
			val := int(code[ip])
			ip++
			stack = append(stack, val)
		case OP_ADD:
			a := stack[len(stack)-1]
			b := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, a+b)
		case OP_SUB:
			a := stack[len(stack)-1]
			b := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, a-b)
		case OP_MUL:
			a := stack[len(stack)-1]
			b := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, a*b)
		case OP_DIV:
			a := stack[len(stack)-1]
			b := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, a/b)
		case OP_POP:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			} else {
				fmt.Println("Erro: Empty stack")
			}
		case OP_DUP:
			val := stack[len(stack)-1]
			stack = append(stack, val)
		case OP_SWAP:
			a := stack[len(stack)-1]
			b := stack[len(stack)-2]

			stack[len(stack)-1] = b
			stack[len(stack)-2] = a
		case OP_PRINT:
			val := stack[len(stack)-1]
			fmt.Println(val)
		case OP_NOP:
		case OP_JMP:
			ip++
			destiny := int(code[ip])
			ip = destiny
		case OP_JMP_IF:
			ip++
			destiny := int(code[ip])

			cond := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if cond != 0 {
				ip = destiny
			} else {
				ip++
			}
		case OP_CALL:
			destiny := int(code[ip])
			ip++
			callStack = append(callStack, ip)
			retStack = append(retStack, ip)
			ip = destiny
		case OP_RET:
			returnAddr := callStack[len(callStack)-1]
			callStack = callStack[:len(callStack)-1]
			ip = returnAddr
		case OP_STORE:
			ip++
			key := int(code[ip])
			val := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			storage[key] = val
		case OP_SLOAD:
			key := int(code[ip])
			val, ok := storage[key]

			if !ok {
				val = 0
			}

			stack = append(stack, val)
		case OP_MSTORE:
			ip++
			key := int(code[ip])
			val := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			mstorage[key] = val
		case OP_MLOAD:
			key := int(code[ip])
			val, ok := mstorage[key]

			if !ok {
				val = 0
			}

			stack = append(stack, val)
		case OP_DELETE:
			ip++
			key := int(code[ip])
			delete(storage, key)
		case OP_HALT:
			fmt.Println("Fim do programa.")
			return
		default:
			fmt.Println("Opcode desconhecido:", op)
			return
		}
	}
}
