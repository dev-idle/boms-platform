package email

import (
	"fmt"
	"time"

	domainaccount "github.com/boms/backend/internal/domain/account"
	"github.com/boms/backend/internal/port"
)

// Where the storefront takes a link (ROUTE.verifyEmail and ROUTE.resetPassword
// in frontend/src/constants/routes.ts). The token rides in the fragment: a
// browser never sends that to a server, so it stays out of access logs and
// Referer headers, and the page hands it to the API itself.
const (
	verifyEmailPath   = "/verify-email"
	resetPasswordPath = "/reset-password"
)

// accountCopy is what one account email says.
type accountCopy struct {
	path       string
	subject    string
	preheader  string
	eyebrow    string
	heading    string
	lead       []string
	action     string
	validity   string // formatted with how long the link works
	footerNote string
}

var accountCopies = map[domainaccount.Purpose]accountCopy{
	domainaccount.PurposeVerifyEmail: {
		path:      verifyEmailPath,
		subject:   "Confirm your email address for " + choux.Name,
		preheader: "Confirm your address to order for pickup.",
		eyebrow:   "Confirm your email",
		heading:   "Confirm your email address",
		lead: []string{
			"Confirm that this address is yours to place orders and get updates about them.",
		},
		action:     "Confirm email",
		validity:   "The link works for %s.",
		footerNote: "You are receiving this email because an account was created with this address at " + choux.Name + ". If that was not you, ignore it: nothing happens without the link.",
	},
	domainaccount.PurposeResetPassword: {
		path:      resetPasswordPath,
		subject:   "Reset your " + choux.Name + " password",
		preheader: "Choose a new password for your " + choux.Name + " account.",
		eyebrow:   "Password reset",
		heading:   "Choose a new password",
		lead: []string{
			"Someone asked to reset the password of the " + choux.Name + " account for this address. If it was you, choose a new one with the link below; doing so signs you out on every device.",
		},
		action:     "Choose a new password",
		validity:   "The link works for %s, and only once.",
		footerNote: "If you did not ask for this, ignore this email: your password stays as it is.",
	},
}

// AccountComposer writes confirmation and password reset emails.
type AccountComposer struct {
	siteURL   string
	templates emailTemplates
}

// NewAccountComposer parses the templates once; siteURL is the storefront
// origin links point to.
func NewAccountComposer(siteURL string) (*AccountComposer, error) {
	parsed, err := parseEmail("account")
	if err != nil {
		return nil, err
	}
	return &AccountComposer{siteURL: siteURL, templates: parsed}, nil
}

type accountView struct {
	frame
	Validity string
}

// ComposeAccountEmail implements port.AccountEmailComposer.
func (c *AccountComposer) ComposeAccountEmail(msg port.AccountEmail) (port.Email, error) {
	words, ok := accountCopies[msg.Purpose]
	if !ok {
		return port.Email{}, fmt.Errorf("no email for account link %q", msg.Purpose)
	}
	if msg.Token == "" {
		return port.Email{}, fmt.Errorf("an account email needs its link token")
	}
	view := accountView{
		frame: frame{
			Brand:       choux,
			Subject:     words.subject,
			Preheader:   words.preheader,
			Eyebrow:     words.eyebrow,
			Heading:     words.heading,
			Greeting:    "Hello,",
			Lead:        words.lead,
			ActionLabel: words.action,
			ActionURL:   c.siteURL + words.path + "#token=" + msg.Token,
			FooterNote:  words.footerNote,
		},
		Validity: fmt.Sprintf(words.validity, lifetime(msg.Purpose.TTL())),
	}
	return c.templates.render(view, msg.To, "", view.Subject)
}

// lifetime reads a link's TTL as a person would say it: "48 hours", "30 minutes".
func lifetime(d time.Duration) string {
	if d >= time.Hour && d%time.Hour == 0 {
		return count(int(d/time.Hour), "hour")
	}
	return count(int(d/time.Minute), "minute")
}

func count(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", n, unit)
}

var _ port.AccountEmailComposer = (*AccountComposer)(nil)
