package errors

import "fmt"

type NotFoundError struct {
    Resource string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s not found", e.Resource)
}

type ValidationError struct {
    Field string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("Invalid value for field: %s", e.Field)
}
