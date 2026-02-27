package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCConfig holds the OIDC provider and OAuth2 configuration for SSO.
type OIDCConfig struct {
	Provider *oidc.Provider
	OAuth2   oauth2.Config
	Verifier *oidc.IDTokenVerifier
}

// NewOIDCConfig initializes OIDC from environment variables.
// Returns nil if SSO_ISSUER is not set (SSO disabled).
func NewOIDCConfig(ctx context.Context) (*OIDCConfig, error) {
	issuer := os.Getenv("SSO_ISSUER")
	if issuer == "" {
		return nil, nil
	}

	clientID := os.Getenv("SSO_CLIENT_ID")
	clientSecret := os.Getenv("SSO_CLIENT_SECRET")
	redirectURI := os.Getenv("SSO_REDIRECT_URI")

	if clientID == "" || clientSecret == "" || redirectURI == "" {
		return nil, fmt.Errorf("SSO_CLIENT_ID, SSO_CLIENT_SECRET, and SSO_REDIRECT_URI are required when SSO_ISSUER is set")
	}

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc provider: %w", err)
	}

	return &OIDCConfig{
		Provider: provider,
		OAuth2: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURI,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
		},
		Verifier: provider.Verifier(&oidc.Config{ClientID: clientID}),
	}, nil
}
