package vm

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/peiblow/vvm/compiler"
)

// VM representa a máquina virtual
type VM struct {
	compiler  *compiler.Compiler
	stack     []interface{}
	callStack []int
	storage   map[int]interface{}
	memory    map[int]interface{}
	ip        int
}

// New cria uma nova instância da VM
func New(c *compiler.Compiler) *VM {
	return &VM{
		compiler:  c,
		stack:     []interface{}{},
		callStack: []int{},
		storage:   make(map[int]interface{}),
		memory:    make(map[int]interface{}),
		ip:        0,
	}
}

// Run executa o programa compilado
func (vm *VM) Run() {
	code := vm.compiler.Code

	for {
		op := code[vm.ip]
		vm.ip++

		switch op {
		case compiler.OP_CONST:
			vm.execConst(code)
		case compiler.OP_PUSH:
			vm.execPush(code)
		case compiler.OP_ADD:
			vm.execAdd()
		case compiler.OP_SUB:
			vm.execSub()
		case compiler.OP_MUL:
			vm.execMul()
		case compiler.OP_DIV:
			vm.execDiv()
		case compiler.OP_GT:
			vm.execGt()
		case compiler.OP_GT_EQ:
			vm.execGtEq()
		case compiler.OP_LT:
			vm.execLt()
		case compiler.OP_LT_EQ:
			vm.execLtEq()
		case compiler.OP_EQ:
			vm.execEq()
		case compiler.OP_DIFF:
			vm.execDiff()
		case compiler.OP_SWAP:
			vm.execSwap()
		case compiler.OP_PRINT:
			vm.execPrint()
		case compiler.OP_NOP:
			// No operation
		case compiler.OP_JMP:
			vm.execJmp(code)
		case compiler.OP_JMP_IF:
			vm.execJmpIf(code)
		case compiler.OP_CALL:
			vm.execCall(code)
		case compiler.OP_RET:
			vm.execRet()
		case compiler.OP_ACCESS:
			vm.execAccess()
		case compiler.OP_GET_PROPERTY:
			vm.execGetProperty()
		case compiler.OP_SET_PROPERTY:
			vm.execSetProperty()
		case compiler.OP_NULL:
			vm.push(nil)
		case compiler.OP_LENGTH:
			vm.execLength()
		case compiler.OP_STORE:
			vm.execStore(code)
		case compiler.OP_SLOAD:
			vm.execSload(code)
		case compiler.OP_REGISTRY_DECLARE:
			vm.execRegistry(code)
		case compiler.OP_REGISTRY_GET:
			vm.execRegistryGet(code)
		case compiler.OP_POLICY_DECLARE:
			vm.execPolicyDeclare(code)
		case compiler.OP_REQUIRE:
			vm.execRequire()
		case compiler.OP_ERR:
			vm.execErr()
		case compiler.OP_DELETE:
			vm.execDelete(code)
		case compiler.OP_PUSH_OBJECT:
			vm.push(make(map[string]interface{}))
		case compiler.OP_TRANSFER:
			fmt.Println("Transfer has been made :D")
		case compiler.OP_BALANCE_OF:
			fmt.Println("BalanceOf has been made :D")
		case compiler.OP_HALT:
			fmt.Println("End of program.")
			return
		default:
			fmt.Printf("Opcode desconhecido: 0x%02X\n", op)
			return
		}
	}
}

// Stack operations

func (vm *VM) push(val interface{}) {
	vm.stack = append(vm.stack, val)
}

func (vm *VM) pop(context string) interface{} {
	if len(vm.stack) == 0 {
		panic(fmt.Sprintf("stack underflow while executing %s", context))
	}
	val := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return val
}

// Instruction implementations

func (vm *VM) execConst(code []byte) {
	idx := int(code[vm.ip])
	vm.ip++
	val := vm.compiler.ConstPool[idx]
	vm.push(val)
}

func (vm *VM) execPush(code []byte) {
	val := int(code[vm.ip])
	vm.ip++
	vm.push(val)
}

