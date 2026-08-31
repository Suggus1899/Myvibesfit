package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"myvibesfit/api/internal/platform"
	"myvibesfit/api/internal/transport/http/handler"
	appmw "myvibesfit/api/internal/transport/http/middleware"
)

type Handlers struct {
	Auth         *handler.AuthHandler
	Exercise     *handler.ExerciseHandler
	Program      *handler.ProgramHandler
	Assignment   *handler.AssignmentHandler
	Sync         *handler.SyncHandler
	Progress     *handler.ProgressHandler
	Habit        *handler.HabitHandler
	Gamification *handler.GamificationHandler
	Coach        *handler.CoachHandler
	AISuggestion *handler.AISuggestionHandler
	Profile      *handler.ProfileHandler
	DeviceToken  *handler.DeviceTokenHandler
}

func NewRouter(h Handlers, signer *platform.JWTSigner, corsOrigins []string) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(httprate.LimitByIP(300, time.Minute))
	if len(corsOrigins) > 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   corsOrigins,
			AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Authorization", "Content-Type"},
			AllowCredentials: false,
			MaxAge:           300,
		}))
	}

	r.Get("/health", handler.Health)

	r.Route("/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(httprate.LimitByIP(10, time.Minute))
			r.Post("/auth/register", h.Auth.Register)
			r.Post("/auth/login", h.Auth.Login)
			r.Post("/auth/refresh", h.Auth.Refresh)
		})

		r.Group(func(r chi.Router) {
			r.Use(appmw.OptionalAuth(signer))
			r.Get("/exercises", h.Exercise.List)
			r.Get("/exercises/{id}", h.Exercise.Get)
		})

		r.Group(func(r chi.Router) {
			r.Use(appmw.Auth(signer))

			r.Get("/me", h.Auth.Me)
			r.Get("/me/profile", h.Profile.Get)
			r.Put("/me/profile", h.Profile.Save)
			r.Post("/me/device-tokens", h.DeviceToken.Register)
			r.Delete("/me/device-tokens", h.DeviceToken.Unregister)
			r.Post("/orgs", h.Auth.CreateOrg)
			r.Post("/orgs/join", h.Auth.JoinOrg)
			r.Get("/me/assignment/current", h.Assignment.CurrentForMe)
			r.Get("/me/stats", h.Gamification.Stats)
			r.Get("/me/achievements", h.Gamification.Achievements)

			r.Post("/sync/sessions", h.Sync.Sync)

			r.Get("/progress/exercises/{id}", h.Progress.ExerciseHistory)
			r.Get("/progress/records", h.Progress.Records)
			r.Get("/progress/volume", h.Progress.Volume)

			r.Post("/body-metrics", h.Progress.LogBodyMetric)
			r.Get("/body-metrics", h.Progress.ListBodyMetrics)

			r.Get("/habits", h.Habit.List)
			r.Get("/habits/logs", h.Habit.LogsForDate)
			r.Get("/me/habits", h.Habit.MyHabits)
			r.Post("/habits/subscribe", h.Habit.Subscribe)
			r.Delete("/habits/{id}", h.Habit.Unsubscribe)
			r.Post("/habits/{id}/log", h.Habit.Log)

			r.Group(func(r chi.Router) {
				r.Use(appmw.RequireRole("owner", "admin", "coach"))
				r.Post("/exercises", h.Exercise.Create)

				r.Group(func(r chi.Router) {
					r.Use(appmw.RequireOrg)
					r.Patch("/exercises/{id}", h.Exercise.Update)
					r.Delete("/exercises/{id}", h.Exercise.Delete)
				})
			})

			r.Group(func(r chi.Router) {
				r.Use(appmw.RequireRole("owner", "admin", "coach"))
				r.Use(appmw.RequireOrg)

				r.Get("/org", h.Auth.GetOrg)
				r.Get("/org/members", h.Auth.ListMembers)
				r.Group(func(r chi.Router) {
					r.Use(appmw.RequireRole("owner"))
					r.Patch("/org/members/{id}/role", h.Auth.UpdateMemberRole)
				})

				r.Get("/progression-rules", h.Program.ListProgressionRules)
				r.Post("/progression-rules", h.Program.CreateProgressionRule)

				r.Post("/programs", h.Program.Create)
				r.Get("/programs", h.Program.List)
				r.Get("/programs/{id}", h.Program.Get)
				r.Patch("/programs/{id}", h.Program.Update)
				r.Post("/programs/{id}/publish", h.Program.Publish)
				r.Post("/programs/{id}/archive", h.Program.Archive)
				r.Post("/programs/{id}/assign", h.Assignment.Assign)

				r.Post("/programs/{id}/workouts", h.Program.CreateWorkout)
				r.Get("/programs/{id}/workouts", h.Program.ListWorkouts)
				r.Patch("/workouts/{id}", h.Program.UpdateWorkout)
				r.Delete("/workouts/{id}", h.Program.DeleteWorkout)

				r.Post("/workouts/{id}/exercises", h.Program.CreateExercise)
				r.Get("/workouts/{id}/exercises", h.Program.ListExercises)
				r.Patch("/program-exercises/{id}", h.Program.UpdateExercise)
				r.Delete("/program-exercises/{id}", h.Program.DeleteExercise)

				r.Post("/assignments/{id}/cancel", h.Assignment.Cancel)

				r.Get("/coach/clients", h.Coach.Clients)
				r.Get("/coach/clients/{id}/progress", h.Coach.ClientProgress)

				r.Get("/coach/suggestions", h.AISuggestion.ListPending)
				r.Post("/ai-suggestions/{id}/approve", h.AISuggestion.Approve)
				r.Post("/ai-suggestions/{id}/reject", h.AISuggestion.Reject)
			})
		})
	})

	return r
}
