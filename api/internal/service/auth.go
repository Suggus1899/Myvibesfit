package service

import (
	"context"
	"fmt"
	"log"
	"strings"
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
	if err := s.repo.TouchUserLogin(ctx, user.ID); err != nil {
		log.Printf("auth: touch last_login_at failed for user %s: %v", user.ID, err)
	}

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
	// Rotacion: el refresh usado queda inutilizable de inmediato. Si la
	// revocacion falla el token viejo sigue vivo — no aborta el refresh, pero
	// tiene que quedar rastro para poder detectarlo.
	if err := s.repo.RevokeRefreshTokenByHash(ctx, hash); err != nil {
		log.Printf("auth: revoke refresh token failed for user %s: %v", rt.UserID, err)
	}

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

// CreateOrganization crea un gimnasio nuevo y deja al creador como owner.
// Reemite tokens con el org_id/role nuevos, igual que JoinOrganization,
// para que el cliente no tenga que hacer un segundo refresh.
func (s *AuthService) CreateOrganization(ctx context.Context, userID uuid.UUID, name string) (*AuthResult, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: el nombre del gimnasio es obligatorio", domain.ErrInvalidInput)
	}

	joinCode, err := platform.NewJoinCode()
	if err != nil {
		return nil, err
	}
	slug := slugify(name) + "-" + strings.ToLower(joinCode[:6])

	org, err := s.repo.CreateOrganization(ctx, domain.CreateOrganizationInput{
		Name: name, Slug: slug, JoinCode: joinCode, BrandColor: domain.DefaultBrandColor,
	})
	if err != nil {
		return nil, err
	}

	if _, err := s.repo.CreateMembership(ctx, domain.CreateMembershipInput{
		OrgID: org.ID, UserID: userID, Role: domain.MemberRoleOwner,
	}); err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user, &org.ID, domain.MemberRoleOwner)
}

func (s *AuthService) GetOrganization(ctx context.Context, orgID uuid.UUID) (domain.Organization, error) {
	return s.repo.GetOrganizationByID(ctx, orgID)
}

func (s *AuthService) ListOrgMembers(ctx context.Context, orgID uuid.UUID) ([]domain.OrgMember, error) {
	return s.repo.ListOrgMembers(ctx, orgID)
}

var updatableMemberRoles = map[string]bool{"admin": true, "coach": true, domain.MemberRoleClient: true}

// UpdateMemberRole no acepta "owner": promover a owner no es un cambio de
// rol casual desde el panel, y el creador del gimnasio ya lo es.
func (s *AuthService) UpdateMemberRole(ctx context.Context, orgID, membershipID uuid.UUID, role string) (domain.Membership, error) {
	if !updatableMemberRoles[role] {
		return domain.Membership{}, fmt.Errorf("%w: rol invalido", domain.ErrInvalidInput)
	}
	return s.repo.UpdateMemberRole(ctx, membershipID, orgID, role)
}

// slugify normaliza un nombre a minusculas/guiones para usar como slug de
// organizacion; la unicidad la garantiza el sufijo aleatorio del caller.
func slugify(name string) string {
	var b strings.Builder
	dash := true // evita guion inicial
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		case !dash:
			b.WriteByte('-')
			dash = true
		}
	}
	s := strings.TrimRight(b.String(), "-")
	if s == "" {
		return "gym"
	}
	return s
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
