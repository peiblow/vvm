package compiler

import (
	"fmt"

	"github.com/peiblow/vvm/ast"
)

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
	case ast.ObjectAssignmentExpr:
		c.compileObjectLiteral(e)
	case ast.ThisExpr:
		c.emit(OP_SLOAD, 0)
	case ast.NullExpr:
		c.emit(OP_NULL)
	}
}

func (c *Compiler) compileNumber(e ast.NumberExpr) {
	if e.Value <= 256 {
		c.emit(OP_PUSH, byte(e.Value))
	} else {
		idx := c.addConst(e.Value)
		c.emit(OP_CONST, idx)
	}
}

func (c *Compiler) compileString(e ast.StringExpr) {
	idx := c.addConst(e.Value)
	c.emit(OP_CONST, idx)
}

func (c *Compiler) compileSymbol(e ast.SymbolExpr) {
	slot := c.Symbols[e.Value]
	c.emit(OP_SLOAD, byte(slot))
}

func (c *Compiler) compileArrayLiteral(e ast.ArrayLiteralExpr) {
	items := c.convertArrayItems(e.Items)
	idx := c.addConst(items)
	c.emit(OP_CONST, idx)
}

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
		case ast.ObjectAssignmentExpr:
			result[i] = c.convertObjectToMap(it)
		default:
			panic(fmt.Sprintf("Array item type not supported: %T", item))
		}
	}
	return result
}

func (c *Compiler) convertObjectToMap(obj ast.ObjectAssignmentExpr) map[string]interface{} {
	result := make(map[string]interface{})
	for _, field := range obj.Fields {
		var key string
		switch k := field.Key.(type) {
		case ast.SymbolExpr:
			key = k.Value
		case ast.StringExpr:
			key = k.Value
		default:
			panic(fmt.Sprintf("Object key type not supported: %T", field.Key))
		}

		// Extract value
		switch v := field.Value.(type) {
		case ast.NumberExpr:
			result[key] = v.Value
		case ast.StringExpr:
			result[key] = v.Value
		case ast.ArrayLiteralExpr:
			result[key] = c.convertArrayItems(v.Items)
		case ast.ObjectAssignmentExpr:
			result[key] = c.convertObjectToMap(v)
		default:
			panic(fmt.Sprintf("Object property value type not supported: %T", field.Value))
		}
	}
	return result
}

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

func (c *Compiler) compileSymbolAssignment(name string, right ast.Expr) {
	slot := c.getSlot(name)

	if objExpr, ok := right.(ast.ObjectAssignmentExpr); ok {
		c.compileObjectAssignment(objExpr, slot)
	} else {
		c.compileExpr(right)
	}

	c.emit(OP_STORE, byte(slot))
}

func (c *Compiler) compileMemberAssignment(member ast.MemberExpr, right ast.Expr) {
	// Carrega o objeto
	if _, ok := member.Object.(ast.ThisExpr); ok {
		c.emit(OP_SLOAD, 0)
	} else if sym, ok := member.Object.(ast.SymbolExpr); ok {
		c.emit(OP_SLOAD, byte(c.Symbols[sym.Value]))
	}

	idx := c.addConst(member.Property.(ast.SymbolExpr).Value)
	c.emit(OP_CONST, idx)

	c.compileExpr(right)
	c.emit(OP_SET_PROPERTY)
}

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

func (c *Compiler) compileObjectLiteral(obj ast.ObjectAssignmentExpr) {
	c.emit(OP_PUSH_OBJECT)

	for _, prop := range obj.Fields {
		key := prop.Key.(ast.SymbolExpr).Value
		idx := c.addConst(key)
		c.emit(OP_CONST, idx)
		c.compileExpr(prop.Value)
		c.emit(OP_SET_PROPERTY)
	}
}

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

func (c *Compiler) compileCall(e ast.CallExpr) {
	if callee, ok := e.Calle.(ast.SymbolExpr); ok {
		if err := c.ValidateFunctionCall(callee.Value, e.Arguments); err != nil {
			panic(fmt.Sprintf("Type error: %s", err.Error()))
		}
	}

	for _, arg := range e.Arguments {
		c.compileExpr(arg)
	}

	if callee, ok := e.Calle.(ast.SymbolExpr); ok {
		c.compileBuiltinOrUserCall(callee.Value)
	}
}

func (c *Compiler) compileBuiltinOrUserCall(name string) {
	switch name {
	case "print":
		c.emit(OP_PRINT)
	case "len", "length":
		c.emit(OP_LENGTH)
	case "_transfer":
		c.emit(OP_TRANSFER)
	case "balanceOf":
		c.emit(OP_BALANCE_OF)
	case "require":
		c.emit(OP_REQUIRE)
	default:
		addr := c.Functions[name]
		c.emit(OP_CALL, byte(addr.Addr>>8), byte(addr.Addr&0xFF))
	}
}

func (c *Compiler) compilePrefix(e ast.PrefixExpr) {
	c.compileExpr(e.RightExpr)
	if e.Operator.Literal == "-" {
		c.emit(OP_PUSH, 0)
		c.emit(OP_SWAP)
		c.emit(OP_SUB)
	}
}

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

func (c *Compiler) compileArrayAccess(e ast.ArrayAccessItemExpr) {
	c.compileExpr(e.Array)
	c.compileExpr(e.Index)
	c.emit(OP_ACCESS)
}

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
