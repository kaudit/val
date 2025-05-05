# val - Thread-Safe Validation Library

## Overview

`val` is a Go package providing a thread-safe validation mechanism using the [`go-playground/validator/v10`](https://github.com/go-playground/validator) library. It enables struct and tag-based validation, along with support for custom validation rules. The package is designed as a singleton, ensuring that a single validator instance is shared across the application without requiring additional setup.

## Features

- **Thread-Safe Validation**: Ensures safe concurrent access.
- **Struct and Field Validation**: Supports validating entire structs and individual fields.
- **Custom Validation Rules**: Enables defining and registering custom validation tags.
- **Singleton Pattern**: No need to instantiate multiple validators.
- **Comprehensive Error Handling**: Provides structured error messages for failed validations.

## Installation

### Dependencies Installation

To use this package, ensure you have Go installed and set up.

```sh
# Install the go-playground validator library
go get github.com/go-playground/validator/v10
```

### Package Installation

To install `val` in your Go project, run:

```sh
go get github.com/kaudit/val
```

## Usage

### Validating a Struct

```go
package main

import (
	"fmt"
	"github.com/kaudit/val"
)

type User struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=18,lte=65"`
}

func main() {
	user := User{Name: "", Email: "invalid-email", Age: 70}

	err := val.ValidateStruct(user)
	if err != nil {
		fmt.Println("Validation error:", err)
	} else {
		fmt.Println("Validation passed!")
	}
}
```

### Validating a Single Field with a Tag

```go
package main

import (
    "fmt"
    "github.com/kaudit/val"
)

func main() {
    email := "test@example.com"
    err := val.ValidateWithTag(email, "email")
    if err != nil {
        fmt.Println("Validation error:", err)
    } else {
        fmt.Println("Validation passed!")
    }
}
```

### Registering a Custom Validation Rule

```go
package main

import (
    "fmt"
    "github.com/go-playground/validator/v10"
    "github.com/kaudit/val"
)

func main() {
    // Register a custom validation rule "is-even"
    err := val.RegisterValidation("is-even", func(fl validator.FieldLevel) bool {
        return fl.Field().Int()%2 == 0
    })
    if err != nil {
        fmt.Println("Failed to register validation rule:", err)
        return
    }

    // Validate a field using the custom rule
    number := 4
    err = val.ValidateWithTag(number, "is-even")
    if err != nil {
        fmt.Println("Validation error:", err)
    } else {
        fmt.Println("Validation passed!")
    }
}
```

## API Documentation

### Available Functions

#### `ValidateStruct(s any) error`
Validates a struct based on its validation tags.  
Ensures the input is a valid struct or a pointer to a struct.  
Validates the struct fields based on their tags.  
Returns detailed, formatted errors for each validation failure.

#### `ValidateWithTag(variable any, tag string) error`
Validates a single variable using a specified validation tag.  
Uses the go-playground validator to validate the `variable` against the provided `tag`.  
If validation fails, it processes and returns a structured error.

#### `RegisterValidation(tag string, fn validator.Func) error`
Registers a custom validation function for a specific tag.

## Custom Validation Rules

### `url_prefix`
Ensures that a string starts with either `http://` or `https://`.

Example usage:

```go
package main

import (
    "fmt"
    "github.com/kaudit/val"
)

func main() {
    url := "https://example.com"
    err := val.ValidateWithTag(url, "url_prefix")
    if err != nil {
        fmt.Println("Validation error:", err)
    } else {
        fmt.Println("Validation passed!")
    }
}
```

### `k8s_label_selector`
Ensures that a string represents a valid Kubernetes label selector. This validator rejects empty values and checks the syntax against Kubernetes label-selector parsing rules.

Example usage:

```go
package main

import (
    "fmt"
    "github.com/kaudit/val"
)

func main() {
    selector := "app=myapp,env=dev"
    err := val.ValidateWithTag(selector, "k8s_label_selector")
    if err != nil {
        fmt.Println("Validation error:", err)
    } else {
        fmt.Println("Validation passed!")
    }
}
```

### `k8s_field_selector`
Ensures that a string represents a valid Kubernetes field selector. This validator also enforces a whitelist of recognized field keys to prevent invalid fields.

The following field keys are currently allowed:
- `metadata.name`
- `metadata.namespace`
- `status.phase`
- `spec.nodeName`
- `spec.unschedulable`
- `status.hostIP`
- `status.podIP`
- `spec.type`

Any other field keys will cause the validation to fail.

Example usage:

```go
package main

import (
    "fmt"
    "github.com/kaudit/val"
)

func main() {
    fieldSel := "metadata.name=my-pod"
    err := val.ValidateWithTag(fieldSel, "k8s_field_selector")
    if err != nil {
        fmt.Println("Validation error:", err)
    } else {
        fmt.Println("Validation passed!")
    }
}
```

## License

This project is licensed under the MIT License. See the [LICENSE](https://opensource.org/licenses/MIT) file for details.

## Thanks

Special thanks to the contributors and maintainers of:
- [`go-playground/validator`](https://github.com/go-playground/validator) for providing the validation framework.
