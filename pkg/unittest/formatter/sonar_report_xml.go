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
}

// SonarError is set when a test case have an error.
type SonarError struct {
	XMLName xml.Name `xml:"error"`
}

// SonarFailure is set when a test case fails.
type SonarFailure struct {
	XMLName xml.Name `xml:"failure"`
}

type SonarReportXML struct{}

// NewSonarReportXML Constructor
func NewSonarReportXML() Formatter {
	return &SonarReportXML{}
}

// WriteTestOutput writes a Sonar xml representation of the given report
// in the format described at https://docs.Sonar.org/latest/analyzing-source-code/test-coverage/generic-test-data/#generic-test-execution
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
					testCase.Error = j.createSonarError()
				} else {
					testCase.Failure = j.createSonarFailure()
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

func (j *SonarReportXML) createSonarError() *SonarError {
	return &SonarError{}
}

func (j *SonarReportXML) createSonarFailure() *SonarFailure {
	return &SonarFailure{}
}

/* skip status currently not supported
func (j *SonarReportXML) createSonarSkipped() *SonarSkipped {
	return &SonarSkipped{}
}
*/
