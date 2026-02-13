package unittest

import (
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/mitchellh/copystructure"
	log "github.com/sirupsen/logrus"

	v3chart "helm.sh/helm/v3/pkg/chart"
)

const templatePrefix string = "templates"
const subchartPrefix string = "charts"
const crdsPrefix string = "crds"
const multiWildcard string = "**"
const singleWildcard string = "*"

// getTemplateFileName,
// Validate if prefix templates is not there,
// used for backward compatibility of old unittests.
func getTemplateFileName(fileName string) string {
	if !strings.HasPrefix(fileName, templatePrefix) &&
		!strings.HasPrefix(fileName, subchartPrefix) &&
		!strings.HasPrefix(fileName, crdsPrefix) &&
		!strings.HasPrefix(fileName, multiWildcard) {

		// Within templates unix separators are always used.
		return filepath.ToSlash(filepath.Join(templatePrefix, fileName))
	}
	return fileName
}

// getTemplateFileNamePattern,
// converts a template file name to a regular expression pattern
func getTemplateFileNamePattern(fileName string) string {
	// escape all other regex special characters except for the ones that are already used
	pattern := strings.ReplaceAll(fileName, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "+", "\\+")
	pattern = strings.ReplaceAll(pattern, "[", "\\[")
	pattern = strings.ReplaceAll(pattern, "]", "\\]")
	pattern = strings.ReplaceAll(pattern, multiWildcard, "[0-9a-zA-Z_\\-/\\.]+")
	pattern = strings.ReplaceAll(pattern, singleWildcard, "[0-9a-zA-Z_\\-_/\\.]*")
	pattern = strings.ReplaceAll(pattern, "?", "\\?")
	pattern = strings.ReplaceAll(pattern, "(", "\\(")
	pattern = strings.ReplaceAll(pattern, ")", "\\)")
	pattern = strings.ReplaceAll(pattern, "|", "\\|")
	pattern = strings.ReplaceAll(pattern, "{", "\\{")
	pattern = strings.ReplaceAll(pattern, "}", "\\}")
	pattern = strings.ReplaceAll(pattern, "^", "\\^")
	pattern = strings.ReplaceAll(pattern, "$", "\\$")

	return pattern
}

func CopySet(setValues map[string]any) map[string]any {
	copiedSet, err := copystructure.Copy(setValues)
	if err != nil {
		panic(err)
	}

	copiedSetValues := copiedSet.(map[string]any)
	// if we have an empty map, make sure it is initialized
	if copiedSetValues == nil {
		copiedSetValues = make(map[string]any)
	}

	return copiedSetValues
}

// Copy the V3Chart and its dependencies with partials and optional selected test files.
func FullCopyV3Chart(chartRoute, currentRoute string, targetChart *v3chart.Chart) *v3chart.Chart {
	copiedChart := new(v3chart.Chart)

	// Copy
	for _, rawFile := range targetChart.Raw {
		copiedRawFile := new(v3chart.File)
		copiedRawFile.Name = rawFile.Name
		copiedRawFile.Data = rawFile.Data
		copiedChart.Raw = append(copiedChart.Raw, copiedRawFile)
	}

	copiedChart.Metadata = new(v3chart.Metadata)
	copiedChart.Metadata.Name = targetChart.Metadata.Name
	copiedChart.Metadata.Home = targetChart.Metadata.Home
	copiedChart.Metadata.Sources = targetChart.Metadata.Sources
	copiedChart.Metadata.Version = targetChart.Metadata.Version
	copiedChart.Metadata.Description = targetChart.Metadata.Description
	copiedChart.Metadata.Keywords = targetChart.Metadata.Keywords
	copiedChart.Metadata.Icon = targetChart.Metadata.Icon
	copiedChart.Metadata.APIVersion = targetChart.Metadata.APIVersion
	copiedChart.Metadata.Condition = targetChart.Metadata.Condition
	copiedChart.Metadata.Tags = targetChart.Metadata.Tags
	copiedChart.Metadata.AppVersion = targetChart.Metadata.AppVersion
	copiedChart.Metadata.KubeVersion = targetChart.Metadata.KubeVersion
	copiedChart.Metadata.Type = targetChart.Metadata.Type
	copiedChart.Metadata.Annotations = maps.Clone(targetChart.Metadata.Annotations)

	for _, maintainer := range targetChart.Metadata.Maintainers {
		copiedMaintainer := new(v3chart.Maintainer)
		copiedMaintainer.Name = maintainer.Name
		copiedMaintainer.Email = maintainer.Email
		copiedMaintainer.URL = maintainer.URL
		copiedChart.Metadata.Maintainers = append(copiedChart.Metadata.Maintainers, copiedMaintainer)
	}

	for _, template := range targetChart.Templates {
		copiedTemplate := new(v3chart.File)
		copiedTemplate.Name = template.Name
		copiedTemplate.Data = template.Data
		copiedChart.Templates = append(copiedChart.Templates, copiedTemplate)
	}

	copiedChart.Values = CopySet(targetChart.Values)

	copiedChart.Schema = targetChart.Schema

	for _, file := range targetChart.Files {
		copiedFile := new(v3chart.File)
		copiedFile.Name = file.Name
		copiedFile.Data = file.Data
		copiedChart.Files = append(copiedChart.Files, copiedFile)
	}

	for _, dependency := range targetChart.Metadata.Dependencies {
		copiedDependency := new(v3chart.Dependency)
		copiedDependency.Name = dependency.Name
		copiedDependency.Version = dependency.Version
		copiedDependency.Repository = dependency.Repository
		copiedDependency.Condition = dependency.Condition
		copiedDependency.Tags = dependency.Tags
		copiedDependency.Enabled = dependency.Enabled
		copiedDependency.ImportValues = dependency.ImportValues
		copiedDependency.Alias = dependency.Alias
		copiedChart.Metadata.Dependencies = append(copiedChart.Metadata.Dependencies, copiedDependency)
	}

	// Recreate the dependencies
	// Filter trough dependencies.
	copiedChartDependencies := make([]*v3chart.Chart, 0)
	for _, dependency := range targetChart.Dependencies() {
		copiedChartRoute := filepath.Join(currentRoute, subchartPrefix, dependency.Name())
		copiedDependency := FullCopyV3Chart(chartRoute, copiedChartRoute, dependency)
		copiedChartDependencies = append(copiedChartDependencies, copiedDependency)
	}
	copiedChart.SetDependencies(copiedChartDependencies...)

	return copiedChart
}

