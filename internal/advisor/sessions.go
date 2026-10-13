package advisor

import (
	"sync"
	"time"
)

// conversation is one chat session's history in a provider's message type.
type conversation[M any] struct {
	mu       sync.Mutex // one in-flight turn per session
	messages []M
	used     time.Time
}

// sessionStore keeps up to maxSessions conversations in memory, evicting the
// least recently used; the zero value is ready to use.
type sessionStore[M any] struct {
	mu sync.Mutex
	m  map[string]*conversation[M]
}

func (st *sessionStore[M]) get(id string) *conversation[M] {
	st.mu.Lock()
	defer st.mu.Unlock()
	if st.m == nil {
		st.m = map[string]*conversation[M]{}
	}
	s, ok := st.m[id]
	if !ok {
		if len(st.m) >= maxSessions {
			st.evictOldest()
		}
		s = &conversation[M]{}
		st.m[id] = s
	}
	s.used = time.Now()
	return s
}

func (st *sessionStore[M]) evictOldest() {
	var oldest string
	var at time.Time
	for id, s := range st.m {
		if oldest == "" || s.used.Before(at) {
			oldest, at = id, s.used
		}
	}
	delete(st.m, oldest)
}
