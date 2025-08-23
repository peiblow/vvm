package ast

type BlockStmt struct {
	Body []Stmt
}

func (n BlockStmt) stmt() {}

type ExpressionStmt struct {
	Expression Expr
}

func (n ExpressionStmt) stmt() {}

type VarDeclStmt struct {
	Identifier    string
	Constant      bool
	AssignedValue Expr
	// ExplicityType Type
}

func (n VarDeclStmt) stmt() {}
