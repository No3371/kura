// Package fx converts JPY prices to TWD. Rates come from open.er-api.com (free,
// no key, daily), are persisted so every price shown can cite its rate date,
// and fall back to a configured constant when nothing has been fetched yet.
package fx

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/No3371/kura/internal/store"
)

const endpoint = "https://open.er-api.com/v6/latest/JPY"

type Converter struct {
	Store    *store.Store
	Client   *http.Client
	Fallback float64 // TWD per JPY used before the first successful fetch

	mu  sync.RWMutex
	cur store.Rate
}

// Load restores the latest stored rate.
func (c *Converter) Load(ctx context.Context) error {
	r, err := c.Store.LatestRate(ctx, "JPY", "TWD")
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	c.set(r)
	return nil
}

// Refresh fetches today's rate and stores it.
func (c *Converter) Refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var body struct {
		Result    string             `json:"result"`
		UpdatedAt int64              `json:"time_last_update_unix"`
		Rates     map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}
	twd := body.Rates["TWD"]
	if body.Result != "success" || twd <= 0 {
		return fmt.Errorf("fx: unexpected response (result=%q, TWD=%v)", body.Result, twd)
	}
	r := store.Rate{Base: "JPY", Quote: "TWD", Rate: twd, AsOf: time.Unix(body.UpdatedAt, 0), Provider: "open.er-api.com"}
	if err := c.Store.SaveRate(ctx, r); err != nil {
		return err
	}
	c.set(r)
	return nil
}

func (c *Converter) set(r store.Rate) {
	c.mu.Lock()
	c.cur = r
	c.mu.Unlock()
}

// Current returns the rate in use (the fallback if none was ever fetched).
func (c *Converter) Current() store.Rate {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cur.Rate > 0 {
		return c.cur
	}
	return store.Rate{Base: "JPY", Quote: "TWD", Rate: c.Fallback, Provider: "fallback"}
}

// TWD converts a JPY amount, rounded to the nearest dollar.
func (c *Converter) TWD(jpy int64) int64 {
	return int64(math.Round(float64(jpy) * c.Current().Rate))
}

// JPY converts a TWD amount back (used for budget filters).
func (c *Converter) JPY(twd int64) int64 {
	r := c.Current().Rate
	if r <= 0 {
		return 0
	}
	return int64(math.Round(float64(twd) / r))
}
