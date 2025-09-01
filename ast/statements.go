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
	ExplicityType Type
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
	Body      []Stmt
}

func (n WhileStmt) stmt() {}

type ForStmt struct {
	Init      Stmt
	Condition Expr
	Post      Stmt
	Body      []Stmt
}

func (n ForStmt) stmt() {}

type FuncStmt struct {
	Name       Expr
	Arguments  Stmt
	Body       Stmt
	ReturnType Type
}

func (n FuncStmt) stmt() {}

type ArrayItemAssignmentStmt struct {
	Name  Expr
	Index Expr
	Value Expr
}

func (n ArrayItemAssignmentStmt) stmt() {}

type ReturnStmt struct {
	Value Expr
}

func (n ReturnStmt) stmt() {}
