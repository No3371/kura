package advisor

import (
	"slices"
	"testing"
)

// Each turn here is "u<n>" plus one or two replies; commit must keep the last
// maxTurns turns whole, starting at a user message.
func TestCommitKeepsLastTurns(t *testing.T) {
	var s conversation[string]
	for n := range maxTurns + 3 {
		start := len(s.messages)
		s.messages = append(s.messages, "u"+string(rune('a'+n)), "a")
		if n%2 == 0 {
			s.messages = append(s.messages, "tool")
		}
		s.commit(start)
	}
	if len(s.turns) != maxTurns {
		t.Fatalf("turns = %d, want %d", len(s.turns), maxTurns)
	}
	if s.turns[0] != 0 || s.messages[0] != "u"+string(rune('a'+3)) {
		t.Fatalf("history starts at %v (turns %v), want the 4th turn's user message", s.messages[0], s.turns)
	}
	for _, i := range s.turns {
		if s.messages[i][0] != 'u' {
			t.Fatalf("turn index %d points at %q", i, s.messages[i])
		}
	}
	if !slices.IsSorted(s.turns) {
		t.Fatalf("turns out of order: %v", s.turns)
	}
}
