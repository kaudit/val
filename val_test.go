package val

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	Field1 int64  `validate:"numeric,required,gt=1024,lt=65536"`
	Field2 string `validate:"oneof=debug info warn error"`
}

type labelSelectorInput struct {
	Label string `validate:"k8s_label_selector"`
}

type fieldSelectorInput struct {
	Field string `validate:"k8s_field_selector"`
}

func TestHandleValidatorError(t *testing.T) {
	t.Run("correct error", func(t *testing.T) {
		a := TestStruct{Field2: "test"}
		expectedErr := "validation failed: TestStruct.Field1 (required=), TestStruct.Field2 (oneof=debug info warn error)"

		err := v.Struct(a)
		resultErr := handleValidatorError(err)

		require.Error(t, resultErr)
		assert.Contains(t, resultErr.Error(), expectedErr)
	})

	t.Run("unexpected error", func(t *testing.T) {
		expectedErr := "unexpected validation error: assert.AnError general error for testing"
		err := handleValidatorError(assert.AnError)

		require.Error(t, err)
		assert.Equal(t, expectedErr, err.Error())
	})
}

func TestValidateInputStruct(t *testing.T) {
	t.Run("correct input", func(t *testing.T) {
		a := TestStruct{Field2: "test"}
		resultErr := validateInputStruct(a)

		require.NoError(t, resultErr)
	})

	t.Run("nil value", func(t *testing.T) {
		expectedErr := "input is nil"
		err := validateInputStruct(nil)

		require.Error(t, err)
		assert.Equal(t, expectedErr, err.Error())
	})

	t.Run("nil pointer value", func(t *testing.T) {
		expectedErr := "input is a nil pointer"
		err := validateInputStruct((*any)(nil))

		require.Error(t, err)
		assert.Equal(t, expectedErr, err.Error())
	})
}

func TestValidateStruct(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		a := TestStruct{Field1: 1025, Field2: "info"}

		err := ValidateStruct(a)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		a := TestStruct{Field2: "test"}
		expectedErr := "validation failed: TestStruct.Field1 (required=), TestStruct.Field2 (oneof=debug info warn error)"

		err := ValidateStruct(a)
		require.Error(t, err)
		assert.Equal(t, expectedErr, err.Error())
	})

	t.Run("invalid input", func(t *testing.T) {
		t.Run("nil value", func(t *testing.T) {
			expectedErr := "input is nil"
			err := ValidateStruct(nil)

			require.Error(t, err)
			assert.Equal(t, expectedErr, err.Error())
		})

		t.Run("nil pointer value", func(t *testing.T) {
			expectedErr := "input is a nil pointer"
			err := ValidateStruct((*any)(nil))

			require.Error(t, err)
			assert.Equal(t, expectedErr, err.Error())
		})

		t.Run("not a struct value", func(t *testing.T) {
			expectedErr := "unexpected validation error: validator: (nil int)"
			err := ValidateStruct(1)

			require.Error(t, err)
			assert.Equal(t, expectedErr, err.Error())
		})
	})
}

func TestValidateWithTag(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		err := ValidateWithTag("debug", "oneof=debug info warn error")
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := "validation failed: string qwe (oneof=debug info warn error)"

		err := ValidateWithTag("qwe", "oneof=debug info warn error")
		require.Error(t, err)
		assert.Equal(t, expectedErr, err.Error())
	})

	t.Run("nil value", func(t *testing.T) {
		expectedErr := "validation failed: nil value (oneof=debug info warn error)"

		err := ValidateWithTag(nil, "oneof=debug info warn error")
		require.Error(t, err)
		assert.Equal(t, expectedErr, err.Error())
	})
}

func TestUrlPrefixValidator(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		err := ValidateWithTag("https://localhost:8081", "url_prefix")
		require.NoError(t, err)
	})

	t.Run("negative", func(t *testing.T) {
		expectedErr := "validation failed: string localhost:8081 (url_prefix=)"

		err := ValidateWithTag("localhost:8081", "url_prefix")
		require.Error(t, err)
		assert.Equal(t, expectedErr, err.Error())
	})
}

