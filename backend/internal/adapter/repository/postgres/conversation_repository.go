package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainconversation "github.com/boms/backend/internal/domain/conversation"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
)

// ConversationRepository keeps the conversations between customers and the
// counter.
type ConversationRepository struct {
	queries *sqlcgen.Queries
}

func NewConversationRepository(pool *Pool) *ConversationRepository {
	return &ConversationRepository{queries: pool.Queries()}
}

func (r *ConversationRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

// OpenForCustomerMessage implements port.ConversationRepository.
func (r *ConversationRepository) OpenForCustomerMessage(ctx context.Context, thread port.ConversationThread) (uuid.UUID, error) {
	var id uuid.UUID
	var err error
	if thread.OrderID == nil {
		id, err = r.q(ctx).OpenGeneralConversationForCustomerMessage(ctx, thread.CustomerID)
	} else {
		id, err = r.q(ctx).OpenConversationForCustomerMessage(ctx, sqlcgen.OpenConversationForCustomerMessageParams{
			CustomerID: thread.CustomerID,
			OrderID:    thread.OrderID,
		})
	}
	if err != nil {
		return uuid.Nil, mapRepoError(err, "open conversation for customer message")
	}
	return id, nil
}

// OpenForStaffMessage implements port.ConversationRepository.
func (r *ConversationRepository) OpenForStaffMessage(
	ctx context.Context,
	thread port.ConversationThread,
	staffID uuid.UUID,
) (uuid.UUID, error) {
	var id uuid.UUID
	var err error
	if thread.OrderID == nil {
		id, err = r.q(ctx).OpenGeneralConversationForStaffMessage(ctx, sqlcgen.OpenGeneralConversationForStaffMessageParams{
			CustomerID: thread.CustomerID,
			StaffID:    &staffID,
		})
	} else {
		id, err = r.q(ctx).OpenConversationForStaffMessage(ctx, sqlcgen.OpenConversationForStaffMessageParams{
			CustomerID: thread.CustomerID,
			OrderID:    thread.OrderID,
			StaffID:    &staffID,
		})
	}
	if err != nil {
		return uuid.Nil, mapRepoError(err, "open conversation for staff message")
	}
	return id, nil
}

// CreateMessage implements port.ConversationRepository.
func (r *ConversationRepository) CreateMessage(
	ctx context.Context,
	conversationID, authorID uuid.UUID,
	authorRole domainuser.Role,
	body string,
	viaAssistant bool,
) (*domainconversation.Message, error) {
	row, err := r.q(ctx).CreateMessage(ctx, sqlcgen.CreateMessageParams{
		ConversationID: conversationID,
		AuthorID:       authorID,
		AuthorRole:     sqlcgen.UserRole(authorRole),
		Body:           body,
		ViaAssistant:   viaAssistant,
	})
	if err != nil {
		return nil, mapRepoError(err, "create message")
	}
	return &domainconversation.Message{
		ID:         row.ID,
		AuthorRole: domainuser.Role(row.AuthorRole),
		AuthorName: row.AuthorName,
		Body:       row.Body,
		CreatedAt:  row.CreatedAt,
	}, nil
}

// Get implements port.ConversationRepository.
func (r *ConversationRepository) Get(ctx context.Context, thread port.ConversationThread) (*domainconversation.Conversation, error) {
	row, err := r.q(ctx).GetConversation(ctx, sqlcgen.GetConversationParams{
		CustomerID: thread.CustomerID,
		OrderID:    thread.OrderID,
	})
	if err != nil {
		return nil, mapRepoError(err, "get conversation")
	}
	return &domainconversation.Conversation{
		Status:            domainconversation.Status(row.Status),
		UnreadByCustomer:  row.UnreadByCustomer,
		UnreadByStaff:     row.UnreadByStaff,
		AssignedStaffName: row.AssignedStaffName,
	}, nil
}

// GeneralCustomer implements port.ConversationRepository.
func (r *ConversationRepository) GeneralCustomer(ctx context.Context, customerID uuid.UUID) (*port.ConversationCustomer, error) {
	row, err := r.q(ctx).GetGeneralConversationCustomer(ctx, customerID)
	if err != nil {
		return nil, mapRepoError(err, "get general conversation customer")
	}
	return &port.ConversationCustomer{Name: row.CustomerName, Email: row.CustomerEmail}, nil
}

// ListMessages implements port.ConversationRepository.
func (r *ConversationRepository) ListMessages(
	ctx context.Context,
	thread port.ConversationThread,
	beforeID *uuid.UUID,
	limit int32,
) ([]domainconversation.Message, error) {
	rows, err := r.q(ctx).ListMessages(ctx, sqlcgen.ListMessagesParams{
		CustomerID: thread.CustomerID,
		OrderID:    thread.OrderID,
		BeforeID:   beforeID,
		Limit:      limit,
	})
	if err != nil {
		return nil, mapRepoError(err, "list messages")
	}
	out := make([]domainconversation.Message, 0, len(rows))
	for _, row := range rows {
		out = append(out, domainconversation.Message{
			ID:         row.ID,
			AuthorRole: domainuser.Role(row.AuthorRole),
			AuthorName: row.AuthorName,
			Body:       row.Body,
			CreatedAt:  row.CreatedAt,
		})
	}
	return out, nil
}

// MarkReadByCustomer implements port.ConversationRepository.
func (r *ConversationRepository) MarkReadByCustomer(ctx context.Context, thread port.ConversationThread) (bool, error) {
	n, err := r.q(ctx).MarkConversationReadByCustomer(ctx, sqlcgen.MarkConversationReadByCustomerParams{
		CustomerID: thread.CustomerID,
		OrderID:    thread.OrderID,
	})
	if err != nil {
		return false, mapRepoError(err, "mark conversation read by customer")
	}
	return n > 0, nil
}

// MarkReadByStaff implements port.ConversationRepository.
func (r *ConversationRepository) MarkReadByStaff(ctx context.Context, thread port.ConversationThread) (bool, error) {
	n, err := r.q(ctx).MarkConversationReadByStaff(ctx, sqlcgen.MarkConversationReadByStaffParams{
		CustomerID: thread.CustomerID,
		OrderID:    thread.OrderID,
	})
	if err != nil {
		return false, mapRepoError(err, "mark conversation read by staff")
	}
	return n > 0, nil
}

// SetStatus implements port.ConversationRepository.
func (r *ConversationRepository) SetStatus(
	ctx context.Context,
	thread port.ConversationThread,
	status domainconversation.Status,
) (bool, error) {
	n, err := r.q(ctx).SetConversationStatus(ctx, sqlcgen.SetConversationStatusParams{
		Status:     sqlcgen.ConversationStatus(status),
		CustomerID: thread.CustomerID,
		OrderID:    thread.OrderID,
	})
	if err != nil {
		return false, mapRepoError(err, "set conversation status")
	}
	return n > 0, nil
}

// StaffList implements port.ConversationRepository.
func (r *ConversationRepository) StaffList(ctx context.Context, params port.StaffListConversationsParams) ([]port.StaffConversation, error) {
	rows, err := r.q(ctx).StaffListConversations(ctx, sqlcgen.StaffListConversationsParams{
		Status: conversationStatusOrNil(params.Status),
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, mapRepoError(err, "staff list conversations")
	}
	out := make([]port.StaffConversation, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.StaffConversation{
			CustomerID:    row.CustomerID,
			OrderID:       row.OrderID,
			OrderCode:     row.OrderCode,
			Status:        domainconversation.Status(row.Status),
			UnreadByStaff: row.UnreadByStaff,
			LastMessageAt: row.LastMessageAt,
			CustomerName:  row.CustomerName,
			CustomerEmail: row.CustomerEmail,
			LastBody:      row.LastBody,
		})
	}
	return out, nil
}

// StaffListCount implements port.ConversationRepository.
func (r *ConversationRepository) StaffListCount(ctx context.Context, status *domainconversation.Status) (int64, error) {
	n, err := r.q(ctx).StaffListConversationsCount(ctx, conversationStatusOrNil(status))
	if err != nil {
		return 0, mapRepoError(err, "staff count conversations")
	}
	return n, nil
}

// StaffCounts implements port.ConversationRepository.
func (r *ConversationRepository) StaffCounts(ctx context.Context) (port.StaffConversationCounts, error) {
	row, err := r.q(ctx).StaffConversationCounts(ctx)
	if err != nil {
		return port.StaffConversationCounts{}, mapRepoError(err, "staff conversation counts")
	}
	return port.StaffConversationCounts{Open: row.OpenCount, Unread: row.UnreadCount}, nil
}

// UnreadByCustomer implements port.ConversationRepository.
func (r *ConversationRepository) UnreadByCustomer(ctx context.Context, orderIDs []uuid.UUID) (map[uuid.UUID]int32, error) {
	rows, err := r.q(ctx).ListUnreadByCustomer(ctx, orderIDs)
	if err != nil {
		return nil, mapRepoError(err, "list unread by customer")
	}
	out := make(map[uuid.UUID]int32, len(rows))
	for _, row := range rows {
		out[*row.OrderID] = row.UnreadByCustomer
	}
	return out, nil
}

// ListMessagesByOrderIDs implements port.ConversationRepository.
func (r *ConversationRepository) ListMessagesByOrderIDs(
	ctx context.Context,
	orderIDs []uuid.UUID,
) (map[uuid.UUID][]domainconversation.Message, error) {
	rows, err := r.q(ctx).ListMessagesByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, mapRepoError(err, "list messages by orders")
	}
	out := make(map[uuid.UUID][]domainconversation.Message)
	for _, row := range rows {
		out[*row.OrderID] = append(out[*row.OrderID], domainconversation.Message{
			ID:         row.ID,
			AuthorRole: domainuser.Role(row.AuthorRole),
			Body:       row.Body,
			CreatedAt:  row.CreatedAt,
		})
	}
	return out, nil
}

