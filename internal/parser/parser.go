package parser

import (
	"fmt"
	"kscript/internal/lexer"
	"strings"

	"github.com/fatih/color"
)

type Parser struct {
	line, column, pos int
	file              string

	toks      []lexer.Token
	numErrors int

	bindings map[lexer.TokenType]Binding
}

func NewParser(toks []lexer.Token, file string) Parser {
	return Parser{
		line:   1,
		column: 1,
		pos:    0,
		file:   file,

		toks:      toks,
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
	if p.pos >= len(p.toks) {
		return p.toks[len(p.toks)-1]
	}

	return p.toks[p.pos]
}

func (p *Parser) consume() lexer.Token {
	t := p.toks[p.pos]
	p.pos++

	fmt.Printf("consumed %s\n", t)
	return t
}

func (p *Parser) expect(types ...lexer.TokenType) error {
	t := p.current()

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
	err := p.expect(types...)
	if err != nil {
		return err
	}

	p.consume()
	return nil
}

func (p *Parser) parseExpressionWithMinBp(minBp int) (*Node, error) {
	err := p.expect(lexer.TokenNumLiteral, lexer.TokenOpenParen, lexer.TokenFloatLiteral, lexer.TokenLiteral)
	var left *Node
	if err != nil {
		return nil, err
	}

	if p.current().Type == lexer.TokenOpenParen {
		p.consume()
		l, err := p.parseExpressionWithMinBp(0)
		if err != nil {
			return nil, err
		}

		left = l // Since Go is such a bitch I have to do this instead, very ugly and stuff

		err = p.skip(lexer.TokenClosedParen)
		if err != nil {
			return nil, err
		}

	} else {
		num := p.consume()
		left = &Node{
			line:   num.Line,
			column: num.Column,
			Type:   NodeNumLit,

			Lexeme: num.Value,
		}
	}

	for {
		op := p.current()
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

		p.consume()
		nodeOp := &Node{
			line:   op.Line,
			column: op.Column,
			Type:   NodeBinOp,

			Lexeme: op.Type.String(),
		}

		var right *Node
		var err error

		if bp.BindingSide == RightSide {
			right, err = p.parseExpressionWithMinBp(int(bp.BindingPrecedence - 1))
			if err != nil {
				return nil, err
			}
		} else {

			right, err = p.parseExpressionWithMinBp(int(bp.BindingPrecedence))
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
	expr, err := p.parseExpressionWithMinBp(0)
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

func (p *Parser) parseVar() (*Node, error) {
	var node *Node

	start := p.consume() // Will use this for the line, col fields
	err := p.expect(lexer.TokenLiteral)
	if err != nil {
		return nil, err
	}

	name := p.consume() // Save this for later

	// Check if it's = or :
	// Since it can be:
	// var krabba = 12;
	// var krabba: i32 = 12;
	err = p.expect(lexer.TokenEq, lexer.TokenColon)
	if err != nil {
		return nil, err
	}

	colEq := p.consume()
	if colEq.Type == lexer.TokenColon {
		typ := p.consume() // Save the type

		_, err := p.convertType(typ)
		if err != nil {
			return nil, err
		}

		err = p.expect(lexer.TokenEq, lexer.TokenSemi)
		if err != nil {
			return nil, err
		}

		// Something like var krabba: i32;
		if p.current().Type == lexer.TokenSemi {
			p.consume() // Skip the semicolon

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

		} else if p.current().Type == lexer.TokenEq {
			p.consume() // Skip the =

			expr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			err = p.skip(lexer.TokenSemi)
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
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		err = p.skip(lexer.TokenSemi)
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

func (p *Parser) parseVal() (*Node, error) {
	var node *Node

	start := p.consume() // Will use this for the line, col fields
	err := p.expect(lexer.TokenLiteral)
	if err != nil {
		return nil, err
	}

	name := p.consume() // Save this for later

	// Check if it's = or :
	// Since it can be:
	// val krabba = 12;
	// val krabba: i32 = 12;
	err = p.expect(lexer.TokenEq, lexer.TokenColon)
	if err != nil {
		return nil, err
	}

	colEq := p.consume()
	if colEq.Type == lexer.TokenColon {
		typ := p.consume() // Save the type

		_, err := p.convertType(typ)
		if err != nil {
			return nil, err
		}

		err = p.expect(lexer.TokenEq)
		if err != nil {
			return nil, err
		}

		p.consume() // Skip the =

		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		err = p.skip(lexer.TokenSemi)
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
	} else { // Equals
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		err = p.skip(lexer.TokenSemi)
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

func (p *Parser) appendToBlock(block *Node, node *Node) {
	switch v := block.ExtraInfo.(type) {
	case *NodeInfoBlock:
		v.Elements = append(v.Elements, node)
	}
}

func (p *Parser) GetErrors() int {
	return p.numErrors
}

func (p *Parser) Parse() *Node {
	ast := &Node{
		Type:      NodeRoot,
		ExtraInfo: &NodeInfoBlock{},
	}

loop:
	for {
		t := p.current()

		switch t.Type {
		case lexer.TokenEof:
			break loop

		case lexer.TokenVar:
			n, err := p.parseVar()
			if err != nil {
				fmt.Printf("kscript: %s: %v\n", color.RedString("error"), err)

				p.numErrors++
				return nil
			}

			p.appendToBlock(ast, n)
		case lexer.TokenVal:
			n, err := p.parseVal()
			if err != nil {
				fmt.Printf("kscript: %s: %v\n", color.RedString("error"), err)

				p.numErrors++
				return nil
			}

			p.appendToBlock(ast, n)
		}
	}

	return ast
}
