package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/service"
	"myvibesfit/api/internal/transport/http/dto"
	"myvibesfit/api/internal/transport/http/middleware"
)

type SyncHandler struct {
	svc *service.SyncService
}

func NewSyncHandler(svc *service.SyncService) *SyncHandler {
	return &SyncHandler{svc: svc}
}

func (h *SyncHandler) Sync(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}

	var req dto.SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Sessions) == 0 {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	var orgID *uuid.UUID
	if id, ok := middleware.OrgID(r.Context()); ok {
		orgID = &id
	}

	sessions := make([]service.SyncSessionInput, len(req.Sessions))
	for i, s := range req.Sessions {
		exercises := make([]service.SyncExerciseInput, len(s.Exercises))
		for j, e := range s.Exercises {
			sets := make([]service.SyncSetInput, len(e.Sets))
			for k, set := range e.Sets {
				sets[k] = service.SyncSetInput{
					ClientLocalID: set.ClientLocalID, SetNumber: set.SetNumber, Type: set.Type,
					WeightKg: set.WeightKg, Reps: set.Reps, RPE: set.RPE, RIR: set.RIR,
					DurationSeconds: set.DurationSeconds, DistanceM: set.DistanceM,
					RestTakenSeconds: set.RestTakenSeconds, IsCompleted: set.IsCompleted, PerformedAt: set.PerformedAt,
				}
			}
			exercises[j] = service.SyncExerciseInput{
				ExerciseID: e.ExerciseID, AssignedExerciseID: e.AssignedExerciseID, OrderIndex: e.OrderIndex,
				SupersetGroup: e.SupersetGroup, Note: e.Note, Sets: sets,
			}
		}
		sessions[i] = service.SyncSessionInput{
			ClientLocalID: s.ClientLocalID, Name: s.Name, Status: s.Status, AssignedWorkoutID: s.AssignedWorkoutID,
			StartedAt: s.StartedAt, EndedAt: s.EndedAt, DurationSeconds: s.DurationSeconds,
			PerceivedEffort: s.PerceivedEffort, Mood: s.Mood, Notes: s.Notes, Exercises: exercises,
		}
	}

	result, err := h.svc.SyncSessions(r.Context(), userID, orgID, sessions)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := dto.SyncResponse{
		Sessions:           make([]dto.SyncedSessionDTO, len(result.Sessions)),
		NewPersonalRecords: result.NewPersonalRecords,
	}
	for i, s := range result.Sessions {
		resp.Sessions[i] = dto.SyncedSessionDTO{
			ID: s.ID, ClientLocalID: s.ClientLocalID, Status: string(s.Status), TotalVolumeKg: s.TotalVolumeKg,
		}
	}
	for _, a := range result.UnlockedAchievements {
		resp.UnlockedAchievements = append(resp.UnlockedAchievements, achievementDTO(a))
	}
	writeJSON(w, http.StatusOK, resp)
}

func achievementDTO(a domain.Achievement) dto.AchievementDTO {
	return dto.AchievementDTO{
		ID: a.ID, Code: a.Code, Name: a.Name, Description: a.Description,
		Icon: a.Icon, Tier: int(a.Tier), XPReward: int(a.XPReward),
	}
}
