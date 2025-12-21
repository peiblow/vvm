package compiler

import (
	"github.com/peiblow/vvm/ast"
)

type FunctionMeta struct {
	Addr int
	Args []int
}

type Compiler struct {
	Code         []byte
	Symbols      map[string]int
	ConstPool    []interface{}
	Functions    map[string]FunctionMeta
	FunctionName map[int]string
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
		NextSlot:     0,
	}
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

// patchJump corrige um salto em uma posição específica
func (c *Compiler) patchJump(pos int, target int) {
	c.Code[pos] = byte(target)
}
