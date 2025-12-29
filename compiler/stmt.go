package compiler

import (
	"fmt"

	"github.com/peiblow/vvm/ast"
)

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
	case ast.RequireStmt:
		c.compileRequire(s)
	case ast.RegistryDeclareStmt:
		c.compileRegistryDeclare(s)
	case ast.AgentStmt:
		c.compileAgentStmt(s)
	case ast.PolicyStmt:
		c.compilePolicyStmt(s)

	default:
		fmt.Printf("Unrecognized statement type: %T\n", s)
	}
}

func (c *Compiler) compileContract(s ast.ContractStmt) {
	for _, stmt := range s.Body {
		c.compileStmt(stmt)
	}
}

func (c *Compiler) compileVarDecl(s ast.VarDeclStmt) {
	if s.AssignedValue == nil {
		return
	}
	c.compileExpr(s.AssignedValue)
	slot := c.allocSlot(s.Identifier)
	c.emit(OP_STORE, byte(slot))
}

func (c *Compiler) compileReturn(s ast.ReturnStmt) {
	c.compileExpr(s.Value)
	if c.isInFunction {
		c.emit(OP_RET)
	} else {
		c.emit(OP_PRINT)
	}
}

func (c *Compiler) compileFunc(s ast.FuncStmt) {
	funcName := c.extractFuncName(s.Name)

	prevInFunction := c.isInFunction
	c.isInFunction = true

	skipFuncPos := c.currentPos()
	c.emit(OP_JMP, 0)

	funcMeta := FunctionMeta{
		Addr: c.currentPos(),
		Args: []int{},
	}

	funcMeta.Args = c.compileFuncArgs(s.Arguments)

	c.Functions[funcName] = funcMeta
	c.FunctionName[c.currentPos()] = funcName

	c.compileFuncBody(s.Body)

	c.emitFuncReturn(s.ReturnType)

	c.patchJump(skipFuncPos+1, c.currentPos())
	c.isInFunction = prevInFunction
}

func (c *Compiler) extractFuncName(name ast.Expr) string {
	switch n := name.(type) {
	case ast.SymbolExpr:
		return n.Value
	case ast.ExpressionStmt:
		if sym, ok := n.Expression.(ast.SymbolExpr); ok {
			return sym.Value
		}
	}
	panic("Unsupported function name expression type")
}

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

func (c *Compiler) compileFuncBody(body ast.Stmt) {
	if block, ok := body.(ast.BlockStmt); ok {
		c.compileBlock(block)
	} else {
		c.compileStmt(body)
	}
}

func (c *Compiler) emitFuncReturn(returnType ast.Type) {
	if t, ok := returnType.(ast.SymbolType); ok && t.Name == "void" {
		c.emit(OP_RET)
	}
}

func (c *Compiler) compileIf(s ast.IfStmt) {
	c.compileExpr(s.Condition)

	jmpIfPos := c.currentPos()
	c.emit(OP_JMP_IF, 0)

	c.compileIfBlock(s.Then)

	if s.Else != nil {
		jmpPos := c.currentPos()
		c.emit(OP_JMP, 0)

		elsePos := c.currentPos()
		c.compileIfBlock(s.Else)

		c.patchJump(jmpIfPos+1, elsePos)
		c.patchJump(jmpPos+1, c.currentPos())
	} else {
		c.patchJump(jmpIfPos+1, c.currentPos())
	}
}

func (c *Compiler) compileIfBlock(block ast.Stmt) {
	if b, ok := block.(ast.BlockStmt); ok {
		for _, stmt := range b.Body {
			c.compileStmt(stmt)
		}
	} else {
		c.compileStmt(block)
	}
}

func (c *Compiler) compileFor(s ast.ForStmt) {
	c.compileStmt(s.Init)

	condPos := c.currentPos()
	c.compileExpr(s.Condition)
	c.emit(OP_JMP_IF, 0)
	jmpExitPos := c.currentPos()

	for _, stmt := range s.Body {
		c.compileStmt(stmt)
	}

	c.compileStmt(s.Post)

	c.emit(OP_JMP, byte(condPos))

	c.patchJump(jmpExitPos-1, c.currentPos())
}

func (c *Compiler) compileWhile(s ast.WhileStmt) {
	condPos := c.currentPos()
	c.compileExpr(s.Condition)
	c.emit(OP_JMP_IF, 0)
	jmpExitPos := c.currentPos()

	for _, stmt := range s.Body {
		c.compileStmt(stmt)
	}

	c.emit(OP_JMP, byte(condPos))

	c.patchJump(jmpExitPos-1, c.currentPos())
}

func (c *Compiler) compileRequire(s ast.RequireStmt) {
	c.compileExpr(s.Condition)

	jmpToEndPos := c.currentPos()
	c.emit(OP_JMP_IF, 0)

	jmpPastErrorPos := c.currentPos()
	c.emit(OP_JMP, 0)

	errorBlockPos := c.currentPos()
	errIdx := c.addConst(s.Message)
	c.emit(OP_CONST, errIdx)
	c.emit(OP_ERR)

	endPos := c.currentPos()

	c.patchJump(jmpToEndPos+1, errorBlockPos)
	c.patchJump(jmpPastErrorPos+1, endPos)
}

func (c *Compiler) compileRegistryDeclare(s ast.RegistryDeclareStmt) {
	kindIdx := c.addConst(s.Kind)
	c.emit(OP_CONST, kindIdx)
	nameIdx := c.addConst(s.Name)
	c.emit(OP_CONST, nameIdx)
	versionIdx := c.addConst(s.Version)
	c.emit(OP_CONST, versionIdx)
	ownerIdx := c.addConst(s.Owner)
	c.emit(OP_CONST, ownerIdx)
	purposeIdx := c.addConst(s.Purpose)
	c.emit(OP_CONST, purposeIdx)

	c.emit(
		OP_REGISTRY_DECLARE,
		byte(kindIdx),
		byte(nameIdx),
		byte(versionIdx),
		byte(ownerIdx),
		byte(purposeIdx),
	)
}

func (c *Compiler) compileAgentStmt(s ast.AgentStmt) {
	c.allocSlot(s.Identifier.(ast.SymbolExpr).Value)
	identifierIdx := c.findConst(s.Identifier)
	c.emit(OP_CONST, identifierIdx)

	c.emit(OP_REGISTRY_GET, byte(identifierIdx))
	c.emit(OP_STORE, 0)
}

func (c *Compiler) compilePolicyStmt(s ast.PolicyStmt) {
	c.allocSlot(s.Identifier.(ast.SymbolExpr).Value)

	identifierIdx := c.findConst(s.Identifier)
	if identifierIdx == 255 {
		identifierIdx = c.addConst(s.Identifier)
	}
	c.emit(OP_CONST, identifierIdx)

	c.emit(OP_PUSH_OBJECT)
	for key, value := range s.Rules {
		keyIdx := c.addConst(key)
		c.emit(OP_CONST, keyIdx)

		switch v := value.(type) {
		case ast.NumberExpr:
			valIdx := c.addConst(v.Value)
			c.emit(OP_CONST, valIdx)
		case ast.StringExpr:
			valIdx := c.addConst(v.Value)
			c.emit(OP_CONST, valIdx)
		default:
			panic(fmt.Sprintf("Unsupported policy rule value type: %T", v))
		}

		c.emit(OP_SET_PROPERTY)
	}

	c.emit(OP_POLICY_DECLARE, byte(identifierIdx))
	c.emit(OP_STORE, byte(c.Symbols[s.Identifier.(ast.SymbolExpr).Value]))
}
