package email

import (
	"context"
	"fmt"

	"ngasihtau/internal/common/config"
)

type ProviderType string

const (
	ProviderSendGrid ProviderType = "sendgrid"
	ProviderSES      ProviderType = "ses"
	ProviderSMTP     ProviderType = "smtp"
)

func NewProvider(ctx context.Context, cfg config.EmailConfig) (Provider, error) {
	switch ProviderType(cfg.Provider) {
	case ProviderSendGrid:
		return NewSendGridProvider(cfg.SendGridKey, cfg.FromEmail, cfg.FromName)

	case ProviderSES:
		return NewSESProvider(ctx, cfg.SESRegion, cfg.SESAccessKey, cfg.SESSecretKey, cfg.FromEmail, cfg.FromName)

	case ProviderSMTP:
		return nil, fmt.Errorf("%w: SMTP provider not yet implemented", ErrProviderNotConfigured)

	default:
		return nil, fmt.Errorf("%w: unknown provider type: %s", ErrInvalidConfig, cfg.Provider)
	}
}
