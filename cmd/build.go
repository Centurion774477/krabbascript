package cmd

import (
	"fmt"
	"kscript/internal/lexer"
	"kscript/internal/parser"
	"kscript/tools"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [directory/file]",
	Short: "Build a KrabbaScript project/file",

	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]

		f, err := tools.IsFile(dir)
		if err != nil {
			return err
		}

		if f == false {
			return fmt.Errorf("building directories is not supported")
		}

		l, _ := lexer.NewLexer(dir)
		toks := l.Scan()

		for _, tok := range toks {
			fmt.Println(tok)
		}

		if errs := l.GetErrors(); errs != 0 {
			fmt.Print("kscript: ")

			color.Set(color.FgRed)
			defer color.Unset()
			fmt.Printf("compilation failed with %d error(s)\n", errs)

			os.Exit(1)
		}

		p := parser.NewParser(toks, dir)
		ast := p.Parse()

		if errs := p.GetErrors(); errs != 0 {
			fmt.Print("kscript: ")

			color.Set(color.FgRed)
			defer color.Unset()
			fmt.Printf("compilation failed with %d error(s)\n", errs)

			os.Exit(1)
		}

		ast.Print()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
