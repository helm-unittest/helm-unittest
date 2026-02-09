package validators_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSnapshotValidatorWhenOk(t *testing.T) {
	data := common.K8sManifest{"a": "b"}
	validator := MatchSnapshotValidator{Path: "a"}

	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", "b").Return(&snapshot.CompareResult{
		Passed: true,
	})

	pass, diff := validator.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{data},
		SnapshotComparer: mockComparer,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	mockComparer.AssertExpectations(t)
}

func TestSnapshotValidatorWhenMultiDocOk(t *testing.T) {
	data := common.K8sManifest{"a": "b"}
	validator := MatchSnapshotValidator{Path: "a"}

	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", "b").Return(&snapshot.CompareResult{
		Passed: true,
	})

	pass, diff := validator.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{data, data},
		SnapshotComparer: mockComparer,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	mockComparer.AssertExpectations(t)
}

func TestSnapshotValidatorWhenNegativeAndOk(t *testing.T) {
	data := common.K8sManifest{"a": "b"}
	validator := MatchSnapshotValidator{Path: "a"}

	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", "b").Return(&snapshot.CompareResult{
		Passed:         false,
		CachedSnapshot: "a:\n  b: c\n",
		NewSnapshot:    "x:\n  y: x\n",
	})

	pass, diff := validator.Validate(&ValidateContext{
		Negative:         true,
		Docs:             []common.K8sManifest{data},
		SnapshotComparer: mockComparer,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	mockComparer.AssertExpectations(t)
}

func TestSnapshotValidatorWhenFail(t *testing.T) {
	data := common.K8sManifest{"a": "b"}

	log.SetLevel(log.DebugLevel)

	validator := MatchSnapshotValidator{Path: "a"}

	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", "b").Return(&snapshot.CompareResult{
		Passed:         false,
		CachedSnapshot: "a:\n  b: c\n",
		NewSnapshot:    "x:\n  y: x\n",
	})

	pass, diff := validator.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{data},
		SnapshotComparer: mockComparer,
		FailFast:         true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a",
		"Expected to match snapshot 0:",
		"	--- Expected",
		"	+++ Actual",
		"	@@ -1,3 +1,3 @@",
		"	-a:",
		"	-  b: c",
		"	+x:",
		"	+  y: x",
	}, diff)

	mockComparer.AssertExpectations(t)
}

func TestSnapshotValidatorWhenNegativeAndFail(t *testing.T) {
	data := common.K8sManifest{"a": "b"}
	validator := MatchSnapshotValidator{Path: "a"}

	cached := "a:\n  b: c\n"
	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", "b").Return(&snapshot.CompareResult{
		Passed:         true,
		CachedSnapshot: cached,
		NewSnapshot:    cached,
	})

	pass, diff := validator.Validate(&ValidateContext{
		Negative:         true,
		Docs:             []common.K8sManifest{data},
		SnapshotComparer: mockComparer,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a",
		"Expected NOT to match snapshot 0:",
		"	a:",
		"	  b: c",
	}, diff)

	mockComparer.AssertExpectations(t)
}

func TestSnapshotValidatorWhenInvalidPath(t *testing.T) {
	manifest := makeManifest("a: b")

	cached := "a:\n  b: c\n"
	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", "b").Return(&snapshot.CompareResult{
		Passed:         true,
		CachedSnapshot: cached,
		NewSnapshot:    cached,
	})

	validator := MatchSnapshotValidator{Path: "a[b]"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{manifest},
		SnapshotComparer: mockComparer,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	invalid array index [b] before position 4: non-integer array index",
	}, diff)
}

func TestSnapshotValidatorWhenUnknownPath(t *testing.T) {
	manifest := makeManifest("a: b")

	cached := "a:\n  b: c\n"
	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", "b").Return(&snapshot.CompareResult{
		Passed:         true,
		CachedSnapshot: cached,
		NewSnapshot:    cached,
	})

	validator := MatchSnapshotValidator{Path: "a[3]"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{manifest},
		SnapshotComparer: mockComparer,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	unknown path a[3]",
	}, diff)
}

