package helmtest

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type TestSuite struct {
	Name  string `yaml:"suite"`
	Files []string
	Tests []TestJob
}

func ParseTestSuiteFile(path string) (TestSuite, error) {
	var suite TestSuite
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return TestSuite{}, err
	}

	if err := yaml.Unmarshal(content, &suite); err != nil {
		return TestSuite{}, err
	}

	return suite, nil
}

func (s TestSuite) Run() {

}
