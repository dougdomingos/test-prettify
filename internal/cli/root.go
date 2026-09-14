package cli

import (
	"fmt"
	"time"

	"github.com/dougdomingos/test-prettify/internal/parser"
	"github.com/dougdomingos/test-prettify/internal/report"
	"github.com/dougdomingos/test-prettify/internal/templates"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the main CLI command. It accepts the input file path
// either as an argument or from stdin.
func NewRootCmd() *cobra.Command {
	opts := options{outputDir: defaultOutputDir}
	cmd := &cobra.Command{
		Use:   "test-prettify [path/to/file.json]",
		Short: "Format and generate visual reports from Go test output",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()

			reader, cleanup, err := ResolveInput(args, cmd.InOrStdin())
			if err != nil {
				return err
			}
			defer cleanup()

			fmt.Print("Parsing test events...\t")
			parseStart := time.Now()
			packages, err := parser.ParseTestJSON(reader)
			if err != nil {
				return err
			}

			totalTests := 0
			for _, pkg := range packages {
				totalTests += int(pkg.Metrics.TotalTests)
			}

			fmt.Printf("Parsed in %s: %d packages, %d tests\n",
				time.Since(parseStart).Round(time.Millisecond), len(packages), totalTests)

			for _, pkg := range packages {
				fmt.Printf("[*] %02d tests from %s\n", len(pkg.Tests), pkg.Path)
			}

			r, err := templates.NewPageRenderer()
			if err != nil {
				return err
			}

			data := report.BuildReport(packages)
			data.ProjectName = "Project Name"
			data.Timestamp = time.Now().Format("2006-01-02 15:04:05")

			fmt.Println("Generating HTML reports...")
			if err := r.GenerateReports(opts.outputDir, data); err != nil {
				return err
			}

			fmt.Printf("\nDone in %s. Output written to %s\n", time.Since(start).Round(time.Millisecond), opts.outputDir)
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.outputDir, "output-dir", "o", opts.outputDir, "Output directory for HTML reports")
	return cmd
}

// Execute runs the CLI application.
func Execute() error {
	return NewRootCmd().Execute()
}
