package main

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

type TestJob struct {
	Name       string `yaml:"it"`
	Values     []string
	Set        map[string]interface{}
	Assertions []Assertion `yaml:"asserts"`
}

type Assertion struct {
	DocumentIndex int
	Not           bool
	AssertType    string
	asserter      Assertable
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

	for assertName, asserterDef := range asserterMapping {
		if params, ok := assertDef[assertName]; ok {
			if a.asserter != nil {
				return fmt.Errorf("Assertion type `%s` and `%s` is declared duplicated", a.AssertType, assertName)
			}
			a.AssertType = assertName
			a.configure(asserterDef.T, params, asserterDef.N)
		}
	}

	return nil
}

func (a *Assertion) configure(asserterType reflect.Type, params interface{}, not bool) error {
	asserter := reflect.New(asserterType).Interface()
	if err := mapstructure.Decode(params, asserter); err != nil {
		return err
	}
	a.asserter = asserter.(Assertable)
	a.Not = a.Not != not
	return nil
}

type AsserterDef struct {
	T reflect.Type
	N bool
}

var asserterMapping = map[string]AsserterDef{
	// "matchSnapshot": AsserterDef{reflect.TypeOf(MatchSnapshotAsserter{}), false},
	"equal":         AsserterDef{reflect.TypeOf(EqualAsserter{}), false},
	"notEqual":      AsserterDef{reflect.TypeOf(EqualAsserter{}), true},
	"matchRegex":    AsserterDef{reflect.TypeOf(MatchRegexAsserter{}), false},
	"notMatchRegex": AsserterDef{reflect.TypeOf(MatchRegexAsserter{}), true},
	"contains":      AsserterDef{reflect.TypeOf(ContainsAsserter{}), false},
	"notContains":   AsserterDef{reflect.TypeOf(ContainsAsserter{}), true},
	"isNull":        AsserterDef{reflect.TypeOf(IsNullAsserter{}), false},
	"isNotNull":     AsserterDef{reflect.TypeOf(IsNullAsserter{}), true},
	"isEmpty":       AsserterDef{reflect.TypeOf(IsEmptyAsserter{}), false},
	"isNotEmpty":    AsserterDef{reflect.TypeOf(IsEmptyAsserter{}), true},
	"isKind":        AsserterDef{reflect.TypeOf(IsKindAsserter{}), false},
	"isAPIVersion":  AsserterDef{reflect.TypeOf(IsAPIVersionAsserter{}), false},
	"hasDocuments":  AsserterDef{reflect.TypeOf(HasDocumentsAsserter{}), false},
}
