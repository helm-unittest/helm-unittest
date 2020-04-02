package unittest

import (
	"bufio"
	"encoding/xml"
	"io"
	"strings"
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
func (j *jUnitReportXML) WriteTestOutput(testSuiteResults []*TestSuiteResult, noXMLHeader bool, w io.Writer) error {
	suites := JUnitTestSuites{}

	// convert TestSuiteResults to JUnit test suites
	for _, testSuiteResult := range testSuiteResults {
		ts := JUnitTestSuite{
			Tests:      len(testSuiteResult.TestsResult),
			Failures:   0,
			Time:       formatDuration(testSuiteResult.calculateTestSuiteDuration()),
			Name:       testSuiteResult.DisplayName,
			Properties: []JUnitProperty{},
			TestCases:  []JUnitTestCase{},
		}

		classname := testSuiteResult.DisplayName
		if idx := strings.LastIndex(classname, "/"); idx > -1 && idx < len(testSuiteResult.DisplayName) {
			classname = testSuiteResult.DisplayName[idx+1:]
		}

		// properties
		ts.Properties = append(ts.Properties, JUnitProperty{"helm-unittest.version", "1.6"})

		// individual test cases
		for _, test := range testSuiteResult.TestsResult {
			testCase := JUnitTestCase{
				Classname: classname,
				Name:      test.DisplayName,
				Time:      formatDuration(test.Duration),
				Failure:   nil,
			}

			// Write when a test is failed
			if !test.Passed {
				ts.Failures++
				testCase.Failure = &JUnitFailure{
					Message:  "Failed",
					Type:     "",
					Contents: test.stringify(),
				}
			}

			ts.TestCases = append(ts.TestCases, testCase)
		}

		suites.Suites = append(suites.Suites, ts)
	}

	// to xml
	bytes, err := xml.MarshalIndent(suites, "", "\t")
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
