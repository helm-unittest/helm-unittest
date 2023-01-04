package validators_test

import (
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/lrills/helm-unittest/internal/common"
	. "github.com/lrills/helm-unittest/pkg/unittest/validators"

	"github.com/stretchr/testify/assert"
)

var docToTestContains = `
a:
  b:
    - c: hello world
    - d: foo bar
    - e: bar
    - e: bar
`

var docToTestContains2 = `
a:
  b:
    - d: foo bar
`

func TestContainsValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"d": "foo bar"},
		nil,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMultiManifestContainsValidatorWhenOk(t *testing.T) {
	manifest1 := makeManifest(docToTestContains)
	manifest2 := makeManifest(docToTestContains2)

	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"d": "foo bar"},
		nil,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:  []common.K8sManifest{manifest1, manifest2},
		Index: -1,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWithValueOnlyWhenOk(t *testing.T) {
	docToTestContainsValueOnly := `
a:
  b:
    - VALUE1
    - VALUE2
`
	manifest := makeManifest(docToTestContainsValueOnly)

	validator := ContainsValidator{
		"a.b",
		"VALUE1",
		nil,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWithAnyWhenOk(t *testing.T) {
	docToTestContainsAny := `
a:
  b:
    - name: VALUE1
      value: bla
    - name: VALUE2
      value: bla2
`
	manifest := makeManifest(docToTestContainsAny)

	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"name": "VALUE1"},
		nil,
		true,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWithAnyWhenNotFoundOk(t *testing.T) {
	docToTestContainsAny := `
a:
  b:
    - name: VALUE1
      value: bla
    - name: VALUE2
      value: bla2
`
	manifest := makeManifest(docToTestContainsAny)

	// Enable debug logging
	log.SetLevel(log.DebugLevel)

	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"name": "VALUE3"},
		nil,
		true,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	- name: VALUE3",
		"Actual:",
		"	- name: VALUE1",
		"	  value: bla",
		"	- name: VALUE2",
		"	  value: bla2",
	}, diff)
}

func TestContainsValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"d": "hello bar"},
		nil,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"e": "bar bar"},
		nil,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	- e: bar bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
		"	- e: bar",
		"	- e: bar",
	}, diff)
}

func TestContainsValidatorMultiManifestWhenFail(t *testing.T) {
	manifest1 := makeManifest(docToTestContains)
	extraDoc := `
a:
  b:
    - c: hello world
`
	manifest2 := makeManifest(extraDoc)
	manifests := []common.K8sManifest{manifest1, manifest2}

	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"d": "foo bar"},
		nil,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:  manifests,
		Index: -1,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	1",
		"Path:	a.b",
		"Expected to contain:",
		"	- d: foo bar",
		"Actual:",
		"	- c: hello world",
	}, diff)
}

func TestContainsValidatorMultiManifestWhenBothFail(t *testing.T) {
	manifest1 := makeManifest(docToTestContains)
	manifests := []common.K8sManifest{manifest1, manifest1}

	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"e": "foo bar"},
		nil,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:  manifests,
		Index: -1,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	- e: foo bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
		"	- e: bar",
		"	- e: bar",
		"DocumentIndex:	1",
		"Path:	a.b",
		"Expected to contain:",
		"	- e: foo bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
		"	- e: bar",
		"	- e: bar",
	}, diff)
}

func TestContainsValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"d": "foo bar"},
		nil,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b",
		"Expected NOT to contain:",
		"	- d: foo bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
		"	- e: bar",
		"	- e: bar",
	}, diff)
}

func TestMatchContainsValidatorWhenNotAnArray(t *testing.T) {
	manifestDocNotArray := `
a:
  b:
    c: hello world
    d: foo bar
`
	manifest := makeManifest(manifestDocNotArray)

	validator := ContainsValidator{
		"a.b",
		common.K8sManifest{"d": "foo bar"},
		nil,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	expect 'a.b' to be an array, got:",
		"	c: hello world",
		"	d: foo bar",
	}, diff)
}

func TestContainsValidatorWhenInvalidIndex(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		"a.b",
		common.K8sManifest{"d": "foo bar"},
		nil,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:  []common.K8sManifest{manifest},
		Index: 2,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Error:",
		"	documentIndex 2 out of range",
	}, diff)
}

func TestContainsValidatorWhenInvalidPath(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		"a.b.e",
		common.K8sManifest{"e": "bar"},
		nil,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	can't get [\"e\"] from a non map type:",
		"	- c: hello world",
		"	- d: foo bar",
		"	- e: bar",
		"	- e: bar",
	}, diff)
}

func TestContainsValidatorWhenMultipleTimesInArray(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	counter := new(int)
	*counter = 2
	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"e": "bar"},
		counter,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorInverseWhenNotMultipleTimesInArray(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	counter := new(int)
	*counter = 1
	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"e": "bar"},
		counter,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWhenNotMultipleTimesInArray(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	counter := new(int)
	*counter = 1
	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"e": "bar"},
		counter,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	expect count 1 in 'a.b' to be in array, got 2:",
		"	- c: hello world",
		"	- d: foo bar",
		"	- e: bar",
		"	- e: bar",
	}, diff)
}

func TestContainsValidatorWhenNotFoundMultipleTimesInArray(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	counter := new(int)
	*counter = 1
	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"f": "bar"},
		counter,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	- f: bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
		"	- e: bar",
		"	- e: bar",
	}, diff)
}

func TestContainsValidatorInverseWhenNotFoundMultipleTimesInArray(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	counter := new(int)
	*counter = 1
	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"f": "bar"},
		counter,
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}
