package middleware

import (
	"context"
	"net/http"
	"strings"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/httpx"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type AccessTokenVerifier interface {
	VerifyAccessToken(ctx context.Context, token string) (*entity.AccessClaims, error)
}

type claimsContextKey struct{}

func RequireAccessToken(verifier AccessTokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrAccessTokenMissing)
				return
			}
			claims, err := verifier.VerifyAccessToken(r.Context(), token)
			if err != nil {
				httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeFromError(err))
				return
			}
			ctx := context.WithValue(r.Context(), claimsContextKey{}, *claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClaimsFromContext(ctx context.Context) (entity.AccessClaims, bool) {
	claims, ok := ctx.Value(claimsContextKey{}).(entity.AccessClaims)
	return claims, ok
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}
