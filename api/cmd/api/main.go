package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

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

	ctx := context.Background()
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
	uow := postgres.NewUnitOfWork(pool)

	authSvc := service.NewAuthService(identityRepo, signer, cfg.JWTRefreshTTL)
	exerciseSvc := service.NewExerciseService(exerciseRepo)
	programSvc := service.NewProgramService(programRepo, auditLogger)
	assignmentSvc := service.NewAssignmentService(assignmentRepo, programRepo, uow, auditLogger)
	gamificationSvc := domain.NewGamificationService()
	syncSvc := service.NewSyncService(uow, gamificationSvc)
	progressSvc := service.NewProgressService(progressRepo)
	habitSvc := service.NewHabitService(habitRepo, uow, gamificationSvc)
	coachSvc := service.NewCoachService(coachRepo)
	aiSuggestionSvc := service.NewAISuggestionService(aiSuggestionRepo, auditLogger)

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
	}, signer, cfg.CORSAllowedOrigins)

	logger.Info("myvibesfit api starting", "port", cfg.Port, "env", cfg.Env)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
