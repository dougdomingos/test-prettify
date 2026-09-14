package cli

const defaultOutputDir = "./reports/"

// options holds the values of all CLI flags for the root command.
type options struct {
	outputDir string
}