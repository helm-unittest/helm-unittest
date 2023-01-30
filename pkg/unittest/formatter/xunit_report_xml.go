package formatter

import (
	"encoding/xml"
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/lrills/helm-unittest/pkg/unittest/results"
)

// XUnitValidationMethod the default name for Helm XUnit validation.
const XUnitValidationMethod string = "Helm-Validation"

// XUnitAssemblies the top level of the document.
type XUnitAssemblies struct {
	XMLName  xml.Name        `xml:"assemblies"`
	Assembly []XUnitAssembly `xml:"assembly,omitempty"`
}

// XUnitAssembly is a run of a single test assembly.
type XUnitAssembly struct {
	XMLName       xml.Name       `xml:"assembly"`
	Name          string         `xml:"name,attr"`
	ConfigFile    string         `xml:"config-file,attr"`
	TestFramework string         `xml:"test-framework,attr"`
	Environment   string         `xml:"environment"`
	RunDate       string         `xml:"run-date,attr"`
	RunTime       string         `xml:"run-time,attr"`
	Time          string         `xml:"time,attr"`
	TotalTests    int            `xml:"total,attr"`
	PassedTests   int            `xml:"passed,attr"`
	FailedTests   int            `xml:"failed,attr"`
	SkippedTests  int            `xml:"skipped,attr"`
	ErrorsTests   int            `xml:"errors,attr"`
	Errors        []XUnitError   `xml:"errors>error,omitempty"`
	TestRuns      []XUnitTestRun `xml:"collection,omitempty"`
}

// XUnitTestRun is a single XUnit test suite which may contain many
// testcases.
type XUnitTestRun struct {
	XMLName      xml.Name        `xml:"collection"`
	Name         string          `xml:"name,attr"`
	Time         string          `xml:"time,attr"`
	TotalTests   int             `xml:"total,attr"`
	PassedTests  int             `xml:"passed,attr"`
	FailedTests  int             `xml:"failed,attr"`
	SkippedTests int             `xml:"skipped,attr"`
	TestCases    []XUnitTestCase `xml:"test"`
}

// XUnitErrors is a testsuitecategory
type XUnitErrors struct {
	XMLName xml.Name     `xml:"errors"`
	Errors  []XUnitError `xml:"error,omitempty"`
}

// XUnitError contains error information.
type XUnitError struct {
	Type    string        `xml:"type,attr"`
	Name    string        `xml:"name,attr"`
	Failure *XUnitFailure `xml:"failure"`
}

// XUnitFailure contains data related to a failed test.
type XUnitFailure struct {
	XMLName       xml.Name                `xml:"failure"`
	ExceptionType string                  `xml:"exception-type,attr"`
	Message       *XUnitFailureMessage    `xml:"message,omitempty"`
	StackTrace    *XUnitFailureStackTrace `xml:"stack-trace,omitempty"`
}

// XUnitFailureMessage contains the message of the failed test.
type XUnitFailureMessage struct {
	XMLName xml.Name `xml:"message"`
	Data    string   `xml:",cdata"`
}

// XUnitFailureStackTrace constains the stacktrace of the failed test.
type XUnitFailureStackTrace struct {
	XMLName xml.Name `xml:"stack-trace"`
	Data    string   `xml:",cdata"`
}

// XUnitTestCase is a single test case with its result.
type XUnitTestCase struct {
	XMLName xml.Name      `xml:"test"`
	Name    string        `xml:"name,attr"`
	Type    string        `xml:"type,attr"`
	Method  string        `xml:"method,attr"`
	Time    string        `xml:"time,attr"`
	Result  string        `xml:"result,attr"`
	Traits  []XUnitTrait  `xml:"traits>trait,omitempty"`
	Failure *XUnitFailure `xml:"failure,omitempty"`
	Reason  *XUnitReason  `xml:"reason,omitempty"`
}

// XUnitReason contains reason why a test is skipped.
type XUnitReason struct {
	XMLName xml.Name `xml:"reason"`
	Reason  string   `xml:",cdata"`
}

// XUnitTrait contains a name/value pair.
type XUnitTrait struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type xUnitReportXML struct{}

// NewXUnitReportXML Constructor
func NewXUnitReportXML() Formatter {
	return &xUnitReportXML{}
}

