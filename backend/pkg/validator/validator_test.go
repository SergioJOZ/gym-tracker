package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
	Age      int    `validate:"gte=0,lte=150"`
	Name     string `validate:"required,max=100"`
}

func TestValidate_ValidStruct(t *testing.T) {
	s := TestStruct{
		Email:    "test@example.com",
		Password: "password123",
		Age:      25,
		Name:     "John Doe",
	}

	err := Validate(s)
	assert.NoError(t, err)
}

func TestValidate_InvalidEmail(t *testing.T) {
	s := TestStruct{
		Email:    "not-an-email",
		Password: "password123",
		Age:      25,
		Name:     "John Doe",
	}

	err := Validate(s)
	assert.Error(t, err)

	validationErr, ok := err.(*ValidationErrors)
	assert.True(t, ok)
	assert.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "email", validationErr.Errors[0].Field)
}

func TestValidate_PasswordTooShort(t *testing.T) {
	s := TestStruct{
		Email:    "test@example.com",
		Password: "short",
		Age:      25,
		Name:     "John Doe",
	}

	err := Validate(s)
	assert.Error(t, err)

	validationErr, ok := err.(*ValidationErrors)
	assert.True(t, ok)
	assert.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "password", validationErr.Errors[0].Field)
	assert.Contains(t, validationErr.Errors[0].Message, "8")
}

func TestValidate_MultipleErrors(t *testing.T) {
	s := TestStruct{
		Email:    "invalid",
		Password: "short",
		Age:      200,
		Name:     "",
	}

	err := Validate(s)
	assert.Error(t, err)

	validationErr, ok := err.(*ValidationErrors)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(validationErr.Errors), 3)
}

func TestValidate_MissingRequired(t *testing.T) {
	s := TestStruct{
		Email:    "",
		Password: "",
		Age:      25,
		Name:     "",
	}

	err := Validate(s)
	assert.Error(t, err)

	validationErr, ok := err.(*ValidationErrors)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(validationErr.Errors), 3)
}

func TestValidate_AgeOutOfRange(t *testing.T) {
	s := TestStruct{
		Email:    "test@example.com",
		Password: "password123",
		Age:      -5,
		Name:     "John Doe",
	}

	err := Validate(s)
	assert.Error(t, err)

	validationErr, ok := err.(*ValidationErrors)
	assert.True(t, ok)
	assert.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "age", validationErr.Errors[0].Field)
}

func TestValidate_NameTooLong(t *testing.T) {
	s := TestStruct{
		Email:    "test@example.com",
		Password: "password123",
		Age:      25,
		Name:     string(make([]byte, 101)),
	}

	err := Validate(s)
	assert.Error(t, err)

	validationErr, ok := err.(*ValidationErrors)
	assert.True(t, ok)
	assert.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "name", validationErr.Errors[0].Field)
}

func TestValidate_NonStruct(t *testing.T) {
	err := Validate("not a struct")
	assert.Error(t, err)
}

func TestValidate_Pointer(t *testing.T) {
	s := &TestStruct{
		Email:    "test@example.com",
		Password: "password123",
		Age:      25,
		Name:     "John Doe",
	}

	err := Validate(s)
	assert.NoError(t, err)
}

func TestValidate_NilPointer(t *testing.T) {
	var s *TestStruct
	err := Validate(s)
	assert.Error(t, err)
}

func TestValidationErrors_Error(t *testing.T) {
	validationErr := &ValidationErrors{
		Errors: []FieldError{
			{Field: "email", Message: "invalid email"},
			{Field: "password", Message: "too short"},
		},
	}

	errMsg := validationErr.Error()
	assert.Contains(t, errMsg, "email")
	assert.Contains(t, errMsg, "password")
}

func TestValidate_EmptyStruct(t *testing.T) {
	type EmptyStruct struct{}
	s := EmptyStruct{}
	err := Validate(s)
	assert.NoError(t, err)
}