// ListGeneralMessages implements port.ConversationRepository.
func (r *ConversationRepository) ListGeneralMessages(ctx context.Context, customerID uuid.UUID) ([]domainconversation.Message, error) {
	rows, err := r.q(ctx).ListGeneralMessages(ctx, customerID)
	if err != nil {
		return nil, mapRepoError(err, "list general messages")
	}
	out := make([]domainconversation.Message, 0, len(rows))
	for _, row := range rows {
		out = append(out, domainconversation.Message{
			ID:         row.ID,
			AuthorRole: domainuser.Role(row.AuthorRole),
			Body:       row.Body,
			CreatedAt:  row.CreatedAt,
		})
	}
	return out, nil
}

// EraseForCustomer implements port.ConversationRepository.
func (r *ConversationRepository) EraseForCustomer(ctx context.Context, userID uuid.UUID) error {
	if err := r.q(ctx).EraseCustomerMessages(ctx, userID); err != nil {
		return mapRepoError(err, "erase customer messages")
	}
	if err := r.q(ctx).CloseCustomerConversations(ctx, userID); err != nil {
		return mapRepoError(err, "close customer conversations")
	}
	return nil
}

func conversationStatusOrNil(status *domainconversation.Status) *sqlcgen.ConversationStatus {
	if status == nil {
		return nil
	}
	s := sqlcgen.ConversationStatus(*status)
	return &s
}

var _ port.ConversationRepository = (*ConversationRepository)(nil)
