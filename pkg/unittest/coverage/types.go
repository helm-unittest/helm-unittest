package coverage

// ProbeKind classifies what kind of template construct a probe is tracking.
type ProbeKind string

const (
	// ProbeAction is a plain {{ ... }} action (variable read, assignment, function call, template invocation).
	ProbeAction ProbeKind = "action"
	// ProbeBranch is one branch of an if/else/with construct.
	ProbeBranch ProbeKind = "branch"
	// ProbeLoop is the body (or else-body) of a range construct.
	ProbeLoop ProbeKind = "loop"
)

// Probe describes a single instrumentation point in a template.
type Probe struct {
	Kind         ProbeKind
	TemplateName string // path of the template file, e.g. "templates/deployment.yaml"
	Line         int    // 1-based line in the source
	Col          int    // 1-based column in the source
	Label        string // short human label, e.g. "if", "else", "range-body", "range-else"
}

// TemplateMeta describes the instrumentation result for one template file.
type TemplateMeta struct {
	Name       string  // template file path within the chart
	ProbeIdxs  []int   // indexes into Tracker.Probes for probes belonging to this file (in source order)
	ParseError error   // non-nil if the template could not be parsed
	Source     []byte  // original file content (kept for line lookup in reports)
}

// FileCoverage is the aggregated coverage for a single template file.
type FileCoverage struct {
	Name        string
	ParseError  error
	Actions     CountStat
	Branches    CountStat
	Loops       CountStat
	MissedLines []int          // 1-based source lines that were never executed
	Lines       []LineCoverage // per-line breakdown, sorted by Line
}

// LineCoverage is the aggregated coverage for a single source line within a
// template file. Hits is the maximum hit count across every probe that sits on
// the line — it represents "how many times this line executed". Branches lists
// each branch/loop probe on the line individually, which is what LCOV and
// Cobertura need in order to emit per-condition data.
type LineCoverage struct {
	Line     int
	Hits     int
	Branches []BranchCoverage
}

// BranchCoverage is a single branch-style probe (if / else / with / range body
// / range else) on a line.
type BranchCoverage struct {
	Label string
	Hits  int
}

// CountStat is covered/total for a coverage dimension. Hits is the sum of
// actual execution counts across all probes — useful for loops where the
// difference between "ran 1 time" and "ran 1000 times" matters. For action /
// branch probes Hits is informational and typically >= Covered.
type CountStat struct {
	Covered int
	Total   int
	Hits    int64
}

// Pct returns the coverage percentage [0,100], or -1 when Total == 0 (N/A).
func (c CountStat) Pct() float64 {
	if c.Total == 0 {
		return -1
	}
	return 100.0 * float64(c.Covered) / float64(c.Total)
}

// Coverage is the top-level coverage snapshot.
type Coverage struct {
	ChartName string
	Files     []FileCoverage
	Totals    struct {
		Actions  CountStat
		Branches CountStat
		Loops    CountStat
	}
}
