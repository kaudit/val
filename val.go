// Package val provides a thread-safe validation mechanism using the go-playground/validator/v10 library as a singleton.
// Validator initialized internally and ready to use without any preparation steps.
package val

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	mtx sync.Mutex
	v   *validator.Validate
)

func init() {
	v = newValidator()
}

// RegisterValidation registers a custom validation function for a specific tag.
// Example usage:
//
//	err := validator.RegisterValidation("is-even", func(fl validator.FieldLevel) bool {
//	    return fl.Field().Int()%2 == 0
//	})
//
// This function is thread-safe.
func RegisterValidation(tag string, fn validator.Func) error {
	mtx.Lock()
	defer mtx.Unlock()
	return v.RegisterValidation(tag, fn)
}

// ValidateWithTag validates a single variable using a specified validation tag.
// Uses the go-playground validator to validate the `variable` against the provided `tag`.
// If validation fails, it processes and returns a structured error.
//
// Example Usage:
//
// err := ValidateWithTag("test@example.com", "email")
//
//	if err != nil {
//	    fmt.Println("Validation failed:", err)
//	}
//
// This function is thread-safe.
func ValidateWithTag(variable any, tag string) error {
	if err := v.Var(variable, tag); err != nil {
		return handleValidatorError(err)
	}
	return nil
}

// ValidateStruct validates a struct based on its validation tags.
// Ensures the input is a valid struct or a pointer to a struct.
// Validates the struct fields based on their tags.
// Returns detailed, formatted errors for each validation failure.
//
// Example:
//
//	type TestStruct struct {
//	    Field1 string `validate:"required"`
//	    Field2 int    `validate:"gte=0,lte=10"`
//	}
//
// obj := TestStruct{Field2: 15}
//
// err := ValidateStruct(obj)
//
//	if err != nil {
//	    fmt.Println(err) // Output: "validation failed: Field2 (gte=0, lte=10)"
//	}
//
// This function is thread-safe.
func ValidateStruct(s any) error {
	if err := validateInputStruct(s); err != nil {
		return err
	}

	if err := v.Struct(s); err != nil {
		return handleValidatorError(err)
	}
	return nil
}

// newValidator initializes and configures a new instance of the go-playground validator.
// This function is typically called during package initialization to set up the validator instance.
func newValidator() *validator.Validate {
	val := validator.New(validator.WithRequiredStructEnabled())

	urlPrefixValidator(val)
	labelSelectorValidator(val)
	fieldSelectorValidator(val)

	return val
}

// validateInputStruct ensures that the input is valid for struct-based validation.
//
// Validation Rules:
//   - The input mustn't be nil.
//   - If the input is a pointer, it mustn't be uninitialized (nil pointer).
func validateInputStruct(s any) error {
	if s == nil {
		return fmt.Errorf("input is nil")
	}

	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return fmt.Errorf("input is a nil pointer")
		}
	}

	return nil
}

// handleValidatorError processes and formats validation errors returned by the go-playground/validator.
// It extracts detailed, field-specific error messages for structured reporting.
//
// Behavior:
//   - If the error contains field-specific validation errors, they're formatted with
//     field names, tags, and parameters where applicable.
//   - If the error is not related to validation, it is returned as an unexpected error.
func handleValidatorError(err error) error {
	var valErr validator.ValidationErrors
	if errors.As(err, &valErr) {
		var detailedErrors []string
		for _, fe := range valErr {
			if fe.StructField() != "" {
				detailedErrors = append(
					detailedErrors,
					fmt.Sprintf("%s (%s=%s)", fe.StructNamespace(), fe.ActualTag(), fe.Param()),
				)
				continue
			}
			if fe.Value() == nil {
				detailedErrors = append(
					detailedErrors,
					fmt.Sprintf("nil value (%s=%s)", fe.ActualTag(), fe.Param()),
				)
				continue
			}
			detailedErrors = append(
				detailedErrors,
				fmt.Sprintf("%s %s (%s=%s)", fe.Type(), fe.Value(), fe.ActualTag(), fe.Param()),
			)
		}
		return fmt.Errorf("validation failed: %s", strings.Join(detailedErrors, ", "))
	}
	return fmt.Errorf("unexpected validation error: %w", err)
}
