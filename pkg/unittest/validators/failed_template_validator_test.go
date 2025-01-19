package validators_test

import (
	"errors"
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var failedTemplate = `
raw: A field should be required
`

func TestFailedTemplateValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(failedTemplate)
	tests := []struct {
		name      string
		validator FailedTemplateValidator
	}{
		{
			name:      "test case 1: with error message",
			validator: FailedTemplateValidator{ErrorMessage: "A field should not be required"},
		},
		{
			name:      "test case 2: with empty error message",
			validator: FailedTemplateValidator{},
		},
		{
			name:      "test case 3: with error message that match substring",
			validator: FailedTemplateValidator{ErrorPattern: "should not be"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, diff := tt.validator.Validate(&ValidateContext{
				Docs:     []common.K8sManifest{manifest},
				Negative: true,
			})

			assert.True(t, pass)
			assert.Equal(t, []string{}, diff)
		})
	}
}

func TestFailedTemplateValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(failedTemplate)
	tests := []struct {
		name      string
		validator FailedTemplateValidator
	}{
		{
			name:      "test case 1: with error message",
			validator: FailedTemplateValidator{ErrorMessage: "A field should not be required"},
		},
		{
			name:      "test case 2: with empty error message",
			validator: FailedTemplateValidator{},
		},
		{
			name:      "test case 3: with error message that match substring",
			validator: FailedTemplateValidator{ErrorPattern: "should not be"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, diff := tt.validator.Validate(&ValidateContext{
				Docs:     []common.K8sManifest{manifest},
				Negative: true,
			})

			assert.True(t, pass)
			assert.Equal(t, []string{}, diff)
		})
	}
}

func TestFailedTemplateValidatorWhenEmptyFail(t *testing.T) {
	tests := []struct {
		name      string
		validator FailedTemplateValidator
		expected  []string
	}{
		{
			name:      "test case 1: with error message",
			validator: FailedTemplateValidator{ErrorMessage: "A field should not be required"},
			expected:  []string{"Expected to equal:", "\tA field should not be required", "Actual:", "\tNo failed document"},
		},
		{
			name:      "test case 2: with empty error message",
			validator: FailedTemplateValidator{},
			expected:  []string{"Expected to equal:", "\t", "Actual:", "\tNo failed document"},
		},
		{
			name:      "test case 3: with error message that match substring",
			validator: FailedTemplateValidator{ErrorPattern: "should not be"},
			expected:  []string{"Expected to match:", "\tshould not be", "Actual:", "\tNo failed document"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, diff := tt.validator.Validate(&ValidateContext{
				Docs:     []common.K8sManifest{},
				Negative: false,
			})

			assert.False(t, pass)
			assert.Equal(t, tt.expected, diff)
		})
	}
}

func TestFailedTemplateValidatorWhenEmptyNegativeAndOk(t *testing.T) {
	tests := []struct {
		name      string
		validator FailedTemplateValidator
	}{
		{
			name:      "test case 1: with error message",
			validator: FailedTemplateValidator{ErrorMessage: "A field should not be required"},
		},
		{
			name:      "test case 2: empty error message",
			validator: FailedTemplateValidator{},
		},
		{
			name:      "test case 3: with error message that match substring",
			validator: FailedTemplateValidator{ErrorPattern: "should not be"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, diff := tt.validator.Validate(&ValidateContext{
				Docs:     []common.K8sManifest{},
				Negative: true,
			})

			assert.True(t, pass)
			assert.Equal(t, []string{}, diff)
		})
	}
}

func TestFailedTemplateValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(failedTemplate)

	log.SetLevel(log.DebugLevel)

	tests := []struct {
		name      string
		validator FailedTemplateValidator
		expected  []string
	}{
		{
			name:      "test case 1: incorrect error message",
			validator: FailedTemplateValidator{ErrorMessage: "A field should not be required"},
			expected: []string{
				"DocumentIndex:	0",
				"Expected to equal:",
				"	A field should not be required",
				"Actual:",
				"	A field should be required",
			},
		},
		{
			name:      "test case 2: empty error message",
			validator: FailedTemplateValidator{},
			expected:  []string{},
		},
		{
			name:      "test case 3: incorrect error message",
			validator: FailedTemplateValidator{ErrorPattern: "should not be required"},
			expected: []string{
				"DocumentIndex:	0",
				"Expected to match:",
				"	should not be required",
				"Actual:",
				"	A field should be required",
			},
		},
		{
			name:      "test case 4: with error message that give compile pattern error",
			validator: FailedTemplateValidator{ErrorPattern: "+"},
			expected:  []string{"DocumentIndex:\t0", "Expected to match:", "\t+", "Actual:", "\tA field should be required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, diff := tt.validator.Validate(&ValidateContext{
				Docs: []common.K8sManifest{manifest},
			})

			if len(tt.expected) > 0 {
				assert.False(t, pass)
			} else {
				assert.True(t, pass)
			}

			assert.Equal(t, tt.expected, diff)
		})
	}
}

func TestFailedTemplateValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(failedTemplate)
	v := FailedTemplateValidator{ErrorMessage: "A field should be required"}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Expected NOT to equal:",
		"	A field should be required",
	}, diff)
}

func TestFailedTemplateValidatorWhenRenderError(t *testing.T) {
	testRenderError := errors.New("values don't meet the specifications of the schema(s)")
	tests := []struct {
		name      string
		validator FailedTemplateValidator
	}{
		{
			name:      "Test case 1: with error pattern",
			validator: FailedTemplateValidator{ErrorPattern: "schema"},
		},
		{
			name:      "Test case 2: with error message",
			validator: FailedTemplateValidator{ErrorMessage: "values don't meet the specifications of the schema(s)"},
		},
		{
			name:      "Test case 2: empty error message",
			validator: FailedTemplateValidator{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			pass, diff := tt.validator.Validate(&ValidateContext{
				Docs:        []common.K8sManifest{},
				RenderError: testRenderError,
			})

			assert.True(t, pass)
			assert.Equal(t, []string{}, diff)
		})
	}
}

func TestFailedTemplateValidatorWhenErrorMessageAndErrorPatternSet(t *testing.T) {
	manifest := makeManifest(failedTemplate)
	v := FailedTemplateValidator{ErrorMessage: "A field should be required", ErrorPattern: "pattern is set"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Error:",
		"	single attribute 'errorMessage' or 'errorPattern' supported at the same time",
	}, diff)
}

func TestFailedTemplateValidatorShowsAllErrors(t *testing.T) {
	manifest := makeManifest(failedTemplate)
	v := FailedTemplateValidator{ErrorMessage: "A field should be required"}

	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest, manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:\t0",
		"Expected NOT to equal:",
		"\tA field should be required",
		"DocumentIndex:\t1",
		"Expected NOT to equal:",
		"\tA field should be required",
	}, diff)
}

func TestFailedTemplateValidatorFailFast(t *testing.T) {
	manifest := makeManifest(failedTemplate)
	v := FailedTemplateValidator{ErrorMessage: "A field should be required"}

	pass, diff := v.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Expected NOT to equal:",
		"	A field should be required",
	}, diff)
}

func TestFailedTemplateValidator_ErrorPattern_SpecialCharactersAndEscapes_OK(t *testing.T) {
	var template = "raw: |-\n    " + "`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` (root)"
	manifest := makeManifest(template)

	cases := []struct {
		name    string
		pattern string
	}{
		{
			name:    "pattern with backticks and escapes",
			pattern: "`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` \\(root\\)",
		},
		{
			name:    "pattern with backticks",
			pattern: "`runAsNonRoot`",
		},
		{
			name:    "pattern with escape",
			pattern: "\\(root\\)",
		},
		{
			name:    "pattern contains metacharacters without escape",
			pattern: "(root)",
		},
		{
			name:    "pattern without escape",
			pattern: "(root)",
		},
		{
			name:    "pattern with meta characters and no explicit escape handling",
			pattern: "`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` (root)",
		},
		{
			name:    "empty range",
			pattern: "[\\w--]",
		},
		{
			name:    "shorthand escape sequences",
			pattern: "[[:digit:]-[:upper:]]",
		},
		{
			name:    "mis-ordered character range",
			pattern: "[\\w-\\d]",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			v := FailedTemplateValidator{ErrorPattern: tt.pattern}
			pass, _ := v.Validate(&ValidateContext{
				Docs: []common.K8sManifest{manifest},
			})
			assert.True(t, pass)
		})
	}
}

