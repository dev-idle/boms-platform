package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domaincart "github.com/boms/backend/internal/domain/cart"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// AccountErasureUsecase erases a customer's account at their request — the
// right to erasure. Once nothing they ordered is still to be made or handed
// over, their personal details go for good: the account is closed, its email
// and profile cleared, the cart and their saved lists emptied, the messages
// about their orders and their reviews erased and the personal data taken out
// of the audit trail. Their orders stay as the bakery's anonymous sales
// records, which accounting law requires it to keep.
type AccountErasureUsecase struct {
	tx            port.TxManager
	users         port.UserRepository
	customers     port.CustomerProfileRepository
	carts         port.CartRepository
	orders        port.OrderRepository
	conversations port.ConversationRepository
	saved         port.SavedProductRepository
	reviews       port.ReviewRepository
	scrubber      port.AuditScrubber
	tokens        port.UserTokenRepository
	sessions      port.SessionStore
	audit         *auditlogger.Service
	hasher        port.PasswordHasher
}

func NewAccountErasureUsecase(
	tx port.TxManager,
	users port.UserRepository,
	customers port.CustomerProfileRepository,
	carts port.CartRepository,
	orders port.OrderRepository,
	conversations port.ConversationRepository,
	saved port.SavedProductRepository,
	reviews port.ReviewRepository,
	scrubber port.AuditScrubber,
	tokens port.UserTokenRepository,
	sessions port.SessionStore,
	audit *auditlogger.Service,
	hasher port.PasswordHasher,
) *AccountErasureUsecase {
	return &AccountErasureUsecase{
		tx:            tx,
		users:         users,
		customers:     customers,
		carts:         carts,
		orders:        orders,
		conversations: conversations,
		saved:         saved,
		reviews:       reviews,
		scrubber:      scrubber,
		tokens:        tokens,
		sessions:      sessions,
		audit:         audit,
		hasher:        hasher,
	}
}

// Erase erases the signed-in customer's account in one transaction, then ends
// every session it has. Staff accounts are closed by an administrator instead.
//
// Erasure cannot be undone, so it asks for the password again: a session left
// open on a shared device is not enough to erase someone's account.
func (u *AccountErasureUsecase) Erase(ctx context.Context, userID uuid.UUID, password string) error {
	user, err := u.users.GetByID(ctx, userID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return u.endSessionsOfErased(ctx, userID)
	}
	if err != nil {
		return err
	}
	if user.Role != domainuser.RoleCustomer {
		return domainuser.ErrSelfDeleteCustomerOnly
	}
	if err := u.hasher.Verify(user.PasswordHash, password); err != nil {
		return apperrors.ErrInvalidCredentials
	}
	if err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		return u.erase(txCtx, user)
	}); err != nil {
		return err
	}
	return u.sessions.DeleteAllForUser(ctx, userID.String())
}

// endSessionsOfErased answers a caller whose account is already closed. When
// an erasure committed but ending its sessions failed, the retry ends them now,
// so asking again always finishes the job; any other closed account is gone.
func (u *AccountErasureUsecase) endSessionsOfErased(ctx context.Context, userID uuid.UUID) error {
	account, err := u.users.AdminGetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return ErrMeNotFound
		}
		return err
	}
	if !account.Erased() {
		return ErrMeNotFound
	}
	return u.sessions.DeleteAllForUser(ctx, userID.String())
}

func (u *AccountErasureUsecase) erase(txCtx context.Context, user *domainuser.User) error {
	// Locks in the order checkout takes them — the cart, then the account — so
	// the two never deadlock. A checkout or profile change already running holds
	// the account, and the check below waits to see its order; one that starts
	// now waits on the cart and finds it emptied, or on the account and finds it
	// closed. Emailed links go before the account, the order redeeming one
	// takes them.
	cart, err := u.carts.GetByUserIDForUpdate(txCtx, user.ID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return err
	}
	if err := u.tokens.DeleteForUser(txCtx, user.ID); err != nil {
		return err
	}
	if _, err := u.users.GetByIDForUpdate(txCtx, user.ID); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return ErrMeNotFound
		}
		return err
	}
	open, err := u.orders.HasOpen(txCtx, user.ID)
	if err != nil {
		return err
	}
	if open {
		return domainuser.ErrAccountHasOpenOrders
	}
	if err := u.emptyCart(txCtx, cart); err != nil {
		return err
	}
	if err := u.conversations.EraseForCustomer(txCtx, user.ID); err != nil {
		return err
	}
	if err := u.orders.EraseIncidentNotes(txCtx, user.ID); err != nil {
		return err
	}
	if err := u.saved.RemoveAll(txCtx, user.ID); err != nil {
		return err
	}
	if err := u.reviews.EraseForCustomer(txCtx, user.ID); err != nil {
		return err
	}
	if err := u.customers.Erase(txCtx, user.ID); err != nil {
		return err
	}
	if err := u.users.Erase(txCtx, user.ID); err != nil {
		return err
	}
	// The erasure itself stays on record — it is the proof the request was
	// honoured, so it commits with the erasure or not at all. The scrub below
	// removes the network and browser it came from, like those of every other
	// action of the account.
	if err := u.audit.Log(txCtx, domainuser.AuditActionMeErasedAccount, user.ID, user.Role, &user.ID, "user", nil, nil); err != nil {
		return apperrors.Errorf("record erasure: %w", err)
	}
	return u.scrubber.ScrubSubject(txCtx, user.ID)
}

func (u *AccountErasureUsecase) emptyCart(txCtx context.Context, cart *domaincart.Cart) error {
	if cart == nil {
		return nil
	}
	if err := u.carts.DeleteAllItems(txCtx, cart.ID); err != nil {
		return err
	}
	return u.carts.ClearDiscountCode(txCtx, cart.ID)
}
