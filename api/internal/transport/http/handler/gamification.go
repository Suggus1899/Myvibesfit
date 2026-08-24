package handler

import (
	"net/http"
	"time"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
	"myvibesfit/api/internal/service"
	"myvibesfit/api/internal/transport/http/dto"
	"myvibesfit/api/internal/transport/http/middleware"
)

type GamificationHandler struct {
	q   db.Querier
	gam *service.GamificationService
}

func NewGamificationHandler(q db.Querier, gam *service.GamificationService) *GamificationHandler {
	return &GamificationHandler{q: q, gam: gam}
}

func (h *GamificationHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}

	stats, streaks, err := h.gam.Stats(r.Context(), h.q, userID)
	if err != nil {
		writeError(w, err)
		return
	}

	out := dto.UserStatsDTO{
		TotalXP: int(stats.TotalXp), Level: int(stats.Level),
		TotalSessions: int(stats.TotalSessions), TotalVolumeKg: stats.TotalVolumeKg,
	}
	for _, s := range streaks {
		out.Streaks = append(out.Streaks, dto.StreakDTO{
			Kind: string(s.Kind), CurrentCount: int(s.CurrentCount), LongestCount: int(s.LongestCount),
			LastActiveOn: datePtrToString(s.LastActiveOn),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *GamificationHandler) Achievements(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}

	rows, err := h.gam.MyAchievements(r.Context(), h.q, userID)
	if err != nil {
		writeError(w, err)
		return
	}

	out := make([]dto.UserAchievementDTO, len(rows))
	for i, row := range rows {
		earnedAt := ""
		if row.EarnedAt != nil {
			earnedAt = row.EarnedAt.Format(time.RFC3339)
		}
		out[i] = dto.UserAchievementDTO{
			Code: row.Code, Name: row.Name, Description: row.Description, Icon: row.Icon,
			Tier: int(row.Tier), Progress: row.Progress, EarnedAt: earnedAt,
		}
	}
	writeJSON(w, http.StatusOK, out)
}
