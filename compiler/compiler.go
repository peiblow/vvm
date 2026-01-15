package compiler

import (
	"fmt"

	"github.com/peiblow/vvm/ast"
)

// ArgMeta holds argument metadata including name and type
type ArgMeta struct {
	Slot     int
	TypeName string
}

type FunctionMeta struct {
	Addr    int
	Args    []int
	ArgMeta []ArgMeta // Argument metadata with types
}

// TypeMeta holds type declaration metadata
type TypeMeta struct {
	Fields map[string]string // field name -> type name
}

type Compiler struct {
	Code         []byte
	Symbols      map[string]int
	ConstPool    []interface{}
	Functions    map[string]FunctionMeta
	FunctionName map[int]string
	Types        map[string]TypeMeta // Registered types
	NextSlot     int
	isInFunction bool
}

func New() *Compiler {
	return &Compiler{
		Code:         []byte{},
		Symbols:      make(map[string]int),
		ConstPool:    make([]interface{}, 0),
		Functions:    make(map[string]FunctionMeta),
		FunctionName: make(map[int]string),
		Types:        make(map[string]TypeMeta),
		NextSlot:     0,
	}
}

// GetExpectedArgType returns the expected type for a function argument
func (c *Compiler) GetExpectedArgType(funcName string, argIndex int) string {
	if meta, ok := c.Functions[funcName]; ok {
		if argIndex < len(meta.ArgMeta) {
			return meta.ArgMeta[argIndex].TypeName
		}
	}
	return ""
}

// GetActualType infers the type of an expression
func (c *Compiler) GetActualType(expr ast.Expr) string {
	switch e := expr.(type) {
	case ast.NumberExpr:
		if e.Value == float64(int(e.Value)) {
			return "Int"
		}
		return "Float"
	case ast.StringExpr:
		return "String"
	case ast.SymbolExpr:
		// Check if it's a known type instance
		if _, ok := c.Types[e.Value]; ok {
			return e.Value
		}
		// Could be a variable, return unknown for now
		return "Unknown"
	case ast.ObjectAssignmentExpr:
		if sym, ok := e.Name.(ast.SymbolExpr); ok {
			return sym.Value
		}
		return "Object"
	default:
		return "Unknown"
	}
}

// TypesAreCompatible checks if actualType is compatible with expectedType
func (c *Compiler) TypesAreCompatible(expectedType, actualType string) bool {
	// If types match exactly, they're compatible
	if expectedType == actualType {
		return true
	}

	// If actual type is unknown, we can't validate at compile time
	if actualType == "Unknown" {
		return true
	}

	// Object literals can be compatible with custom types (structural typing)
	if actualType == "Object" {
		// If expected type is a declared custom type, we'll do structural validation
		if _, isCustomType := c.Types[expectedType]; isCustomType {
			return true // Structural validation happens in ValidateObjectAgainstType
		}
	}

	// Primitive type aliases
	primitiveAliases := map[string][]string{
		"Int":     {"UInt", "Float", "Number"},
		"UInt":    {"Int", "Float", "Number"},
		"Float":   {"Int", "UInt", "Number"},
		"Number":  {"Int", "UInt", "Float"},
		"String":  {},
		"Address": {"String", "Int", "UInt"}, // Addresses can be hex numbers (parsed as Int)
		"Proof":   {"String"},
	}

	if aliases, ok := primitiveAliases[expectedType]; ok {
		for _, alias := range aliases {
			if alias == actualType {
				return true
			}
		}
	}

	// Check if expectedType is a declared type and actualType is a primitive
	if _, isCustomType := c.Types[expectedType]; isCustomType {
		// Custom types are not compatible with primitives
		switch actualType {
		case "Int", "UInt", "Float", "Number", "String":
			return false
		}
	}

	return false
}

