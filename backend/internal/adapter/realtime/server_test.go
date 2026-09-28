package realtime

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/boms/backend/internal/adapter/eventbus"
	"github.com/boms/backend/internal/config"
	domainrealtime "github.com/boms/backend/internal/domain/realtime"
	domainsession "github.com/boms/backend/internal/domain/session"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

const testOrigin = "http://localhost:3000"

type fakeTickets struct {
	mu      sync.Mutex
	tickets map[string]domainrealtime.Ticket
	err     error
}

func (f *fakeTickets) Issue(_ context.Context, t domainrealtime.Ticket, _ time.Duration) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	token := wellFormedToken()
	f.tickets[token] = t
	return token, nil
}

// wellFormedToken has the shape of an issued ticket; whether it was issued is
// up to the store.
func wellFormedToken() string {
	raw := make([]byte, domainrealtime.TicketBytes)
	_, _ = rand.Read(raw)
	return base64.RawURLEncoding.EncodeToString(raw)
}

func (f *fakeTickets) Consume(_ context.Context, token string) (domainrealtime.Ticket, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return domainrealtime.Ticket{}, f.err
	}
	t, ok := f.tickets[token]
	if !ok {
		return domainrealtime.Ticket{}, domainrealtime.ErrTicketInvalid
	}
	delete(f.tickets, token)
	return t, nil
}

func (f *fakeTickets) fail(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.err = err
}

// fakeSessions answers Get only; a socket never needs the rest of the store.
type fakeSessions struct {
	port.SessionStore
	mu    sync.Mutex
	alive map[uuid.UUID]bool
	// failing answers the next this many lookups with an error; -1 fails them all.
	failing int
	lookups int
}

func (f *fakeSessions) Get(_ context.Context, _, sessionID string) (domainsession.SessionMeta, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lookups++
	if f.failing != 0 {
		if f.failing > 0 {
			f.failing--
		}
		return domainsession.SessionMeta{}, errors.New("redis down")
	}
	if !f.alive[uuid.MustParse(sessionID)] {
		return domainsession.SessionMeta{}, apperrors.ErrNotFound
	}
	return domainsession.SessionMeta{}, nil
}

func (f *fakeSessions) setAlive(sessionID uuid.UUID, alive bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.alive[sessionID] = alive
}

func (f *fakeSessions) fail(lookups int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failing = lookups
}

func (f *fakeSessions) lookupCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lookups
}

type harness struct {
	hub      *Hub
	tickets  *fakeTickets
	sessions *fakeSessions
	server   *Server
	url      string
}

func newHarness(t *testing.T, tune func(*config.RealtimeConfig)) *harness {
	t.Helper()
	cfg := config.RealtimeConfig{
		AllowedOrigins:          []string{testOrigin},
		SessionCheckInterval:    time.Hour,
		MaxLifetime:             time.Hour,
		PingInterval:            time.Hour,
		WriteTimeout:            time.Second,
		SendBuffer:              4,
		MaxConnsPerUser:         2,
		MaxConns:                100,
		AdmissionRate:           1000,
		AdmissionRatePerAddress: 1000,
	}
	if tune != nil {
		tune(&cfg)
	}
	h := &harness{
		hub:      NewHub(cfg.MaxConnsPerUser),
		tickets:  &fakeTickets{tickets: map[string]domainrealtime.Ticket{}},
		sessions: &fakeSessions{alive: map[uuid.UUID]bool{}},
	}
	h.server = NewServer(h.hub, h.tickets, h.sessions, cfg, zap.NewNop())
	ts := httptest.NewServer(h.server.Handler())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.server.Close(ctx)
		ts.Close()
	})
	h.url = "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	return h
}

// ticket issues a ticket for a new user of role whose session is alive.
func (h *harness) ticket(t *testing.T, role domainuser.Role) (string, domainrealtime.Ticket) {
	t.Helper()
	ticket := domainrealtime.Ticket{UserID: uuid.New(), Role: role, SessionID: uuid.New()}
	h.sessions.setAlive(ticket.SessionID, true)
	token, err := h.tickets.Issue(context.Background(), ticket, time.Minute)
	require.NoError(t, err)
	return token, ticket
}

// dial opens a socket, or returns the status the handshake was refused with.
func (h *harness) dial(t *testing.T, token, origin string) (*websocket.Conn, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	header := http.Header{}
	if origin != "" {
		header.Set("Origin", origin)
	}
	conn, resp, err := websocket.Dial(ctx, h.url+"?ticket="+url.QueryEscape(token), &websocket.DialOptions{HTTPHeader: header})
	status := 0
	if resp != nil {
		status = resp.StatusCode
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}
	if err != nil {
		return nil, status
	}
	t.Cleanup(func() { _ = conn.CloseNow() })
	return conn, status
}

