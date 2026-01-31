package email

import (
	"context"
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridProvider struct {
	client    *sendgrid.Client
	fromEmail string
	fromName  string
}

func NewSendGridProvider(apiKey, fromEmail, fromName string) (*SendGridProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("%w: SendGrid API key is required", ErrInvalidConfig)
	}

	if fromEmail == "" {
		return nil, fmt.Errorf("%w: SendGrid from email is required", ErrInvalidConfig)
	}

	return &SendGridProvider{
		client:    sendgrid.NewSendClient(apiKey),
		fromEmail: fromEmail,
		fromName:  fromName,
	}, nil
}

func (p *SendGridProvider) Send(ctx context.Context, email *Email) error {
	from := mail.NewEmail(p.fromName, p.fromEmail)
	to := mail.NewEmail("", email.To)

	message := mail.NewSingleEmail(from, email.Subject, to, email.TextBody, email.HTMLBody)

	response, err := p.client.SendWithContext(ctx, message)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("%w: SendGrid returned status %d: %s", ErrSendFailed, response.StatusCode, response.Body)
	}

	return nil
}
