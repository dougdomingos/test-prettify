package cli

import (
	"fmt"

	"github.com/dougdomingos/test-prettify/internal/parser"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the main CLI command. It accepts the input file path
// either as an argument or from stdin.
func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test-prettify [path/to/file.json]",
		Short: "Format and generate visual reports from Go test output",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reader, cleanup, err := ResolveInput(args, cmd.InOrStdin())
			if err != nil {
				return err
			}
			defer cleanup()

			packages, err := parser.ParseTestJSON(reader)
			if err != nil {
				return err
			}

			fmt.Printf("Read %d packages:\n", len(packages))
			for _, pkg := range packages {
				fmt.Printf("\t=> %02d tests from %s\n", len(pkg.Tests), pkg.Path)
			}

			return nil
		},
	}
}

// Execute runs the CLI application.
func Execute() error {
	return NewRootCmd().Execute()
}
