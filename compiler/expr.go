package compiler

import (
	"github.com/peiblow/vvm/ast"
)

// compileExpr compila uma expressão
func (c *Compiler) compileExpr(expr ast.Expr) {
	switch e := expr.(type) {
	case ast.NumberExpr:
		c.compileNumber(e)
	case ast.StringExpr:
		c.compileString(e)
	case ast.SymbolExpr:
		c.compileSymbol(e)
	case ast.ArrayLiteralExpr:
		c.compileArrayLiteral(e)
	case ast.AssignmentExpr:
		c.compileAssignment(e)
	case ast.BinaryExpr:
		c.compileBinary(e)
	case ast.CallExpr:
		c.compileCall(e)
	case ast.PrefixExpr:
		c.compilePrefix(e)
	case ast.IncDecExpr:
		c.compileIncDec(e)
	case ast.ArrayAccessItemExpr:
		c.compileArrayAccess(e)
	case ast.MemberExpr:
		c.compileMember(e)
	case ast.ThisExpr:
		c.emit(OP_SLOAD, 0)
	case ast.NullExpr:
		c.emit(OP_NULL)
	}
}

// compileNumber compila um número literal
func (c *Compiler) compileNumber(e ast.NumberExpr) {
	c.emit(OP_PUSH, byte(e.Value))
}

// compileString compila uma string literal
func (c *Compiler) compileString(e ast.StringExpr) {
	idx := c.addConst(e.Value)
	c.emit(OP_CONST, idx)
}

// compileSymbol compila uma referência a símbolo/variável
func (c *Compiler) compileSymbol(e ast.SymbolExpr) {
	slot := c.Symbols[e.Value]
	c.emit(OP_SLOAD, byte(slot))
}

// compileArrayLiteral compila um array literal
func (c *Compiler) compileArrayLiteral(e ast.ArrayLiteralExpr) {
	items := c.convertArrayItems(e.Items)
	idx := c.addConst(items)
	c.emit(OP_CONST, idx)
}

// convertArrayItems converte itens da AST para valores Go
func (c *Compiler) convertArrayItems(items []ast.Expr) []interface{} {
	result := make([]interface{}, len(items))
	for i, item := range items {
		switch it := item.(type) {
		case ast.NumberExpr:
			result[i] = it.Value
		case ast.StringExpr:
			result[i] = it.Value
		case ast.ArrayLiteralExpr:
			result[i] = c.convertArrayItems(it.Items)
		default:
			panic("Array item type not supported")
		}
	}
	return result
}

// compileAssignment compila uma expressão de atribuição
func (c *Compiler) compileAssignment(e ast.AssignmentExpr) {
	switch left := e.Left.(type) {
	case ast.SymbolExpr:
		c.compileSymbolAssignment(left.Value, e.Right)
	case ast.ExpressionStmt:
		if sym, ok := left.Expression.(ast.SymbolExpr); ok {
			c.compileSymbolAssignment(sym.Value, e.Right)
		} else {
			panic("Esperava SymbolExpr em ExpressionStmt para atribuição")
		}
	case ast.MemberExpr:
		c.compileMemberAssignment(left, e.Right)
	default:
		panic("Expression type not supported for left assignment")
	}
}

// compileSymbolAssignment compila atribuição a um símbolo
func (c *Compiler) compileSymbolAssignment(name string, right ast.Expr) {
	slot := c.getSlot(name)
	c.emit(OP_STORE, byte(slot))

	if objExpr, ok := right.(ast.ObjectAssignmentExpr); ok {
		c.compileObjectAssignment(objExpr, slot)
	} else {
		c.compileExpr(right)
	}
}

// compileMemberAssignment compila atribuição a membro de objeto
func (c *Compiler) compileMemberAssignment(member ast.MemberExpr, right ast.Expr) {
	// Carrega o objeto
	if _, ok := member.Object.(ast.ThisExpr); ok {
		c.emit(OP_SLOAD, 0)
	} else if sym, ok := member.Object.(ast.SymbolExpr); ok {
		c.emit(OP_SLOAD, byte(c.Symbols[sym.Value]))
	}

	// Carrega nome da propriedade
	idx := c.addConst(member.Property.(ast.SymbolExpr).Value)
	c.emit(OP_CONST, idx)

	// Compila valor e define propriedade
	c.compileExpr(right)
	c.emit(OP_SET_PROPERTY)
}

