package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/dougdomingos/test-prettify/internal/coverage"
	"github.com/dougdomingos/test-prettify/internal/model"
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

			if opts.coverageSrc != "" {
				cov, err := loadCoverage(opts)
				if err != nil {
					return err
				}
				
				report.ApplyCoverage(data, cov)
			}

			fmt.Println("Generating HTML reports...")
			if err := r.GenerateReports(opts.outputDir, data); err != nil {
				return err
			}

			fmt.Printf("\nDone in %s. Output written to %s\n", time.Since(start).Round(time.Millisecond), opts.outputDir)
			return nil
		},
	}

	setCmdFlags(cmd, &opts)
	return cmd
}

// loadCoverage parses the coverage profile and renders the coverage report.
func loadCoverage(opts options) (*model.CoverageReportData, error) {
	f, err := os.Open(opts.coverageSrc)
	if err != nil {
		return nil, fmt.Errorf("opening coverage profile: %w", err)
	}
	defer f.Close()

	profile, err := parser.ParseCoverageProfile(f)
	if err != nil {
		return nil, fmt.Errorf("parsing coverage profile: %w", err)
	}

	cov, err := coverage.RenderProfile(profile, coverage.RenderOptions{RootModule: opts.coverageRoot})
	if err != nil {
		return nil, fmt.Errorf("rendering coverage report: %w", err)
	}

	return cov, nil
}

// Execute runs the CLI application.
func Execute() error {
	return NewRootCmd().Execute()
}