// Copy the V3Chart and its dependencies with partials and optional selected test files.
func CopyV3Chart(chartRoute, currentRoute string, templatesToAssert []string, templatesToSkip []string, targetChart *v3chart.Chart) *v3chart.Chart {
	copiedChart := new(v3chart.Chart)
	*copiedChart = *targetChart

	// Clean all parts and rebuild the chart which is needed
	copiedChart.Templates = nil

	// Filter the templates based on the templates to Assert
	// To filter templates ensure only the original chartname is used.
	copiedChart.Templates = filterV3Templates(chartRoute, currentRoute, templatesToAssert, templatesToSkip, targetChart)

	// Recreate the dependencies
	// Filter trough dependencies.
	copiedChartDependencies := make([]*v3chart.Chart, 0)
	for _, dependency := range targetChart.Dependencies() {
		copiedChartRoute := filepath.Join(currentRoute, subchartPrefix, dependency.Name())
		copiedDependency := CopyV3Chart(chartRoute, copiedChartRoute, templatesToAssert, templatesToSkip, dependency)
		copiedChartDependencies = append(copiedChartDependencies, copiedDependency)
	}
	copiedChart.SetDependencies(copiedChartDependencies...)

	return copiedChart
}

// filterV3Templates, Filter the V3Templates with only the partials and selected test files.
func filterV3Templates(chartRoute, currentRoute string, templateToAssert []string, templatesToSkip []string, targetChart *v3chart.Chart) []*v3chart.File {
	filteredV3Template := make([]*v3chart.File, 0)

	log.WithField("filterV3Templates", "chartRoute").Debugln("expected chartRoute:", chartRoute)
	log.WithField("filterV3Templates", "currentRoute").Debugln("expected currentRoute:", currentRoute)
	log.WithField("filterV3Templates", "templateToAssert").Debugln("expected templateToAssert:", templateToAssert)

	// check templates in chart
	for _, fileName := range templateToAssert {
		selectedV3TemplateNamePattern := getTemplateFileNamePattern(filepath.ToSlash(filepath.Join(chartRoute, getTemplateFileName(fileName))))

		for _, template := range targetChart.Templates {
			foundV3TemplateName := filepath.ToSlash(filepath.Join(currentRoute, template.Name))

			if ok, _ := regexp.MatchString(selectedV3TemplateNamePattern, foundV3TemplateName); ok {
				filteredV3Template = append(filteredV3Template, template)
			}
		}
	}

	// remove excluded templates
	filteredV3Template = slices.DeleteFunc(filteredV3Template, func(template *v3chart.File) bool {
		foundV3TemplateName := filepath.ToSlash(filepath.Join(currentRoute, template.Name))

		return slices.ContainsFunc(templatesToSkip, func(fileName string) bool {
			selectedV3TemplateNamePattern := getTemplateFileNamePattern(filepath.ToSlash(filepath.Join(chartRoute, getTemplateFileName(fileName))))

			ok, _ := regexp.MatchString(selectedV3TemplateNamePattern, foundV3TemplateName)
			return ok
		})
	})

	// add partial templates
	for _, template := range targetChart.Templates {
		if strings.HasPrefix(filepath.Base(template.Name), "_") {
			filteredV3Template = append(filteredV3Template, template)
		}
	}

	return filteredV3Template
}
