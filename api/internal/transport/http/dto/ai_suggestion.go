package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AISuggestionDTO struct {
	ID           uuid.UUID       `json:"id"`
	ClientUserID uuid.UUID       `json:"client_user_id"`
	ClientName   string          `json:"client_name"`
	AssignmentID *uuid.UUID      `json:"assignment_id,omitempty"`
	Kind         string          `json:"kind"`
	Payload      json.RawMessage `json:"payload"`
	Rationale    string          `json:"rationale"`
	Confidence   *float64        `json:"confidence,omitempty"`
	Model        string          `json:"model"`
	Status       string          `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	ExpiresAt    time.Time       `json:"expires_at"`
}
