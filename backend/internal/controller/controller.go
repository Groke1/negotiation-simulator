package controller

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/auth"
	ctrlmw "sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/middleware"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/organizations"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/user"
)

type controller struct {
	auth          *auth.Controller
	user          *user.Controller
	organizations *organizations.Controller
	verifier      ctrlmw.AccessTokenVerifier
}

func NewController(
	logger *zap.Logger,
	authService auth.Service,
	userService user.Service,
	organizationsService organizations.Service,
	verifier ctrlmw.AccessTokenVerifier,
	cookieCfg ctrlmw.CookieConfig,
) *controller {
	return &controller{
		auth:          auth.NewController(logger, authService, cookieCfg),
		user:          user.NewController(logger, userService),
		organizations: organizations.NewController(logger, organizationsService),
		verifier:      verifier,
	}
}

func (c *controller) InitRoutes(r chi.Router) {
	r.Route("/api/v1", func(r chi.Router) {
		c.auth.InitRoutes(r)
		r.Group(func(r chi.Router) {
			r.Use(ctrlmw.RequireAccessToken(c.verifier))
			c.user.InitRoutes(r)
			c.organizations.InitRoutes(r)
		})
	})
	r.Route("/internal/v1", func(r chi.Router) {
		c.user.InitInternalRoutes(r)
		c.organizations.InitInternalRoutes(r)
	})
}
