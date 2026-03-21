package vm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/peiblow/vvm/compiler"
)

type VM struct {
	compiler  *compiler.Compiler
	stack     []interface{}
	callStack []int
	storage   map[int]interface{}
	memory    map[int]interface{}
	ip        int
	errors    []error
	journal   []JournalEvent
}

type JournalEvent struct {
	Type      string
	Payload   map[string]interface{}
	Hash      string
	Timestamp int64
}

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

type ExecutionResult struct {
	Success bool
	Journal []JournalEvent
	Error   error
}

func NewFromArtifact(artifact *compiler.ContractArtifact) *VM {
	cmpl := &compiler.Compiler{
		Code:         artifact.Bytecode,
		ConstPool:    artifact.ConstPool,
		Functions:    artifact.Functions,
		FunctionName: artifact.FunctionName,
		Types:        artifact.Types,
	}
	vm := New(cmpl)

	// Deep copy InitStorage to ensure execution doesn't modify the artifact
	if artifact.InitStorage != nil {
		for k, v := range artifact.InitStorage {
			vm.storage[k] = deepCopy(v)
		}
	}
	return vm
}

func deepCopy(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case map[string]interface{}:
		copied := make(map[string]interface{}, len(val))
		for k, v := range val {
			copied[k] = deepCopy(v)
		}
		return copied
	case []interface{}:
		copied := make([]interface{}, len(val))
		for i, v := range val {
			copied[i] = deepCopy(v)
		}
		return copied
	default:
		return val
	}
}

func (vm *VM) GetStorage() map[int]interface{} {
	storageCopy := make(map[int]interface{})
	for k, v := range vm.storage {
		storageCopy[k] = deepCopy(v)
	}
	return storageCopy
}

func (vm *VM) Run() ExecutionResult {
	return vm.execute()
}

// RunFunction executes a specific function by name with the given arguments
func (vm *VM) RunFunction(funcName string, args ...interface{}) ExecutionResult {
	funcMeta, exists := vm.compiler.Functions[funcName]
	if !exists {
		return ExecutionResult{
			Success: false,
			Journal: vm.journal,
			Error:   fmt.Errorf("function '%s' not found in contract", funcName),
		}
	}

	if len(args) != len(funcMeta.Args) {
		return ExecutionResult{
			Success: false,
			Journal: vm.journal,
			Error:   fmt.Errorf("function '%s' expects %d argument(s), got %d", funcName, len(funcMeta.Args), len(args)),
		}
	}

	for i, arg := range args {
		slot := funcMeta.Args[i]
		vm.storage[slot] = arg
	}

	haltAddr := len(vm.compiler.Code) - 1
	vm.callStack = append(vm.callStack, haltAddr)

	vm.ip = funcMeta.Addr

	vmResult := vm.execute()

	if len(vm.errors) > 0 {
		return ExecutionResult{
			Success: false,
			Journal: vm.journal,
			Error:   fmt.Errorf("execution aborted due to previous errors: %v", vm.errors),
		}
	} else {
		return vmResult
	}

	// return vm.execute()
}

