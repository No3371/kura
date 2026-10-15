package web

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ChatLimiter is a token bucket per client IP for POST /api/chat. Content
// classification cannot stop a flood of harmless-looking turns, and every
// turn spends advisor tokens, so turns are rate-limited before anything
// else looks at them. A ~40-line map stands in for golang.org/x/time/rate,
// which would add a dependency for one call site.
type ChatLimiter struct {
	rate  float64 // tokens per second
	burst float64

	mu      sync.Mutex
	clients map[string]*bucket
	swept   time.Time
	now     func() time.Time // for tests; nil = time.Now
}

type bucket struct {
	tokens float64
	at     time.Time
}

// sweepAt is the client count above which idle buckets are dropped (at most
// once a minute), so a scan from many addresses cannot grow the map forever.
const sweepAt = 4096

// NewChatLimiter allows perMinute turns per client, with bursts up to burst.
func NewChatLimiter(perMinute, burst int) *ChatLimiter {
	return &ChatLimiter{rate: float64(perMinute) / 60, burst: float64(max(burst, 1)), clients: map[string]*bucket{}}
}

// Allow takes one token from client's bucket, reporting whether it had one.
func (l *ChatLimiter) Allow(client string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	if l.now != nil {
		now = l.now()
	}
	b, ok := l.clients[client]
	if !ok {
		if len(l.clients) >= sweepAt && now.Sub(l.swept) > time.Minute {
			l.sweep(now)
		}
		b = &bucket{tokens: l.burst, at: now}
		l.clients[client] = b
	}
	b.tokens = min(l.burst, b.tokens+now.Sub(b.at).Seconds()*l.rate)
	b.at = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// sweep drops buckets that have refilled: a full bucket and a missing one
// behave the same.
func (l *ChatLimiter) sweep(now time.Time) {
	for k, b := range l.clients {
		if b.tokens+now.Sub(b.at).Seconds()*l.rate >= l.burst {
			delete(l.clients, k)
		}
	}
	l.swept = now
}

// clientIP is the address the limiter keys on. Behind a proxy (trustProxy)
// it is the last X-Forwarded-For hop, the one the proxy itself appended;
// earlier hops are whatever the client sent. Without a proxy the header is
// ignored, since any client could set it to dodge its own bucket.
func clientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if vs := r.Header.Values("X-Forwarded-For"); len(vs) > 0 {
			hops := strings.Split(vs[len(vs)-1], ",")
			if ip := strings.TrimSpace(hops[len(hops)-1]); ip != "" {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