// compileObjectAssignment compila atribuição de objeto literal
func (c *Compiler) compileObjectAssignment(obj ast.ObjectAssignmentExpr, slot int) {
	c.emit(OP_PUSH_OBJECT)

	for _, prop := range obj.Fields {
		key := prop.Key.(ast.SymbolExpr).Value
		idx := c.addConst(key)
		c.emit(OP_CONST, idx)
		c.compileExpr(prop.Value)
		c.emit(OP_SET_PROPERTY)
	}

	c.emit(OP_STORE, byte(slot))
}

// compileBinary compila uma expressão binária
func (c *Compiler) compileBinary(e ast.BinaryExpr) {
	c.compileExpr(e.Left)
	c.compileExpr(e.Right)

	switch e.Operator.Literal {
	case "+":
		c.emit(OP_SWAP)
		c.emit(OP_ADD)
	case "-":
		c.emit(OP_SUB)
	case "*":
		c.emit(OP_MUL)
	case "/":
		c.emit(OP_DIV)
	case ">":
		c.emit(OP_GT)
	case "<":
		c.emit(OP_LT)
	case "==":
		c.emit(OP_EQ)
	case ">=":
		c.emit(OP_GT_EQ)
	case "<=":
		c.emit(OP_LT_EQ)
	case "!=":
		c.emit(OP_DIFF)
	case "+=":
		c.emit(OP_PLUS_EQ)
	}
}

// compileCall compila uma chamada de função
func (c *Compiler) compileCall(e ast.CallExpr) {
	// Compila argumentos
	for _, arg := range e.Arguments {
		c.compileExpr(arg)
	}

	// Compila chamada
	if callee, ok := e.Calle.(ast.SymbolExpr); ok {
		c.compileBuiltinOrUserCall(callee.Value)
	}
}

func (c *Compiler) compileBuiltinOrUserCall(name string) {
	switch name {
	case "print":
		c.emit(OP_PRINT)
	case "length":
		c.emit(OP_LENGTH)
	case "_transfer":
		c.emit(OP_TRANSFER)
	case "balanceOf":
		c.emit(OP_BALANCE_OF)
	case "require":
		c.emit(OP_REQUIRE)
	default:
		// Chamada de função de usuário
		addr := c.Functions[name]
		c.emit(OP_CALL, byte(addr.Addr))
	}
}

// compilePrefix compila uma expressão de prefixo
func (c *Compiler) compilePrefix(e ast.PrefixExpr) {
	c.compileExpr(e.RightExpr)
	if e.Operator.Literal == "-" {
		c.emit(OP_PUSH, 0)
		c.emit(OP_SWAP)
		c.emit(OP_SUB)
	}
}

// compileIncDec compila uma expressão de incremento/decremento
func (c *Compiler) compileIncDec(e ast.IncDecExpr) {
	sym, ok := e.Left.(ast.SymbolExpr)
	if !ok {
		return
	}

	slot := c.Symbols[sym.Value]
	c.emit(OP_SLOAD, byte(slot))
	c.emit(OP_PUSH, 1)

	if e.Operator.Literal == "++" {
		c.emit(OP_ADD)
	} else {
		c.emit(OP_SUB)
	}

	c.emit(OP_STORE, byte(slot))
}

// compileArrayAccess compila acesso a item de array
func (c *Compiler) compileArrayAccess(e ast.ArrayAccessItemExpr) {
	c.compileExpr(e.Array)
	c.compileExpr(e.Index)
	c.emit(OP_ACCESS)
}

// compileMember compila acesso a membro de objeto
func (c *Compiler) compileMember(e ast.MemberExpr) {
	if _, ok := e.Object.(ast.ThisExpr); ok {
		c.emit(OP_SLOAD, 0)
	} else {
		c.compileExpr(e.Object)
	}

	if prop, ok := e.Property.(ast.SymbolExpr); ok {
		idx := c.addConst(prop.Value)
		c.emit(OP_CONST, idx)
	} else {
		c.compileExpr(e.Property)
	}

	c.emit(OP_GET_PROPERTY)
}
