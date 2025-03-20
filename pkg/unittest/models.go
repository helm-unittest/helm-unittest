package unittest

import (
	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	v3chart "helm.sh/helm/v3/pkg/chart"
)

type TestConfig struct {
	targetChart  *v3chart.Chart
	cache        *snapshot.Cache
	renderPath   string
	failFast     bool
	postRenderer PostRendererConfig
}

func NewTestConfig(chart *v3chart.Chart, cache *snapshot.Cache, options ...func(*TestConfig)) *TestConfig {
	config := &TestConfig{
		targetChart:  chart,
		cache:        cache,
		renderPath:   "",
		failFast:     false,
		postRenderer: PostRendererConfig{},
	}
	for _, option := range options {
		option(config)
	}
	return config
}

type LoadTestOptionsFunc func(*TestConfig)

func WithFailFast(failFast bool) LoadTestOptionsFunc {
	return func(c *TestConfig) {
		c.failFast = failFast
	}
}

func WithRenderPath(path string) LoadTestOptionsFunc {
	return func(c *TestConfig) {
		c.renderPath = path
	}
}

func WithPostRendererConfig(config PostRendererConfig) LoadTestOptionsFunc {
	return func(c *TestConfig) {
		c.postRenderer = config
	}
}

type AssertionConfig struct {
	templatesResult  map[string][]common.K8sManifest
	snapshotComparer validators.SnapshotComparer
	renderSucceed    bool
	failFast         bool
	didPostRender    bool
	renderError      error
}

// AssertionConfigBuilder Required to simplify tests
type AssertionConfigBuilder struct {
	TemplatesResult  map[string][]common.K8sManifest
	SnapshotComparer validators.SnapshotComparer
	RenderSucceed    bool
	FailFast         bool
	DidPostRender    bool
	RenderError      error
}

func (b AssertionConfigBuilder) Build() AssertionConfig {
	return AssertionConfig{
		templatesResult:  b.TemplatesResult,
		snapshotComparer: b.SnapshotComparer,
		renderSucceed:    b.RenderSucceed,
		failFast:         b.FailFast,
		didPostRender:    b.DidPostRender,
		renderError:      b.RenderError,
	}
}
