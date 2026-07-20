package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FieldError represents a validation error for a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents a collection of field validation errors.
type ValidationErrors struct {
	Errors []FieldError `json:"errors"`
}

// Error implements the error interface.
func (e *ValidationErrors) Error() string {
	messages := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		messages[i] = fmt.Sprintf("%s: %s", err.Field, err.Message)
	}
	return strings.Join(messages, "; ")
}

// defaultValidator is the shared validator instance.
var defaultValidator = validator.New()

// Validate validates a struct using the default validator.
// Returns *ValidationErrors if validation fails, or nil if valid.
func Validate(s interface{}) error {
	err := defaultValidator.Struct(s)
	if err == nil {
		return nil
	}

	// Check if it's a validation error
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	// Convert to our FieldError format
	fieldErrors := make([]FieldError, 0, len(validationErrs))
	for _, ve := range validationErrs {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   toSnakeCase(ve.Field()),
			Message: formatMessage(ve),
		})
	}

	return &ValidationErrors{Errors: fieldErrors}
}

// toSnakeCase converts a field name to snake_case.
func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, []rune{r}...)
	}
	return strings.ToLower(string(result))
}

// formatMessage creates a human-readable error message from a validation error.
func formatMessage(ve validator.FieldError) string {
	switch ve.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", ve.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", ve.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", ve.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", ve.Param())
	default:
		return fmt.Sprintf("failed on '%s' validation", ve.Tag())
	}
}