func TestRegisterValidation(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		err := RegisterValidation("is-even", func(fl validator.FieldLevel) bool {
			return fl.Field().Int()%2 == 0
		})
		require.NoError(t, err)

		t.Run("is-even positive", func(t *testing.T) {
			err := ValidateWithTag(2, "is-even")
			require.NoError(t, err)
		})

		t.Run("is-even negative", func(t *testing.T) {
			expectedErr := "validation failed: int %!s(int=1) (is-even=)"

			err := ValidateWithTag(1, "is-even")
			require.Error(t, err)
			assert.Equal(t, expectedErr, err.Error())
		})
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("tag empty", func(t *testing.T) {
			expectedErr := "function Key cannot be empty"

			err := RegisterValidation("", func(fl validator.FieldLevel) bool {
				return fl.Field().Int()%2 == 0
			})
			require.Error(t, err)
			assert.Equal(t, expectedErr, err.Error())
		})

		t.Run("func empty", func(t *testing.T) {
			expectedErr := "function cannot be empty"

			err := RegisterValidation("is-even", nil)
			require.Error(t, err)
			assert.Equal(t, expectedErr, err.Error())
		})
	})
}

func TestLabelSelectorValidator_AllSyntax(t *testing.T) {
	v := validator.New()
	labelSelectorValidator(v)

	tests := []struct {
		name  string
		input labelSelectorInput
		valid bool
	}{
		// valid syntax
		{"Equal", labelSelectorInput{"env=prod"}, true},
		{"DoubleEqual", labelSelectorInput{"env==prod"}, true},
		{"NotEqual", labelSelectorInput{"env!=prod"}, true},
		{"InOperator", labelSelectorInput{"env in (prod,dev)"}, true},
		{"NotInOperator", labelSelectorInput{"env notin (qa,stage)"}, true},
		{"KeyExists", labelSelectorInput{"team"}, true},
		{"KeyNotExists", labelSelectorInput{"!deprecated"}, true},
		{"MultipleAND", labelSelectorInput{"env=prod,team=platform"}, true},

		// invalid syntax
		{"InvalidOperator", labelSelectorInput{"env~prod"}, false},
		{"InvalidExpression", labelSelectorInput{"env prod"}, false},
		{"InvalidKey", labelSelectorInput{"@@=value"}, false},
		{"InvalidSet", labelSelectorInput{"team in prod"}, false},
		{"Empty", labelSelectorInput{""}, false},
		{"TrailingComma", labelSelectorInput{"env=prod,"}, false},
	}

	for _, tt := range tests {
		err := v.Struct(tt.input)
		if tt.valid {
			assert.NoError(t, err, tt.name)
		} else {
			assert.Error(t, err, tt.name)
		}
	}
}

func TestFieldSelectorValidator_AllSyntax(t *testing.T) {
	v := validator.New()
	fieldSelectorValidator(v)

	tests := []struct {
		name  string
		input fieldSelectorInput
		valid bool
	}{
		// valid syntax
		{"Equal", fieldSelectorInput{"metadata.name=default"}, true},
		{"DoubleEqual", fieldSelectorInput{"metadata.name==default"}, true},
		{"NotEqual", fieldSelectorInput{"metadata.name!=kube-system"}, true},
		{"MultipleAND", fieldSelectorInput{"metadata.name=default,status.phase=Running"}, true},

		// invalid syntax
		{"InvalidOperator", fieldSelectorInput{"metadata.name~value"}, false},
		{"InvalidField", fieldSelectorInput{"spec.replicas=3"}, false}, // not indexable
		{"InvalidFormat", fieldSelectorInput{"metadata.name default"}, false},
		{"Empty", fieldSelectorInput{""}, false},
		{"TrailingComma", fieldSelectorInput{"metadata.name=default,"}, false},
	}

	for _, tt := range tests {
		err := v.Struct(tt.input)
		if tt.valid {
			assert.NoError(t, err, tt.name)
		} else {
			assert.Error(t, err, tt.name)
		}
	}
}
