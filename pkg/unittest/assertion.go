package unittest

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"

	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
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
	config               AssertionConfig
}

func (a *Assertion) WithConfig(config AssertionConfig) {
	a.config = config
}

func (a *Assertion) configOrDefault() AssertionConfig {
	if a.config.templatesResult == nil {
		a.config.templatesResult = make(map[string][]common.K8sManifest)
	}
	return a.config
}

// Assert validate the rendered manifests with validator
func (a *Assertion) Assert(
	result *results.AssertionResult,
) *results.AssertionResult {
	result.AssertType = a.AssertType
	result.Not = a.Not

	var templates = a.computeTemplatesWithPostRender()

	// TODO: This could be optimised and computed once for the test suite
	selectedDocsByTemplate, indexError := a.selectDocumentsForAssertion(templates)
	selectedTemplates := a.getKeys(selectedDocsByTemplate)

	// Sort templates to ensure a consistent output
	sort.Strings(selectedTemplates)

	if indexError != nil {
		return a.handleIndexError(result, indexError)
	}

	if a.shouldSkipAssertion(selectedTemplates) {
		return a.skipAssertion(result)
	}

	if len(selectedTemplates) == 0 {
		return a.evaluateEmptyTemplates(result)
	}

	return a.evaluateTemplates(result, selectedTemplates, selectedDocsByTemplate)
}

// handleIndexError handles the error when the document index is out of range
// It sets the assertion result to failed and includes the error message
func (a *Assertion) handleIndexError(result *results.AssertionResult, indexError error) *results.AssertionResult {
	result.Passed = false
	result.FailInfo = []string{"Error:", indexError.Error()}
	return result
}

// shouldSkipAssertion checks if the assertion should be skipped based on the configuration
func (a *Assertion) shouldSkipAssertion(selectedTemplates []string) bool {
	return a.configOrDefault().isSkipEmptyTemplate && len(selectedTemplates) == 0
}

// skipAssertion marks the assertion as skipped and sets the reason
// It returns the assertion result with the skip status and reason
func (a *Assertion) skipAssertion(result *results.AssertionResult) *results.AssertionResult {
	result.Skipped = true
	result.SkipReason = "skipped as 'documentSelector.skipEmptyTemplates: true' and 'selectedTemplates: empty'"
	result.Passed = true
	log.WithField(common.LOG_TEST_ASSERTION, "assert").Debugln("skip assertion", result.SkipReason)
	return result
}

// evaluateEmptyTemplates evaluates the assertion when there are no selected templates
// It returns the assertion result with the validation status and failure information
func (a *Assertion) evaluateEmptyTemplates(
	result *results.AssertionResult,
) *results.AssertionResult {
	validatePassed := false
	failInfo := make([]string, 0)

	if a.requireRenderSuccess != a.configOrDefault().renderSucceed {
		invalidRender := "Error: rendered manifest is empty"
		failInfo = append(failInfo, invalidRender)
	} else {
		var emptyTemplate []common.K8sManifest
		_, validatePassed, failInfo = a.validateTemplate(emptyTemplate, emptyTemplate)
	}

	result.Passed = validatePassed
	result.FailInfo = failInfo
	return result
}

// evaluateTemplates evaluates the assertion for each selected template
// It processes the templates and validates them using the configured validator
// It returns the assertion result with the validation status and failure information
func (a *Assertion) evaluateTemplates(
	result *results.AssertionResult,
	selectedTemplates []string,
	selectedDocsByTemplate map[string][]common.K8sManifest,
) *results.AssertionResult {
	assertionPassed := false
	failInfo := make([]string, 0)

	for idx, template := range selectedTemplates {
		addTemplate, validatePassed, singleFailInfo := a.processTemplate(template, selectedDocsByTemplate[template])
		if !validatePassed {
			if addTemplate {
				failInfo = append(failInfo, fmt.Sprintf("Template:\t%s", template))
			}
			failInfo = append(failInfo, singleFailInfo...)
		}

		if idx == 0 {
			assertionPassed = validatePassed
		}
		assertionPassed = assertionPassed && validatePassed
	}

	result.Passed = assertionPassed
	result.FailInfo = failInfo
	return result
}

