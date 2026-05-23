package port

// AccessTokenClaims are signed into access JWTs (token_use=access, includes role).
type AccessTokenClaims struct {
	Subject   string
	Role      string
	SessionID string
	JTI       string
}

// RefreshTokenClaims are signed into refresh JWTs (token_use=refresh, no role).
type RefreshTokenClaims struct {
	Subject   string
	SessionID string
	JTI       string
}

// TokenSigner issues and validates EdDSA JWTs (infrastructure/jwt).
type TokenSigner interface {
	SignAccess(claims AccessTokenClaims) (string, error)
	SignRefresh(claims RefreshTokenClaims) (string, error)
	ParseAccess(token string) (AccessTokenClaims, error)
	ParseRefresh(token string) (RefreshTokenClaims, error)
}
