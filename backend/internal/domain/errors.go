package domain

// AppError represents a domain-level error with HTTP status and error code.
type AppError struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Status  int          `json:"-"`
	Details []FieldError `json:"details,omitempty"`
}

// FieldError represents a validation error for a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	return e.Message
}

// Sentinel errors for common domain cases.
var (
	ErrNotFound     = &AppError{Code: "NOT_FOUND", Message: "resource not found", Status: 404}
	ErrConflict     = &AppError{Code: "CONFLICT", Message: "resource already exists", Status: 409}
	ErrUnauthorized = &AppError{Code: "UNAUTHORIZED", Message: "invalid credentials", Status: 401}
	ErrForbidden    = &AppError{Code: "FORBIDDEN", Message: "access denied", Status: 403}
)

// NewAppError creates a new AppError with the given code, message, and status.
func NewAppError(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// NewValidationError creates a validation error with field-level details.
func NewValidationError(message string, details []FieldError) *AppError {
	return &AppError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Status:  400,
		Details: details,
	}
}
