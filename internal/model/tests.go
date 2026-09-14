package model

// TestResult represents the final result of a test.
type TestResult string

const (
	TestPassed  TestResult = "PASSED"
	TestFailed  TestResult = "FAILED"
	TestSkipped TestResult = "SKIPPED"
)

// Test serves as a record for an executed test case.
type Test struct {

	// Name refers to the name of the test function.
	Name string

	// Status indicates the test's final result.
	Status TestResult

	// ElapsedTime indicates the runtime in seconds.
	ElapsedTime float64

	// Output holds all logs produced by a test throughout the execution.
	Output string
}

// PackageMetrics tracks relevant statistics about the test results of a
// package.
type PackageMetrics struct {

	// TotalTests tracks the number of tests that belong to this package.
	TotalTests uint

	// Results registers the frequency of each result type (e.g., "pass",
	// "fail", "skip") within this package.
	Results map[TestResult]uint
}

// TestPackage groups tests which belong to the same package.
type TestPackage struct {

	// Metrics tracks the number of passing, failing and skipped tests in this
	// package.
	Metrics PackageMetrics

	// Path refers to the package's full path (ex: "github.com/user/repo/pkg").
	Path string

	// Tests holds all tests that belong to this package.
	Tests []*Test
}

// NewTestPackage initializes a new TestPackage instance, with an empty test list
// and clean state for test metrics.
func NewTestPackage(pkgName string) *TestPackage {
	return &TestPackage{
		Path:  pkgName,
		Tests: make([]*Test, 0),
		Metrics: PackageMetrics{
			Results: map[TestResult]uint{
				TestPassed:  0,
				TestFailed:  0,
				TestSkipped: 0,
			},
		},
	}
}

// InsertTest updates the test list with a new entry and updates the current
// metrics for this package instance.
func (pkg *TestPackage) InsertTest(test *Test) {
	pkg.Tests = append(pkg.Tests, test)
	pkg.Metrics.TotalTests++
	pkg.Metrics.Results[test.Status]++
}

// GetResultRate computes the rate of a result type over all tests registered
// in this package.
func (pkg *TestPackage) GetResultRate(resultType TestResult) float64 {
	return float64(pkg.Metrics.Results[resultType] / pkg.Metrics.TotalTests)
}

// TestReportData packs all data required to render the tests page.
type TestReportData struct {

	// Overview holds the high-level summary cards.
	Overview OverviewData

	// Packages holds per-package data for the "Tests by package" section.
	Packages []*PackageReport
}

// OverviewData holds the summary card values for the overview grid.
type OverviewData struct {

	// TotalRunned is the total number of executed tests.
	TotalRunned uint

	// Passed is the number of passed tests.
	Passed uint

	// Failed is the number of failed tests.
	Failed uint

	// Skipped is the number of skipped tests.
	Skipped uint
}

// PackageReport holds the per-package data for the "Tests by package"
// section.
type PackageReport struct {

	// Name is the short package name (e.g. "auth").
	Name string

	// TestCount is the number of tests in this package.
	TestCount uint

	// ElapsedTime is the package's execution time (e.g. "97ms").
	ElapsedTime string

	// Tests holds the individual test results within this package.
	Tests []*TestReport
}

// TestReport holds the rendering data of a single test.
type TestReport struct {

	// Name is the test function name.
	Name string

	// Passed indicates whether the test passed (PASS/FAIL badge).
	Passed bool

	// Output holds error output (shown in test-error-box), empty if passed.
	Output string
}
