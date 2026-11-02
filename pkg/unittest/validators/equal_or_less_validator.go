package validators

// EqualOrLessValidator validate whether the value of Path is less or equal to Value
type EqualOrLessValidator struct {
	Path         string
	Value        any
	ParseOptions `mapstructure:",squash"`
}

// Validate implement Validatable
func (l EqualOrLessValidator) Validate(context *ValidateContext) (bool, []string) {
	operatorValidator := operatorValidator{
		Path:           l.Path,
		Value:          l.Value,
		ComparisonType: "less",
		ParseOptions:   l.ParseOptions,
	}

	return operatorValidator.Validate(context)
}
