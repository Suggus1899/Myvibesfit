package platform

import "testing"

func TestHashPasswordVerifyRoundTrip(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if !VerifyPassword(hash, "correct-horse-battery-staple") {
		t.Error("VerifyPassword() with correct password: want true, got false")
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Error("VerifyPassword() with wrong password: want false, got true")
	}
}
