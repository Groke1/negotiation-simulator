package middleware

import (
	"context"
	"net/http"
	"strings"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/httpx"
)

type AccessTokenValidator interface {
	ValidateAccessToken(ctx context.Context, token string) (userID string, err error)
}

type contextKey struct{}

func RequireAccessToken(validator AccessTokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrAccessTokenMissing)
				return
			}
			userID, err := validator.ValidateAccessToken(r.Context(), token)
			if err != nil || userID == "" {
				httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrAccessTokenInvalid)
				return
			}
			ctx := context.WithValue(r.Context(), contextKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(contextKey{}).(string)
	return userID, ok && userID != ""
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}
