package email

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"strconv"
	"strings"
	texttemplate "text/template"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/port"
)

//go:embed templates/order.html templates/order.txt
var orderTemplates embed.FS

// orderPath is where the storefront shows one of a customer's orders
// (ROUTE.orderDetail in frontend/src/constants/routes.ts).
const orderPath = "/orders/"

// pickupLayout reads as "Tuesday, September 30 at 11:00 AM", in the bakery's
// own time zone whatever the reader's.
const pickupLayout = "Monday, January 2 at 3:04 PM"

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
}

// OrderComposer writes order notices in the bakery's style, as HTML with a
// plain-text alternative.
type OrderComposer struct {
	siteURL string
	html    *htmltemplate.Template
	text    *texttemplate.Template
}

// NewOrderComposer parses the templates once; siteURL is the storefront origin
// links point to.
func NewOrderComposer(siteURL string) (*OrderComposer, error) {
	html, err := htmltemplate.ParseFS(orderTemplates, "templates/order.html")
	if err != nil {
		return nil, fmt.Errorf("parse order email html: %w", err)
	}
	text, err := texttemplate.ParseFS(orderTemplates, "templates/order.txt")
	if err != nil {
		return nil, fmt.Errorf("parse order email text: %w", err)
	}
	return &OrderComposer{siteURL: siteURL, html: html, text: text}, nil
}

type orderLine struct {
	Name     string
	Quantity int32
	Total    string
}

type orderView struct {
	Brand        brand
	Subject      string
	Preheader    string
	Eyebrow      string
	Heading      string
	Greeting     string
	Lead         []string
	Code         string
	Pickup       string
	Receipt      bool
	Lines        []orderLine
	Subtotal     string
	Discount     string
	DiscountCode string
	Total        string
	OrderURL     string
}

// ComposeOrderEmail implements port.OrderEmailComposer.
func (c *OrderComposer) ComposeOrderEmail(msg port.OrderEmail) (port.Email, error) {
	words, ok := orderCopies[msg.Notice]
	if !ok {
		return port.Email{}, fmt.Errorf("no email for order notice %q", msg.Notice)
	}
	view := c.view(words, msg)
	var html, text bytes.Buffer
	if err := c.html.Execute(&html, view); err != nil {
		return port.Email{}, fmt.Errorf("render order email html: %w", err)
	}
	if err := c.text.Execute(&text, view); err != nil {
		return port.Email{}, fmt.Errorf("render order email text: %w", err)
	}
	return port.Email{
		To:      msg.To,
		ToName:  msg.CustomerName,
		Subject: view.Subject,
		Text:    text.String(),
		HTML:    html.String(),
	}, nil
}

func (c *OrderComposer) view(words orderCopy, msg port.OrderEmail) orderView {
	order := msg.Order
	view := orderView{
		Brand:     choux,
		Subject:   fmt.Sprintf(words.subject, order.Code),
		Preheader: words.preheader,
		Eyebrow:   words.eyebrow,
		Heading:   words.heading,
		Greeting:  "Hello,",
		Lead:      words.lead,
		Code:      order.Code,
		Receipt:   words.receipt,
		Subtotal:  money(order.SubtotalCents),
		Total:     money(order.TotalCents),
		OrderURL:  c.siteURL + orderPath + order.ID.String(),
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
