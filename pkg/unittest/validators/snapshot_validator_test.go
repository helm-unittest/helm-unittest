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
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
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
		"Path:	a",
		"Expected NOT to match snapshot 0:",
		"	a:",
		"	  b: c",
	}, diff)

	mockComparer.AssertExpectations(t)
}

func TestSnapshotValidatorWhenInvalidPath(t *testing.T) {
	manifest := makeManifest("a:b")

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
	manifest := makeManifest("a:b")

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
