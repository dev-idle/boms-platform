package email

import (
	"bufio"
	"context"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/port"
)

// fakeSMTP is an SMTP server that speaks just enough of the protocol for one
// client at a time. rcptReply is its answer to RCPT TO; dropAt names the step
// ("RCPT", "DATA" once the message is read, "QUIT") at which it hangs up
// without answering.
type fakeSMTP struct {
	addr      string
	rcptReply string
	dropAt    string

	mu       sync.Mutex
	messages []string
	rcpts    []string
}

func startFakeSMTP(t *testing.T, rcptReply, dropAt string) *fakeSMTP {
	t.Helper()
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	server := &fakeSMTP{addr: listener.Addr().String(), rcptReply: rcptReply, dropAt: dropAt}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.serve(conn)
		}
	}()
	return server
}

func (s *fakeSMTP) serve(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	r := bufio.NewReader(conn)
	reply := func(line string) { _, _ = conn.Write([]byte(line + "\r\n")) }
	reply("220 fake.example ESMTP")
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		cmd := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(cmd, "EHLO"):
			reply("250-fake.example")
			reply("250 8BITMIME")
		case strings.HasPrefix(cmd, "RCPT TO"):
			if s.dropAt == "RCPT" {
				return
			}
			s.mu.Lock()
			s.rcpts = append(s.rcpts, strings.TrimSpace(line))
			s.mu.Unlock()
			reply(s.rcptReply)
		case cmd == "DATA":
			reply("354 go ahead")
			var body strings.Builder
			for {
				data, err := r.ReadString('\n')
				if err != nil {
					return
				}
				if data == ".\r\n" {
					break
				}
				body.WriteString(data)
			}
			if s.dropAt == "DATA" {
				return
			}
			s.mu.Lock()
			s.messages = append(s.messages, body.String())
			s.mu.Unlock()
			reply("250 queued")
		case cmd == "QUIT":
			if s.dropAt == "QUIT" {
				return
			}
			reply("221 bye")
			return
		default:
			reply("250 ok")
		}
	}
}

func (s *fakeSMTP) delivered() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.messages...)
}

func (s *fakeSMTP) mailer(t *testing.T, tls string) *SMTPMailer {
	t.Helper()
	host, portText, err := net.SplitHostPort(s.addr)
	require.NoError(t, err)
	port, err := strconv.Atoi(portText)
	require.NoError(t, err)
	return NewSMTPMailer(config.MailConfig{
		SMTPHost: host, SMTPPort: port, SMTPTLS: tls,
		FromAddress: "orders@chouxbakery.example", FromName: "Choux",
		ReplyTo: "hello@chouxbakery.example", SendTimeout: 5 * time.Second,
	})
}

var sample = port.Email{
	To: "mai@example.com", ToName: "Mai", Subject: "Your order CH-260930-007 is ready to collect",
	Text: "Your order is ready.", HTML: "<p>Your order is ready.</p>",
}

