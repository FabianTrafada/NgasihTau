package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return fmt.Sprintf("configuration validation failed:\n  - %s", strings.Join(msgs, "\n  - "))
}

// Validate validates the configuration struct and returns detailed errors.
func Validate(cfg *Config) error {
	validate := validator.New()

	if err := validate.Struct(cfg); err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			return formatValidationErrors(validationErrs)
		}
		return err
	}

	// Additional custom validations
	if errs := customValidations(cfg); len(errs) > 0 {
		return errs
	}

	return nil
}

// formatValidationErrors converts validator errors to our custom format.
func formatValidationErrors(errs validator.ValidationErrors) ValidationErrors {
	var result ValidationErrors
	for _, err := range errs {
		result = append(result, ValidationError{
			Field:   formatFieldName(err.Namespace()),
			Message: formatValidationMessage(err),
		})
	}
	return result
}

// formatFieldName converts struct field path to config-style path.
func formatFieldName(namespace string) string {
	// Remove "Config." prefix
	name := strings.TrimPrefix(namespace, "Config.")
	// Convert to lowercase with dots
	return strings.ToLower(name)
}

// formatValidationMessage creates a human-readable validation message.
func formatValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "is required"
	case "min":
		return fmt.Sprintf("must be at least %s", err.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", err.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", err.Param())
	case "url":
		return "must be a valid URL"
	case "email":
		return "must be a valid email address"
	default:
		return fmt.Sprintf("failed validation: %s", err.Tag())
	}
}

// customValidations performs additional validation logic beyond struct tags.
func customValidations(cfg *Config) ValidationErrors {
	var errs ValidationErrors

	// Validate JWT secret in production
	if cfg.App.Env == "production" {
		if cfg.JWT.Secret == "" {
			errs = append(errs, ValidationError{
				Field:   "jwt.secret",
				Message: "is required in production environment",
			})
		}
		if len(cfg.JWT.Secret) < 32 {
			errs = append(errs, ValidationError{
				Field:   "jwt.secret",
				Message: "must be at least 32 characters in production",
			})
		}
	}

	// Validate OpenAI API key if AI service is configured
	if cfg.AISvc.Port > 0 && cfg.OpenAI.APIKey == "" && cfg.App.Env == "production" {
		errs = append(errs, ValidationError{
			Field:   "openai.api_key",
			Message: "is required when AI service is enabled in production",
		})
	}

	// Validate chunk sizes
	if cfg.OpenAI.ChunkSizeMin > cfg.OpenAI.ChunkSizeMax {
		errs = append(errs, ValidationError{
			Field:   "openai.chunk_size_min",
			Message: "must be less than or equal to chunk_size_max",
		})
	}

	// Validate refresh token expiry is greater than access token expiry
	if cfg.JWT.RefreshTokenExpiry <= cfg.JWT.AccessTokenExpiry {
		errs = append(errs, ValidationError{
			Field:   "jwt.refresh_token_expiry",
			Message: "must be greater than access_token_expiry",
		})
	}

	// Validate database pool settings
	validateDBPool(&cfg.UserDB, "user_db", &errs)
	validateDBPool(&cfg.PodDB, "pod_db", &errs)
	validateDBPool(&cfg.MaterialDB, "material_db", &errs)
	validateDBPool(&cfg.AIDB, "ai_db", &errs)
	validateDBPool(&cfg.NotifDB, "notification_db", &errs)

	return errs
}

// validateDBPool validates database connection pool settings.
func validateDBPool(db *DatabaseConfig, prefix string, errs *ValidationErrors) {
	if db.MaxIdleConns > db.MaxOpenConns {
		*errs = append(*errs, ValidationError{
			Field:   prefix + ".max_idle_conns",
			Message: "must be less than or equal to max_open_conns",
		})
	}
}

// MustLoad loads configuration and panics on error.
// Use this in main() for fail-fast behavior.
func MustLoad(configPath string) *Config {
	cfg, err := Load(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration: %v", err))
	}
	return cfg
}

// MustLoadFromEnv loads configuration from environment and panics on error.
func MustLoadFromEnv() *Config {
	return MustLoad("")
}
