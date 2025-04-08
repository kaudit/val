package val

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

// urlPrefixValidator registers custom validation rules with the validator instance.
//
// Custom Validators:
//   - "url_prefix": Validates that a string starts with "http://" or "https://".
//
// This function is intended to be called during validator initialization to
// ensure the custom rules are consistently available across the application.
func urlPrefixValidator(v *validator.Validate) {
	_ = v.RegisterValidation(
		"url_prefix",
		func(fl validator.FieldLevel) bool {
			value := fl.Field().String()
			return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
		},
	)
}
