package web

import (
	"sync"
	"time"
)

// banWindow is how long a strike counts toward a ban: after an hour without
// a new strike an address starts over.
const banWindow = time.Hour

// Banlist bans a client IP for a long time once the chat gate has blocked
// `after` of its turns within banWindow. The gate's own strikes follow the
// session, which a visitor resets by opening a new one; this follows the
// address. It lives in memory, so a restart lifts every ban.
type Banlist struct {
	after int
	ban   time.Duration

	mu    sync.Mutex
	ips   map[string]*offender
	swept time.Time
	now   func() time.Time // for tests; nil = time.Now
}

type offender struct {
	strikes int
	last    time.Time // latest strike
	until   time.Time // zero unless banned
}

// NewBanlist bans an address for ban after `after` blocked turns.
func NewBanlist(after int, ban time.Duration) *Banlist {
	return &Banlist{after: after, ban: ban, ips: map[string]*offender{}}
}

func (b *Banlist) clock() time.Time {
	if b.now != nil {
		return b.now()
	}
	return time.Now()
}

// Banned reports whether ip is banned and for how much longer.
func (b *Banlist) Banned(ip string) (time.Duration, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	o, ok := b.ips[ip]
	if !ok {
		return 0, false
	}
	if left := o.until.Sub(b.clock()); left > 0 {
		return left, true
	}
	return 0, false
}

// Strike records one blocked turn from ip, reporting whether it started a ban.
func (b *Banlist) Strike(ip string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.clock()
	o, ok := b.ips[ip]
	if !ok {
		if len(b.ips) >= sweepAt && now.Sub(b.swept) > time.Minute {
			b.sweep(now)
		}
		o = &offender{}
		b.ips[ip] = o
	}
	if now.Before(o.until) {
		return false // already banned; a turn in flight when the ban began
	}
	if now.Sub(o.last) > banWindow {
		o.strikes = 0
	}
	o.strikes++
	o.last = now
	if o.strikes < b.after {
		return false
	}
	o.strikes, o.until = 0, now.Add(b.ban)
	return true
}

// sweep drops addresses that are neither banned nor inside the strike window.
func (b *Banlist) sweep(now time.Time) {
	for k, o := range b.ips {
		if !now.Before(o.until) && now.Sub(o.last) > banWindow {
			delete(b.ips, k)
		}
	}
	b.swept = now
}
