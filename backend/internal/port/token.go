package port

// TokenIssuer defines outbound auth token operations (implement in adapter when adding auth).
type TokenIssuer interface {
	IssueAccessToken(subject string, claims map[string]any) (string, error)
	IssueRefreshToken(subject string) (string, error)
}

// TokenValidator validates JWTs without embedding business rules.
type TokenValidator interface {
	ParseAccessToken(token string) (subject string, err error)
	ParseRefreshToken(token string) (subject string, err error)
}
