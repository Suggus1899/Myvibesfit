package middleware

import "net/http"

// RequireRole exige que el rol del JWT (dentro del gimnasio activo) este
// entre los permitidos. Se monta despues de Auth.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !allowed[Role(r.Context())] {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireOrg exige que el JWT traiga org_id resuelto (usuario vinculado a
// un gimnasio). Se monta despues de Auth, antes de handlers que operan
// sobre recursos siempre scoped a organizacion (programas, asignaciones).
func RequireOrg(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := OrgID(r.Context()); !ok {
			http.Error(w, `{"error":"organization context required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
