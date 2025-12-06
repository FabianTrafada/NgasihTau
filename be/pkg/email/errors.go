package email

import "errors"

var (
	ErrTemplateNotFound      = errors.New("email template not found")
	ErrSendFailed            = errors.New("failed to send email")
	ErrInvalidConfig         = errors.New("invalid email configuration")
	ErrProviderNotConfigured = errors.New("email provider not configured")
)
