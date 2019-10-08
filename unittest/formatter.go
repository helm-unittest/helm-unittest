package unittest

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// Formatter Interface.
type Formatter interface {
	WriteTestOutput(testSuiteResults []*TestSuiteResult, noXMLHeader bool, w io.Writer) error
}

// NewFormatter create a new Formatter.
func NewFormatter(outputFile string) Formatter {
	if outputFile != "" {
		// Ensure the directory of the outputFile is created
		outputDirectory := filepath.Dir(outputFile)
		err := os.MkdirAll(outputDirectory, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		return NewJUnitReportXML()
	}

	return nil
}
