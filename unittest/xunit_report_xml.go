package unittest

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"
)

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
func (x *xUnitReportXML) WriteTestOutput(testSuiteResults []*TestSuiteResult, noXMLHeader bool, w io.Writer) error {
	currentTime := time.Now()
	testAssemblies := []XUnitAssembly{}

	// convert TestSuiteResults to NUnit test suites
	for _, testSuiteResult := range testSuiteResults {
		totalTime := formatDuration(testSuiteResult.calculateTestSuiteDuration())

		ts := XUnitAssembly{
			Name:          testSuiteResult.FilePath,
			ConfigFile:    testSuiteResult.FilePath,
			TestFramework: "helm-unittest",
			Environment:   fmt.Sprintf("%s.%s-%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
			RunDate:       formatDate(currentTime),
			RunTime:       formatTime(currentTime),
			Time:          totalTime,
			TotalTests:    0,
			PassedTests:   0,
			FailedTests:   0,
			SkippedTests:  0,
			ErrorsTests:   0,
		}

		classname := testSuiteResult.DisplayName
		if idx := strings.LastIndex(classname, "/"); idx > -1 && idx < len(testSuiteResult.DisplayName) {
			classname = testSuiteResult.DisplayName[idx+1:]
		}

		if testSuiteResult.ExecError != nil {
			ts.TotalTests++
			ts.ErrorsTests++
			ts.Errors = []XUnitError{
				{
					Type: "Error",
					Name: "Error",
					Failure: &XUnitFailure{
						ExceptionType: "Helm-Validation-Error",
						Message: &XUnitFailureMessage{
							Data: "Error",
						},
						StackTrace: &XUnitFailureStackTrace{
							Data: testSuiteResult.ExecError.Error(),
						},
					},
				},
			}
		} else {
			ts.TestRuns = []XUnitTestRun{
				{
					Name:         testSuiteResult.DisplayName,
					Time:         totalTime,
					TotalTests:   0,
					PassedTests:  0,
					FailedTests:  0,
					SkippedTests: 0,
					TestCases:    []XUnitTestCase{},
				},
			}

			// individual test cases
			for _, test := range testSuiteResult.TestsResult {
				ts.TotalTests++
				ts.TestRuns[0].TotalTests++

				testCase := XUnitTestCase{
					Name:    test.DisplayName,
					Type:    classname,
					Method:  "Helm-Validation",
					Time:    formatDuration(test.Duration),
					Result:  x.formatResult(test.Passed),
					Failure: nil,
				}

				// Write when a test is failed
				if !test.Passed {
					// Update error count
					if test.ExecError != nil {
						ts.ErrorsTests++
					} else {
						ts.FailedTests++
						ts.TestRuns[0].FailedTests++
					}

					testCase.Failure = &XUnitFailure{
						ExceptionType: "Helm-Validation",
						Message: &XUnitFailureMessage{
							Data: "Failed",
						},
						StackTrace: &XUnitFailureStackTrace{
							Data: test.stringify(),
						},
					}
				} else {
					ts.PassedTests++
					ts.TestRuns[0].PassedTests++
				}

				ts.TestRuns[0].TestCases = append(ts.TestRuns[0].TestCases, testCase)
			}
		}

		testAssemblies = append(testAssemblies, ts)
	}

	xunitResult := XUnitAssemblies{
		Assembly: testAssemblies,
	}

	// to xml
	bytes, err := xml.MarshalIndent(xunitResult, "", "\t")
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(w)

	if !noXMLHeader {
		writer.WriteString(xml.Header)
	}

	writer.Write(bytes)
	writer.WriteByte('\n')
	writer.Flush()

	return nil
}

func (x *xUnitReportXML) formatResult(b bool) string {
	if !b {
		return "Fail"
	}
	return "Pass"
}
