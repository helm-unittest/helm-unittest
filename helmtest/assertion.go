package helmtest

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

type Assertion struct {
	File          string
	DocumentIndex int
	Not           bool
	AssertType    string
	validator     Validatable
	antonym       bool
}

func (a Assertion) Assert(docs map[string][]K8sManifest, result *AssertionResult) *AssertionResult {
	result.AssertType = a.AssertType
	result.Not = a.Not

	if file, ok := docs[a.File]; ok {
		result.Passed, result.FailInfo = a.validator.Validate(
			file,
			a.DocumentIndex,
			a.Not != a.antonym,
		)
		return result
	}

	result.FailInfo = []string{"Error:", a.noFileErrMessage()}
	return result
}

func (a Assertion) noFileErrMessage() string {
	if a.File != "" {
		return fmt.Sprintf(
			"\tfile \"%s\" not exists or not selected in test suite",
			a.File,
		)
	}
	return "\tassertion.file must be given if testsuite.templates is empty"
}

func (a *Assertion) UnmarshalYAML(unmarshal func(interface{}) error) error {
	assertDef := make(map[string]interface{})
	if err := unmarshal(&assertDef); err != nil {
		return err
	}

	if documentIndex, ok := assertDef["documentIndex"].(int); ok {
		a.DocumentIndex = documentIndex
	}
	if not, ok := assertDef["not"].(bool); ok {
		a.Not = not
	}
	if file, ok := assertDef["file"].(string); ok {
		a.File = file
	}

	if err := a.constructValidator(assertDef); err != nil {
		return err
	}

	if a.validator == nil {
		for key := range assertDef {
			if key != "file" && key != "documentIndex" && key != "not" {
				return fmt.Errorf("Assertion type `%s` is invalid", key)
			}
		}
		return fmt.Errorf("No assertion type defined")
	}

	return nil
}

func (a *Assertion) constructValidator(assertDef map[string]interface{}) error {
	for assertName, correspondDef := range assertTypeMapping {
		if params, ok := assertDef[assertName]; ok {
			if a.validator != nil {
				return fmt.Errorf(
					"Assertion type `%s` and `%s` is declared duplicately",
					a.AssertType,
					assertName,
				)
			}

			validator := reflect.New(correspondDef.validatorType).Interface()
			if err := mapstructure.Decode(params, validator); err != nil {
				return err
			}

			a.AssertType = assertName
			a.validator = validator.(Validatable)
			a.antonym = correspondDef.antonym
		}
	}
	return nil
}

type assertTypeDef struct {
	validatorType reflect.Type
	antonym       bool
}

var assertTypeMapping = map[string]assertTypeDef{
	// "matchSnapshot": assertTypeDef{reflect.TypeOf(MatchSnapshotValidator{}), false},
	"equal":         assertTypeDef{reflect.TypeOf(EqualValidator{}), false},
	"notEqual":      assertTypeDef{reflect.TypeOf(EqualValidator{}), true},
	"matchRegex":    assertTypeDef{reflect.TypeOf(MatchRegexValidator{}), false},
	"notMatchRegex": assertTypeDef{reflect.TypeOf(MatchRegexValidator{}), true},
	"contains":      assertTypeDef{reflect.TypeOf(ContainsValidator{}), false},
	"notContains":   assertTypeDef{reflect.TypeOf(ContainsValidator{}), true},
	"isNull":        assertTypeDef{reflect.TypeOf(IsNullValidator{}), false},
	"isNotNull":     assertTypeDef{reflect.TypeOf(IsNullValidator{}), true},
	"isEmpty":       assertTypeDef{reflect.TypeOf(IsEmptyValidator{}), false},
	"isNotEmpty":    assertTypeDef{reflect.TypeOf(IsEmptyValidator{}), true},
	"isKind":        assertTypeDef{reflect.TypeOf(IsKindValidator{}), false},
	"isAPIVersion":  assertTypeDef{reflect.TypeOf(IsAPIVersionValidator{}), false},
	"hasDocuments":  assertTypeDef{reflect.TypeOf(HasDocumentsValidator{}), false},
}
