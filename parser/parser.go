package parser

import (
	"github.com/peiblow/vvm/ast"
	"github.com/peiblow/vvm/lexer"
)

type parser struct {
	tokens []lexer.Token
	pos    int
}

func (p *parser) currentToken() lexer.Token {
	return p.tokens[p.pos]
}

func (p *parser) advance() lexer.Token {
	tk := p.currentToken()
	p.pos++
	return tk
}

func (p *parser) currentTokenType() lexer.TokenType {
	return p.tokens[p.pos].Type
}

func (p *parser) hasTokens() bool {
	return p.pos < len(p.tokens) && p.currentTokenType() != lexer.EOF
}

func (p *parser) expect(token lexer.TokenType) lexer.Token {
	tk := p.currentToken()
	p.pos++
	return tk
}

func createParser(tokens []lexer.Token) *parser {
	createTokenLookups()
	return &parser{tokens: tokens}
}

func Parse(tokens []lexer.Token) ast.BlockStmt {
	Body := make([]ast.Stmt, 0)
	p := createParser(tokens)

	for p.hasTokens() {
		Body = append(Body, parse_stmt(p))
	}

	return ast.BlockStmt{
		Body: Body,
	}
}
