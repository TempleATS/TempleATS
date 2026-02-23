package auth

import (
	"testing"
)

func TestGenerateAndValidateToken(t *testing.T) {
	token, err := GenerateToken("user-1", "org-1", "acme", "admin")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != "user-1" {
		t.Errorf("expected userId user-1, got %s", claims.UserID)
	}
	if claims.OrgID != "org-1" {
		t.Errorf("expected orgId org-1, got %s", claims.OrgID)
	}
	if claims.OrgSlug != "acme" {
		t.Errorf("expected orgSlug acme, got %s", claims.OrgSlug)
	}
	if claims.Role != "admin" {
		t.Errorf("expected role admin, got %s", claims.Role)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	_, err := ValidateToken("garbage-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestValidateToken_Tampered(t *testing.T) {
	token, _ := GenerateToken("user-1", "org-1", "acme", "admin")
	// Flip a character in the signature
	tampered := token[:len(token)-1] + "X"
	_, err := ValidateToken(tampered)
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
}
