package unittest

import (
	"fmt"
)

// ConvertIToString The convertToString function takes an interface{} value as input and returns a string representation of it.
// If the input value is nil, it returns an empty string.
func ConvertIToString(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	default:
		return ""
	}
}