// processTemplate processes the template and validates it using the configured validator
// It returns a boolean indicating if the template needs to be added in the failure information,
// a boolean indicating if the validation passed, and a slice of failure information
func (a *Assertion) processTemplate(template string, selectedDocs []common.K8sManifest) (bool, bool, []string) {
	rendered, ok := a.configOrDefault().templatesResult[template]
	if !ok && a.requireRenderSuccess {
		return false, false, []string{"Error:", a.noFileErrMessage(template)}
	}

	if a.requireRenderSuccess != a.configOrDefault().renderSucceed {
		return true, false, a.handleRenderError(rendered)
	}

	return a.validateTemplate(rendered, selectedDocs)
}

// handleRenderError handles the error when the rendered manifest is empty
// It returns a slice of failure information
func (a *Assertion) handleRenderError(rendered []common.K8sManifest) []string {
	if len(rendered) > 0 {
		return []string{fmt.Sprintf("Error: Invalid rendering: %s", rendered[0][common.RAW])}
	}
	return []string{"Error: rendered manifest is empty"}
}

// validateTemplate validates the rendered template using the configured validator
// It returns a boolean indicating if the template needs to be added in the failure information,
// a boolean indicating if the template is valid and a slice of failure information
func (a *Assertion) validateTemplate(rendered []common.K8sManifest, selectedDocs []common.K8sManifest) (bool, bool, []string) {
	var validatePassed bool
	var singleFailInfo []string

	validatePassed, singleFailInfo = a.validator.Validate(&validators.ValidateContext{
		Docs:             rendered,
		SelectedDocs:     &selectedDocs,
		Negative:         a.Not != a.antonym,
		SnapshotComparer: a.configOrDefault().snapshotComparer,
		RenderError:      a.configOrDefault().renderError,
		FailFast:         a.configOrDefault().failFast,
	})

	return true, validatePassed, singleFailInfo
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
			return map[string][]common.K8sManifest{}, fmt.Errorf("document index %d is out of range", a.DocumentIndex)
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

// UnmarshalYAML implements yaml.Unmarshaler, constructing the validator according to the assert type.
func (a *Assertion) UnmarshalYAML(unmarshal func(interface{}) error) error {
	assertDef := make(map[string]interface{})
	if err := unmarshal(&assertDef); err != nil {
		return err
	}

	a.parseBasicFields(assertDef)
	if err := a.parseDocumentSelector(assertDef); err != nil {
		return err
	}

	if err := a.constructValidator(assertDef); err != nil {
		return err
	}

	if a.validator == nil {
		return a.validateAssertionType(assertDef)
	}

	return nil
}

// parseBasicFields parses basic fields like documentIndex, not, and template.
func (a *Assertion) parseBasicFields(assertDef map[string]interface{}) {
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
}

// parseDocumentSelector parses the documentSelector field if present.
func (a *Assertion) parseDocumentSelector(assertDef map[string]interface{}) error {
	if documentSelector, ok := assertDef["documentSelector"].(map[string]interface{}); ok {
		s, err := valueutils.NewDocumentSelector(documentSelector)
		if err != nil {
			return err
		}
		a.DocumentSelector = s
	}
	return nil
}

// validateAssertionType validates the assertion type and ensures at least one is defined.
func (a *Assertion) validateAssertionType(assertDef map[string]interface{}) error {
	for key := range assertDef {
		if key != "template" && key != "documentIndex" && key != "not" {
			return fmt.Errorf("Assertion type `%s` is invalid", key)
		}
	}
	return fmt.Errorf("no assertion type defined")
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

func (a *Assertion) computeTemplatesWithPostRender() map[string][]common.K8sManifest {
	// If we PostRendered, there's no guarantee the post-renderer will preserve our file mapping.  If it doesn't, the
	// parser just puts the whole manifest in one "manifest.yaml" so handle that case:
	var templates map[string][]common.K8sManifest
	templatesResult := a.configOrDefault().templatesResult
	if a.configOrDefault().didPostRender && len(templatesResult) == 1 {
		for key := range templatesResult {
			if strings.HasSuffix(key, "manifest.yaml") {
				a.Template = key
				templates = map[string][]common.K8sManifest{key: templatesResult[key]}
			} else {
				templates = a.getDocumentsByDefaultTemplates(templatesResult)
			}
			break
		}
	} else {
		templates = a.getDocumentsByDefaultTemplates(templatesResult)
	}
	return templates
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
