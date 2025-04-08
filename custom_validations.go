package val

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
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

// labelSelectorValidator registers a custom validation rule "k8s_label_selector"
// with the provided validator instance.
//
// Validation Rule:
//   - The value must be a non-empty string.
//   - The string must conform to Kubernetes label selector syntax as defined by
//     the k8s.io/apimachinery/pkg/labels package.
//   - Invalid selectors (e.g., malformed keys, illegal operators) will fail validation.
func labelSelectorValidator(v *validator.Validate) {
	_ = v.RegisterValidation("k8s_label_selector", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()

		// Reject empty selectors explicitly.
		if value == "" {
			return false
		}
		_, err := labels.Parse(value)
		return err == nil
	})
}

// fieldSelectorValidator registers a custom validation rule "k8s_field_selector"
// with the given validator instance.
//
// Validation Rule:
//   - The field must be a non-empty string.
//   - The string must conform to Kubernetes field selector syntax,
//     as parsed by k8s.io/apimachinery/pkg/fields.
func fieldSelectorValidator(v *validator.Validate) {
	_ = v.RegisterValidation("k8s_field_selector", func(fl validator.FieldLevel) bool {
		var allowedFieldKeys = map[string]struct{}{
			"metadata.name":      {},
			"metadata.namespace": {},
			"status.phase":       {},
			"spec.nodeName":      {},
			"spec.unschedulable": {},
			"status.hostIP":      {},
			"status.podIP":       {},
		}

		value := fl.Field().String()
		if value == "" {
			return false
		}

		// Parse for syntax only; fail early on malformed input
		if _, err := fields.ParseSelector(value); err != nil {
			return false
		}

		// Enforce only known indexable field keys
		requirements := strings.Split(value, ",")
		for _, r := range requirements {
			r = strings.TrimSpace(r)
			var key string

			switch {
			case strings.Contains(r, "!="):
				key = strings.SplitN(r, "!=", 2)[0]
			case strings.Contains(r, "="):
				key = strings.SplitN(r, "=", 2)[0]
			default:
				return false // invalid or unsupported syntax
			}

			key = strings.TrimSpace(key)
			if _, ok := allowedFieldKeys[key]; !ok {
				return false
			}
		}

		return true
	})
}
