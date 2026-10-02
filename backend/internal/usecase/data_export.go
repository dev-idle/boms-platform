package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	domaincart "github.com/boms/backend/internal/domain/cart"
	domainconversation "github.com/boms/backend/internal/domain/conversation"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainpolicy "github.com/boms/backend/internal/domain/policy"
	domainsession "github.com/boms/backend/internal/domain/session"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/profilesvc"
	"github.com/boms/backend/internal/shared/ctxmeta"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// dataExportPage is how many rows one export read takes: an export holds every
// order and every recorded change, read a bounded keyset page at a time.
const dataExportPage int32 = 100

// DataExportUsecase gathers what the bakery holds about one person for them to
// download (the right of access): their account and profile, the policies they
// accepted, their sign-in sessions, the changes recorded to their account and,
// for a customer, their cart, their favorites and wishlist, their reviews, and
// every order with its lines, history, messages and incidents.
type DataExportUsecase struct {
	users         port.UserRepository
	profiles      *profilesvc.Service
	orders        port.OrderRepository
	conversations port.ConversationRepository
	saved         port.SavedProductRepository
	reviews       port.ReviewRepository
	carts         port.CartRepository
	sessions      port.SessionLister
	activity      port.AccountActivityReader
}

func NewDataExportUsecase(
	users port.UserRepository,
	customers port.CustomerProfileRepository,
	staff port.StaffProfileRepository,
	admins port.AdminProfileRepository,
	orders port.OrderRepository,
	conversations port.ConversationRepository,
	saved port.SavedProductRepository,
	reviews port.ReviewRepository,
	carts port.CartRepository,
	sessions port.SessionLister,
	activity port.AccountActivityReader,
) *DataExportUsecase {
	return &DataExportUsecase{
		users:         users,
		profiles:      profilesvc.NewService(customers, staff, admins),
		orders:        orders,
		conversations: conversations,
		saved:         saved,
		reviews:       reviews,
		carts:         carts,
		sessions:      sessions,
		activity:      activity,
	}
}

// DataExport is one person's data before it is written out; the account and
// profile are shaped as GET /me shapes them.
type DataExport struct {
	User     *domainuser.User
	Profile  any
	Terms    *dto.TermsAcceptanceResponse
	Sessions []dto.DataExportSessionResponse
	Activity []dto.AccountActivityResponse
	Cart     []dto.DataExportCartItemResponse
	Saved    []dto.DataExportSavedProductResponse
	Reviews  []dto.DataExportReviewResponse
	Orders   []dto.DataExportOrderResponse
}