func TestFailedTemplateValidator_ErrorPattern_SpecialCharactersAndEscapes_Diff(t *testing.T) {
	var template = "raw: |-\n    " + "`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` (root)"
	manifest := makeManifest(template)

	cases := []struct {
		name    string
		pattern string
		diff    interface{}
	}{
		{
			name:    "pattern with incorrect regex escape",
			pattern: `\(root)`,
			diff: []string{
				"DocumentIndex:\t0",
				"Expected to match:",
				"\t\\(root)", "Actual:",
				"\t`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` (root)",
			},
		},
		{
			name:    "pattern with incorrect regex escape",
			pattern: `\\`,
			diff: []string{
				"DocumentIndex:\t0",
				"Expected to match:",
				"\t\\\\", "Actual:",
				"\t`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` (root)",
			},
		},
		{
			name:    "pattern with incorrect regex escape",
			pattern: "\x8A",
			diff:    []string{"Error:", "\terror parsing regexp: invalid UTF-8: `\x8a`"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			v := FailedTemplateValidator{ErrorPattern: tt.pattern}
			_, diff := v.Validate(&ValidateContext{
				Docs: []common.K8sManifest{manifest},
			})
			assert.Equal(t, diff, tt.diff)
		})
	}
}

func TestFailedTemplateValidator_HandleMetaCharacters_InvalidRegex1(t *testing.T) {
	var template = "raw: |-\n    " + "`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` (root)"
	manifest := makeManifest(template)
	regexPatterns := []string{
		"[a-z",                     // missing closing bracket
		"\\",                       // trailing backslash
		"(?P<name>",                // incomplete named group
		"(?",                       // incomplete group
		"(*",                       // invalid group
		"(?<",                      // invalid named group
		"(?<name>",                 // incomplete named group
		"(?P<name>[a-z",            // incomplete named group with character class
		"(?P<name>[",               // incomplete named group with character class
		"(?P<name>",                // incomplete named group
		"(?P<name",                 // incomplete named group
		"(?P<",                     // incomplete named group
		"(?P",                      // incomplete named group
		"[",                        // Unmatched opening bracket
		"]",                        // Unmatched closing bracket
		"{",                        // Unmatched opening curly brace
		"}",                        // Unmatched closing curly brace
		"\\",                       // Trailing backslash
		"(?",                       // Invalid group syntax
		"(*)",                      // Invalid usage of '*'
		"[a-z",                     // Unterminated character class
		"[a-z-",                    // Invalid character range
		"\\x",                      // Incomplete escape sequence
		"\\xG",                     // Invalid hex escape sequence
		"\\u",                      // Incomplete Unicode escape sequence
		"\\uZZZZ",                  // Invalid Unicode escape sequence
		"a{2,1}",                   // Invalid quantifier range
		"[[:invalid:]]",            // Invalid POSIX character class
		"(?<name>abc(?<name>d))",   // Duplicate named group
		"a(?P<name>bc(?P<name>d))", // Duplicate named group syntax
		"\\p{InvalidProperty}",     // Invalid Unicode property
		"**",                       // Invalid usage of repeated quantifiers
		"abc\\q",                   // Invalid escape sequence
		"a(?<![a-z])",              // Invalid negative lookbehind (not supported in Go)
		"a(?<=a)b(?<!c)",           // Multiple invalid lookbehind assertions
		"[[.symbol.]]",             // Unsupported POSIX equivalence class
		"(?i)a[b-Z]",               // Invalid range inside a case-insensitive group
		"[[:alpha]abc]",            // POSIX class mixed with characters
		"(a|",                      // Unterminated alternation
		"(?P<name1>a)(?P<name1>b)", // Duplicate named group (alternate syntax)
		"(?x",                      // Unterminated extended mode
		"(?#)",                     // Invalid comment syntax
		"\\k<name>",                // Invalid backreference
		"\\c?",                     // Invalid control character escape
		"\\cZ",                     // Invalid control character
		"\\p{UnknownProperty}",     // Unsupported Unicode property
		"\\p",                      // Incomplete property syntax
		"a\\y",                     // Unsupported escape sequence
		"(?P<name>.+\\k<name>)",    // Invalid backreference in the same group
		"\\u{FFFFFFF}",             // Exceeds valid Unicode range
		"(?!(?i:a))",               // Invalid nested group reference
		"[--]",                     // Empty character class range
		"a[a-\\d]",                 // Invalid range mix
		"[\\z]",                    // Unsupported escape
		"[\\*]",                    // Escaped special character not required
		"[\\x1Z]",                  // Invalid hex escape in a range
		"(abc",                     // Unbalanced parentheses
		"(a*",                      // Incomplete quantifier
		"(.*?)(*.)",                // Invalid wildcard group
		"(a)(?(1)B)",               // Conditional groups unsupported
		"([a-Z])",                  // Invalid range
		"\\w++",                    // Invalid possessive quantifier
		"[abc\\Qd\\E]",             // Unsupported literal escape
		"(?z:)",                    // Unsupported group flag
		"a{,2}",                    // Missing lower quantifier bound
		"{",                        // Lone opening brace
		"}a",                       // Unexpected closing brace
		"(a)(b(?>c)",               // Unterminated atomic group
		"a{",                       // Unterminated quantifier
		"a{2,}",                    // Missing upper quantifier
		"a(?:bc",                   // Unterminated non-capturing group
		"(?<a>",                    // Unterminated named group
		"a[b-c",                    // Unterminated range
		"[[:alpha",                 // Unterminated POSIX class
		"[abc\\",                   // Unterminated escape in range
		"(?'name'a)",               // Incorrect group naming
		"(?(1)a|)",                 // Empty conditional branch
		"(?:",                      // Unclosed non-capturing group
		"\\m",                      // Invalid escape character
		"(?()abc)",                 // Invalid conditional structure
		"(?a)b(?-i)c",              // Invalid inline flag nesting
		"(?(?=a))",                 // Conditional missing branch
		"[\\ud800]",                // Invalid surrogate
		"(?>a)*",                   // Invalid repeat on atomic
		"(?>a{2,}?)",               // Atomic group with lazy quantifier
		"(?=",                      // Unclosed positive lookahead
		"[\\]",                     // Escape in empty range
		"[[:alnum",                 // Unterminated POSIX class
		"\\P{Unknown}",             // Invalid Unicode negated property
		"[[:xdigit]]",              // Missing ':'
		"(?(1)a|bc",                // Unterminated conditional group
		"\\g<1>",                   // Unsupported backreference syntax
		"((a)",                     // Mismatched parentheses
		"a[[:digit:]-]",            // Invalid POSIX range ending
		"(a)(b(c)",                 // Missing close parenthesis
		"[a--b]",                   // Invalid range structure
		"\\U0001F600",              // Invalid Unicode property format
		"[[:badclass:]]",           // Invalid POSIX class name
		"(a)(*b)",                  // Invalid group usage
		"(a)b)c",                   // Extra closing parenthesis
		"\\k'name'",                // Invalid group reference
		"[[:punct][-]]",            // Invalid range mix
	}

	for _, invalidRegex := range regexPatterns {
		v := FailedTemplateValidator{ErrorPattern: invalidRegex}
		pass, _ := v.Validate(&ValidateContext{
			Docs: []common.K8sManifest{manifest},
		})
		assert.False(t, pass, "Regex pattern: %s", invalidRegex)
	}
}

