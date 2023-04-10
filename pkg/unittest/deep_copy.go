package unittest

import (
	"path/filepath"
	"regexp"
	"strings"

	v3chart "helm.sh/helm/v3/pkg/chart"
)

const templatePrefix string = "templates"
const subchartPrefix string = "charts"
const multiWildcard string = "**"
const singleWildcard string = "*"

// getTemplateFileName,
// Validate if prefix templates is not there,
// used for backward compatibility of old unittests.
func getTemplateFileName(fileName string) string {
	if !strings.HasPrefix(fileName, templatePrefix) &&
		!strings.HasPrefix(fileName, subchartPrefix) &&
		!strings.HasPrefix(fileName, multiWildcard) {

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
	copiedChartDependencies := make([]*v3chart.Chart, 0)
	for _, dependency := range targetChart.Dependencies() {
		copiedChartRoute := filepath.Join(chartRoute, subchartPrefix, dependency.Name())
		copiedDependency := CopyV3Chart(copiedChartRoute, templatesToAssert, dependency)
		copiedChartDependencies = append(copiedChartDependencies, copiedDependency)
	}
	copiedChart.SetDependencies(copiedChartDependencies...)

	return copiedChart
}

// filterV3Templates, Filter the V3Templates with only the partials and selected test files.
func filterV3Templates(chartRoute string, templateToAssert []string, targetChart *v3chart.Chart) []*v3chart.File {
	filteredV3Template := make([]*v3chart.File, 0)
	selectedV3TemplateName := chartRoute

	if len(templateToAssert) == 0 {
		// Set all files
		templateToAssert = []string{multiWildcard}
	}

	// check templates in chart
	for _, fileName := range templateToAssert {
		if strings.Contains(fileName, singleWildcard) {
			selectedV3TemplateName = filepath.ToSlash(filepath.Join(targetChart.Root().Name(), getTemplateFileName(fileName)))
		}

		// Set selectedV3TemplateName as regular expression to search
		selectedV3TemplateNamePattern := strings.ReplaceAll(selectedV3TemplateName, multiWildcard, "[a-zA-Z/\\.]+")
		selectedV3TemplateNamePattern = strings.ReplaceAll(selectedV3TemplateNamePattern, singleWildcard, "[a-zA-Z/\\.]*")
		selectedV3TemplateNamePattern = strings.ReplaceAll(selectedV3TemplateNamePattern, ".", "\\.")

		for _, template := range targetChart.Templates {
			foundV3TemplateName := filepath.ToSlash(filepath.Join(chartRoute, template.Name))

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
