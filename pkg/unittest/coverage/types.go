package coverage

type ProbeKind string

const (
	ProbeAction ProbeKind = "action"
	ProbeBranch ProbeKind = "branch"
	ProbeLoop   ProbeKind = "loop"
)

type Probe struct {
	Kind         ProbeKind
	TemplateName string
	Line         int
	Col          int
	Label        string
}

type TemplateMeta struct {
	Name       string
	ProbeIdxs  []int
	ParseError error
	Source     []byte
}

type FileCoverage struct {
	Name        string
	ParseError  error
	Actions     CountStat
	Branches    CountStat
	Loops       CountStat
	MissedLines []int
	Lines       []LineCoverage
	// Source is kept only for the HTML report's source view; JSON omits it.
	Source []byte `json:"-"`
	// Rendered is true if the template produced output itself OR any of its probes
	// were exercised (the latter catches _*.tpl partials called via include).
	Rendered bool
}

type LineCoverage struct {
	Line int
	// Hits is the max hit count across the probes on this line.
	Hits          int
	Branches      []BranchCoverage
	ProbesCovered int
	ProbesTotal   int
}

type BranchCoverage struct {
	Label string
	Hits  int
}

type CountStat struct {
	Covered int
	Total   int
	// Hits is the summed execution count (loops distinguish 1 vs many); for action/branch probes it is >= Covered.
	Hits int64
}

// Pct returns the coverage percentage [0,100], or -1 when Total == 0 (N/A).
func (c CountStat) Pct() float64 {
	if c.Total == 0 {
		return -1
	}
	return 100.0 * float64(c.Covered) / float64(c.Total)
}

type Coverage struct {
	ChartName string
	Files     []FileCoverage
	Totals    struct {
		Actions  CountStat
		Branches CountStat
		Loops    CountStat
	}
}
