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
	// each page's bag listings and the source ids of listings the page shows as
	// sold (nil when the shop's list pages show none). Sold ids are never
	// stored as listings; they only mark gone a listing already held.
	// complete is true when the last page was reached (as opposed to stopping
	// at maxPages, 0 = unlimited).
	Sweep(ctx context.Context, target string, maxPages int, visit func(page int, items []listing.Listing, sold []string) error) (complete bool, err error)
	// Detail fetches a product page. gone is true if the page no longer exists
	// or says the item is sold.
	Detail(ctx context.Context, url string) (d listing.Detail, gone bool, err error)
}

type Crawler struct {
	Store *store.Store
	// Sources are swept one after another, never concurrently: SQLite has one
	// writer, and each source's Fetcher already paces requests to its own site.
	Sources []Source
	Log     *slog.Logger
	// MaxPages caps pages per target per sweep (0 = all). A capped sweep never
	// counts a listing it did not see as missed (no SweepMissing), because it
	// can't tell "sold" from "not reached"; sold ids on the pages it read and
	// re-checks (RecheckAfter) still mark listings gone.
	MaxPages int
	// DetailBudget caps product-page fetches per source per run; list pages
	// already carry everything the radar needs, details only enrich (colour,
	// size, accessories).
	// Rejected: one budget shared by all sources — a shop with a large backlog
	// (its first crawl) would use it all and starve the others' details. Each
	// Fetcher paces only its own site, so total fetches per run scale with the
	// number of sources (N sources → up to N × DetailBudget).
	// Re-checks (RecheckAfter) come out of the same budget, after first fetches.
	DetailBudget int
	// RecheckAfter is the least time between two product-page reads of one
	// listing when re-checking (0 = no re-checks). Re-checks cover listings of
	// targets whose sweep did not complete in this crawl, which SweepMissing
	// cannot mark gone, that the crawl's list pages did not show, and spend
	// only what first fetches leave of DetailBudget: a run stays within
	// DetailBudget product pages per source, and while re-check candidates
	// exist it spends up to all of it, not only what new listings need.
	// Rejected: a separate re-check budget
	// (more requests per run than -details states); re-checking every source
	// (complete sweeps already mark gone after GoneAfterMissedSweeps); a fixed
	// per-source "never completes" flag (GALLERY RARE brands with one page of
	// bags do complete, and a BRAND OFF sweep that is capped or fails does not).
	RecheckAfter time.Duration
	Now          func() time.Time
}

func (c *Crawler) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// RunAll sweeps every target of every source, then spends each source's
// detail budget (first fetches, then re-checks of the targets whose sweep did
// not complete). A failed sweep is logged and the rest still run.
func (c *Crawler) RunAll(ctx context.Context) error {
	start := c.now()
	incomplete := map[string][]string{} // source -> targets whose sweep did not complete
	for _, src := range c.Sources {
		for _, t := range src.Targets() {
			run, err := c.SweepTarget(ctx, src, t)
			if !run.Complete {
				incomplete[src.Name()] = append(incomplete[src.Name()], t)
			}
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				c.Log.Error("sweep failed", "source", src.Name(), "target", t, "err", err)
			}
		}
	}
	for _, src := range c.Sources {
		if err := c.FetchDetails(ctx, src, incomplete[src.Name()], start); err != nil {
			return err
		}
	}
	return nil
}

// SweepTarget sweeps one target of src and records the run.
func (c *Crawler) SweepTarget(ctx context.Context, src Source, target string) (store.Run, error) {
	start := c.now()
	run := store.Run{Source: src.Name(), Target: target, StartedAt: start}
	id, err := c.Store.StartRun(ctx, run.Source, target, start)
	if err != nil {
		return run, err
	}
	run.ID = id
	c.Log.Info("sweep start", "source", run.Source, "target", target)

	seen := map[string]bool{} // newest-first paging can repeat items across pages
	complete, sweepErr := src.Sweep(ctx, target, c.MaxPages, func(page int, items []listing.Listing, sold []string) error {
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
		for _, sid := range sold {
			id := listing.Listing{Source: run.Source, SourceID: sid}.ID()
			marked, err := c.Store.MarkGone(ctx, id, c.now())
			if err != nil {
				return err
			}
			if marked {
				run.Gone++
				c.Log.Info("sold on list page", "id", id)
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
		run.Gone += n // after list-row sold marks, which it must not overwrite
	}
	if sweepErr != nil {
		run.Error = sweepErr.Error()
	}
	run.FinishedAt = c.now()
	// Record the run even if the caller's context was cancelled mid-sweep.
	if err := c.Store.FinishRun(context.WithoutCancel(ctx), run); err != nil {
		return run, err
	}
	c.Log.Info("sweep done", "source", run.Source, "target", target, "pages", run.Pages, "seen", run.Seen, "new", run.New,
		"price_changes", run.PriceChanges, "gone", run.Gone, "complete", run.Complete,
		"took", run.FinishedAt.Sub(start).Round(time.Second))
	if sweepErr != nil {
		return run, fmt.Errorf("sweep %s %s: %w", run.Source, target, sweepErr)
	}
	return run, nil
}

// FetchDetails spends src's detail budget: first on listings without
// product-page data, then, with what is left and when RecheckAfter is set, on
// re-checks (store.RecheckDue) of listings in recheck — targets whose sweep did
// not complete in this crawl — that no list page showed since crawlStart.
// A read's photos replace the listing's gallery (store.SaveDetail), so
// galleries cost no request beyond the product page itself.
func (c *Crawler) FetchDetails(ctx context.Context, src Source, recheck []string, crawlStart time.Time) error {
	if c.DetailBudget <= 0 {
		return nil
	}
	todo, err := c.Store.PendingDetails(ctx, src.Name(), c.DetailBudget)
	if err != nil {
		return err
	}
	first := len(todo)
	if left := c.DetailBudget - first; left > 0 && c.RecheckAfter > 0 {
		due, err := c.Store.RecheckDue(ctx, src.Name(), recheck, crawlStart, c.now().Add(-c.RecheckAfter), left)
		if err != nil {
			return err
		}
		todo = append(todo, due...)
	}
	var ok, gone, failed, galleries int
	for _, t := range todo {
		d, isGone, err := src.Detail(ctx, t.URL)
		var saveErr error
		switch {
		case ctx.Err() != nil:
			return ctx.Err()
		case err != nil:
			failed++
			c.Log.Warn("detail failed", "id", t.ID, "err", err)
		case isGone:
			gone++
			_, saveErr = c.Store.MarkGone(ctx, t.ID, c.now())
		default:
			ok++
			if len(d.Images) > 0 {
				galleries++
			}
			saveErr = c.Store.SaveDetail(ctx, t.ID, d, c.now())
		}
		if saveErr != nil {
			return saveErr
		}
	}
	c.Log.Info("details done", "source", src.Name(), "first", first, "rechecks", len(todo)-first,
		"fetched", ok, "galleries", galleries, "gone", gone, "failed", failed)
	return nil
}
