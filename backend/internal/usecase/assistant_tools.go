package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	domainassistant "github.com/boms/backend/internal/domain/assistant"
	domainconversation "github.com/boms/backend/internal/domain/conversation"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
)

const (
	// assistantSearchLimit is how many products one search returns; with no
	// match, the menu's first assistantMenuLimit products stand in.
	assistantSearchLimit int32 = 5
	assistantMenuLimit   int32 = 10
	// assistantQueryLength is the longest search the catalogue takes.
	assistantQueryLength = 100
	// assistantDayLayout and assistantTimeLayout write days and times the way
	// the model repeats them to a visitor, in the bakery's time zone.
	assistantDayLayout  = "Monday 2 January 2006"
	assistantTimeLayout = "Monday 2 January 2006, 15:04"
)

// assistantStatusMeaning is what each status tells a customer, so the model
// repeats the bakery's words rather than its own guess.
var assistantStatusMeaning = map[domainorder.Status]string{
	domainorder.StatusAwaitingPayment: "waiting for the customer to pay online",
	domainorder.StatusPending:         "paid; the bakery has yet to accept the custom request",
	domainorder.StatusConfirmed:       "accepted; the bakery has not started making it",
	domainorder.StatusInProduction:    "being made",
	domainorder.StatusReady:           "ready to collect at the shop",
	domainorder.StatusFulfilled:       "collected",
	domainorder.StatusCancelled:       "cancelled",
	domainorder.StatusExpired:         "cancelled because it was not paid in time",
	domainorder.StatusNoShow:          "not collected on its pickup day",
}

// assistantTools are the tools offered to a visitor: the catalogue and the
// opening hours to anyone, their orders and the counter to a signed-in customer.
func assistantTools(signedIn bool) []port.LLMTool {
	tools := []port.LLMTool{
		{
			Name:        domainassistant.ToolSearchProducts,
			Description: "Search the products the bakery sells now, with prices and whether they sold out today.",
			Parameters: objectSchema(map[string]any{
				"query": map[string]any{"type": "string", "description": "What the customer is looking for, such as 'matcha croissant' or 'birthday cake'."},
			}, "query"),
		},
		{
			Name:        domainassistant.ToolStoreHours,
			Description: "Opening hours, closed days, pickup slot length and how far ahead orders must be placed.",
			Parameters:  objectSchema(map[string]any{}),
		},
	}
	if !signedIn {
		return tools
	}
	return append(tools,
		port.LLMTool{
			Name:        domainassistant.ToolOrderStatus,
			Description: "Status and pickup time of one of the signed-in customer's own orders.",
			Parameters: objectSchema(map[string]any{
				"order_code": map[string]any{"type": "string", "description": "The order code, such as CH-261124-003."},
			}, "order_code"),
		},
		port.LLMTool{
			Name:        domainassistant.ToolHandOff,
			Description: "Pass the customer's question to the bakery's staff, who reply in the customer's messages.",
			Parameters: objectSchema(map[string]any{
				"order_code": map[string]any{"type": "string", "description": "The code of the order the question is about, if it is about one."},
			}),
		},
	)
}

func objectSchema(properties map[string]any, required ...string) map[string]any {
	return map[string]any{"type": "object", "properties": properties, "required": append([]string{}, required...)}
}

// runTool runs one tool the model called and returns its result as JSON for
// the model. A tool the visitor may not use, or arguments it cannot read,
// come back to the model as an error to work around. The customer is passed
// on once: a second hand-off is told it is done.
func (q *AssistantQuestion) runTool(ctx context.Context, call port.LLMToolCall) (string, error) {
	signedIn := q.userID != nil
	args, ok := readToolArgs(call.Arguments)
	if !ok {
		return toolResult(map[string]string{"error": "arguments are not valid JSON"}), nil
	}
	switch {
	case call.Name == domainassistant.ToolSearchProducts:
		return q.searchProducts(ctx, args.Query)
	case call.Name == domainassistant.ToolStoreHours:
		return q.storeHours(ctx)
	case call.Name == domainassistant.ToolOrderStatus && signedIn:
		return q.orderStatus(ctx, args.OrderCode)
	case call.Name == domainassistant.ToolHandOff && signedIn:
		if q.handedOff == nil {
			handOff, err := q.handOff(ctx, args.OrderCode)
			if err != nil {
				return "", err
			}
			q.handedOff = handOff
		}
		return toolResult(map[string]bool{"passed_to_staff": true}), nil
	}
	return toolResult(map[string]string{"error": "no such tool"}), nil
}

// toolArgs are the arguments any of the tools takes.
type toolArgs struct {
	Query     string `json:"query"`
	OrderCode string `json:"order_code"`
}

// readToolArgs reads a call's arguments, none when it sent none; false when
// they are not a JSON object.
func readToolArgs(arguments string) (toolArgs, bool) {
	var args toolArgs
	if arguments == "" {
		return args, true
	}
	return args, json.Unmarshal([]byte(arguments), &args) == nil
}

