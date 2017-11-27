package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseTestSuiteFileOk(t *testing.T) {
	a := assert.New(t)
	s, err := ParseTestSuiteFile("./__fixtures__/tests/basic.yaml")

	a.Nil(err)
	a.Equal(s.Name, "test suite name")
	a.Equal(s.Files, []string{"a.yaml", "b.yaml"})
	a.Equal(s.Tests, []TestJob{
		TestJob{
			Name:   "should ...",
			Values: []string{"values.yaml"},
			Set: map[string]interface{}{
				"a.b.c": "ABC",
				"x.y.z": "XYZ",
			},
			Assertions: []Assertion{
				Assertion{DocumentIndex: 1, Not: true},
				Assertion{DocumentIndex: 0, Not: false},
				Assertion{DocumentIndex: 0, Not: false},
				Assertion{DocumentIndex: 0, Not: false},
				Assertion{DocumentIndex: 0, Not: false},
				Assertion{DocumentIndex: 0, Not: false},
				Assertion{DocumentIndex: 0, Not: false},
				Assertion{DocumentIndex: 0, Not: false},
				Assertion{DocumentIndex: 0, Not: false},
				Assertion{DocumentIndex: 0, Not: false},
			},
		},
	})
}
