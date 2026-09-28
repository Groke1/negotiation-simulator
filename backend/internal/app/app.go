package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/ai"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/ai/ollama"
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
	profileserv "sourcecraft.dev/lct-hackaton-2026.arena/internal/usecase/profile"
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

	var aiProvider ai.Provider
	switch strings.ToLower(strings.TrimSpace(cfg.AI.Provider)) {
	case "ollama":
		aiProvider = ollama.NewProvider(ollama.Config{BaseURL: cfg.AI.BaseURL, Timeout: cfg.AI.Timeout, Think: cfg.AI.Think, Stream: cfg.AI.Stream, Format: cfg.AI.Format, NumPredict: cfg.AI.NumPredict, NumCtx: cfg.AI.NumCtx})
	default:
		return fmt.Errorf("unsupported AI_PROVIDER %q", cfg.AI.Provider)
	}
	systemPrompts, err := loadAISystemPrompts(cfg.AI.PromptDir)
	if err != nil {
		logger.Error("can not load ai prompts", zap.Error(err))
		return err
	}
	aiService, err := ai.NewService(aiProvider, ai.Config{
		ScenarioArchitectTemp:      cfg.AI.ScenarioArchitectTemperature,
		BlockDesignerTemp:          cfg.AI.BlockDesignerTemperature,
		DialogueBuilderTemp:        cfg.AI.DialogueBuilderTemperature,
		ActionClassifierTemp:       cfg.AI.ActionClassifierTemperature,
		VirtualPartnerTemp:         cfg.AI.VirtualPartnerTemperature,
		VirtualPartnerRuntimeTemp:  cfg.AI.VirtualPartnerRuntimeTemperature,
		TrainingStateTemp:          cfg.AI.TrainingStateTemperature,
		TrainingAssessmentTemp:     cfg.AI.TrainingAssessmentTemperature,
		BlockDesignerThink:         boolPtr(cfg.AI.BlockDesignerThink),
		DialogueBuilderThink:       boolPtr(cfg.AI.DialogueBuilderThink),
		VirtualPartnerRuntimeThink: boolPtr(cfg.AI.VirtualPartnerRuntimeThink),
		TrainingStateThink:         boolPtr(cfg.AI.TrainingStateThink),
		TrainingAssessmentThink:    boolPtr(cfg.AI.TrainingAssessmentThink),
		MaxStructuredOutputRetry:   cfg.AI.MaxStructuredOutputRetry,
		SystemPrompts:              systemPrompts,
	})
	if err != nil {
		logger.Error("can not initialize ai service", zap.Error(err))
		return err
	}

	authService := authserv.NewAuthService(userRepository, authRepository, tx, authserv.Config{
		JWTSecret:       []byte(cfg.Settings.JWTSecret),
		AccessTokenTTL:  cfg.Settings.AccessTokenTTL,
		RefreshTokenTTL: cfg.Settings.RefreshTokenTTL,
	})
	organizationService := orgserv.NewOrganizationService(organizationRepository)
	userService := userserv.NewUserService(userRepository, organizationRepository, authRepository, authService)
	profileService := profileserv.NewService(profileRepository)
	if err := profileService.WarmCache(ctx); err != nil {
		logger.Error("can not warm profile dictionaries cache", zap.Error(err))
		return err
	}
	scenarioService := scenarioserv.NewScenarioService(scenarioRepository, profileService, userRepository, tx, aiService)
	mediaService := mediaserv.NewService(mediaRepository, cfg.Settings.MediaMaxBytes)
	trainingService := trainingserv.NewService(trainingRepository, scenarioRepository, profileService, tx, aiService)
	curatorService := curatorserv.NewService(curatorRepository, userRepository, scenarioRepository)
	arenaService := arenaserv.NewService(arenaRepository, userRepository, scenarioRepository, profileService, tx)

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

func loadAISystemPrompts(dir string) (map[ai.Role]string, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, nil
	}
	files := map[ai.Role]string{
		ai.RoleScenarioArchitect:     "scenario_architect.system.md",
		ai.RoleBlockDesigner:         "block_designer.system.md",
		ai.RoleDialogueBuilder:       "deterministic_dialogue_builder.system.md",
		ai.RoleVirtualPartnerRuntime: "virtual_partner_runtime.system.md",
		ai.RoleTrainingAssessment:    "training_assessment_evaluator.system.md",
	}
	prompts := make(map[ai.Role]string, len(files))
	for role, name := range files {
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		if prompt := strings.TrimSpace(string(content)); prompt != "" {
			prompts[role] = prompt
		}
	}
	return prompts, nil
}

func boolPtr(value bool) *bool {
	return &value
}
