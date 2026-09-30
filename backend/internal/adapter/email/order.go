package email

import (
	"fmt"
	"strconv"
	"strings"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/port"
)

// pickupLayout reads as "Tuesday, September 30 at 11:00 AM", in the bakery's
// own time zone whatever the reader's.
const pickupLayout = "Monday, January 2 at 3:04 PM"

var orderFooterNote = "You are receiving this email about an order you placed with " + choux.Name + " for pickup."

// orderCopy is what one notice says.
type orderCopy struct {
	subject   string // formatted with the order code
	preheader string
	eyebrow   string
	heading   string
	lead      []string
	receipt   bool
}

var orderCopies = map[domainorder.Notice]orderCopy{
	domainorder.NoticePlaced: {
		subject:   "We received your order %s",
		preheader: "Your order is in. We will email you again when it is ready to collect.",
		eyebrow:   "Order received",
		heading:   "Thank you for your order",
		lead: []string{
			"We have your order and will have it ready for your pickup time. We will email you again when it is ready to collect.",
		},
		receipt: true,
	},
	domainorder.NoticeReady: {
		subject:   "Your order %s is ready to collect",
		preheader: "Come in during your pickup time and give your order code at the counter.",
		eyebrow:   "Ready for pickup",
		heading:   "Your order is ready",
		lead: []string{
			"Come in during your pickup time and give your order code at the counter.",
		},
	},
	domainorder.NoticeCancelled: {
		subject:   "Your order %s was cancelled",
		preheader: "Your order will not be made.",
		eyebrow:   "Order cancelled",
		heading:   "Your order was cancelled",
		lead: []string{
			"We are sorry: your order has been cancelled and will not be made.",
			"If you did not expect this, reply to this email or call us and we will help.",
		},
	},
	domainorder.NoticeExpired: {
		subject:   "Your order %s was not paid in time",
		preheader: "We held your pickup time, but the payment did not come through.",
		eyebrow:   "Order expired",
		heading:   "Your order has expired",
		lead: []string{
			"We held your pickup time for you, but the payment did not come through in time, so the order will not be made. Nothing was charged.",
			"You are welcome to order again whenever you like.",
		},
	},
}

// OrderComposer writes order notices in the bakery's style, as HTML with a
// plain-text alternative.
type OrderComposer struct {
	siteURL   string
	templates emailTemplates
}

// NewOrderComposer parses the templates once; siteURL is the storefront origin
// links point to.
func NewOrderComposer(siteURL string) (*OrderComposer, error) {
	parsed, err := parseEmail("order")
	if err != nil {
		return nil, err
	}
	return &OrderComposer{siteURL: siteURL, templates: parsed}, nil
}

type orderLine struct {
	Name     string
	Quantity int32
	Total    string
}

type orderView struct {
	frame
	Code         string
	Pickup       string
	Receipt      bool
	Lines        []orderLine
	Subtotal     string
	Discount     string
	DiscountCode string
	Total        string
}

// ComposeOrderEmail implements port.OrderEmailComposer.
func (c *OrderComposer) ComposeOrderEmail(msg port.OrderEmail) (port.Email, error) {
	words, ok := orderCopies[msg.Notice]
	if !ok {
		return port.Email{}, fmt.Errorf("no email for order notice %q", msg.Notice)
	}
	view := c.view(words, msg)
	return c.templates.render(view, msg.To, msg.CustomerName, view.Subject)
}

func (c *OrderComposer) view(words orderCopy, msg port.OrderEmail) orderView {
	order := msg.Order
	view := orderView{
		frame: frame{
			Brand:       choux,
			Subject:     fmt.Sprintf(words.subject, order.Code),
			Preheader:   words.preheader,
			Eyebrow:     words.eyebrow,
			Heading:     words.heading,
			Greeting:    "Hello,",
			Lead:        words.lead,
			ActionLabel: "View your order",
			ActionURL:   c.siteURL + domainorder.PagePath(order.ID),
			FooterNote:  orderFooterNote,
		},
		Code:     order.Code,
		Receipt:  words.receipt,
		Subtotal: money(order.SubtotalCents),
		Total:    money(order.TotalCents),
	}
	if msg.CustomerName != "" {
		view.Greeting = "Hi " + msg.CustomerName + ","
	}
	if order.PickupAt != nil {
		view.Pickup = order.PickupAt.In(domainstore.Location).Format(pickupLayout)
	}
	if order.DiscountCents > 0 {
		view.Discount = "−" + money(order.DiscountCents)
		if order.DiscountCodeSnapshot != nil {
			view.DiscountCode = *order.DiscountCodeSnapshot
		}
	}
	view.Lines = make([]orderLine, 0, len(msg.Items))
	for _, item := range msg.Items {
		view.Lines = append(view.Lines, orderLine{
			Name:     item.Name,
			Quantity: item.Quantity,
			Total:    money(item.LineTotalCents),
		})
	}
	return view
}

// money writes cents as US dollars, the currency the bakery prices in, the way
// the storefront does: $1,234.50.
func money(cents int64) string {
	dollars := strconv.FormatInt(cents/100, 10)
	var grouped strings.Builder
	for i, digit := range dollars {
		if i > 0 && (len(dollars)-i)%3 == 0 {
			grouped.WriteByte(',')
		}
		grouped.WriteRune(digit)
	}
	return fmt.Sprintf("$%s.%02d", grouped.String(), cents%100)
}

var _ port.OrderEmailComposer = (*OrderComposer)(nil)
