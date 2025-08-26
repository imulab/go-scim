package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/imulab/go-scim/cmd/internal/args"
)

// AuthMiddleware wraps an HTTP handler with authentication
func AuthMiddleware(authConfig *args.Auth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for health check and metadata endpoints
			if isPublicEndpoint(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Authenticate the request
			if err := authenticateRequest(r, authConfig); err != nil {
				writeUnauthorizedError(w, fmt.Sprintf("Authentication failed: %v", err))
				return
			}

			// Request is authenticated, continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// isPublicEndpoint returns true for endpoints that don't require authentication
func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/health",
		"/ServiceProviderConfig",
		"/Schemas",
		"/ResourceTypes",
	}

	for _, publicPath := range publicPaths {
		if path == publicPath || strings.HasPrefix(path, publicPath+"/") {
			return true
		}
	}
	return false
}

// authenticateRequest validates the request authentication
func authenticateRequest(r *http.Request, authConfig *args.Auth) error {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return fmt.Errorf("missing Authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return fmt.Errorf("authorization header must be Bearer token")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return fmt.Errorf("bearer token is empty")
	}

	// Try OAuth2 validation first if enabled
	if authConfig.OAuth2Enabled {
		return validateOAuth2Token(r.Context(), token, authConfig)
	}

	// Fall back to bearer token validation
	return validateBearerToken(token, authConfig.GetBearerTokens())
}

// validateBearerToken validates a simple bearer token
func validateBearerToken(token string, validTokens []string) error {
	for _, validToken := range validTokens {
		if token == validToken {
			return nil
		}
	}
	return fmt.Errorf("invalid bearer token")
}

// validateOAuth2Token validates an OAuth2 token (simplified for mock server)
func validateOAuth2Token(_ context.Context, token string, authConfig *args.Auth) error {
	// Check cache first for performance
	if authConfig.TokenCache != nil {
		if expiry, exists := authConfig.TokenCache[token]; exists && time.Now().Before(expiry) {
			return nil
		}
	}

	// Basic token format validation (sufficient for mock/testing)
	if len(token) < 10 {
		return fmt.Errorf("token too short")
	}

	// For mock server, accept tokens that look like valid JWTs or have basic format
	if !strings.Contains(token, ".") && len(token) < 20 {
		return fmt.Errorf("invalid token format")
	}

	// Cache the token for 1 hour
	if authConfig.TokenCache == nil {
		authConfig.TokenCache = make(map[string]time.Time)
	}
	authConfig.TokenCache[token] = time.Now().Add(1 * time.Hour)

	return nil
}

// writeUnauthorizedError writes a SCIM-compliant error response
func writeUnauthorizedError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/scim+json")
	w.WriteHeader(http.StatusUnauthorized)

	errorResponse := map[string]interface{}{
		"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		"detail":  message,
		"status":  "401",
	}

	json.NewEncoder(w).Encode(errorResponse)
}
