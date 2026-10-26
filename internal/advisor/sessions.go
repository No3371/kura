package advisor

import (
	"sync"
	"time"
)

// maxTurns is how many committed turns a session replays. Every turn resends
// the whole history, so without a cap one session's cost per turn keeps
// growing; older turns are dropped whole, never splitting a tool call from
// its result.
const maxTurns = 10

// conversation is one chat session's history in a provider's message type.
type conversation[M any] struct {
	mu       sync.Mutex // one in-flight turn per session
	messages []M
	turns    []int // index in messages where each committed turn starts
	used     time.Time
}

// commit records a finished turn that began at messages[start] and drops the
// oldest turns beyond maxTurns. A turn starts with the shopper's message, so
// the trimmed history still opens with a user message.
func (s *conversation[M]) commit(start int) {
	s.turns = append(s.turns, start)
	if len(s.turns) <= maxTurns {
		return
	}
	cut := s.turns[len(s.turns)-maxTurns]
	// Copy rather than reslice, so the dropped turns are freed.
	s.messages = append([]M(nil), s.messages[cut:]...)
	s.turns = s.turns[len(s.turns)-maxTurns:]
	for i := range s.turns {
		s.turns[i] -= cut
	}
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
