package unittest

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"

	"github.com/mitchellh/mapstructure"
)

// Assertion defines target and metrics to validate rendered result
type Assertion struct {
	Template             string
	DocumentSelector     *valueutils.DocumentSelector
	DocumentIndex        int
	Not                  bool
	AssertType           string
	validator            validators.Validatable
	requireRenderSuccess bool
	antonym              bool
	defaultTemplates     []string
}

// Assert validate the rendered manifests with validator
func (a *Assertion) Assert(
	templatesResult map[string][]common.K8sManifest,
	snapshotComparer validators.SnapshotComparer,
	renderSucceed bool,
	renderError error,
	result *results.AssertionResult,
	failfast bool,
	didPostRender bool,
) *results.AssertionResult {
	result.AssertType = a.AssertType
	result.Not = a.Not

	// Ensure assertion is succeeding or failing based on templates to test.
	assertionPassed := false
	failInfo := make([]string, 0)

	var selectedDocsByTemplate map[string][]common.K8sManifest
	var indexError error

	// If we PostRendered, there's no guarantee the post-renderer will preserve our file mapping.  If it doesn't, the
	// parser just puts the whole manifest in one "manifest.yaml" so handle that case:
	if val, ok := templatesResult["manifest.yaml"]; didPostRender && len(templatesResult) == 1 && ok {
		println("Found non-preserving post-render of manifest.yaml only")
		selectedDocsByTemplate["manifest.yaml"] = val
	} else {
		selectedDocsByTemplate, indexError = a.selectDocumentsForAssertion(a.getDocumentsByDefaultTemplates(templatesResult))
	}
	selectedTemplates := a.getKeys(selectedDocsByTemplate)

	// Sort templates to ensure a consistent output
	sort.Strings(selectedTemplates)

	if indexError != nil {
		invalidDocumentIndex := []string{"Error:", indexError.Error()}
		failInfo = append(failInfo, invalidDocumentIndex...)
	} else {
		// Check for failed templates when no documents are found
		if len(selectedTemplates) == 0 {
			var validatePassed bool
			var singleFailInfo []string

			if a.requireRenderSuccess != renderSucceed {
				invalidRender := "Error: rendered manifest is empty"
				failInfo = append(failInfo, invalidRender)
			} else {
				var emptyTemplate []common.K8sManifest
				validatePassed, singleFailInfo = a.validateTemplate(emptyTemplate, emptyTemplate, snapshotComparer, renderError, failfast)
			}

			assertionPassed = validatePassed
			failInfo = append(failInfo, singleFailInfo...)
		}

		for idx, template := range selectedTemplates {
			rendered, ok := templatesResult[template]
			var validatePassed bool
			var singleFailInfo []string

			if !ok && a.requireRenderSuccess {

				noFile := []string{"Error:", a.noFileErrMessage(template)}
				failInfo = append(failInfo, noFile...)
				assertionPassed = false
				break
			}

			if a.requireRenderSuccess != renderSucceed {
				invalidRender := ""
				if len(rendered) > 0 {
					invalidRender = fmt.Sprintf("Error: Invalid rendering: %s", rendered[0][common.RAW])
				} else {
					invalidRender = "Error: rendered manifest is empty"
				}
				failInfo = append(failInfo, invalidRender)
				break
			}

			singleTemplateResult := make(map[string][]common.K8sManifest)
			singleTemplateResult[template] = rendered

			selectedDocs := selectedDocsByTemplate[template]
			validatePassed, singleFailInfo = a.validateTemplate(rendered, selectedDocs, snapshotComparer, renderError, failfast)

			if !validatePassed {
				failInfoTemplate := []string{fmt.Sprintf("Template:\t%s", template)}
				singleFailInfo = append(failInfoTemplate, singleFailInfo...)
			}

			if idx == 0 {
				assertionPassed = validatePassed
			}

			assertionPassed = assertionPassed && validatePassed
			failInfo = append(failInfo, singleFailInfo...)
		}
	}

	result.Passed = assertionPassed
	result.FailInfo = failInfo

	return result
}