func TestSMTPMailer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("sends_one_message_with_both_parts", func(t *testing.T) {
		t.Parallel()
		server := startFakeSMTP(t, "250 ok", "")

		require.NoError(t, server.mailer(t, config.SMTPTLSNone).Send(ctx, sample))

		messages := server.delivered()
		require.Len(t, messages, 1)
		msg := messages[0]
		assert.Contains(t, server.rcpts[0], "<mai@example.com>")
		assert.Contains(t, msg, `From: "Choux" <orders@chouxbakery.example>`)
		assert.Contains(t, msg, `To: "Mai" <mai@example.com>`)
		assert.Contains(t, msg, "Reply-To: <hello@chouxbakery.example>")
		assert.Contains(t, msg, "Subject: Your order CH-260930-007 is ready to collect")
		assert.Contains(t, msg, "multipart/alternative")
		assert.Contains(t, msg, "text/plain")
		assert.Contains(t, msg, "text/html")
		assert.Contains(t, msg, "Message-ID:")
	})

	t.Run("names_its_unsubscribe_page_only_when_it_has_one", func(t *testing.T) {
		t.Parallel()
		server := startFakeSMTP(t, "250 ok", "")
		promotion := sample
		promotion.ListUnsubscribe = "https://shop.example/unsubscribe#token=tok"

		mailer := server.mailer(t, config.SMTPTLSNone)
		require.NoError(t, mailer.Send(ctx, promotion))
		require.NoError(t, mailer.Send(ctx, sample))

		messages := server.delivered()
		require.Len(t, messages, 2)
		assert.Contains(t, messages[0], "List-Unsubscribe: <https://shop.example/unsubscribe#token=tok>")
		assert.NotContains(t, messages[1], "List-Unsubscribe", "an email the recipient asked for has none")
	})

	t.Run("a_display_name_is_quoted_not_parsed", func(t *testing.T) {
		t.Parallel()
		server := startFakeSMTP(t, "250 ok", "")
		named := sample
		named.ToName = `Anna "Annie" Lee\`

		require.NoError(t, server.mailer(t, config.SMTPTLSNone).Send(ctx, named), "an unusual name still gets the email")

		messages := server.delivered()
		require.Len(t, messages, 1)
		assert.Contains(t, server.rcpts[0], "<mai@example.com>", "the name never changes the recipient")
		assert.Contains(t, messages[0], `To: "Anna \"Annie\" Lee\\" <mai@example.com>`)
	})

	t.Run("a_final_refusal_is_undeliverable", func(t *testing.T) {
		t.Parallel()
		server := startFakeSMTP(t, "550 5.1.1 no such user", "")

		err := server.mailer(t, config.SMTPTLSNone).Send(ctx, sample)

		require.ErrorIs(t, err, port.ErrEmailUndeliverable)
		assert.Contains(t, err.Error(), "550")
		assert.NotContains(t, err.Error(), "mai@example.com", "errors reach logs and the queue: no customer address")
	})

	t.Run("try_again_later_is_retried", func(t *testing.T) {
		t.Parallel()
		server := startFakeSMTP(t, "451 4.3.0 try again later", "")

		err := server.mailer(t, config.SMTPTLSNone).Send(ctx, sample)

		require.Error(t, err)
		assert.NotErrorIs(t, err, port.ErrEmailUndeliverable, "the queue retries it")
		assert.NotContains(t, err.Error(), "mai@example.com")
	})

	t.Run("a_dropped_connection_is_retried", func(t *testing.T) {
		t.Parallel()
		for _, step := range []string{"RCPT", "DATA"} {
			server := startFakeSMTP(t, "250 ok", step)

			err := server.mailer(t, config.SMTPTLSNone).Send(ctx, sample)

			require.Error(t, err, step)
			assert.NotErrorIs(t, err, port.ErrEmailUndeliverable, "a connection lost at %s is not a refusal", step)
			assert.NotContains(t, err.Error(), "mai@example.com")
		}
	})

	t.Run("a_message_the_server_took_is_sent_even_if_hanging_up_fails", func(t *testing.T) {
		t.Parallel()
		server := startFakeSMTP(t, "250 ok", "QUIT")

		require.NoError(t, server.mailer(t, config.SMTPTLSNone).Send(ctx, sample), "retrying would mail the customer twice")
		assert.Len(t, server.delivered(), 1)
	})

	t.Run("an_unreachable_server_is_retried", func(t *testing.T) {
		t.Parallel()
		listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
		require.NoError(t, err)
		server := &fakeSMTP{addr: listener.Addr().String()}
		require.NoError(t, listener.Close())

		err = server.mailer(t, config.SMTPTLSNone).Send(ctx, sample)

		require.Error(t, err)
		assert.NotErrorIs(t, err, port.ErrEmailUndeliverable)
	})

	t.Run("starttls_refuses_a_server_that_cannot_encrypt", func(t *testing.T) {
		t.Parallel()
		server := startFakeSMTP(t, "250 ok", "")

		err := server.mailer(t, config.SMTPTLSStartTLS).Send(ctx, sample)

		require.Error(t, err, "the message never crosses the network in the clear")
		assert.Empty(t, server.delivered())
	})

	t.Run("a_malformed_address_is_undeliverable_before_connecting", func(t *testing.T) {
		t.Parallel()
		server := startFakeSMTP(t, "250 ok", "")
		bad := sample
		bad.To = "not an address"

		err := server.mailer(t, config.SMTPTLSNone).Send(ctx, bad)

		require.ErrorIs(t, err, port.ErrEmailUndeliverable)
		assert.NotContains(t, err.Error(), "not an address")
		assert.Empty(t, server.delivered())
	})
}
