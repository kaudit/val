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

func TestHandleValidatorError(t *testing.T) {
	t.Run("correct error", func(t *testing.T) {
		a := TestStruct{Field2: "test"}
		expectedErr := "validation failed: TestStruct.Field1 (required=), TestStruct.Field2 (oneof=debug info warn error)"

		err := v.Struct(a)
		resultErr := handleValidatorError(err)

		require.Error(t, resultErr)
		assert.Equal(t, expectedErr, expectedErr)
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

	t.Run("not a struct value", func(t *testing.T) {
		expectedErr := "input is not a struct: got int"
		err := validateInputStruct(1)

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
			expectedErr := "input is not a struct: got int"
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
		err := ValidateWithTag("https://localhost:8081", "urlprefix")
		require.NoError(t, err)
	})

	t.Run("negative", func(t *testing.T) {
		expectedErr := "validation failed: string localhost:8081 (urlprefix=)"

		err := ValidateWithTag("localhost:8081", "urlprefix")
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