func read(t *testing.T, conn *websocket.Conn) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, payload, err := conn.Read(ctx)
	return string(payload), err
}

func TestServerAdmission(t *testing.T) {
	t.Parallel()

	t.Run("opens_a_socket_for_a_valid_ticket", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, nil)
		token, _ := h.ticket(t, domainuser.RoleCustomer)

		_, status := h.dial(t, token, testOrigin)

		assert.Equal(t, http.StatusSwitchingProtocols, status)
	})

	t.Run("refuses_a_malformed_ticket", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, nil)

		conn, status := h.dial(t, "not-a-ticket", testOrigin)

		assert.Nil(t, conn)
		assert.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("refuses_an_unknown_ticket", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, nil)

		conn, status := h.dial(t, wellFormedToken(), testOrigin)

		assert.Nil(t, conn)
		assert.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("refuses_a_ticket_used_twice", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, nil)
		token, _ := h.ticket(t, domainuser.RoleCustomer)
		_, first := h.dial(t, token, testOrigin)
		require.Equal(t, http.StatusSwitchingProtocols, first)

		_, second := h.dial(t, token, testOrigin)

		assert.Equal(t, http.StatusUnauthorized, second)
	})

	t.Run("refuses_a_foreign_origin_without_spending_the_ticket", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, nil)
		token, _ := h.ticket(t, domainuser.RoleCustomer)

		conn, status := h.dial(t, token, "https://evil.example")

		assert.Nil(t, conn)
		assert.Equal(t, http.StatusForbidden, status)
		_, status = h.dial(t, token, testOrigin)
		assert.Equal(t, http.StatusSwitchingProtocols, status)
	})

	t.Run("refuses_a_handshake_without_origin", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, nil)
		token, _ := h.ticket(t, domainuser.RoleCustomer)

		_, status := h.dial(t, token, "")

		assert.Equal(t, http.StatusForbidden, status)
	})

	t.Run("refuses_a_ticket_whose_session_has_ended", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, nil)
		token, ticket := h.ticket(t, domainuser.RoleStaff)
		h.sessions.setAlive(ticket.SessionID, false)

		_, status := h.dial(t, token, testOrigin)

		assert.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("answers_503_when_the_session_cannot_be_checked", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, nil)
		token, _ := h.ticket(t, domainuser.RoleStaff)
		h.sessions.fail(-1)

		_, status := h.dial(t, token, testOrigin)

		assert.Equal(t, http.StatusServiceUnavailable, status)
	})

	t.Run("throttles_redeems_before_touching_redis", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, func(cfg *config.RealtimeConfig) {
			cfg.AdmissionRate = 1
			cfg.AdmissionRatePerAddress = 1
		})
		for range 2 {
			token, _ := h.ticket(t, domainuser.RoleCustomer)
			_, status := h.dial(t, token, testOrigin)
			require.Equal(t, http.StatusSwitchingProtocols, status)
		}
		lookups := h.sessions.lookupCount()
		token, _ := h.ticket(t, domainuser.RoleCustomer)

		_, status := h.dial(t, token, testOrigin)

		assert.Equal(t, http.StatusTooManyRequests, status)
		assert.Equal(t, lookups, h.sessions.lookupCount())
	})

	t.Run("answers_503_when_tickets_cannot_be_checked", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, nil)
		token, _ := h.ticket(t, domainuser.RoleCustomer)
		h.tickets.fail(errors.New("redis down"))

		_, status := h.dial(t, token, testOrigin)

		assert.Equal(t, http.StatusServiceUnavailable, status)
	})

	t.Run("caps_sockets_per_user", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, func(cfg *config.RealtimeConfig) { cfg.MaxConnsPerUser = 1 })
		token, ticket := h.ticket(t, domainuser.RoleStaff)
		_, first := h.dial(t, token, testOrigin)
		require.Equal(t, http.StatusSwitchingProtocols, first)
		again, err := h.tickets.Issue(context.Background(), ticket, time.Minute)
		require.NoError(t, err)

		_, second := h.dial(t, again, testOrigin)

		assert.Equal(t, http.StatusTooManyRequests, second)
	})
}

