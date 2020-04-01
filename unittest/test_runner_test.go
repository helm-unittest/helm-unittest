package unittest_test

import (
	"bytes"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	. "github.com/lrills/helm-unittest/unittest"
	"github.com/stretchr/testify/assert"
)

var sectionBeginPattern = regexp.MustCompile("( PASS | FAIL |\n*###|\n*Charts:|\n*Snapshot Summary:)")
var timePattern = regexp.MustCompile("Time:\\s+([\\d\\.]+)ms")

func makeOutputSnapshotable(originalOutput string) []interface{} {
	output := strings.ReplaceAll(originalOutput, "\\", "/")
	timeLoc := timePattern.FindStringSubmatchIndex(output)[2:4]
	timeAgnosticOutput := output[:timeLoc[0]] + "XX.XXX" + output[timeLoc[1]:]

	sectionBeggingLocs := sectionBeginPattern.FindAllStringIndex(timeAgnosticOutput, -1)
	sections := make([]string, len(sectionBeggingLocs))

	suiteBeginIdx := -1
	for sectionIdx := 0; sectionIdx < len(sections); sectionIdx++ {
		start := sectionBeggingLocs[sectionIdx][0]
		var end int
		if sectionIdx >= len(sections)-1 {
			end = len(timeAgnosticOutput)
		} else {
			end = sectionBeggingLocs[sectionIdx+1][0]
		}

		sectionContent := timeAgnosticOutput[start:end]
		sectionBegin := sectionContent[:6]
		if sectionBegin == " PASS " || sectionBegin == " FAIL " {
			sections[sectionIdx] = strings.TrimRight(sectionContent, "\n")
			if suiteBeginIdx == -1 {
				suiteBeginIdx = sectionIdx
			}
		} else {
			sections[sectionIdx] = sectionContent
			if suiteBeginIdx != -1 {
				sort.Strings(sections[suiteBeginIdx:sectionIdx])
				suiteBeginIdx = -1
			}
		}
	}

	sectionsToRetrun := make([]interface{}, len(sections))
	for idx, section := range sections {
		sectionsToRetrun[idx] = section
	}
	return sectionsToRetrun
}

func TestV2RunnerOkWithPassedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: NewPrinter(buffer, nil),
		Config: TestConfig{
			TestFiles: []string{"tests/*_test.yaml"},
		},
	}
	passed := runner.RunV2([]string{"../__fixtures__/v2/basic"})
	assert.True(t, passed)
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV2RunnerOkWithFailedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: NewPrinter(buffer, nil),
		Config: TestConfig{
			TestFiles: []string{"tests_failed/*_test.yaml"},
		},
	}
	passed := runner.RunV2([]string{"../__fixtures__/v2/basic"})
	assert.False(t, passed)
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV2RunnerWithTestsInSubchart(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: NewPrinter(buffer, nil),
		Config: TestConfig{
			WithSubChart: true,
			TestFiles:    []string{"tests/*_test.yaml"},
		},
	}
	passed := runner.RunV2([]string{"../__fixtures__/v2/with-subchart"})
	assert.True(t, passed)
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV2RunnerWithTestsInSubchartButFlagFalse(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: NewPrinter(buffer, nil),
		Config: TestConfig{
			WithSubChart: false,
			TestFiles:    []string{"tests/*_test.yaml"},
		},
	}
	passed := runner.RunV2([]string{"../__fixtures__/v2/with-subchart"})
	assert.True(t, passed)
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithPassedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: NewPrinter(buffer, nil),
		Config: TestConfig{
			TestFiles: []string{"tests/*_test.yaml"},
		},
	}
	passed := runner.RunV3([]string{"../__fixtures__/v3/basic"})
	assert.True(t, passed)
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithFailedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: NewPrinter(buffer, nil),
		Config: TestConfig{
			TestFiles: []string{"tests_failed/*_test.yaml"},
		},
	}
	passed := runner.RunV3([]string{"../__fixtures__/v3/basic"})
	assert.False(t, passed)
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerWithTestsInSubchart(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: NewPrinter(buffer, nil),
		Config: TestConfig{
			WithSubChart: true,
			TestFiles:    []string{"tests/*_test.yaml"},
		},
	}
	passed := runner.RunV3([]string{"../__fixtures__/v3/with-subchart"})
	assert.True(t, passed)
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerWithTestsInSubchartButFlagFalse(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: NewPrinter(buffer, nil),
		Config: TestConfig{
			WithSubChart: false,
			TestFiles:    []string{"tests/*_test.yaml"},
		},
	}
	passed := runner.RunV3([]string{"../__fixtures__/v3/with-subchart"})
	assert.True(t, passed)
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}
