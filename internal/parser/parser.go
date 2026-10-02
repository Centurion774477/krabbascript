package parser

import (
	"fmt"
	"kscript/internal/lexer"
	"strings"

	"github.com/fatih/color"
)

type TokenStream struct {
	toks []lexer.Token
	pos  int
}

type Parser struct {
	line, column int
	file         string

	toks      TokenStream
	numErrors int

	bindings map[lexer.TokenType]Binding
}

func NewParser(toks []lexer.Token, file string) Parser {
	return Parser{
		line:   1,
		column: 1,
		file:   file,

		toks: TokenStream{
			pos:  0,
			toks: toks,
		},
		numErrors: 0,
		bindings: map[lexer.TokenType]Binding{
			// Addition and subtraction (+,-)
			lexer.TokenPlus:  {BindingPrecedence: PrecedenceTerm, BindingSide: LeftSide},
			lexer.TokenMinus: {BindingPrecedence: PrecedenceTerm, BindingSide: LeftSide},
			// Multiplication and division (*,/)
			lexer.TokenMul: {BindingPrecedence: PrecedenceFactor, BindingSide: LeftSide},
			lexer.TokenDiv: {BindingPrecedence: PrecedenceFactor, BindingSide: LeftSide},
			// Compasion (==, !=, <, >, >=, <=)
			lexer.TokenEqEq:        {BindingPrecedence: PrecedenceComparison, BindingSide: LeftSide},
			lexer.TokenNotEq:       {BindingPrecedence: PrecedenceComparison, BindingSide: LeftSide},
			lexer.TokenLessThan:    {BindingPrecedence: PrecedenceComparison, BindingSide: LeftSide},
			lexer.TokenGreaterThan: {BindingPrecedence: PrecedenceComparison, BindingSide: LeftSide},
			lexer.TokenLessEq:      {BindingPrecedence: PrecedenceComparison, BindingSide: LeftSide},
			lexer.TokenGreaterEq:   {BindingPrecedence: PrecedenceComparison, BindingSide: LeftSide},
			// Bitwise (&,|,^)
			lexer.TokenBand: {BindingPrecedence: PrecedenceBitwise, BindingSide: LeftSide},
			lexer.TokenBor:  {BindingPrecedence: PrecedenceBitwise, BindingSide: LeftSide},
			lexer.TokenBxor: {BindingPrecedence: PrecedenceBitwise, BindingSide: LeftSide},
			// Assignment (=,+=,-=,/=,*=,&=,|=,^=)
			lexer.TokenEq:      {BindingPrecedence: PrecedenceAssignment, BindingSide: RightSide},
			lexer.TokenPlusEq:  {BindingPrecedence: PrecedenceAssignment, BindingSide: RightSide},
			lexer.TokenMinusEq: {BindingPrecedence: PrecedenceAssignment, BindingSide: RightSide},
			lexer.TokenDivEq:   {BindingPrecedence: PrecedenceAssignment, BindingSide: RightSide},
			lexer.TokenMulEq:   {BindingPrecedence: PrecedenceAssignment, BindingSide: RightSide},
			lexer.TokenBandEq:  {BindingPrecedence: PrecedenceAssignment, BindingSide: RightSide},
			lexer.TokenBorEq:   {BindingPrecedence: PrecedenceAssignment, BindingSide: RightSide},
			lexer.TokenBxorEq:  {BindingPrecedence: PrecedenceAssignment, BindingSide: RightSide},
			// Postfix ((), [],.)
			lexer.TokenOpenParen:   {BindingPrecedence: PrecedencePostfix, BindingSide: LeftSide},
			lexer.TokenOpenSqBrack: {BindingPrecedence: PrecedencePostfix, BindingSide: LeftSide},
			lexer.TokenDot:         {BindingPrecedence: PrecedencePostfix, BindingSide: LeftSide},
		},
	}
}

func (p *Parser) current() lexer.Token {
	return p.currentToks(&p.toks)
}

func (p *Parser) currentToks(s *TokenStream) lexer.Token {
	if s.pos >= len(s.toks) {
		return s.toks[len(s.toks)-1]
	}

	return s.toks[s.pos]
}

func (p *Parser) consume() lexer.Token {
	return p.consumeToks(&p.toks)
}

func (p *Parser) consumeToks(s *TokenStream) lexer.Token {
	t := s.toks[s.pos]
	s.pos++

	return t
}

func (p *Parser) expect(types ...lexer.TokenType) error {
	return p.expectToks(&p.toks, types...)
}

