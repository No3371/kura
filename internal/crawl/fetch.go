// Package crawl contains the polite HTTP fetcher and the crawler that turns
// source sweeps into stored listings and price observations.
package crawl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/temoto/robotstxt"
)

// DefaultUserAgent identifies the crawler honestly. Sites that refuse
// self-identified bots are not crawled (see README: "Choosing a source").
const DefaultUserAgent = "Mozilla/5.0 (compatible; KuraBot/0.1; +https://github.com/No3371/kura)"

const robotsAgent = "KuraBot"

var ErrDisallowed = errors.New("crawl: disallowed by robots.txt")

// Fetcher is a per-host rate-limited HTTP client that honours robots.txt
// (including Crawl-delay) and retries transient failures with backoff.
type Fetcher struct {
	Client    *http.Client
	UserAgent string
	MinDelay  time.Duration // minimum gap between requests to the same host
	Retries   int
	MaxBody   int64

	mu     sync.Mutex
	hosts  map[string]*hostState
	Counts Counters
}

// Counters are cumulative and safe to read concurrently.
type Counters struct {
	Requests, Bytes, Retries, Errors atomic.Int64
}

type hostState struct {
	mu     sync.Mutex // serialises requests to one host
	next   time.Time
	robots *robotstxt.RobotsData
	delay  time.Duration
}

func NewFetcher(minDelay time.Duration) *Fetcher {
	return &Fetcher{
		Client:    &http.Client{Timeout: 30 * time.Second},
		UserAgent: DefaultUserAgent,
		MinDelay:  minDelay,
		Retries:   3,
		MaxBody:   8 << 20,
		hosts:     map[string]*hostState{},
	}
}

// Response is a fully-read HTTP response.
type Response struct {
	Status int
	URL    *url.URL // final URL after redirects
	Body   []byte
}

// Get fetches rawURL, waiting for the host's rate limit and checking robots.txt.
func (f *Fetcher) Get(ctx context.Context, rawURL string) (*Response, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	h, err := f.host(ctx, u)
	if err != nil {
		return nil, err
	}
	if h.robots != nil && !h.robots.TestAgent(u.RequestURI(), robotsAgent) {
		return nil, fmt.Errorf("%w: %s", ErrDisallowed, u.Path)
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	var lastErr error
	for attempt := 0; attempt <= f.Retries; attempt++ {
		if attempt > 0 {
			f.Counts.Retries.Add(1)
		}
		if err := h.wait(ctx); err != nil {
			return nil, err
		}
		resp, retryAfter, err := f.do(ctx, u)
		h.schedule(retryAfter)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !retryable(err) {
			break
		}
		backoff := time.Duration(1<<attempt) * 2 * time.Second
		h.schedule(backoff)
	}
	f.Counts.Errors.Add(1)
	return nil, lastErr
}

type statusError struct{ code int }

func (e statusError) Error() string { return "http status " + strconv.Itoa(e.code) }

func retryable(err error) bool {
	if se, ok := errors.AsType[statusError](err); ok {
		return se.code == http.StatusTooManyRequests || se.code >= 500
	}
	return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
}

// do performs a single request. 404 is returned as a Response, not an error,
// because "the product page is gone" is a meaningful answer for the crawler.
func (f *Fetcher) do(ctx context.Context, u *url.URL) (*Response, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", f.UserAgent)
	req.Header.Set("Accept-Language", "ja,en;q=0.5")
	f.Counts.Requests.Add(1)
	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, f.MaxBody))
	if err != nil {
		return nil, 0, err
	}
	f.Counts.Bytes.Add(int64(len(body)))
	retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
	if resp.StatusCode >= 400 && resp.StatusCode != http.StatusNotFound {
		return nil, retryAfter, statusError{resp.StatusCode}
	}
	return &Response{Status: resp.StatusCode, URL: resp.Request.URL, Body: body}, retryAfter, nil
}

func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	if s, err := strconv.Atoi(v); err == nil {
		return time.Duration(s) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		return time.Until(t)
	}
	return 0
}

// host returns per-host state, fetching robots.txt on first contact.
func (f *Fetcher) host(ctx context.Context, u *url.URL) (*hostState, error) {
	f.mu.Lock()
	h, ok := f.hosts[u.Host]
	if !ok {
		h = &hostState{delay: f.MinDelay}
		f.hosts[u.Host] = h
	}
	f.mu.Unlock()
	if ok {
		return h, nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	robotsURL := &url.URL{Scheme: u.Scheme, Host: u.Host, Path: "/robots.txt"}
	resp, _, err := f.do(ctx, robotsURL)
	h.schedule(0)
	// If robots.txt is unreachable, proceed at the default delay.
	if err == nil {
		if rd, perr := robotstxt.FromStatusAndBytes(resp.Status, resp.Body); perr == nil {
			h.robots = rd
			if g := rd.FindGroup(robotsAgent); g != nil {
				h.delay = max(h.delay, g.CrawlDelay)
			}
		}
	}
	return h, nil
}

func (h *hostState) wait(ctx context.Context) error {
	d := time.Until(h.next)
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// schedule pushes the next allowed request time out by the host delay plus up
// to 25% jitter, or by extra if the server asked for longer.
func (h *hostState) schedule(extra time.Duration) {
	d := max(extra, h.delay+time.Duration(rand.Int64N(int64(h.delay)/4+1)))
	if n := time.Now().Add(d); n.After(h.next) {
		h.next = n
	}
}
