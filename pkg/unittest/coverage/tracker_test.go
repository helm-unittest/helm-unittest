package coverage

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v3chart "helm.sh/helm/v3/pkg/chart"
	v3util "helm.sh/helm/v3/pkg/chartutil"
	v3engine "helm.sh/helm/v3/pkg/engine"
)

// buildTestChart builds an in-memory chart with the supplied templates.
// Each template entry maps a Name (e.g. "templates/foo.yaml") to its raw body.
func buildTestChart(name string, templates map[string]string) *v3chart.Chart {
	c := &v3chart.Chart{
		Metadata: &v3chart.Metadata{
			Name:       name,
			Version:    "0.0.1",
			APIVersion: v3chart.APIVersionV2,
		},
	}
	for n, body := range templates {
		c.Templates = append(c.Templates, &v3chart.File{Name: n, Data: []byte(body)})
	}
	return c
}

func TestTracker_BranchHitsDifferByValues(t *testing.T) {
	chart := buildTestChart("covtest", map[string]string{
		"templates/cm.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
{{- if .Values.enabled }}
  flag: "on"
{{- else }}
  flag: "off"
{{- end }}
{{- range $k, $v := .Values.items }}
  {{ $k }}: {{ $v | quote }}
{{- end }}
`,
	})

	tracker := NewTracker(chart)
	require.True(t, tracker.HasProbes())

	render := func(values map[string]any) {
		instrumented := tracker.InstrumentedChart()
		vals, err := v3util.ToRenderValues(instrumented, values, v3util.ReleaseOptions{
			Name:      "rel",
			Namespace: "ns",
			IsInstall: true,
		}, nil)
		require.NoError(t, err)
		out, err := v3engine.Render(instrumented, vals)
		require.NoError(t, err)
		tracker.Absorb(out)
	}

	render(map[string]any{"enabled": true, "items": map[string]any{"a": "1"}})
	render(map[string]any{"enabled": false})

	cov := tracker.Snapshot()
	require.Len(t, cov.Files, 1)
	f := cov.Files[0]

	// Both if and else branches were hit across the two renders.
	assert.Equal(t, 2, f.Branches.Covered, "if + else should both have been hit")
	assert.Equal(t, 2, f.Branches.Total)
	// Range body was hit at least once (in the first render).
	assert.Equal(t, 1, f.Loops.Covered)
	assert.Equal(t, 1, f.Loops.Total)
	// Totals should mirror file aggregates for a single-file chart.
	assert.Equal(t, f.Actions.Total, cov.Totals.Actions.Total)
	assert.Equal(t, f.Branches.Covered, cov.Totals.Branches.Covered)
}

func TestTracker_UncoveredLinesReported(t *testing.T) {
	chart := buildTestChart("covtest", map[string]string{
		"templates/cm.yaml": `apiVersion: v1
kind: ConfigMap
data:
{{- if .Values.enabled }}
  yes: "1"
{{- else }}
  no: "1"
{{- end }}
`,
	})

	tracker := NewTracker(chart)
	instrumented := tracker.InstrumentedChart()

	vals, err := v3util.ToRenderValues(instrumented, map[string]any{"enabled": true}, v3util.ReleaseOptions{
		Name:      "rel",
		Namespace: "ns",
		IsInstall: true,
	}, nil)
	require.NoError(t, err)
	out, err := v3engine.Render(instrumented, vals)
	require.NoError(t, err)
	tracker.Absorb(out)

	cov := tracker.Snapshot()
	require.Len(t, cov.Files, 1)
	// The else branch was not exercised; its source line must be reported.
	assert.NotEmpty(t, cov.Files[0].MissedLines)
}

func TestTracker_RenderedFlag(t *testing.T) {
	// Chart with three templates:
	//   live.yaml  — renders unconditionally (always non-empty)
	//   gated.yaml — content wrapped in `{{ if .Values.on }}…{{ end }}`
	//   static.yaml — plain YAML with no Go-template actions at all
	chart := buildTestChart("covtest", map[string]string{
		"templates/live.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: live
`,
		"templates/gated.yaml": `{{- if .Values.on }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: gated
{{- end }}
`,
		"templates/static.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: static
`,
	})

	tracker := NewTracker(chart)
	instrumented := tracker.InstrumentedChart()

	// Render with .Values.on absent so gated.yaml produces no content.
	vals, err := v3util.ToRenderValues(instrumented, map[string]any{}, v3util.ReleaseOptions{
		Name: "rel", Namespace: "ns", IsInstall: true,
	}, nil)
	require.NoError(t, err)
	out, err := v3engine.Render(instrumented, vals)
	require.NoError(t, err)
	tracker.Absorb(out)

	cov := tracker.Snapshot()
	byName := map[string]FileCoverage{}
	for _, f := range cov.Files {
		byName[f.Name] = f
	}

	assert.True(t, byName["covtest/templates/live.yaml"].Rendered,
		"live.yaml renders content unconditionally")
	assert.True(t, byName["covtest/templates/static.yaml"].Rendered,
		"static.yaml has no probes but renders non-empty content")
	assert.False(t, byName["covtest/templates/gated.yaml"].Rendered,
		"gated.yaml renders empty when .Values.on is false")
}

func TestTracker_WithSubcharts_FalseSkipsSubchartTemplates(t *testing.T) {
	// Parent chart that includes a helper defined in a subchart.
	parent := buildTestChart("parentchart", map[string]string{
		"templates/cm.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    team: {{ include "sub.label" . }}
`,
	})
	sub := buildTestChart("subchart", map[string]string{
		"templates/_helpers.tpl": `{{- define "sub.label" -}}
{{- .Values.team | default "default" -}}
{{- end -}}`,
		"templates/svc.yaml": `apiVersion: v1
kind: Service
metadata:
  name: sub-svc
`,
	})
	parent.SetDependencies(sub)

	tracker := NewTracker(parent, WithSubcharts(false))
	instrumented := tracker.InstrumentedChart()

	// The instrumented chart must still carry the subchart so the parent's
	// `include` call resolves at render time.
	require.Len(t, instrumented.Dependencies(), 1, "subchart still attached")

	vals, err := v3util.ToRenderValues(instrumented, map[string]any{}, v3util.ReleaseOptions{
		Name: "rel", Namespace: "ns", IsInstall: true,
	}, nil)
	require.NoError(t, err)
	out, err := v3engine.Render(instrumented, vals)
	require.NoError(t, err)
	tracker.Absorb(out)

	cov := tracker.Snapshot()
	for _, f := range cov.Files {
		assert.NotContains(t, f.Name, "/charts/subchart/",
			"subchart template %q must not appear when WithSubcharts(false)", f.Name)
	}
	// Parent chart's own template is still there.
	var sawParent bool
	for _, f := range cov.Files {
		if f.Name == "parentchart/templates/cm.yaml" {
			sawParent = true
		}
	}
	assert.True(t, sawParent, "parent template should still be reported")
}

