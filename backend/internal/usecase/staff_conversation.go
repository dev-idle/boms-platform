package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	domainconversation "github.com/boms/backend/internal/domain/conversation"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
)

// StaffConversationUsecase is the counter's inbox: every customer writing
// about an order, answered by whoever is at the counter.
type StaffConversationUsecase struct {
	users         port.UserRepository
	orders        port.OrderRepository
	conversations port.ConversationRepository
	tx            port.TxManager
	events        port.EventOutbox
}

func NewStaffConversationUsecase(
	users port.UserRepository,
	orders port.OrderRepository,
	conversations port.ConversationRepository,
	tx port.TxManager,
	events port.EventOutbox,
) *StaffConversationUsecase {
	return &StaffConversationUsecase{users: users, orders: orders, conversations: conversations, tx: tx, events: events}
}

// List returns a page of the inbox, latest message first. statusFilter
// narrows it to open or closed conversations.
func (u *StaffConversationUsecase) List(
	ctx context.Context,
	page, pageSize int32,
	statusFilter string,
) ([]dto.StaffInboxConversationResponse, int64, int32, int32, error) {
	page, pageSize = normalizeOrderListPage(page, pageSize)
	var status *domainconversation.Status
	if trimmed := strings.TrimSpace(statusFilter); trimmed != "" {
		parsed := domainconversation.Status(trimmed)
		if !parsed.Valid() {
			return nil, 0, page, pageSize, apperrors.ErrValidation.WithDetail("status", "must be open or closed")
		}
		status = &parsed
	}
	rows, total, err := listWithTotal(ctx,
		func(ctx context.Context) ([]port.StaffConversation, error) {
			return u.conversations.StaffList(ctx, port.StaffListConversationsParams{
				Status: status,
				Limit:  pageSize,
				Offset: utils.PageOffset(page, pageSize),
			})
		},
		func(ctx context.Context) (int64, error) {
			return u.conversations.StaffListCount(ctx, status)
		},
	)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]dto.StaffInboxConversationResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.StaffInboxConversationResponse{
			OrderID:       row.OrderID.String(),
			OrderCode:     row.OrderCode,
			Status:        string(row.Status),
			Unread:        row.UnreadByStaff,
			LastMessageAt: row.LastMessageAt,
			CustomerName:  row.CustomerName,
			CustomerEmail: row.CustomerEmail,
			Preview:       row.LastBody,
		})
	}
	return out, total, page, pageSize, nil
}

// Counts returns how many conversations are open and how many hold messages
// nobody at the counter has read.
func (u *StaffConversationUsecase) Counts(ctx context.Context) (*dto.StaffConversationCountsResponse, error) {
	counts, err := u.conversations.StaffCounts(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.StaffConversationCountsResponse{Open: counts.Open, Unread: counts.Unread}, nil
}

// Thread returns a page of the messages on an order: the latest, or those
// before beforeID.
func (u *StaffConversationUsecase) Thread(ctx context.Context, orderID uuid.UUID, beforeID *uuid.UUID) (*dto.StaffThreadResponse, error) {
	if err := u.checkOrder(ctx, orderID); err != nil {
		return nil, err
	}
	conversation, messages, hasMore, err := readThread(ctx, u.conversations, orderID, beforeID)
	if err != nil {
		return nil, err
	}
	resp := &dto.StaffThreadResponse{Messages: toMessageResponses(messages, true), HasMore: hasMore}
	if conversation != nil {
		resp.Conversation = toStaffConversationResponse(*conversation)
	}
	return resp, nil
}

// Post writes a staff member's message on an order. The writer takes the
// conversation over, and the customer hears of it at once.
func (u *StaffConversationUsecase) Post(
	ctx context.Context,
	actorID, orderID uuid.UUID,
	req dto.PostMessageRequest,
) (*dto.MessageResponse, error) {
	body, err := domainconversation.NewBody(req.Body)
	if err != nil {
		return nil, err
	}
	row, err := u.order(ctx, orderID)
	if err != nil {
		return nil, err
	}
	var message *domainconversation.Message
	err = u.tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := holdCustomer(txCtx, u.users, *row.Order.UserID); err != nil {
			return err
		}
		conversationID, err := u.conversations.OpenForStaffMessage(txCtx, orderID, actorID)
		if err != nil {
			return err
		}
		if message, err = u.conversations.CreateMessage(txCtx, conversationID, actorID, domainuser.RoleStaff, body); err != nil {
			return err
		}
		return u.events.Add(txCtx, domainconversation.MessageCreatedEvent(orderID, *row.Order.UserID))
	})
	if err != nil {
		return nil, err
	}
	resp := toMessageResponse(*message, true)
	return &resp, nil
}

// MarkRead records that the counter read every message on an order.
func (u *StaffConversationUsecase) MarkRead(ctx context.Context, orderID uuid.UUID) error {
	if err := u.checkOrder(ctx, orderID); err != nil {
		return err
	}
	return u.tx.WithTx(ctx, func(txCtx context.Context) error {
		read, err := u.conversations.MarkReadByStaff(txCtx, orderID)
		if err != nil || !read {
			return err
		}
		return u.events.Add(txCtx, domainconversation.ChangedAtCounterEvent(orderID))
	})
}

// SetStatus resolves an order's conversation, or opens it again.
func (u *StaffConversationUsecase) SetStatus(
	ctx context.Context,
	orderID uuid.UUID,
	req dto.PatchConversationRequest,
) (*dto.StaffConversationResponse, error) {
	status := domainconversation.Status(req.Status)
	if !status.Valid() {
		return nil, apperrors.ErrValidation.WithDetail("status", "must be open or closed")
	}
	if err := u.checkOrder(ctx, orderID); err != nil {
		return nil, err
	}
	var conversation *domainconversation.Conversation
	err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		changed, err := u.conversations.SetStatus(txCtx, orderID, status)
		if err != nil {
			return err
		}
		if conversation, err = u.conversations.GetByOrderID(txCtx, orderID); err != nil {
			return err
		}
		if !changed {
			return nil
		}
		return u.events.Add(txCtx, domainconversation.ChangedAtCounterEvent(orderID))
	})
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, domainconversation.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toStaffConversationResponse(*conversation), nil
}

func (u *StaffConversationUsecase) checkOrder(ctx context.Context, orderID uuid.UUID) error {
	_, err := u.order(ctx, orderID)
	return err
}

// order reads an order the counter may write about: one the bakery sees,
// placed from an account. An order of a closed account is not found.
func (u *StaffConversationUsecase) order(ctx context.Context, orderID uuid.UUID) (*port.StaffOrderListRow, error) {
	row, err := u.orders.StaffGetByID(ctx, orderID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, domainorder.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	seen, err := seenByStaff(ctx, u.orders, row.Order)
	if err != nil {
		return nil, err
	}
	if !seen {
		return nil, domainorder.ErrNotFound
	}
	if row.Order.UserID == nil {
		return nil, domainconversation.ErrUnavailable
	}
	return row, nil
}

func toStaffConversationResponse(conversation domainconversation.Conversation) *dto.StaffConversationResponse {
	return &dto.StaffConversationResponse{
		Status:            string(conversation.Status),
		Unread:            conversation.UnreadByStaff,
		AssignedStaffName: conversation.AssignedStaffName,
	}
}