func (vm *VM) execAdd() {
	if len(vm.stack) < 2 {
		panic("stack underflow on ADD")
	}

	a := vm.pop("OP_ADD")
	b := vm.pop("OP_ADD")

	switch av := a.(type) {
	case int:
		switch bv := b.(type) {
		case int:
			vm.push(av + bv)
		case string:
			vm.push(strconv.Itoa(av) + bv)
		default:
			panic("[INT] unsupported ADD type")
		}
	case string:
		switch bv := b.(type) {
		case int:
			vm.push(av + strconv.Itoa(bv))
		case string:
			vm.push(av + bv)
		case float64:
			vm.push(av + strconv.FormatFloat(bv, 'f', 0, 64))
		default:
			fmt.Println(reflect.TypeOf(bv))
			panic("[STR] unsupported ADD type")
		}
	default:
		panic("[DFT] unsupported ADD type")
	}
}

func (vm *VM) execSub() {
	b := vm.pop("OP_SUB").(int)
	a := vm.pop("OP_SUB").(int)
	vm.push(a - b)
}

func (vm *VM) execMul() {
	b := vm.pop("OP_MUL").(int)
	a := vm.pop("OP_MUL").(int)
	vm.push(a * b)
}

func (vm *VM) execDiv() {
	b := vm.pop("OP_DIV").(int)
	a := vm.pop("OP_DIV").(int)
	vm.push(a / b)
}

func (vm *VM) execGt() {
	b := vm.pop("OP_GT").(int)
	a := vm.pop("OP_GT").(int)
	if a > b {
		vm.push(1)
	} else {
		vm.push(0)
	}
}

func (vm *VM) execGtEq() {
	b := asNumber(vm.pop("OP_GT_EQ"))
	a := asNumber(vm.pop("OP_GT_EQ"))

	if a >= b {
		vm.push(1)
	} else {
		vm.push(0)
	}
}

func (vm *VM) execLt() {
	b := vm.pop("OP_LT").(int)
	a := vm.pop("OP_LT").(int)
	if a < b {
		vm.push(1)
	} else {
		vm.push(0)
	}
}

func (vm *VM) execLtEq() {
	b := vm.pop("OP_LT_EQ").(int)
	a := vm.pop("OP_LT_EQ").(int)
	if a <= b {
		vm.push(1)
	} else {
		vm.push(0)
	}
}

func asNumber(v interface{}) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case float64:
		return n
	default:
		panic("expected numeric value")
	}
}

func (vm *VM) execEq() {
	b := vm.pop("OP_EQ")
	a := vm.pop("OP_EQ")
	if a == b {
		vm.push(1)
	} else {
		vm.push(0)
	}
}

func (vm *VM) execDiff() {
	b := vm.pop("OP_DIFF")
	a := vm.pop("OP_DIFF")
	if a != b {
		vm.push(1)
	} else {
		vm.push(0)
	}
}

func (vm *VM) execSwap() {
	a := vm.stack[len(vm.stack)-1]
	b := vm.stack[len(vm.stack)-2]
	vm.stack[len(vm.stack)-1] = b
	vm.stack[len(vm.stack)-2] = a
}

func (vm *VM) execPrint() {
	if len(vm.stack) == 0 {
		fmt.Println("Warning: OP_PRINT with empty stack, ignoring")
		return
	}

	val := vm.pop("OP_PRINT")
	fmt.Println(val)
}

func (vm *VM) execJmp(code []byte) {
	destiny := int(code[vm.ip])
	vm.ip = destiny
}

func (vm *VM) execJmpIf(code []byte) {
	destiny := int(code[vm.ip])
	vm.ip++
	cond := vm.pop("OP_JMP_IF")
	if cond == 0 {
		vm.ip = destiny
	}
}

func (vm *VM) execCall(code []byte) {
	destiny := int(code[vm.ip])
	vm.ip++

	funcArgs := vm.compiler.GetFuncArgs(destiny)
	for i := len(funcArgs) - 1; i >= 0; i-- {
		val := vm.pop("OP_CALL")
		slot := funcArgs[i]
		vm.storage[slot] = val
	}

	vm.callStack = append(vm.callStack, vm.ip)
	vm.ip = destiny
}

func (vm *VM) execRet() {
	if len(vm.callStack) == 0 {
		return
	}
	returnAddr := vm.callStack[len(vm.callStack)-1]
	vm.callStack = vm.callStack[:len(vm.callStack)-1]
	vm.ip = returnAddr
}

