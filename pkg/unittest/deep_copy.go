package unittest

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mitchellh/copystructure"
	log "github.com/sirupsen/logrus"

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

func copySet(setValues map[string]interface{}) map[string]interface{} {
	copiedSet, err := copystructure.Copy(setValues)
	if err != nil {
		panic(err)
	}

	copiedSetValues := copiedSet.(map[string]interface{})
	// if we have an empty map, make sure it is initialized
	if copiedSetValues == nil {
		copiedSetValues = make(map[string]interface{})
	}

	return copiedSetValues
}

// Copy the V3Chart and its dependencies with partials and optional selected test files.
func CopyV3Chart(chartRoute, currentRoute string, templatesToAssert []string, targetChart *v3chart.Chart) *v3chart.Chart {
	copiedChart := new(v3chart.Chart)
	*copiedChart = *targetChart

	// Clean all parts and rebuild the chart which is needed
	copiedChart.Templates = nil

	// Filter the templates based on the templates to Assert
	// To filter templates ensure only the original chartname is used.
	copiedChart.Templates = filterV3Templates(chartRoute, currentRoute, templatesToAssert, targetChart)

	// Recreate the dependencies
	// Filter trough dependencies.
	copiedChartDependencies := make([]*v3chart.Chart, 0)
	for _, dependency := range targetChart.Dependencies() {
		copiedChartRoute := filepath.Join(currentRoute, subchartPrefix, dependency.Name())
		copiedDependency := CopyV3Chart(chartRoute, copiedChartRoute, templatesToAssert, dependency)
		copiedChartDependencies = append(copiedChartDependencies, copiedDependency)
	}
	copiedChart.SetDependencies(copiedChartDependencies...)

	return copiedChart
}

// filterV3Templates, Filter the V3Templates with only the partials and selected test files.
func filterV3Templates(chartRoute, currentRoute string, templateToAssert []string, targetChart *v3chart.Chart) []*v3chart.File {
	filteredV3Template := make([]*v3chart.File, 0)

	log.WithField("filterV3Templates", "chartRoute").Debugln("expected chartRoute:", chartRoute)
	log.WithField("filterV3Templates", "currentRoute").Debugln("expected currentRoute:", currentRoute)
	log.WithField("filterV3Templates", "templateToAssert").Debugln("expected templateToAssert:", templateToAssert)

	// check templates in chart
	for _, fileName := range templateToAssert {
		selectedV3TemplateName := filepath.ToSlash(filepath.Join(chartRoute, getTemplateFileName(fileName)))

		// Set selectedV3TemplateName as regular expression to search
		selectedV3TemplateNamePattern := strings.ReplaceAll(selectedV3TemplateName, multiWildcard, "[0-9a-zA-Z_\\-/\\.]+")
		selectedV3TemplateNamePattern = strings.ReplaceAll(selectedV3TemplateNamePattern, singleWildcard, "[0-9a-zA-Z_\\-_/\\.]*")
		selectedV3TemplateNamePattern = strings.ReplaceAll(selectedV3TemplateNamePattern, ".", "\\.")

		for _, template := range targetChart.Templates {
			foundV3TemplateName := filepath.ToSlash(filepath.Join(currentRoute, template.Name))

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
