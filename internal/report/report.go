package report

import (
	"fmt"
	"path"
	"sort"

	"github.com/dougdomingos/test-prettify/internal/model"
)

// BuildReport converts the parsed test packages into the report source of
// truth. Values that cannot be derived from `go test -json` (project name,
// timestamp, coverage) are left for the caller to fill in.
func BuildReport(packages map[string]*model.TestPackage) *model.ReportData {
	data := &model.ReportData{
		Sidebar: model.SidebarData{
			TestEntries:     make([]model.SidebarEntry, 0, len(packages)),
			CoverageEntries: make([]model.SidebarEntry, 0),
		},
	}

	names := make([]string, 0, len(packages))
	for name := range packages {
		names = append(names, name)
	}
	sort.Strings(names)

	var totalRunned, totalPassed, totalFailed, totalSkipped float64
	var totalElapsed float64
	pkgReports := make([]*model.PackageReport, 0, len(packages))

	for _, name := range names {
		pkg, ok := packages[name]
		if !ok || pkg == nil {
			continue
		}

		shortName := shortName(pkg.Path)
		var pkgElapsed, pkgPassed, pkgFailed float64

		tests := make([]*model.TestReport, 0, len(pkg.Tests))
		for _, t := range pkg.Tests {
			report := &model.TestReport{
				Name:   t.Name,
				Passed: t.Status == model.TestPassed,
				Output: t.Output,
			}
			tests = append(tests, report)

			pkgElapsed += t.ElapsedTime
			switch t.Status {
			case model.TestPassed:
				pkgPassed++
			case model.TestFailed:
				pkgFailed++
			}
		}

		totalElapsed += pkgElapsed
		totalRunned += float64(pkg.Metrics.TotalTests)
		totalPassed += pkgPassed
		totalFailed += pkgFailed
		totalSkipped += float64(pkg.Metrics.Results[model.TestSkipped])

		pkgReports = append(pkgReports, &model.PackageReport{
			Name:        shortName,
			TestCount:   pkg.Metrics.TotalTests,
			ElapsedTime: formatDuration(pkgElapsed),
			Tests:       tests,
		})

		data.Sidebar.TestEntries = append(data.Sidebar.TestEntries, model.SidebarEntry{
			Name:    shortName,
			Success: pkgFailed == 0,
			Value:   ratioString(pkgPassed, pkg.Metrics.TotalTests),
		})
	}

	data.Sidebar.TotalTime = formatDuration(totalElapsed)
	data.TestReport.Overview = model.OverviewData{
		TotalRunned: uint(totalRunned),
		Passed:      uint(totalPassed),
		Failed:      uint(totalFailed),
		Skipped:     uint(totalSkipped),
	}
	data.TestReport.Packages = pkgReports

	return data
}

// shortName returns the last segment of a package path, falling back to the
// full path when the base is not a usable name.
func shortName(pkgPath string) string {
	if pkgPath == "" {
		return ""
	}

	if base := path.Base(pkgPath); base != "." && base != "/" && base != "" {
		return base
	}

	return pkgPath
}

// ratioString renders the passed count over the total (ex: "10/10").
func ratioString(passed float64, total uint) string {
	return fmt.Sprintf("%d/%d", uint(passed), total)
}