func (a *Assertion) validateTemplate(rendered []common.K8sManifest, selectedDocs []common.K8sManifest, snapshotComparer validators.SnapshotComparer, renderError error, failfast bool) (bool, []string) {
	var validatePassed bool
	var singleFailInfo []string

	validatePassed, singleFailInfo = a.validator.Validate(&validators.ValidateContext{
		Docs:             rendered,
		SelectedDocs:     &selectedDocs,
		Negative:         a.Not != a.antonym,
		SnapshotComparer: snapshotComparer,
		RenderError:      renderError,
		FailFast:         failfast,
	})

	return validatePassed, singleFailInfo
}

func (a *Assertion) getDocumentsByDefaultTemplates(templatesResult map[string][]common.K8sManifest) map[string][]common.K8sManifest {
	documentsByDefaultTemplates := map[string][]common.K8sManifest{}

	for _, template := range a.defaultTemplates {
		documentsByDefaultTemplates[template] = templatesResult[template]
	}

	return documentsByDefaultTemplates
}

func (a *Assertion) getKeys(docs map[string][]common.K8sManifest) []string {
	var keys []string

	for key := range docs {
		keys = append(keys, key)
	}

	return keys
}

func (a *Assertion) selectDocumentsForAssertion(docs map[string][]common.K8sManifest) (map[string][]common.K8sManifest, error) {
	if a.DocumentSelector != nil && a.DocumentSelector.Path != "" {
		return a.DocumentSelector.SelectDocuments(docs)
	}

	if a.DocumentIndex != -1 {
		return a.selectDocumentsByIndex(a.DocumentIndex, docs)
	}

	return docs, nil
}

func (a *Assertion) selectDocumentsByIndex(index int, docs map[string][]common.K8sManifest) (map[string][]common.K8sManifest, error) {
	selectedDocs := map[string][]common.K8sManifest{}

	for template, manifests := range docs {
		if index >= len(manifests) {
			return map[string][]common.K8sManifest{}, fmt.Errorf("document index %d is out of rage", a.DocumentIndex)
		}

		selectedDocs[template] = []common.K8sManifest{manifests[index]}
	}

	return selectedDocs, nil
}

func (a *Assertion) noFileErrMessage(template string) string {
	if template != "" {
		return fmt.Sprintf(
			"\ttemplate \"%s\" not exists or not selected in test suite",
			template,
		)
	}

	return "\tassertion.template must be given if testsuite.templates is empty"
}

// UnmarshalYAML implement yaml.Unmalshaler, construct validator according to the assert type
func (a *Assertion) UnmarshalYAML(unmarshal func(interface{}) error) error {
	assertDef := make(map[string]interface{})
	if err := unmarshal(&assertDef); err != nil {
		return err
	}

	if documentIndex, ok := assertDef["documentIndex"].(int); ok {
		a.DocumentIndex = documentIndex
	} else {
		a.DocumentIndex = -1
	}

	if not, ok := assertDef["not"].(bool); ok {
		a.Not = not
	}

	if template, ok := assertDef["template"].(string); ok {
		a.Template = template
	}

	if documentSelector, ok := assertDef["documentSelector"].(map[string]interface{}); ok {
		s, err := valueutils.NewDocumentSelector(documentSelector)
		if err != nil {
			return err
		}
		a.DocumentSelector = s
	}

	if err := a.constructValidator(assertDef); err != nil {
		return err
	}

	if a.validator == nil {
		for key := range assertDef {
			if key != "template" && key != "documentIndex" && key != "not" {
				return fmt.Errorf("Assertion type `%s` is invalid", key)
			}
		}
		return fmt.Errorf("no assertion type defined")
	}

	return nil
}

func (a *Assertion) constructValidator(assertDef map[string]interface{}) error {
	for assertName, correspondDef := range assertTypeMapping {
		if params, ok := assertDef[assertName]; ok {
			if a.validator != nil {
				return fmt.Errorf(
					"assertion type `%s` and `%s` is declared duplicately",
					a.AssertType,
					assertName,
				)
			}

			validator := reflect.New(correspondDef.validatorType).Interface()
			if err := mapstructure.Decode(params, validator); err != nil {
				return err
			}

			a.AssertType = assertName
			a.validator = validator.(validators.Validatable)
			a.requireRenderSuccess = correspondDef.expectRenderSuccess
			a.antonym = correspondDef.antonym
			a.defaultTemplates = []string{a.Template}
		}
	}
	return nil
}

