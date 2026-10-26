package web

import (
	"net/http"
	"testing"
	"time"
)

func TestBanlistBansAfterStrikes(t *testing.T) {
	b := NewBanlist(3, 24*time.Hour)
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	b.now = func() time.Time { return now }

	for i := range 2 {
		if b.Strike("a") {
			t.Fatalf("strike %d banned early", i+1)
		}
	}
	if _, banned := b.Banned("a"); banned {
		t.Fatal("banned before the third strike")
	}
	if !b.Strike("a") {
		t.Fatal("third strike did not ban")
	}
	left, banned := b.Banned("a")
	if !banned || left != 24*time.Hour {
		t.Fatalf("Banned = %v, %v; want 24h, true", left, banned)
	}
	if _, banned := b.Banned("b"); banned {
		t.Error("another address banned")
	}
	if b.Strike("a") {
		t.Error("a strike during a ban started another ban")
	}

	now = now.Add(23 * time.Hour)
	if _, banned := b.Banned("a"); !banned {
		t.Error("ban lifted early")
	}
	now = now.Add(time.Hour + time.Second)
	if _, banned := b.Banned("a"); banned {
		t.Error("ban outlived its duration")
	}
	// A lifted ban leaves a clean record: it takes the full count again.
	b.Strike("a")
	if b.Strike("a") {
		t.Error("two strikes after a ban banned again")
	}
}

func TestBanlistStrikesExpire(t *testing.T) {
	b := NewBanlist(3, 24*time.Hour)
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	b.now = func() time.Time { return now }

	b.Strike("a")
	b.Strike("a")
	now = now.Add(banWindow + time.Minute)
	if b.Strike("a") {
		t.Error("strikes older than the window still counted")
	}
	// A strike just inside the window does count.
	now = now.Add(banWindow - time.Minute)
	b.Strike("a")
	if !b.Strike("a") {
		t.Error("three strikes within the window did not ban")
	}
}

func TestBanlistSweep(t *testing.T) {
	b := NewBanlist(2, time.Hour)
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	b.now = func() time.Time { return now }
	b.Strike("idle")
	b.Strike("banned")
	b.Strike("banned")

	now = now.Add(banWindow / 2)
	b.sweep(now)
	if len(b.ips) != 2 {
		t.Fatalf("swept inside the window: %d left, want 2", len(b.ips))
	}
	now = now.Add(banWindow)
	b.sweep(now)
	if len(b.ips) != 0 {
		t.Errorf("after the window and the ban: %d left, want 0", len(b.ips))
	}
}

func TestChatBanned(t *testing.T) {
	adv := &countingAdvisor{}
	s := testServer(t, adv)
	s.Bans = NewBanlist(1, 12*time.Hour)
	h := s.Handler()

	s.Bans.Strike("203.0.113.7")
	rec := postChat(h, "203.0.113.7:5000", "", "愛馬仕")
	if rec.Code != http.StatusForbidden || errorText(t, rec) != "此連線因多次違規使用，已暫停使用顧問，約 12 小時後可再試" {
		t.Fatalf("banned client: %d %s", rec.Code, rec.Body)
	}
	if rec.Header().Get("Retry-After") != "43201" {
		t.Errorf("Retry-After = %q, want 43201", rec.Header().Get("Retry-After"))
	}
	if adv.calls.Load() != 0 {
		t.Errorf("advisor calls = %d, want 0 (a banned turn never reaches it)", adv.calls.Load())
	}
	if rec := postChat(h, "198.51.100.2:5000", "", "愛馬仕"); rec.Code != 200 {
		t.Errorf("another client: %d", rec.Code)
	}
}
