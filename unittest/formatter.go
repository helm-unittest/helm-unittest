package unittest

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

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
