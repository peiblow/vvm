package main

import (
	"fmt"
	"reflect"
	"strconv"
)

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

			a := pop(&stack, "OP_ADD")
			b := pop(&stack, "OP_ADD")

			switch av := a.(type) {
			case int:
				switch bv := b.(type) {
				case int:
					push(&stack, av+bv)
				case string:
					push(&stack, strconv.Itoa(av)+bv)
				default:
					panic("unsupported ADD type")
				}
			case string:
				switch bv := b.(type) {
				case int:
					push(&stack, av+strconv.Itoa(bv))
				case string:
					push(&stack, av+bv)
				case float64:
					push(&stack, av+strconv.FormatFloat(bv, 'f', 0, 64))
				default:
					fmt.Println(reflect.TypeOf(bv))
					panic("unsupported ADD type")
				}
			default:
				panic("unsupported ADD type")
			}

		case OP_SUB:
			b := pop(&stack, "OP_SUB").(int)
			a := pop(&stack, "OP_SUB").(int)

			stack = append(stack, a-b)
		case OP_MUL:
			b := pop(&stack, "OP_MUL").(int)
			a := pop(&stack, "OP_MUL").(int)

			stack = append(stack, a*b)
		case OP_DIV:
			b := pop(&stack, "OP_DIV").(int)
			a := pop(&stack, "OP_DIV").(int)

			stack = append(stack, a/b)
		case OP_GT:
			b := pop(&stack, "OP_GT").(int)
			a := pop(&stack, "OP_GT").(int)

			if a > b {
				stack = append(stack, 1)
			} else {
				stack = append(stack, 0)
			}
		case OP_GT_EQ:
			b := pop(&stack, "OP_GT_EQ").(int)
			a := pop(&stack, "OP_GT_EQ").(int)

			if a >= b {
				stack = append(stack, 1)
			} else {
				stack = append(stack, 0)
			}
		case OP_LT:
			b := pop(&stack, "OP_LT").(int)
			a := pop(&stack, "OP_LT").(int)

			if a < b {
				stack = append(stack, 1)
			} else {
				stack = append(stack, 0)
			}
		case OP_LT_EQ:
			b := pop(&stack, "OP_LT_EQ").(int)
			a := pop(&stack, "OP_LT_EQ").(int)

			if a <= b {
				stack = append(stack, 1)
			} else {
				stack = append(stack, 0)
			}
		case OP_DIFF:
			b := pop(&stack, "OP_DIFF")
			a := pop(&stack, "OP_DIFF")

			if a != b {
				stack = append(stack, 1)
			} else {
				fmt.Println("Erro: Empty stack")
			}
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
			val := pop(&stack, "OP_PRINT")
			fmt.Println(val)
		case OP_NOP:
		case OP_JMP:
			destiny := int(code[ip])
			ip++
			ip = destiny
		case OP_JMP_IF:
			destiny := int(code[ip])
			ip++

			cond := pop(&stack, "OP_JMP_IF")

			if cond == 0 {
				ip = destiny
			}
		case OP_CALL:
			destiny := int(code[ip])
			ip++

			funcArgs := compile.GetFuncArgs(destiny)
			for i := len(funcArgs) - 1; i >= 0; i-- {
				val := pop(&stack, "OP_CALL")
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
			key := pop(&stack, "OP_ACCESS")
			target := pop(&stack, "OP_ACCESS")

			switch obj := target.(type) {
			case []interface{}: // array access
				idx, ok := key.(int)
				if !ok {
					panic(fmt.Sprintf("Array index must be int, got %T", key))
				}
				if idx < 0 || idx >= len(obj) {
					panic(fmt.Sprintf("Array index out of bounds: %d", idx))
				}
				push(&stack, obj[idx])
			case map[string]interface{}: // object access
				prop, ok := key.(string)
				if !ok {
					panic(fmt.Sprintf("Object property key must be string, got %T", key))
				}
				val, exists := obj[prop]
				if !exists {
					panic(fmt.Sprintf("Property '%s' not found in object", prop))
				}

				push(&stack, val)
			default:
				panic(fmt.Sprintf("OP_ACCESS: unsupported target type %T", target))
			}
		case OP_GET_PROPERTY:
			key := pop(&stack, "OP_GET_PROPERTY")
			target := pop(&stack, "OP_GET_PROPERTY")

			switch obj := target.(type) {
			case map[string]interface{}:
				prop, ok := key.(string)
				if !ok {
					panic(fmt.Sprintf("Expected property key as string, got %T", key))
				}

				val, exists := obj[prop]
				if !exists {
					panic(fmt.Sprintf("Property '%s' not found in object", prop))
				}

				push(&stack, val)
				// fmt.Printf("[DEBUG] OP_GET_PROPERTY: %s = %#v\n", prop, val)

			default:
				panic(fmt.Sprintf("OP_GET_PROPERTY: unsupported target type %T", target))
			}
		case OP_SET_PROPERTY:
			value := pop(&stack, "SET_PROPERTY")
			key := pop(&stack, "SET_PROPERTY")
			object := pop(&stack, "SET_PROPERTY")

			objMap, ok := object.(map[string]interface{})

			if !ok {
				panic(fmt.Sprintf("SET_PROPERTY expected map[string]interface{}, got %T", object))
			}

			keyStr, ok := key.(string)
			if !ok {
				panic(fmt.Sprintf("SET_PROPERTY expected string key, got %T", key))
			}

			objMap[keyStr] = value
			push(&stack, objMap)
		case OP_NULL:
			stack = append(stack, nil)
		case OP_LENGTH:
			arr := pop(&stack, "OP_LENGTH")

			switch v := arr.(type) {
			case string:
				stack = append(stack, len(v))
			case []interface{}:
				stack = append(stack, len(v))
			default:
				panic(fmt.Sprintf("OP_LENGTH: invalid type %T", v))
			}
		case OP_STORE:
			key := int(code[ip])
			ip++

			val := pop(&stack, "OP_STORE")

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
			val := pop(&stack, "OP_MSTORE")
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
		case OP_PUSH_OBJECT:
			push(&stack, make(map[string]interface{}))
		case OP_TRANSFER:
			fmt.Println("Transfer has been made :D")
		case OP_BALANCE_OF:
			fmt.Println("BalanceOf has been made :D")
		case OP_REQUIRE:
			fmt.Println("Require has been made :D")
		case OP_HALT:
			fmt.Println("End of program.")
			return
		default:
			fmt.Println("Opcode desconhecido:", op)
			return
		}
	}
}

func pop(stack *[]interface{}, context string) interface{} {
	if len(*stack) == 0 {
		panic(fmt.Sprintf("stack underflow while executing %s", context))
	}
	val := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]
	return val
}

func push(stack *[]interface{}, val interface{}) {
	*stack = append(*stack, val)
}
