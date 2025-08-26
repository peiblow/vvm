package ast

type BlockStmt struct {
	Body []Stmt
}

func (n BlockStmt) stmt() {}

type ContractStmt struct {
	Identifier string
	Body       []Stmt
}

func (n ContractStmt) stmt() {}

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

type IfStmt struct {
	Condition Expr
	Then      Stmt
	Else      Stmt
}

func (n IfStmt) stmt() {}

type WhileStmt struct {
	Condition Expr
	Body      Stmt
}

func (n WhileStmt) stmt() {}

type ForStmt struct {
	Init      Stmt
	Condition Expr
	Post      Stmt
	Body      Stmt
}

func (n ForStmt) stmt() {}
