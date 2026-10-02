package email

import (
	"strings"

	"github.com/boms/backend/internal/port"
)

// Where the storefront shows the catalog (ROUTE.products) and stops promotion
// emails (ROUTE.unsubscribe). The unsubscribe token rides in the fragment, as
// account links do.
const (
	shopPath        = "/products"
	unsubscribePath = "/unsubscribe"
)

// PromotionComposer writes promotion emails.
type PromotionComposer struct {
	siteURL   string
	templates emailTemplates
}

// NewPromotionComposer parses the templates once; siteURL is the storefront
// origin links point to.
func NewPromotionComposer(siteURL string) (*PromotionComposer, error) {
	parsed, err := parseEmail("promotion")
	if err != nil {
		return nil, err
	}
	return &PromotionComposer{siteURL: siteURL, templates: parsed}, nil
}

type promotionView struct {
	frame
	UnsubscribeURL string
}

// ComposePromotionEmail implements port.PromotionEmailComposer: the manager's
// subject as the heading, each line of their message a paragraph, and a link
// that stops these emails, which mail clients also offer as their own button.
func (c *PromotionComposer) ComposePromotionEmail(msg port.PromotionEmail) (port.Email, error) {
	lines := paragraphs(msg.Body)
	unsubscribeURL := c.siteURL + unsubscribePath + "#token=" + msg.UnsubscribeToken
	view := promotionView{
		frame: frame{
			Brand:       choux,
			Subject:     msg.Subject,
			Preheader:   lines[0],
			Eyebrow:     "News from " + choux.Name,
			Heading:     msg.Subject,
			Greeting:    "Hello,",
			Lead:        lines,
			ActionLabel: "Visit the shop",
			ActionURL:   c.siteURL + shopPath,
			FooterNote:  "You are receiving this email because you chose to hear about promotions from " + choux.Name + ".",
		},
		UnsubscribeURL: unsubscribeURL,
	}
	email, err := c.templates.render(view, msg.To, msg.ToName, msg.Subject)
	if err != nil {
		return port.Email{}, err
	}
	email.ListUnsubscribe = unsubscribeURL
	return email, nil
}

// paragraphs is each line of a message that holds anything; a message holds at
// least one.
func paragraphs(body string) []string {
	var out []string
	for line := range strings.SplitSeq(body, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, line)
		}
	}
	return out
}

var _ port.PromotionEmailComposer = (*PromotionComposer)(nil)
