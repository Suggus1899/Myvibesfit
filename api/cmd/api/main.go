package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"myvibesfit/api/internal/config"
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

	authSvc := service.NewAuthService(queries, signer, cfg.JWTRefreshTTL)
	exerciseSvc := service.NewExerciseService(queries)
	programSvc := service.NewProgramService(queries)
	assignmentSvc := service.NewAssignmentService(queries, pool)
	gamificationSvc := service.NewGamificationService()
	syncSvc := service.NewSyncService(queries, pool, gamificationSvc)
	progressSvc := service.NewProgressService(queries)
	habitSvc := service.NewHabitService(queries, pool, gamificationSvc)

	router := apphttp.NewRouter(apphttp.Handlers{
		Auth:         handler.NewAuthHandler(authSvc),
		Exercise:     handler.NewExerciseHandler(exerciseSvc),
		Program:      handler.NewProgramHandler(programSvc),
		Assignment:   handler.NewAssignmentHandler(assignmentSvc),
		Sync:         handler.NewSyncHandler(syncSvc),
		Progress:     handler.NewProgressHandler(progressSvc),
		Habit:        handler.NewHabitHandler(habitSvc),
		Gamification: handler.NewGamificationHandler(queries, gamificationSvc),
	}, signer)

	logger.Info("myvibesfit api starting", "port", cfg.Port, "env", cfg.Env)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
