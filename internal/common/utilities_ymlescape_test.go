package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscape_InputAsRunes_EscapeBackslash_Code92(t *testing.T) {
	y := &YmlEscapeHandlers{}
	cases := []struct {
		name     string
		content  []byte
		expected []byte
	}{
		{
			name: "single and double backslashes",
			// paradox \(root\\)
			content:  []byte{10, 9, 112, 97, 114, 97, 100, 111, 120, 32, bsCode, 40, 114, 111, 111, 116, bsCode, bsCode, 41, 10, 9},
			expected: []byte{10, 9, 112, 97, 114, 97, 100, 111, 120, 32, bsCode, bsCode, 40, 114, 111, 111, 116, bsCode, bsCode, 41, 10, 9},
		},
		{
			name: "single and triple backslashes with special characters",
			// `runAsUser` is set to `0` \(root\\\
			content: []byte{96, 114, 117, 110, 65, 115, 85, 115, 101, 114, 96, 32, 105, 115, 32, 115, 101, 116, 32, 116, 111, 32, 96, 48, 96, 32, bsCode, 40, 114, 111, 111, 116, bsCode, bsCode, bsCode},
			// `runAsUser` is set to `0` \\(root\\\\
			expected: []byte{96, 114, 117, 110, 65, 115, 85, 115, 101, 114, 96, 32, 105, 115, 32, 115, 101, 116, 32, 116, 111, 32, 96, 48, 96, 32, bsCode, bsCode, 40, 114, 111, 111, 116, bsCode, bsCode, bsCode, bsCode},
		},
		{
			name: "single and triple backslashes with special characters",
			// `runAsUser` is set to `0` \(root\\)
			content: []byte{96, 114, 117, 110, 65, 115, 85, 115, 101, 114, 96, 32, 105, 115, 32, 115, 101, 116, 32, 116, 111, 32, 96, 48, 96, 32, bsCode, 40, 114, 111, 111, 116, bsCode, bsCode, bsCode},
			// `runAsUser` is set to `0` \\(root\\\\
			expected: []byte{96, 114, 117, 110, 65, 115, 85, 115, 101, 114, 96, 32, 105, 115, 32, 115, 101, 116, 32, 116, 111, 32, 96, 48, 96, 32, bsCode, bsCode, 40, 114, 111, 111, 116, bsCode, bsCode, bsCode, bsCode},
		},
		{
			name: "special characters and backslashes",
			// `run` is set \\0\\\ \(root\\\)\\\\
			content: []byte{96, 114, 117, 110, 96, 32, 105, 115, 32, 115, 101, 116, 32, bsCode, 48, bsCode, bsCode, bsCode, 32, bsCode, 40, 114, 111, 111, 116, bsCode, bsCode, bsCode, 41, bsCode, bsCode, bsCode, bsCode},
			// `run` is set \\0\\\\ \\(root\\\\)\\\\
			expected: []byte{96, 114, 117, 110, 96, 32, 105, 115, 32, 115, 101, 116, 32, bsCode, bsCode, 48, bsCode, bsCode, bsCode, bsCode, 32, bsCode, bsCode, 40, 114, 111, 111, 116, bsCode, bsCode, bsCode, bsCode, 41, bsCode, bsCode, bsCode, bsCode},
		},
		{
			name:     "empty",
			content:  []byte{},
			expected: nil,
		},
		{
			name:     "Mixed Backslashes",
			content:  []byte("hello\\world\\\\"),
			expected: []byte("hello\\\\world\\\\"),
		},
		{
			name:     "no backslashes",
			content:  []byte("abrakadabra"),
			expected: nil,
		},
		{
			name:     "contains meta characters but no slashes",
			content:  []byte("some text with [value] needs escaping"),
			expected: []byte("some text with [value] needs escaping"),
		},
		{
			name:     "backslashes at the end",
			content:  []byte("hello world\\"),
			expected: []byte("hello world\\\\"),
		},
		{
			name:     "backslashes in the middle",
			content:  []byte("hello\\world\\"),
			expected: []byte("hello\\\\world\\\\"),
		},
		{
			name:     "even number of backslashes",
			content:  []byte("hello" + `\\\` + "world"),
			expected: []byte("hello" + `\\\\` + "world"),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual := y.Escape(string(tt.content))

			assert.Equal(t, string(tt.expected), string(actual))
			assert.Equal(t, tt.expected, actual)
		})
	}
}
