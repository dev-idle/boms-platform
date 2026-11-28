package email

import (
	"errors"
	"fmt"

	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
)

var replyFooterNote = "You are receiving this email because our staff at " + choux.Name +
	" wrote to you. We email you once while their messages wait unread, and not again until you have read them."

// ReplyComposer writes the email telling a customer the counter replied.
type ReplyComposer struct {
	siteURL   string
	templates emailTemplates
}

// NewReplyComposer parses the templates once; siteURL is the storefront
// origin links point to.
func NewReplyComposer(siteURL string) (*ReplyComposer, error) {
	parsed, err := parseEmail("reply")
	if err != nil {
		return nil, err
	}
	return &ReplyComposer{siteURL: siteURL, templates: parsed}, nil
}

// ComposeReplyEmail implements port.ReplyEmailComposer. Messages about an
// order — which staff may have written first — link to its page, where its
// conversation is; replies in the general conversation, which the customer
// always starts, to the shop, where it opens from the assistant in the corner
// of a page.
func (c *ReplyComposer) ComposeReplyEmail(msg port.ReplyEmail) (port.Email, error) {
	if msg.To == "" || msg.Count < 1 {
		return port.Email{}, errors.New("a reply email needs its address and a reply")
	}
	view := frame{
		Brand:       choux,
		Subject:     "Our staff replied to your message",
		Preheader:   "Our staff wrote back to you.",
		Eyebrow:     "From our staff",
		Heading:     "Our staff replied",
		Greeting:    "Hello,",
		ActionLabel: "Read and reply",
		ActionURL:   c.siteURL + "/",
		FooterNote:  replyFooterNote,
		Lead: []string{
			"Our staff wrote you " + counted(msg.Count, "a reply", "replies") + " that you have not read yet.",
			"Sign in and open Ask " + choux.Name + ", in the corner of any page, to read and write back.",
		},
	}
	if msg.OrderID != nil {
		view.Subject = "A message from our staff about your order " + msg.OrderCode
		view.Preheader = "Our staff wrote to you about your order."
		view.Heading = "Our staff wrote to you"
		view.Lead = []string{
			"About your order " + msg.OrderCode + ", our staff wrote you " + counted(msg.Count, "a message", "messages") +
				" that you have not read yet.",
		}
		view.ActionURL = c.siteURL + domainorder.PagePath(*msg.OrderID)
	}
	if msg.ToName != "" {
		view.Greeting = "Hello " + msg.ToName + ","
	}
	return c.templates.render(view, msg.To, msg.ToName, view.Subject)
}

// counted reads n of a thing: one as given, more with their number.
func counted(n int32, one, many string) string {
	if n == 1 {
		return one
	}
	return fmt.Sprintf("%d %s", n, many)
}

var _ port.ReplyEmailComposer = (*ReplyComposer)(nil)