func (q *AssistantQuestion) searchProducts(ctx context.Context, query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" || utf8.RuneCountInString(query) > assistantQueryLength {
		return toolResult(map[string]string{"error": "query must be 1 to 100 characters"}), nil
	}
	found, err := q.u.products.CatalogList(ctx, port.CatalogListProductsParams{Search: &query, Limit: assistantSearchLimit})
	if err != nil {
		return "", err
	}
	noMatch := len(found) == 0
	if noMatch {
		if found, err = q.u.products.CatalogList(ctx, port.CatalogListProductsParams{Limit: assistantMenuLimit}); err != nil {
			return "", err
		}
	}
	type product struct {
		Name         string `json:"name"`
		Category     string `json:"category"`
		Price        string `json:"price"`
		SoldOutToday bool   `json:"sold_out_today"`
		Customizable bool   `json:"customizable"`
	}
	products := make([]product, 0, len(found))
	for _, p := range found {
		products = append(products, product{
			Name:         p.Name,
			Category:     p.CategoryName,
			Price:        utils.Dollars(p.PriceCents),
			SoldOutToday: soldOutToday(p.SoldOutOn),
			Customizable: p.IsCustomizable,
		})
	}
	return toolResult(map[string]any{"no_exact_match": noMatch, "products": products}), nil
}

func (q *AssistantQuestion) storeHours(ctx context.Context) (string, error) {
	now := time.Now()
	settings, closed, err := readPickupRules(ctx, q.u.store, now)
	if err != nil {
		return "", err
	}
	closedDays := make([]map[string]string, 0, len(closed))
	for _, day := range closed {
		closedDays = append(closedDays, map[string]string{"date": day.Day.Format(assistantDayLayout), "reason": day.Reason})
	}
	return toolResult(map[string]any{
		"today":                        domainstore.DayOf(now).Format(assistantDayLayout),
		"open_every_day_from":          clockOf(settings.OpensAt),
		"open_every_day_until":         clockOf(settings.ClosesAt),
		"closed_dates":                 closedDays,
		"pickup_slot_minutes":          int(settings.SlotLength / time.Minute),
		"made_to_order_notice_minutes": int(settings.PreorderMinLead / time.Minute),
		"ready_now_items_wait_minutes": int(settings.InstantPrep / time.Minute),
		"orders_up_to_days_ahead":      settings.MaxAdvanceDays,
	}), nil
}

func (q *AssistantQuestion) orderStatus(ctx context.Context, code string) (string, error) {
	order, err := q.customerOrderByCode(ctx, code)
	if err != nil || order == nil {
		return toolResult(map[string]bool{"found": false}), err
	}
	result := map[string]any{
		"found":   true,
		"code":    order.Code,
		"status":  assistantStatusMeaning[order.Status],
		"type":    string(order.Type),
		"channel": string(order.Channel),
	}
	if order.PickupAt != nil {
		result["pickup_at"] = order.PickupAt.In(domainstore.Location).Format(assistantTimeLayout)
	}
	return toolResult(result), nil
}

// handOff writes the customer's questions to the counter: on the order the
// model named, when it is theirs and the bakery has taken it, or else in
// their general conversation.
func (q *AssistantQuestion) handOff(ctx context.Context, code string) (*dto.AssistantHandOffResponse, error) {
	req := dto.PostMessageRequest{Body: q.handOffBody()}
	order, err := q.customerOrderByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if order != nil {
		_, err := q.u.messages.Post(ctx, *q.userID, &order.ID, req)
		if err == nil {
			orderID := order.ID.String()
			return &dto.AssistantHandOffResponse{OrderID: &orderID, OrderCode: &order.Code}, nil
		}
		if !errors.Is(err, domainconversation.ErrUnavailable) {
			return nil, err
		}
	}
	if _, err := q.u.messages.Post(ctx, *q.userID, nil, req); err != nil {
		return nil, err
	}
	return &dto.AssistantHandOffResponse{}, nil
}

// customerOrderByCode reads one of the signed-in customer's orders by the
// code the model passed; nil when the code names none of theirs.
func (q *AssistantQuestion) customerOrderByCode(ctx context.Context, code string) (*domainorder.Order, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil, nil
	}
	order, err := q.u.orders.GetByCodeForUser(ctx, *q.userID, code)
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, nil
	}
	return order, err
}

// clockOf writes a time of day, held as the time since midnight, as 07:30.
func clockOf(d time.Duration) string {
	return fmt.Sprintf("%02d:%02d", int(d/time.Hour), int(d%time.Hour/time.Minute))
}

// toolResult writes a result for the model. Results are maps and structs of
// strings, numbers and bools, which always encode, so there is no error to handle.
func toolResult(v any) string {
	out, _ := json.Marshal(v)
	return string(out)
}
