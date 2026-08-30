package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type IdentityRepository struct{ q db.Querier }

func NewIdentityRepository(q db.Querier) *IdentityRepository { return &IdentityRepository{q: q} }

var _ domain.IdentityRepository = (*IdentityRepository)(nil)

func (r *IdentityRepository) CreateUser(ctx context.Context, in domain.CreateUserInput) (domain.User, error) {
	row, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Email: in.Email, PasswordHash: stringToPgText(in.PasswordHash), FullName: in.FullName,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, domain.ErrAlreadyExists
		}
		return domain.User{}, err
	}
	return toDomainUser(row), nil
}

func (r *IdentityRepository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return toDomainUser(row), nil
}

func (r *IdentityRepository) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return toDomainUser(row), nil
}

func (r *IdentityRepository) TouchUserLogin(ctx context.Context, id uuid.UUID) error {
	return r.q.TouchUserLogin(ctx, id)
}

func (r *IdentityRepository) CreateOrganization(ctx context.Context, in domain.CreateOrganizationInput) (domain.Organization, error) {
	row, err := r.q.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name: in.Name, Slug: in.Slug, JoinCode: in.JoinCode, BrandColor: in.BrandColor,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Organization{}, domain.ErrAlreadyExists
		}
		return domain.Organization{}, err
	}
	return toDomainOrganization(row), nil
}

func (r *IdentityRepository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (domain.Organization, error) {
	row, err := r.q.GetOrganizationByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Organization{}, domain.ErrNotFound
		}
		return domain.Organization{}, err
	}
	return toDomainOrganization(row), nil
}

func (r *IdentityRepository) GetOrganizationByJoinCode(ctx context.Context, joinCode string) (domain.Organization, error) {
	row, err := r.q.GetOrganizationByJoinCode(ctx, joinCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Organization{}, domain.ErrNotFound
		}
		return domain.Organization{}, err
	}
	return toDomainOrganization(row), nil
}

func (r *IdentityRepository) CreateMembership(ctx context.Context, in domain.CreateMembershipInput) (domain.Membership, error) {
	row, err := r.q.CreateMembership(ctx, db.CreateMembershipParams{
		OrgID: in.OrgID, UserID: in.UserID, Role: db.MemberRole(in.Role),
	})
	if err != nil {
		return domain.Membership{}, err
	}
	return toDomainMembership(row), nil
}

func (r *IdentityRepository) GetActiveMembershipByUser(ctx context.Context, userID uuid.UUID) (domain.Membership, error) {
	row, err := r.q.GetActiveMembershipByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Membership{}, domain.ErrNotFound
		}
		return domain.Membership{}, err
	}
	return toDomainMembership(row), nil
}

func (r *IdentityRepository) ListOrgMembers(ctx context.Context, orgID uuid.UUID) ([]domain.OrgMember, error) {
	rows, err := r.q.ListOrgMembers(ctx, orgID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.OrgMember, len(rows))
	for i, row := range rows {
		out[i] = domain.OrgMember{
			MembershipID: row.MembershipID, UserID: row.UserID, Role: string(row.Role), Status: string(row.Status),
			JoinedAt: row.JoinedAt, FullName: row.FullName, Email: row.Email, AvatarURL: pgTextToString(row.AvatarUrl),
		}
	}
	return out, nil
}

func (r *IdentityRepository) UpdateMemberRole(ctx context.Context, membershipID, orgID uuid.UUID, role string) (domain.Membership, error) {
	row, err := r.q.UpdateMemberRole(ctx, db.UpdateMemberRoleParams{ID: membershipID, OrgID: orgID, Role: db.MemberRole(role)})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Membership{}, domain.ErrNotFound
		}
		return domain.Membership{}, err
	}
	return toDomainMembership(row), nil
}

func (r *IdentityRepository) CreateRefreshToken(ctx context.Context, in domain.CreateRefreshTokenInput) (domain.RefreshToken, error) {
	row, err := r.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID: in.UserID, TokenHash: in.TokenHash, DeviceLabel: stringToPgText(in.DeviceLabel), ExpiresAt: in.ExpiresAt,
	})
	if err != nil {
		return domain.RefreshToken{}, err
	}
	return toDomainRefreshToken(row), nil
}

func (r *IdentityRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error) {
	row, err := r.q.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.RefreshToken{}, domain.ErrNotFound
		}
		return domain.RefreshToken{}, err
	}
	return toDomainRefreshToken(row), nil
}

func (r *IdentityRepository) RevokeRefreshTokenByHash(ctx context.Context, tokenHash string) error {
	return r.q.RevokeRefreshTokenByHash(ctx, tokenHash)
}

func toDomainUser(u db.AppUser) domain.User {
	return domain.User{
		ID: u.ID, Email: u.Email, PasswordHash: pgTextToString(u.PasswordHash), FullName: u.FullName,
		AvatarURL: pgTextToString(u.AvatarUrl), Locale: u.Locale, Timezone: u.Timezone,
		EmailVerifiedAt: u.EmailVerifiedAt, IsPlatformAdmin: u.IsPlatformAdmin, LastLoginAt: u.LastLoginAt,
		CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt, DeletedAt: u.DeletedAt,
	}
}

func toDomainOrganization(o db.Organization) domain.Organization {
	return domain.Organization{
		ID: o.ID, Name: o.Name, Slug: o.Slug, JoinCode: o.JoinCode, LogoURL: pgTextToString(o.LogoUrl),
		BrandColor: o.BrandColor, Timezone: o.Timezone, Status: string(o.Status),
		CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt,
	}
}

func toDomainMembership(m db.Membership) domain.Membership {
	return domain.Membership{
		ID: m.ID, OrgID: m.OrgID, UserID: m.UserID, Role: string(m.Role), Status: string(m.Status),
		InvitedBy: pgUUIDToPtr(m.InvitedBy), JoinedAt: m.JoinedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func toDomainRefreshToken(t db.RefreshToken) domain.RefreshToken {
	return domain.RefreshToken{
		ID: t.ID, UserID: t.UserID, TokenHash: t.TokenHash, DeviceLabel: pgTextToString(t.DeviceLabel),
		ExpiresAt: t.ExpiresAt, RevokedAt: t.RevokedAt, CreatedAt: t.CreatedAt,
	}
}
