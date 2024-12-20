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

	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
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

func formatDateTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05")
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

func formatDurationMilliSeconds(d time.Duration) string {
	return fmt.Sprintf("%d", d.Milliseconds())
}

func writeContentToFile(noXMLHeader bool, content interface{}, w io.Writer) error {

	// to xml
	bytes, err := xml.MarshalIndent(content, "", "\t")
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(w)

	if !noXMLHeader {
		_, headerErr := writer.WriteString(xml.Header)
		if headerErr != nil {
			return headerErr
		}
	}

	_, writeErr := writer.Write(bytes)
	if writeErr != nil {
		return writeErr
	}
	byteErr := writer.WriteByte('\n')
	if byteErr != nil {
		return byteErr
	}

	return writer.Flush()
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
		fmt.Println("outputDirectory: ", outputDirectory)
		err := os.MkdirAll(outputDirectory, os.ModePerm)
		if err != nil {
			fmt.Println("NewFormatter Error creating output directory: ", err)
			log.Fatal(err)
		}

		switch strings.ToLower(outputType) {
		case "junit":
			return NewJUnitReportXML()
		case "nunit":
			return NewNUnitReportXML()
		case "xunit":
			return NewXUnitReportXML()
		case "sonar":
			return NewSonarReportXML()
		default:
			return nil
		}
	}

	return nil
}
