package email

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type SESProvider struct {
	client    *ses.Client
	fromEmail string
	fromName  string
}

func NewSESProvider(ctx context.Context, region, accessKey, secretKey, fromEmail, fromName string) (*SESProvider, error) {
	if region == "" {
		return nil, fmt.Errorf("%w: AWS SES region is required", ErrInvalidConfig)
	}
	if fromEmail == "" {
		return nil, fmt.Errorf("%w: AWS SES email is required", ErrInvalidConfig)
	}

	var cfg aws.Config
	var err error

	if accessKey != "" && secretKey != "" {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("%w: failed to load AWS config: %v", ErrInvalidConfig, err)
	}

	return &SESProvider{
		client:    ses.NewFromConfig(cfg),
		fromEmail: fromEmail,
		fromName:  fromName,
	}, nil
}

func (p *SESProvider) Send(ctx context.Context, email *Email) error {
	fromAddress := p.fromEmail
	if p.fromName != "" {
		fromAddress = fmt.Sprintf("%s <%s>", fromAddress, p.fromName)
	}

	input := &ses.SendEmailInput{
		Source: aws.String(fromAddress),
		Destination: &types.Destination{
			ToAddresses: []string{email.To},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(email.Subject),
			},
			Body: &types.Body{},
		},
	}

	if email.HTMLBody != "" {
		input.Message.Body.Html = &types.Content{
			Charset: aws.String("UTF-8"),
			Data:    aws.String(email.HTMLBody),
		}
	}

	if email.TextBody != "" {
		input.Message.Body.Text = &types.Content{
			Charset: aws.String("UTF-8"),
			Data:    aws.String(email.TextBody),
		}
	}

	_, err := p.client.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	return nil
}