// Export reads the person's data, every part at once; like listWithTotal it
// refuses to run inside a transaction, whose single connection cannot serve
// concurrent reads.
func (u *DataExportUsecase) Export(ctx context.Context, userID uuid.UUID) (*DataExport, error) {
	if ctxmeta.InTransaction(ctx) {
		return nil, apperrors.Errorf("data export: cannot run inside a transaction")
	}
	user, err := u.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, ErrMeNotFound
		}
		return nil, err
	}
	out := &DataExport{
		User:    user,
		Cart:    []dto.DataExportCartItemResponse{},
		Saved:   []dto.DataExportSavedProductResponse{},
		Reviews: []dto.DataExportReviewResponse{},
		Orders:  []dto.DataExportOrderResponse{},
	}
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		profile, err := u.profiles.GetByUserID(groupCtx, user.ID, user.Role)
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		out.Profile = profile
		return err
	})
	group.Go(func() error {
		terms, err := u.users.TermsAcceptance(groupCtx, user.ID)
		out.Terms = termsAcceptanceToDTO(terms)
		return err
	})
	group.Go(func() error {
		sessions, err := u.sessions.ListForUser(groupCtx, user.ID.String())
		out.Sessions = mapExportSessions(sessions)
		return err
	})
	group.Go(func() error {
		activity, err := u.exportActivity(groupCtx, user.ID)
		out.Activity = activity
		return err
	})
	if user.Role == domainuser.RoleCustomer {
		group.Go(func() error {
			cart, err := u.exportCart(groupCtx, user.ID)
			out.Cart = cart
			return err
		})
		group.Go(func() error {
			saved, err := u.saved.ListForExport(groupCtx, user.ID)
			out.Saved = mapExportSavedProducts(saved)
			return err
		})
		group.Go(func() error {
			reviews, err := u.reviews.ListForExport(groupCtx, user.ID)
			out.Reviews = mapExportReviews(reviews)
			return err
		})
		group.Go(func() error {
			orders, err := u.exportOrders(groupCtx, user.ID)
			out.Orders = orders
			return err
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}

// exportOrders reads every order the customer placed, newest first, by keyset,
// with each page's lines, histories, messages and incidents read at once.
func (u *DataExportUsecase) exportOrders(ctx context.Context, userID uuid.UUID) ([]dto.DataExportOrderResponse, error) {
	out := make([]dto.DataExportOrderResponse, 0)
	var before *port.PageCursor
	for {
		orders, err := u.orders.ListByUserBefore(ctx, userID, before, dataExportPage)
		if err != nil {
			return nil, err
		}
		if len(orders) == 0 {
			return out, nil
		}
		page, err := u.exportOrderPage(ctx, orders)
		if err != nil {
			return nil, err
		}
		out = append(out, page...)
		if len(orders) < int(dataExportPage) {
			return out, nil
		}
		last := orders[len(orders)-1]
		before = &port.PageCursor{At: last.CreatedAt, ID: last.ID}
	}
}

func (u *DataExportUsecase) exportOrderPage(ctx context.Context, orders []domainorder.Order) ([]dto.DataExportOrderResponse, error) {
	ids := make([]uuid.UUID, 0, len(orders))
	for _, order := range orders {
		ids = append(ids, order.ID)
	}
	var items map[uuid.UUID][]domainorder.Item
	var timelines map[uuid.UUID][]domainorder.StatusEvent
	var messages map[uuid.UUID][]domainconversation.Message
	var incidents map[uuid.UUID][]domainorder.Incident
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		items, err = u.orders.ListItemsByOrderIDs(groupCtx, ids)
		return err
	})
	group.Go(func() error {
		var err error
		timelines, err = u.orders.ListStatusEventsByOrderIDs(groupCtx, ids)
		return err
	})
	group.Go(func() error {
		var err error
		messages, err = u.conversations.ListMessagesByOrderIDs(groupCtx, ids)
		return err
	})
	group.Go(func() error {
		var err error
		incidents, err = u.orders.ListIncidentsByOrderIDs(groupCtx, ids)
		return err
	})
	if err := group.Wait(); err != nil {
		return nil, err
	}
	out := make([]dto.DataExportOrderResponse, 0, len(orders))
	for _, order := range orders {
		out = append(out, dto.DataExportOrderResponse{
			ID:                   order.ID.String(),
			Code:                 order.Code,
			Status:               string(order.Status),
			OrderType:            string(order.Type),
			SubtotalCents:        order.SubtotalCents,
			DiscountCents:        order.DiscountCents,
			TotalCents:           order.TotalCents,
			DiscountCodeSnapshot: order.DiscountCodeSnapshot,
			PickupAt:             order.PickupAt,
			TermsAcceptance:      termsAcceptanceToDTO(order.Terms),
			Items:                mapOrderItemsToDTO(items[order.ID]),
			Timeline:             mapOrderTimelineToDTO(timelines[order.ID]),
			Messages:             toMessageResponses(messages[order.ID], false),
			Incidents:            toOrderIncidentResponses(incidents[order.ID]),
			CreatedAt:            order.CreatedAt,
			UpdatedAt:            order.UpdatedAt,
		})
	}
	return out, nil
}

