package crawl

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/No3371/kura/internal/listing"
	"github.com/No3371/kura/internal/store"
)

// Source is a resale site adapter.
type Source interface {
	Name() string
	// Targets lists the crawl targets (brands) this source is configured for.
	Targets() []string
	// Sweep pages through target's catalogue, newest first, calling visit with
	// each page's bag listings. complete is true when the last page was reached
	// (as opposed to stopping at maxPages, 0 = unlimited).
	Sweep(ctx context.Context, target string, maxPages int, visit func(page int, items []listing.Listing) error) (complete bool, err error)
	// Detail fetches a product page. gone is true if the page no longer exists.
	Detail(ctx context.Context, url string) (d listing.Detail, gone bool, err error)
}

type Crawler struct {
	Store  *store.Store
	Source Source
	Log    *slog.Logger
	// MaxPages caps pages per target per sweep (0 = all). Capped sweeps never
	// mark anything gone because they can't tell "sold" from "not reached".
	MaxPages int
	// DetailBudget caps product-page fetches per run; list pages already carry
	// everything the radar needs, details only enrich (colour, size, accessories).
	DetailBudget int
	Now          func() time.Time
}

func (c *Crawler) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// RunAll sweeps every configured target, then spends the detail budget.
func (c *Crawler) RunAll(ctx context.Context) error {
	for _, t := range c.Source.Targets() {
		if _, err := c.SweepTarget(ctx, t); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			c.Log.Error("sweep failed", "target", t, "err", err)
		}
	}
	return c.FetchDetails(ctx)
}

// SweepTarget sweeps one target and records the run.
func (c *Crawler) SweepTarget(ctx context.Context, target string) (store.Run, error) {
	start := c.now()
	run := store.Run{Source: c.Source.Name(), Target: target, StartedAt: start}
	id, err := c.Store.StartRun(ctx, run.Source, target, start)
	if err != nil {
		return run, err
	}
	run.ID = id
	c.Log.Info("sweep start", "source", run.Source, "target", target)

	seen := map[string]bool{} // newest-first paging can repeat items across pages
	complete, sweepErr := c.Source.Sweep(ctx, target, c.MaxPages, func(page int, items []listing.Listing) error {
		run.Pages = page
		for _, l := range items {
			if seen[l.ID()] {
				continue
			}
			seen[l.ID()] = true
			res, err := c.Store.RecordSeen(ctx, target, l, c.now())
			if err != nil {
				return err
			}
			run.Seen++
			if res.New {
				run.New++
			}
			if res.OldPrice != 0 {
				run.PriceChanges++
				c.Log.Info("price change", "id", l.ID(), "from", res.OldPrice, "to", l.PriceJPY)
			}
		}
		c.Log.Debug("page", "target", target, "page", page, "bags", len(items))
		return nil
	})
	run.Complete = complete && sweepErr == nil
	if run.Complete {
		n, err := c.Store.SweepMissing(ctx, run.Source, target, start, c.now())
		if err != nil {
			sweepErr = err
		}
		run.Gone = n
	}
	if sweepErr != nil {
		run.Error = sweepErr.Error()
	}
	run.FinishedAt = c.now()
	// Record the run even if the caller's context was cancelled mid-sweep.
	if err := c.Store.FinishRun(context.WithoutCancel(ctx), run); err != nil {
		return run, err
	}
	c.Log.Info("sweep done", "target", target, "pages", run.Pages, "seen", run.Seen, "new", run.New,
		"price_changes", run.PriceChanges, "gone", run.Gone, "complete", run.Complete,
		"took", run.FinishedAt.Sub(start).Round(time.Second))
	if sweepErr != nil {
		return run, fmt.Errorf("sweep %s: %w", target, sweepErr)
	}
	return run, nil
}

// FetchDetails enriches up to DetailBudget listings that lack product-page data.
func (c *Crawler) FetchDetails(ctx context.Context) error {
	if c.DetailBudget <= 0 {
		return nil
	}
	todo, err := c.Store.PendingDetails(ctx, c.Source.Name(), c.DetailBudget)
	if err != nil {
		return err
	}
	var ok, gone, failed int
	for _, t := range todo {
		d, isGone, err := c.Source.Detail(ctx, t.URL)
		var saveErr error
		switch {
		case ctx.Err() != nil:
			return ctx.Err()
		case err != nil:
			failed++
			c.Log.Warn("detail failed", "id", t.ID, "err", err)
		case isGone:
			gone++
			saveErr = c.Store.MarkGone(ctx, t.ID, c.now())
		default:
			ok++
			saveErr = c.Store.SaveDetail(ctx, t.ID, d, c.now())
		}
		if saveErr != nil {
			return saveErr
		}
	}
	c.Log.Info("details done", "fetched", ok, "gone", gone, "failed", failed)
	return nil
}
