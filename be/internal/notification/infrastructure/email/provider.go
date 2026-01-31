package email

import (
	"context"
	"fmt"

	pkgemail "ngasihtau/pkg/email"
)

// Email represents an email message to be sent.
type Email = pkgemail.Email

// Provider defines the interface for email providers.
type Provider = pkgemail.Provider

// TemplateData holds common data for email templates.
type TemplateData = pkgemail.TemplateData

// TemplateRenderer renders email templates.
type TemplateRenderer = pkgemail.TemplateRenderer

// NewTemplateRenderer creates a new template renderer.
func NewTemplateRenderer() *TemplateRenderer {
	return pkgemail.NewTemplateRenderer()
}

// Service wraps the email provider with template rendering capabilities.
type Service struct {
	provider Provider
	renderer *TemplateRenderer
}

// NewService creates a new email service.
func NewService(provider Provider) *Service {
	return &Service{
		provider: provider,
		renderer: pkgemail.NewTemplateRenderer(),
	}
}

// SendVerificationEmail sends an email verification email.
func (s *Service) SendVerificationEmail(ctx context.Context, to, recipientName, verificationURL string) error {
	html, text, err := s.renderer.Render("verification", TemplateData{
		RecipientName: recipientName,
		ActionURL:     verificationURL,
		AppName:       "NgasihTau",
		SupportEmail:  "support@ngasihtau.com",
		ExpiryHours:   24,
	})
	if err != nil {
		return err
	}

	return s.provider.Send(ctx, &Email{
		To:       to,
		Subject:  "Verify Your Email - NgasihTau",
		HTMLBody: html,
		TextBody: text,
	})
}

// SendPasswordResetEmail sends a password reset email.
func (s *Service) SendPasswordResetEmail(ctx context.Context, to, recipientName, resetURL string) error {
	html, text, err := s.renderer.Render("password_reset", TemplateData{
		RecipientName: recipientName,
		ActionURL:     resetURL,
		AppName:       "NgasihTau",
		SupportEmail:  "support@ngasihtau.com",
		ExpiryHours:   1,
	})
	if err != nil {
		return err
	}

	return s.provider.Send(ctx, &Email{
		To:       to,
		Subject:  "Reset Your Password - NgasihTau",
		HTMLBody: html,
		TextBody: text,
	})
}

// SendCollaboratorInviteEmail sends a collaborator invitation email.
func (s *Service) SendCollaboratorInviteEmail(ctx context.Context, to, recipientName, inviterName, podName, inviteURL string) error {
	html, text, err := s.renderer.Render("collaborator_invite", TemplateData{
		RecipientName: recipientName,
		InviterName:   inviterName,
		PodName:       podName,
		ActionURL:     inviteURL,
		AppName:       "NgasihTau",
		SupportEmail:  "support@ngasihtau.com",
	})
	if err != nil {
		return err
	}

	return s.provider.Send(ctx, &Email{
		To:       to,
		Subject:  fmt.Sprintf("You're Invited to Collaborate on %s - NgasihTau", podName),
		HTMLBody: html,
		TextBody: text,
	})
}
