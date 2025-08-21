package main

var binaryOperators = map[string]bool{
	"+":  true,
	"-":  true,
	"*":  true,
	"/":  true,
	"%":  true,
	"==": true,
	"!=": true,
	">":  true,
	"<":  true,
	">=": true,
	"<=": true,
}

func parseContract(tokens []Token, start int) (Contract, int) {
	ct := Contract{
		Name: tokens[start+1].Literal,
	}

	body, newI := parseBlock(tokens, start+3)
	ct.Body = body

	return ct, newI
}

func parseFunction(tokens []Token, start int) (Function, int) {
	fn := Function{
		Name:       tokens[start+1].Literal,
		ReturnType: tokens[start+3].Literal,
	}

	body, newI := parseBlock(tokens, start+5)
	fn.Body = body
	return fn, newI
}

func parseExpression(tokens []Token, start int, stopTokens map[string]bool) (Node, int) {
	if start >= len(tokens) {
		panic("unexpected end of tokens in parseExpression")
	}

	tok := tokens[start]
	i := start
	var left Node

	switch tok.Type {
	case NUMBER, STRING:
		left = Literal{Value: tok.Literal}
		i++
	case IDENT:
		if i+1 < len(tokens) && tokens[i+1].Literal == "=" {
			value, newI := parseExpression(tokens, i+2, stopTokens)
			assign := VarAssign{
				Name:  tok.Literal,
				Value: value,
			}
			return assign, newI
		}
		left = Literal{Value: tok.Literal}
		i++
	case SYMBOL:
		if tok.Literal == "(" {
			inner, newI := parseExpression(tokens, i+1, map[string]bool{")": true})
			if newI >= len(tokens) || tokens[newI].Literal != ")" {
				panic("expected closing )")
			}
			left = inner
			i = newI + 1
		} else {
			panic("unexpected symbol: " + tok.Literal)
		}
	default:
		panic("unsupported token in expression: " + tok.Literal)
	}

	// BinaryOp
	for i < len(tokens) {
		tok := tokens[i]

		if stopTokens != nil && stopTokens[tok.Literal] {
			break
		}

		if tok.Type == SYMBOL && binaryOperators[tok.Literal] {
			op := tok.Literal
			i++
			right, newI := parseExpression(tokens, i, stopTokens)
			left = BinaryOp{Left: left, Op: op, Right: right}
			i = newI
		} else {
			break
		}
	}

	return left, i
}

func parseIf(tokens []Token, start int) (IfStmt, int) {
	ifstmt := IfStmt{}

	if tokens[start+1].Literal != "(" {
		panic("expected ( after if")
	}

	condNode, i := parseExpression(tokens, start+2, map[string]bool{")": true})
	if i >= len(tokens) || tokens[i].Literal != ")" {
		panic("expected ) after if condition")
	}
	i++
	ifstmt.Condition = condNode

	thenBlock, i := parseBlock(tokens, i)
	ifstmt.Then = thenBlock

	if i < len(tokens) && tokens[i].Literal == "else" {
		i++
		elseBlock, newI := parseBlock(tokens, i)
		ifstmt.Else = elseBlock
		i = newI
	}

	return ifstmt, i
}

func parseBlock(tokens []Token, start int) ([]Node, int) {
	nodes := []Node{}
	i := start
	for i < len(tokens) {
		tok := tokens[i]

		if tok.Literal == "}" {
			return nodes, i + 1
		}

		switch tok.Type {
		case KEYWORD:
			if tok.Literal == "contract" {
				ct, newI := parseContract(tokens, i)
				nodes = append(nodes, ct)
				i = newI
				continue
			} else if tok.Literal == "func" {
				fn, newI := parseFunction(tokens, i)
				nodes = append(nodes, fn)
				i = newI
				continue
			} else if tok.Literal == "if" {
				ifstmt, newI := parseIf(tokens, i)
				nodes = append(nodes, ifstmt)
				i = newI
				continue
			}
		case IDENT:
			if i+1 < len(tokens) && tokens[i+1].Literal == "=" {
				assign := VarAssign{
					Name:  tok.Literal,
					Value: Literal{Value: tokens[i+2].Literal},
				}
				nodes = append(nodes, assign)
				i += 3
				continue
			}
		}

		i++
	}

	return nodes, i
}
