package formatter

import (
	"encoding/xml"
	"io"

	"github.com/lrills/helm-unittest/pkg/unittest/results"
)

// JUnitTestSuites is a collection of JUnit test suites.
type JUnitTestSuites struct {
	XMLName xml.Name         `xml:"testsuites"`
	Suites  []JUnitTestSuite `xml:"testsuite"`
}

// JUnitTestSuite is a single JUnit test suite which may contain many
// testcases.
type JUnitTestSuite struct {
	XMLName    xml.Name        `xml:"testsuite"`
	Tests      int             `xml:"tests,attr"`
	Failures   int             `xml:"failures,attr"`
	Time       string          `xml:"time,attr"`
	Name       string          `xml:"name,attr"`
	Properties []JUnitProperty `xml:"properties>property,omitempty"`
	TestCases  []JUnitTestCase `xml:"testcase"`
}

// JUnitTestCase is a single test case with its result.
type JUnitTestCase struct {
	XMLName     xml.Name          `xml:"testcase"`
	Classname   string            `xml:"classname,attr"`
	Name        string            `xml:"name,attr"`
	Time        string            `xml:"time,attr"`
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

// JUnitFailure contains data related to a failed test.
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
	for _, testSuiteResult := range testSuiteResults {
		ts := j.createJUnitTestSuite(testSuiteResult)

		// properties
		ts.Properties = append(ts.Properties, JUnitProperty{"helm-unittest.version", "1.6"})

		// individual test cases
		for _, test := range testSuiteResult.TestsResult {
			testCase := j.createJUnitTestCase(determineClassnameFromDisplayName(testSuiteResult.DisplayName), test)

			// Write when a test is failed
			if !test.Passed {
				ts.Failures++
				testCase.Failure = j.createJUnitFailure("Failed", "", test.Stringify())
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

func (j *jUnitReportXML) createJUnitTestSuite(testSuiteResult *results.TestSuiteResult) JUnitTestSuite {
	return JUnitTestSuite{
		Tests:      len(testSuiteResult.TestsResult),
		Failures:   0,
		Time:       formatDuration(testSuiteResult.CalculateTestSuiteDuration()),
		Name:       testSuiteResult.DisplayName,
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
