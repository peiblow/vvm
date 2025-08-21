package main

type Node interface{}

type Contract struct {
	Name string
	Body []Node
}

type VarAssign struct {
	Name  string
	Value Node
}

type Literal struct {
	Value interface{}
}

type Identifier struct {
	Name string
}

type Function struct {
	Name       string
	ReturnType string
	Body       []Node
}

type BinaryOp struct {
	Left  Node
	Op    string
	Right Node
}

type IfStmt struct {
	Condition Node
	Then      Node
	Else      Node
}

type PrintStmt struct {
	Expr Node
}

type ForLoop struct {
	Init Node
	Cond Node
	Post Node
	Body []Node
}

type ArrayLiteral struct {
	Elements []Node
}

type ArrayAccess struct {
	Array Node
	Index Node
}
