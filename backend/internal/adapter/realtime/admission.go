package realtime

import (
	"net/http"
	"net/netip"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// addressIdle is how long an address's bucket outlives its last handshake. By
// then it has refilled, so forgetting it loses nothing.
const addressIdle = time.Minute

// admission meters handshakes before any Redis work. Every client address has
// a small bucket and the process one shared bucket: a single host cannot hold
// the listener's budget, and a crowd cannot starve the Redis pool the API
// authenticates with.
type admission struct {
	mu        sync.Mutex
	global    *rate.Limiter
	perAddr   map[netip.Addr]*addressBucket
	addrRate  rate.Limit
	addrBurst int
	lastSweep time.Time
}

type addressBucket struct {
	limiter *rate.Limiter
	seen    time.Time
}

// newAdmission allows globalRate handshakes a second in all, perAddressRate of
// them from one address. Each bucket absorbs a burst of twice its rate, so a
// visitor reopening a window of tabs is not refused.
func newAdmission(globalRate, perAddressRate int) *admission {
	return &admission{
		global:    rate.NewLimiter(rate.Limit(globalRate), 2*globalRate),
		perAddr:   map[netip.Addr]*addressBucket{},
		addrRate:  rate.Limit(perAddressRate),
		addrBurst: 2 * perAddressRate,
	}
}

func (a *admission) allow(addr netip.Addr, now time.Time) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.forgetIdle(now)
	bucket, ok := a.perAddr[addr]
	if !ok {
		bucket = &addressBucket{limiter: rate.NewLimiter(a.addrRate, a.addrBurst)}
		a.perAddr[addr] = bucket
	}
	bucket.seen = now
	return bucket.limiter.AllowN(now, 1) && a.global.AllowN(now, 1)
}

func (a *admission) forgetIdle(now time.Time) {
	if now.Sub(a.lastSweep) < addressIdle {
		return
	}
	a.lastSweep = now
	for addr, bucket := range a.perAddr {
		if now.Sub(bucket.seen) >= addressIdle {
			delete(a.perAddr, addr)
		}
	}
}

// clientAddress is where a handshake came from: the peer itself, or, when the
// peer is a trusted proxy, the nearest forwarded address that is not one of
// ours. X-Forwarded-For is read right to left, so a client cannot pick its own
// bucket by sending the header; a malformed hop leaves the request with the
// proxy's address.
func clientAddress(r *http.Request, trusted []netip.Prefix) netip.Addr {
	peer, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}
	}
	addr := peer.Addr().Unmap()
	if !isTrustedProxy(addr, trusted) {
		return addr
	}
	hops := strings.Split(strings.Join(r.Header.Values("X-Forwarded-For"), ","), ",")
	for _, raw := range slices.Backward(hops) {
		hop, err := netip.ParseAddr(strings.TrimSpace(raw))
		if err != nil {
			return addr
		}
		hop = hop.Unmap()
		if !isTrustedProxy(hop, trusted) {
			return hop
		}
	}
	return addr
}

func isTrustedProxy(addr netip.Addr, trusted []netip.Prefix) bool {
	return slices.ContainsFunc(trusted, func(p netip.Prefix) bool { return p.Contains(addr) })
}
