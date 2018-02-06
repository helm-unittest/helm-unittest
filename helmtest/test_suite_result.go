package helmtest

type TestSuiteResult struct {
	Passed      bool
	ExecError   error
	TestsResult []TestJobResult
}
