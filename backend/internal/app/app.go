package app

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/config"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller"
	ctrlmw "sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/middleware"
	arenarepo "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/arena"
	authrepo "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/auth"
	curatorrepo "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/curator"
	mediarepo "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/media"
	orgrepo "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/organizations"
	profilerepo "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/profile"
	scenariorepo "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/scenario"
	trainingrepo "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/training"
	userrepo "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/user"
	arenaserv "sourcecraft.dev/lct-hackaton-2026.arena/internal/usecase/arena"
	authserv "sourcecraft.dev/lct-hackaton-2026.arena/internal/usecase/auth"
	curatorserv "sourcecraft.dev/lct-hackaton-2026.arena/internal/usecase/curator"
	mediaserv "sourcecraft.dev/lct-hackaton-2026.arena/internal/usecase/media"
	orgserv "sourcecraft.dev/lct-hackaton-2026.arena/internal/usecase/organizations"
	scenarioserv "sourcecraft.dev/lct-hackaton-2026.arena/internal/usecase/scenario"
	trainingserv "sourcecraft.dev/lct-hackaton-2026.arena/internal/usecase/training"
	userserv "sourcecraft.dev/lct-hackaton-2026.arena/internal/usecase/user"
	"sourcecraft.dev/lct-hackaton-2026.arena/migrations"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/db"
	cors "sourcecraft.dev/lct-hackaton-2026.arena/pkg/middleware"
)

func Run(logger *zap.Logger, cfg *config.Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pgxcfg, err := pgxpool.ParseConfig(cfg.ConstructPostgresURL())
	if err != nil {
		logger.Error("can not create pgxpool cfg", zap.Error(err))
		return err
	}

	pgxcfg.MaxConns = cfg.PG.MaxConns
	pgxcfg.MinConns = cfg.PG.MinConns
	pgxcfg.HealthCheckPeriod = cfg.Settings.PGHealthCheckPeriod
	pgxcfg.MaxConnLifetime = cfg.Settings.PGMaxConnLifetime
	pgxcfg.MaxConnIdleTime = cfg.Settings.PGMaxConnIdleTime

	dbPool, err := pgxpool.NewWithConfig(ctx, pgxcfg)
	if err != nil {
		logger.Error("can not create pgxpool", zap.Error(err))
		return err
	}
	defer dbPool.Close()

	migrations.SetupPostgres(dbPool, logger)

	tx := db.NewTransactor(dbPool)
	authRepository := authrepo.NewAuthRepository(dbPool)
	userRepository := userrepo.NewUserRepository(dbPool)
	organizationRepository := orgrepo.NewOrganizationRepository(dbPool)
	scenarioRepository := scenariorepo.NewScenarioRepository(dbPool)
	profileRepository := profilerepo.NewRepository(dbPool)
	mediaRepository := mediarepo.NewRepository(dbPool)
	trainingRepository := trainingrepo.NewRepository(dbPool)
	curatorRepository := curatorrepo.NewRepository(dbPool)
	arenaRepository := arenarepo.NewRepository(dbPool)

	authService := authserv.NewAuthService(userRepository, authRepository, tx, authserv.Config{
		JWTSecret:       []byte(cfg.Settings.JWTSecret),
		AccessTokenTTL:  cfg.Settings.AccessTokenTTL,
		RefreshTokenTTL: cfg.Settings.RefreshTokenTTL,
	})
	organizationService := orgserv.NewOrganizationService(organizationRepository)
	userService := userserv.NewUserService(userRepository, organizationRepository, authRepository, authService)
	scenarioService := scenarioserv.NewScenarioService(scenarioRepository, profileRepository, userRepository, tx)
	mediaService := mediaserv.NewService(mediaRepository, cfg.Settings.MediaMaxBytes)
	trainingService := trainingserv.NewService(trainingRepository, scenarioRepository, profileRepository, tx)
	curatorService := curatorserv.NewService(curatorRepository, userRepository, scenarioRepository)
	arenaService := arenaserv.NewService(arenaRepository, userRepository, scenarioRepository, profileRepository, tx)

	router := chi.NewRouter()
	router.Use(middleware.StripSlashes, middleware.Logger, middleware.Recoverer)

	contr := controller.NewController(
		logger,
		authService,
		userService,
		organizationService,
		scenarioService,
		mediaService,
		trainingService,
		curatorService,
		arenaService,
		authService,
		ctrlmw.CookieConfig{RefreshTokenTTL: cfg.Settings.RefreshTokenTTL},
		cfg.Settings.MediaMaxBytes,
	)
	contr.InitRoutes(router)

	server := &http.Server{
		Addr:              ":" + cfg.HTTP.Port,
		Handler:           cors.CorsHandler(router),
		ReadHeaderTimeout: cfg.Settings.ServerReadHeaderTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("starting arena service", zap.String("addr", server.Addr))
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server failed", zap.Error(err))
			return err
		}
	case <-ctx.Done():
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Settings.ShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("graceful shutdown failed", zap.Error(err))
		return err
	}

	return nil
}
