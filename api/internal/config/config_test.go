package config

import (
	"strings"
	"testing"
)

// setEnv deja el entorno limpio entre casos: Load lee de variables globales y
// un residuo de un caso anterior haria pasar un test que deberia fallar.
func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for _, k := range []string{"ENV", "DATABASE_URL", "JWT_ACCESS_SECRET", "PORT", "CORS_ALLOWED_ORIGINS", "FCM_CREDENTIALS_JSON", "FCM_PROJECT_ID"} {
		t.Setenv(k, "")
	}
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

const realSecret = "un-secreto-de-produccion-suficientemente-largo"

func TestLoadRechazaElSecretoDeDesarrolloEnProduccion(t *testing.T) {
	setEnv(t, map[string]string{
		"ENV":               "production",
		"DATABASE_URL":      "postgres://x/y",
		"JWT_ACCESS_SECRET": "dev-local-secret-do-not-use-in-production",
	})

	_, err := Load()
	if err == nil {
		t.Fatal("produccion con el secreto de desarrollo tiene que fallar al arrancar")
	}
	if !strings.Contains(err.Error(), "desarrollo") {
		t.Fatalf("el error deberia explicar cual es el problema, dijo: %v", err)
	}
}

func TestLoadExigeSecretoLargoEnProduccion(t *testing.T) {
	setEnv(t, map[string]string{
		"ENV":               "production",
		"DATABASE_URL":      "postgres://x/y",
		"JWT_ACCESS_SECRET": "corto",
	})

	if _, err := Load(); err == nil {
		t.Fatal("un secreto de 5 caracteres no puede pasar en produccion")
	}
}

// El mismo secreto corto en desarrollo no molesta: la friccion solo tiene
// sentido donde hay usuarios reales.
func TestLoadNoMolestaEnDesarrollo(t *testing.T) {
	setEnv(t, map[string]string{
		"DATABASE_URL":      "postgres://x/y",
		"JWT_ACCESS_SECRET": "dev-local-secret-do-not-use-in-production",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("desarrollo no deberia exigir un secreto de produccion: %v", err)
	}
	if cfg.Env != "development" {
		t.Fatalf("ENV vacio deberia caer en development, dio %q", cfg.Env)
	}
}

func TestLoadAceptaProduccionBienConfigurada(t *testing.T) {
	setEnv(t, map[string]string{
		"ENV":                  "production",
		"DATABASE_URL":         "postgres://x/y",
		"JWT_ACCESS_SECRET":    realSecret,
		"CORS_ALLOWED_ORIGINS": "https://panel.example.com",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if len(cfg.CORSAllowedOrigins) != 1 || cfg.CORSAllowedOrigins[0] != "https://panel.example.com" {
		t.Fatalf("origenes CORS mal parseados: %v", cfg.CORSAllowedOrigins)
	}
}

// En produccion no hay red de seguridad de CORS: si nadie los declara, la
// lista queda vacia y el navegador bloquea, que es lo correcto. El default
// permisivo es solo de desarrollo.
func TestCORSNoTieneDefaultEnProduccion(t *testing.T) {
	setEnv(t, map[string]string{
		"ENV":               "production",
		"DATABASE_URL":      "postgres://x/y",
		"JWT_ACCESS_SECRET": realSecret,
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if len(cfg.CORSAllowedOrigins) != 0 {
		t.Fatalf("produccion no deberia inventar origenes, dio %v", cfg.CORSAllowedOrigins)
	}
}
