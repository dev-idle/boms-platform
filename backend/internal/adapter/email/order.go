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
	domainorder.NoticeRequested: {
		subject:   "We received your request %s",
		preheader: "We will check your custom order and email you once we accept it.",
		eyebrow:   "Request received",
		heading:   "Thank you for your request",
		lead: []string{
			"We have your order and your payment. Before we start, we check that we can make it as you asked, and we will email you once we accept it.",
			"If we cannot make it, we will tell you why and refund you in full.",
		},
		receipt: true,
	},
	domainorder.NoticeAccepted: {
		subject:   "We accepted your order %s",
		preheader: "Your custom order is confirmed for your pickup time.",
		eyebrow:   "Order accepted",
		heading:   "Your order is confirmed",
		lead: []string{
			"We checked your custom order and will have it ready for your pickup time. We will email you again when it is ready to collect.",
		},
	},
	domainorder.NoticeReady: {
		subject:   "Your order %s is ready to collect",
		preheader: "Come in during your pickup time and give your pickup code at the counter.",
		eyebrow:   "Ready for pickup",
		heading:   "Your order is ready",
		lead: []string{
			"Come in during your pickup time and give your pickup code at the counter.",
		},
	},
	// The lead says who cancelled and why, and what comes back (cancelledLead).
	domainorder.NoticeCancelled: {
		subject:   "Your order %s was cancelled",
		preheader: "Your order will not be made.",
		eyebrow:   "Order cancelled",
		heading:   "Your order was cancelled",
	},
	domainorder.NoticeNoShow: {
		subject:   "Your order %s was not collected",
		preheader: "We kept your order until closing time on its pickup day.",
		eyebrow:   "Missed pickup",
		heading:   "Your order was not collected",
		lead: []string{
			"Your order was ready, but it was not collected by closing time on its pickup day, so we recorded it as a missed pickup.",
			"As our refund policy explains, a missed pickup is not refunded: it was made for you and cannot be sold again.",
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
	PickupCode   string
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
		Code:       order.Code,
		PickupCode: msg.PickupCode,
		Receipt:    words.receipt,
		Subtotal:   money(order.SubtotalCents),
		Total:      money(order.TotalCents),
	}
	if msg.CustomerName != "" {
		view.Greeting = "Hi " + msg.CustomerName + ","
	}
	if msg.Notice == domainorder.NoticeCancelled {
		view.Lead = cancelledLead(msg)
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

// cancelledLead says who cancelled the order and why, and what goes back.
func cancelledLead(msg port.OrderEmail) []string {
	lead := []string{"Your order has been cancelled as you asked, and will not be made."}
	if msg.Reason != "" {
		lead = []string{"We are sorry: we had to cancel your order, and it will not be made.", "Our reason: " + msg.Reason}
	}
	if msg.RefundCents > 0 {
		lead = append(lead, "We are returning "+money(msg.RefundCents)+
			" to your PayPal account. When it arrives is up to PayPal and your bank.")
	}
	return append(lead, "If you did not expect this, reply to this email or call us and we will help.")
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
