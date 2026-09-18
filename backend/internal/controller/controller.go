package controller

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/arena"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/auth"
	ctrlmw "sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/middleware"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/organizations"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/scenario"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/training"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/user"
)

type controller struct {
	auth          *auth.Controller
	user          *user.Controller
	organizations *organizations.Controller
	scenario      *scenario.Controller
	training      *training.Controller
	arena         *arena.Controller
	verifier      ctrlmw.AccessTokenVerifier
}

func NewController(
	logger *zap.Logger,
	authService auth.Service,
	userService user.Service,
	organizationsService organizations.Service,
	scenarioService scenario.Service,
	trainingService training.Service,
	arenaService arena.Service,
	verifier ctrlmw.AccessTokenVerifier,
	cookieCfg ctrlmw.CookieConfig,
) *controller {
	return &controller{
		auth:          auth.NewController(logger, authService, cookieCfg),
		user:          user.NewController(logger, userService),
		organizations: organizations.NewController(logger, organizationsService),
		scenario:      scenario.NewController(logger, scenarioService),
		training:      training.NewController(logger, trainingService),
		arena:         arena.NewController(logger, arenaService),
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
			c.scenario.InitRoutes(r)
			c.training.InitRoutes(r)
			c.arena.InitRoutes(r)
		})
	})
	r.Group(func(r chi.Router) {
		r.Use(ctrlmw.RequireWebSocketAccessToken(c.verifier))
		c.arena.InitWebsocketRoutes(r)
	})
	r.Route("/internal/v1", func(r chi.Router) {
		c.user.InitInternalRoutes(r)
		c.organizations.InitInternalRoutes(r)
	})
}
