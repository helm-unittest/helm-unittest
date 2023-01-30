package formatter

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/jibber_jabber"
	"github.com/lrills/helm-unittest/pkg/unittest/results"
)

// OSVersion the default value of the OS.
const OSVersion = "unknown"

// CLRVersion the default value of the CLR.
const CLRVersion = "unknown"

// NUnitVersion the version of NUnit
const NUnitVersion = "2.5.8.0"

// TestFixture the default fixture of a test.
const TestFixture = "TestFixture"

// NUnitTestResults is a collection of NUnit test suites.
type NUnitTestResults struct {
	XMLName      xml.Name         `xml:"test-results"`
	Environment  NUnitEnvironment `xml:"environment"`
	CultureInfo  NUnitCultureInfo `xml:"culture-info"`
	TestSuite    []NUnitTestSuite `xml:"test-suite"`
	Name         string           `xml:"name,attr"`
	Total        int              `xml:"total,attr"`
	Errors       int              `xml:"errors,attr"`
	Failures     int              `xml:"failures,attr"`
	Inconclusive int              `xml:"inconclusive,attr"`
	NotRun       int              `xml:"not-run,attr"`
	Ignored      int              `xml:"ignored,attr"`
	Skipped      int              `xml:"skipped,attr"`
	Invalid      int              `xml:"invalid,attr"`
	Date         string           `xml:"date,attr"`
	Time         string           `xml:"time,attr"`
}

// NUnitEnvironment is the environment settings.
type NUnitEnvironment struct {
	XMLName      xml.Name `xml:"environment"`
	NUnitVersion string   `xml:"nunit-version,attr"`
	CLRVersion   string   `xml:"clr-version,attr"`
	OSVersion    string   `xml:"os-version,attr"`
	Platform     string   `xml:"platform,attr"`
	Cwd          string   `xml:"cwd,attr"`
	MachineName  string   `xml:"machine-name,attr"`
	User         string   `xml:"user,attr"`
	UserDomain   string   `xml:"user-domain,attr"`
}

// NUnitCultureInfo is the environment settings.
type NUnitCultureInfo struct {
	XMLName          xml.Name `xml:"culture-info"`
	CurrentCulture   string   `xml:"current-culture,attr"`
	CurrentUICulture string   `xml:"current-uiculture,attr"`
}

// NUnitTestSuite is a single NUnit test suite which may contain many
// testcases.
type NUnitTestSuite struct {
	XMLName     xml.Name         `xml:"test-suite"`
	Failure     *NUnitFailure    `xml:"failure,omitempty"`
	Reason      *NUnitReason     `xml:"reason,omitempty"`
	TestSuites  []NUnitTestSuite `xml:"results>test-suite,omitempty"`
	TestCases   []NUnitTestCase  `xml:"results>test-case,omitempty"`
	Type        string           `xml:"type,attr"`
	Name        string           `xml:"name,attr"`
	Description string           `xml:"description,attr"`
	Success     string           `xml:"success,attr"`
	Time        string           `xml:"time,attr"`
	Executed    string           `xml:"executed,attr"`
	Asserts     string           `xml:"asserts,attr"`
	Result      string           `xml:"result,attr"`
}

// NUnitTestCase is a single test case with its result.
type NUnitTestCase struct {
	XMLName     xml.Name      `xml:"test-case"`
	Failure     *NUnitFailure `xml:"failure,omitempty"`
	Reason      *NUnitReason  `xml:"reason,omitempty"`
	Name        string        `xml:"name,attr"`
	Description string        `xml:"description,attr"`
	Success     string        `xml:"success,attr"`
	Time        string        `xml:"time,attr"`
	Executed    string        `xml:"executed,attr"`
	Asserts     string        `xml:"asserts,attr"`
	Result      string        `xml:"result,attr"`
}

// NUnitCategory is a testsuitecategory
type NUnitCategory struct {
	Name string `xml:"name,attr"`
}

// NUnitProperty represents a key/value pair used to define properties.
type NUnitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// NUnitFailure contains data related to a failed test.
type NUnitFailure struct {
	Message    string `xml:"message"`
	StackTrace string `xml:"stack-trace"`
}

// NUnitReason contains data related to a failed test.
type NUnitReason struct {
	Message string `xml:"message"`
}

type nUnitReportXML struct{}

// NewNUnitReportXML Constructor
func NewNUnitReportXML() Formatter {
	return &nUnitReportXML{}
}

