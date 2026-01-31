package email

import (
	"bytes"
	"context"
	"html/template"
)

type Email struct {
	To       string
	Subject  string
	HTMLBody string
	TextBody string
}

type Provider interface {
	Send(ctx context.Context, email *Email) error
}

type TemplateData struct {
	RecipientName string
	ActionURL     string
	AppName       string
	SupportEmail  string
	ExpiryHours   int
	// Additional
	PodName      string
	InviterName  string
	MaterialName string
}

type TemplateRenderer struct {
	templates map[string]*template.Template
}

func NewTemplateRenderer() *TemplateRenderer {
	r := &TemplateRenderer{
		templates: make(map[string]*template.Template),
	}
	r.loadTemplates()
	return r
}

func (r *TemplateRenderer) loadTemplates() {
	r.templates["verification"] = template.Must(template.New("verification").Parse(verificationTemplate))
	r.templates["password_reset"] = template.Must(template.New("password_reset").Parse(passwordResetTemplate))
	r.templates["collaborator_invite"] = template.Must(template.New("collaborator_invite").Parse(collaboratorInviteTemplate))
}

func (r *TemplateRenderer) Render(templateName string, data TemplateData) (html string, text string, err error) {
	tmpl, ok := r.templates[templateName]
	if !ok {
		return "", "", ErrTemplateNotFound
	}

	var htmlBuf bytes.Buffer
	if err := tmpl.Execute(&htmlBuf, data); err != nil {
		return "", "", err
	}

	textContent := generatePlainText(templateName, data)

	return htmlBuf.String(), textContent, nil
}

func generatePlainText(templateName string, data TemplateData) string {
	switch templateName {
	case "verification":
		return "Hi " + data.RecipientName + ",\n\n" +
			"Please verify your email address by clicking the link below:\n\n" +
			data.ActionURL + "\n\n" +
			"This link will expire in " + string(rune(data.ExpiryHours)) + " hours.\n\n" +
			"If you didn't create an account, please ignore this email.\n\n" +
			"Best regards,\n" + data.AppName + " Team"
	case "password_reset":
		return "Hi " + data.RecipientName + ",\n\n" +
			"You requested to reset your password. Click the link below:\n\n" +
			data.ActionURL + "\n\n" +
			"This link will expire in " + string(rune(data.ExpiryHours)) + " hours.\n\n" +
			"If you didn't request this, please ignore this email.\n\n" +
			"Best regards,\n" + data.AppName + " Team"
	case "collaborator_invite":
		return "Hi " + data.RecipientName + ",\n\n" +
			data.InviterName + " has invited you to collaborate on \"" + data.PodName + "\".\n\n" +
			"Click the link below to accept the invitation:\n\n" +
			data.ActionURL + "\n\n" +
			"Best regards,\n" + data.AppName + " Team"
	default:
		return ""
	}
}
