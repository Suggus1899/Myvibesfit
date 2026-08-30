package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID
	Email           string
	PasswordHash    string
	FullName        string
	AvatarURL       string
	Locale          string
	Timezone        string
	EmailVerifiedAt *time.Time
	IsPlatformAdmin bool
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type Organization struct {
	ID         uuid.UUID
	Name       string
	Slug       string
	JoinCode   string
	LogoURL    string
	BrandColor string
	Timezone   string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Membership.Role es MemberRole en Postgres (owner/admin/coach/client), pero
// ningun service rama sobre su valor (RBAC ya opera en string via
// middleware.Role) asi que se mantiene como string plano.
const (
	MemberRoleClient  = "client"
	MemberRoleOwner   = "owner"
	DefaultBrandColor = "#C6FF4F"
)

type Membership struct {
	ID        uuid.UUID
	OrgID     uuid.UUID
	UserID    uuid.UUID
	Role      string
	Status    string
	InvitedBy *uuid.UUID
	JoinedAt  *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RefreshToken struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	TokenHash   string
	DeviceLabel string
	ExpiresAt   time.Time
	RevokedAt   *time.Time
	CreatedAt   time.Time
}

type CreateUserInput struct {
	Email        string
	PasswordHash string
	FullName     string
}

type CreateOrganizationInput struct {
	Name       string
	Slug       string
	JoinCode   string
	BrandColor string
}

type CreateMembershipInput struct {
	OrgID  uuid.UUID
	UserID uuid.UUID
	Role   string
}

// OrgMember es el read-model de ListOrgMembers: membership + datos del
// usuario, para la pantalla de miembros del panel.
type OrgMember struct {
	MembershipID uuid.UUID
	UserID       uuid.UUID
	FullName     string
	Email        string
	AvatarURL    string
	Role         string
	Status       string
	JoinedAt     *time.Time
}

type CreateRefreshTokenInput struct {
	UserID      uuid.UUID
	TokenHash   string
	DeviceLabel string
	ExpiresAt   time.Time
}

// IdentityRepository espeja identity.sql.go (12 metodos): usuarios,
// organizaciones, membresias y refresh tokens.
type IdentityRepository interface {
	CreateUser(ctx context.Context, in CreateUserInput) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	TouchUserLogin(ctx context.Context, id uuid.UUID) error

	CreateOrganization(ctx context.Context, in CreateOrganizationInput) (Organization, error)
	GetOrganizationByID(ctx context.Context, id uuid.UUID) (Organization, error)
	GetOrganizationByJoinCode(ctx context.Context, joinCode string) (Organization, error)

	CreateMembership(ctx context.Context, in CreateMembershipInput) (Membership, error)
	GetActiveMembershipByUser(ctx context.Context, userID uuid.UUID) (Membership, error)
	ListOrgMembers(ctx context.Context, orgID uuid.UUID) ([]OrgMember, error)
	UpdateMemberRole(ctx context.Context, membershipID, orgID uuid.UUID, role string) (Membership, error)

	CreateRefreshToken(ctx context.Context, in CreateRefreshTokenInput) (RefreshToken, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (RefreshToken, error)
	RevokeRefreshTokenByHash(ctx context.Context, tokenHash string) error
}
