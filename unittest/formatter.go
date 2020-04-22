package unittest

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// testFramework the default name of the test framework.
const testFramework = "helm-unittest"

func determineClassnameFromDisplayName(displayName string) string {
	classname := displayName
	if idx := strings.LastIndex(classname, "/"); idx > -1 && idx < len(displayName) {
		classname = displayName[idx+1:]
	}
	return classname
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func formatTime(t time.Time) string {
	return t.Format("15:04:05")
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.3f", d.Seconds())
}

// Formatter Interface.
type Formatter interface {
	WriteTestOutput(testSuiteResults []*TestSuiteResult, noXMLHeader bool, w io.Writer) error
}

// NewFormatter create a new Formatter.
func NewFormatter(outputFile, outputType string) Formatter {
	if outputFile != "" {
		// Ensure the directory of the outputFile is created
		outputDirectory := filepath.Dir(outputFile)
		err := os.MkdirAll(outputDirectory, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		switch strings.ToLower(outputType) {
		case "junit":
			return NewJUnitReportXML()
		case "nunit":
			return NewNUnitReportXML()
		case "xunit":
			return NewXUnitReportXML()
		default:
			return nil
		}
	}

	return nil
}
