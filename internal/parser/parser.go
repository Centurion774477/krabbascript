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

func newBindings() map[lexer.TokenType]Binding {
	return map[lexer.TokenType]Binding{
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
	}
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
		bindings:  newBindings(),
	}
}

func (p *Parser) currentToks(s *TokenStream) lexer.Token {
	if s.pos >= len(s.toks) {
		if len(s.toks) == 0 {
			return lexer.Token{
				Line:   1,
				Column: 1,
				Type:   lexer.TokenEof,
			}
		}

		return s.toks[len(s.toks)-1]
	}

	return s.toks[s.pos]
}

func (p *Parser) peekToks(s *TokenStream) lexer.Token {
	if s.pos+1 >= len(s.toks) {
		if len(s.toks) == 0 {
			return lexer.Token{
				Line:   1,
				Column: 1,
				Type:   lexer.TokenEof,
			}
		}

		return s.toks[len(s.toks)-1]
	}

	return s.toks[s.pos+1]
}

func (p *Parser) consumeToks(s *TokenStream) lexer.Token {
	if s.pos >= len(s.toks) {
		if len(s.toks) == 0 {
			return lexer.Token{
				Line:   1,
				Column: 1,
				Type:   lexer.TokenEof,
			}
		}

		return s.toks[len(s.toks)-1]
	}

	t := s.toks[s.pos]
	s.pos++

	return t
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

func (p *Parser) expectLitToks(s *TokenStream) error {
	return p.expectToks(s, lexer.TokenLiteral, lexer.TokenFloatLiteral, lexer.TokenStrLiteral, lexer.TokenNumLiteral)
}

func (p *Parser) expectLitToksExtra(s *TokenStream, types ...lexer.TokenType) error {
	base := []lexer.TokenType{
		lexer.TokenLiteral,
		lexer.TokenFloatLiteral,
		lexer.TokenStrLiteral,
		lexer.TokenNumLiteral,
	}

	toks := append(base, types...)

	return p.expectToks(s, toks...)
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

	err := p.expectLitToksExtra(
		s,
		lexer.TokenOpenParen,
		lexer.TokenOpenBrack)
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

	} else if p.currentToks(s).Type == lexer.TokenOpenBrack {
		start := p.currentToks(s)

		var buffer []lexer.Token
		depth := 0

		// Collect everything inside {}
		for {
			t := p.currentToks(s)

			if t.Type == lexer.TokenEof {
				return nil, fmt.Errorf(
					"%s:%d:%d: unclosed bracket",
					p.file, start.Line, start.Column,
				)
			}

			if t.Type == lexer.TokenOpenBrack {
				depth++
			} else if t.Type == lexer.TokenClosedBrack {
				depth--
			}

			buffer = append(buffer, t)
			p.consumeToks(s)

			if depth == 0 {
				break
			}
		}

		buffer = append(buffer, lexer.Token{
			Line:   start.Line,
			Column: start.Column,
			Type:   lexer.TokenEof,
		})

		stream := TokenStream{
			toks: buffer,
			pos:  0,
		}

		trimmed, err := p.trimListInit(&stream)
		if err != nil {
			return nil, err
		}

		for i, stream := range trimmed {
			fmt.Printf("stream %d\n", i)
			for _, el := range stream.toks {
				fmt.Printf("	%s\n", el)
			}
		}

		left = &Node{
			line:      start.Line,
			column:    start.Column,
			Type:      NodeListInit,
			ExtraInfo: &NodeInfoBlock{},
		}

		for _, stream := range trimmed {
			node, err := p.parseExpressionWithMinBpToks(&stream, 0)
			if err != nil {
				return nil, err
			}

			p.appendToBlock(left, node)
		}
	} else if p.currentToks(s).Type == lexer.TokenLiteral && p.peekToks(s).Type == lexer.TokenOpenBrack {

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
	case lexer.TokenFloatLiteral:
		n.Type = NodeFloatLit
		n.Lexeme = typ.Value
	case lexer.TokenNumLiteral:
		n.Type = NodeNumLit
		n.Lexeme = typ.Value
	case lexer.TokenStrLiteral:
		n.Type = NodeStrLit
		n.Lexeme = typ.Value
	default:
		return nil, fmt.Errorf("%s:%d:%d: expected a type, got %s", p.file, typ.Line, typ.Column, typ.Type)
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
	err := p.skipToks(s, lexer.TokenOpenBrack)
	if err != nil {
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

	err = p.skipToks(s, lexer.TokenClosedBrack)
	if err != nil {
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

func (p *Parser) parseStructFieldToks(s *TokenStream) (*Node, error) {
	err := p.expectToks(s, lexer.TokenLiteral)
	if err != nil {
		return nil, err
	}

	name := p.consumeToks(s)

	err = p.skipToks(s, lexer.TokenColon)
	if err != nil {
		return nil, err
	}

	typ, err := p.convertType(p.currentToks(s))
	if err != nil {
		return nil, err
	}

	p.consumeToks(s)

	err = p.skipToks(s, lexer.TokenSemi)
	if err != nil {
		return nil, err
	}

	n := Node{
		line:   name.Line,
		column: name.Column,

		Type:   NodeFieldDec,
		Lexeme: name.Value,

		Left: typ,
	}

	return &n, nil
}

func (p *Parser) parseStructFieldsToks(s *TokenStream) (*Node, error) {
	node := &Node{
		ExtraInfo: &NodeInfoBlock{},
		Type:      NodeStructDec,
	}
	for {
		t := p.currentToks(s)
		if t.Type == lexer.TokenEof {
			return node, nil
		}

		n, err := p.parseStructFieldToks(s)
		if err != nil {
			return nil, err
		}

		p.appendToBlock(node, n)
	}
}

func (p *Parser) parseStructScopeToks(s *TokenStream) (*Node, error) {
	err := p.skipToks(s, lexer.TokenOpenBrack)
	if err != nil {
		return nil, err
	}

	var toks []lexer.Token
	for {
		t := p.currentToks(s)
		if t.Type == lexer.TokenEof || t.Type == lexer.TokenClosedBrack {
			break
		}

		toks = append(toks, t)
		p.consumeToks(s)
	}

	err = p.skipToks(s, lexer.TokenClosedBrack)
	if err != nil {
		return nil, err
	}

	for _, t := range toks {
		fmt.Println(t)
	}

	if len(toks) != 0 {

		last := toks[len(toks)-1]
		toks = append(toks, lexer.Token{
			Line:   last.Line,
			Column: last.Column,
			Type:   lexer.TokenEof,
		})
	}

	stream := TokenStream{
		toks: toks,
		pos:  0,
	}
	return p.parseStructFieldsToks(&stream)
}

func (p *Parser) parseStructToks(s *TokenStream) (*Node, error) {
	start := p.consumeToks(s) // Save this for line, col fields

	err := p.expectToks(s, lexer.TokenLiteral)
	if err != nil {
		return nil, err
	}

	name := p.consumeToks(s) // Save the name too

	n, err := p.parseStructScopeToks(s)
	if err != nil {
		return nil, err
	}

	if n == nil {
		return nil, nil
	}

	n.line = start.Line
	n.column = start.Column

	n.Lexeme = name.Value

	return n, nil
}

/*
Some code for trimming { 1, 2 }, 3 for example

Get current
Depth = 0

While current != EOF
    While Depth > 0 || current != , && current != EOF
        If current == {
            Depth++
        Else if current == }
            If Depth == 0 Error out
            Depth--

        Add to buffer
        Update current

    If Depth != 0 Error out

    Add buffer to the array of token streams
    Clear buffer

    If current == EOF break
    Else skip ,

Return trimmed
*/

func (p *Parser) trimListInit(s *TokenStream) ([]TokenStream, error) {
	var trimmed []TokenStream
	var buffer []lexer.Token
	depth := 0

	// Delete the { if present
	if p.currentToks(s).Type == lexer.TokenOpenBrack {
		p.consumeToks(s)
	}

	for {
		t := p.currentToks(s)

		if t.Type == lexer.TokenEof || (t.Type == lexer.TokenClosedBrack && depth == 0) {
			break
		}

		if t.Type == lexer.TokenComma && depth == 0 {
			p.consumeToks(s)

			start := p.currentToks(s) // Save this for error handling
			if start.Type == lexer.TokenClosedBrack {
				return nil, fmt.Errorf("%s:%d:%d: expected an element, got {", p.file, start.Line, start.Column)
			}
			if len(buffer) > 0 {
				buffer = append(buffer, lexer.Token{
					Line:   t.Line,
					Column: t.Column,
					Type:   lexer.TokenEof,
				})

				trimmed = append(trimmed, TokenStream{
					toks: buffer,
					pos:  0,
				})

				buffer = nil
			}

			continue
		}
		if t.Type == lexer.TokenOpenBrack {
			depth++
		} else if t.Type == lexer.TokenClosedBrack {
			depth--
			if depth < 0 {
				return nil, fmt.Errorf("%s:%d:%d: unexpected closing bracket", p.file, t.Line, t.Column)
			}
		}

		buffer = append(buffer, t)
		p.consumeToks(s)
	}

	if depth != 0 {
		t := p.currentToks(s)
		return nil, fmt.Errorf("%s:%d:%d: unclosed bracket", p.file, t.Line, t.Column)
	}

	if len(buffer) > 0 {
		buffer = append(buffer, lexer.Token{
			Line:   1,
			Column: 1,
			Type:   lexer.TokenEof,
		})

		trimmed = append(trimmed, TokenStream{
			toks: buffer,
			pos:  0,
		})
	}

	// Same thing for }
	if p.currentToks(s).Type == lexer.TokenClosedBrack {
		p.consumeToks(s)
	}

	return trimmed, nil
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

// Will advance the stream until a safe place to continue parsing is reached
func (p *Parser) sync(s *TokenStream) {
	p.consumeToks(s)

	for p.currentToks(s).Type != lexer.TokenEof {
		c := p.currentToks(s).Type
		if c == lexer.TokenSemi || c == lexer.TokenClosedBrack {
			p.consumeToks(s)
			return
		}

		switch p.currentToks(s).Type {
		case lexer.TokenVar, lexer.TokenVal, lexer.TokenIf,
			lexer.TokenElsif, lexer.TokenElse, lexer.TokenStruct:
			return
		}

		p.consumeToks(s)
	}
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
				p.sync(s)
			}

			p.appendToBlock(ast, n)
		case lexer.TokenVal:
			n, err := p.parseValToks(s)
			if err != nil {
				fmt.Printf("kscript: %s: %v\n", color.RedString("error"), err)

				p.numErrors++
				p.sync(s)
			}

			p.appendToBlock(ast, n)
		case lexer.TokenIf:
			n, err := p.parseIfToks(s)
			if err != nil {
				fmt.Printf("kscript: %s: %v\n", color.RedString("error"), err)

				p.numErrors++
				p.sync(s)
			}

			p.appendToBlock(ast, n)
		case lexer.TokenElsif:
			n, err := p.parseElsifToks(s)
			if err != nil {
				fmt.Printf("kscript: %s: %v\n", color.RedString("error"), err)

				p.numErrors++
				p.sync(s)
			}

			p.appendToBlock(ast, n)
		case lexer.TokenElse:
			n, err := p.parseElseToks(s)
			if err != nil {
				fmt.Printf("kscript: %s: %v\n", color.RedString("error"), err)

				p.numErrors++
				p.sync(s)
			}

			p.appendToBlock(ast, n)
		case lexer.TokenStruct:
			n, err := p.parseStructToks(s)
			if err != nil {
				fmt.Printf("kscript: %s: %v\n", color.RedString("error"), err)

				p.numErrors++
				p.sync(s)
			}

			p.appendToBlock(ast, n)
		default:
			fmt.Printf("kscript: %s: %s:%d:%d: unexpected token %s\n", color.RedString("error"), p.file, t.Line, t.Column, t.Type)
			p.numErrors++

			p.sync(s)
		}
	}

	return ast
}

func (p *Parser) Parse() *Node {
	return p.ParseToks(&p.toks)
}
