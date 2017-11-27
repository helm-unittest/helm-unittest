package main

type Assertable interface {
	assert(manifest map[string]interface{}, not bool) error
}

type TestJob struct {
	Name       string `yaml:"it"`
	Values     []string
	Set        map[string]interface{}
	Assertions []Assertion `yaml:"asserts"`
}

type Assertion struct {
	DocumentIndex int
	Not           bool
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

	if _, ok := assertDef["matchSnapshot"]; ok {
		// a.asserter :=
	} else if _, ok := assertDef["matchValue"]; ok {
		// a.asserter :=
	} else if _, ok := assertDef["matchPattern"]; ok {
		// a.asserter :=
	} else if _, ok := assertDef["contain"]; ok {
		// a.asserter :=
	} else if _, ok := assertDef["containMap"]; ok {
		// a.asserter :=
	} else if _, ok := assertDef["isNotNull"]; ok {
		// a.asserter :=
	} else if _, ok := assertDef["isNotEmpty"]; ok {
		// a.asserter :=
	} else if _, ok := assertDef["isKindOf"]; ok {
		// a.asserter :=
	} else if _, ok := assertDef["isApiVersion"]; ok {
		// a.asserter :=
	} else if _, ok := assertDef["hasDocumentsCount"]; ok {
		// a.asserter :=
	}
	return nil
}
