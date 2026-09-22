package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newAuthUC(t *testing.T, users *mockUserRepo, customerProfiles *mockCustomerProfileRepo, sessions *mockSessionStore, hasher *mockHasher, signer *mockSigner) *usecase.AuthUsecase {
	t.Helper()
	hasher.On("Hash", usecase.TimingSafeDummySeed).Return("dummy-hash", nil).Once()
	uc, err := usecase.NewAuthUsecase(users, customerProfiles, passthroughTxManager{}, sessions, hasher, signer, nil)
	require.NoError(t, err)
	return uc
}

func TestAuthUsecase_RegisterDuplicateEmail(t *testing.T) {
	t.Parallel()
	users := new(mockUserRepo)
	customerProfiles := new(mockCustomerProfileRepo)
	hasher := new(mockHasher)
	signer := new(mockSigner)
	sessions := new(mockSessionStore)
	uc := newAuthUC(t, users, customerProfiles, sessions, hasher, signer)

	hasher.On("Hash", "Password1").Return("hash", nil)
	users.On("Create", mock.Anything, mock.Anything).Return(nil, apperrors.ErrConflict)

	_, err := uc.Register(context.Background(), dto.RegisterRequest{Email: "a@b.com", Password: "Password1"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrEmailExists))
}

func TestAuthUsecase_LoginWrongPassword(t *testing.T) {
	t.Parallel()
	users := new(mockUserRepo)
	customerProfiles := new(mockCustomerProfileRepo)
	hasher := new(mockHasher)
	signer := new(mockSigner)
	sessions := new(mockSessionStore)
	uc := newAuthUC(t, users, customerProfiles, sessions, hasher, signer)

	user := &domainuser.User{ID: uuid.New(), Email: "a@b.com", PasswordHash: "real-hash", Role: domainuser.RoleCustomer}
	users.On("GetByEmail", mock.Anything, "a@b.com").Return(user, nil)
	hasher.On("Verify", "real-hash", "wrong").Return(apperrors.ErrInvalidCredentials)

	_, _, _, err := uc.Login(context.Background(), dto.LoginRequest{Email: "a@b.com", Password: "wrong"}, "ua", "127.0.0.1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrInvalidCredentials))
	hasher.AssertNumberOfCalls(t, "Verify", 1)
}

func TestAuthUsecase_LoginUserNotFoundTimingSafe(t *testing.T) {
	t.Parallel()
	users := new(mockUserRepo)
	customerProfiles := new(mockCustomerProfileRepo)
	hasher := new(mockHasher)
	signer := new(mockSigner)
	sessions := new(mockSessionStore)
	uc := newAuthUC(t, users, customerProfiles, sessions, hasher, signer)

	users.On("GetByEmail", mock.Anything, "missing@b.com").Return(nil, apperrors.ErrNotFound)
	hasher.On("Verify", "dummy-hash", "Password1").Return(apperrors.ErrInvalidCredentials)

	_, _, _, err := uc.Login(context.Background(), dto.LoginRequest{Email: "missing@b.com", Password: "Password1"}, "ua", "127.0.0.1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrInvalidCredentials))
	hasher.AssertCalled(t, "Verify", "dummy-hash", "Password1")
	hasher.AssertNumberOfCalls(t, "Verify", 1)
}

func TestAuthUsecase_RefreshHappyPath(t *testing.T) {
	t.Parallel()
	users := new(mockUserRepo)
	hasher := new(mockHasher)
	signer := new(mockSigner)
	sessions := new(mockSessionStore)
	uc := newAuthUC(t, users, new(mockCustomerProfileRepo), sessions, hasher, signer)

	uid := uuid.New()
	oldSid := uuid.NewString()
	oldJti := uuid.NewString()
	users.On("GetByID", mock.Anything, uid).Return(&domainuser.User{ID: uid, Role: domainuser.RoleCustomer}, nil)
	signer.On("ParseRefresh", "refresh-raw").Return(port.RefreshTokenClaims{
		Subject: uid.String(), SessionID: oldSid, JTI: oldJti,
	}, nil)
	sessions.On("Rotate", mock.Anything, uid.String(), oldSid, mock.AnythingOfType("string"), oldJti, mock.Anything).Return(nil)
	signer.On("SignAccess", mock.Anything).Return("access-new", nil)
	signer.On("SignRefresh", mock.Anything).Return("refresh-new", nil)

	access, refresh, mustChange, err := uc.Refresh(context.Background(), "refresh-raw", "ua", "1.2.3.4")
	require.NoError(t, err)
	assert.Equal(t, "access-new", access)
	assert.Equal(t, "refresh-new", refresh)
	assert.False(t, mustChange)
	sessions.AssertCalled(t, "Rotate", mock.Anything, uid.String(), oldSid, mock.AnythingOfType("string"), oldJti, mock.Anything)
}

func TestAuthUsecase_RefreshReuseJTImismatch(t *testing.T) {
	t.Parallel()
	users := new(mockUserRepo)
	sessions := new(mockSessionStore)
	signer := new(mockSigner)
	uc := newAuthUC(t, users, new(mockCustomerProfileRepo), sessions, new(mockHasher), signer)

	uid := uuid.New()
	signer.On("ParseRefresh", "rt").Return(port.RefreshTokenClaims{Subject: uid.String(), SessionID: "s1", JTI: "jti-a"}, nil)
	users.On("GetByID", mock.Anything, uid).Return(&domainuser.User{ID: uid, Role: domainuser.RoleCustomer}, nil)
	sessions.On("Rotate", mock.Anything, uid.String(), "s1", mock.AnythingOfType("string"), "jti-a", mock.Anything).Return(apperrors.ErrConflict)
	sessions.On("DeleteAllForUser", mock.Anything, uid.String()).Return(nil)

	_, _, _, err := uc.Refresh(context.Background(), "rt", "ua", "ip")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrSessionRevoked))
	sessions.AssertCalled(t, "DeleteAllForUser", mock.Anything, uid.String())
}

