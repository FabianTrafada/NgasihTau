package application

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"ngasihtau/pkg/email"
	natspkg "ngasihtau/pkg/nats"
)

type EmailWorkerConfig struct {
	AppName        string
	AppUrl         string
	SupportEmail   string
	MaxRetries     int
	RetryBaseDelay time.Duration
}

func DefaultEmailWorkerConfig() EmailWorkerConfig {
	return EmailWorkerConfig{
		AppName:        "NgasihTau",
		AppUrl:         "http://localhost:3000", // Override via config.App.FrontendURL
		SupportEmail:   "support@ngasihtau.com",
		MaxRetries:     3,
		RetryBaseDelay: time.Second,
	}
}

type EmailWorker struct {
	natsClient       *natspkg.Client
	emailProvider    email.Provider
	templateRenderer *email.TemplateRenderer
	config           EmailWorkerConfig
	subscriptions    []*natspkg.Subscription
}

func NewEmailWorker(
	natsClient *natspkg.Client,
	emailProvider email.Provider,
	config EmailWorkerConfig,
) *EmailWorker {
	return &EmailWorker{
		natsClient:       natsClient,
		emailProvider:    emailProvider,
		templateRenderer: email.NewTemplateRenderer(),
		config:           config,
		subscriptions:    make([]*natspkg.Subscription, 0),
	}
}

func (w *EmailWorker) Start(ctx context.Context) error {
	if err := w.natsClient.EnsureStream(ctx, natspkg.StreamEmail, natspkg.StreamSubjects[natspkg.StreamEmail]); err != nil {
		return fmt.Errorf("failed to ensure Email stream: %w", err)
	}

	verificationSub, err := w.natsClient.SubscribeSimple(
		ctx,
		natspkg.DefaultSubscribeConfig(
			natspkg.StreamEmail,
			"notification-email-verification",
			[]string{natspkg.SubjectEmailVerification},
		),
		w.handleEmailVerification,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to email verification events: %w", err)
	}
	w.subscriptions = append(w.subscriptions, verificationSub)
	log.Info().Msg("subscribed to email verification events")

	passwordResetSub, err := w.natsClient.SubscribeSimple(
		ctx,
		natspkg.DefaultSubscribeConfig(
			natspkg.StreamEmail,
			"notification-email-password-reset",
			[]string{natspkg.SubjectEmailPasswordReset},
		),
		w.handlePasswordReset,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to password reset events: %w", err)
	}

	w.subscriptions = append(w.subscriptions, passwordResetSub)
	log.Info().Msg("subscribed to email.password_reset events")

	inviteSub, err := w.natsClient.SubscribeSimple(
		ctx,
		natspkg.DefaultSubscribeConfig(
			natspkg.StreamEmail,
			"notification-email-collaborator-invite",
			[]string{natspkg.SubjectEmailCollaboratorInvite},
		),
		w.handleCollaboratorInvite,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to collaborator invite events: %w", err)
	}
	w.subscriptions = append(w.subscriptions, inviteSub)
	log.Info().Msg("subscribed to email.collaborator_invite events")

	log.Info().Msg("email worker started successfully")
	return nil
}

func (w *EmailWorker) Stop() {
	for _, sub := range w.subscriptions {
		sub.Stop()
	}
	log.Info().Msg("email worker stopped")
}

