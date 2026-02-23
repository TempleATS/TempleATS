package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims for an authenticated user.
type Claims struct {
	UserID  string `json:"userId"`
	OrgID   string `json:"orgId"`
	OrgSlug string `json:"orgSlug"`
	Role    string `json:"role"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

func getSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}
	return []byte(secret)
}

// GenerateToken creates a signed JWT for the given user.
func GenerateToken(userID, orgID, orgSlug, role string) (string, error) {
	claims := Claims{
		UserID:  userID,
		OrgID:   orgID,
		OrgSlug: orgSlug,
		Role:    role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecret())
}

// ValidateToken parses and validates a JWT string, returning the claims.
func ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return getSecret(), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