func (p *Parser) expectToks(s *TokenStream, types ...lexer.TokenType) error {
	t := p.currentToks(s)

	for _, tt := range types {
		if t.Type == tt {
			return nil
		}
	}

	expected := make([]string, len(types))
	for i, tt := range types {
		expected[i] = fmt.Sprintf("'%s'", tt)
	}

	return fmt.Errorf("%s:%d:%d: expected %s, got %s",
		p.file, t.Line, t.Column, strings.Join(expected, ", "), t.Type)
}

func (p *Parser) skip(types ...lexer.TokenType) error {
	return p.skipToks(&p.toks, types...)
}

func (p *Parser) skipToks(s *TokenStream, types ...lexer.TokenType) error {
	err := p.expectToks(s, types...)
	if err != nil {
		return err
	}

	p.consumeToks(s)
	return nil
}

func (p *Parser) parseExpressionWithMinBpToks(s *TokenStream, minBp int) (*Node, error) {
	err := p.expectToks(
		s,
		lexer.TokenNumLiteral,
		lexer.TokenOpenParen,
		lexer.TokenFloatLiteral,
		lexer.TokenLiteral,
		lexer.TokenStrLiteral) // Expect a str literal too since KrabbaScript supports string concatting
	var left *Node
	if err != nil {
		return nil, err
	}

	if p.currentToks(s).Type == lexer.TokenOpenParen {
		p.consumeToks(s)
		l, err := p.parseExpressionWithMinBpToks(s, 0)
		if err != nil {
			return nil, err
		}

		left = l // Since Go is such a bitch I have to do this instead, very ugly and stuff

		err = p.skipToks(s, lexer.TokenClosedParen)
		if err != nil {
			return nil, err
		}

	} else {
		num := p.consumeToks(s)
		left = &Node{
			line:   num.Line,
			column: num.Column,
			Type:   NodeNumLit,

			Lexeme: num.Value,
		}
	}

	for {
		op := p.currentToks(s)
		if op.Type == lexer.TokenEof {
			break
		}

		bp, ok := p.bindings[op.Type]
		if !ok {
			break
		}

		if int(bp.BindingPrecedence) <= minBp {
			break
		}

		p.consumeToks(s)
		nodeOp := &Node{
			line:   op.Line,
			column: op.Column,
			Type:   NodeBinOp,

			Lexeme: op.Type.String(),
		}

		var right *Node
		var err error

		if bp.BindingSide == RightSide {
			right, err = p.parseExpressionWithMinBpToks(s, int(bp.BindingPrecedence-1))
			if err != nil {
				return nil, err
			}
		} else {

			right, err = p.parseExpressionWithMinBpToks(s, int(bp.BindingPrecedence))
			if err != nil {
				return nil, err
			}
		}

		nodeOp.Left = left
		nodeOp.Right = right

		left = nodeOp
	}

	return left, nil
}

func (p *Parser) parseExpression() (*Node, error) {
	expr, err := p.parseExpressionWithMinBpToks(&p.toks, 0)
	return expr, err
}

func (p *Parser) convertType(typ lexer.Token) (*Node, error) {
	n := &Node{
		line:   typ.Line,
		column: typ.Column,
	}

	// Assign a type
	switch typ.Type {
	case lexer.TokenI64:
		n.Type = NodeI64Type
	case lexer.TokenI32:
		n.Type = NodeI32Type
	case lexer.TokenI16:
		n.Type = NodeI16Type
	case lexer.TokenI8:
		n.Type = NodeI8Type

	case lexer.TokenU64:
		n.Type = NodeU64Type
	case lexer.TokenU32:
		n.Type = NodeU32Type
	case lexer.TokenU16:
		n.Type = NodeU16Type
	case lexer.TokenU8:
		n.Type = NodeU8Type

	case lexer.TokenStr:
		n.Type = NodeStrType
	case lexer.TokenAny:
		n.Type = NodeAnyType
	case lexer.TokenBool:
		n.Type = NodeBoolType
	case lexer.TokenLiteral:
		n.Type = NodeLit
		n.Lexeme = typ.Value
	default:
		return nil, fmt.Errorf("expected a type, got %s", typ.Type)
	}

	return n, nil
}

