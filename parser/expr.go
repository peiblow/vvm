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

func parse_literal_array_expr(p *parser) ast.Expr {
	p.expect(lexer.OPEN_BRACKET)

	args := []ast.Expr{}
	for p.currentTokenType() != lexer.CLOSE_BRACKET {
		arg := parse_expr(p, defalt_bp)
		args = append(args, arg)

		if p.currentTokenType() == lexer.COMMA {
			p.advance()
		}
	}

	p.expect(lexer.CLOSE_BRACKET)
	return ast.ArrayLiteralExpr{
		Items: args,
	}
}

func parse_array_access_item_expr(p *parser, identifier ast.Expr, bp binding_power) ast.Expr {
	p.advance()
	index := parse_expr(p, defalt_bp)
	p.expect(lexer.CLOSE_BRACKET)

	return ast.ArrayAccessItemExpr{
		Array: identifier,
		Index: index,
	}
}

func parse_obj_item_assignment_expr(p *parser) ast.ObjectPropertyExpr {
	key := parse_expr(p, defalt_bp)
	p.expect(lexer.COLON)

	value := parse_expr(p, defalt_bp)
	p.expect(lexer.COMMA)

	return ast.ObjectPropertyExpr{
		Key:   key,
		Value: value,
	}
}

func parse_obj_assignment_expr(p *parser) ast.Expr {
	p.expect(lexer.OPEN_CURLY)

	fields := make([]ast.ObjectPropertyExpr, 0)
	for p.currentTokenType() != lexer.CLOSE_CURLY {
		prop := parse_obj_item_assignment_expr(p)
		fields = append(fields, prop)

		if p.currentToken().Type == lexer.COMMA {
			p.advance()
		}
	}

	p.expect(lexer.CLOSE_CURLY)
	return ast.ObjectAssignmentExpr{
		Name:   nil,
		Fields: fields,
	}
}

func parse_member_expr(p *parser, callee ast.Expr, bp binding_power) ast.Expr {
	p.expect(lexer.DOT)

	prop := parse_expr(p, defalt_bp)

	return ast.MemberExpr{
		Object:   callee,
		Property: prop,
	}
}

func parse_call_expr(p *parser, callee ast.Expr, bp binding_power) ast.Expr {
	p.expect(lexer.OPEN_PAREN)

	args := []ast.Expr{}
	for p.currentTokenType() != lexer.CLOSE_PAREN {
		arg := parse_expr(p, defalt_bp)
		args = append(args, arg)

		if p.currentTokenType() == lexer.COMMA {
			p.advance()
		}
	}
	p.expect(lexer.CLOSE_PAREN)

	return ast.CallExpr{
		Calle:     callee,
		Arguments: args,
	}
}
