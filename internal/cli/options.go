package cli

import "github.com/spf13/cobra"

const defaultOutputDir = "./reports/"

// options holds the values of all CLI flags for the root command.
type options struct {
	outputDir string

	// coverageSrc is the path to a Go coverage profile (coverage.out). When
	// set, a coverage report is generated alongside the test report.
	coverageSrc string

	// coverageRoot is the module/source root used to resolve the source files
	// referenced by the coverage profile. Defaults to the working directory.
	coverageRoot string
}

// setCmdFlags configures the avaliable CLI flags for the provided
// cobra.Command instance.
func setCmdFlags(cmd *cobra.Command, opts *options) {
	cmd.Flags().StringVarP(
		&opts.outputDir,
		"output-dir",
		"o",
		opts.outputDir,
		"Output directory for HTML reports")

	cmd.Flags().StringVar(
		&opts.coverageSrc,
		"cov-prof",
		"",
		"Path to a Go coverage profile (coverage.out) to include a coverage report")

	cmd.Flags().StringVar(
		&opts.coverageRoot,
		"cov-root",
		"",
		"Module/source root used to resolve coverage files (default: current directory)")
}
