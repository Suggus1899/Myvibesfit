package dto

import (
	"time"

	"github.com/google/uuid"
)

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

type CreateOrgRequest struct {
	Name string `json:"name"`
}

type OrgDTO struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	JoinCode   string    `json:"join_code"`
	BrandColor string    `json:"brand_color"`
}

type OrgMemberDTO struct {
	MembershipID uuid.UUID  `json:"membership_id"`
	UserID       uuid.UUID  `json:"user_id"`
	FullName     string     `json:"full_name"`
	Email        string     `json:"email"`
	AvatarURL    string     `json:"avatar_url,omitempty"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	JoinedAt     *time.Time `json:"joined_at,omitempty"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role"`
}