func TestAuthUsecase_RefreshMissingSession(t *testing.T) {
	t.Parallel()
	users := new(mockUserRepo)
	sessions := new(mockSessionStore)
	signer := new(mockSigner)
	uc := newAuthUC(t, users, new(mockCustomerProfileRepo), sessions, new(mockHasher), signer)

	uid := uuid.New()
	signer.On("ParseRefresh", "rt").Return(port.RefreshTokenClaims{Subject: uid.String(), SessionID: "s1", JTI: "j1"}, nil)
	users.On("GetByID", mock.Anything, uid).Return(&domainuser.User{ID: uid, Role: domainuser.RoleCustomer}, nil)
	sessions.On("Rotate", mock.Anything, uid.String(), "s1", mock.AnythingOfType("string"), "j1", mock.Anything).Return(apperrors.ErrNotFound)
	sessions.On("DeleteAllForUser", mock.Anything, uid.String()).Return(nil)

	_, _, _, err := uc.Refresh(context.Background(), "rt", "ua", "ip")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrSessionRevoked))
}

func TestAuthUsecase_RefreshAccessTokenRejected(t *testing.T) {
	t.Parallel()
	signer := new(mockSigner)
	uc := newAuthUC(t, new(mockUserRepo), new(mockCustomerProfileRepo), new(mockSessionStore), new(mockHasher), signer)

	signer.On("ParseRefresh", "access-token").Return(port.RefreshTokenClaims{}, apperrors.ErrUnauthorized)

	_, _, _, err := uc.Refresh(context.Background(), "access-token", "ua", "ip")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrInvalidRefreshToken))
	signer.AssertCalled(t, "ParseRefresh", "access-token")
}

func TestAuthUsecase_RefreshUserSoftDeleted(t *testing.T) {
	t.Parallel()
	users := new(mockUserRepo)
	sessions := new(mockSessionStore)
	signer := new(mockSigner)
	uc := newAuthUC(t, users, new(mockCustomerProfileRepo), sessions, new(mockHasher), signer)

	uid := uuid.New()
	signer.On("ParseRefresh", "rt").Return(port.RefreshTokenClaims{Subject: uid.String(), SessionID: "s", JTI: "j"}, nil)
	users.On("GetByID", mock.Anything, uid).Return(nil, apperrors.ErrNotFound)
	sessions.On("DeleteAllForUser", mock.Anything, uid.String()).Return(nil)

	_, _, _, err := uc.Refresh(context.Background(), "rt", "ua", "ip")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrSessionRevoked))
	sessions.AssertNotCalled(t, "Rotate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthUsecase_LoginSuccess(t *testing.T) {
	t.Parallel()
	users := new(mockUserRepo)
	hasher := new(mockHasher)
	signer := new(mockSigner)
	sessions := new(mockSessionStore)
	uc := newAuthUC(t, users, new(mockCustomerProfileRepo), sessions, hasher, signer)

	user := &domainuser.User{ID: uuid.New(), Email: "a@b.com", PasswordHash: "hash", Role: domainuser.RoleCustomer, CreatedAt: time.Now()}
	users.On("GetByEmail", mock.Anything, "a@b.com").Return(user, nil)
	hasher.On("Verify", "hash", "Password1").Return(nil)
	sessions.On("Create", mock.Anything, user.ID.String(), mock.AnythingOfType("string"), mock.Anything).Return(nil)
	signer.On("SignAccess", mock.Anything).Return("at", nil)
	signer.On("SignRefresh", mock.Anything).Return("rt", nil)

	at, rt, u, err := uc.Login(context.Background(), dto.LoginRequest{Email: "a@b.com", Password: "Password1"}, "ua", "ip")
	require.NoError(t, err)
	assert.Equal(t, "at", at)
	assert.Equal(t, "rt", rt)
	assert.Equal(t, user.ID, u.ID)
}

