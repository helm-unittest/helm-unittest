package coverage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	chartcommon "helm.sh/helm/v4/pkg/chart/common"
	chartcommonutil "helm.sh/helm/v4/pkg/chart/common/util"
	v4engine "helm.sh/helm/v4/pkg/engine"
)

func renderChart(t *testing.T, tracker *Tracker, values map[string]any) map[string]string {
	t.Helper()
	chart := tracker.InstrumentedChart()
	vals, err := chartcommonutil.ToRenderValues(chart, values, chartcommon.ReleaseOptions{
		Name: "rel", Namespace: "ns", IsInstall: true,
	}, nil)
	require.NoError(t, err)
	eng := v4engine.Engine{CustomTemplateFuncs: tracker.ProbeFuncMap()}
	out, err := eng.Render(chart, vals)
	require.NoError(t, err)
	return out
}

func TestImplicitBranch_DefaultPipeForm(t *testing.T) {
	chart := buildTestChart("demo", map[string]string{
		"templates/cm.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Values.name | default "fallback-name" }}
`,
	})
	tracker := NewTracker(chart)

	// First render: .Values.name is set → primary branch.
	tracker.Absorb(renderChart(t, tracker, map[string]any{"name": "real"}))
	// Second render: .Values.name empty → fallback branch.
	tracker.Absorb(renderChart(t, tracker, map[string]any{}))

	cov := tracker.Snapshot()
	require.Len(t, cov.Files, 1)
	f := cov.Files[0]

	// One implicit pair (primary + fallback) should both be hit.
	assert.Equal(t, 2, f.Branches.Covered)
	assert.Equal(t, 2, f.Branches.Total)
}

func TestImplicitBranch_DefaultDirectForm(t *testing.T) {
	chart := buildTestChart("demo", map[string]string{
		"templates/cm.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ default "fallback-name" .Values.name }}
`,
	})
	tracker := NewTracker(chart)

	tracker.Absorb(renderChart(t, tracker, map[string]any{}))

	cov := tracker.Snapshot()
	require.Len(t, cov.Files, 1)
	f := cov.Files[0]
	// Only the fallback was exercised: 1/2 implicit branches covered.
	assert.Equal(t, 1, f.Branches.Covered)
	assert.Equal(t, 2, f.Branches.Total)
}

func TestImplicitBranch_TernaryBothPaths(t *testing.T) {
	chart := buildTestChart("demo", map[string]string{
		"templates/cm.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ ternary "yes" "no" .Values.flag }}
`,
	})
	tracker := NewTracker(chart)

	tracker.Absorb(renderChart(t, tracker, map[string]any{"flag": true}))
	tracker.Absorb(renderChart(t, tracker, map[string]any{"flag": false}))

	cov := tracker.Snapshot()
	f := cov.Files[0]
	assert.Equal(t, 2, f.Branches.Covered, "both ternary branches should be hit")
	assert.Equal(t, 2, f.Branches.Total)
}

func TestImplicitBranch_SkipsUnsafePipelines(t *testing.T) {
	chart := buildTestChart("demo", map[string]string{
		// `upper .Values.name` is a function call in the prefix — we MUST
		// NOT inject a side probe here, otherwise we'd run `upper` twice.
		"templates/cm.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ upper .Values.name | default "fallback" }}
`,
	})
	tracker := NewTracker(chart)

	cov := tracker.Snapshot()
	f := cov.Files[0]
	// Only the action probe exists; no implicit-branch pair was added.
	assert.Equal(t, 0, f.Branches.Total, "unsafe pipeline must not be instrumented")
	assert.Equal(t, 1, f.Actions.Total)
}

func TestLoops_IterationHitsTracked(t *testing.T) {
	chart := buildTestChart("demo", map[string]string{
		"templates/cm.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
{{- range $i, $v := .Values.items }}
  {{ $i }}: "{{ $v }}"
{{- end }}
`,
	})
	tracker := NewTracker(chart)

	tracker.Absorb(renderChart(t, tracker, map[string]any{
		"items": map[string]any{"a": "1", "b": "2", "c": "3"},
	}))

	cov := tracker.Snapshot()
	f := cov.Files[0]
	// 1 loop body, covered once (single loop probe), but Hits captures the
	// real iteration count.
	assert.Equal(t, 1, f.Loops.Total)
	assert.Equal(t, 1, f.Loops.Covered)
	assert.Equal(t, int64(3), f.Loops.Hits, "should record one hit per iteration")
}
