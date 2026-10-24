package parser

import (
	"fmt"

	"github.com/fatih/color"
)

type ParserError struct {
	What error
	Fix  string
}

func (e ParserError) String() string {
	if e.Fix != "" {
		return fmt.Sprintf("%s\n\t%s: %s", e.What, color.BlueString("fix"), e.Fix)
	} else {
		return fmt.Sprintf("%s", e.What)
	}
}
