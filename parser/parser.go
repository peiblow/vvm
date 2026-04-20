package parser

import (
	"fmt"

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
	return p.expectError(token, nil)
}

func (p *parser) expectError(expectedType lexer.TokenType, err any) lexer.Token {
	token := p.currentToken()
	tokenType := token.Type

	if tokenType != expectedType {
		if err == nil {
			err = fmt.Sprintf("Expected %s but received %s instead\n", lexer.TokenTypeString(expectedType), lexer.TokenTypeString(tokenType))
			fmt.Println(token)
		}

		panic(err)
	}

	return p.advance()
}

// expectIdentifierOrKeyword consumes the current token if it is an IDENTIFIER
// or any reserved keyword, returning its literal as a name. This allows
// keywords like "nonce" or "hash" to be used as field/property names.
func (p *parser) expectIdentifierOrKeyword(errMsg string) string {
	token := p.currentToken()
	if token.Type == lexer.IDENTIFIER || lexer.IsKeyword(token.Type) {
		p.advance()
		return token.Literal
	}
	panic(errMsg)
}

func createParser(tokens []lexer.Token) *parser {
	createTokenLookups()
	createTokenTypeLookups()

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
