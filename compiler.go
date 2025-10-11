package main

import (
	"fmt"

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

func NewCompiler() *Compiler {
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

func (c *Compiler) GetFuncArgs(addr int) []int {
	funcName := c.FunctionName[addr]
	args := c.Functions[funcName].Args
	return args
}

func (c *Compiler) CompileBlock(block ast.BlockStmt) {
	for _, stmt := range block.Body {
		c.compile_stmt(stmt)
	}
	c.emit(OP_HALT)
}

func (c *Compiler) compile_block(block ast.BlockStmt) {
	for _, stmt := range block.Body {
		c.compile_stmt(stmt)
	}
}

func (c *Compiler) compile_stmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case ast.ContractStmt:
		for _, b := range s.Body {
			c.compile_stmt(b)
		}
	case ast.ExpressionStmt:
		c.compile_expr(s.Expression)
	case ast.VarDeclStmt:
		if s.AssignedValue != nil {
			c.compile_expr(s.AssignedValue)
			c.Symbols[s.Identifier] = c.NextSlot
			c.emit(OP_STORE, byte(c.NextSlot))
			c.NextSlot++
		}
	case ast.ReturnStmt:
		c.compile_expr(s.Value)
		if c.isInFunction {
			c.emit(OP_RET)
		} else {
			c.emit(OP_PRINT)
		}
	case ast.FuncStmt:
		var funcName string
		switch n := s.Name.(type) {
		case ast.SymbolExpr:
			funcName = n.Value
		case ast.ExpressionStmt:
			if sym, ok := n.Expression.(ast.SymbolExpr); ok {
				funcName = sym.Value
			} else {
				panic("Expected SymbolExpr inside ExpressionStmt for function name")
			}
		default:
			panic("Unsupported expression type for function name")
		}

		prevInFunction := c.isInFunction
		c.isInFunction = true

		skipFuncPos := len(c.Code)
		c.emit(OP_JMP, 0)

		funcMeta := FunctionMeta{
			Addr: len(c.Code),
			Args: []int{},
		}

		if block, ok := s.Arguments.(ast.BlockStmt); ok {
			for _, arg := range block.Body {
				if exprStmt, ok := arg.(ast.ExpressionStmt); ok {
					if sym, ok := exprStmt.Expression.(ast.SymbolExpr); ok {
						slot := c.NextSlot
						c.Symbols[sym.Value] = slot
						c.NextSlot++
						funcMeta.Args = append(funcMeta.Args, slot)
					}
				}
			}
		}

		c.Functions[funcName] = funcMeta
		c.FunctionName[len(c.Code)] = funcName

		if block, ok := s.Body.(ast.BlockStmt); ok {
			c.compile_block(block)
		} else {
			c.compile_stmt(s.Body)
		}

		switch t := s.ReturnType.(type) {
		case ast.SymbolType:
			if t.Name == "void" {
				c.emit(OP_RET)
			}
		default:
			c.emit(OP_REQUIRE)
		}

		afterFuncPos := len(c.Code)
		c.Code[skipFuncPos+1] = byte(afterFuncPos)
		c.isInFunction = prevInFunction
	case ast.IfStmt:
		c.compile_expr(s.Condition)

		jmpIfPos := len(c.Code)
		c.emit(OP_JMP_IF, 0)

		if block, ok := s.Then.(ast.BlockStmt); ok {
			for _, stmt := range block.Body {
				c.compile_stmt(stmt)
			}
		} else {
			c.compile_stmt(s.Then)
		}

		if s.Else != nil {
			jmpPos := len(c.Code)
			c.emit(OP_JMP, 0)

			elsePos := len(c.Code)
			if block, ok := s.Else.(ast.BlockStmt); ok {
				if len(block.Body) > 0 {
					for _, stmt := range block.Body {
						c.compile_stmt(stmt)
					}
				}
			} else {
				c.compile_stmt(s.Else)
			}

			// resolve saltos
			c.Code[jmpIfPos+1] = byte(elsePos)
			endPos := len(c.Code)
			c.Code[jmpPos+1] = byte(endPos)
		} else {
			endPos := len(c.Code)
			c.Code[jmpIfPos+1] = byte(endPos)
		}
	case ast.ForStmt:
		c.compile_stmt(s.Init)

		condpos := len(c.Code)
		c.compile_expr(s.Condition)
		c.emit(OP_JMP_IF, 0)

		jmpExitPos := len(c.Code)

		for _, block := range s.Body {
			c.compile_stmt(block)
		}

		c.compile_stmt(s.Post)
		c.emit(OP_JMP, byte(condpos))

		endPos := len(c.Code)
		c.Code[jmpExitPos-1] = byte(endPos)
	case ast.WhileStmt:
		condpos := len(c.Code)
		c.compile_expr(s.Condition)
		c.emit(OP_JMP_IF, 0)

		jmpExitPos := len(c.Code)

		for _, block := range s.Body {
			c.compile_stmt(block)
		}

		c.emit(OP_JMP, byte(condpos))

		endPos := len(c.Code)
		c.Code[jmpExitPos-1] = byte(endPos)
	case ast.ConstructorStmt:
		c.emit(OP_PUSH_OBJECT)
		c.emit(OP_STORE, 0)

		if block, ok := s.Arguments.(ast.BlockStmt); ok {
			for _, arg := range block.Body {
				if exprStmt, ok := arg.(ast.ExpressionStmt); ok {
					if sym, ok := exprStmt.Expression.(ast.SymbolExpr); ok {
						slot := c.NextSlot
						c.Symbols[sym.Value] = slot
						c.NextSlot++
					}
				}
			}
		}

		c.NextSlot++
		c.compile_block(s.Body)
	default:
		fmt.Println("Statement not found 404: ", s)
	}
}

