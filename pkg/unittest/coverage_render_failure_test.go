package unittest

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression: a coverage-render failure must be logged at Warn or above,
// since Debug-only logging hides it unless --debugPlugin is set.
func TestV4RunnerCoverageRenderFailureLogsAtWarn(t *testing.T) {
	chartDir := t.TempDir()
	writeFile := func(rel, content string) {
		path := filepath.Join(chartDir, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
	writeFile("Chart.yaml", "apiVersion: v2\nname: coverage-render-failure\nversion: 0.1.0\n")
	writeFile("values.yaml", "{}\n")
	writeFile("templates/cm.yaml", "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: {{ required \"boom\" nil }}\n")
	writeFile("tests/cm_test.yaml", "suite: cm\ntemplates:\n  - templates/cm.yaml\ntests:\n  - it: fails\n    asserts:\n      - failedTemplate: {}\n")

	logBuf := new(bytes.Buffer)
	origOut, origLevel := log.StandardLogger().Out, log.GetLevel()
	log.SetOutput(logBuf)
	log.SetLevel(log.WarnLevel)
	defer func() {
		log.SetOutput(origOut)
		log.SetLevel(origLevel)
	}()

	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:      printer.NewPrinter(buffer, nil),
		Coverage:     true,
		WithSubChart: true,
		TestFiles:    []string{"tests/*_test.yaml"},
	}
	runner.RunV4([]string{chartDir})

	assert.Contains(t, logBuf.String(), "coverage render failed",
		"a coverage-render failure must be logged at Warn or above, not just Debug")
}
