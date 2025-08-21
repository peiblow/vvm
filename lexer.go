package main

import (
	"strings"
	"unicode"
)

type TokenType string

const (
	IDENT   TokenType = "IDENT"
	NUMBER  TokenType = "NUMBER"
	STRING  TokenType = "STRING"
	KEYWORD TokenType = "KEYWORD"
	SYMBOL  TokenType = "SYMBOL"
)

type Token struct {
	Type    TokenType
	Literal string
}

func tokenize(input string) Node {
	var tokens []Token
	var current strings.Builder

	isKeyword := func(word string) bool {
		keywords := []string{"contract", "func", "if", "else", "for", "int", "string", "print"}

		for _, kw := range keywords {
			if kw == word {
				return true
			}
		}

		return false
	}

	isNumber := func(num string) bool {
		for _, n := range num {
			if !unicode.IsDigit(n) {
				return false
			}
		}
		return len(num) > 0
	}

	flush := func() {
		if current.Len() > 0 {
			word := current.String()
			if isKeyword(word) {
				tokens = append(tokens, Token{Type: KEYWORD, Literal: word})
			} else if isNumber(word) {
				tokens = append(tokens, Token{Type: NUMBER, Literal: word})
			} else {
				tokens = append(tokens, Token{Type: IDENT, Literal: word})
			}
			current.Reset()
		}
	}

	for _, r := range input {
		switch {
		case unicode.IsLetter(r):
			current.WriteRune(r)
		case unicode.IsDigit(r):
			current.WriteRune(r)
		case unicode.IsSpace(r):
			flush()
		case r == '"' || r == '\'':
			flush()
			current.Reset()
		case strings.ContainsRune("{}();=,+-<>[]", r):
			flush()
			tokens = append(tokens, Token{Type: SYMBOL, Literal: string(r)})
		default:
			flush()
		}
	}

	flush()
	contract, _ := parseBlock(tokens, 0)

	return contract
}
