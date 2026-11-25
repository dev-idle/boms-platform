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
// orders, once the bakery has taken it, or about anything else in their
// general conversation.
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

// repliesBatch is how many conversations one run emails about; the next run
// takes the rest.
const repliesBatch = 50

// EmailUnreadReplies asks for an email to each customer who has left the
// counter's replies unread for domainconversation.RepliesUnreadEmailAfter —
// once until they read them — a batch at a time, and returns how many it asked
// for. Each conversation stands alone: one that fails is reported with the
// rest while the others go on.
func (u *ConversationUsecase) EmailUnreadReplies(ctx context.Context) (int, error) {
	due, err := u.conversations.ListUnreadRepliesDue(ctx, domainconversation.RepliesUnreadEmailAfter, repliesBatch)
	if err != nil {
		return 0, err
	}
	asked := 0
	var failures []error
	for _, id := range due {
		marked := false
		err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
			var err error
			// Not marked: read meanwhile, or another run took it.
			if marked, err = u.conversations.MarkRepliesEmailed(txCtx, id); err != nil || !marked {
				return err
			}
			return u.events.Add(txCtx, domainconversation.RepliesUnreadEvent(id))
		})
		if err != nil {
			failures = append(failures, apperrors.Errorf("email unread replies of %s: %w", id, err))
			continue
		}
		if marked {
			asked++
		}
	}
	return asked, errors.Join(failures...)
}

// Thread returns a page of the customer's messages about their order, or of
// their general conversation when orderID is nil: the latest, or those
// before beforeID.
func (u *ConversationUsecase) Thread(
	ctx context.Context,
	userID uuid.UUID,
	orderID, beforeID *uuid.UUID,
) (*dto.ThreadResponse, error) {
	thread, err := u.thread(ctx, userID, orderID, false)
	if err != nil {
		return nil, err
	}
	conversation, messages, hasMore, err := readThread(ctx, u.conversations, thread, beforeID)
	if err != nil {
		return nil, err
	}
	resp := &dto.ThreadResponse{Messages: toMessageResponses(messages, false), HasMore: hasMore}
	if conversation != nil {
		resp.Unread = conversation.UnreadByCustomer
	}
	return resp, nil
}

// Post writes the customer's message about their order, or on their general
// conversation when orderID is nil. The counter hears of it at once, and has
// read nothing of it yet.
func (u *ConversationUsecase) Post(
	ctx context.Context,
	userID uuid.UUID,
	orderID *uuid.UUID,
	req dto.PostMessageRequest,
) (*dto.MessageResponse, error) {
	return u.post(ctx, userID, orderID, req, false)
}

// post writes a customer's message; viaAssistant marks the questions the
// assistant passed on, which the engagement report does not count as the
// customer writing to the counter.
func (u *ConversationUsecase) post(
	ctx context.Context,
	userID uuid.UUID,
	orderID *uuid.UUID,
	req dto.PostMessageRequest,
	viaAssistant bool,
) (*dto.MessageResponse, error) {
	body, err := domainconversation.NewBody(req.Body)
	if err != nil {
		return nil, err
	}
	thread, err := u.thread(ctx, userID, orderID, true)
	if err != nil {
		return nil, err
	}
	var message *domainconversation.Message
	err = u.tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := holdCustomer(txCtx, u.users, userID); err != nil {
			return err
		}
		conversationID, err := u.conversations.OpenForCustomerMessage(txCtx, thread)
		if err != nil {
			return err
		}
		if message, err = u.conversations.CreateMessage(txCtx, conversationID, userID, domainuser.RoleCustomer, body, viaAssistant); err != nil {
			return err
		}
		return u.events.Add(txCtx, domainconversation.MessageCreatedEvent(userID, orderID))
	})
	if err != nil {
		return nil, err
	}
	resp := toMessageResponse(*message, false)
	return &resp, nil
}

// MarkRead records that the customer read every message about their order,
// or of their general conversation when orderID is nil.
func (u *ConversationUsecase) MarkRead(ctx context.Context, userID uuid.UUID, orderID *uuid.UUID) error {
	thread, err := u.thread(ctx, userID, orderID, false)
	if err != nil {
		return err
	}
	return u.tx.WithTx(ctx, func(txCtx context.Context) error {
		read, err := u.conversations.MarkReadByCustomer(txCtx, thread)
		if err != nil || !read {
			return err
		}
		return u.events.Add(txCtx, domainconversation.ReadByCustomerEvent(userID, orderID))
	})
}

// thread names the customer's conversation about orderID — one of their
// orders, which they may write about only once the bakery has taken it — or
// their general conversation when orderID is nil.
func (u *ConversationUsecase) thread(
	ctx context.Context,
	userID uuid.UUID,
	orderID *uuid.UUID,
	writing bool,
) (port.ConversationThread, error) {
	thread := port.ConversationThread{CustomerID: userID, OrderID: orderID}
	if orderID == nil {
		return thread, nil
	}
	order, err := u.customerOrder(ctx, userID, *orderID)
	if err != nil || !writing {
		return thread, err
	}
	seen, err := seenByStaff(ctx, u.orders, *order)
	if err != nil {
		return thread, err
	}
	if !seen {
		return thread, domainconversation.ErrUnavailable
	}
	return thread, nil
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

// readThread reads where the conversation stands, nil before its first
// message, and a page of its messages oldest first, with whether older ones
// exist.
func readThread(
	ctx context.Context,
	conversations port.ConversationRepository,
	thread port.ConversationThread,
	beforeID *uuid.UUID,
) (*domainconversation.Conversation, []domainconversation.Message, bool, error) {
	var conversation *domainconversation.Conversation
	var messages []domainconversation.Message
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		conversation, err = conversations.Get(groupCtx, thread)
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		return err
	})
	group.Go(func() error {
		var err error
		messages, err = conversations.ListMessages(groupCtx, thread, beforeID, threadPageSize+1)
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