func TestSnapshotValidatorWhenUnknownPathNegative(t *testing.T) {
	manifest := makeManifest("a: b")

	cached := "a:\n  b: c\n"
	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", "b").Return(&snapshot.CompareResult{
		Passed:         true,
		CachedSnapshot: cached,
		NewSnapshot:    cached,
	})

	validator := MatchSnapshotValidator{Path: "a[3]"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{manifest},
		SnapshotComparer: mockComparer,
		Negative:         true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestSnapshotValidatorWhenNoManifestOk(t *testing.T) {
	validator := MatchSnapshotValidator{Path: "a"}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestSnapshotValidatorWhenNoManifestNegativeFail(t *testing.T) {
	validator := MatchSnapshotValidator{Path: "a"}

	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestSnapshotValidatorWhenMatchRegexSuccess(t *testing.T) {
	manifest := makeManifest("a: abrakadabra")

	cached := "a: abrakadabra"
	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", manifest).Return(&snapshot.CompareResult{
		Passed:         true,
		CachedSnapshot: cached,
		NewSnapshot:    cached,
	})

	validator := MatchSnapshotValidator{MatchRegex: &MatchRegex{Pattern: "abrakadabra"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{manifest},
		SnapshotComparer: mockComparer,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestSnapshotValidatorWhenMatchRegexFail(t *testing.T) {
	manifest := makeManifest("a: abrakadabra")

	cached := "a: abrakadabra"
	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", manifest).Return(&snapshot.CompareResult{
		Passed:         false,
		CachedSnapshot: cached,
		NewSnapshot:    cached,
		Msg:            " pattern 'hello' not found in snapshot",
	})

	validator := MatchSnapshotValidator{MatchRegex: &MatchRegex{Pattern: "hello"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{manifest},
		SnapshotComparer: mockComparer,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{"DocumentIndex:\t0", "ValuesIndex:\t0", "Expected pattern 'hello' not found in snapshot:", "\ta: abrakadabra"}, diff)
}

func TestSnapshotValidatorWheNotMatchRegexSuccess(t *testing.T) {
	manifest := makeManifest("a: abrakadabra")

	cached := "a: abrakadabra"
	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", manifest).Return(&snapshot.CompareResult{
		Passed:         true,
		CachedSnapshot: cached,
		NewSnapshot:    cached,
	})

	validator := MatchSnapshotValidator{NotMatchRegex: &NotMatchRegex{Pattern: "one-two-three"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{manifest},
		SnapshotComparer: mockComparer,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestSnapshotValidatorWheNotMatchRegexFail(t *testing.T) {
	manifest := makeManifest("a: abrakadabra")

	cached := "a: abrakadabra"
	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", manifest).Return(&snapshot.CompareResult{
		Passed:         false,
		CachedSnapshot: cached,
		NewSnapshot:    cached,
		Msg:            " pattern '.*abra.*' should not be in snapshot",
	})

	validator := MatchSnapshotValidator{NotMatchRegex: &NotMatchRegex{Pattern: ".*abra.*"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{manifest},
		SnapshotComparer: mockComparer,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{"DocumentIndex:\t0", "ValuesIndex:\t0", "Expected pattern '.*abra.*' should not be in snapshot:", "\ta: abrakadabra"}, diff)
}

// Tests for YAML format

func TestSnapshotValidatorYAMLFormatWhenOk(t *testing.T) {
	data1 := common.K8sManifest{"apiVersion": "v1", "kind": "ConfigMap", "metadata": map[string]any{"name": "test"}}
	data2 := common.K8sManifest{"apiVersion": "v1", "kind": "Service", "metadata": map[string]any{"name": "test-svc"}}
	validator := MatchSnapshotValidator{Format: "yaml"}

	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshotYAML", []any{data1, data2}).Return(&snapshot.CompareResult{
		Passed: true,
	})

	pass, diff := validator.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{data1, data2},
		SnapshotComparer: mockComparer,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	mockComparer.AssertExpectations(t)
}

func TestSnapshotValidatorYAMLFormatWhenFail(t *testing.T) {
	data1 := common.K8sManifest{"apiVersion": "v1", "kind": "ConfigMap"}
	data2 := common.K8sManifest{"apiVersion": "v1", "kind": "Service"}
	validator := MatchSnapshotValidator{Format: "yaml"}

	cachedSnapshot := "---\napiVersion: v1\nkind: ConfigMap\n---\napiVersion: v1\nkind: Deployment\n"
	newSnapshot := "---\napiVersion: v1\nkind: ConfigMap\n---\napiVersion: v1\nkind: Service\n"

	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshotYAML", []any{data1, data2}).Return(&snapshot.CompareResult{
		Passed:         false,
		CachedSnapshot: cachedSnapshot,
		NewSnapshot:    newSnapshot,
	})

	pass, diff := validator.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{data1, data2},
		SnapshotComparer: mockComparer,
	})

	assert.False(t, pass)
	assert.NotEmpty(t, diff)

	mockComparer.AssertExpectations(t)
}

func TestSnapshotValidatorYAMLFormatWithPath(t *testing.T) {
	data := common.K8sManifest{"spec": map[string]any{"replicas": 3}}
	validator := MatchSnapshotValidator{Path: "spec", Format: "yaml"}

	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshotYAML", []any{map[string]any{"replicas": 3}}).Return(&snapshot.CompareResult{
		Passed: true,
	})

	pass, diff := validator.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{data},
		SnapshotComparer: mockComparer,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	mockComparer.AssertExpectations(t)
}

func TestSnapshotValidatorYAMLFormatWhenNoManifests(t *testing.T) {
	validator := MatchSnapshotValidator{Format: "yaml"}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{},
	})

	assert.False(t, pass)
	assert.Contains(t, diff[0], "Error")
}

func TestSnapshotValidatorYAMLFormatNegativeAndOk(t *testing.T) {
	data := common.K8sManifest{"apiVersion": "v1", "kind": "ConfigMap"}
	validator := MatchSnapshotValidator{Format: "yaml"}

	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshotYAML", []any{data}).Return(&snapshot.CompareResult{
		Passed:         false,
		CachedSnapshot: "old-content",
		NewSnapshot:    "new-content",
	})

	pass, diff := validator.Validate(&ValidateContext{
		Negative:         true,
		Docs:             []common.K8sManifest{data},
		SnapshotComparer: mockComparer,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	mockComparer.AssertExpectations(t)
}