type assertTypeDef struct {
	validatorType       reflect.Type
	antonym             bool
	expectRenderSuccess bool
}

var assertTypeMapping = map[string]assertTypeDef{
	"matchSnapshot":     {reflect.TypeOf(validators.MatchSnapshotValidator{}), false, true},
	"matchSnapshotRaw":  {reflect.TypeOf(validators.MatchSnapshotRawValidator{}), false, true},
	"equal":             {reflect.TypeOf(validators.EqualValidator{}), false, true},
	"notEqual":          {reflect.TypeOf(validators.EqualValidator{}), true, true},
	"greaterOrEqual":    {reflect.TypeOf(validators.EqualOrGreaterValidator{}), false, true},
	"notGreaterOrEqual": {reflect.TypeOf(validators.EqualOrGreaterValidator{}), true, true},
	"lessOrEqual":       {reflect.TypeOf(validators.EqualOrLessValidator{}), false, true},
	"notLessOrEqual":    {reflect.TypeOf(validators.EqualOrLessValidator{}), true, true},
	"equalRaw":          {reflect.TypeOf(validators.EqualRawValidator{}), false, true},
	"notEqualRaw":       {reflect.TypeOf(validators.EqualRawValidator{}), true, true},
	"exists":            {reflect.TypeOf(validators.ExistsValidator{}), false, true},
	"notExists":         {reflect.TypeOf(validators.ExistsValidator{}), true, true},
	"matchRegex":        {reflect.TypeOf(validators.MatchRegexValidator{}), false, true},
	"notMatchRegex":     {reflect.TypeOf(validators.MatchRegexValidator{}), true, true},
	"matchRegexRaw":     {reflect.TypeOf(validators.MatchRegexRawValidator{}), false, true},
	"notMatchRegexRaw":  {reflect.TypeOf(validators.MatchRegexRawValidator{}), true, true},
	"contains":          {reflect.TypeOf(validators.ContainsValidator{}), false, true},
	"notContains":       {reflect.TypeOf(validators.ContainsValidator{}), true, true},
	"isKind":            {reflect.TypeOf(validators.IsKindValidator{}), false, true},
	"isAPIVersion":      {reflect.TypeOf(validators.IsAPIVersionValidator{}), false, true},
	"hasDocuments":      {reflect.TypeOf(validators.HasDocumentsValidator{}), false, true},
	"isSubset":          {reflect.TypeOf(validators.IsSubsetValidator{}), false, true},
	"isNotSubset":       {reflect.TypeOf(validators.IsSubsetValidator{}), true, true},
	"isNullOrEmpty":     {reflect.TypeOf(validators.IsNullOrEmptyValidator{}), false, true},
	"isNotNullOrEmpty":  {reflect.TypeOf(validators.IsNullOrEmptyValidator{}), true, true},
	"failedTemplate":    {reflect.TypeOf(validators.FailedTemplateValidator{}), false, false},
	"notFailedTemplate": {reflect.TypeOf(validators.FailedTemplateValidator{}), true, true},
	"containsDocument":  {reflect.TypeOf(validators.ContainsDocumentValidator{}), false, true},
	"lengthEqual":       {reflect.TypeOf(validators.LengthEqualDocumentsValidator{}), false, true},
	"notLengthEqual":    {reflect.TypeOf(validators.LengthEqualDocumentsValidator{}), true, true},
	"isNull":            {reflect.TypeOf(validators.ExistsValidator{}), true, true},
	"isNotNull":         {reflect.TypeOf(validators.ExistsValidator{}), false, true},
	"isEmpty":           {reflect.TypeOf(validators.IsNullOrEmptyValidator{}), false, true},
	"isNotEmpty":        {reflect.TypeOf(validators.IsNullOrEmptyValidator{}), true, true},
	"isType":            {reflect.TypeOf(validators.IsTypeValidator{}), false, true},
	"isNotType":         {reflect.TypeOf(validators.IsTypeValidator{}), true, true},
}
