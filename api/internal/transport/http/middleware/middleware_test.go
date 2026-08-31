package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"myvibesfit/api/internal/platform"
)

const testSecret = "secreto-de-prueba-con-mas-de-32-caracteres"

// llegaHandler marca si la request atraveso el middleware.
func llegaHandler(llego *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*llego = true
		w.WriteHeader(http.StatusOK)
	})
}

func TestAuthRechazaLoQueNoEsUnBearerValido(t *testing.T) {
	signer := platform.NewJWTSigner(testSecret, time.Hour)
	otroSigner := platform.NewJWTSigner("otro-secreto-completamente-distinto", time.Hour)
	vencido := platform.NewJWTSigner(testSecret, -time.Minute)

	valido, err := signer.Sign(uuid.New(), nil, "client")
	if err != nil {
		t.Fatalf("no se pudo firmar: %v", err)
	}
	firmadoConOtroSecreto, err := otroSigner.Sign(uuid.New(), nil, "owner")
	if err != nil {
		t.Fatalf("no se pudo firmar: %v", err)
	}
	yaVencido, err := vencido.Sign(uuid.New(), nil, "client")
	if err != nil {
		t.Fatalf("no se pudo firmar: %v", err)
	}

	cases := []struct {
		name       string
		header     string
		wantStatus int
	}{
		{"sin header", "", http.StatusUnauthorized},
		{"esquema equivocado", "Basic abc", http.StatusUnauthorized},
		{"bearer vacio", "Bearer ", http.StatusUnauthorized},
		{"token basura", "Bearer no-es-un-jwt", http.StatusUnauthorized},
		{"firmado con otro secreto", "Bearer " + firmadoConOtroSecreto, http.StatusUnauthorized},
		{"token expirado", "Bearer " + yaVencido, http.StatusUnauthorized},
		{"token valido", "Bearer " + valido, http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var llego bool
			h := Auth(signer)(llegaHandler(&llego))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status %d, esperaba %d", rec.Code, tc.wantStatus)
			}
			if quiso := tc.wantStatus == http.StatusOK; llego != quiso {
				t.Fatalf("handler ejecutado = %v, esperaba %v", llego, quiso)
			}
		})
	}
}

func TestAuthVuelcaLosClaimsEnElContexto(t *testing.T) {
	signer := platform.NewJWTSigner(testSecret, time.Hour)
	userID, orgID := uuid.New(), uuid.New()
	token, err := signer.Sign(userID, &orgID, "coach")
	if err != nil {
		t.Fatalf("no se pudo firmar: %v", err)
	}

	var gotUser, gotOrg uuid.UUID
	var gotRole string
	var teniaOrg bool
	h := Auth(signer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, _ = UserID(r.Context())
		gotOrg, teniaOrg = OrgID(r.Context())
		gotRole = Role(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if gotUser != userID {
		t.Fatalf("user_id %v, esperaba %v", gotUser, userID)
	}
	if !teniaOrg || gotOrg != orgID {
		t.Fatalf("org_id %v (presente=%v), esperaba %v", gotOrg, teniaOrg, orgID)
	}
	if gotRole != "coach" {
		t.Fatalf("role %q, esperaba coach", gotRole)
	}
}

// Un JWT sin gimnasio no debe dejar un org_id fantasma en el contexto: es lo
// que RequireOrg mira para decidir si la request puede tocar recursos del gym.
func TestAuthSinOrgNoDejaOrgIDEnElContexto(t *testing.T) {
	signer := platform.NewJWTSigner(testSecret, time.Hour)
	token, err := signer.Sign(uuid.New(), nil, "client")
	if err != nil {
		t.Fatalf("no se pudo firmar: %v", err)
	}

	var teniaOrg bool
	h := Auth(signer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, teniaOrg = OrgID(r.Context())
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if teniaOrg {
		t.Fatal("un token sin org_id no deberia poblar el contexto")
	}
}

// OptionalAuth nunca rechaza: un token invalido tiene que pasar como anonimo
// y no como 401, o el catalogo publico de ejercicios dejaria de responder.
func TestOptionalAuthNuncaRechaza(t *testing.T) {
	signer := platform.NewJWTSigner(testSecret, time.Hour)
	valido, err := signer.Sign(uuid.New(), nil, "client")
	if err != nil {
		t.Fatalf("no se pudo firmar: %v", err)
	}

	cases := []struct {
		name          string
		header        string
		esperaUsuario bool
	}{
		{"sin header", "", false},
		{"token basura", "Bearer basura", false},
		{"token valido", "Bearer " + valido, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var hayUsuario bool
			h := OptionalAuth(signer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, hayUsuario = UserID(r.Context())
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("OptionalAuth nunca debe rechazar, devolvio %d", rec.Code)
			}
			if hayUsuario != tc.esperaUsuario {
				t.Fatalf("usuario en contexto = %v, esperaba %v", hayUsuario, tc.esperaUsuario)
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	signer := platform.NewJWTSigner(testSecret, time.Hour)

	cases := []struct {
		name       string
		permitidos []string
		rol        string // "" = request sin token, sin rol en contexto
		wantStatus int
	}{
		{"rol permitido", []string{"owner", "coach"}, "coach", http.StatusOK},
		{"rol no permitido", []string{"owner", "coach"}, "client", http.StatusForbidden},
		{"sin rol en contexto", []string{"owner"}, "", http.StatusForbidden},
		{"lista vacia no permite nada", nil, "owner", http.StatusForbidden},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var llego bool
			guard := RequireRole(tc.permitidos...)(llegaHandler(&llego))
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			// Con rol, la cadena real es Auth -> RequireRole; sin rol se prueba
			// RequireRole solo, que es como queda si Auth no poblo el contexto.
			h := guard
			if tc.rol != "" {
				token, err := signer.Sign(uuid.New(), nil, tc.rol)
				if err != nil {
					t.Fatalf("no se pudo firmar: %v", err)
				}
				req.Header.Set("Authorization", "Bearer "+token)
				h = Auth(signer)(guard)
			}

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status %d, esperaba %d", rec.Code, tc.wantStatus)
			}
			if quiso := tc.wantStatus == http.StatusOK; llego != quiso {
				t.Fatalf("handler ejecutado = %v, esperaba %v", llego, quiso)
			}
		})
	}
}

func TestRequireOrg(t *testing.T) {
	signer := platform.NewJWTSigner(testSecret, time.Hour)
	orgID := uuid.New()
	conOrg, err := signer.Sign(uuid.New(), &orgID, "coach")
	if err != nil {
		t.Fatalf("no se pudo firmar: %v", err)
	}
	sinOrg, err := signer.Sign(uuid.New(), nil, "client")
	if err != nil {
		t.Fatalf("no se pudo firmar: %v", err)
	}

	cases := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{"con gimnasio", conOrg, http.StatusOK},
		{"sin gimnasio", sinOrg, http.StatusForbidden},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var llego bool
			h := Auth(signer)(RequireOrg(llegaHandler(&llego)))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status %d, esperaba %d", rec.Code, tc.wantStatus)
			}
			if quiso := tc.wantStatus == http.StatusOK; llego != quiso {
				t.Fatalf("handler ejecutado = %v, esperaba %v", llego, quiso)
			}
		})
	}
}