func (p *Parser) parseVarToks(s *TokenStream) (*Node, error) {
	var node *Node

	start := p.consumeToks(s) // Will use this for the line, col fields
	err := p.expectToks(s, lexer.TokenLiteral)
	if err != nil {
		return nil, err
	}

	name := p.consumeToks(s) // Save this for later

	// Check if it's = or :
	// Since it can be:
	// var krabba = 12;
	// var krabba: i32 = 12;
	err = p.expectToks(s, lexer.TokenEq, lexer.TokenColon)
	if err != nil {
		return nil, err
	}

	colEq := p.consumeToks(s)
	if colEq.Type == lexer.TokenColon {
		typ := p.consumeToks(s) // Save the type

		_, err := p.convertType(typ)
		if err != nil {
			return nil, err
		}

		err = p.expectToks(s, lexer.TokenEq, lexer.TokenSemi)
		if err != nil {
			return nil, err
		}

		// Something like var krabba: i32;
		if p.currentToks(s).Type == lexer.TokenSemi {
			p.consumeToks(s) // Skip the semicolon

			nType, err := p.convertType(typ)
			if err != nil {
				return nil, err
			}

			node = &Node{
				line:   start.Line,
				column: start.Column,
				Type:   NodeVariableDec,
				Lexeme: name.Value,
				Left:   nType,
			}
			goto exit

		} else if p.currentToks(s).Type == lexer.TokenEq {
			p.consumeToks(s) // Skip the =

			expr, err := p.parseExpressionWithMinBpToks(s, 0)
			if err != nil {
				return nil, err
			}

			err = p.skipToks(s, lexer.TokenSemi)
			if err != nil {
				return nil, err
			}

			nType, err := p.convertType(typ)
			if err != nil {
				return nil, err
			}

			node = &Node{
				line:   start.Line,
				column: start.Column,
				Type:   NodeVariableDef,
				Lexeme: name.Value,
				Left:   nType,
				Right:  expr,
			}
			goto exit
		}
	} else { // Equals
		expr, err := p.parseExpressionWithMinBpToks(s, 0)
		if err != nil {
			return nil, err
		}

		err = p.skipToks(s, lexer.TokenSemi)
		if err != nil {
			return nil, err
		}

		node = &Node{
			line:   start.Line,
			column: start.Column,
			Type:   NodeVariableDef,
			Lexeme: name.Value,
			Right:  expr,
		}
		goto exit
	}

exit:
	return node, nil
}

func (p *Parser) parseValToks(s *TokenStream) (*Node, error) {
	var node *Node

	start := p.consumeToks(s) // Will use this for the line, col fields
	err := p.expectToks(s, lexer.TokenLiteral)
	if err != nil {
		return nil, err
	}

	name := p.consumeToks(s) // Save this for later

	// Check if it's = or :
	// Since it can be:
	// val krabba = 12;
	// val krabba: i32 = 12;
	err = p.expectToks(s, lexer.TokenEq, lexer.TokenColon)
	if err != nil {
		return nil, err
	}

	colEq := p.consumeToks(s)
	if colEq.Type == lexer.TokenColon {
		typ := p.consumeToks(s) // Save the type

		_, err := p.convertType(typ)
		if err != nil {
			return nil, err
		}

		err = p.expectToks(s, lexer.TokenEq)
		if err != nil {
			return nil, err
		}

		p.consumeToks(s) // Skip the =

		expr, err := p.parseExpressionWithMinBpToks(s, 0)
		if err != nil {
			return nil, err
		}

		err = p.skipToks(s, lexer.TokenSemi)
		if err != nil {
			return nil, err
		}

		nType, err := p.convertType(typ)
		if err != nil {
			return nil, err
		}

		node = &Node{
			line:   start.Line,
			column: start.Column,
			Type:   NodeValueDef,
			Lexeme: name.Value,
			Left:   nType,
			Right:  expr,
		}
		goto exit
	} else { // Equals
		expr, err := p.parseExpressionWithMinBpToks(s, 0)
		if err != nil {
			return nil, err
		}

		err = p.skipToks(s, lexer.TokenSemi)
		if err != nil {
			return nil, err
		}

		node = &Node{
			line:   start.Line,
			column: start.Column,
			Type:   NodeValueDef,
			Lexeme: name.Value,
			Right:  expr,
		}
		goto exit
	}

exit:
	return node, nil
}

func (p *Parser) parseScopeToks(s *TokenStream, what string) (*Node, error) {
	if err := p.skipToks(s, lexer.TokenOpenBrack); err != nil {
		return nil, err
	}

	var toks []lexer.Token
	depth := 0
	for {
		t := p.currentToks(s)
		if t.Type == lexer.TokenEof {
			break
		}

		if t.Type == lexer.TokenOpenBrack {
			depth++
		} else if t.Type == lexer.TokenClosedBrack {
			if depth == 0 {
				break
			}
			depth--
		}

		toks = append(toks, t)
		p.consumeToks(s)
	}

	if err := p.skipToks(s, lexer.TokenClosedBrack); err != nil {
		return nil, err
	}

	if len(toks) == 0 {
		// Nothing to parse...
		return nil, nil
	}

	last := toks[len(toks)-1]
	toks = append(toks, lexer.Token{
		Line:   last.Line,
		Column: last.Column,
		Type:   lexer.TokenEof,
	})

	n := p.ParseToks(&TokenStream{toks: toks, pos: 0})
	if n == nil && p.GetErrors() > 0 {
		return nil, fmt.Errorf("failed to parse %s body", what)
	}

	return n, nil
}