func (vm *VM) execAccess() {
	key := vm.pop("OP_ACCESS")
	target := vm.pop("OP_ACCESS")

	switch obj := target.(type) {
	case []interface{}:
		idx, ok := key.(int)
		if !ok {
			panic(fmt.Sprintf("Array index must be int, got %T", key))
		}
		if idx < 0 || idx >= len(obj) {
			panic(fmt.Sprintf("Array index out of bounds: %d", idx))
		}
		vm.push(obj[idx])
	case map[string]interface{}:
		prop, ok := key.(string)
		if !ok {
			panic(fmt.Sprintf("Object property key must be string, got %T", key))
		}
		val, exists := obj[prop]
		if !exists {
			panic(fmt.Sprintf("Property '%s' not found in object", prop))
		}
		vm.push(val)
	default:
		panic(fmt.Sprintf("OP_ACCESS: unsupported target type %T", target))
	}
}

func (vm *VM) execGetProperty() {
	key := vm.pop("OP_GET_PROPERTY")
	target := vm.pop("OP_GET_PROPERTY")

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
		vm.push(val)
	default:
		panic(fmt.Sprintf("OP_GET_PROPERTY: unsupported target type %T", target))
	}
}

func (vm *VM) execSetProperty() {
	value := vm.pop("SET_PROPERTY")
	key := vm.pop("SET_PROPERTY")
	object := vm.pop("SET_PROPERTY")

	objMap, ok := object.(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("SET_PROPERTY expected map[string]interface{}, got %T", object))
	}

	keyStr, ok := key.(string)
	if !ok {
		panic(fmt.Sprintf("SET_PROPERTY expected string key, got %T", key))
	}

	objMap[keyStr] = value
	vm.push(objMap)
}

func (vm *VM) execLength() {
	arr := vm.pop("OP_LENGTH")

	switch v := arr.(type) {
	case string:
		vm.push(len(v))
	case []interface{}:
		vm.push(len(v))
	default:
		panic(fmt.Sprintf("OP_LENGTH: invalid type %T", v))
	}
}

func (vm *VM) execStore(code []byte) {
	key := int(code[vm.ip])
	vm.ip++
	val := vm.pop("OP_STORE")
	vm.storage[key] = val
}

func (vm *VM) execSload(code []byte) {
	key := int(code[vm.ip])
	vm.ip++
	val, ok := vm.storage[key]
	if !ok {
		val = 0
	}
	vm.push(val)
}

func (vm *VM) execRegistry(code []byte) {
	purpose := vm.pop("OP_REGISTRY_DECLARE")
	vm.ip++

	owner := vm.pop("OP_REGISTRY_DECLARE")
	vm.ip++

	version := vm.pop("OP_REGISTRY_DECLARE")
	vm.ip++

	name := vm.pop("OP_REGISTRY_DECLARE")
	vm.ip++

	kind := vm.pop("OP_REGISTRY_DECLARE")
	vm.ip++

	key := len(vm.storage) + 1
	vm.storage[key] = map[string]interface{}{
		"kind":    kind,
		"name":    name,
		"version": version,
		"owner":   owner,
		"purpose": purpose,
	}
}

func (vm *VM) execRegistryGet(code []byte) {
	identifierIdx := int(code[vm.ip])
	vm.ip++

	val, ok := vm.storage[identifierIdx]
	if !ok {
		panic(fmt.Sprintf("Registry with identifier %d not found", identifierIdx))
	}

	// fmt.Println("Registry Get:", val)
	vm.push(val)
}

func (vm *VM) execPolicyDeclare(code []byte) {
	// For simplicity, we just pop the identifier and store an empty policy
	identifierIdx := int(code[vm.ip])
	vm.ip++

	key := len(vm.storage) + 1
	vm.storage[key] = map[string]interface{}{
		"identifier": identifierIdx,
		"rules":      map[string]interface{}{},
	}
}

func (vm *VM) execDelete(code []byte) {
	vm.ip++
	key := int(code[vm.ip])
	delete(vm.storage, key)
}

func (vm *VM) execRequire() {
	condition := vm.pop("OP_REQUIRE")
	condInt, ok := condition.(int)
	if !ok || condInt == 0 {
		panic("Require condition failed")
	}
}

func (vm *VM) execErr() {
	message := vm.pop("OP_ERR")
	panic(fmt.Sprintf("Error raised: %v", message))
}