// NUnitReportXML writes a NUnit xml representation of the given report to w
// in the format described at https://github.com/nunit/docs/wiki/XML-Formats
func (n *nUnitReportXML) WriteTestOutput(testSuiteResults []*results.TestSuiteResult, noXMLHeader bool, w io.Writer) error {
	currentTime := time.Now()
	totalTests := 0
	totalErrors := 0
	totalFailures := 0
	totalSuccess := true
	testSuites := []NUnitTestSuite{}

	// convert TestSuiteResults to NUnit test suites
	for _, testSuiteResult := range testSuiteResults {
		totalSuccess = totalSuccess && testSuiteResult.Passed

		ts := n.createNUnitTestSuite(testSuiteResult)

		// In case the testsuite failed with an error
		// direct append to the list and iterate to the next suite.
		if testSuiteResult.ExecError != nil {
			totalTests++
			totalErrors++
			ts.Failure = n.createNUnitFailure("Error", testSuiteResult.ExecError.Error())
			testSuites = append(testSuites, ts)
			continue
		}
		// individual test cases
		for _, test := range testSuiteResult.TestsResult {
			totalTests++
			testCase := n.createNUnitTestCase(determineClassnameFromDisplayName(testSuiteResult.DisplayName), test)

			// Write when a test is failed
			if !test.Passed {
				// Update total counts
				if test.ExecError != nil {
					totalErrors++
				} else {
					totalFailures++
				}

				testCase.Failure = n.createNUnitFailure("Failed", test.Stringify())
			}

			ts.TestCases = append(ts.TestCases, testCase)
		}

		testSuites = append(testSuites, ts)
	}

	nunitResult := n.createNUnitTestResults(currentTime, totalTests, totalErrors, totalFailures, totalSuccess, testSuites)

	// to xml
	if err := writeContentToFile(noXMLHeader, nunitResult, w); err != nil {
		return err
	}

	return nil
}

func (n *nUnitReportXML) formatUserAndDomain() (domainName, userName string) {
	userSettings, _ := user.Current()
	userDomainName := strings.Split(userSettings.Username, "\\")

	if len(userDomainName) == 2 {
		domainName = userDomainName[0]
		userName = userDomainName[1]
	} else {
		userName = userDomainName[0]
	}

	return domainName, userName
}

func (n *nUnitReportXML) formatResult(b bool) string {
	if !b {
		return "Failed"
	}
	return "Success"
}

func (n *nUnitReportXML) createNUnitTestResults(
	currentTime time.Time,
	totalTests, totalErrors, totalFailures int,
	totalSuccess bool,
	testSuites []NUnitTestSuite) NUnitTestResults {
	domainName, userName := n.formatUserAndDomain()
	cwd, _ := os.Getwd()
	hostName, _ := os.Hostname()
	currentCulture, _ := jibber_jabber.DetectLanguage()
	currentUICulture, _ := jibber_jabber.DetectIETF()

	return NUnitTestResults{
		Environment: NUnitEnvironment{
			NUnitVersion: NUnitVersion,
			CLRVersion:   CLRVersion,
			OSVersion:    OSVersion,
			Platform:     fmt.Sprintf("%s.%s-%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
			Cwd:          cwd,
			MachineName:  hostName,
			User:         userName,
			UserDomain:   domainName,
		},
		CultureInfo: NUnitCultureInfo{
			CurrentCulture:   currentCulture,
			CurrentUICulture: currentUICulture,
		},
		TestSuite:    testSuites,
		Name:         testFramework,
		Total:        totalTests,
		Errors:       totalErrors,
		Failures:     totalFailures,
		NotRun:       0,
		Inconclusive: 0,
		Ignored:      0,
		Skipped:      0,
		Invalid:      0,
		Date:         formatDate(currentTime),
		Time:         formatTime(currentTime),
	}
}

func (n *nUnitReportXML) createNUnitTestSuite(testSuiteResult *results.TestSuiteResult) NUnitTestSuite {
	return NUnitTestSuite{
		Type:        TestFixture,
		Name:        testSuiteResult.DisplayName,
		Description: testSuiteResult.FilePath,
		Success:     strconv.FormatBool(testSuiteResult.Passed),
		Time:        formatDuration(testSuiteResult.CalculateTestSuiteDuration()),
		Executed:    strconv.FormatBool(testSuiteResult.ExecError == nil),
		Result:      n.formatResult(testSuiteResult.Passed),
	}
}

func (n *nUnitReportXML) createNUnitTestCase(className string, testJobResult *results.TestJobResult) NUnitTestCase {
	return NUnitTestCase{
		Failure:     nil,
		Name:        testJobResult.DisplayName,
		Description: fmt.Sprintf("%s.%s", className, testJobResult.DisplayName),
		Success:     strconv.FormatBool(testJobResult.Passed),
		Time:        formatDuration(testJobResult.Duration),
		Executed:    strconv.FormatBool(testJobResult.ExecError == nil),
		Asserts:     "0",
		Result:      n.formatResult(testJobResult.Passed),
	}
}

func (n *nUnitReportXML) createNUnitFailure(errorMessage, stackTrace string) *NUnitFailure {
	return &NUnitFailure{
		Message:    errorMessage,
		StackTrace: stackTrace,
	}
}
