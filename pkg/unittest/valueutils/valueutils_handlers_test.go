package valueutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockTraverser struct {
	err error
	buildTraverser
	traverseMapKeyCount  int
	traverseListIdxCount int
}

func newMockTraverser() *mockTraverser {
	return &mockTraverser{
		buildTraverser: buildTraverser{},
	}
}

func (m *mockTraverser) traverseMapKey(key string) error {
	m.traverseMapKeyCount++
	err := m.buildTraverser.traverseMapKey(key)
	if err != nil {
		return err
	}
	if key == "key-error" {
		m.err = fmt.Errorf("traverser error")
	}
	return m.err
}

func (m *mockTraverser) traverseListIdx(idx int) error {
	m.traverseListIdxCount++
	err := m.buildTraverser.traverseListIdx(idx)
	if err != nil {
		return err
	}
	if idx == 6897 {
		m.err = fmt.Errorf("traverser error")
	}
	return m.err
}

func TestHandleExpectIndex_ErrorCases(t *testing.T) {
	tr := newMockTraverser()
	tests := []struct {
		name                 string
		runes                []rune
		lastRune             rune
		errMsg               string
		traverseListIdxCount int
	}{
		{
			name:                 "missing index value",
			errMsg:               "missing index value",
			runes:                []rune(""),
			lastRune:             'a',
			traverseListIdxCount: 0,
		},
		{
			name:                 "invalid index value",
			errMsg:               "invalid syntax",
			runes:                []rune("abc"),
			lastRune:             ']',
			traverseListIdxCount: 0,
		},
		{
			name:                 "invalid traverser",
			errMsg:               "traverser error",
			runes:                []rune("6897"),
			lastRune:             ']',
			traverseListIdxCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := handleExpectIndex(tt.runes, tt.lastRune, tr)
			assert.Error(t, err)
			assert.Equal(t, tt.traverseListIdxCount, tr.traverseListIdxCount)
			assert.Equal(t, -1, state)
		})
	}
}

func TestHandleExpectIndex_OkCases(t *testing.T) {
	tr := &mockTraverser{}
	tests := []struct {
		name     string
		runes    []rune
		lastRune rune
	}{
		{
			name:     "valid index value",
			runes:    []rune("123"),
			lastRune: ']',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := handleExpectIndex(tt.runes, tt.lastRune, tr)
			assert.NoError(t, err)
			assert.Equal(t, expectDenotation, state)
		})
	}
}

func TestHandleExpectDenotation_ValidDotToken(t *testing.T) {
	tests := []struct {
		name     string
		lastRune rune
		expected int
	}{
		{
			name:     "valid token",
			lastRune: '.',
			expected: expectKey,
		},
		{
			name:     "valid open bracket token",
			lastRune: '[',
			expected: expectIndex,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := handleExpectDenotation(tt.lastRune)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, state)
		})
	}
}

func TestHandleExpectDenotation_InvalidToken(t *testing.T) {
	state, err := handleExpectDenotation('a')
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid denotation token a")
	assert.Equal(t, -1, state)
}

func TestHandleExpectKey_OkCases(t *testing.T) {
	tests := []struct {
		name     string
		runes    []rune
		lastRune rune
		expected int
	}{
		{
			name:     "valid dot token",
			lastRune: '.',
			runes:    []rune("key"),
			expected: expectKey,
		},
		{
			name:     "valid open bracket token with empty key",
			lastRune: '[',
			runes:    []rune(""),
			expected: expectEscaping,
		},
		{
			name:     "valid open bracket token with non-empty key",
			lastRune: '[',
			runes:    []rune("key"),
			expected: expectIndex,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := newMockTraverser()
			state, err := handleExpectKey(tt.runes, tt.lastRune, tr)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, state)
		})
	}
}

func TestHandleExpectKey_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		runes       []rune
		lastRune    rune
		errorString string
	}{
		{
			name:        "invalid token",
			lastRune:    'a',
			runes:       []rune("key"),
			errorString: "invalid key a",
		},
		{
			name:        "traverser error on dot token",
			lastRune:    '.',
			runes:       []rune("key-error"),
			errorString: "traverser error",
		},
		{
			name:        "error on error on open bracket token",
			lastRune:    '[',
			runes:       []rune("key-error"),
			errorString: "traverser error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := newMockTraverser()
			state, err := handleExpectKey(tt.runes, tt.lastRune, tr)
			assert.Error(t, err)
			assert.Equal(t, tt.errorString, err.Error())
			assert.Equal(t, -1, state)
		})
	}
}

// last
func TestHandleExpectEscaping_OkCases(t *testing.T) {
	tests := []struct {
		name        string
		runes       []rune
		lastRune    rune
		expected    int
		bufferedKey string
	}{
		{
			name:        "valid dot token",
			lastRune:    '.',
			runes:       []rune("key"),
			expected:    expectEscaping,
			bufferedKey: "key.",
		},
		{
			name:        "valid close bracket token",
			lastRune:    ']',
			runes:       []rune("key"),
			expected:    expectDenotation,
			bufferedKey: "key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bufferedMapKey = ""
			tr := newMockTraverser()
			state, err := handleExpectEscaping(tt.runes, tt.lastRune, tr)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, state)
			assert.Equal(t, tt.bufferedKey, bufferedMapKey)
		})
	}
}

func TestHandleExpectEscaping_InvalidToken(t *testing.T) {
	tr := &mockTraverser{}
	bufferedMapKey = ""
	state, err := handleExpectEscaping([]rune("key"), 'a', tr)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid escaping token a")
	assert.Equal(t, -1, state)
}

func TestHandleExpectEscaping_TraverserError(t *testing.T) {
	tr := &mockTraverser{err: fmt.Errorf("traverser error")}
	bufferedMapKey = ""
	state, err := handleExpectEscaping([]rune("key"), ']', tr)
	assert.Error(t, err)
	assert.EqualError(t, err, "traverser error")
	assert.Equal(t, -1, state)
}
