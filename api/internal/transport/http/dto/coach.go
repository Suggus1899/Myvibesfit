package dto

import (
	"time"

	"github.com/google/uuid"
)

type CoachClientDTO struct {
	ClientUserID   uuid.UUID  `json:"client_user_id"`
	FullName       string     `json:"full_name"`
	AvatarURL      string     `json:"avatar_url,omitempty"`
	AssignmentID   *uuid.UUID `json:"assignment_id,omitempty"`
	AssignmentName string     `json:"assignment_name,omitempty"`
	HasAssignment  bool       `json:"has_assignment"`
	LastSessionAt  *time.Time `json:"last_session_at,omitempty"`
	StreakDays     int        `json:"streak_days"`
	RecentPRs      int        `json:"recent_prs"`
	NeedsAttention bool       `json:"needs_attention"`
}
