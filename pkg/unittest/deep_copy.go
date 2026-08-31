package unittest

import (
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/mitchellh/copystructure"
	log "github.com/sirupsen/logrus"

	chartcommon "helm.sh/helm/v4/pkg/chart/common"
	v2chart "helm.sh/helm/v4/pkg/chart/v2"
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

// Copy the V2Chart and its dependencies with partials and optional selected test files.
func FullCopyV2Chart(chartRoute, currentRoute string, targetChart *v2chart.Chart) *v2chart.Chart {
	copiedChart := new(v2chart.Chart)

	// Copy
	for _, rawFile := range targetChart.Raw {
		copiedRawFile := new(chartcommon.File)
		copiedRawFile.Name = rawFile.Name
		copiedRawFile.ModTime = rawFile.ModTime
		copiedRawFile.Data = rawFile.Data
		copiedChart.Raw = append(copiedChart.Raw, copiedRawFile)
	}

	copiedChart.Metadata = new(v2chart.Metadata)
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
		copiedMaintainer := new(v2chart.Maintainer)
		copiedMaintainer.Name = maintainer.Name
		copiedMaintainer.Email = maintainer.Email
		copiedMaintainer.URL = maintainer.URL
		copiedChart.Metadata.Maintainers = append(copiedChart.Metadata.Maintainers, copiedMaintainer)
	}

	for _, template := range targetChart.Templates {
		copiedTemplate := new(chartcommon.File)
		copiedTemplate.Name = template.Name
		copiedTemplate.ModTime = template.ModTime
		copiedTemplate.Data = template.Data
		copiedChart.Templates = append(copiedChart.Templates, copiedTemplate)
	}

	copiedChart.Values = CopySet(targetChart.Values)

	copiedChart.Schema = targetChart.Schema

	for _, file := range targetChart.Files {
		copiedFile := new(chartcommon.File)
		copiedFile.Name = file.Name
		copiedFile.ModTime = file.ModTime
		copiedFile.Data = file.Data
		copiedChart.Files = append(copiedChart.Files, copiedFile)
	}

	for _, dependency := range targetChart.Metadata.Dependencies {
		copiedDependency := new(v2chart.Dependency)
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
	copiedChartDependencies := make([]*v2chart.Chart, 0)
	for _, dependency := range targetChart.Dependencies() {
		copiedChartRoute := filepath.Join(currentRoute, subchartPrefix, dependency.Name())
		copiedDependency := FullCopyV2Chart(chartRoute, copiedChartRoute, dependency)
		copiedChartDependencies = append(copiedChartDependencies, copiedDependency)
	}
	copiedChart.SetDependencies(copiedChartDependencies...)

	return copiedChart
}

// Copy the V4Chart and its dependencies with partials and optional selected test files.
func CopyV2Chart(chartRoute, currentRoute string, templatesToAssert []string, templatesToSkip []string, targetChart *v2chart.Chart) *v2chart.Chart {
	copiedChart := new(v2chart.Chart)
	*copiedChart = *targetChart

	// Clean all parts and rebuild the chart which is needed
	copiedChart.Templates = nil

	// Filter the templates based on the templates to Assert
	// To filter templates ensure only the original chartname is used.
	copiedChart.Templates = filterV2Templates(chartRoute, currentRoute, templatesToAssert, templatesToSkip, targetChart)

	// Recreate the dependencies
	// Filter trough dependencies.
	copiedChartDependencies := make([]*v2chart.Chart, 0)
	for _, dependency := range targetChart.Dependencies() {
		copiedChartRoute := filepath.Join(currentRoute, subchartPrefix, dependency.Name())
		copiedDependency := CopyV2Chart(chartRoute, copiedChartRoute, templatesToAssert, templatesToSkip, dependency)
		copiedChartDependencies = append(copiedChartDependencies, copiedDependency)
	}
	copiedChart.SetDependencies(copiedChartDependencies...)

	return copiedChart
}

// filterV2Templates, Filter the V2Templates with only the partials and selected test files.
func filterV2Templates(chartRoute, currentRoute string, templateToAssert []string, templatesToSkip []string, targetChart *v2chart.Chart) []*chartcommon.File {
	filteredV2Template := make([]*chartcommon.File, 0)

	log.WithField("filterV2Templates", "chartRoute").Debugln("expected chartRoute:", chartRoute)
	log.WithField("filterV2Templates", "currentRoute").Debugln("expected currentRoute:", currentRoute)
	log.WithField("filterV2Templates", "templateToAssert").Debugln("expected templateToAssert:", templateToAssert)

	// check templates in chart
	for _, fileName := range templateToAssert {
		selectedV2TemplateNamePattern := getTemplateFileNamePattern(filepath.ToSlash(filepath.Join(chartRoute, getTemplateFileName(fileName))))

		for _, template := range targetChart.Templates {
			foundV2TemplateName := filepath.ToSlash(filepath.Join(currentRoute, template.Name))

			if ok, _ := regexp.MatchString(selectedV2TemplateNamePattern, foundV2TemplateName); ok {
				filteredV2Template = append(filteredV2Template, template)
			}
		}
	}

	// remove excluded templates
	filteredV2Template = slices.DeleteFunc(filteredV2Template, func(template *chartcommon.File) bool {
		foundV2TemplateName := filepath.ToSlash(filepath.Join(currentRoute, template.Name))

		return slices.ContainsFunc(templatesToSkip, func(fileName string) bool {
			selectedV2TemplateNamePattern := getTemplateFileNamePattern(filepath.ToSlash(filepath.Join(chartRoute, getTemplateFileName(fileName))))

			ok, _ := regexp.MatchString(selectedV2TemplateNamePattern, foundV2TemplateName)
			return ok
		})
	})

	// add partial templates
	for _, template := range targetChart.Templates {
		if strings.HasPrefix(filepath.Base(template.Name), "_") {
			filteredV2Template = append(filteredV2Template, template)
		}
	}

	return filteredV2Template
}