func (c *Compiler) compile_expr(expr ast.Expr) {
	switch e := expr.(type) {
	case ast.NumberExpr:
		c.emit(OP_PUSH, byte(e.Value))
	case ast.StringExpr:
		idx := len(c.ConstPool)
		c.ConstPool = append(c.ConstPool, e.Value)
		c.emit(OP_CONST, byte(idx))
	case ast.SymbolExpr:
		slot := c.Symbols[e.Value]
		c.emit(OP_SLOAD, byte(slot))
	case ast.ArrayLiteralExpr:
		items := make([]interface{}, len(e.Items))
		for i, item := range e.Items {
			switch it := item.(type) {
			case ast.NumberExpr:
				items[i] = it.Value
			case ast.StringExpr:
				items[i] = it.Value
			case ast.ArrayLiteralExpr:
				inner := make([]interface{}, len(it.Items))
				for j, innerItem := range it.Items {
					switch ii := innerItem.(type) {
					case ast.NumberExpr:
						inner[j] = ii.Value
					case ast.StringExpr:
						inner[j] = ii.Value
					default:
						panic("Unsupported nested array item")
					}
				}
				items[i] = inner
			default:
				panic("Unsupported array item type")
			}
		}

		idx := len(c.ConstPool)
		c.ConstPool = append(c.ConstPool, items)
		c.emit(OP_CONST, byte(idx))
	case ast.AssignmentExpr:
		if objExpr, ok := e.Right.(ast.ObjectAssignmentExpr); ok {
			obj := make(map[string]interface{})
			for _, property := range objExpr.Fields {
				key := property.Key.(ast.SymbolExpr).Value
				switch v := property.Value.(type) {
				case ast.StringExpr:
					obj[key] = v.Value
				case ast.NumberExpr:
					obj[key] = v.Value
				case ast.SymbolExpr:
					symName := c.Symbols[v.Value]
					obj[key] = c.ConstPool[symName]
				default:
					panic(fmt.Sprintf("unsupported value type in object: %T", v))
				}
			}

			idx := len(c.ConstPool)
			c.ConstPool = append(c.ConstPool, obj)
			c.emit(OP_CONST, byte(idx))
		} else {
			c.compile_expr(e.Right)
		}

		var symName string
		switch l := e.Left.(type) {
		case ast.SymbolExpr:
			symName = l.Value
		case ast.ExpressionStmt:
			if se, ok := l.Expression.(ast.SymbolExpr); ok {
				symName = se.Value
			} else {
				panic("Expected SymbolExpr in ExpressionStmt for assignment left")
			}
		case ast.MemberExpr:
			if _, ok := l.Object.(ast.ThisExpr); ok {
				c.emit(OP_SLOAD, 0)
			} else if se, ok := l.Object.(ast.SymbolExpr); ok {
				c.emit(OP_SLOAD, byte(c.Symbols[se.Value]))
			}

			idx := len(c.ConstPool)
			c.ConstPool = append(c.ConstPool, l.Property.(ast.SymbolExpr).Value)
			c.emit(OP_CONST, byte(idx))
			c.compile_expr(e.Right)
			c.emit(OP_SET_PROPERTY)
		default:
			panic("Unsupported assignment left expression type")
		}

		slot, ok := c.Symbols[symName]
		if !ok {
			slot = c.NextSlot
			c.Symbols[symName] = slot
			c.NextSlot++
		}

		c.emit(OP_STORE, byte(slot))
	case ast.BinaryExpr:
		c.compile_expr(e.Left)
		c.compile_expr(e.Right)
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
	case ast.CallExpr:
		for _, arg := range e.Arguments {
			c.compile_expr(arg)
		}

		if calle, ok := e.Calle.(ast.SymbolExpr); ok {
			switch calle.Value {
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
				addr := c.Functions[calle.Value]
				c.emit(OP_CALL, byte(addr.Addr))
			}
		}
	case ast.PrefixExpr:
		c.compile_expr(e.RightExpr)
		switch e.Operator.Literal {
		case "-":
			c.emit(OP_PUSH, 0)
			c.emit(OP_SWAP)
			c.emit(OP_SUB)
		}
	case ast.IncDecExpr:
		if sym, ok := e.Left.(ast.SymbolExpr); ok {
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
	case ast.ArrayAccessItemExpr:
		c.compile_expr(e.Array)
		c.compile_expr(e.Index)

		c.emit(OP_ACCESS)
	case ast.MemberExpr:
		if _, ok := e.Object.(ast.ThisExpr); ok {
			c.emit(OP_SLOAD, 0)
		} else {
			c.compile_expr(e.Object)
		}

		if prop, ok := e.Property.(ast.SymbolExpr); ok {
			idx := len(c.ConstPool)
			c.ConstPool = append(c.ConstPool, prop.Value)
			c.emit(OP_CONST, byte(idx))
		} else {
			c.compile_expr(e.Property)
		}

		c.emit(OP_GET_PROPERTY)
	case ast.ThisExpr:
		c.emit(OP_SLOAD, 0)
	case ast.NullExpr:
		c.emit(OP_NULL)
	}
}
