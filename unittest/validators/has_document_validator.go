package validators

import (
	"strconv"
)

// HasDocumentsValidator validate whether the count of manifests rendered form template is Count
type HasDocumentsValidator struct {
	Count int
}

func (a HasDocumentsValidator) failInfo(actual int, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT to be"
	}
	hasDocumentsFailFormat := "Expected documents count" + notAnnotation + ":%s"
	if not {
		return splitInfof(hasDocumentsFailFormat, strconv.Itoa(a.Count))
	}
	return splitInfof(
		hasDocumentsFailFormat+"\nActual:%s",
		strconv.Itoa(a.Count),
		strconv.Itoa(actual),
	)
}

// Validate implement Validatable
func (a HasDocumentsValidator) Validate(context *ValidateContext) (bool, []string) {
	if len(context.Docs) == a.Count != context.Negative {
		return true, []string{}
	}
	return false, a.failInfo(len(context.Docs), context.Negative)
}
