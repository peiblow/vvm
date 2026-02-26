package lexer

import (
	"fmt"
	"regexp"
)

// ─────────────────────────────────────────────────────────────────────────────
// Handler types
// ─────────────────────────────────────────────────────────────────────────────

type regexHandler func(lex *lexer, regex *regexp.Regexp)

type regexPattern struct {
	regex   *regexp.Regexp
	handler regexHandler
}

// ─────────────────────────────────────────────────────────────────────────────
// Lexer struct
// ─────────────────────────────────────────────────────────────────────────────

type lexer struct {
	patterns []regexPattern
	tokens   []Token
	source   string
	pos      int
	line     int
	errors   []LexerError
}

type LexerError struct {
	Line    int
	Message string
}

func (e LexerError) Error() string {
	return fmt.Sprintf("[linha %d] %s", e.Line, e.Message)
}

// ─────────────────────────────────────────────────────────────────────────────
// Lexer helpers
// ─────────────────────────────────────────────────────────────────────────────

func (lex *lexer) advanceN(n int) {
	lex.pos += n
}

func (lex *lexer) push(token Token) {
	lex.tokens = append(lex.tokens, token)
}

func (lex *lexer) remainder() string {
	return lex.source[lex.pos:]
}

func (lex *lexer) at() byte {
	return lex.source[lex.pos]
}

func (lex *lexer) at_eof() bool {
	return lex.pos >= len(lex.source)
}

func (lex *lexer) addError(msg string) {
	lex.errors = append(lex.errors, LexerError{
		Line:    lex.line,
		Message: msg,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Handlers
// ─────────────────────────────────────────────────────────────────────────────
func defaultHandler(tp TokenType, value string) regexHandler {
	return func(lex *lexer, regex *regexp.Regexp) {
		lex.advanceN(len(value))
		lex.push(NewToken(tp, value))
	}
}

func skipHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindStringIndex(lex.remainder())
	lex.advanceN(match[1])
}

func newlineHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindStringIndex(lex.remainder())
	lex.advanceN(match[1])
	lex.line++
}

func lineCommentHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindStringIndex(lex.remainder())
	if match != nil {
		lex.advanceN(match[1])
	}
}

func blockCommentHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindStringIndex(lex.remainder())
	if match == nil {
		lex.addError("comentário de bloco /* não fechado")
		lex.advanceN(len(lex.remainder()))
		return
	}
	comment := lex.remainder()[match[0]:match[1]]
	for _, ch := range comment {
		if ch == '\n' {
			lex.line++
		}
	}
	lex.advanceN(match[1])
}

func numberHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindString(lex.remainder())
	lex.push(NewToken(NUMBER, match))
	lex.advanceN(len(match))
}

func hexNumberHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindString(lex.remainder())
	lex.push(NewToken(HEX_NUMBER, match))
	lex.advanceN(len(match))
}

func stringHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindStringIndex(lex.remainder())
	if match == nil {
		lex.addError("string não fechada")
		lex.advanceN(1)
		return
	}
	raw := lex.remainder()[match[0]:match[1]]
	literal := raw[1 : len(raw)-1]
	lex.push(NewToken(STRING, literal))
	lex.advanceN(len(raw))
}

func unclosedStringHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindString(lex.remainder())
	lex.addError(fmt.Sprintf("string não fechada: %s", match))
	lex.advanceN(len(match))
}

func symbolHandler(lex *lexer, regex *regexp.Regexp) {
	value := regex.FindString(lex.remainder())

	if tp, exists := reserved_lu[value]; exists {
		lex.push(NewToken(tp, value))
	} else {
		lex.push(NewToken(IDENTIFIER, value))
	}

	lex.advanceN(len(value))
}

