// Package oauth provides OAuth client implementations for external authentication providers.
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleUserInfo represents the user information returned by Google's userinfo API.
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// GoogleClient handles Google OAuth authentication.
type GoogleClient struct {
	config *oauth2.Config
}

// GoogleConfig holds the configuration for Google OAuth.
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// NewGoogleClient creates a new Google OAuth client.
func NewGoogleClient(cfg GoogleConfig) *GoogleClient {
	return &GoogleClient{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

// GetAuthURL returns the URL to redirect users to for Google OAuth consent.
func (c *GoogleClient) GetAuthURL(state string) string {
	return c.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges an authorization code for tokens and returns user info.
// The redirectURI parameter allows overriding the redirect URL for the token exchange,
// which is useful when the frontend uses a different callback URL.
func (c *GoogleClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*GoogleUserInfo, error) {
	// Create a copy of the config with the provided redirect URI if specified
	config := c.config
	if redirectURI != "" {
		configCopy := *c.config
		configCopy.RedirectURL = redirectURI
		config = &configCopy
	}

	// Exchange authorization code for tokens
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from Google
	userInfo, err := c.getUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return userInfo, nil
}

// getUserInfo fetches user information from Google's userinfo API.
func (c *GoogleClient) getUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := c.config.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google API error: status=%d body=%s", resp.StatusCode, string(body))
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// ValidateConfig checks if the Google OAuth configuration is valid.
func (c *GoogleClient) ValidateConfig() error {
	if c.config.ClientID == "" {
		return fmt.Errorf("google client ID is required")
	}
	if c.config.ClientSecret == "" {
		return fmt.Errorf("google client secret is required")
	}
	return nil
}