func TestServerDelivery(t *testing.T) {
	t.Parallel()

	t.Run("customer_hears_only_its_own_channel", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, nil)
		token, ticket := h.ticket(t, domainuser.RoleCustomer)
		conn, _ := h.dial(t, token, testOrigin)

		h.hub.Deliver(eventbus.RoleChannel(domainuser.RoleCustomer), []byte("broadcast"))
		h.hub.Deliver(eventbus.RoleChannel(domainuser.RoleStaff), []byte("staff"))
		h.hub.Deliver(eventbus.UserChannel(uuid.New()), []byte("someone else"))
		h.hub.Deliver(eventbus.UserChannel(ticket.UserID), []byte("mine"))

		payload, err := read(t, conn)
		require.NoError(t, err)
		assert.Equal(t, "mine", payload)
	})

	t.Run("staff_hears_its_role_channel", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, nil)
		token, _ := h.ticket(t, domainuser.RoleStaff)
		conn, _ := h.dial(t, token, testOrigin)

		h.hub.Deliver(eventbus.RoleChannel(domainuser.RoleBaker), []byte("baker"))
		h.hub.Deliver(eventbus.RoleChannel(domainuser.RoleStaff), []byte("staff"))

		payload, err := read(t, conn)
		require.NoError(t, err)
		assert.Equal(t, "staff", payload)
	})

	t.Run("drops_a_socket_that_fell_behind", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, nil)
		token, ticket := h.ticket(t, domainuser.RoleCustomer)
		conn, _ := h.dial(t, token, testOrigin)

		h.hub.mu.RLock()
		for c := range h.hub.byChannel[eventbus.UserChannel(ticket.UserID)] {
			c.markSlow()
		}
		h.hub.mu.RUnlock()

		_, err := read(t, conn)
		assert.Equal(t, websocket.StatusTryAgainLater, websocket.CloseStatus(err))
	})

	t.Run("closes_a_socket_that_sends_data", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, nil)
		token, _ := h.ticket(t, domainuser.RoleCustomer)
		conn, _ := h.dial(t, token, testOrigin)

		require.NoError(t, conn.Write(context.Background(), websocket.MessageText, []byte("hello")))

		_, err := read(t, conn)
		assert.Equal(t, websocket.StatusPolicyViolation, websocket.CloseStatus(err))
	})
}

func TestServerLifecycle(t *testing.T) {
	t.Parallel()

	t.Run("asks_for_a_new_ticket_when_the_session_ends", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, func(cfg *config.RealtimeConfig) { cfg.SessionCheckInterval = 20 * time.Millisecond })
		token, ticket := h.ticket(t, domainuser.RoleBaker)
		conn, _ := h.dial(t, token, testOrigin)

		h.sessions.setAlive(ticket.SessionID, false)

		_, err := read(t, conn)
		assert.Equal(t, StatusReauthenticate, websocket.CloseStatus(err))
	})

	t.Run("rides_out_a_few_unanswered_checks", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, func(cfg *config.RealtimeConfig) { cfg.SessionCheckInterval = 10 * time.Millisecond })
		token, ticket := h.ticket(t, domainuser.RoleBaker)
		conn, _ := h.dial(t, token, testOrigin)
		before := h.sessions.lookupCount()

		h.sessions.fail(maxFailedSessionChecks - 1)
		require.Eventually(t, func() bool { return h.sessions.lookupCount() >= before+maxFailedSessionChecks+2 },
			5*time.Second, 5*time.Millisecond, "failures then answered checks")
		h.hub.Deliver(eventbus.UserChannel(ticket.UserID), []byte("still here"))

		payload, err := read(t, conn)
		require.NoError(t, err)
		assert.Equal(t, "still here", payload)
	})

	t.Run("gives_up_when_the_session_stays_unverifiable", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, func(cfg *config.RealtimeConfig) { cfg.SessionCheckInterval = 10 * time.Millisecond })
		token, _ := h.ticket(t, domainuser.RoleBaker)
		conn, _ := h.dial(t, token, testOrigin)

		h.sessions.fail(-1)

		_, err := read(t, conn)
		assert.Equal(t, websocket.StatusTryAgainLater, websocket.CloseStatus(err))
	})

	t.Run("asks_for_a_new_ticket_at_the_maximum_lifetime", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, func(cfg *config.RealtimeConfig) { cfg.MaxLifetime = 20 * time.Millisecond })
		token, _ := h.ticket(t, domainuser.RoleCustomer)
		conn, _ := h.dial(t, token, testOrigin)

		_, err := read(t, conn)
		assert.Equal(t, StatusReauthenticate, websocket.CloseStatus(err))
	})

	t.Run("says_going_away_on_shutdown_and_refuses_new_sockets", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, nil)
		token, _ := h.ticket(t, domainuser.RoleCustomer)
		conn, _ := h.dial(t, token, testOrigin)
		late, _ := h.ticket(t, domainuser.RoleCustomer)

		closed := make(chan struct{})
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			h.server.Close(ctx)
			close(closed)
		}()

		_, err := read(t, conn)
		assert.Equal(t, websocket.StatusGoingAway, websocket.CloseStatus(err))
		<-closed
		_, status := h.dial(t, late, testOrigin)
		assert.Equal(t, http.StatusServiceUnavailable, status)
	})
}
