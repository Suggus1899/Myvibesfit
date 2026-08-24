package dto

import "github.com/google/uuid"

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type JoinOrgRequest struct {
	JoinCode string `json:"join_code"`
}

type UserDTO struct {
	ID       uuid.UUID  `json:"id"`
	Email    string     `json:"email"`
	FullName string     `json:"full_name"`
	OrgID    *uuid.UUID `json:"org_id,omitempty"`
	Role     string     `json:"role,omitempty"`
}

type AuthResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	User         UserDTO `json:"user"`
}