func (p *Parser) parseIfToks(s *TokenStream) (*Node, error) {
	start := p.consumeToks(s) // Will use this for the line, col fields

	expr, err := p.parseExpressionWithMinBpToks(s, 0)
	if err != nil {
		return nil, err
	}

	body, err := p.parseScopeToks(s, "if")
	if err != nil {
		return nil, err
	}

	node := &Node{
		line:      start.Line,
		column:    start.Column,
		Type:      NodeIfStatement,
		ExtraInfo: &NodeInfoBlock{},
		Left:      expr,
	}

	if body != nil {
		p.appendToBlock(node, body)
	}

	return node, nil
}

func (p *Parser) parseElsifToks(s *TokenStream) (*Node, error) {
	start := p.consumeToks(s)

	expr, err := p.parseExpressionWithMinBpToks(s, 0)
	if err != nil {
		return nil, err
	}

	body, err := p.parseScopeToks(s, "elsif")
	if err != nil {
		return nil, err
	}

	node := &Node{
		line:      start.Line,
		column:    start.Column,
		Type:      NodeElsifStatement,
		ExtraInfo: &NodeInfoBlock{},
		Left:      expr,
	}

	if body != nil {
		p.appendToBlock(node, body)
	}

	return node, nil
}

func (p *Parser) parseElseToks(s *TokenStream) (*Node, error) {
	start := p.consumeToks(s)

	body, err := p.parseScopeToks(s, "else")
	if err != nil {
		return nil, err
	}

	node := &Node{
		line:      start.Line,
		column:    start.Column,
		Type:      NodeElseStatement,
		ExtraInfo: &NodeInfoBlock{},
	}

	if body != nil {
		p.appendToBlock(node, body)
	}

	return node, nil
}

func (p *Parser) appendToBlock(block *Node, node *Node) {
	switch v := block.ExtraInfo.(type) {
	case *NodeInfoBlock:
		v.Elements = append(v.Elements, node)
	}
}

func (p *Parser) GetErrors() int {
	return p.numErrors
}

func (p *Parser) ParseToks(s *TokenStream) *Node {
	ast := &Node{
		Type:      NodeRoot,
		ExtraInfo: &NodeInfoBlock{},
	}

loop:
	for {
		t := p.currentToks(s)

		switch t.Type {
		case lexer.TokenEof:
			break loop

		case lexer.TokenVar:
			n, err := p.parseVarToks(s)
			if err != nil {
				fmt.Printf("kscript: %s: %v\n", color.RedString("error"), err)

				p.numErrors++
				return nil
			}

			p.appendToBlock(ast, n)
		case lexer.TokenVal:
			n, err := p.parseValToks(s)
			if err != nil {
				fmt.Printf("kscript: %s: %v\n", color.RedString("error"), err)

				p.numErrors++
				return nil
			}

			p.appendToBlock(ast, n)
		case lexer.TokenIf:
			n, err := p.parseIfToks(s)
			if err != nil {
				fmt.Printf("kscript: %s: %v\n", color.RedString("error"), err)

				p.numErrors++
				return nil
			}

			p.appendToBlock(ast, n)
		case lexer.TokenElsif:
			n, err := p.parseElsifToks(s)
			if err != nil {
				fmt.Printf("kscript: %s: %v\n", color.RedString("error"), err)

				p.numErrors++
				return nil
			}

			p.appendToBlock(ast, n)
		case lexer.TokenElse:
			n, err := p.parseElseToks(s)
			if err != nil {
				fmt.Printf("kscript: %s: %v\n", color.RedString("error"), err)

				p.numErrors++
				return nil
			}

			p.appendToBlock(ast, n)
		default:
			fmt.Printf("kscript: %s: %s:%d:%d: unknown token %s\n", color.RedString("error"), p.file, t.Line, t.Column, t.Type)
			p.numErrors++

			return nil
		}
	}

	return ast
}

func (p *Parser) Parse() *Node {
	return p.ParseToks(&p.toks)
}