func TestFailedTemplateValidator_HandleMetaCharacters_InvalidUtfSequence(t *testing.T) {
	var template = "raw: |-\n    " + "`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` (root)"
	manifest := makeManifest(template)
	invalidUTF8Sequences := []string{
		"\xC0\x80",                 // Overlong encoding (U+0000)
		"\xE0\x80\x80",             // Overlong encoding (U+0000)
		"\xF0\x80\x80\x80",         // Overlong encoding (U+0000)
		"\xED\xA0\x80",             // Leading surrogate (U+D800)
		"\xED\xBF\xBF",             // Trailing surrogate (U+DFFF)
		"\x80",                     // Standalone continuation byte
		"\xBF",                     // Standalone continuation byte
		"\xC2",                     // Incomplete 2-byte sequence
		"\xE2\x82",                 // Incomplete 3-byte sequence
		"\xF0\xA4",                 // Incomplete 4-byte sequence
		"\xF5\x80\x80\x80",         // Out of range (beyond U+10FFFF)
		"\xFE",                     // Invalid UTF-8 byte
		"\xFF",                     // Invalid UTF-8 byte
		"\xF8\x80\x80\x80\x80",     // 5-byte overlong encoding
		"\xFC\x80\x80\x80\x80\x80", // 6-byte overlong encoding
		"\xC1\xBF",                 // Overlong encoding (U+007F)
		"\xE0\x9F\xBF",             // Overlong encoding (U+07FF)
		"\xF0\x8F\xBF\xBF",         // Overlong encoding (U+FFFF)
		"\xF4\x90\x80\x80",         // Out of Unicode range
		"\xE2\x28\xA1",             // Invalid UTF-8 (not continuation byte)
	}

	for _, invalidRegex := range invalidUTF8Sequences {
		v := FailedTemplateValidator{ErrorPattern: string(invalidRegex)}
		pass, _ := v.Validate(&ValidateContext{
			Docs: []common.K8sManifest{manifest},
		})
		assert.False(t, pass)
	}
}
