package model

import "html/template"

// Mode describes the coverage accounting mode declared in the profile header.
type Mode string

// Supported coverage modes. `set` is the default produced by `go test
// -coverprofile`. `count` and `atomic` behave like `set` for rendering
// purposes (any count greater than zero means the block was executed).
const (
	ModeSet    Mode = "set"
	ModeCount  Mode = "count"
	ModeAtomic Mode = "atomic"
)

// Block is a single record in a Go coverage profile. The position uses
// 1-based line and byte-column offsets; StartCol is inclusive and EndCol is
// exclusive (one past the last byte of the statement).
type Block struct {

	// FilePath is the import path of the source file, as recorded by the Go
	// toolchain (e.g. "example.com/mod/internal/foo/a.go").
	FilePath string

	// StartLine and StartCol locate the first byte of the block.
	StartLine int
	StartCol  int

	// EndLine and EndCol locate the byte right after the last byte of the
	// block.
	EndLine int
	EndCol  int

	// NumStmt is the number of statements attributed to this block.
	NumStmt int

	// Count is the number of times the block executed (mode-dependent).
	Count int
}

// Profile is the result of parsing a coverage profile. It retains the header
// mode plus the block records grouped by file.
type Profile struct {

	// Mode is the accounting mode declared in the header.
	Mode Mode

	// Blocks holds every parsed block in file order.
	Blocks []Block

	// BlocksByFile indexes blocks by their source file import path.
	BlocksByFile map[string][]Block
}

// LineStatus classifies how a source line is painted in the coverage viewer.
type LineStatus int

// Rendered source lines can be fully covered (hit), fully uncovered (miss),
// partially covered (mixed, shown via inline sub-line spans) or neutral when
// the line carries no executable statements.
const (
	LineNeutral LineStatus = iota
	LineCovered
	LineUncovered
	LineMixed
)

// CoveredLine holds the render data of a single source line in the coverage
// viewer.
type CoveredLine struct {

	// LineNum is the 1-based source line number.
	LineNum int

	// Status drives the row's background color.
	Status LineStatus

	// HTML is the escaped source text, with covered/uncovered segments wrapped
	// in inline highlight spans.
	HTML template.HTML
}

// CoverageFileReport packs the coverage data of a single source file.
type CoverageFileReport struct {

	// Path is the module-relative display path (e.g. "internal/auth/auth.go").
	Path string

	// Percent is the file's statement coverage percentage.
	Percent float64

	// Unavailable marks files whose source could not be located on disk, so
	// their code viewer is replaced by a notice.
	Unavailable bool

	// Lines holds the highlighted source lines for the code viewer.
	Lines []CoveredLine
}

// CoveragePackageReport groups the covered files that belong to the same
// source package.
type CoveragePackageReport struct {

	// Name is the short package name (e.g. "auth").
	Name string

	// Percent is the package's overall statement coverage percentage.
	Percent float64

	// Files lists the covered files within this package.
	Files []*CoverageFileReport
}

// CoverageFileSummary is a lightweight reference to a covered file, used by
// the lowest-coverage overview card.
type CoverageFileSummary struct {

	// Path is the module-relative display path (e.g. "internal/auth/auth.go").
	Path string

	// Percent is the file's statement coverage percentage.
	Percent float64
}

// CoverageOverview holds the global coverage summary cards.
type CoverageOverview struct {

	// GlobalPercent is the overall statement coverage percentage.
	GlobalPercent float64

	// CoveredStmts is the number of executed statements.
	CoveredStmts int

	// TotalStmts is the total number of statements measured.
	TotalStmts int

	// FileCount is the number of files included in the report.
	FileCount int

	// PackageCount is the number of packages included in the report.
	PackageCount int

	// LowestRate lists the files with the lowest coverage percentages, used
	// by the "Lowest coverage" overview card.
	LowestRate []CoverageFileSummary
}

// CoverageGrade buckets a coverage percentage into an indicator color name:
// "success" for c >= 75, "warning" for 50 <= c < 75 and "danger" for c < 50.
func CoverageGrade(percent float64) string {
	switch {
	case percent >= 75:
		return "success"
	case percent >= 50:
		return "warning"
	default:
		return "danger"
	}
}

// CoverageReportData packs all data required to render the coverage page.
type CoverageReportData struct {

	// Overview holds the high-level summary cards.
	Overview CoverageOverview

	// Packages holds per-package data for the coverage file cards.
	Packages []*CoveragePackageReport
}
