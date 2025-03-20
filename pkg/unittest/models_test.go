package unittest

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/stretchr/testify/assert"
	v3chart "helm.sh/helm/v3/pkg/chart"
)

func TestNewTestConfig(t *testing.T) {
	chart := &v3chart.Chart{}
	cache := &snapshot.Cache{}
	config := NewTestConfig(chart, cache)

	assert.NotNil(t, config)
	assert.Equal(t, chart, config.targetChart)
	assert.Equal(t, cache, config.cache)
	assert.Equal(t, "", config.renderPath)
	assert.False(t, config.failFast)
	assert.Equal(t, PostRendererConfig{}, config.postRenderer)
}

func TestWithFailFast(t *testing.T) {
	chart := &v3chart.Chart{}
	cache := &snapshot.Cache{}
	config := NewTestConfig(chart, cache, WithFailFast(true))

	assert.True(t, config.failFast)
}

func TestWithRenderPath(t *testing.T) {
	chart := &v3chart.Chart{}
	cache := &snapshot.Cache{}
	config := NewTestConfig(chart, cache, WithRenderPath("/path/to/render"))

	assert.Equal(t, "/path/to/render", config.renderPath)
}

func TestWithPostRendererConfig(t *testing.T) {
	chart := &v3chart.Chart{}
	cache := &snapshot.Cache{}
	postRendererConfig := PostRendererConfig{}
	config := NewTestConfig(chart, cache, WithPostRendererConfig(postRendererConfig))

	assert.Equal(t, postRendererConfig, config.postRenderer)
}

func TestAssertionConfigBuilder(t *testing.T) {
	builder := AssertionConfigBuilder{
		TemplatesResult:  map[string][]common.K8sManifest{},
		SnapshotComparer: nil,
		RenderSucceed:    true,
		FailFast:         true,
		DidPostRender:    true,
		RenderError:      nil,
	}

	config := builder.Build()

	assert.NotNil(t, config)
	assert.Equal(t, builder.TemplatesResult, config.templatesResult)
	assert.Equal(t, builder.SnapshotComparer, config.snapshotComparer)
	assert.True(t, config.renderSucceed)
	assert.True(t, config.failFast)
	assert.True(t, config.didPostRender)
	assert.Nil(t, config.renderError)
}
