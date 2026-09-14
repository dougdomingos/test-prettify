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

// TestPackage groups tests which belong to the same package.
type TestPackage struct {
	
	// Path refers to the package's full path (ex: "github.com/user/repo/pkg").
	Path string
	
	// Tests holds all tests that belong to this package.
	Tests []*Test
}