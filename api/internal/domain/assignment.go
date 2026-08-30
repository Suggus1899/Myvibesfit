package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Assignment struct {
	ID           uuid.UUID
	OrgID        uuid.UUID
	ProgramID    *uuid.UUID
	ClientUserID uuid.UUID
	CoachUserID  *uuid.UUID
	Name         string
	StartDate    time.Time
	EndDate      *time.Time
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type AssignedWorkout struct {
	ID              uuid.UUID
	AssignmentID    uuid.UUID
	SourceWorkoutID *uuid.UUID
	WeekNumber      int16
	DayIndex        int16
	Name            string
	Note            string
	IsDeload        bool
	ScheduledOn     *time.Time
	Status          string
}

type AssignedExercise struct {
	ID                uuid.UUID
	AssignedWorkoutID uuid.UUID
	ExerciseID        uuid.UUID
	OrderIndex        int16
	SupersetGroup     *int
	TargetSets        int16
	TargetRepsMin     *int
	TargetRepsMax     *int
	TargetRPE         *float64
	TargetWeightKg    *float64
	RestSeconds       int16
	Tempo             string
	Note              string
	ProgressionRuleID *uuid.UUID

	// OverrideSource marca que este target lo escribio algo que no es el
	// motor de progresion (hoy solo "ai_suggestion"). El motor lo respeta una
	// sola vez y limpia el flag — ver OverrideSourceAISuggestion.
	OverrideSource *string
}

// OverrideSourceAISuggestion: una sugerencia de IA aprobada por el coach
// escribio este target. La proxima vez que el motor de progresion tocaria
// este ejercicio, lo deja intacto y limpia el flag, para que el ajuste del
// coach sobreviva hasta que el cliente entrene esa sesion puntual.
const OverrideSourceAISuggestion = "ai_suggestion"

// AssignmentRepository espeja assignment.sql.go (9 metodos). Incluye
// GetOrgMembership porque solo lo usa AssignmentService.Assign para validar
// que el cliente pertenece al gimnasio antes de asignarle un programa.
type AssignmentRepository interface {
	Create(ctx context.Context, a Assignment) (Assignment, error)
	GetByID(ctx context.Context, id uuid.UUID) (Assignment, error)
	GetActiveByClient(ctx context.Context, clientUserID uuid.UUID) (Assignment, error)
	Cancel(ctx context.Context, id, orgID uuid.UUID) (Assignment, error)

	CreateWorkout(ctx context.Context, w AssignedWorkout) (AssignedWorkout, error)
	ListWorkouts(ctx context.Context, assignmentID uuid.UUID) ([]AssignedWorkout, error)
	CreateExercise(ctx context.Context, e AssignedExercise) (AssignedExercise, error)
	ListExercises(ctx context.Context, assignedWorkoutID uuid.UUID) ([]AssignedExercise, error)

	GetOrgMembership(ctx context.Context, orgID, userID uuid.UUID) (Membership, error)

	// --- motor de progresion ---

	// MarkWorkoutCompleted cierra el dia entrenado. Sin esto nada marca un
	// assigned_workout como hecho, y "la proxima ocurrencia" devolveria el
	// dia que el cliente acaba de terminar.
	// GetWorkoutAssignmentID resuelve a que asignacion pertenece un dia. El
	// movil solo manda assigned_workout_id, no la asignacion.
	GetWorkoutAssignmentID(ctx context.Context, assignedWorkoutID uuid.UUID) (uuid.UUID, error)

	MarkWorkoutCompleted(ctx context.Context, assignedWorkoutID uuid.UUID) error

	// FindNextExerciseOccurrence busca la siguiente vez que este ejercicio
	// aparece en el plan, saltando el dia recien entrenado y los completados.
	// El bool es false cuando no queda ninguna (plan terminado o cancelado):
	// no es un error, simplemente no hay nada que progresar.
	FindNextExerciseOccurrence(ctx context.Context, assignmentID, exerciseID, excludeWorkoutID uuid.UUID) (AssignedExercise, bool, error)

	UpdateExerciseTargets(ctx context.Context, id uuid.UUID, weightKg *float64, repsMin, repsMax *int, overrideSource *string) (AssignedExercise, error)
}
