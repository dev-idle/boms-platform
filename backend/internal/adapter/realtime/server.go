package realtime

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"go.uber.org/zap"

	"github.com/boms/backend/internal/adapter/eventbus"
	"github.com/boms/backend/internal/config"
	domainrealtime "github.com/boms/backend/internal/domain/realtime"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

const (
	// StatusReauthenticate asks the browser for a new ticket: its session ended,
	// was rotated by a token refresh, or the socket reached its maximum lifetime.
	StatusReauthenticate websocket.StatusCode = 4001
	// maxInboundBytes is all a push-only socket accepts from a browser; any data
	// frame closes the socket anyway.
	maxInboundBytes = 512
	// maxFailedSessionChecks ends a socket whose session cannot be confirmed for
	// this many checks in a row: a Redis hiccup is tolerated, an outage is not.
	maxFailedSessionChecks = 3
)

// Server admits browsers with a ticket and runs their sockets.
type Server struct {
	hub       *Hub
	tickets   port.RealtimeTicketStore
	sessions  port.SessionStore
	cfg       config.RealtimeConfig
	log       *zap.Logger
	origins   map[string]struct{}
	admission *admission

	mu       sync.Mutex
	closed   bool
	stopping chan struct{}
	sockets  sync.WaitGroup
}

// NewServer returns a server whose sockets hear what hub delivers.
func NewServer(
	hub *Hub,
	tickets port.RealtimeTicketStore,
	sessions port.SessionStore,
	cfg config.RealtimeConfig,
	log *zap.Logger,
) *Server {
	origins := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		origins[origin] = struct{}{}
	}
	return &Server{
		hub:       hub,
		tickets:   tickets,
		sessions:  sessions,
		cfg:       cfg,
		log:       log,
		origins:   origins,
		admission: newAdmission(cfg.AdmissionRate, cfg.AdmissionRatePerAddress),
		stopping:  make(chan struct{}),
	}
}

// Handler serves GET /ws and nothing else.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", s.serveSocket)
	return mux
}

// Close tells every socket the server is going away and waits for them to
// finish, or for ctx to end. Sockets are hijacked connections, which
// http.Server.Shutdown neither closes nor waits for.
func (s *Server) Close(ctx context.Context) {
	s.mu.Lock()
	if !s.closed {
		s.closed = true
		close(s.stopping)
	}
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		s.sockets.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		s.log.Warn("realtime_sockets_left_open", zap.Error(ctx.Err()))
	}
}

// admit counts a socket in, unless the server is closing.
func (s *Server) admit() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	s.sockets.Add(1)
	return true
}

func (s *Server) serveSocket(w http.ResponseWriter, r *http.Request) {
	if !s.admit() {
		http.Error(w, "server shutting down", http.StatusServiceUnavailable)
		return
	}
	defer s.sockets.Done()

	// The cheap refusals come before any Redis work: only the site's own pages
	// may connect (a browser always sends Origin on a WebSocket handshake), a
	// token must look issued, and handshakes are metered per address and per
	// process.
	if _, ok := s.origins[r.Header.Get("Origin")]; !ok {
		http.Error(w, "origin not allowed", http.StatusForbidden)
		return
	}
	token := r.URL.Query().Get("ticket")
	if !domainrealtime.WellFormedToken(token) {
		http.Error(w, "invalid or used ticket", http.StatusUnauthorized)
		return
	}
	if !s.admission.allow(clientAddress(r, s.cfg.TrustedProxies), time.Now()) {
		w.Header().Set("Retry-After", "1")
		http.Error(w, "too many connection attempts", http.StatusTooManyRequests)
		return
	}

	ticket, err := s.redeem(r.Context(), token)
	if err != nil {
		if errors.Is(err, domainrealtime.ErrTicketInvalid) || errors.Is(err, apperrors.ErrNotFound) {
			http.Error(w, "invalid or used ticket", http.StatusUnauthorized)
			return
		}
		s.log.Error("realtime_ticket_redeem_failed", zap.Error(err))
		http.Error(w, "realtime unavailable", http.StatusServiceUnavailable)
		return
	}

	c := newClient(ticket.UserID, channelsFor(ticket), s.cfg.SendBuffer)
	if !s.hub.add(c) {
		http.Error(w, "too many open connections", http.StatusTooManyRequests)
		return
	}
	defer s.hub.remove(c)

	// The listener runs on its own host or port, so the site's pages are cross
	// origin to it: Accept admits them by the same list checked above.
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: s.cfg.AllowedOrigins})
	if err != nil {
		return // Accept has already answered, e.g. 403 for a foreign origin.
	}
	defer func() { _ = conn.CloseNow() }()
	conn.SetReadLimit(maxInboundBytes)

	s.run(r.Context(), conn, c, ticket)
}

