package coverage

import (
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	v3chart "helm.sh/helm/v3/pkg/chart"
)

const logField = "coverage"

// Tracker instruments a chart's templates and aggregates probe hits across
// every render performed against the instrumented chart.
//
// The tracker is goroutine-safe: probes are registered once at chart setup
// time, and Absorb may be called concurrently from multiple test jobs.
type Tracker struct {
	chartName string

	mu     sync.Mutex
	probes []Probe
	hits   []int64

	// templateMetas keeps per-template information (probe indexes, parse
	// errors). Keys are the chart-rooted template paths used in chart.File.Name
	// (e.g. "templates/deployment.yaml"). For subchart templates we use the
	// subchart-relative path to avoid collisions across charts; sub-chart files
	// always retain their own chart-root path.
	templateMetas map[string]TemplateMeta
	// templateOrder preserves the insertion order so reports list files in a
	// stable, source-like order.
	templateOrder []string

	// instrumentedChart is a deep clone of the input chart whose Templates
	// (and all subcharts' Templates) have been swapped for instrumented bytes.
	// We retain it so test jobs can reuse it without re-instrumenting per run.
	instrumentedChart *v3chart.Chart
}

// NewTracker constructs a tracker by instrumenting every template in the chart
// and its dependency subcharts. Templates that fail to parse are kept verbatim
// in the cloned chart but recorded with a ParseError in their metadata so the
// final report can surface them.
//
// The returned tracker owns its own deep copy of the chart; the input is not
// modified.
func NewTracker(chart *v3chart.Chart) *Tracker {
	t := &Tracker{
		chartName:     chart.Name(),
		templateMetas: map[string]TemplateMeta{},
	}
	t.instrumentedChart = t.deepCopyAndInstrument(chart, chart.Name())
	return t
}

// deepCopyAndInstrument mirrors FullCopyV3Chart but rewrites template Data with
// instrumented bytes. It must produce a chart whose structure is otherwise
// identical to FullCopyV3Chart's output so the engine treats it the same way.
func (t *Tracker) deepCopyAndInstrument(in *v3chart.Chart, route string) *v3chart.Chart {
	out := new(v3chart.Chart)

	// Raw files (Chart.yaml, values.yaml, ...) — copy as-is.
	for _, rf := range in.Raw {
		c := *rf
		out.Raw = append(out.Raw, &c)
	}

	if in.Metadata != nil {
		md := *in.Metadata
		out.Metadata = &md
	}

	out.Values = in.Values
	out.Schema = in.Schema

	// Static files (anything not under templates/).
	for _, f := range in.Files {
		c := *f
		out.Files = append(out.Files, &c)
	}

	// Templates — instrument each.
	for _, tmpl := range in.Templates {
		copied := &v3chart.File{Name: tmpl.Name}

		// Only instrument YAML / TPL templates. Other files (txt, json, etc.)
		// might appear in templates/ and we don't tamper with them.
		if shouldInstrument(tmpl.Name) {
			key := metaKey(route, tmpl.Name)
			instrumented, meta := t.Instrument(key, tmpl.Data)
			copied.Data = instrumented
			t.registerTemplateMeta(key, meta)
		} else {
			copied.Data = append([]byte(nil), tmpl.Data...)
		}
		out.Templates = append(out.Templates, copied)
	}

	// Recurse into dependencies, mirroring FullCopyV3Chart's route convention.
	depRoute := func(dep *v3chart.Chart) string {
		return filepath.ToSlash(filepath.Join(route, "charts", dep.Name()))
	}
	deps := make([]*v3chart.Chart, 0, len(in.Dependencies()))
	for _, d := range in.Dependencies() {
		deps = append(deps, t.deepCopyAndInstrument(d, depRoute(d)))
	}
	out.SetDependencies(deps...)

	return out
}

// metaKey is the unique per-template key used to look up coverage metadata.
// Always slash-separated so it sorts/searches predictably across platforms.
func metaKey(chartRoute, templateName string) string {
	return filepath.ToSlash(filepath.Join(chartRoute, templateName))
}

func shouldInstrument(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".yaml" || ext == ".yml" || ext == ".tpl"
}

func (t *Tracker) registerProbe(p Probe) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.probes = append(t.probes, p)
	t.hits = append(t.hits, 0)
	return len(t.probes) - 1
}

func (t *Tracker) registerTemplateMeta(key string, meta TemplateMeta) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.templateMetas[key] = meta
	t.templateOrder = append(t.templateOrder, key)
}

