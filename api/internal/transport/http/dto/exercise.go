package dto

import "github.com/google/uuid"

type ExerciseDTO struct {
	ID               uuid.UUID  `json:"id"`
	OrgID            *uuid.UUID `json:"org_id,omitempty"`
	Slug             string     `json:"slug"`
	Name             string     `json:"name"`
	Description      string     `json:"description,omitempty"`
	Instructions     []string   `json:"instructions"`
	Pattern          string     `json:"pattern"`
	Mechanic         string     `json:"mechanic"`
	PrimaryMuscle    string     `json:"primary_muscle"`
	SecondaryMuscles []string   `json:"secondary_muscles"`
	Equipment        []string   `json:"equipment"`
	Difficulty       string     `json:"difficulty"`
	Tracking         string     `json:"tracking"`
	IsUnilateral     bool       `json:"is_unilateral"`
	VideoURL         string     `json:"video_url,omitempty"`
	ThumbnailURL     string     `json:"thumbnail_url,omitempty"`
}

type CreateExerciseRequest struct {
	Slug             string   `json:"slug"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Instructions     []string `json:"instructions"`
	Pattern          string   `json:"pattern"`
	Mechanic         string   `json:"mechanic"`
	PrimaryMuscle    string   `json:"primary_muscle"`
	SecondaryMuscles []string `json:"secondary_muscles"`
	Equipment        []string `json:"equipment"`
	Difficulty       string   `json:"difficulty"`
	Tracking         string   `json:"tracking"`
	IsUnilateral     bool     `json:"is_unilateral"`
	VideoURL         string   `json:"video_url"`
	ThumbnailURL     string   `json:"thumbnail_url"`
}

type UpdateExerciseRequest = CreateExerciseRequest