func (w *EmailWorker) handleEmailVerification(ctx context.Context, event natspkg.CloudEvent) error {
	data, err := natspkg.ParseEventData[natspkg.EmailVerificationEvent](event)
	if err != nil {
		log.Error().Err(err).Str("event_id", event.ID).Msg("failed to parse email verification event")
		return err
	}

	log.Info().
		Str("user_id", data.UserID.String()).
		Str("email", data.Email).
		Msg("processing email verification event")

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", w.config.AppUrl, data.Token)

	htmlBody, textBody, err := w.templateRenderer.Render("verification", email.TemplateData{
		RecipientName: data.Name,
		ActionURL:     verificationURL,
		AppName:       w.config.AppName,
		SupportEmail:  w.config.SupportEmail,
		ExpiryHours:   24,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to render verification email template")
		return err
	}

	emailMsg := &email.Email{
		To:       data.Email,
		Subject:  fmt.Sprintf("Verify your %s account", w.config.AppName),
		HTMLBody: htmlBody,
		TextBody: textBody,
	}

	if err := w.sendWithRetry(ctx, emailMsg); err != nil {
		log.Error().Err(err).
			Str("email", data.Email).
			Msg("failed to send verification email after retries")
		return err
	}

	log.Info().
		Str("email", data.Email).
		Msg("verification email sent successfully")
	return nil
}

func (w *EmailWorker) handlePasswordReset(ctx context.Context, event natspkg.CloudEvent) error {
	data, err := natspkg.ParseEventData[natspkg.EmailPasswordResetEvent](event)
	if err != nil {
		log.Error().Err(err).Str("event_id", event.ID).Msg("failed to parse password reset event")
		return err
	}

	log.Info().
		Str("user_id", data.UserID.String()).
		Str("email", data.Email).
		Msg("processing password reset event")

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", w.config.AppUrl, data.Token)

	htmlBody, textBody, err := w.templateRenderer.Render("password_reset", email.TemplateData{
		RecipientName: data.Name,
		ActionURL:     resetURL,
		AppName:       w.config.AppName,
		SupportEmail:  w.config.SupportEmail,
		ExpiryHours:   1,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to render password reset email template")
		return err
	}

	emailMsg := &email.Email{
		To:       data.Email,
		Subject:  fmt.Sprintf("Reset your %s password", w.config.AppName),
		HTMLBody: htmlBody,
		TextBody: textBody,
	}

	if err := w.sendWithRetry(ctx, emailMsg); err != nil {
		log.Error().Err(err).
			Str("email", data.Email).
			Msg("failed to send password reset email after retries")
		return err
	}

	log.Info().
		Str("email", data.Email).
		Msg("password reset email sent successfully")
	return nil
}

func (w *EmailWorker) handleCollaboratorInvite(ctx context.Context, event natspkg.CloudEvent) error {
	data, err := natspkg.ParseEventData[natspkg.EmailCollaboratorInviteEvent](event)
	if err != nil {
		log.Error().Err(err).Str("event_id", event.ID).Msg("failed to parse collaborator invite event")
		return err
	}

	log.Info().
		Str("invitee_id", data.InviteeID.String()).
		Str("email", data.Email).
		Str("pod_id", data.PodID.String()).
		Msg("processing collaborator invite event")

	inviteURL := fmt.Sprintf("%s/pods/%s/accept-invite", w.config.AppUrl, data.PodID.String())

	htmlBody, textBody, err := w.templateRenderer.Render("collaborator_invite", email.TemplateData{
		RecipientName: "", // We don't have the invitee name in the event
		ActionURL:     inviteURL,
		AppName:       w.config.AppName,
		SupportEmail:  w.config.SupportEmail,
		PodName:       data.PodName,
		InviterName:   data.InviterName,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to render collaborator invite email template")
		return err
	}

	emailMsg := &email.Email{
		To:       data.Email,
		Subject:  fmt.Sprintf("%s invited you to collaborate on %s", data.InviterName, data.PodName),
		HTMLBody: htmlBody,
		TextBody: textBody,
	}

	if err := w.sendWithRetry(ctx, emailMsg); err != nil {
		log.Error().Err(err).
			Str("email", data.Email).
			Msg("failed to send collaborator invite email after retries")
		return err
	}

	log.Info().
		Str("email", data.Email).
		Str("pod_name", data.PodName).
		Msg("collaborator invite email sent successfully")
	return nil
}

func (w *EmailWorker) sendWithRetry(ctx context.Context, emailMsg *email.Email) error {
	var lastErr error

	for attempt := 0; attempt < w.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, ...
			delay := w.config.RetryBaseDelay * time.Duration(1<<uint(attempt-1))
			log.Info().
				Int("attempt", attempt+1).
				Dur("delay", delay).
				Str("to", emailMsg.To).
				Msg("retrying email send")

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := w.emailProvider.Send(ctx, emailMsg)
		if err == nil {
			return nil
		}

		lastErr = err
		log.Warn().Err(err).
			Int("attempt", attempt+1).
			Int("max_retries", w.config.MaxRetries).
			Str("to", emailMsg.To).
			Msg("email send attempt failed")
	}

	return fmt.Errorf("failed to send email after %d attempts: %w", w.config.MaxRetries, lastErr)
}