func (vm *VM) execute() ExecutionResult {
	code := vm.compiler.Code

	for {
		op := code[vm.ip]
		vm.ip++

		if len(vm.errors) > 0 {
			fmt.Printf("Execution halted due to errors: %v\n", vm.errors)
			return ExecutionResult{
				Success: false,
				Journal: vm.journal,
				Error:   fmt.Errorf("execution aborted due to previous errors: %v", vm.errors),
			}
		}

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
		case compiler.OP_DUP:
			vm.execDup()
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
		case compiler.OP_AGENT_DECLARE:
			vm.execAgentDeclare()
		case compiler.OP_AGENT_GET:
			vm.execAgentGet(code)
		case compiler.OP_AGENT_VALIDATE:
			vm.execAgentValidate()
		case compiler.OP_POLICY_DECLARE:
			vm.execPolicyDeclare(code)
		case compiler.OP_TYPE_DECLARE:
			vm.execTypeDeclare(code)
		case compiler.OP_REQUIRE:
			vm.execRequire()
		case compiler.OP_EMIT:
			vm.execEmitEvent()
		case compiler.OP_GET_ENV:
			vm.execGetEnv()
		case compiler.OP_NONCE:
			vm.execNonce()
		case compiler.OP_HASH:
			vm.execHash()
		case compiler.OP_ERR:
			vm.execErr()
		case compiler.OP_DELETE:
			vm.execDelete(code)
		case compiler.OP_PUSH_OBJECT:
			vm.push(make(map[string]interface{}))
		case compiler.OP_HALT:
			return ExecutionResult{
				Success: true,
				Journal: vm.journal,
				Error:   nil,
			}
		default:
			return ExecutionResult{
				Success: false,
				Journal: vm.journal,
				Error:   fmt.Errorf("unknown opcode: 0x%02X", op),
			}
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
	b := asNumber(vm.pop("OP_SUB"))
	a := asNumber(vm.pop("OP_SUB"))
	vm.push(a - b)
}

func (vm *VM) execMul() {
	b := asNumber(vm.pop("OP_MUL"))
	a := asNumber(vm.pop("OP_MUL"))
	vm.push(a * b)
}

func (vm *VM) execDiv() {
	b := asNumber(vm.pop("OP_DIV"))
	a := asNumber(vm.pop("OP_DIV"))
	vm.push(a / b)
}

func (vm *VM) execGt() {
	b := asNumber(vm.pop("OP_GT"))
	a := asNumber(vm.pop("OP_GT"))
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
	b := asNumber(vm.pop("OP_LT"))
	a := asNumber(vm.pop("OP_LT"))
	if a < b {
		vm.push(1)
	} else {
		vm.push(0)
	}
}

func (vm *VM) execLtEq() {
	b := asNumber(vm.pop("OP_LT_EQ"))
	a := asNumber(vm.pop("OP_LT_EQ"))
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

func (vm *VM) execDup() {
	if len(vm.stack) == 0 {
		panic("OP_DUP: stack underflow")
	}
	val := vm.stack[len(vm.stack)-1]
	vm.push(val)
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
	high := int(code[vm.ip])
	low := int(code[vm.ip+1])
	destiny := (high << 8) | low
	vm.ip = destiny
}

func (vm *VM) execJmpIf(code []byte) {
	high := int(code[vm.ip])
	low := int(code[vm.ip+1])
	destiny := (high << 8) | low
	vm.ip += 2
	cond := vm.pop("OP_JMP_IF")
	if cond == 0 {
		vm.ip = destiny
	}
}

func (vm *VM) execCall(code []byte) {
	// Read 2-byte address (high byte, low byte)
	high := int(code[vm.ip])
	low := int(code[vm.ip+1])
	destiny := (high << 8) | low
	vm.ip += 2

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

func (vm *VM) execAgentDeclare() {
	purpose := vm.pop("OP_AGENT_DECLARE")
	vm.ip++

	owner := vm.pop("OP_AGENT_DECLARE")
	vm.ip++

	version := vm.pop("OP_AGENT_DECLARE")
	vm.ip++

	name := vm.pop("OP_AGENT_DECLARE")
	vm.ip++

	// Generate hash from registry data
	hashInput := fmt.Sprintf("%v:%v:%v:%v:%v", name, version, owner, purpose, time.Now().UnixMilli())
	hashBytes := sha256.Sum256([]byte(hashInput))
	hash := "0x" + hex.EncodeToString(hashBytes[:])

	fmt.Printf("Agent '%v' created with hash: %s\n", name, hash)

	key := len(vm.storage) + 1
	vm.storage[key] = map[string]interface{}{
		"hash":    hash,
		"name":    name,
		"version": version,
		"owner":   owner,
		"purpose": purpose,
	}
	fmt.Println("Agent stored at key:", vm.storage)
}

func (vm *VM) execAgentGet(code []byte) {
	identifierIdx := int(code[vm.ip])
	vm.ip++

	// Pop the identifier from stack
	identifier := vm.pop("OP_REGISTRY_GET")
	identifierStr := extractValue(identifier)

	// Find the registry with the same name
	var registry map[string]interface{}
	var registryFound bool
	for _, val := range vm.storage {
		if reg, ok := val.(map[string]interface{}); ok {
			if regName, exists := reg["name"]; exists {
				regNameStr := extractValue(regName)
				if regNameStr == identifierStr {
					registry = reg
					registryFound = true
					break
				}
			}
		}
	}

	if !registryFound {
		panic(fmt.Sprintf("Registry '%v' not found (identifier: %d)", identifierStr, identifierIdx))
	}

	vm.push(registry)
}

func (vm *VM) execAgentValidate() {
	agentOwner := vm.pop("AGENT_VALIDATE")
	agentVersion := vm.pop("AGENT_VALIDATE")
	registry := vm.pop("AGENT_VALIDATE").(map[string]interface{})

	agentOwnerStr := extractValue(agentOwner)
	agentVersionStr := extractValue(agentVersion)
	agentNameStr := extractValue(registry["name"])

	// Validate version
	registryVersion := extractValue(registry["version"])
	if registryVersion != agentVersionStr {
		panic(fmt.Sprintf("Agent validation failed: Version not found for '%v'\n  Registry version: %s\n  Agent version: %s", agentNameStr, registryVersion, agentVersionStr))
	}

	// Validate owner
	registryOwner := extractValue(registry["owner"])
	if registryOwner != agentOwnerStr {
		panic(fmt.Sprintf("Agent validation failed: Owner mismatch for '%v'\n  Expected: %s\n  Got: %s", agentNameStr, registryOwner, agentOwnerStr))
	}

	agentHashStr := extractValue(registry["hash"])

	fmt.Printf("Agent '%v' validated successfully (hash: %s..., owner: %s, version: %v)\n", agentNameStr, agentHashStr[:18], agentOwnerStr, agentVersionStr)
	agentData := map[string]interface{}{
		"name":    agentNameStr,
		"hash":    agentHashStr,
		"version": agentVersionStr,
		"owner":   agentOwnerStr,
	}
	vm.push(agentData)
}

func (vm *VM) execPolicyDeclare(code []byte) {
	vm.ip++

	policyObj := vm.pop("OP_POLICY_DECLARE")
	vm.pop("OP_POLICY_DECLARE")
	vm.push(policyObj)
}

func (vm *VM) execTypeDeclare(code []byte) {
	vm.ip++
	typeObj := vm.pop("OP_TYPE_DECLARE")
	vm.pop("OP_TYPE_DECLARE")
	vm.push(typeObj)
}

func (vm *VM) execGetEnv() {
	vm.ip++
	variableName := vm.pop("OP_GET_ENV")
	variableNameStr := extractValue(variableName)

	value, ok := os.LookupEnv(variableNameStr)
	if !ok {
		panic(fmt.Sprintf("Environment variable '%s' not found", variableNameStr))
	}

	fmt.Printf("Environment variable '%s' accessed with value: %s\n", variableNameStr, value)
	vm.push(value)
}

func (vm *VM) execNonce() {
	sizeVal := vm.pop("OP_NONCE")
	size, ok := sizeVal.(int)
	if !ok {
		panic(fmt.Sprintf("OP_NONCE expected int size, got %T", sizeVal))
	}

	nonceBytes := make([]byte, size)
	for i := 0; i < size; i++ {
		nonceBytes[i] = byte(time.Now().UnixNano() >> (i * 8))
	}

	nonceHex := "0x" + hex.EncodeToString(nonceBytes)
	fmt.Printf("Generated nonce of size %d: %s\n", size, nonceHex)
	vm.push(nonceHex)
}

func (vm *VM) execHash() {
	dataVal := vm.pop("OP_HASH")
	hashTypeVal := vm.pop("OP_HASH")

	hashTypeStr := extractValue(hashTypeVal)
	dataStr := extractValue(dataVal)

	var hashBytes []byte
	switch hashTypeStr {
	case "SHA256":
		sum := sha256.Sum256([]byte(dataStr))
		hashBytes = sum[:]
	default:
		panic(fmt.Sprintf("Unsupported hash type: %s", hashTypeStr))
	}

	hashHex := "0x" + hex.EncodeToString(hashBytes)
	fmt.Printf("Hashed data using %s: %s\n", hashTypeStr, hashHex)
	vm.push(hashHex)
}

func (vm *VM) execEmitEvent() {
	vm.ip++
	eventPayload := vm.pop("OP_EMIT_EVENT")
	eventType := vm.pop("OP_EMIT_EVENT")

	eventData := fmt.Sprintf("%v:%v", eventType, eventPayload)
	hashBytes := sha256.Sum256([]byte(eventData))
	hash := "0x" + hex.EncodeToString(hashBytes[:])

	journalEvent := JournalEvent{
		Type:    extractValue(eventType),
		Payload: map[string]interface{}{"data": eventPayload},
		Hash:    hash,
	}

	vm.journal = append(vm.journal, journalEvent)
	fmt.Printf("Event emitted: Type=%s, Hash=%s\n", journalEvent.Type, journalEvent.Hash)
}

// extractValue extracts the actual value from AST expressions or returns string representation
func extractValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%.0f", val)
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Struct {
			valueField := rv.FieldByName("Value")
			if valueField.IsValid() {
				return fmt.Sprintf("%v", valueField.Interface())
			}
		}
		return fmt.Sprintf("%v", v)
	}
}

func (vm *VM) execDelete(code []byte) {
	vm.ip++
	key := int(code[vm.ip])
	delete(vm.storage, key)
}

func (vm *VM) execRequire() {
	condition := vm.pop("OP_REQUIRE")

	message := vm.pop("OP_REQUIRE")
	messageStr, ok := message.(string)
	if !ok {
		messageStr = fmt.Sprintf("%v", message)
	}

	condInt, ok := condition.(int)
	if !ok || condInt == 0 {
		vm.errors = append(vm.errors, fmt.Errorf("require failed: %s", messageStr))
	}
}

func (vm *VM) execErr() {
	message := vm.pop("OP_ERR")
	vm.errors = append(vm.errors, fmt.Errorf("execution error: %v", message))
}
