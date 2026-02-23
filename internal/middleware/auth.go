package middleware

import (
	"context"
	"net/http"

	"github.com/temple-ats/TempleATS/internal/auth"
)

type contextKey string

const (
	UserIDKey  contextKey = "userId"
	OrgIDKey   contextKey = "orgId"
	OrgSlugKey contextKey = "orgSlug"
	RoleKey    contextKey = "role"
)

// RequireAuth is a middleware that validates the JWT from the "token" cookie
// and injects user claims into the request context.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateToken(cookie.Value)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, OrgIDKey, claims.OrgID)
		ctx = context.WithValue(ctx, OrgSlugKey, claims.OrgSlug)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts the user ID from the request context.
func GetUserID(ctx context.Context) string {
	v, _ := ctx.Value(UserIDKey).(string)
	return v
}

// GetOrgID extracts the organization ID from the request context.
func GetOrgID(ctx context.Context) string {
	v, _ := ctx.Value(OrgIDKey).(string)
	return v
}

// GetOrgSlug extracts the organization slug from the request context.
func GetOrgSlug(ctx context.Context) string {
	v, _ := ctx.Value(OrgSlugKey).(string)
	return v
}

// GetRole extracts the user role from the request context.
func GetRole(ctx context.Context) string {
	v, _ := ctx.Value(RoleKey).(string)
	return v
}
