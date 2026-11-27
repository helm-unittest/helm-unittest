package validators

// EqualOrLessValidator validate whether the value of Path is less or equal to Value
type EqualOrLessValidator struct {
	Path         string
	Value        any
	ParseOptions `mapstructure:",squash"`
}

// Validate implement Validatable
func (elv EqualOrLessValidator) Validate(context *ValidateContext) (bool, []string) {
	operatorValidator := operatorValidator{
		Path:           elv.Path,
		Value:          elv.Value,
		ComparisonType: "less",
		ParseOptions:   elv.ParseOptions,
	}

	return operatorValidator.Validate(context)
}