// XUnitReportXML writes a NUnit xml representation of the given report to w
// in the format described at https://github.com/nunit/docs/wiki/XML-Formats
func (x *xUnitReportXML) WriteTestOutput(testSuiteResults []*results.TestSuiteResult, noXMLHeader bool, w io.Writer) error {
	currentTime := time.Now()
	testAssemblies := []XUnitAssembly{}

	// convert TestSuiteResults to NUnit test suites
	for _, testSuiteResult := range testSuiteResults {
		ts := x.createXUnitAssembly(currentTime, testSuiteResult)

		// When ExecError found, direct create error and
		// add to the list and iterater trough next testSuiteResult.
		if testSuiteResult.ExecError != nil {
			ts.TotalTests++
			ts.ErrorsTests++
			ts.Errors = []XUnitError{
				x.createXUnitError(
					"Error",
					"Error",
					x.createXUnitFailure(
						fmt.Sprintf("%s-%s", XUnitValidationMethod, "Error"),
						"Error",
						testSuiteResult.ExecError.Error(),
					),
				),
			}

			testAssemblies = append(testAssemblies, ts)
			continue
		}

		ts.TestRuns = []XUnitTestRun{
			x.createXUnitTestRun(testSuiteResult),
		}

		// individual test cases
		for _, test := range testSuiteResult.TestsResult {
			ts.TotalTests++
			ts.TestRuns[0].TotalTests++

			testCase := x.createXUnitTestCase(determineClassnameFromDisplayName(testSuiteResult.DisplayName), test)

			// Write when a test is failed
			if !test.Passed {
				testCase.Failure = x.createXUnitFailure(XUnitValidationMethod, "Failed", test.Stringify())

				// Update error count and ExceptionType
				if test.ExecError != nil {
					ts.ErrorsTests++
					testCase.Failure.ExceptionType = fmt.Sprintf("%s-%s", XUnitValidationMethod, "Error")
				} else {
					ts.FailedTests++
					ts.TestRuns[0].FailedTests++
				}
			} else {
				ts.PassedTests++
				ts.TestRuns[0].PassedTests++
			}

			ts.TestRuns[0].TestCases = append(ts.TestRuns[0].TestCases, testCase)
		}

		testAssemblies = append(testAssemblies, ts)
	}

	xunitResult := XUnitAssemblies{
		Assembly: testAssemblies,
	}

	// to xml
	if err := writeContentToFile(noXMLHeader, xunitResult, w); err != nil {
		return err
	}

	return nil
}

func (x *xUnitReportXML) formatResult(b bool) string {
	if !b {
		return "Fail"
	}
	return "Pass"
}

func (x *xUnitReportXML) createXUnitAssembly(currentTime time.Time, testSuiteResult *results.TestSuiteResult) XUnitAssembly {
	return XUnitAssembly{
		Name:          testSuiteResult.FilePath,
		ConfigFile:    testSuiteResult.FilePath,
		TestFramework: testFramework,
		Environment:   fmt.Sprintf("%s.%s-%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
		RunDate:       formatDate(currentTime),
		RunTime:       formatTime(currentTime),
		Time:          formatDuration(testSuiteResult.CalculateTestSuiteDuration()),
		TotalTests:    0,
		PassedTests:   0,
		FailedTests:   0,
		SkippedTests:  0,
		ErrorsTests:   0,
	}
}

func (x *xUnitReportXML) createXUnitTestRun(testSuiteResult *results.TestSuiteResult) XUnitTestRun {
	return XUnitTestRun{
		Name:         testSuiteResult.DisplayName,
		Time:         formatDuration(testSuiteResult.CalculateTestSuiteDuration()),
		TotalTests:   0,
		PassedTests:  0,
		FailedTests:  0,
		SkippedTests: 0,
		TestCases:    []XUnitTestCase{},
	}
}

func (x *xUnitReportXML) createXUnitTestCase(className string, testJobResult *results.TestJobResult) XUnitTestCase {
	return XUnitTestCase{
		Name:    testJobResult.DisplayName,
		Type:    className,
		Method:  XUnitValidationMethod,
		Time:    formatDuration(testJobResult.Duration),
		Result:  x.formatResult(testJobResult.Passed),
		Failure: nil,
	}
}

func (x *xUnitReportXML) createXUnitFailure(exceptionType, failureMessage, stackTrace string) *XUnitFailure {
	return &XUnitFailure{
		ExceptionType: exceptionType,
		Message: &XUnitFailureMessage{
			Data: failureMessage,
		},
		StackTrace: &XUnitFailureStackTrace{
			Data: stackTrace,
		},
	}
}

func (x *xUnitReportXML) createXUnitError(errorType, errorName string, xunitFailure *XUnitFailure) XUnitError {
	return XUnitError{
		Type:    errorType,
		Name:    errorName,
		Failure: xunitFailure,
	}
}
