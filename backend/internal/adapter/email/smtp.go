package email

import (
	"context"
	"errors"
	"fmt"
	netmail "net/mail"
	"strings"

	"github.com/wneessen/go-mail"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/port"
)

// SMTPMailer sends email through an SMTP server, one connection per message:
// the worker sends a few at a time, and a connection held open between
// messages would only be dropped by the server.
//
// Its errors never name the recipient. go-mail and mail servers put the
// address in theirs, and a send error reaches the logs and the email queue's
// records, which must not keep a customer's contact details.
type SMTPMailer struct {
	cfg  config.MailConfig
	opts []mail.Option
}

// NewSMTPMailer returns a mailer for cfg, which config validates.
func NewSMTPMailer(cfg config.MailConfig) *SMTPMailer {
	opts := []mail.Option{mail.WithPort(cfg.SMTPPort), mail.WithTimeout(cfg.SendTimeout)}
	switch cfg.SMTPTLS {
	case config.SMTPTLSImplicit:
		opts = append(opts, mail.WithSSL())
	case config.SMTPTLSNone:
		opts = append(opts, mail.WithTLSPolicy(mail.NoTLS))
	default:
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	}
	if cfg.SMTPUsername != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
			mail.WithUsername(cfg.SMTPUsername),
			mail.WithPassword(cfg.SMTPPassword),
		)
	}
	return &SMTPMailer{cfg: cfg, opts: opts}
}

// Send implements port.Mailer. Only the server refusing the address itself
// wraps port.ErrEmailUndeliverable; a full mailbox, a sending quota, a policy
// refusal, a dropped connection, a timeout or a "try again later" may pass on
// a retry.
//
// go-mail applies SendTimeout once to connecting and once to sending, so one
// call takes up to twice that.
func (m *SMTPMailer) Send(ctx context.Context, e port.Email) error {
	msg, err := m.message(e)
	if err != nil {
		return err
	}
	client, err := mail.NewClient(m.cfg.SMTPHost, m.opts...)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, m.cfg.SendTimeout)
	defer cancel()
	if err := client.DialAndSendWithContext(ctx, msg); err != nil {
		// The server took the message and only closing the connection failed:
		// retrying would mail the customer twice.
		if msg.IsDelivered() {
			return nil
		}
		return sendFailure(err)
	}
	return nil
}

// message builds the email from structured addresses, so a display name is
// quoted and encoded as a header needs and can never change a recipient.
func (m *SMTPMailer) message(e port.Email) (*mail.Msg, error) {
	to, err := netmail.ParseAddress(e.To)
	if err != nil || to.Address != e.To {
		return nil, fmt.Errorf("%w: recipient address is not valid", port.ErrEmailUndeliverable)
	}
	to.Name = e.ToName
	msg := mail.NewMsg()
	msg.FromMailAddress(&netmail.Address{Name: m.cfg.FromName, Address: m.cfg.FromAddress})
	msg.AddToMailAddress(to)
	msg.ReplyToMailAddress(&netmail.Address{Address: m.cfg.ReplyTo})
	msg.Subject(e.Subject)
	msg.SetDate()
	msg.SetMessageID()
	if e.ListUnsubscribe != "" {
		msg.SetGenHeaderPreformatted(mail.HeaderListUnsubscribe, "<"+e.ListUnsubscribe+">")
	}
	msg.SetBodyString(mail.TypeTextPlain, e.Text)
	msg.AddAlternativeString(mail.TypeTextHTML, e.HTML)
	return msg, nil
}

// sendFailure describes a failed send by what the server answered, never by
// its text, which names the recipient. Connection and authentication failures
// carry no recipient and pass through.
func sendFailure(err error) error {
	var refused *mail.SendError
	if !errors.As(err, &refused) {
		return fmt.Errorf("send mail: %w", err)
	}
	code, enhanced := refused.ErrorCode(), refused.EnhancedStatusCode()
	described := fmt.Errorf("smtp %s failed: %d %s", refused.Reason, code, enhanced)
	if addressRefused(code, enhanced) {
		return fmt.Errorf("%w: %w", port.ErrEmailUndeliverable, described)
	}
	return described
}

// addressRefused reports whether the server refused the address itself — a
// mailbox or domain that does not exist — which no retry changes: an enhanced
// status of class 5.1, or, from a server that gives none, a bare 550, 551 or
// 553. Any other refusal, a 5xx included, may pass later: a full mailbox
// (5.2.2), a sending quota (5.4.5, 4.7.x) or a policy (5.7.x). go-mail reports
// a dropped connection as a failed step with no reply code.
func addressRefused(code int, enhanced string) bool {
	if enhanced != "" {
		return strings.HasPrefix(enhanced, "5.1.")
	}
	return code == 550 || code == 551 || code == 553
}

var _ port.Mailer = (*SMTPMailer)(nil)
