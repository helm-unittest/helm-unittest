package validators

import (
	"strconv"
)

// HasDocumentsValidator validate whether the count of manifests rendered form template is Count
type HasDocumentsValidator struct {
	Count int
}

func (v HasDocumentsValidator) failInfo(actual int, not bool) []string {
	customMessage := " documents count to be"
	if not {
		return splitInfof(
			setFailFormat(not, false, false, false, customMessage),
			-1,
			strconv.Itoa(v.Count),
		)
	}
	return splitInfof(
		setFailFormat(not, false, true, false, customMessage),
		-1,
		strconv.Itoa(v.Count),
		strconv.Itoa(actual),
	)
}

// Validate implement Validatable
func (v HasDocumentsValidator) Validate(context *ValidateContext) (bool, []string) {
	if len(context.Docs) == v.Count != context.Negative {
		return true, []string{}
	}
	return false, v.failInfo(len(context.Docs), context.Negative)
}
