package parser

import (
	"github.com/peiblow/vvm/ast"
	"github.com/peiblow/vvm/lexer"
)

func parse_stmt(p *parser) ast.Stmt {
	stmt_fn, exists := stmt_lu[p.currentTokenType()]

	if exists {
		return stmt_fn(p)
	}

	expr := parse_expr(p, defalt_bp)
	p.expect(lexer.BREAK_LINE)

	return ast.ExpressionStmt{
		Expression: expr,
	}
}

func parse_var_decl(p *parser) ast.Stmt {
	isConst := p.advance().Type == lexer.CONST
	varName := p.expectError(lexer.IDENTIFIER, "Inside variable declaration expected to find variable name").Literal

	p.expect(lexer.ASSIGNMENT)
	assignmentValue := parse_expr(p, assignment)
	p.expect(lexer.BREAK_LINE)

	return ast.VarDeclStmt{
		Identifier:    varName,
		AssignedValue: assignmentValue,
		Constant:      isConst,
	}
}
