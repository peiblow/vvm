package ast

import "github.com/peiblow/vvm/lexer"

// Literals Expressions
type NumberExpr struct {
	Value float64
}

func (n NumberExpr) expr() {}

type StringExpr struct {
	Value string
}

func (n StringExpr) expr() {}

type SymbolExpr struct {
	Value string
}

func (n SymbolExpr) expr() {}

// Complex Expressions
type BinaryExpr struct {
	Left     Expr
	Operator lexer.Token
	Right    Expr
}

func (n BinaryExpr) expr() {}

type PrefixExpr struct {
	Operator  lexer.Token
	RightExpr Expr
}

func (n PrefixExpr) expr() {}

func (n ExpressionStmt) expr() {}

type IncDecExpr struct {
	Left     Expr
	Operator lexer.Token
}

func (n IncDecExpr) expr() {}

type AssignmentExpr struct {
	Left     Expr
	Operator lexer.Token
	Right    Expr
}

func (n AssignmentExpr) expr() {}
