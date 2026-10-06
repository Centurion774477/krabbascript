package lexer

import "fmt"

type TokenType int

const (
	// Keywords
	TokenVar TokenType = iota
	TokenVal
	TokenFunc
	TokenMod
	TokenRet

	TokenIf
	TokenElsif
	TokenElse

	TokenWhile
	TokenFor
	TokenLoop
	TokenRepeat
	TokenUntil
	TokenWhen
	TokenStruct
	TokenImport

	// Types
	TokenBool
	TokenStr

	TokenI64
	TokenI32
	TokenI16
	TokenI8

	TokenU64
	TokenU32
	TokenU16
	TokenU8

	TokenAny
	TokenArr

	// Symbols
	TokenPlus
	TokenPlusPlus
	TokenPlusEq

	TokenMinus
	TokenMinusMinus
	TokenMinusEq

	TokenMul // AKA TokenStar
	TokenMulEq

	TokenDiv
	TokenDivEq

	TokenEq
	TokenEqEq

	TokenNotEq
	TokenNot

	TokenBand
	TokenBor
	TokenBxor
	TokenBnot

	TokenBandEq
	TokenBorEq
	TokenBxorEq

	TokenArrow

	TokenComma
	TokenSemi
	TokenColon

	TokenOpenParen
	TokenClosedParen

	TokenOpenBrack
	TokenClosedBrack

	TokenOpenSqBrack
	TokenClosedSqBrack

	TokenDot
	TokenGreaterThan
	TokenLessThan

	TokenLessEq
	TokenGreaterEq

	// Compiler stuff
	TokenLiteral
	TokenStrLiteral
	TokenNumLiteral
	TokenFloatLiteral

	TokenNone
	TokenEof
)

func (tt TokenType) String() string {
	tokens := [...]string{
		// Keywords
		TokenVar:    "var",
		TokenVal:    "val",
		TokenFunc:   "func",
		TokenMod:    "mod",
		TokenRet:    "ret",
		TokenIf:     "if",
		TokenElsif:  "elsif",
		TokenElse:   "else",
		TokenWhile:  "while",
		TokenFor:    "for",
		TokenLoop:   "loop",
		TokenRepeat: "repeat",
		TokenUntil:  "until",
		TokenWhen:   "when",
		TokenStruct: "struct",
		TokenImport: "import",

		// Types
		TokenBool: "Bool",
		TokenStr:  "Str",
		TokenI64:  "I64",
		TokenI32:  "I32",
		TokenI16:  "I16",
		TokenI8:   "I8",
		TokenU64:  "U64",
		TokenU32:  "U32",
		TokenU16:  "U16",
		TokenU8:   "U8",
		TokenAny:  "Any",
		TokenArr:  "Arr",

		// Symbols
		TokenPlus:          "+",
		TokenPlusPlus:      "++",
		TokenPlusEq:        "+=",
		TokenMinus:         "-",
		TokenMinusMinus:    "--",
		TokenMinusEq:       "-=",
		TokenMul:           "*",
		TokenMulEq:         "*=",
		TokenDiv:           "/",
		TokenDivEq:         "/=",
		TokenEq:            "=",
		TokenEqEq:          "==",
		TokenNotEq:         "!=",
		TokenNot:           "!",
		TokenBand:          "&",
		TokenBor:           "|",
		TokenBxor:          "^",
		TokenBnot:          "~",
		TokenBandEq:        "&=",
		TokenBorEq:         "|=",
		TokenBxorEq:        "^=",
		TokenArrow:         "->",
		TokenComma:         ",",
		TokenSemi:          ";",
		TokenColon:         ":",
		TokenOpenParen:     "(",
		TokenClosedParen:   ")",
		TokenOpenBrack:     "{",
		TokenClosedBrack:   "}",
		TokenOpenSqBrack:   "[",
		TokenClosedSqBrack: "]",
		TokenDot:           ".",
		TokenLessThan:      "<",
		TokenGreaterThan:   ">",
		TokenLessEq:        "<=",
		TokenGreaterEq:     ">=",

		// Compiler stuff
		TokenLiteral:      "literal",
		TokenStrLiteral:   "str literal",
		TokenNumLiteral:   "num literal",
		TokenFloatLiteral: "float literal",
		TokenNone:         "none",
		TokenEof:          "eof",
	}

	if int(tt) >= 0 && int(tt) < len(tokens) {
		return tokens[tt]
	}
	return fmt.Sprintf("unknown token(%d)", tt)
}

type Token struct {
	Type         TokenType
	Line, Column int

	Value string
}

func (t Token) String() string {
	if t.Value != "" {
		return fmt.Sprintf("token %s with value %q at %d:%d", t.Type, t.Value, t.Line, t.Column)
	} else {
		return fmt.Sprintf("token %s at %d:%d", t.Type, t.Line, t.Column)
	}
}
