package parser

import (
	"fmt"
	"strconv"

	"github.com/peiblow/vvm/ast"
	"github.com/peiblow/vvm/lexer"
)

func parse_expr(p *parser, bp binding_power) ast.Expr {
	tokenType := p.currentTokenType()
	nud_fn, exists := nud_lu[tokenType]

	if !exists {
		panic(fmt.Sprintf("NUD handler expected for token %v - %v", tokenType, p.currentToken().Literal))
	}

	left := nud_fn(p)

	for bp_lu[p.currentTokenType()] > bp {
		tokenType = p.currentTokenType()
		led_fn, exists := led_lu[tokenType]

		if !exists {
			panic(fmt.Sprintf("LED handler expected for token %v", tokenType))
		}

		left = led_fn(p, left, bp_lu[p.currentTokenType()])
	}

	return left
}

func parse_primary_expr(p *parser) ast.Expr {
	switch p.currentTokenType() {
	case lexer.NUMBER:
		number, _ := strconv.ParseFloat(p.advance().Literal, 64)
		return ast.NumberExpr{
			Value: number,
		}
	case lexer.STRING:
		return ast.StringExpr{
			Value: p.advance().Literal,
		}
	case lexer.IDENTIFIER:
		return ast.SymbolExpr{
			Value: p.advance().Literal,
		}
	case lexer.PLUS_PLUS:
		return ast.SymbolExpr{
			Value: p.advance().Literal,
		}
	default:
		panic("Cannot create primary expr")
	}
}

func parse_incdec_expr(p *parser, left ast.Expr, bp binding_power) ast.Expr {
	op := p.advance()
	return ast.IncDecExpr{
		Left:     left,
		Operator: op,
	}
}

func parse_binary_expr(p *parser, left ast.Expr, bp binding_power) ast.Expr {
	operatorToken := p.advance()
	right := parse_expr(p, bp)

	return ast.BinaryExpr{
		Left:     left,
		Operator: operatorToken,
		Right:    right,
	}
}

func parse_assignment(p *parser, left ast.Expr, bp binding_power) ast.Expr {
	operatorToken := p.currentToken()
	p.advance()
	right := parse_expr(p, assignment)

	return ast.AssignmentExpr{
		Left:     left,
		Operator: operatorToken,
		Right:    right,
	}
}

func parse_prefix_expr(p *parser) ast.Expr {
	return ast.PrefixExpr{
		Operator:  p.advance(),
		RightExpr: parse_expr(p, defalt_bp),
	}
}

func grouping_expr(p *parser) ast.Expr {
	p.advance()

	expr := parse_expr(p, defalt_bp)
	p.expect(lexer.CLOSE_PAREN)
	return expr
}
