package parser

import (
	"fmt"

	"github.com/peiblow/vvm/ast"
	"github.com/peiblow/vvm/lexer"
)

func parse_block(p *parser) ast.BlockStmt {
	p.expect(lexer.OPEN_CURLY)

	body := make([]ast.Stmt, 0)
	for p.hasTokens() && p.currentTokenType() != lexer.CLOSE_CURLY {
		body = append(body, parse_stmt(p))
	}

	p.expect(lexer.CLOSE_CURLY)

	return ast.BlockStmt{
		Body: body,
	}
}

func parse_stmt(p *parser) ast.Stmt {
	stmt_fn, exists := stmt_lu[p.currentTokenType()]

	if exists {
		return stmt_fn(p)
	}

	expr := parse_expr(p, defalt_bp)

	return ast.ExpressionStmt{
		Expression: expr,
	}
}

func parse_arguments(p *parser) ast.Stmt {
	p.expect(lexer.OPEN_PAREN)
	body := []ast.Stmt{}

	for p.currentTokenType() != lexer.CLOSE_PAREN {
		expr := parse_expr(p, defalt_bp)
		body = append(body, ast.ExpressionStmt{Expression: expr})

		if p.currentTokenType() == lexer.COMMA {
			p.advance()
		}
	}

	p.expect(lexer.CLOSE_PAREN)
	return ast.BlockStmt{Body: body}
}

func parse_contract_decl(p *parser) ast.Stmt {
	p.expect(lexer.CONTRACT)
	contractName := p.currentToken().Literal
	p.advance()

	body := parse_block(p)

	return ast.ContractStmt{
		Identifier: contractName,
		Body:       body.Body,
	}
}

func parse_var_decl(p *parser) ast.Stmt {
	isConst := p.advance().Type == lexer.CONST
	varName := p.expectError(lexer.IDENTIFIER, "Inside variable declaration expected to find variable name").Literal

	p.expect(lexer.ASSIGNMENT)
	assignmentValue := parse_expr(p, assignment)

	return ast.VarDeclStmt{
		Identifier:    varName,
		AssignedValue: assignmentValue,
		Constant:      isConst,
	}
}

func parse_if_stmt(p *parser) ast.Stmt {
	p.expect(lexer.IF)
	p.expect(lexer.OPEN_PAREN)
	condition := parse_expr(p, defalt_bp)
	p.expect(lexer.CLOSE_PAREN)

	thenBlock := parse_block(p)
	var elseBlock ast.BlockStmt
	if p.currentTokenType() == lexer.ELSE {
		p.advance()
		elseBlock = parse_block(p)
	}

	return ast.IfStmt{
		Condition: condition,
		Then:      thenBlock,
		Else:      elseBlock,
	}
}

func parse_while_loop_stmt(p *parser) ast.Stmt {
	p.advance()
	p.expect(lexer.OPEN_PAREN)
	cond := parse_expr(p, defalt_bp)
	p.expect(lexer.CLOSE_PAREN)
	body := parse_block(p)

	return ast.WhileStmt{
		Condition: cond,
		Body:      body,
	}
}

func parse_for_loop_stmt(p *parser) ast.Stmt {
	p.advance()
	p.expect(lexer.OPEN_PAREN)

	init := parse_stmt(p)
	p.expect(lexer.SEMI_COLON)

	cond := parse_expr(p, defalt_bp)
	p.expect(lexer.SEMI_COLON)

	post := parse_stmt(p)
	p.expect(lexer.CLOSE_PAREN)

	body := parse_block(p)

	return ast.ForStmt{
		Init:      init,
		Condition: cond,
		Post:      post,
		Body:      body,
	}
}

func parse_func_stmt(p *parser) ast.Stmt {
	p.expect(lexer.FUNC)
	name := parse_expr(p, defalt_bp)
	args := parse_arguments(p)

	fmt.Println(p.currentToken().Literal)
	body := parse_block(p)

	return ast.FuncStmt{
		Name:      name,
		Arguments: args,
		Body:      body,
	}
}
