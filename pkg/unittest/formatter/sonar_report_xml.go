package formatter

import (
	"encoding/xml"
	"io"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
)

// SonarTestExecutions is a collection of Sonar files.
type SonarTestExecutions struct {
	XMLName xml.Name    `xml:"testExecutions"`
	Version int         `xml:"version,attr"`
	Files   []SonarFile `xml:"file"`
}

// SonarFile is a single Sonar test file suite which may contain many
// testCases.
type SonarFile struct {
	XMLName   xml.Name        `xml:"file"`
	Path      string          `xml:"path,attr"`
	TestCases []SonarTestCase `xml:"testCase"`
}

// SonarTestCase is a single test case with its result.
type SonarTestCase struct {
	XMLName  xml.Name      `xml:"testCase"`
	Name     string        `xml:"name,attr"`
	Duration string        `xml:"duration,attr"`
	Error    *SonarError   `xml:"error,omitempty"`
	Skipped  *SonarSkipped `xml:"skipped,omitempty"`
	Failure  *SonarFailure `xml:"failure,omitempty"`
}

// SonarSkipped is set when a test case was skipped.
type SonarSkipped struct {
	XMLName xml.Name `xml:"skipped"`
	Message string   `xml:"message,attr"`
	Reason  string   `xml:",cdata"`
}

// SonarError is set when a test case have an error.
type SonarError struct {
	XMLName    xml.Name `xml:"error"`
	Message    string   `xml:"message,attr"`
	Stacktrace string   `xml:",cdata"`
}

// SonarFailure is set when a test case fails.
type SonarFailure struct {
	XMLName    xml.Name `xml:"failure"`
	Message    string   `xml:"message,attr"`
	Stacktrace string   `xml:",cdata"`
}

type SonarReportXML struct{}

// NewSonarReportXML Constructor
func NewSonarReportXML() Formatter {
	return &SonarReportXML{}
}

// WriteTestOutput writes a Sonar xml representation of the given report
// in the format described at https://docs.sonarqube.org/8.9/analyzing-source-code/generic-test-data/
func (j *SonarReportXML) WriteTestOutput(testSuiteResults []*results.TestSuiteResult, noXMLHeader bool, w io.Writer) error {
	suites := SonarTestExecutions{}
	suites.Version = 1

	// convert TestSuiteResults to Sonar test suites
	for _, testSuiteResult := range testSuiteResults {
		ts := j.createSonarTestSuite(testSuiteResult)

		// individual test cases
		for _, test := range testSuiteResult.TestsResult {
			testCase := j.createSonarTestCase(test)

			if !test.Passed {
				if test.ExecError != nil {
					testCase.Error = j.createSonarError("Error", test.ExecError.Error())
				} else {
					testCase.Failure = j.createSonarFailure("Failed", test.Stringify())
				}
			}

			// skip status currently not supported
			// testCase.Skipped = j.createSonarSkipped()

			ts.TestCases = append(ts.TestCases, testCase)
		}

		suites.Files = append(suites.Files, ts)
	}

	// to xml
	if err := writeContentToFile(noXMLHeader, suites, w); err != nil {
		return err
	}

	return nil
}

func (j *SonarReportXML) createSonarTestSuite(testSuiteResult *results.TestSuiteResult) SonarFile {
	return SonarFile{
		Path:      testSuiteResult.FilePath,
		TestCases: []SonarTestCase{},
	}
}

func (j *SonarReportXML) createSonarTestCase(testJobResult *results.TestJobResult) SonarTestCase {
	return SonarTestCase{
		Name:     testJobResult.DisplayName,
		Duration: formatDurationMilliSeconds(testJobResult.Duration),
		Failure:  nil,
	}
}

func (j *SonarReportXML) createSonarError(message string, stacktrace string) *SonarError {
	return &SonarError{
		Message:    message,
		Stacktrace: stacktrace,
	}
}

func (j *SonarReportXML) createSonarFailure(message string, stacktrace string) *SonarFailure {
	return &SonarFailure{
		Message:    message,
		Stacktrace: stacktrace,
	}
}

/* skip status currently not supported
func (j *SonarReportXML) createSonarSkipped(message string, reason string) *SonarSkipped {
	return &SonarSkipped{
		Message:    message,
		Reason: reason,
	}
}
*/
