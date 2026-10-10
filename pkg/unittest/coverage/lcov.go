package coverage

import (
	"bufio"
	"fmt"
	"os"
)

// WriteLCOV emits an LCOV-format coverage report at path. Coveralls, Codecov
// and most IDE gutter integrations (VS Code "Coverage Gutters", JetBrains
// "Run with Coverage" import, etc.) understand this format directly.
//
// Each template file becomes one LCOV record:
//   - DA lines record per-line hit counts (sourced from action probes).
//   - BRDA lines record each branch / loop probe individually.
//   - LF/LH and BRF/BRH carry the summary totals.
//
// Templates that failed to parse are emitted as empty records so consumers
// still see them in the file list.
func WriteLCOV(path string, cov Coverage) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)
	defer func() { _ = w.Flush() }()

	for _, file := range cov.Files {
		if _, err := fmt.Fprintln(w, "TN:helm-unittest"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "SF:%s\n", file.Name); err != nil {
			return err
		}

		if file.ParseError != nil {
			// Empty record so the file still appears in the file list.
			if _, err := fmt.Fprintln(w, "LF:0\nLH:0\nBRF:0\nBRH:0\nend_of_record"); err != nil {
				return err
			}
			continue
		}

		linesFound := 0
		linesHit := 0
		branchesFound := 0
		branchesHit := 0
		// BRDA blocks are grouped by source line; "block" is the line number,
		// "branch" is the running index within that line.
		for _, ln := range file.Lines {
			if _, err := fmt.Fprintf(w, "DA:%d,%d\n", ln.Line, ln.Hits); err != nil {
				return err
			}
			linesFound++
			if ln.Hits > 0 {
				linesHit++
			}
			for bi, br := range ln.Branches {
				taken := "-"
				if br.Hits > 0 {
					taken = fmt.Sprintf("%d", br.Hits)
					branchesHit++
				}
				if _, err := fmt.Fprintf(w, "BRDA:%d,%d,%d,%s\n", ln.Line, ln.Line, bi, taken); err != nil {
					return err
				}
				branchesFound++
			}
		}

		if _, err := fmt.Fprintf(w, "LF:%d\nLH:%d\nBRF:%d\nBRH:%d\nend_of_record\n", linesFound, linesHit, branchesFound, branchesHit); err != nil {
			return err
		}
	}
	return nil
}