// redeem consumes a ticket and confirms its session still exists: a ticket
// issued just before a revocation must not open a socket after it.
func (s *Server) redeem(ctx context.Context, token string) (domainrealtime.Ticket, error) {
	ticket, err := s.tickets.Consume(ctx, token)
	if err != nil {
		return domainrealtime.Ticket{}, err
	}
	if err := s.checkSession(ctx, ticket); err != nil {
		return domainrealtime.Ticket{}, err
	}
	return ticket, nil
}

// channelsFor picks what a ticket may hear: its own user channel and the public
// channel always, and the role channel of the bakery's roles, each sharing its
// work: the counter's and the kitchen's orders, what the managers watch.
// Nothing private travels on a channel a customer shares with anyone.
func channelsFor(t domainrealtime.Ticket) []string {
	channels := []string{eventbus.UserChannel(t.UserID), eventbus.PublicChannel}
	switch t.Role {
	case domainuser.RoleStaff, domainuser.RoleBaker, domainuser.RoleManager:
		channels = append(channels, eventbus.RoleChannel(t.Role))
	}
	return channels
}

// run writes queued events to the socket until the browser leaves, the session
// ends, the socket reaches its lifetime, it falls behind, or the server stops.
func (s *Server) run(ctx context.Context, conn *websocket.Conn, c *client, t domainrealtime.Ticket) {
	open := rejectInbound(ctx, conn)

	ping := time.NewTicker(s.cfg.PingInterval)
	defer ping.Stop()
	check := time.NewTicker(s.cfg.SessionCheckInterval)
	defer check.Stop()
	lifetime := time.NewTimer(s.cfg.MaxLifetime)
	defer lifetime.Stop()
	failedChecks := 0

	for {
		select {
		case <-open.Done():
			return
		case <-s.stopping:
			_ = conn.Close(websocket.StatusGoingAway, "server shutting down")
			return
		case <-c.slow:
			_ = conn.Close(websocket.StatusTryAgainLater, "fell behind")
			return
		case payload := <-c.send:
			if err := s.write(open, conn, payload); err != nil {
				return
			}
		case <-ping.C:
			if err := s.ping(open, conn); err != nil {
				return
			}
		case <-check.C:
			if code, reason, end := s.recheckSession(open, t, &failedChecks); end {
				_ = conn.Close(code, reason)
				return
			}
		case <-lifetime.C:
			_ = conn.Close(StatusReauthenticate, "socket lifetime reached")
			return
		}
	}
}

// rejectInbound reads the socket so control frames are answered, and closes it
// with StatusPolicyViolation when the browser sends data: the socket only
// pushes. Unlike websocket.Conn.CloseRead it reads the whole message before
// closing, so the close handshake completes at once instead of timing out. The
// returned context ends with the socket.
func rejectInbound(ctx context.Context, conn *websocket.Conn) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		if _, _, err := conn.Read(ctx); err == nil {
			_ = conn.Close(websocket.StatusPolicyViolation, "this socket only sends")
		}
	}()
	return ctx
}

func (s *Server) write(ctx context.Context, conn *websocket.Conn, payload []byte) error {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.WriteTimeout)
	defer cancel()
	return conn.Write(ctx, websocket.MessageText, payload)
}

// ping proves the browser is still there; a socket behind a dead network path
// would otherwise hold its slot until the lifetime ends.
func (s *Server) ping(ctx context.Context, conn *websocket.Conn) error {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.WriteTimeout)
	defer cancel()
	return conn.Ping(ctx)
}

// recheckSession runs one periodic check and says whether, and how, the socket
// must end: at once when the session is gone, and after maxFailedSessionChecks
// checks in a row that Redis could not answer. failed carries that count.
func (s *Server) recheckSession(ctx context.Context, t domainrealtime.Ticket, failed *int) (websocket.StatusCode, string, bool) {
	err := s.checkSession(ctx, t)
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		return StatusReauthenticate, "session ended", true
	case err != nil:
		*failed++
		s.log.Warn("realtime_session_check_failed", zap.Int("in_a_row", *failed), zap.Error(err))
		if *failed >= maxFailedSessionChecks {
			return websocket.StatusTryAgainLater, "session unverifiable", true
		}
	default:
		*failed = 0
	}
	return 0, "", false
}

// checkSession returns apperrors.ErrNotFound once the ticket's session is gone
// (logout, revocation, or the rotation every token refresh makes).
func (s *Server) checkSession(ctx context.Context, t domainrealtime.Ticket) error {
	_, err := s.sessions.Get(ctx, t.UserID.String(), t.SessionID.String())
	return err
}
