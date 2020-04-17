package unittest

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// TestFramework the default name of the test framework.
const TestFramework = "helm-unittest"

const templatePrefix string = "templates"
const subchartPrefix string = "charts"

// getTemplateFileName,
// Validate if prefix templates is not there,
// used for backward compatibility of old unittests.
func getTemplateFileName(fileName string) string {
	if !strings.HasPrefix(fileName, templatePrefix) && !strings.HasPrefix(fileName, subchartPrefix) {
		// Within templates unix separators are always used.
		return filepath.ToSlash(filepath.Join(templatePrefix, fileName))
	}
	return fileName
}

func spliteChartRoutes(routePath string) []string {
	splited := strings.Split(routePath, string(filepath.Separator))
	routes := make([]string, len(splited)/2+1)
	for r := 0; r < len(routes); r++ {
		routes[r] = splited[r*2]
	}
	return routes
}

func scopeValuesWithRoutes(routes []string, values map[interface{}]interface{}) map[interface{}]interface{} {
	if len(routes) > 1 {
		return scopeValuesWithRoutes(
			routes[:len(routes)-1],
			map[interface{}]interface{}{
				routes[len(routes)-1]: values,
			},
		)
	}
	return values
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
