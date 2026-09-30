package parser

type Precedence int

const (
	PrecedenceNone Precedence = iota // Little trick to make precedences start with 1
	PrecedenceAssignment
	PrecedenceComparison
	PrecedenceBitwise
	PrecedenceTerm
	PrecedenceFactor
	PrecedenceUnary
	PrecedencePostfix
)

const (
	LeftSide  Side = true
	RightSide Side = false
)

type Side bool

type Binding struct {
	BindingPrecedence Precedence
	BindingSide       Side
}
