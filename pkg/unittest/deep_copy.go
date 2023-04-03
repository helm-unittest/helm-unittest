package unittest

import (
	"path/filepath"
	"regexp"
	"strings"

	v3chart "helm.sh/helm/v3/pkg/chart"
)

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

// Copy the V3Chart and its dependencies with partials and optional selected test files.
func CopyV3Chart(chartRoute string, templatesToAssert []string, targetChart *v3chart.Chart) *v3chart.Chart {
	copiedChart := new(v3chart.Chart)
	*copiedChart = *targetChart

	// Clean all parts and rebuild the chart which is needed
	copiedChart.Files = nil
	copiedChart.Raw = nil
	copiedChart.Templates = nil

	// Filter the templates based on the templates to Assert
	copiedChart.Templates = filterV3Templates(chartRoute, templatesToAssert, targetChart)

	// Recreate the dependencies
	// Filter trough dependencies.
	for _, dependency := range targetChart.Dependencies() {
		copiedChartRoute := filepath.Join(chartRoute, subchartPrefix, dependency.Name())
		copiedDependency := CopyV3Chart(copiedChartRoute, templatesToAssert, dependency)
		copiedChart.AddDependency(copiedDependency)
	}

	return copiedChart
}

// filterV3Templates, Filter the V3Templates with only the partials and selected test files.
func filterV3Templates(chartRoute string, templateToAssert []string, targetChart *v3chart.Chart) []*v3chart.File {
	filteredV3Template := make([]*v3chart.File, 0)

	if len(templateToAssert) == 0 {
		// Set all files
		templateToAssert = []string{"**"}
	}

	// check templates in chart
	for _, fileName := range templateToAssert {
		for _, template := range targetChart.Templates {
			selectedV3TemplateName := filepath.ToSlash(filepath.Join(chartRoute, getTemplateFileName(fileName)))
			foundV3TemplateName := filepath.ToSlash(filepath.Join(chartRoute, template.Name))

			// Set selectedV3TemplateName as regular expression to search
			selectedV3TemplateNamePattern := strings.ReplaceAll(selectedV3TemplateName, "**", "[a-zA-Z/\\.]*")
			selectedV3TemplateNamePattern = strings.ReplaceAll(selectedV3TemplateNamePattern, "*", "[a-zA-Z]*")
			selectedV3TemplateNamePattern = strings.ReplaceAll(selectedV3TemplateNamePattern, ".", "\\.")

			if ok, _ := regexp.MatchString(selectedV3TemplateNamePattern, foundV3TemplateName); ok {
				filteredV3Template = append(filteredV3Template, template)
			}
		}
	}

	// add partial templates
	for _, template := range targetChart.Templates {
		if strings.HasPrefix(filepath.Base(template.Name), "_") {
			filteredV3Template = append(filteredV3Template, template)
		}
	}

	return filteredV3Template
}
