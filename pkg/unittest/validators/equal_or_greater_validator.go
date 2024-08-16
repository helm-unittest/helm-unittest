package validators

// EqualOrGreaterValidator validate whether the value of Path is greater or equal to Value
type EqualOrGreaterValidator struct {
	Path  string
	Value interface{}
}

// Validate implement Validatable
func (g EqualOrGreaterValidator) Validate(context *ValidateContext) (bool, []string) {
	operatorValidator := operatorValidator{
		Path:           g.Path,
		Value:          g.Value,
		ComparisonType: "greater",
	}

	return operatorValidator.Validate(context)
}
