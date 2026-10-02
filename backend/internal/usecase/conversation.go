package usecase

import (
	"context"
	"errors"
	"slices"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	domainconversation "github.com/boms/backend/internal/domain/conversation"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// threadPageSize is how many messages one page of a thread holds.
const threadPageSize int32 = 50

// ConversationUsecase is a customer writing to the counter about one of their
// orders, once the bakery has taken it.
type ConversationUsecase struct {
	users         port.UserRepository
	orders        port.OrderRepository
	conversations port.ConversationRepository
	tx            port.TxManager
	events        port.EventOutbox
}

func NewConversationUsecase(
	users port.UserRepository,
	orders port.OrderRepository,
	conversations port.ConversationRepository,
	tx port.TxManager,
	events port.EventOutbox,
) *ConversationUsecase {
	return &ConversationUsecase{users: users, orders: orders, conversations: conversations, tx: tx, events: events}
}

// Thread returns a page of the messages on the customer's order: the latest,
// or those before beforeID.
func (u *ConversationUsecase) Thread(ctx context.Context, userID, orderID uuid.UUID, beforeID *uuid.UUID) (*dto.ThreadResponse, error) {
	if _, err := u.customerOrder(ctx, userID, orderID); err != nil {
		return nil, err
	}
	conversation, messages, hasMore, err := readThread(ctx, u.conversations, orderID, beforeID)
	if err != nil {
		return nil, err
	}
	resp := &dto.ThreadResponse{Messages: toMessageResponses(messages, false), HasMore: hasMore}
	if conversation != nil {
		resp.Unread = conversation.UnreadByCustomer
	}
	return resp, nil
}

// Post writes the customer's message on their order. The counter hears of it
// at once, and has read nothing of it yet.
func (u *ConversationUsecase) Post(ctx context.Context, userID, orderID uuid.UUID, req dto.PostMessageRequest) (*dto.MessageResponse, error) {
	body, err := domainconversation.NewBody(req.Body)
	if err != nil {
		return nil, err
	}
	order, err := u.customerOrder(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}
	seen, err := seenByStaff(ctx, u.orders, *order)
	if err != nil {
		return nil, err
	}
	if !seen {
		return nil, domainconversation.ErrUnavailable
	}
	var message *domainconversation.Message
	err = u.tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := holdCustomer(txCtx, u.users, userID); err != nil {
			return err
		}
		conversationID, err := u.conversations.OpenForCustomerMessage(txCtx, orderID)
		if err != nil {
			return err
		}
		if message, err = u.conversations.CreateMessage(txCtx, conversationID, userID, domainuser.RoleCustomer, body); err != nil {
			return err
		}
		return u.events.Add(txCtx, domainconversation.MessageCreatedEvent(orderID, userID))
	})
	if err != nil {
		return nil, err
	}
	resp := toMessageResponse(*message, false)
	return &resp, nil
}

// MarkRead records that the customer read every message on their order.
func (u *ConversationUsecase) MarkRead(ctx context.Context, userID, orderID uuid.UUID) error {
	if _, err := u.customerOrder(ctx, userID, orderID); err != nil {
		return err
	}
	return u.tx.WithTx(ctx, func(txCtx context.Context) error {
		read, err := u.conversations.MarkReadByCustomer(txCtx, orderID)
		if err != nil || !read {
			return err
		}
		return u.events.Add(txCtx, domainconversation.ReadByCustomerEvent(orderID, userID))
	})
}

func (u *ConversationUsecase) customerOrder(ctx context.Context, userID, orderID uuid.UUID) (*domainorder.Order, error) {
	order, err := u.orders.GetByIDForUser(ctx, userID, orderID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, domainorder.ErrNotFound
	}
	return order, err
}

// holdCustomer holds the account a message, review or staff report is about
// until it commits, as a checkout holds its customer: erasing the account waits, then
// erases it with the rest; one that waited on an erasure finds no account.
func holdCustomer(txCtx context.Context, users port.UserRepository, customerID uuid.UUID) error {
	_, err := users.GetByIDForShare(txCtx, customerID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return domainorder.ErrNotFound
	}
	return err
}

// seenByStaff reports whether the bakery sees order, which it must for
// anyone to write about it: one cancelled only if the bakery took it first.
func seenByStaff(ctx context.Context, orders port.OrderRepository, order domainorder.Order) (bool, error) {
	if order.Status != domainorder.StatusCancelled {
		return order.Status.VisibleToStaff(), nil
	}
	history, err := orders.ListStatusEvents(ctx, order.ID)
	if err != nil {
		return false, err
	}
	return domainorder.SeenByStaff(order.Status, history), nil
}

// readThread reads where the order's conversation stands, nil before its
// first message, and a page of its messages oldest first, with whether older
// ones exist.
func readThread(
	ctx context.Context,
	conversations port.ConversationRepository,
	orderID uuid.UUID,
	beforeID *uuid.UUID,
) (*domainconversation.Conversation, []domainconversation.Message, bool, error) {
	var conversation *domainconversation.Conversation
	var messages []domainconversation.Message
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		conversation, err = conversations.GetByOrderID(groupCtx, orderID)
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		return err
	})
	group.Go(func() error {
		var err error
		messages, err = conversations.ListMessages(groupCtx, orderID, beforeID, threadPageSize+1)
		return err
	})
	if err := group.Wait(); err != nil {
		return nil, nil, false, err
	}
	hasMore := len(messages) > int(threadPageSize)
	if hasMore {
		messages = messages[:threadPageSize]
	}
	slices.Reverse(messages)
	return conversation, messages, hasMore, nil
}
