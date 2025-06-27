package helmutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFail(t *testing.T) {
	rule := RulesWithDefaults()
	rules := []string{`templates/.?*`, `tests/*.yaml`, `!tests/__snapshot__/`}
	for _, r := range rules {
		err := rule.parseRule(r)
		assert.NoError(t, err)
	}
	assert.NotNil(t, rule)
	assert.Len(t, rule.getPatterns(), 3)
}
