package coverage

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WriteCobertura emits a Cobertura-compatible XML report at path. The format
// follows the Cobertura 04 DTD that Codecov, GitLab, Jenkins, SonarQube and
// Azure DevOps all consume. Action probes are mapped to Cobertura "lines" and
// every branch/loop probe contributes a Cobertura "condition" on its line.
func WriteCobertura(path string, cov Coverage) error {
	doc := buildCobertura(cov)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if _, err := io.WriteString(f, xml.Header); err != nil {
		return err
	}
	if _, err := io.WriteString(f, coberturaDoctype); err != nil {
		return err
	}
	enc := xml.NewEncoder(f)
	enc.Indent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return err
	}
	_, err = io.WriteString(f, "\n")
	return err
}

const coberturaDoctype = `<!DOCTYPE coverage SYSTEM "http://cobertura.sourceforge.net/xml/coverage-04.dtd">` + "\n"

type coberturaCoverage struct {
	XMLName         xml.Name           `xml:"coverage"`
	LineRate        string             `xml:"line-rate,attr"`
	BranchRate      string             `xml:"branch-rate,attr"`
	LinesCovered    int                `xml:"lines-covered,attr"`
	LinesValid      int                `xml:"lines-valid,attr"`
	BranchesCovered int                `xml:"branches-covered,attr"`
	BranchesValid   int                `xml:"branches-valid,attr"`
	Complexity      string             `xml:"complexity,attr"`
	Version         string             `xml:"version,attr"`
	Timestamp       int64              `xml:"timestamp,attr"`
	Sources         coberturaSources   `xml:"sources"`
	Packages        coberturaPackages  `xml:"packages"`
}

type coberturaSources struct {
	Sources []string `xml:"source"`
}

type coberturaPackages struct {
	Packages []coberturaPackage `xml:"package"`
}

type coberturaPackage struct {
	Name       string            `xml:"name,attr"`
	LineRate   string            `xml:"line-rate,attr"`
	BranchRate string            `xml:"branch-rate,attr"`
	Complexity string            `xml:"complexity,attr"`
	Classes    coberturaClasses  `xml:"classes"`
}

type coberturaClasses struct {
	Classes []coberturaClass `xml:"class"`
}

type coberturaClass struct {
	Name       string            `xml:"name,attr"`
	Filename   string            `xml:"filename,attr"`
	LineRate   string            `xml:"line-rate,attr"`
	BranchRate string            `xml:"branch-rate,attr"`
	Complexity string            `xml:"complexity,attr"`
	// HelmUnittestRendered is a helm-unittest extension attribute. Standard
	// Cobertura consumers ignore unknown attributes, so this is safe; tooling
	// that knows to look for it can use it to surface unused templates.
	HelmUnittestRendered string             `xml:"helm-unittest-rendered,attr"`
	Methods              coberturaMethods   `xml:"methods"`
	Lines                coberturaLineList  `xml:"lines"`
}

type coberturaMethods struct{}

type coberturaLineList struct {
	Lines []coberturaLine `xml:"line"`
}

type coberturaLine struct {
	Number            int    `xml:"number,attr"`
	Hits              int    `xml:"hits,attr"`
	Branch            string `xml:"branch,attr"`
	ConditionCoverage string `xml:"condition-coverage,attr,omitempty"`
}

func buildCobertura(cov Coverage) coberturaCoverage {
	doc := coberturaCoverage{
		Version:   "helm-unittest",
		Timestamp: time.Now().Unix() * 1000,
		Sources:   coberturaSources{Sources: []string{"."}},
		Complexity: "0",
	}

	// Cobertura's top-level rates are computed across all files. We use:
	//   line-rate   = action coverage (each action probe is a "line of code")
	//   branch-rate = combined branch + loop coverage
	totalLines := cov.Totals.Actions
	totalBranches := CountStat{
		Covered: cov.Totals.Branches.Covered + cov.Totals.Loops.Covered,
		Total:   cov.Totals.Branches.Total + cov.Totals.Loops.Total,
	}
	doc.LinesCovered = totalLines.Covered
	doc.LinesValid = totalLines.Total
	doc.BranchesCovered = totalBranches.Covered
	doc.BranchesValid = totalBranches.Total
	doc.LineRate = formatRate(totalLines)
	doc.BranchRate = formatRate(totalBranches)

	pkg := coberturaPackage{
		Name:       cov.ChartName,
		LineRate:   formatRate(totalLines),
		BranchRate: formatRate(totalBranches),
		Complexity: "0",
	}

	for _, f := range cov.Files {
		if f.ParseError != nil {
			continue
		}
		class := coberturaClass{
			Name:                 classNameFromPath(f.Name),
			Filename:             f.Name,
			LineRate:             formatRate(f.Actions),
			BranchRate:           formatRate(CountStat{Covered: f.Branches.Covered + f.Loops.Covered, Total: f.Branches.Total + f.Loops.Total}),
			Complexity:           "0",
			HelmUnittestRendered: fmt.Sprintf("%t", f.Rendered),
		}
		for _, ln := range f.Lines {
			cl := coberturaLine{
				Number: ln.Line,
				Hits:   ln.Hits,
				Branch: "false",
			}
			if n := len(ln.Branches); n > 0 {
				cl.Branch = "true"
				taken := 0
				for _, b := range ln.Branches {
					if b.Hits > 0 {
						taken++
					}
				}
				cl.ConditionCoverage = fmt.Sprintf("%d%% (%d/%d)", percentInt(taken, n), taken, n)
			}
			class.Lines.Lines = append(class.Lines.Lines, cl)
		}
		pkg.Classes.Classes = append(pkg.Classes.Classes, class)
	}
	doc.Packages.Packages = append(doc.Packages.Packages, pkg)
	return doc
}

// classNameFromPath converts a slash-separated template path into a
// dot-separated Cobertura class name. e.g.
//
//	chart-subchart-v1/templates/configmap.yaml
//	-> chart-subchart-v1.templates.configmap_yaml
func classNameFromPath(p string) string {
	dir := filepath.ToSlash(filepath.Dir(p))
	base := filepath.Base(p)
	base = strings.ReplaceAll(base, ".", "_")
	if dir == "." || dir == "" {
		return base
	}
	return strings.ReplaceAll(dir, "/", ".") + "." + base
}

func formatRate(s CountStat) string {
	if s.Total == 0 {
		return "1"
	}
	return fmt.Sprintf("%.4f", float64(s.Covered)/float64(s.Total))
}

func percentInt(covered, total int) int {
	if total == 0 {
		return 0
	}
	return int(100 * float64(covered) / float64(total))
}