// InstrumentedChart returns the deep-copied, instrumented chart suitable for
// rendering through helm.sh/helm/v3/pkg/engine.
func (t *Tracker) InstrumentedChart() *v3chart.Chart {
	return t.instrumentedChart
}

// HasProbes reports whether any probe was registered. When false, rendering
// the instrumented chart will still work, but there will be nothing meaningful
// to report.
func (t *Tracker) HasProbes() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.probes) > 0
}

var probeTokenRe = regexp.MustCompile(regexp.QuoteMeta(tokenPrefix) + `(\d+)` + regexp.QuoteMeta(tokenSuffix))

// Absorb scans the output map produced by v3engine.Render for probe tokens and
// increments hit counts. It is safe to call from concurrent goroutines.
func (t *Tracker) Absorb(rendered map[string]string) {
	if len(rendered) == 0 {
		return
	}
	local := make(map[int]int)
	for _, content := range rendered {
		matches := probeTokenRe.FindAllStringSubmatch(content, -1)
		for _, m := range matches {
			idx, err := strconv.Atoi(m[1])
			if err != nil {
				continue
			}
			local[idx]++
		}
	}
	if len(local) == 0 {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	for idx, n := range local {
		if idx >= 0 && idx < len(t.hits) {
			t.hits[idx] += int64(n)
		} else {
			log.WithField(logField, "absorb").Debugf("probe index %d out of range", idx)
		}
	}
}

// Snapshot freezes the current tracker state into a stable Coverage struct.
func (t *Tracker) Snapshot() Coverage {
	t.mu.Lock()
	defer t.mu.Unlock()

	cov := Coverage{ChartName: t.chartName}

	// Group probe indexes per template for fast lookup.
	for _, key := range t.templateOrder {
		meta := t.templateMetas[key]
		fc := FileCoverage{Name: key, ParseError: meta.ParseError, Source: meta.Source}
		uncoveredLines := map[int]struct{}{}
		lineAcc := map[int]*LineCoverage{}
		for _, pi := range meta.ProbeIdxs {
			p := t.probes[pi]
			hits := int(t.hits[pi])
			hit := hits > 0
			switch p.Kind {
			case ProbeAction:
				fc.Actions.Total++
				fc.Actions.Hits += int64(hits)
				if hit {
					fc.Actions.Covered++
				}
			case ProbeBranch:
				fc.Branches.Total++
				fc.Branches.Hits += int64(hits)
				if hit {
					fc.Branches.Covered++
				}
			case ProbeLoop:
				fc.Loops.Total++
				fc.Loops.Hits += int64(hits)
				if hit {
					fc.Loops.Covered++
				}
			}
			if !hit {
				uncoveredLines[p.Line] = struct{}{}
			}

			line := lineAcc[p.Line]
			if line == nil {
				line = &LineCoverage{Line: p.Line}
				lineAcc[p.Line] = line
			}
			if hits > line.Hits {
				line.Hits = hits
			}
			line.ProbesTotal++
			if hit {
				line.ProbesCovered++
			}
			if p.Kind == ProbeBranch || p.Kind == ProbeLoop {
				line.Branches = append(line.Branches, BranchCoverage{Label: p.Label, Hits: hits})
			}
		}
		fc.MissedLines = sortedKeys(uncoveredLines)
		fc.Lines = make([]LineCoverage, 0, len(lineAcc))
		for _, lc := range lineAcc {
			fc.Lines = append(fc.Lines, *lc)
		}
		sort.Slice(fc.Lines, func(i, j int) bool { return fc.Lines[i].Line < fc.Lines[j].Line })

		cov.Totals.Actions.Total += fc.Actions.Total
		cov.Totals.Actions.Covered += fc.Actions.Covered
		cov.Totals.Actions.Hits += fc.Actions.Hits
		cov.Totals.Branches.Total += fc.Branches.Total
		cov.Totals.Branches.Covered += fc.Branches.Covered
		cov.Totals.Branches.Hits += fc.Branches.Hits
		cov.Totals.Loops.Total += fc.Loops.Total
		cov.Totals.Loops.Covered += fc.Loops.Covered
		cov.Totals.Loops.Hits += fc.Loops.Hits

		cov.Files = append(cov.Files, fc)
	}

	// Stable alphabetic order makes the report deterministic across runs.
	sort.Slice(cov.Files, func(i, j int) bool {
		return cov.Files[i].Name < cov.Files[j].Name
	})
	return cov
}

func sortedKeys(m map[int]struct{}) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
