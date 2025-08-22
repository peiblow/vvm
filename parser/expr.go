package parser

import (
	"strconv"

	"github.com/peiblow/vvm/ast"
	"github.com/peiblow/vvm/lexer"
)

func parse_expr(p *parser, bp binding_power) ast.Expr {
	tokenType := p.currentTokenType()
	nud_fn, exists := nud_lu[tokenType]

	if !exists {
		panic("NUD handler expected for token")
	}

	left := nud_fn(p)

	for bp_lu[p.currentTokenType()] > bp {
		tokenType = p.currentTokenType()
		led_fn, exists := led_lu[tokenType]

		if !exists {
			panic("LED handler expected for token")
		}

		left = led_fn(p, left, bp)
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
	default:
		panic("Cannot create primary expr")
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
