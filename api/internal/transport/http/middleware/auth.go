package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"myvibesfit/api/internal/platform"
)

type ctxKey string

const (
	ctxUserID ctxKey = "user_id"
	ctxOrgID  ctxKey = "org_id"
	ctxRole   ctxKey = "role"
)

// Auth exige un access token JWT valido y vuelca sus claims en el contexto.
func Auth(signer *platform.JWTSigner) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				http.Error(w, `{"error":"missing bearer token"}`, http.StatusUnauthorized)
				return
			}

			claims, err := signer.Parse(token)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
			ctx = context.WithValue(ctx, ctxRole, claims.Role)
			if claims.OrgID != nil {
				ctx = context.WithValue(ctx, ctxOrgID, *claims.OrgID)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth vuelca los claims en el contexto si viene un bearer token
// valido, pero nunca rechaza la request. La usan rutas publicas que igual
// quieren personalizar la respuesta por organizacion cuando hay sesion
// (p. ej. el catalogo de ejercicios: global siempre, mas el del gym si
// el caller esta autenticado).
func OptionalAuth(signer *platform.JWTSigner) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := signer.Parse(token)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
			ctx = context.WithValue(ctx, ctxRole, claims.Role)
			if claims.OrgID != nil {
				ctx = context.WithValue(ctx, ctxOrgID, *claims.OrgID)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ctxUserID).(uuid.UUID)
	return id, ok
}

func OrgID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ctxOrgID).(uuid.UUID)
	return id, ok
}

func Role(ctx context.Context) string {
	role, _ := ctx.Value(ctxRole).(string)
	return role
}
