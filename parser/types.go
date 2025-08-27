package parser

import (
	"fmt"

	"github.com/peiblow/vvm/ast"
	"github.com/peiblow/vvm/lexer"
)

type type_nud_handler func(p *parser) ast.Type
type type_led_handler func(p *parser, left ast.Type, bp binding_power) ast.Type

type type_nud_lookup map[lexer.TokenType]type_nud_handler
type type_led_lookup map[lexer.TokenType]type_led_handler
type type_bp_lookup map[lexer.TokenType]binding_power

var type_bp_lu = type_bp_lookup{}
var type_nud_lu = type_nud_lookup{}
var type_led_lu = type_led_lookup{}

func type_led(tp lexer.TokenType, bp binding_power, led_fn type_led_handler) {
	type_bp_lu[tp] = bp
	type_led_lu[tp] = led_fn
}

func type_nud(tp lexer.TokenType, nud_fn type_nud_handler) {
	type_nud_lu[tp] = nud_fn
}

func createTokenTypeLookups() {
	type_nud(lexer.IDENTIFIER, parse_symbol_type)
	type_nud(lexer.OPEN_BRACKET, parse_array_type)
}

func parse_type(p *parser, bp binding_power) ast.Type {
	tokenType := p.currentTokenType()
	nud_fn, exists := type_nud_lu[tokenType]

	if !exists {
		panic(fmt.Sprintf("TYPE_NUD handler expected for token %v - %v", tokenType, p.currentToken().Literal))
	}

	left := nud_fn(p)

	for type_bp_lu[p.currentTokenType()] > bp {
		tokenType = p.currentTokenType()
		led_fn, exists := type_led_lu[tokenType]

		if !exists {
			panic(fmt.Sprintf("TYPE_LED handler expected for token %v", tokenType))
		}

		left = led_fn(p, left, type_bp_lu[p.currentTokenType()])
	}

	return left
}

func parse_symbol_type(p *parser) ast.Type {
	return ast.SymbolType{
		Name: p.expect(lexer.IDENTIFIER).Literal,
	}
}

func parse_array_type(p *parser) ast.Type {
	p.advance()
	p.expect(lexer.CLOSE_BRACKET)
	var underyinType = parse_type(p, defalt_bp)
	return ast.ArrayType{
		Underlying: underyinType,
	}
}