// ValidateObjectAgainstType validates that an object literal matches a custom type's structure
func (c *Compiler) ValidateObjectAgainstType(obj ast.ObjectAssignmentExpr, typeName string) error {
	typeMeta, exists := c.Types[typeName]
	if !exists {
		return fmt.Errorf("unknown type '%s'", typeName)
	}

	// Build a map of fields provided in the object literal
	providedFields := make(map[string]ast.Expr)
	for _, field := range obj.Fields {
		if key, ok := field.Key.(ast.SymbolExpr); ok {
			providedFields[key.Value] = field.Value
		}
	}

	// Check that all required fields are present
	for fieldName, expectedFieldType := range typeMeta.Fields {
		providedValue, exists := providedFields[fieldName]
		if !exists {
			return fmt.Errorf("missing field '%s' of type '%s' in object literal for type '%s'",
				fieldName, expectedFieldType, typeName)
		}

		// Validate field type
		actualFieldType := c.GetActualType(providedValue)
		if !c.TypesAreCompatible(expectedFieldType, actualFieldType) {
			return fmt.Errorf("field '%s' has type '%s', expected '%s' for type '%s'",
				fieldName, actualFieldType, expectedFieldType, typeName)
		}
	}

	return nil
}

// ValidateFunctionCall validates argument types for a function call
func (c *Compiler) ValidateFunctionCall(funcName string, args []ast.Expr) error {
	funcMeta, exists := c.Functions[funcName]
	if !exists {
		// Built-in function or function not yet defined, skip validation
		return nil
	}

	if len(args) != len(funcMeta.ArgMeta) {
		return fmt.Errorf("function '%s' expects %d argument(s), got %d",
			funcName, len(funcMeta.ArgMeta), len(args))
	}

	for i, arg := range args {
		expectedType := funcMeta.ArgMeta[i].TypeName
		actualType := c.GetActualType(arg)

		// Special handling for object literals passed to custom types
		if actualType == "Object" {
			if _, isCustomType := c.Types[expectedType]; isCustomType {
				if objExpr, ok := arg.(ast.ObjectAssignmentExpr); ok {
					if err := c.ValidateObjectAgainstType(objExpr, expectedType); err != nil {
						return err
					}
					continue // Validation passed
				}
			}
		}

		if !c.TypesAreCompatible(expectedType, actualType) {
			return fmt.Errorf("type mismatch in argument %d of function '%s': expected '%s', got '%s'",
				i+1, funcName, expectedType, actualType)
		}
	}

	return nil
}

func (c *Compiler) emit(opcodes ...byte) {
	c.Code = append(c.Code, opcodes...)
}

// addConst adiciona uma constante ao pool e retorna seu índice
func (c *Compiler) addConst(val interface{}) byte {
	idx := len(c.ConstPool)
	c.ConstPool = append(c.ConstPool, val)
	return byte(idx)
}

// procuta na constPool e retorna o índice se encontrado, ou -1 se não encontrado
func (c *Compiler) findConst(val interface{}) byte {
	for i, v := range c.ConstPool {
		if v == val {
			return byte(i)
		}
	}
	return byte(255)
}

// allocSlot aloca um slot para uma variável
func (c *Compiler) allocSlot(name string) int {
	slot := c.NextSlot
	c.Symbols[name] = slot
	c.NextSlot++
	return slot
}

// getSlot obtém o slot de uma variável, alocando se necessário
func (c *Compiler) getSlot(name string) int {
	slot, ok := c.Symbols[name]
	if !ok {
		slot = c.allocSlot(name)
	}
	return slot
}

// GetFuncArgs retorna os argumentos de uma função pelo endereço
func (c *Compiler) GetFuncArgs(addr int) []int {
	funcName := c.FunctionName[addr]
	return c.Functions[funcName].Args
}

// CompileBlock compila um bloco de código principal (adiciona HALT no final)
func (c *Compiler) CompileBlock(block ast.BlockStmt) {
	for _, stmt := range block.Body {
		c.compileStmt(stmt)
	}
	c.emit(OP_HALT)
}

// compileBlock compila um bloco interno (sem HALT)
func (c *Compiler) compileBlock(block ast.BlockStmt) {
	for _, stmt := range block.Body {
		c.compileStmt(stmt)
	}
}

// currentPos retorna a posição atual no código
func (c *Compiler) currentPos() int {
	return len(c.Code)
}

// patchJump corrige um salto em uma posição específica (2 bytes: high, low)
func (c *Compiler) patchJump(pos int, target int) {
	c.Code[pos] = byte(target >> 8)     // high byte
	c.Code[pos+1] = byte(target & 0xFF) // low byte
}
