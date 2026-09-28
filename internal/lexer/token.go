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
	TokenElif
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

	// Compiler stuff
	TokenLiteral
	TokenStrLiteral
	TokenNumLiteral
	TokenFloatLiteral

	TokenNone
	TokenEOF
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
		TokenElif:   "elif",
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
		TokenBool: "bool",
		TokenStr:  "str",
		TokenI64:  "i64",
		TokenI32:  "i32",
		TokenI16:  "i16",
		TokenI8:   "i8",
		TokenU64:  "u64",
		TokenU32:  "u32",
		TokenU16:  "u16",
		TokenU8:   "u8",
		TokenAny:  "any",

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

		// Compiler stuff
		TokenLiteral:      "literal",
		TokenStrLiteral:   "str literal",
		TokenNumLiteral:   "num literal",
		TokenFloatLiteral: "float literal",
		TokenNone:         "none",
		TokenEOF:          "eof",
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
