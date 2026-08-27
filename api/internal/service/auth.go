package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/platform"
)

type AuthService struct {
	repo       domain.IdentityRepository
	access     *platform.JWTSigner
	refreshTTL time.Duration
}

func NewAuthService(repo domain.IdentityRepository, access *platform.JWTSigner, refreshTTL time.Duration) *AuthService {
	return &AuthService{repo: repo, access: access, refreshTTL: refreshTTL}
}

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	User         domain.User
	OrgID        *uuid.UUID
	Role         string
}

func (s *AuthService) Register(ctx context.Context, email, password, fullName string) (*AuthResult, error) {
	hash, err := platform.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.CreateUser(ctx, domain.CreateUserInput{Email: email, PasswordHash: hash, FullName: fullName})
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user, nil, "")
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if user.PasswordHash == "" || !platform.VerifyPassword(user.PasswordHash, password) {
		return nil, domain.ErrInvalidCredentials
	}
	_ = s.repo.TouchUserLogin(ctx, user.ID)

	orgID, role := s.resolveActiveOrg(ctx, user.ID)
	return s.issueTokens(ctx, user, orgID, role)
}

func (s *AuthService) Refresh(ctx context.Context, rawToken string) (*AuthResult, error) {
	if rawToken == "" {
		return nil, domain.ErrUnauthorized
	}
	hash := platform.HashRefreshToken(rawToken)
	rt, err := s.repo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	if time.Now().After(rt.ExpiresAt) {
		return nil, domain.ErrUnauthorized
	}
	// Rotacion: el refresh usado queda inutilizable de inmediato.
	_ = s.repo.RevokeRefreshTokenByHash(ctx, hash)

	user, err := s.repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	orgID, role := s.resolveActiveOrg(ctx, user.ID)
	return s.issueTokens(ctx, user, orgID, role)
}

func (s *AuthService) JoinOrganization(ctx context.Context, userID uuid.UUID, joinCode string) (*AuthResult, error) {
	org, err := s.repo.GetOrganizationByJoinCode(ctx, joinCode)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	if _, err := s.repo.CreateMembership(ctx, domain.CreateMembershipInput{
		OrgID: org.ID, UserID: userID, Role: domain.MemberRoleClient,
	}); err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	return s.issueTokens(ctx, user, &org.ID, domain.MemberRoleClient)
}

func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*AuthResult, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	orgID, role := s.resolveActiveOrg(ctx, userID)
	return &AuthResult{User: user, OrgID: orgID, Role: role}, nil
}

// resolveActiveOrg toma la membresia activa mas reciente del usuario.
// Un usuario sin gimnasio (registro abierto) simplemente no tiene ninguna.
func (s *AuthService) resolveActiveOrg(ctx context.Context, userID uuid.UUID) (*uuid.UUID, string) {
	m, err := s.repo.GetActiveMembershipByUser(ctx, userID)
	if err != nil {
		return nil, ""
	}
	orgID := m.OrgID
	return &orgID, m.Role
}

func (s *AuthService) issueTokens(ctx context.Context, user domain.User, orgID *uuid.UUID, role string) (*AuthResult, error) {
	access, err := s.access.Sign(user.ID, orgID, role)
	if err != nil {
		return nil, err
	}

	rawRefresh, err := platform.NewRefreshTokenValue()
	if err != nil {
		return nil, err
	}

	if _, err := s.repo.CreateRefreshToken(ctx, domain.CreateRefreshTokenInput{
		UserID: user.ID, TokenHash: platform.HashRefreshToken(rawRefresh), ExpiresAt: time.Now().Add(s.refreshTTL),
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
