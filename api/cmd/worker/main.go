// cmd/worker corre el job de sugerencias de IA (fase 9): por cada
// asignacion activa arma un resumen estructurado, le pide a Claude una
// sugerencia y la guarda como 'pending'. Pensado para correr por cron
// (una vez por noche) — no hay scheduler embebido, cada corrida procesa
// todas las asignaciones activas una vez y termina.
package main

import (
	"context"
	"log/slog"
	"os"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/jackc/pgx/v5/pgxpool"

	"myvibesfit/api/internal/adapter/anthropic"
	"myvibesfit/api/internal/adapter/postgres"
	"myvibesfit/api/internal/repository/db"
	"myvibesfit/api/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Error("db connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	queries := db.New(pool)
	client := anthropicsdk.NewClient()
	model := os.Getenv("ANTHROPIC_SUGGESTION_MODEL")
	proposer := anthropic.NewSuggester(&client, model)

	worker := service.NewSuggestionWorkerService(
		postgres.NewAISuggestionRepository(queries),
		postgres.NewGamificationRepository(queries),
		postgres.NewHabitRepository(queries),
		postgres.NewProgressRepository(queries),
		proposer,
		model,
	)

	assignments, err := worker.ListActiveAssignments(ctx)
	if err != nil {
		logger.Error("list active assignments failed", "error", err)
		os.Exit(1)
	}
	logger.Info("worker starting", "active_assignments", len(assignments))

	created, failed := 0, 0
	for _, a := range assignments {
		if a.CoachUserID == nil {
			continue // sin coach asignado, no hay quien apruebe la sugerencia
		}
		if err := worker.ProcessAssignment(ctx, a); err != nil {
			logger.Error("assignment failed", "assignment_id", a.AssignmentID, "error", err)
			failed++
			continue
		}
		created++
	}
	logger.Info("worker finished", "created", created, "failed", failed)
}
