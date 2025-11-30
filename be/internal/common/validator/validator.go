// Package validator provides validation utilities using go-playground/validator
// with custom validation rules for common fields.
// Implements requirement 10.4 for input validation at API boundaries.
package validator

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"

	"ngasihtau/internal/common/errors"
)

var (
	instance *Validator
	once     sync.Once
)

// Validator wraps go-playground/validator with custom rules and error formatting.
type Validator struct {
	validate *validator.Validate
}

// Get returns the singleton Validator instance.
// The validator is initialized with custom validation rules on first call.
func Get() *Validator {
	once.Do(func() {
		instance = newValidator()
	})
	return instance
}

// newValidator creates a new Validator with custom rules registered.
func newValidator() *Validator {
	v := validator.New()

	// Use JSON tag names for field names in error messages
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		if name == "" {
			return fld.Name
		}
		return name
	})

	// Register custom validation rules
	registerCustomRules(v)

	return &Validator{validate: v}
}

// Struct validates a struct and returns an AppError with field details if validation fails.
func (v *Validator) Struct(s any) *errors.AppError {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	return v.formatValidationErrors(err)
}

// Var validates a single variable against a tag.
func (v *Validator) Var(field any, tag string) *errors.AppError {
	err := v.validate.Var(field, tag)
	if err == nil {
		return nil
	}

	return v.formatValidationErrors(err)
}

// VarWithName validates a single variable and uses the provided name in error messages.
func (v *Validator) VarWithName(field any, tag, name string) *errors.AppError {
	err := v.validate.Var(field, tag)
	if err == nil {
		return nil
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return errors.Validation("validation failed")
	}

	details := make([]errors.ErrorDetail, 0, len(validationErrs))
	for _, e := range validationErrs {
		details = append(details, errors.ErrorDetail{
			Field:   name,
			Message: formatErrorMessage(e),
		})
	}

	return errors.Validation("validation failed", details...)
}

// formatValidationErrors converts validator.ValidationErrors to AppError with details.
func (v *Validator) formatValidationErrors(err error) *errors.AppError {
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return errors.Validation("validation failed")
	}

	details := make([]errors.ErrorDetail, 0, len(validationErrs))
	for _, e := range validationErrs {
		details = append(details, errors.ErrorDetail{
			Field:   e.Field(),
			Message: formatErrorMessage(e),
		})
	}

	return errors.Validation("validation failed", details...)
}

// formatErrorMessage creates a human-readable error message for a validation error.
func formatErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "uuid", "uuid4":
		return "must be a valid UUID"
	case "min":
		if e.Kind() == reflect.String {
			return "must be at least " + e.Param() + " characters"
		}
		return "must be at least " + e.Param()
	case "max":
		if e.Kind() == reflect.String {
			return "must be at most " + e.Param() + " characters"
		}
		return "must be at most " + e.Param()
	case "len":
		return "must be exactly " + e.Param() + " characters"
	case "oneof":
		return "must be one of: " + e.Param()
	case "url":
		return "must be a valid URL"
	case "gte":
		return "must be greater than or equal to " + e.Param()
	case "lte":
		return "must be less than or equal to " + e.Param()
	case "gt":
		return "must be greater than " + e.Param()
	case "lt":
		return "must be less than " + e.Param()
	case "alphanum":
		return "must contain only alphanumeric characters"
	case "alpha":
		return "must contain only alphabetic characters"
	case "numeric":
		return "must be a numeric value"
	case "slug":
		return "must be a valid slug (lowercase letters, numbers, and hyphens)"
	case "password":
		return "must be at least 8 characters with uppercase, lowercase, and number"
	case "username":
		return "must be 3-30 characters, alphanumeric with underscores"
	case "safe_string":
		return "contains invalid characters"
	default:
		return "failed validation: " + e.Tag()
	}
}
