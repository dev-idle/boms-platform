package dto

import "time"

// RegisterRequest is the JSON body for POST /api/v1/auth/register.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128,password_complexity"`
	// TermsVersion is the version of the policies the customer accepted; the
	// usecase refuses any but the current one.
	TermsVersion string `json:"terms_version" validate:"max=32"`
}

// LoginRequest is the JSON body for POST /api/v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

// UserResponse is the public user shape (never includes password hash).
type UserResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Role          string    `json:"role"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

// TokenResponse is returned from login with access token metadata and user profile.
type TokenResponse struct {
	AccessToken        string       `json:"access_token"`
	TokenType          string       `json:"token_type"`
	ExpiresIn          int          `json:"expires_in"`
	User               UserResponse `json:"user"`
	MustChangePassword *bool        `json:"must_change_password,omitempty"`
}

// RefreshResponse is returned from refresh (user profile omitted — client already has it).
type RefreshResponse struct {
	AccessToken        string `json:"access_token"`
	TokenType          string `json:"token_type"`
	ExpiresIn          int    `json:"expires_in"`
	MustChangePassword *bool  `json:"must_change_password,omitempty"`
}

// VerifyEmailRequest is POST /auth/verify-email: the token from the link.
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required,max=128"`
}

// PasswordResetRequest is POST /auth/password-reset/request.
type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

// PasswordResetConfirmRequest is POST /auth/password-reset/confirm: the token
// from the link and the new password.
type PasswordResetConfirmRequest struct {
	Token       string `json:"token" validate:"required,max=128"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128,password_complexity"`
}
