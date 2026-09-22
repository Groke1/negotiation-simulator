package controller

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/arena"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/auth"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/curator"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/media"
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
	media         *media.Controller
	training      *training.Controller
	curator       *curator.Controller
	arena         *arena.Controller
	verifier      ctrlmw.AccessTokenVerifier
}

func NewController(
	logger *zap.Logger,
	authService auth.Service,
	userService user.Service,
	organizationsService organizations.Service,
	scenarioService scenario.Service,
	mediaService media.Service,
	trainingService training.Service,
	curatorService curator.Service,
	arenaService arena.Service,
	verifier ctrlmw.AccessTokenVerifier,
	cookieCfg ctrlmw.CookieConfig,
	mediaMaxBytes int64,
) *controller {
	return &controller{
		auth:          auth.NewController(logger, authService, cookieCfg),
		user:          user.NewController(logger, userService),
		organizations: organizations.NewController(logger, organizationsService),
		scenario:      scenario.NewController(logger, scenarioService),
		media:         media.NewController(logger, mediaService, mediaMaxBytes),
		training:      training.NewController(logger, trainingService),
		curator:       curator.NewController(logger, curatorService),
		arena:         arena.NewController(logger, arenaService, verifier),
		verifier:      verifier,
	}
}

func (c *controller) InitRoutes(r chi.Router) {
	c.arena.InitWebSocketRoutes(r)
	r.Route("/api/v1", func(r chi.Router) {
		c.auth.InitRoutes(r)
		r.Group(func(r chi.Router) {
			r.Use(ctrlmw.RequireAccessToken(c.verifier))
			c.user.InitRoutes(r)
			c.organizations.InitRoutes(r)
			c.scenario.InitRoutes(r)
			c.media.InitRoutes(r)
			c.training.InitRoutes(r)
			c.curator.InitRoutes(r)
			c.arena.InitRoutes(r)
		})
	})
	r.Route("/internal/v1", func(r chi.Router) {
		c.user.InitInternalRoutes(r)
		c.organizations.InitInternalRoutes(r)
	})
}
