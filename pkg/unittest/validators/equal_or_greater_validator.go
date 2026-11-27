package validators

// EqualOrGreaterValidator validate whether the value of Path is greater or equal to Value
type EqualOrGreaterValidator struct {
	Path         string
	Value        any
	ParseOptions `mapstructure:",squash"`
}

// Validate implement Validatable
func (egv EqualOrGreaterValidator) Validate(context *ValidateContext) (bool, []string) {
	operatorValidator := operatorValidator{
		Path:           egv.Path,
		Value:          egv.Value,
		ComparisonType: "greater",
		ParseOptions:   egv.ParseOptions,
	}

	return operatorValidator.Validate(context)
}
