package helmtest

type TestJobResult struct {
	Passed        bool
	ExecError     error
	AssertsResult []AssertionResult
}
