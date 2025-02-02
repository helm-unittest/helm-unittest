package coverage

import (
	"strings"

	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	wrap "github.com/mitchellh/go-wordwrap"
	"gopkg.in/yaml.v3"
)

type CoverageReporter struct {
	templates map[string]string
	results   []ResultMap
	coverage  []CoverageEntry
	total     CoverageEntry
	rawOutput string
}

type ResultEntry struct {
	Path     string
	Actions  []int
	Branches []int
	Loops    []int
}

type CoverageEntry struct {
	Path             string
	BranchesCovered  int
	BranchesTotal    int
	BranchesMissed   string
	BranchesCoverage float32 // Percentage
	ActionCovered    int
	ActionTotal      int
	ActionMissed     string
	ActionCoverage   float32 // Percentage
	LoopsCovered     int
	LoopsTotal       int
	LoopsMissed      int
	LoopsCoverage    float32
	NotExecuted      bool
}

type ResultMap map[string]ResultEntry

type ParsedAction struct {
	Total    int
	Executed int
}

func NewTemplateResult(path string) *ResultEntry {
	return &ResultEntry{
		Path:     path,
		Actions:  []int{},
		Branches: []int{},
	}
}

func NewCoverageResult(path string) CoverageEntry {
	return CoverageEntry{
		Path:             path,
		BranchesCovered:  0,
		BranchesTotal:    0,
		BranchesMissed:   "",
		BranchesCoverage: 0,
		ActionCovered:    0,
		ActionTotal:      0,
		ActionMissed:     "",
		ActionCoverage:   0,
		LoopsCovered:     0,
		LoopsTotal:       0,
		LoopsMissed:      0,
		LoopsCoverage:    0,
		NotExecuted:      false,
	}
}

func NewCoverageReporter(coverageResults []ResultMap) *CoverageReporter {
	return &CoverageReporter{results: coverageResults}
}

func (t *ResultEntry) Extract(input string) {
	var templateYaml YamlStruct
	err := yaml.Unmarshal([]byte(input), &templateYaml)
	if err != nil {
		panic(err)
	}

	t.Actions = templateYaml.Actions
	t.Branches = templateYaml.Branches
	t.Loops = templateYaml.Loops

}

func (t *CoverageReporter) ComputeCoverage(resultMap ResultMap) {
	for key, result := range resultMap {

		var actionExecuted int
		var actionTotal int
		var actionMissed []string

		var coverageEntry CoverageEntry
		coverageEntry = NewCoverageResult(key)
		if len(result.Actions) == 0 && len(result.Branches) == 0 && len(result.Loops) == 0 {
			coverageEntry.NotExecuted = true
		}
		for i := 0; i < len(result.Actions); i++ {
			if result.Actions[i] > 0 {
				actionExecuted++
			} else {
				actionMissed = append(actionMissed, fmt.Sprintf("%d", i))
			}
			actionTotal++
		}

		coverageEntry.ActionTotal = actionTotal
		coverageEntry.ActionCovered = actionExecuted
		coverageEntry.ActionMissed = strings.Join(actionMissed, ", ")
		coverageEntry.ActionCoverage = float32(actionExecuted) / float32(actionTotal) * 100

		var branchExecuted int
		var branchTotal int
		var branchMissed []string

		for i := 0; i < len(result.Branches); i++ {
			if result.Branches[i] > 0 {
				branchExecuted++
			} else {
				branchMissed = append(branchMissed, fmt.Sprintf("%d", i))
			}
			branchTotal++
		}

		coverageEntry.BranchesTotal = branchTotal
		coverageEntry.BranchesCovered = branchExecuted
		coverageEntry.BranchesMissed = strings.Join(branchMissed, ", ")
		// We could get NaN if we have 0 branches
		// This is handled in the formatting for now, because of the `float32` type.
		coverageEntry.BranchesCoverage = float32(branchExecuted) / float32(branchTotal) * 100

		var loopExecuted int
		var loopTotal int
		for i := 0; i < len(result.Loops); i++ {
			if result.Loops[i] > 0 {
				loopExecuted++
			}
			loopTotal++
		}

		coverageEntry.LoopsTotal = loopTotal
		coverageEntry.LoopsCovered = loopExecuted
		coverageEntry.LoopsMissed = loopTotal - loopExecuted
		// We could get NaN if we have 0 branches
		// This is handled in the formatting for now, because of the `float32` type.
		coverageEntry.LoopsCoverage = float32(loopExecuted) / float32(loopTotal) * 100

		t.coverage = append(t.coverage, coverageEntry)
	}

}