// exportActivity reads every change recorded to the account, newest first, by
// keyset. The network and browser of a change are shown only when the account
// holder made it: a staff member's are that person's data, not theirs.
func (u *DataExportUsecase) exportActivity(ctx context.Context, userID uuid.UUID) ([]dto.AccountActivityResponse, error) {
	out := make([]dto.AccountActivityResponse, 0)
	var before *port.PageCursor
	for {
		entries, err := u.activity.ListForSubject(ctx, userID, before, dataExportPage)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			out = append(out, mapAccountActivity(entry, userID))
		}
		if len(entries) < int(dataExportPage) {
			return out, nil
		}
		last := entries[len(entries)-1]
		before = &port.PageCursor{At: last.At, ID: last.ID}
	}
}

// exportCart reads the lines waiting in the customer's cart; a customer who
// never added anything has no cart yet.
func (u *DataExportUsecase) exportCart(ctx context.Context, userID uuid.UUID) ([]dto.DataExportCartItemResponse, error) {
	cart, err := u.carts.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return []dto.DataExportCartItemResponse{}, nil
		}
		return nil, err
	}
	items, err := u.carts.ListItemsByCartID(ctx, cart.ID)
	if err != nil {
		return nil, err
	}
	return mapExportCartItems(items), nil
}

func mapExportSessions(sessions []domainsession.SessionMeta) []dto.DataExportSessionResponse {
	out := make([]dto.DataExportSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		out = append(out, dto.DataExportSessionResponse{
			SignedInAt: session.CreatedAt,
			IP:         session.IP,
			UserAgent:  session.UserAgent,
		})
	}
	return out
}

func mapAccountActivity(entry port.AccountActivity, subjectID uuid.UUID) dto.AccountActivityResponse {
	activity := dto.AccountActivityResponse{
		Action: string(entry.Action),
		At:     entry.At,
		By:     string(entry.ActorRole),
		Before: entry.BeforeJSON,
		After:  entry.AfterJSON,
	}
	if entry.ActorID == subjectID {
		activity.By = dto.AccountActivityByYou
		if entry.IP != "" {
			ip := entry.IP
			activity.IP = &ip
		}
		activity.UserAgent = entry.UserAgent
	}
	return activity
}

func mapExportCartItems(items []domaincart.Item) []dto.DataExportCartItemResponse {
	out := make([]dto.DataExportCartItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.DataExportCartItemResponse{
			ID:            item.ID.String(),
			LineType:      string(item.LineType),
			ProductID:     uuidString(item.ProductID),
			ComboID:       uuidString(item.ComboID),
			Quantity:      item.Quantity,
			Configuration: item.Configuration,
			AddedAt:       item.CreatedAt,
		})
	}
	return out
}

func mapExportSavedProducts(saved []port.SavedProductEntry) []dto.DataExportSavedProductResponse {
	out := make([]dto.DataExportSavedProductResponse, 0, len(saved))
	for _, item := range saved {
		out = append(out, dto.DataExportSavedProductResponse{
			List:        string(item.List),
			ProductID:   item.ProductID.String(),
			ProductName: item.ProductName,
			SavedAt:     item.SavedAt,
		})
	}
	return out
}

func mapExportReviews(reviews []port.ReviewEntry) []dto.DataExportReviewResponse {
	out := make([]dto.DataExportReviewResponse, 0, len(reviews))
	for _, review := range reviews {
		out = append(out, dto.DataExportReviewResponse{
			OrderID:     review.OrderID.String(),
			OrderCode:   review.OrderCode,
			ProductID:   review.ProductID.String(),
			ProductName: review.ProductName,
			Rating:      review.Rating,
			Comment:     review.Comment,
			Status:      string(review.Status),
			CreatedAt:   review.CreatedAt,
		})
	}
	return out
}

func uuidString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

func termsAcceptanceToDTO(terms *domainpolicy.Acceptance) *dto.TermsAcceptanceResponse {
	if terms == nil {
		return nil
	}
	return &dto.TermsAcceptanceResponse{Version: terms.Version, AcceptedAt: terms.AcceptedAt}
}
