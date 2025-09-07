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

type NullExpr struct{}

func (n NullExpr) expr() {}

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

type ArrayLiteralExpr struct {
	Items []Expr
}

func (n ArrayLiteralExpr) expr() {}

type ArrayAccessItemExpr struct {
	Array Expr
	Index Expr
}

func (n ArrayAccessItemExpr) expr() {}

type ObjectPropertyExpr struct {
	Key   Expr
	Value Expr
}

func (n ObjectPropertyExpr) expr() {}

type ObjectAssignmentExpr struct {
	Name   Expr
	Fields []ObjectPropertyExpr
}

func (n ObjectAssignmentExpr) expr() {}

type CallExpr struct {
	Calle     Expr
	Arguments []Expr
}

func (n CallExpr) expr() {}

type MemberExpr struct {
	Object   Expr
	Property Expr
}

func (n MemberExpr) expr() {}
