package helmtest

type AssertionResult struct {
	FailInfo   []string
	Passed     bool
	AssertType string
	CustomInfo string
}