// ─────────────────────────────────────────────────────────────────────────────
// Lexer factory
// ─────────────────────────────────────────────────────────────────────────────
func createLexer(input string) *lexer {
	return &lexer{
		pos:    0,
		line:   1,
		source: input,
		tokens: make([]Token, 0),
		errors: make([]LexerError, 0),

		// ── ORDER MATTERS ────────────────────────────────────────────────────
		// Longer/more specific patterns MUST appear before shorter ones.
		// Examples:
		//   `==`  before  `=`
		//   `<=`  before  `<`
		//   `//`  before  `/`   (comment before division operator)
		//   `??=` before  `?`
		//   `0x`  before  `[0-9]`  (hex before decimal)
		// ─────────────────────────────────────────────────────────────────────
		patterns: []regexPattern{
			{regexp.MustCompile(`\n`), newlineHandler},
			{regexp.MustCompile(`[ \t\r]+`), skipHandler},
			{regexp.MustCompile(`\/\/[^\n]*`), lineCommentHandler},
			{regexp.MustCompile(`\/\*[\s\S]*?\*\/`), blockCommentHandler},
			{regexp.MustCompile(`"[^"\n]*"`), stringHandler},
			{regexp.MustCompile(`"[^"\n]*`), unclosedStringHandler},
			{regexp.MustCompile(`0[xX][0-9a-fA-F]+`), hexNumberHandler},
			{regexp.MustCompile(`[0-9]+(\.[0-9]+)?`), numberHandler},
			{regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`), symbolHandler},
			{regexp.MustCompile(`\?\?=`), defaultHandler(NULLISH_ASSIGNMENT, "??=")},
			{regexp.MustCompile(`\+\+`), defaultHandler(PLUS_PLUS, "++")},
			{regexp.MustCompile(`--`), defaultHandler(MINUS_MINUS, "--")},
			{regexp.MustCompile(`\+=`), defaultHandler(PLUS_EQUALS, "+=")},
			{regexp.MustCompile(`-=`), defaultHandler(MINUS_EQUALS, "-=")},
			{regexp.MustCompile(`==`), defaultHandler(EQUALS, "==")},
			{regexp.MustCompile(`!=`), defaultHandler(NOT_EQUALS, "!=")},
			{regexp.MustCompile(`<=`), defaultHandler(LESS_EQUALS, "<=")},
			{regexp.MustCompile(`>=`), defaultHandler(GREATER_EQUALS, ">=")},
			{regexp.MustCompile(`\|\|`), defaultHandler(OR, "||")},
			{regexp.MustCompile(`&&`), defaultHandler(AND, "&&")},
			{regexp.MustCompile(`\.\.`), defaultHandler(DOT_DOT, "..")},
			{regexp.MustCompile(`=`), defaultHandler(ASSIGNMENT, "=")},
			{regexp.MustCompile(`!`), defaultHandler(NOT, "!")},
			{regexp.MustCompile(`<`), defaultHandler(LESS, "<")},
			{regexp.MustCompile(`>`), defaultHandler(GREATER, ">")},
			{regexp.MustCompile(`\?`), defaultHandler(QUESTION, "?")},
			{regexp.MustCompile(`\.`), defaultHandler(DOT, ".")},
			{regexp.MustCompile(`\+`), defaultHandler(PLUS, "+")},
			{regexp.MustCompile(`-`), defaultHandler(DASH, "-")},
			{regexp.MustCompile(`/`), defaultHandler(SLASH, "/")},
			{regexp.MustCompile(`\*`), defaultHandler(STAR, "*")},
			{regexp.MustCompile(`%`), defaultHandler(PERCENT, "%")},
			{regexp.MustCompile(`\[`), defaultHandler(OPEN_BRACKET, "[")},
			{regexp.MustCompile(`\]`), defaultHandler(CLOSE_BRACKET, "]")},
			{regexp.MustCompile(`\{`), defaultHandler(OPEN_CURLY, "{")},
			{regexp.MustCompile(`\}`), defaultHandler(CLOSE_CURLY, "}")},
			{regexp.MustCompile(`\(`), defaultHandler(OPEN_PAREN, "(")},
			{regexp.MustCompile(`\)`), defaultHandler(CLOSE_PAREN, ")")},
			{regexp.MustCompile(`;`), defaultHandler(SEMI_COLON, ";")},
			{regexp.MustCompile(`:`), defaultHandler(COLON, ":")},
			{regexp.MustCompile(`,`), defaultHandler(COMMA, ",")},
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Public API
// ─────────────────────────────────────────────────────────────────────────────
type TokenizeResult struct {
	Tokens []Token
	Errors []LexerError
}

func (r TokenizeResult) HasErrors() bool {
	return len(r.Errors) > 0
}

func (r TokenizeResult) PrintErrors() {
	for _, e := range r.Errors {
		fmt.Println(e.Error())
	}
}

func Tokenize(input string) TokenizeResult {
	lex := createLexer(input)

	for !lex.at_eof() {
		matched := false

		for _, pattern := range lex.patterns {
			loc := pattern.regex.FindStringIndex(lex.remainder())

			if loc != nil && loc[0] == 0 {
				pattern.handler(lex, pattern.regex)
				matched = true
				break
			}
		}

		if !matched {
			badChar := string(lex.at())
			lex.addError(fmt.Sprintf("caractere não reconhecido: '%s'", badChar))
			lex.advanceN(1)
		}
	}

	lex.push(NewToken(EOF, "EOF"))

	return TokenizeResult{
		Tokens: lex.tokens,
		Errors: lex.errors,
	}
}
