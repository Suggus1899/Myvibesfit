package platform

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTSignParseRoundTrip(t *testing.T) {
	signer := NewJWTSigner("test-secret", time.Minute)
	userID := uuid.New()
	orgID := uuid.New()

	token, err := signer.Sign(userID, &orgID, "coach")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	claims, err := signer.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.OrgID == nil || *claims.OrgID != orgID {
		t.Errorf("OrgID = %v, want %v", claims.OrgID, orgID)
	}
	if claims.Role != "coach" {
		t.Errorf("Role = %q, want %q", claims.Role, "coach")
	}
}

func TestJWTExpiredTokenRejected(t *testing.T) {
	signer := NewJWTSigner("test-secret", -time.Minute) // ya vencido al firmar
	token, err := signer.Sign(uuid.New(), nil, "")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if _, err := signer.Parse(token); err == nil {
		t.Error("Parse() on expired token: want error, got nil")
	}
}

func TestJWTWrongSecretRejected(t *testing.T) {
	signer := NewJWTSigner("secret-a", time.Minute)
	token, err := signer.Sign(uuid.New(), nil, "")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	other := NewJWTSigner("secret-b", time.Minute)
	if _, err := other.Parse(token); err == nil {
		t.Error("Parse() with wrong secret: want error, got nil")
	}
}
