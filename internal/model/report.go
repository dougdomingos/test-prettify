package model

// ReportData is the source of truth for every report page. It holds the
// shared metadata (header, sidebar and global totals) alongside each
// page-specific report struct, which are added as new pages ship.
type ReportData struct {

	// ProjectName is the project's display name in the header.
	ProjectName string

	// Timestamp is the report generation time shown in the header.
	Timestamp string

	// Sidebar holds the data to render the sidebar.
	Sidebar SidebarData

	// TestReport packs the data required to render the tests page.
	TestReport TestReportData

	// CoverageReport packs the data required to render the coverage page.
	CoverageReport CoverageReportData
}

// SidebarData holds the data to render the sidebar (tests + coverage
// sections and footer totals).
type SidebarData struct {

	// TestEntries lists the per-package test results.
	TestEntries []SidebarEntry

	// CoverageEntries lists the per-package coverage results.
	CoverageEntries []SidebarEntry

	// TotalTime is the aggregate execution time (e.g. "203ms").
	TotalTime string

	// GlobalCoverage is the overall coverage percentage (e.g. 85.7).
	GlobalCoverage float64
}

// SidebarEntry represents a single item in the sidebar's Tests or Coverage
// list.
type SidebarEntry struct {

	// Name is the short package name (e.g. "auth", "api").
	Name string

	// Success marks the status dot color (true = green, false = red).
	Success bool

	// Grade is the coverage indicator color ("success", "warning" or
	// "danger"), set on coverage entries.
	Grade string

	// Value is the displayed value (e.g. "10/10" or "93.2%").
	Value string
}
