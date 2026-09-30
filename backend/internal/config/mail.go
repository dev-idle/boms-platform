package config

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode"
)

// SMTP connection security, as MAIL_SMTP_TLS names it.
const (
	// SMTPTLSStartTLS upgrades a plain connection before anything is sent — the
	// submission port, 587. A server that does not offer it is refused.
	SMTPTLSStartTLS = "starttls"
	// SMTPTLSImplicit speaks TLS from the first byte — port 465.
	SMTPTLSImplicit = "tls"
	// SMTPTLSNone sends in the clear: only for a local catcher such as Mailpit.
	SMTPTLSNone = "none"
)

const maxMailConcurrency = 64

// maxMailSendTimeout keeps a send — which takes up to twice the timeout —
// inside the email task's one-minute deadline, with room to read the order.
const maxMailSendTimeout = 10 * time.Second

// MailConfig is the outgoing mail server the worker sends through and the
// sender customers see. Only cmd/worker sends mail.
type MailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPTLS      string
	FromAddress  string
	FromName     string
	// ReplyTo is where a customer's reply goes: the bakery's own inbox, not the
	// address the mail is sent from.
	ReplyTo     string
	SendTimeout time.Duration
	// Concurrency is how many emails the worker sends at once.
	Concurrency int
}

func (c MailConfig) validate(env string) error {
	env = strings.ToLower(strings.TrimSpace(env))
	deployed := env == "production" || env == "staging"
	if strings.TrimSpace(c.SMTPHost) == "" {
		return errors.New("mail.smtp_host is required")
	}
	if c.SMTPPort < 1 || c.SMTPPort > 65535 {
		return errors.New("mail.smtp_port must be between 1 and 65535")
	}
	switch c.SMTPTLS {
	case SMTPTLSStartTLS, SMTPTLSImplicit, SMTPTLSNone:
	default:
		return fmt.Errorf("mail.smtp_tls must be %q, %q or %q", SMTPTLSStartTLS, SMTPTLSImplicit, SMTPTLSNone)
	}
	if (c.SMTPUsername == "") != (c.SMTPPassword == "") {
		return errors.New("mail.smtp_username and mail.smtp_password are set together or not at all")
	}
	if err := validateMailbox("mail.from_address", c.FromAddress); err != nil {
		return err
	}
	if err := validateMailbox("mail.reply_to", c.ReplyTo); err != nil {
		return err
	}
	if c.FromName == "" || strings.ContainsFunc(c.FromName, unicode.IsControl) {
		return errors.New("mail.from_name is required and must be one line of text")
	}
	if c.SendTimeout <= 0 || c.SendTimeout > maxMailSendTimeout {
		return fmt.Errorf("mail.send_timeout must be positive and at most %s", maxMailSendTimeout)
	}
	if c.Concurrency < 1 || c.Concurrency > maxMailConcurrency {
		return fmt.Errorf("mail.concurrency must be between 1 and %d", maxMailConcurrency)
	}
	if deployed {
		if c.SMTPTLS == SMTPTLSNone {
			return errors.New("mail.smtp_tls must not be none in staging/production: credentials and customer data would cross the network in the clear")
		}
		if c.SMTPUsername == "" {
			return errors.New("mail.smtp_username and mail.smtp_password are required in staging/production")
		}
		for _, mailbox := range []struct{ name, value string }{
			{"mail.from_address", c.FromAddress},
			{"mail.reply_to", c.ReplyTo},
		} {
			if isReservedMailDomain(mailbox.value) {
				return fmt.Errorf("%s must be an address at the bakery's own domain in staging/production, not a reserved example domain", mailbox.name)
			}
		}
	}
	return nil
}

// isReservedMailDomain reports whether address is at a domain RFC 2606 keeps
// for examples and tests. The development defaults use one; in production it
// means a setting was left unset, and mail would go out from a domain nobody
// owns.
func isReservedMailDomain(address string) bool {
	domain := strings.ToLower(address[strings.LastIndex(address, "@")+1:])
	for _, reserved := range []string{"example", "invalid", "test", "localhost", "example.com", "example.net", "example.org"} {
		if domain == reserved || strings.HasSuffix(domain, "."+reserved) {
			return true
		}
	}
	return false
}

// validateMailbox accepts a bare address, as a header needs it: no display
// name, which has its own setting.
func validateMailbox(name, value string) error {
	addr, err := mail.ParseAddress(value)
	if err != nil || addr.Name != "" || addr.Address != value {
		return fmt.Errorf("%s must be a plain email address", name)
	}
	return nil
}
