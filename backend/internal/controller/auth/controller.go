package auth

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/httpx"
	cookiemw "sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/middleware"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	pkghttpx "sourcecraft.dev/lct-hackaton-2026.arena/pkg/httpx"
)

type Service interface {
	Register(ctx context.Context, params entity.RegisterUser) (*entity.JWT, error)
	Login(ctx context.Context, params entity.LoginUser) (*entity.JWT, error)
	Refresh(ctx context.Context, refreshToken string) (*entity.JWT, error)
	Logout(ctx context.Context, refreshToken string) error
}

type Controller struct {
	logger    *zap.Logger
	service   Service
	cookieCfg cookiemw.CookieConfig
}

func NewController(logger *zap.Logger, service Service, cookieCfg cookiemw.CookieConfig) *Controller {
	return &Controller{logger: logger, service: service, cookieCfg: cookieCfg}
}

func (c *Controller) InitRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", c.Register)
		r.Post("/login", c.Login)
		r.Post("/refresh", c.Refresh)
		r.Post("/logout", c.Logout)
	})
}

func (c *Controller) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || !validRegisterRequest(req) {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	tokens, err := c.service.Register(r.Context(), entity.RegisterUser{Username: normalizeUsername(req.Username), Email: strings.TrimSpace(req.Email), Password: req.Password})
	if err != nil {
		c.writeAuthError(w, r, err)
		return
	}
	cookiemw.SetAuthCookies(w, r, tokens, c.cookieCfg)
	pkghttpx.WriteJSON(w, http.StatusCreated, accessTokenResponse{AccessToken: tokens.AccessToken})
}

func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || !validLoginRequest(req) {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	tokens, err := c.service.Login(r.Context(), entity.LoginUser{Email: strings.TrimSpace(req.Email), Password: req.Password})
	if err != nil {
		c.writeAuthError(w, r, err)
		return
	}
	cookiemw.SetAuthCookies(w, r, tokens, c.cookieCfg)
	pkghttpx.WriteJSON(w, http.StatusOK, accessTokenResponse{AccessToken: tokens.AccessToken})
}

func (c *Controller) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := readRefreshCookie(r)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrRefreshTokenMissing)
		return
	}
	tokens, err := c.service.Refresh(r.Context(), refreshToken)
	if err != nil {
		c.writeAuthError(w, r, err)
		return
	}
	cookiemw.SetAuthCookies(w, r, tokens, c.cookieCfg)
	pkghttpx.WriteJSON(w, http.StatusOK, accessTokenResponse{AccessToken: tokens.AccessToken})
}

func (c *Controller) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := readRefreshCookie(r)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrRefreshTokenMissing)
		return
	}
	if err := c.service.Logout(r.Context(), refreshToken); err != nil {
		c.writeAuthError(w, r, err)
		return
	}
	cookiemw.ClearAuthCookies(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) writeAuthError(w http.ResponseWriter, r *http.Request, err error) {
	code := httpx.ErrorCodeFromError(err)
	status := http.StatusInternalServerError
	switch code {
	case httpx.ErrUsernameAlreadyExists, httpx.ErrEmailAlreadyExists:
		status = http.StatusConflict
	case httpx.ErrInvalidCredentials, httpx.ErrRefreshTokenMissing, httpx.ErrRefreshTokenInvalid, httpx.ErrRefreshTokenExpired, httpx.ErrRefreshTokenRevoked, httpx.ErrRefreshTokenReused:
		status = http.StatusUnauthorized
	case httpx.ErrInvalidRequest:
		status = http.StatusBadRequest
	}
	if status == http.StatusInternalServerError {
		c.logger.Error("auth request failed", zap.Error(err), zap.String("path", r.URL.Path), zap.String("code", string(code)))
	}
	httpx.WriteError(w, status, code)
}

var usernamePattern = regexp.MustCompile(`^[\p{L}\p{N}_.\- ]+$`)

func validRegisterRequest(req registerRequest) bool {
	username := normalizeUsername(req.Username)
	email := strings.TrimSpace(req.Email)
	usernameLen := utf8.RuneCountInString(username)
	return usernameLen >= 3 && usernameLen <= 50 && usernamePattern.MatchString(username) && validEmail(email) && len(req.Password) >= 8 && len(req.Password) <= 128
}

func normalizeUsername(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func validLoginRequest(req loginRequest) bool {
	email := strings.TrimSpace(req.Email)
	return validEmail(email) && len(req.Password) >= 1 && len(req.Password) <= 128
}

func validEmail(value string) bool {
	if value == "" || len(value) > 254 {
		return false
	}
	addr, err := mail.ParseAddress(value)
	return err == nil && addr.Address == value
}

func readRefreshCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookiemw.RefreshTokenCookieName)
	if errors.Is(err, http.ErrNoCookie) || cookie == nil || strings.TrimSpace(cookie.Value) == "" {
		return "", entity.ErrRefreshTokenMissing
	}
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
