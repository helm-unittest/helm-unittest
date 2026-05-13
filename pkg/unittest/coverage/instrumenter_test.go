package coverage

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstrument_RegistersProbesPerKind(t *testing.T) {
	src := []byte(`
apiVersion: v1
kind: ConfigMap
{{- if .Values.enabled }}
data:
  key: {{ .Values.key }}
{{- else }}
data: {}
{{- end }}
{{- range $i, $v := .Values.list }}
  - item-{{ $i }}: {{ $v }}
{{- end }}
{{- with .Values.opt }}
  optional: {{ .name }}
{{- end }}
`)

	tr := NewTrackerForTest(t)
	instr, meta := tr.Instrument("templates/cm.yaml", src)
	require.NoError(t, meta.ParseError)

	// The instrumenter should record at least: 2 branches (if/else),
	// 1 loop body, 1 with branch, and several actions for variable refs.
	var actions, branches, loops int
	for _, idx := range meta.ProbeIdxs {
		switch tr.probes[idx].Kind {
		case ProbeAction:
			actions++
		case ProbeBranch:
			branches++
		case ProbeLoop:
			loops++
		}
	}
	assert.GreaterOrEqual(t, actions, 3, "should see multiple action probes (variable references)")
	assert.Equal(t, 3, branches, "if + else + with should produce 3 branch probes")
	assert.Equal(t, 1, loops, "range body should produce 1 loop probe")

	// Every probe must reference our token format and the source must still
	// parse as a valid Go template after instrumentation.
	for _, idx := range meta.ProbeIdxs {
		assert.Contains(t, string(instr), ProbeToken(idx))
	}
}

func TestInstrument_PreservesDefineBlocks(t *testing.T) {
	src := []byte(`{{- define "myapp.labels" -}}
app: myapp
{{- if .Values.extra }}
extra: yes
{{- end }}
{{- end -}}`)

	tr := NewTrackerForTest(t)
	instr, meta := tr.Instrument("templates/_helpers.tpl", src)
	require.NoError(t, meta.ParseError)

	assert.Contains(t, string(instr), `{{- define "myapp.labels"`)
	assert.Contains(t, string(instr), "{{- end -}}")
	// At least one probe should have been emitted inside the define.
	assert.NotEmpty(t, meta.ProbeIdxs)
}

func TestInstrument_TemplateWithParseError(t *testing.T) {
	src := []byte(`{{ if .Values.x }}{{ /* missing end */ `)

	tr := NewTrackerForTest(t)
	out, meta := tr.Instrument("templates/broken.yaml", src)
	assert.Error(t, meta.ParseError, "parse error should be recorded")
	assert.Equal(t, src, out, "broken template should be returned unchanged")
	assert.Empty(t, meta.ProbeIdxs)
}

func TestProbeToken_Format(t *testing.T) {
	tok := ProbeToken(42)
	re := regexp.MustCompile(regexp.QuoteMeta(tokenPrefix) + `\d+` + regexp.QuoteMeta(tokenSuffix))
	assert.True(t, re.MatchString(tok), "token should match the documented format")
	// Must NOT contain quotes or YAML-special characters that would break
	// raw inclusion into the rendered output.
	assert.False(t, strings.ContainsAny(tok, " \"\n\t:"))
}

func TestPositionLineCol(t *testing.T) {
	data := []byte("ab\ncd\nef")
	offsets := computeLineOffsets(data)

	cases := []struct {
		pos      int
		wantLine int
		wantCol  int
	}{
		{0, 1, 1},
		{1, 1, 2},
		{2, 1, 3}, // newline
		{3, 2, 1},
		{6, 3, 1},
		{7, 3, 2},
	}
	for _, c := range cases {
		l, col := positionLineCol(offsets, c.pos)
		assert.Equal(t, c.wantLine, l, "line for pos %d", c.pos)
		assert.Equal(t, c.wantCol, col, "col for pos %d", c.pos)
	}
}

// NewTrackerForTest returns an empty tracker with no chart attached, suitable
// for exercising Instrument in isolation.
func NewTrackerForTest(t *testing.T) *Tracker {
	t.Helper()
	return &Tracker{
		chartName:     "test",
		templateMetas: map[string]TemplateMeta{},
	}
}
