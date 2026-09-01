package security

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/config"
	"time"
)

type Identity struct{ UserID uuid.UUID }
type TokenManager struct {
	secret           []byte
	issuer, audience string
	ttl              time.Duration
}
type claims struct{ jwt.RegisteredClaims }

func NewTokenManager(cfg config.JWTConfig) *TokenManager {
	return &TokenManager{[]byte(cfg.Secret), cfg.Issuer, cfg.Audience, cfg.AccessTokenTTL}
}
func (m *TokenManager) Issue(userID uuid.UUID) (string, time.Time, error) {
	expiry := time.Now().Add(m.ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{jwt.RegisteredClaims{Issuer: m.issuer, Subject: userID.String(), Audience: jwt.ClaimStrings{m.audience}, IssuedAt: jwt.NewNumericDate(time.Now()), ExpiresAt: jwt.NewNumericDate(expiry)}})
	value, err := token.SignedString(m.secret)
	return value, expiry, err
}
func (m *TokenManager) Verify(raw string) (Identity, error) {
	var c claims
	token, err := jwt.ParseWithClaims(raw, &c, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("invalid signing method")
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithAudience(m.audience))
	if err != nil || !token.Valid {
		return Identity{}, errors.New("invalid access token")
	}
	id, err := uuid.Parse(c.Subject)
	if err != nil {
		return Identity{}, errors.New("invalid subject")
	}
	return Identity{UserID: id}, nil
}
