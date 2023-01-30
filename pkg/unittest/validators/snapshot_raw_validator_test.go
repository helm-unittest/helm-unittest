package validators_test

import (
	"testing"

	"github.com/lrills/helm-unittest/internal/common"
	"github.com/lrills/helm-unittest/pkg/unittest/snapshot"
	. "github.com/lrills/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSnapshotRawValidatorWhenOk(t *testing.T) {
	data := common.K8sManifest{common.RAW: "b"}
	validator := MatchSnapshotRawValidator{}

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

func TestSnapshotRawValidatorWhenNegativeAndOk(t *testing.T) {
	data := common.K8sManifest{common.RAW: "b"}
	validator := MatchSnapshotRawValidator{}

	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", "b").Return(&snapshot.CompareResult{
		Passed:         false,
		CachedSnapshot: "b\n",
		NewSnapshot:    "x\n",
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

func TestSnapshotRawValidatorWhenFail(t *testing.T) {
	data := common.K8sManifest{common.RAW: "b"}

	log.SetLevel(log.DebugLevel)

	validator := MatchSnapshotRawValidator{}

	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", "b").Return(&snapshot.CompareResult{
		Passed:         false,
		CachedSnapshot: "b\n",
		NewSnapshot:    "x\n",
	})

	pass, diff := validator.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{data},
		SnapshotComparer: mockComparer,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected to match snapshot 0:",
		"	--- Expected",
		"	+++ Actual",
		"	@@ -1,2 +1,2 @@",
		"	-b",
		"	+x",
	}, diff)

	mockComparer.AssertExpectations(t)
}

func TestSnapshotRawValidatorWhenNegativeAndFail(t *testing.T) {
	data := common.K8sManifest{common.RAW: "b"}
	validator := MatchSnapshotRawValidator{}

	cached := "b\n"
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
		"Expected NOT to match snapshot 0:",
		"	b",
	}, diff)

	mockComparer.AssertExpectations(t)
}

func TestSnapshotRawValidatorWhenInvalidIndex(t *testing.T) {
	data := common.K8sManifest{common.RAW: "b"}

	validator := MatchSnapshotRawValidator{}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:  []common.K8sManifest{data},
		Index: 2,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Error:",
		"	documentIndex 2 out of range",
	}, diff)
}