func TestTracker_WithSubcharts_DefaultIncludes(t *testing.T) {
	// Same chart but with subcharts enabled (default) — subchart templates
	// must show up in the report.
	parent := buildTestChart("parentchart", map[string]string{
		"templates/cm.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    team: {{ include "sub.label" . }}
`,
	})
	sub := buildTestChart("subchart", map[string]string{
		"templates/_helpers.tpl": `{{- define "sub.label" -}}
{{- .Values.team | default "default" -}}
{{- end -}}`,
	})
	parent.SetDependencies(sub)

	tracker := NewTracker(parent) // default: WithSubcharts(true)
	cov := tracker.Snapshot()

	var sawSub bool
	for _, f := range cov.Files {
		if strings.Contains(f.Name, "/charts/subchart/") {
			sawSub = true
		}
	}
	assert.True(t, sawSub, "subchart template should appear by default")
}

func TestTracker_PartialTemplateInstrumented(t *testing.T) {
	chart := buildTestChart("covtest", map[string]string{
		"templates/_helpers.tpl": `{{- define "covtest.label" -}}
{{- if .Values.team }}{{ .Values.team }}{{- else }}default{{- end }}
{{- end -}}`,
		"templates/cm.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
  labels:
    team: {{ include "covtest.label" . }}
`,
	})

	tracker := NewTracker(chart)
	instrumented := tracker.InstrumentedChart()
	vals, err := v3util.ToRenderValues(instrumented, map[string]any{}, v3util.ReleaseOptions{
		Name:      "rel",
		Namespace: "ns",
		IsInstall: true,
	}, nil)
	require.NoError(t, err)
	out, err := v3engine.Render(instrumented, vals)
	require.NoError(t, err)
	tracker.Absorb(out)

	cov := tracker.Snapshot()
	// Both files should appear in the report.
	require.Len(t, cov.Files, 2)
	// Only the else branch in _helpers.tpl was exercised (no .Values.team).
	var helperFile *FileCoverage
	for i := range cov.Files {
		if cov.Files[i].Name == "covtest/templates/_helpers.tpl" {
			helperFile = &cov.Files[i]
		}
	}
	require.NotNil(t, helperFile, "helpers.tpl should be tracked")
	assert.Equal(t, 1, helperFile.Branches.Covered, "only the else branch ran")
	assert.Equal(t, 2, helperFile.Branches.Total)
}
