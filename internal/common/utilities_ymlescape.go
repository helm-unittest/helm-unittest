package common

import (
	"bytes"
	"strings"
)

const (
	bsCode byte = byte('\\')
)

type YmlEscapeHandlers struct{}

// Escape function is required, as yaml library no longer maintained
// yaml unmaintained library issue https://github.com/go-yaml/yaml/pull/862
func (y *YmlEscapeHandlers) Escape(content string) []byte {
	if !strings.Contains(content, `\`) {
		return nil
	}
	return escapeBackslashes([]byte(content))
}

// escapeBackslashes escapes backslashes in the given byte slice.
// It ensures that an even number of backslashes are present by doubling any single backslash found.
func escapeBackslashes(content []byte) []byte {
	var result bytes.Buffer
	for i := 0; i < len(content); i++ {
		if content[i] != bsCode {
			result.WriteByte(content[i])
			continue
		}
		count := 1
		for i+count < len(content) && content[i+count] == bsCode {
			count++
		}

		times := count
		if count%2 == 1 {
			times = count + 1
		}

		if count > 1 {
			i += count - 1
		}

		for j := 0; j < times; j++ {
			result.WriteByte('\\')
		}

	}
	return result.Bytes()
}
