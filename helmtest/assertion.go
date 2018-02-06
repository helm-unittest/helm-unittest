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
}

func (a Assertion) Assert(docs map[string][]K8sManifest) AssertionResult {
	if file, ok := docs[a.File]; ok {
		passed, info := a.validator.Validate(file, a.DocumentIndex)
		return AssertionResult{
			FailInfo:   info,
			Passed:     passed,
			AssertType: a.AssertType,
		}
	}

	var noFileErrMsg string
	if a.File != "" {
		noFileErrMsg = fmt.Sprintf(
			"\tfile \"%s\" not exists or not selected in test suite",
			a.File,
		)
	} else {
		noFileErrMsg = "\tassertion.file must be given if testsuite.templates is empty"
	}
	return AssertionResult{
		AssertType: a.AssertType,
		FailInfo:   []string{"Error:", noFileErrMsg},
	}
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
	for assertName, correspondDef := range validatorMapping {
		if params, ok := assertDef[assertName]; ok {
			if a.validator != nil {
				return fmt.Errorf("Assertion type `%s` and `%s` is declared duplicated", a.AssertType, assertName)
			}
			a.AssertType = assertName

			validator := reflect.New(correspondDef.Type).Interface()
			if err := mapstructure.Decode(params, validator); err != nil {
				return err
			}
			a.validator = validator.(Validatable)
			a.Not = a.Not != correspondDef.Not
		}
	}
	return nil
}

type validatorDef struct {
	Type reflect.Type
	Not  bool
}

var validatorMapping = map[string]validatorDef{
	// "matchSnapshot": validatorDef{reflect.TypeOf(MatchSnapshotValidator{}), false},
	"equal":         validatorDef{reflect.TypeOf(EqualValidator{}), false},
	"notEqual":      validatorDef{reflect.TypeOf(EqualValidator{}), true},
	"matchRegex":    validatorDef{reflect.TypeOf(MatchRegexValidator{}), false},
	"notMatchRegex": validatorDef{reflect.TypeOf(MatchRegexValidator{}), true},
	"contains":      validatorDef{reflect.TypeOf(ContainsValidator{}), false},
	"notContains":   validatorDef{reflect.TypeOf(ContainsValidator{}), true},
	"isNull":        validatorDef{reflect.TypeOf(IsNullValidator{}), false},
	"isNotNull":     validatorDef{reflect.TypeOf(IsNullValidator{}), true},
	"isEmpty":       validatorDef{reflect.TypeOf(IsEmptyValidator{}), false},
	"isNotEmpty":    validatorDef{reflect.TypeOf(IsEmptyValidator{}), true},
	"isKind":        validatorDef{reflect.TypeOf(IsKindValidator{}), false},
	"isAPIVersion":  validatorDef{reflect.TypeOf(IsAPIVersionValidator{}), false},
	"hasDocuments":  validatorDef{reflect.TypeOf(HasDocumentsValidator{}), false},
}
