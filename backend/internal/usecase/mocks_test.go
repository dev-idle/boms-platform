package usecase_test

// Hand-written testify mocks shared by every usecase test. The assertions below
// make a port change fail the build here, naming the mock that drifted, instead
// of surfacing as a runtime panic in whichever test first calls the new method.

import (
	"context"

	domainaccount "github.com/boms/backend/internal/domain/account"
	domainpolicy "github.com/boms/backend/internal/domain/policy"
	domainprofile "github.com/boms/backend/internal/domain/profile"
	domainsession "github.com/boms/backend/internal/domain/session"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

var (
	_ port.UserRepository            = (*mockUserRepo)(nil)
	_ port.CustomerProfileRepository = (*mockCustomerProfileRepo)(nil)
	_ port.StaffProfileRepository    = (*mockStaffProfileRepo)(nil)
	_ port.TxManager                 = passthroughTxManager{}
	_ port.SessionStore              = (*mockSessionStore)(nil)
	_ port.PasswordHasher            = (*mockHasher)(nil)
	_ port.TokenSigner               = (*mockSigner)(nil)
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
func (m *mockUserRepo) FindCustomerByEmail(ctx context.Context, email string) (*port.CustomerContact, error) {
	args := m.Called(ctx, email)
	c, _ := args.Get(0).(*port.CustomerContact)
	return c, args.Error(1)
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
func (m *mockUserRepo) GetByIDForShare(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
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
func (m *mockUserRepo) AdminGetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
	args := m.Called(ctx, id)
	out, _ := args.Get(0).(*domainuser.User)
	return out, args.Error(1)
}
func (m *mockUserRepo) Restore(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepo) AdminUpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return m.Called(ctx, id, passwordHash).Error(0)
}
func (m *mockUserRepo) AdminList(ctx context.Context, params port.AdminListUsersParams) ([]port.AdminListUser, error) {
	args := m.Called(ctx, params)
	rows, _ := args.Get(0).([]port.AdminListUser)
	return rows, args.Error(1)
}
func (m *mockUserRepo) AdminListCount(ctx context.Context, search string, role *domainuser.Role) (int64, error) {
	args := m.Called(ctx, search, role)
	total, _ := args.Get(0).(int64)
	return total, args.Error(1)
}
func (m *mockUserRepo) ClaimPhone(ctx context.Context, phone string, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, phone, userID)
	return args.Bool(0), args.Error(1)
}
func (m *mockUserRepo) AdminGetByIDForUpdate(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
	args := m.Called(ctx, id)
	out, _ := args.Get(0).(*domainuser.User)
	return out, args.Error(1)
}
func (m *mockUserRepo) ReleasePhone(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockUserRepo) Erase(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockUserRepo) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockUserRepo) ResetPassword(ctx context.Context, id uuid.UUID, hash string) error {
	return m.Called(ctx, id, hash).Error(0)
}
func (m *mockUserRepo) BumpSessionVersion(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockUserRepo) TermsAcceptance(ctx context.Context, userID uuid.UUID) (*domainpolicy.Acceptance, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainpolicy.Acceptance), args.Error(1)
}

type mockCustomerProfileRepo struct{ mock.Mock }

func (m *mockCustomerProfileRepo) Create(ctx context.Context, params port.UpsertCustomerProfileParams) (*domainprofile.Customer, error) {
	args := m.Called(ctx, params)
	out, _ := args.Get(0).(*domainprofile.Customer)
	return out, args.Error(1)
}
func (m *mockCustomerProfileRepo) Erase(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
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

type mockStaffProfileRepo struct{ mock.Mock }

func (m *mockStaffProfileRepo) Create(ctx context.Context, params port.UpsertStaffProfileParams) (*domainprofile.Staff, error) {
	args := m.Called(ctx, params)
	out, _ := args.Get(0).(*domainprofile.Staff)
	return out, args.Error(1)
}
func (m *mockStaffProfileRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domainprofile.Staff, error) {
	args := m.Called(ctx, userID)
	out, _ := args.Get(0).(*domainprofile.Staff)
	return out, args.Error(1)
}
func (m *mockStaffProfileRepo) UpdateByUserID(ctx context.Context, params port.UpsertStaffProfileParams) (*domainprofile.Staff, error) {
	args := m.Called(ctx, params)
	out, _ := args.Get(0).(*domainprofile.Staff)
	return out, args.Error(1)
}
func (m *mockStaffProfileRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *mockStaffProfileRepo) NextEmployeeCode(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

type passthroughTxManager struct{}

func (passthroughTxManager) WithTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	return fn(ctx)
}

// recordingAuditLogs keeps the actions written to the audit trail and whether
// it was scrubbed; any other call panics.
type recordingAuditLogs struct {
	port.AuditLogRepository
	failWith error
	actions  []domainuser.AuditAction
	scrubbed bool
}

func (a *recordingAuditLogs) Create(_ context.Context, params port.CreateAuditLogParams) error {
	if a.failWith != nil {
		return a.failWith
	}
	a.actions = append(a.actions, params.Action)
	return nil
}

func (a *recordingAuditLogs) ScrubSubject(context.Context, uuid.UUID) error {
	a.scrubbed = true
	return nil
}

// fakeTokens stands in for the link token store: Issue hands out next,
// Redeem answers owner (or redeemErr), and every call is recorded.
type fakeTokens struct {
	next      string
	issueErr  error
	issued    []domainaccount.Purpose
	owner     uuid.UUID
	redeemErr error
	redeemed  []string
	deleteErr error
	deleted   []uuid.UUID
}

func (f *fakeTokens) Issue(_ context.Context, _ uuid.UUID, purpose domainaccount.Purpose) (string, error) {
	if f.issueErr != nil {
		return "", f.issueErr
	}
	f.issued = append(f.issued, purpose)
	return f.next, nil
}

func (f *fakeTokens) Redeem(_ context.Context, token string, _ domainaccount.Purpose) (uuid.UUID, error) {
	f.redeemed = append(f.redeemed, token)
	if f.redeemErr != nil {
		return uuid.Nil, f.redeemErr
	}
	return f.owner, nil
}

func (f *fakeTokens) DeleteForUser(_ context.Context, userID uuid.UUID) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deleted = append(f.deleted, userID)
	return nil
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
