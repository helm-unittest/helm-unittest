package helmutils

import (
	"os"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// inspired by https://github.com/helm/helm/blob/663a896f4a815053445eec4153677ddc24a0a361/pkg/chart/loader/load.go#L38

// LoaderWithRules returns a new ChartLoader appropriate for the given chart name
func LoaderWithRules(name string, rules *Rules) (loader.ChartLoader, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return DirLoader{
			path:  name,
			rules: *rules,
		}, nil
	}
	return loader.FileLoader(name), nil
}

// Load takes a string name, tries to resolve it to a file or directory, and then loads it.
//
// This is the preferred way to load a chart. It will discover the chart encoding
// and hand off to the appropriate chart reader.
func Load(name string, rules *Rules) (*chart.Chart, error) {
	l, err := LoaderWithRules(name, rules)
	if err != nil {
		return nil, err
	}
	return l.Load()
}
