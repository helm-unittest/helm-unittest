package formatter

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lrills/helm-unittest/pkg/unittest/results"
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

func writeContentToFile(noXMLHeader bool, content interface{}, w io.Writer) error {

	// to xml
	bytes, err := xml.MarshalIndent(content, "", "\t")
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

// Formatter Interface.
type Formatter interface {
	WriteTestOutput(testSuiteResults []*results.TestSuiteResult, noXMLHeader bool, w io.Writer) error
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
