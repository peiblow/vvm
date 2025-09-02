package main

import "fmt"

func RunProgram(compile *Compiler) {
	code := compile.Code

	stack := []interface{}{}
	callStack := []int{}

	storage := make(map[int]interface{})
	mstorage := make(map[int]interface{})

	ip := 0
	for {
		op := code[ip]
		ip++

		switch op {
		case OP_CONST:
			idx := int(code[ip])
			ip++
			val := compile.ConstPool[idx]
			stack = append(stack, val)
		case OP_PUSH:
			val := int(code[ip])
			ip++
			stack = append(stack, val)
		case OP_ADD:
			if len(stack) < 2 {
				panic("stack underflow on ADD")
			}

			b := stack[len(stack)-1]
			a := stack[len(stack)-2]

			switch av := a.(type) {
			case int:
				if bv, ok := b.(int); ok {
					stack = stack[:len(stack)-2]
					stack = append(stack, av+bv)
				}
			case string:
				if bv, ok := b.(string); ok {
					stack = stack[:len(stack)-2]
					stack = append(stack, av+bv)
				}
			default:
				panic("unsupported ADD type")
			}
		case OP_SUB:
			a := stack[len(stack)-1].(int)
			b := stack[len(stack)-2].(int)
			stack = stack[:len(stack)-2]
			stack = append(stack, a-b)
		case OP_MUL:
			a := stack[len(stack)-1].(int)
			b := stack[len(stack)-2].(int)
			stack = stack[:len(stack)-2]
			stack = append(stack, a*b)
		case OP_DIV:
			a := stack[len(stack)-1].(int)
			b := stack[len(stack)-2].(int)
			stack = stack[:len(stack)-2]
			stack = append(stack, a/b)
		case OP_GT:
			a := stack[len(stack)-2].(int)
			b := stack[len(stack)-1].(int)
			stack = stack[:len(stack)-2]

			if a > b {
				stack = append(stack, 1)
			} else {
				stack = append(stack, 0)
			}
		case OP_GT_EQ:
			a := stack[len(stack)-2].(int)
			b := stack[len(stack)-1].(int)
			stack = stack[:len(stack)-2]

			if a >= b {
				stack = append(stack, 1)
			} else {
				stack = append(stack, 0)
			}
		case OP_LT:
			a := stack[len(stack)-2].(int)
			b := stack[len(stack)-1].(int)

			stack = stack[:len(stack)-2]

			if a < b {
				stack = append(stack, 1)
			} else {
				stack = append(stack, 0)
			}
		case OP_LT_EQ:
			a := stack[len(stack)-2].(int)
			b := stack[len(stack)-1].(int)

			stack = stack[:len(stack)-2]

			if a <= b {
				stack = append(stack, 1)
			} else {
				stack = append(stack, 0)
			}
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
			if len(stack) == 0 {
				fmt.Println("Warning: OP_PRINT with empty stack, ignoring")
				break
			}
			val := stack[len(stack)-1]
			fmt.Println(val)
		case OP_NOP:
		case OP_JMP:
			destiny := int(code[ip])
			ip++
			ip = destiny
		case OP_JMP_IF:
			destiny := int(code[ip])
			ip++

			cond := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if cond == 0 {
				ip = destiny
			}
		case OP_CALL:
			destiny := int(code[ip])
			ip++

			funcArgs := compile.GetFuncArgs(destiny)
			for i := len(funcArgs) - 1; i >= 0; i-- {
				val := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				slot := funcArgs[i]
				storage[slot] = val
			}

			callStack = append(callStack, ip)
			ip = destiny
		case OP_RET:
			if len(callStack) == 0 {
				// fmt.Println("Warning: OP_RET with empty callStack, ignoring")
				break
			}

			returnAddr := callStack[len(callStack)-1]
			callStack = callStack[:len(callStack)-1]
			ip = returnAddr
		case OP_ACCESS:
			idx := stack[len(stack)-1].(int)
			arr := stack[len(stack)-2].([]interface{})
			stack = stack[:len(stack)-2]

			if idx < 0 || idx >= len(arr) {
				panic(fmt.Sprintf("Array index out of bounds: %d", idx))
			}

			val := arr[idx]
			stack = append(stack, val)
		case OP_STORE:
			key := int(code[ip])
			ip++

			val := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			storage[key] = val
		case OP_SLOAD:
			key := int(code[ip])
			ip++
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