func TestAuthUsecase_LogoutHybridBearer(t *testing.T) {
	t.Parallel()
	sessions := new(mockSessionStore)
	uc := newAuthUC(t, new(mockUserRepo), new(mockCustomerProfileRepo), sessions, new(mockHasher), new(mockSigner))

	uid := uuid.New()
	sid := uuid.New()
	sessions.On("Delete", mock.Anything, uid.String(), sid.String()).Return(nil)

	source := uc.LogoutHybrid(context.Background(), uid, sid, true, "refresh-cookie-ignored")
	assert.Equal(t, usecase.LogoutSourceBearer, source)
	sessions.AssertCalled(t, "Delete", mock.Anything, uid.String(), sid.String())
}

func TestAuthUsecase_LogoutHybridCookieFallback(t *testing.T) {
	t.Parallel()
	sessions := new(mockSessionStore)
	signer := new(mockSigner)
	uc := newAuthUC(t, new(mockUserRepo), new(mockCustomerProfileRepo), sessions, new(mockHasher), signer)

	uid := uuid.NewString()
	sid := uuid.NewString()
	signer.On("ParseRefresh", "rt").Return(port.RefreshTokenClaims{Subject: uid, SessionID: sid, JTI: "j1"}, nil)
	sessions.On("Delete", mock.Anything, uid, sid).Return(nil)

	source := uc.LogoutHybrid(context.Background(), uuid.Nil, uuid.Nil, false, "rt")
	assert.Equal(t, usecase.LogoutSourceCookie, source)
}

func TestAuthUsecase_LogoutHybridBearerPriorityOverCookie(t *testing.T) {
	t.Parallel()
	sessions := new(mockSessionStore)
	signer := new(mockSigner)
	uc := newAuthUC(t, new(mockUserRepo), new(mockCustomerProfileRepo), sessions, new(mockHasher), signer)

	uid := uuid.New()
	sid := uuid.New()
	sessions.On("Delete", mock.Anything, uid.String(), sid.String()).Return(nil)

	source := uc.LogoutHybrid(context.Background(), uid, sid, true, "rt-should-not-parse")
	assert.Equal(t, usecase.LogoutSourceBearer, source)
	signer.AssertNotCalled(t, "ParseRefresh", mock.Anything)
}

func TestAuthUsecase_LogoutHybridNone(t *testing.T) {
	t.Parallel()
	uc := newAuthUC(t, new(mockUserRepo), new(mockCustomerProfileRepo), new(mockSessionStore), new(mockHasher), new(mockSigner))

	source := uc.LogoutHybrid(context.Background(), uuid.Nil, uuid.Nil, false, "")
	assert.Equal(t, usecase.LogoutSourceNone, source)
}

func TestAuthUsecase_LogoutHybridInvalidCookie(t *testing.T) {
	t.Parallel()
	signer := new(mockSigner)
	uc := newAuthUC(t, new(mockUserRepo), new(mockCustomerProfileRepo), new(mockSessionStore), new(mockHasher), signer)

	signer.On("ParseRefresh", "bad").Return(port.RefreshTokenClaims{}, apperrors.ErrUnauthorized)

	source := uc.LogoutHybrid(context.Background(), uuid.Nil, uuid.Nil, false, "bad")
	assert.Equal(t, usecase.LogoutSourceNone, source)
}