func (cr *CoverageReporter) ComputeSummary() {
	var globalResultMap = ResultMap{}
	for _, result := range cr.results {
		for templateName, resultEntry := range result {
			globalResultEntry, ok := globalResultMap[templateName]
			if !ok {
				globalResultMap[templateName] = resultEntry
			} else {
				for n, action := range resultEntry.Actions {
					if len(globalResultEntry.Actions) < n+1 {
						globalResultEntry.Actions = append(globalResultEntry.Actions, action)
					} else {
						globalResultEntry.Actions[n] = globalResultEntry.Actions[n] + action
					}
				}
				for n, branch := range resultEntry.Branches {
					if len(globalResultEntry.Branches) < n+1 {
						globalResultEntry.Branches = append(globalResultEntry.Branches, branch)
					} else {
						globalResultEntry.Branches[n] = globalResultEntry.Branches[n] + branch
					}
				}
				for n, loop := range resultEntry.Loops {
					if len(globalResultEntry.Loops) < n+1 {
						globalResultEntry.Loops = append(globalResultEntry.Loops, loop)
					} else {
						globalResultEntry.Loops[n] = globalResultEntry.Loops[n] + loop
					}
				}
				globalResultMap[templateName] = globalResultEntry
			}
		}

	}
	cr.ComputeCoverage(globalResultMap)
}

type YamlStruct struct {
	Source   []string `yaml:"#"`
	Actions  []int    `yaml:"__actions"`
	Loops    []int    `yaml:"__loops"`
	Branches []int    `yaml:"__branches"`
}

func (cr *CoverageReporter) Generate() {
	cr.ComputeSummary()
	cr.RenderTable()
}

func percentageText(perc float32) string {
	c := color.FgGreen
	if strings.Contains(fmt.Sprintf("%f", perc), "NaN") {
		return color.New(color.FgWhite).SprintFunc()("N/A")
	}

	formatted := fmt.Sprintf("%.2f%%", perc)
	if perc <= 33 {
		c = color.FgRed
	} else if perc <= 66 {
		c = color.FgYellow
	}

	colorer := color.New(c).SprintFunc()
	return colorer(formatted)
}

func (cr *CoverageReporter) RenderTable() {
	tbl := table.NewWriter()
	tbl.SetOutputMirror(os.Stdout)
	tbl.AppendHeader(table.Row{"File", "% Action", "% Branch", "% Loop", "Actions", "Actions Missed", "Branches", "Branches Missed", "Loops"})
	tbl.AppendSeparator()
	tbl.SortBy([]table.SortBy{
		{Name: "File", Mode: table.Asc},
	})

	var actionTotal int
	var actionCovered int
	var branchesTotal int
	var branchesCovered int
	var loopsTotal int
	var loopsCovered int
	for _, tr := range cr.coverage {
		actionTotal += tr.ActionTotal
		actionCovered += tr.ActionCovered
		branchesTotal += tr.BranchesTotal
		branchesCovered += tr.BranchesCovered
		loopsTotal += tr.LoopsTotal
		loopsCovered += tr.LoopsCovered
	}
	var actionCoverage = float32(actionCovered) / float32(actionTotal) * 100
	var branchesCoverage = float32(branchesCovered) / float32(branchesTotal) * 100
	var loopsCoverage = float32(loopsCovered) / float32(loopsTotal) * 100

	for _, tr := range cr.coverage {
		if tr.NotExecuted {
			tbl.AppendRow([]interface{}{
				tr.Path,
				"-",
				"-",
				"-",
				"-",
				"-",
				"-",
				"-"},
			)
		} else {

			tbl.AppendRow([]interface{}{
				tr.Path,
				percentageText(tr.ActionCoverage),
				percentageText(float32(tr.BranchesCoverage)),
				percentageText(float32(tr.LoopsCoverage)),
				fmt.Sprintf("%d/%d", tr.ActionCovered, tr.ActionTotal),
				strings.ReplaceAll(wrap.WrapString(tr.ActionMissed, 30), " ", ""),
				fmt.Sprintf("%d/%d", tr.BranchesCovered, tr.BranchesTotal),
				strings.ReplaceAll(wrap.WrapString(tr.BranchesMissed, 30), " ", ""),
				fmt.Sprintf("%d/%d", tr.LoopsCovered, tr.LoopsTotal)},
			)
		}

	}

	// All files
	tbl.AppendFooter([]interface{}{
		"ALL FILES",
		percentageText(actionCoverage),
		percentageText(float32(branchesCoverage)),
		percentageText(float32(loopsCoverage)),
		fmt.Sprintf("%d/%d", actionCovered, actionTotal),
		"-",
		fmt.Sprintf("%d/%d", branchesCovered, branchesTotal),
		"-",
		fmt.Sprintf("%d/%d", loopsCovered, loopsTotal)},
	)
	tbl.Render()
}
