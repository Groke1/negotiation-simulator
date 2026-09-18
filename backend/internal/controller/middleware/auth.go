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
	return requireAccessToken(verifier, func(r *http.Request) (string, bool) {
		return bearerToken(r.Header.Get("Authorization"))
	})
}

// RequireWebSocketAccessToken reads the JWT from a WebSocket subprotocol.
// Browsers cannot set an Authorization header on the WebSocket handshake.
func RequireWebSocketAccessToken(verifier AccessTokenVerifier) func(http.Handler) http.Handler {
	return requireAccessToken(verifier, websocketToken)
}

func requireAccessToken(verifier AccessTokenVerifier, tokenFromRequest func(*http.Request) (string, bool)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := tokenFromRequest(r)
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

func websocketToken(r *http.Request) (string, bool) {
	for _, protocol := range strings.Split(r.Header.Get("Sec-WebSocket-Protocol"), ",") {
		token := strings.TrimSpace(protocol)
		if strings.Count(token, ".") == 2 {
			return token, true
		}
	}
	return "", false
}
