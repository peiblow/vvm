package compiler

import (
	"fmt"

	"github.com/peiblow/vvm/ast"
)

// compileStmt compila um statement
func (c *Compiler) compileStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case ast.ContractStmt:
		c.compileContract(s)
	case ast.ExpressionStmt:
		c.compileExpr(s.Expression)
	case ast.VarDeclStmt:
		c.compileVarDecl(s)
	case ast.ReturnStmt:
		c.compileReturn(s)
	case ast.FuncStmt:
		c.compileFunc(s)
	case ast.IfStmt:
		c.compileIf(s)
	case ast.ForStmt:
		c.compileFor(s)
	case ast.WhileStmt:
		c.compileWhile(s)

	default:
		fmt.Printf("Statement não reconhecido: %T\n", s)
	}
}

// compileContract compila um contrato
func (c *Compiler) compileContract(s ast.ContractStmt) {
	for _, stmt := range s.Body {
		c.compileStmt(stmt)
	}
}

// compileVarDecl compila uma declaração de variável
func (c *Compiler) compileVarDecl(s ast.VarDeclStmt) {
	if s.AssignedValue == nil {
		return
	}
	c.compileExpr(s.AssignedValue)
	slot := c.allocSlot(s.Identifier)
	c.emit(OP_STORE, byte(slot))
}

// compileReturn compila um statement de retorno
func (c *Compiler) compileReturn(s ast.ReturnStmt) {
	c.compileExpr(s.Value)
	if c.isInFunction {
		c.emit(OP_RET)
	} else {
		c.emit(OP_PRINT)
	}
}

// compileFunc compila uma declaração de função
func (c *Compiler) compileFunc(s ast.FuncStmt) {
	funcName := c.extractFuncName(s.Name)

	// Salva estado anterior
	prevInFunction := c.isInFunction
	c.isInFunction = true

	// Emite salto para pular o corpo da função
	skipFuncPos := c.currentPos()
	c.emit(OP_JMP, 0)

	// Registra metadados da função
	funcMeta := FunctionMeta{
		Addr: c.currentPos(),
		Args: []int{},
	}

	// Compila argumentos
	funcMeta.Args = c.compileFuncArgs(s.Arguments)

	c.Functions[funcName] = funcMeta
	c.FunctionName[c.currentPos()] = funcName

	// Compila corpo da função
	c.compileFuncBody(s.Body)

	// Emite retorno baseado no tipo
	c.emitFuncReturn(s.ReturnType)

	// Corrige o salto
	c.patchJump(skipFuncPos+1, c.currentPos())
	c.isInFunction = prevInFunction
}

// extractFuncName extrai o nome de uma função
func (c *Compiler) extractFuncName(name ast.Expr) string {
	switch n := name.(type) {
	case ast.SymbolExpr:
		return n.Value
	case ast.ExpressionStmt:
		if sym, ok := n.Expression.(ast.SymbolExpr); ok {
			return sym.Value
		}
	}
	panic("Tipo de expressão não suportado para nome de função")
}

// compileFuncArgs compila os argumentos de uma função
func (c *Compiler) compileFuncArgs(args ast.Stmt) []int {
	var slots []int
	block, ok := args.(ast.BlockStmt)
	if !ok {
		return slots
	}

	for _, arg := range block.Body {
		exprStmt, ok := arg.(ast.ExpressionStmt)
		if !ok {
			continue
		}
		sym, ok := exprStmt.Expression.(ast.SymbolExpr)
		if !ok {
			continue
		}
		slot := c.allocSlot(sym.Value)
		slots = append(slots, slot)
	}
	return slots
}

// compileFuncBody compila o corpo de uma função
func (c *Compiler) compileFuncBody(body ast.Stmt) {
	if block, ok := body.(ast.BlockStmt); ok {
		c.compileBlock(block)
	} else {
		c.compileStmt(body)
	}
}

// emitFuncReturn emite o retorno apropriado baseado no tipo
func (c *Compiler) emitFuncReturn(returnType ast.Type) {
	// Apenas emite RET para funções void
	// Funções não-void já têm RET emitido pelo statement return
	if t, ok := returnType.(ast.SymbolType); ok && t.Name == "void" {
		c.emit(OP_RET)
	}
}

// compileIf compila um statement if
func (c *Compiler) compileIf(s ast.IfStmt) {
	c.compileExpr(s.Condition)

	jmpIfPos := c.currentPos()
	c.emit(OP_JMP_IF, 0)

	// Compila bloco then
	c.compileIfBlock(s.Then)

	if s.Else != nil {
		// Com else: precisa de dois saltos
		jmpPos := c.currentPos()
		c.emit(OP_JMP, 0)

		elsePos := c.currentPos()
		c.compileIfBlock(s.Else)

		c.patchJump(jmpIfPos+1, elsePos)
		c.patchJump(jmpPos+1, c.currentPos())
	} else {
		// Sem else: apenas um salto
		c.patchJump(jmpIfPos+1, c.currentPos())
	}
}

// compileIfBlock compila um bloco de if/else
func (c *Compiler) compileIfBlock(block ast.Stmt) {
	if b, ok := block.(ast.BlockStmt); ok {
		for _, stmt := range b.Body {
			c.compileStmt(stmt)
		}
	} else {
		c.compileStmt(block)
	}
}

// compileFor compila um loop for
func (c *Compiler) compileFor(s ast.ForStmt) {
	// Inicialização
	c.compileStmt(s.Init)

	// Início do loop (condição)
	condPos := c.currentPos()
	c.compileExpr(s.Condition)
	c.emit(OP_JMP_IF, 0)
	jmpExitPos := c.currentPos()

	// Corpo do loop
	for _, stmt := range s.Body {
		c.compileStmt(stmt)
	}

	// Pós-iteração
	c.compileStmt(s.Post)

	// Volta para condição
	c.emit(OP_JMP, byte(condPos))

	// Corrige salto de saída
	c.patchJump(jmpExitPos-1, c.currentPos())
}

// compileWhile compila um loop while
func (c *Compiler) compileWhile(s ast.WhileStmt) {
	// Início do loop (condição)
	condPos := c.currentPos()
	c.compileExpr(s.Condition)
	c.emit(OP_JMP_IF, 0)
	jmpExitPos := c.currentPos()

	// Corpo do loop
	for _, stmt := range s.Body {
		c.compileStmt(stmt)
	}

	// Volta para condição
	c.emit(OP_JMP, byte(condPos))

	// Corrige salto de saída
	c.patchJump(jmpExitPos-1, c.currentPos())
}


