package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"myvibesfit/api/internal/adapter/postgres"
	"myvibesfit/api/internal/config"
	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/platform"
	"myvibesfit/api/internal/repository/db"
	"myvibesfit/api/internal/service"
	apphttp "myvibesfit/api/internal/transport/http"
	"myvibesfit/api/internal/transport/http/handler"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config load failed", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("db ping failed", "error", err)
		os.Exit(1)
	}

	queries := db.New(pool)
	signer := platform.NewJWTSigner(cfg.JWTAccessSecret, cfg.JWTAccessTTL)

	identityRepo := postgres.NewIdentityRepository(queries)
	auditRepo := postgres.NewAuditRepository(queries)
	auditLogger := service.NewAuditLogger(auditRepo)
	exerciseRepo := postgres.NewExerciseRepository(queries)
	programRepo := postgres.NewProgramRepository(queries)
	assignmentRepo := postgres.NewAssignmentRepository(queries)
	gamificationRepo := postgres.NewGamificationRepository(queries)
	habitRepo := postgres.NewHabitRepository(queries)
	progressRepo := postgres.NewProgressRepository(queries)
	coachRepo := postgres.NewCoachRepository(queries)
	aiSuggestionRepo := postgres.NewAISuggestionRepository(queries)
	profileRepo := postgres.NewProfileRepository(queries)
	uow := postgres.NewUnitOfWork(pool)

	authSvc := service.NewAuthService(identityRepo, signer, cfg.JWTRefreshTTL)
	exerciseSvc := service.NewExerciseService(exerciseRepo)
	programSvc := service.NewProgramService(programRepo, auditLogger)
	assignmentSvc := service.NewAssignmentService(assignmentRepo, programRepo, uow, auditLogger)
	gamificationSvc := domain.NewGamificationService()
	syncSvc := service.NewSyncService(uow, gamificationSvc, assignmentRepo, programRepo)
	progressSvc := service.NewProgressService(progressRepo)
	habitSvc := service.NewHabitService(habitRepo, uow, gamificationSvc)
	coachSvc := service.NewCoachService(coachRepo)
	aiSuggestionSvc := service.NewAISuggestionService(aiSuggestionRepo, uow, auditLogger)
	profileSvc := service.NewProfileService(profileRepo)

	router := apphttp.NewRouter(apphttp.Handlers{
		Auth:         handler.NewAuthHandler(authSvc),
		Exercise:     handler.NewExerciseHandler(exerciseSvc),
		Program:      handler.NewProgramHandler(programSvc),
		Assignment:   handler.NewAssignmentHandler(assignmentSvc),
		Sync:         handler.NewSyncHandler(syncSvc),
		Progress:     handler.NewProgressHandler(progressSvc),
		Habit:        handler.NewHabitHandler(habitSvc),
		Gamification: handler.NewGamificationHandler(gamificationRepo, gamificationSvc),
		Coach:        handler.NewCoachHandler(coachSvc),
		AISuggestion: handler.NewAISuggestionHandler(aiSuggestionSvc),
		Profile:      handler.NewProfileHandler(profileSvc),
	}, signer, cfg.CORSAllowedOrigins)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: router}

	// Apagado ordenado: sin esto un SIGTERM del orquestador mata el proceso
	// con requests en vuelo y sin cerrar el pool.
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("myvibesfit api starting", "port", cfg.Port, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	case <-ctx.Done():
		logger.Info("shutdown signal received, draining")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
	logger.Info("myvibesfit api stopped")
}
