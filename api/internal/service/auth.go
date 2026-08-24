package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/platform"
	"myvibesfit/api/internal/repository/db"
)

type AuthService struct {
	q          db.Querier
	access     *platform.JWTSigner
	refreshTTL time.Duration
}

func NewAuthService(q db.Querier, access *platform.JWTSigner, refreshTTL time.Duration) *AuthService {
	return &AuthService{q: q, access: access, refreshTTL: refreshTTL}
}

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	User         db.AppUser
	OrgID        *uuid.UUID
	Role         string
}

func (s *AuthService) Register(ctx context.Context, email, password, fullName string) (*AuthResult, error) {
	hash, err := platform.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user, err := s.q.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: pgtype.Text{String: hash, Valid: true},
		FullName:     fullName,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrAlreadyExists
		}
		return nil, err
	}

	return s.issueTokens(ctx, user, nil, "")
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if !user.PasswordHash.Valid || !platform.VerifyPassword(user.PasswordHash.String, password) {
		return nil, domain.ErrInvalidCredentials
	}
	_ = s.q.TouchUserLogin(ctx, user.ID)

	orgID, role := s.resolveActiveOrg(ctx, user.ID)
	return s.issueTokens(ctx, user, orgID, role)
}

func (s *AuthService) Refresh(ctx context.Context, rawToken string) (*AuthResult, error) {
	if rawToken == "" {
		return nil, domain.ErrUnauthorized
	}
	hash := platform.HashRefreshToken(rawToken)
	rt, err := s.q.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	if time.Now().After(rt.ExpiresAt) {
		return nil, domain.ErrUnauthorized
	}
	// Rotacion: el refresh usado queda inutilizable de inmediato.
	_ = s.q.RevokeRefreshTokenByHash(ctx, hash)

	user, err := s.q.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	orgID, role := s.resolveActiveOrg(ctx, user.ID)
	return s.issueTokens(ctx, user, orgID, role)
}

func (s *AuthService) JoinOrganization(ctx context.Context, userID uuid.UUID, joinCode string) (*AuthResult, error) {
	org, err := s.q.GetOrganizationByJoinCode(ctx, joinCode)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	if _, err := s.q.CreateMembership(ctx, db.CreateMembershipParams{
		OrgID:  org.ID,
		UserID: userID,
		Role:   db.MemberRoleClient,
	}); err != nil {
		return nil, err
	}

	user, err := s.q.GetUserByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	role := string(db.MemberRoleClient)
	return s.issueTokens(ctx, user, &org.ID, role)
}

func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*AuthResult, error) {
	user, err := s.q.GetUserByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	orgID, role := s.resolveActiveOrg(ctx, userID)
	return &AuthResult{User: user, OrgID: orgID, Role: role}, nil
}

// resolveActiveOrg toma la membresia activa mas reciente del usuario.
// Un usuario sin gimnasio (registro abierto) simplemente no tiene ninguna.
func (s *AuthService) resolveActiveOrg(ctx context.Context, userID uuid.UUID) (*uuid.UUID, string) {
	m, err := s.q.GetActiveMembershipByUser(ctx, userID)
	if err != nil {
		return nil, ""
	}
	orgID := m.OrgID
	return &orgID, string(m.Role)
}

func (s *AuthService) issueTokens(ctx context.Context, user db.AppUser, orgID *uuid.UUID, role string) (*AuthResult, error) {
	access, err := s.access.Sign(user.ID, orgID, role)
	if err != nil {
		return nil, err
	}

	rawRefresh, err := platform.NewRefreshTokenValue()
	if err != nil {
		return nil, err
	}

	if _, err := s.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:      user.ID,
		TokenHash:   platform.HashRefreshToken(rawRefresh),
		DeviceLabel: pgtype.Text{},
		ExpiresAt:   time.Now().Add(s.refreshTTL),
	}); err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  access,
		RefreshToken: rawRefresh,
		User:         user,
		OrgID:        orgID,
		Role:         role,
	}, nil
}
