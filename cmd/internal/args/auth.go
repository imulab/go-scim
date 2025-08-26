package args

import (
	"strings"

	"github.com/urfave/cli/v2"
)

// Auth the configuration options related to authentication.
type Auth struct {
	// Enable OAuth2 authentication
	OAuth2Enabled bool
	// OAuth2 client ID
	OAuth2ClientID string
	// OAuth2 client secret
	OAuth2ClientSecret string
	// OAuth2 token endpoint URL
	OAuth2TokenURL string
	// Bearer tokens for simple authentication (comma-separated)
	BearerTokens string
	// Enable authentication (either OAuth2 or Bearer token)
	AuthEnabled bool
}

// GetBearerTokens returns a slice of valid bearer tokens
func (arg *Auth) GetBearerTokens() []string {
	if arg.BearerTokens == "" {
		return nil
	}
	tokens := strings.Split(arg.BearerTokens, ",")
	var validTokens []string
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token != "" {
			validTokens = append(validTokens, token)
		}
	}
	return validTokens
}

// IsAuthenticationEnabled returns true if any authentication method is enabled
func (arg *Auth) IsAuthenticationEnabled() bool {
	return arg.AuthEnabled || arg.OAuth2Enabled || arg.BearerTokens != ""
}

// Flags returns the CLI flags for authentication configuration
func (arg *Auth) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "auth-enabled",
			Usage:       "Enable authentication for SCIM endpoints",
			EnvVars:     []string{"AUTH_ENABLED"},
			Value:       false,
			Destination: &arg.AuthEnabled,
		},
		&cli.BoolFlag{
			Name:        "auth-oauth2-enabled",
			Usage:       "Enable OAuth2 client credentials authentication",
			EnvVars:     []string{"OAUTH2_ENABLED"},
			Value:       false,
			Destination: &arg.OAuth2Enabled,
		},
		&cli.StringFlag{
			Name:        "auth-oauth2-client-id",
			Usage:       "OAuth2 client ID for authentication",
			EnvVars:     []string{"OAUTH2_CLIENT_ID"},
			Value:       "",
			Destination: &arg.OAuth2ClientID,
		},
		&cli.StringFlag{
			Name:        "auth-oauth2-client-secret",
			Usage:       "OAuth2 client secret for authentication",
			EnvVars:     []string{"OAUTH2_CLIENT_SECRET"},
			Value:       "",
			Destination: &arg.OAuth2ClientSecret,
		},
		&cli.StringFlag{
			Name:        "auth-oauth2-token-url",
			Usage:       "OAuth2 token endpoint URL for authentication",
			EnvVars:     []string{"OAUTH2_TOKEN_URL"},
			Value:       "",
			Destination: &arg.OAuth2TokenURL,
		},
		&cli.StringFlag{
			Name:        "auth-bearer-tokens",
			Usage:       "Valid bearer tokens for authentication (comma-separated)",
			EnvVars:     []string{"BEARER_TOKENS"},
			Value:       "",
			Destination: &arg.BearerTokens,
		},
	}
}
