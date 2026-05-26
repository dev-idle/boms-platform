// Package jwt provides JWT signing adapters (EdDSA / Ed25519).
package jwt

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/golang-jwt/jwt/v5"
)

const (
	claimTokenUse   = "token_use"
	tokenUseAccess  = "access"
	tokenUseRefresh = "refresh"
	claimSessionID  = "sid"
	claimRole       = "role"
	claimJTI        = "jti"
	keyID           = "v1"
	allowedAlg      = "EdDSA"
)

var allowedAlgorithms = []string{allowedAlg}

// EdDSASigner implements port.TokenSigner with Ed25519 (JWT alg EdDSA, kid=v1).
type EdDSASigner struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	issuer     string
	audience   string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewEdDSASigner builds a signer from application JWT config.
func NewEdDSASigner(cfg config.JWTConfig) (*EdDSASigner, error) {
	priv, pub, err := loadKeyPairFromSeed(cfg.Ed25519PrivateKey)
	if err != nil {
		return nil, err
	}
	return &EdDSASigner{
		privateKey: priv,
		publicKey:  pub,
		issuer:     cfg.Issuer,
		audience:   cfg.Audience,
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}, nil
}

// SignAccess implements port.TokenSigner.
func (s *EdDSASigner) SignAccess(in port.AccessTokenClaims) (string, error) {
	if in.Subject == "" {
		return "", apperrors.ErrValidation.WithDetail("field", "subject")
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":          s.issuer,
		"aud":          s.audience,
		"sub":          in.Subject,
		claimRole:      in.Role,
		claimSessionID: in.SessionID,
		claimJTI:       in.JTI,
		claimTokenUse:  tokenUseAccess,
		"iat":          now.Unix(),
		"exp":          now.Add(s.accessTTL).Unix(),
	}
	return s.sign(claims)
}

// SignRefresh implements port.TokenSigner.
func (s *EdDSASigner) SignRefresh(in port.RefreshTokenClaims) (string, error) {
	if in.Subject == "" {
		return "", apperrors.ErrValidation.WithDetail("field", "subject")
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":          s.issuer,
		"aud":          s.audience,
		"sub":          in.Subject,
		claimSessionID: in.SessionID,
		claimJTI:       in.JTI,
		claimTokenUse:  tokenUseRefresh,
		"iat":          now.Unix(),
		"exp":          now.Add(s.refreshTTL).Unix(),
	}
	return s.sign(claims)
}

// ParseAccess implements port.TokenSigner.
func (s *EdDSASigner) ParseAccess(token string) (port.AccessTokenClaims, error) {
	claims, err := s.parseToken(token, tokenUseAccess)
	if err != nil {
		return port.AccessTokenClaims{}, err
	}
	subject, err := subjectFromClaims(claims)
	if err != nil {
		return port.AccessTokenClaims{}, err
	}
	sid := stringClaim(claims, claimSessionID)
	jti := stringClaim(claims, claimJTI)
	if sid == "" || jti == "" {
		return port.AccessTokenClaims{}, apperrors.ErrUnauthorized
	}
	return port.AccessTokenClaims{
		Subject:   subject,
		Role:      stringClaim(claims, claimRole),
		SessionID: sid,
		JTI:       jti,
	}, nil
}

// ParseRefresh implements port.TokenSigner.
func (s *EdDSASigner) ParseRefresh(token string) (port.RefreshTokenClaims, error) {
	claims, err := s.parseToken(token, tokenUseRefresh)
	if err != nil {
		return port.RefreshTokenClaims{}, err
	}
	subject, err := subjectFromClaims(claims)
	if err != nil {
		return port.RefreshTokenClaims{}, err
	}
	if role := stringClaim(claims, claimRole); role != "" {
		return port.RefreshTokenClaims{}, apperrors.ErrUnauthorized
	}
	sid := stringClaim(claims, claimSessionID)
	jti := stringClaim(claims, claimJTI)
	if sid == "" || jti == "" {
		return port.RefreshTokenClaims{}, apperrors.ErrUnauthorized
	}
	return port.RefreshTokenClaims{
		Subject:   subject,
		SessionID: sid,
		JTI:       jti,
	}, nil
}

func (s *EdDSASigner) sign(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = keyID
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}

func (s *EdDSASigner) parseToken(raw, expectedUse string) (jwt.MapClaims, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, apperrors.ErrUnauthorized
	}
	parsed, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if t.Method == nil || t.Method.Alg() != allowedAlg {
			return nil, fmt.Errorf("unexpected alg: %v", t.Header["alg"])
		}
		if alg, _ := t.Header["alg"].(string); strings.EqualFold(alg, "none") {
			return nil, errors.New("alg none rejected")
		}
		return s.publicKey, nil
	}, jwt.WithValidMethods(allowedAlgorithms))
	if err != nil {
		// Distinguish expiry from other validation failures so callers (FE retry / refresh)
		// can react accordingly. All other failures collapse to a generic unauthorized.
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.ErrTokenExpired
		}
		return nil, apperrors.ErrUnauthorized
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return nil, apperrors.ErrUnauthorized
	}
	if use, _ := claims[claimTokenUse].(string); use != expectedUse {
		return nil, apperrors.ErrUnauthorized
	}
	if iss, _ := claims["iss"].(string); iss != s.issuer {
		return nil, apperrors.ErrUnauthorized
	}
	if !audienceMatches(claims, s.audience) {
		return nil, apperrors.ErrUnauthorized
	}
	return claims, nil
}

func audienceMatches(claims jwt.MapClaims, expected string) bool {
	aud, ok := claims["aud"]
	if !ok {
		return false
	}
	switch v := aud.(type) {
	case string:
		return v == expected
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s == expected {
				return true
			}
		}
	}
	return false
}

func subjectFromClaims(claims jwt.MapClaims) (string, error) {
	subject, err := claims.GetSubject()
	if err != nil || subject == "" {
		return "", apperrors.ErrUnauthorized
	}
	return subject, nil
}

func stringClaim(claims jwt.MapClaims, key string) string {
	v, _ := claims[key].(string)
	return v
}

func loadKeyPairFromSeed(seedB64 string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	seedB64 = strings.TrimSpace(seedB64)
	if seedB64 == "" {
		return nil, nil, errors.New("jwt ed25519 private key seed is empty")
	}
	seed, err := base64.StdEncoding.DecodeString(seedB64)
	if err != nil {
		return nil, nil, fmt.Errorf("decode jwt ed25519 seed: %w", err)
	}
	if len(seed) != ed25519.SeedSize {
		return nil, nil, fmt.Errorf("jwt ed25519 seed must be %d bytes", ed25519.SeedSize)
	}
	priv := ed25519.NewKeyFromSeed(seed)
	return priv, priv.Public().(ed25519.PublicKey), nil
}

var _ port.TokenSigner = (*EdDSASigner)(nil)
