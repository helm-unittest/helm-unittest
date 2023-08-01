package formatter

import (
	"encoding/xml"
	"io"
	"os"
	"time"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
)

// JUnitTestSuites is a collection of JUnit test suites.
type JUnitTestSuites struct {
	XMLName xml.Name         `xml:"testsuites"`
	Suites  []JUnitTestSuite `xml:"testsuite"`
}

// JUnitTestSuite is a single JUnit test suite which may contain many
// testcases.
type JUnitTestSuite struct {
	XMLName     xml.Name        `xml:"testsuite"`
	Id          int             `xml:"id,attr"`
	Tests       int             `xml:"tests,attr"`
	Failures    int             `xml:"failures,attr"`
	Errors      int             `xml:"errors,attr"`
	Package     string          `xml:"package,attr"`
	Time        string          `xml:"time,attr"`
	Name        string          `xml:"name,attr"`
	Timestamp   string          `xml:"timestamp,attr"`
	Hostname    string          `xml:"hostname,attr"`
	Properties  []JUnitProperty `xml:"properties>property,omitempty"`
	TestCases   []JUnitTestCase `xml:"testcase"`
	SystemOut   string          `xml:"system-out,omitempty"`
	SystemError string          `xml:"system-err,omitempty"`
}

// JUnitTestCase is a single test case with its result.
type JUnitTestCase struct {
	XMLName     xml.Name          `xml:"testcase"`
	Classname   string            `xml:"classname,attr"`
	Name        string            `xml:"name,attr"`
	Time        string            `xml:"time,attr"`
	Error       *JUnitFailure     `xml:"error,omitempty"`
	SkipMessage *JUnitSkipMessage `xml:"skipped,omitempty"`
	Failure     *JUnitFailure     `xml:"failure,omitempty"`
}

// JUnitSkipMessage contains the reason why a testcase was skipped.
type JUnitSkipMessage struct {
	Message string `xml:"message,attr"`
}

// JUnitProperty represents a key/value pair used to define properties.
type JUnitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// JUnitFailure or error contains data related to a failed or errored test.
type JUnitFailure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

type jUnitReportXML struct{}

// NewJUnitReportXML Constructor
func NewJUnitReportXML() Formatter {
	return &jUnitReportXML{}
}

// JUnitReportXML writes a JUnit xml representation of the given report to w
// in the format described at http://windyroad.org/dl/Open%20Source/JUnit.xsd
func (j *jUnitReportXML) WriteTestOutput(testSuiteResults []*results.TestSuiteResult, noXMLHeader bool, w io.Writer) error {
	suites := JUnitTestSuites{}

	// convert TestSuiteResults to JUnit test suites
	for idx, testSuiteResult := range testSuiteResults {
		ts := j.createJUnitTestSuite(idx, testSuiteResult)

		// properties
		ts.Properties = append(ts.Properties, JUnitProperty{"helm-unittest.version", "1.6"})

		// individual test cases
		for _, test := range testSuiteResult.TestsResult {
			testCase := j.createJUnitTestCase(determineClassnameFromDisplayName(testSuiteResult.DisplayName), test)

			// Write when a test is failed
			if !test.Passed && test.ExecError == nil {
				ts.Failures++
				testCase.Failure = j.createJUnitFailure("Failed", "", test.Stringify())
			}

			if !test.Passed && test.ExecError != nil {
				ts.Errors++
				testCase.Error = j.createJUnitFailure("Error", "", test.ExecError.Error())
			}

			ts.TestCases = append(ts.TestCases, testCase)
		}

		suites.Suites = append(suites.Suites, ts)
	}

	// to xml
	if err := writeContentToFile(noXMLHeader, suites, w); err != nil {
		return err
	}

	return nil
}

func (j *jUnitReportXML) createJUnitTestSuite(idx int, testSuiteResult *results.TestSuiteResult) JUnitTestSuite {
	name, _ := os.Hostname()
	return JUnitTestSuite{
		Tests:      len(testSuiteResult.TestsResult),
		Id:         idx,
		Failures:   0,
		Errors:     0,
		Time:       formatDuration(testSuiteResult.CalculateTestSuiteDuration()),
		Timestamp:  formatDateTime(time.Now()),
		Name:       testSuiteResult.DisplayName,
		Hostname:   name,
		Properties: []JUnitProperty{},
		TestCases:  []JUnitTestCase{},
	}
}

func (j *jUnitReportXML) createJUnitTestCase(className string, testJobResult *results.TestJobResult) JUnitTestCase {
	return JUnitTestCase{
		Classname: className,
		Name:      testJobResult.DisplayName,
		Time:      formatDuration(testJobResult.Duration),
		Failure:   nil,
	}
}

func (j *jUnitReportXML) createJUnitFailure(message, failureType, content string) *JUnitFailure {
	return &JUnitFailure{
		Message:  message,
		Type:     failureType,
		Contents: content,
	}
}
