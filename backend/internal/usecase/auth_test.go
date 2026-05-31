package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domainprofile "github.com/boms/backend/internal/domain/profile"
	domainsession "github.com/boms/backend/internal/domain/session"
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

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, params port.CreateUserParams) (*domainuser.User, error) {
	args := m.Called(ctx, params)
	u, _ := args.Get(0).(*domainuser.User)
	return u, args.Error(1)
}
func (m *mockUserRepo) AdminCreate(ctx context.Context, params port.CreateUserParams) (*domainuser.User, error) {
	args := m.Called(ctx, params)
	u, _ := args.Get(0).(*domainuser.User)
	return u, args.Error(1)
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	args := m.Called(ctx, email)
	u, _ := args.Get(0).(*domainuser.User)
	return u, args.Error(1)
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
	args := m.Called(ctx, id)
	u, _ := args.Get(0).(*domainuser.User)
	return u, args.Error(1)
}
func (m *mockUserRepo) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
	args := m.Called(ctx, id)
	u, _ := args.Get(0).(*domainuser.User)
	return u, args.Error(1)
}
func (m *mockUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	return m.Called(ctx, id, hash).Error(0)
}
func (m *mockUserRepo) UpdateRole(ctx context.Context, id uuid.UUID, role domainuser.Role) error {
	return m.Called(ctx, id, role).Error(0)
}
func (m *mockUserRepo) SetMustChangePassword(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepo) ClearMustChangePassword(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepo) AdminList(ctx context.Context, params port.AdminListUsersParams) ([]port.AdminListUser, int64, error) {
	args := m.Called(ctx, params)
	rows, _ := args.Get(0).([]port.AdminListUser)
	total, _ := args.Get(1).(int64)
	return rows, total, args.Error(2)
}

type mockCustomerProfileRepo struct{ mock.Mock }

func (m *mockCustomerProfileRepo) Create(ctx context.Context, params port.UpsertCustomerProfileParams) (*domainprofile.Customer, error) {
	args := m.Called(ctx, params)
	out, _ := args.Get(0).(*domainprofile.Customer)
	return out, args.Error(1)
}
func (m *mockCustomerProfileRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domainprofile.Customer, error) {
	args := m.Called(ctx, userID)
	out, _ := args.Get(0).(*domainprofile.Customer)
	return out, args.Error(1)
}
func (m *mockCustomerProfileRepo) UpdateByUserID(ctx context.Context, params port.UpsertCustomerProfileParams) (*domainprofile.Customer, error) {
	args := m.Called(ctx, params)
	out, _ := args.Get(0).(*domainprofile.Customer)
	return out, args.Error(1)
}
func (m *mockCustomerProfileRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

type passthroughTxManager struct{}

func (passthroughTxManager) WithTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	return fn(ctx)
}

type mockSessionStore struct{ mock.Mock }

func (m *mockSessionStore) Create(ctx context.Context, userID, sessionID string, meta domainsession.SessionMeta) error {
	return m.Called(ctx, userID, sessionID, meta).Error(0)
}
func (m *mockSessionStore) Get(ctx context.Context, userID, sessionID string) (domainsession.SessionMeta, error) {
	args := m.Called(ctx, userID, sessionID)
	meta, _ := args.Get(0).(domainsession.SessionMeta)
	return meta, args.Error(1)
}
func (m *mockSessionStore) Delete(ctx context.Context, userID, sessionID string) error {
	return m.Called(ctx, userID, sessionID).Error(0)
}
func (m *mockSessionStore) DeleteAllForUser(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *mockSessionStore) Rotate(ctx context.Context, userID, oldSessionID, newSessionID, expectedRefreshJTI string, meta domainsession.SessionMeta) error {
	return m.Called(ctx, userID, oldSessionID, newSessionID, expectedRefreshJTI, meta).Error(0)
}

type mockHasher struct{ mock.Mock }

func (m *mockHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}
func (m *mockHasher) Verify(encoded, password string) error {
	args := m.Called(encoded, password)
	return args.Error(0)
}
func (m *mockHasher) NeedsRehash(encoded string) bool {
	return m.Called(encoded).Bool(0)
}

type mockSigner struct{ mock.Mock }

func (m *mockSigner) SignAccess(in port.AccessTokenClaims) (string, error) {
	args := m.Called(in)
	return args.String(0), args.Error(1)
}
func (m *mockSigner) SignRefresh(in port.RefreshTokenClaims) (string, error) {
	args := m.Called(in)
	return args.String(0), args.Error(1)
}
func (m *mockSigner) ParseAccess(token string) (port.AccessTokenClaims, error) {
	args := m.Called(token)
	claims, _ := args.Get(0).(port.AccessTokenClaims)
	return claims, args.Error(1)
}
func (m *mockSigner) ParseRefresh(token string) (port.RefreshTokenClaims, error) {
	args := m.Called(token)
	claims, _ := args.Get(0).(port.RefreshTokenClaims)
	return claims, args.Error(1)
}

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

func TestAuthUsecase_LogoutIdempotent(t *testing.T) {
	t.Parallel()
	sessions := new(mockSessionStore)
	uc := newAuthUC(t, new(mockUserRepo), new(mockCustomerProfileRepo), sessions, new(mockHasher), new(mockSigner))

	sessions.On("Delete", mock.Anything, "u1", "s1").Return(nil)
	require.NoError(t, uc.Logout(context.Background(), "u1", "s1"))
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
