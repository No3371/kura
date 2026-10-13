# Kura Multi-Source Crawl + Per-Pair Onboarding Epoch

> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 1
> **Status:** Complete
> **Author:** Claude Opus 5.5 (plan, sprint iteration 1)
> **Parent:** 2610131822-kura-def-coverage-sprint.md
> **Source:** Direct request — sprint derive, iteration 1 (uid `2610131822-kura-def-coverage#1`)
> **Related Projex:** 2610131843-kura-multi-source-onboarding-epoch-patch.md (execution, commit `072c15d`) | 2610131822-kura-def-coverage-sprint.md | 2610130449-kura-mock-def.md | 2610130420-kura-roadmap-nav.md (Known Blockers R1) | 2610131037-kura-member-notifications-audit.md (R1 origin) | 2610131048-kura-newest-order-member-feed-patch.md (current newest order)
> Partial Execution: Steps 1–5 + Verification Plan completed via 2610131843-kura-multi-source-onboarding-epoch-patch.md — nothing remains open
> **Worktree:** No — runs on the sprint worktree `kura/.projexwt/2610131822-kura-def-coverage`, branch `projex/sprint/2610131822-kura-def-coverage` (strategy `plan → patch! → audit → [patch]`); no further worktree/branch

---

## Summary

Make crawl → radar multi-source-safe before any second provider lands. (1) `crawl.Crawler` runs a list of `crawl.Source`s; `cmd/kura` `setup` wires that list (BRAND OFF only, this iteration) and backfills grades for every registered `listing.Scheme`. (2) Fix R1: "Kura watched it appear" stops using one global epoch (earliest `first_seen` + 24 h) and uses a per-(source, target) onboarding epoch = `finished_at` of that pair's first complete sweep in `crawl_runs`. 最新上架 order (`Engine.newer`) and days-to-sell (`Engine.stats`) share the one predicate.

**Scope:** `internal/store` (onboarding query + version) | `internal/radar` (watched predicate) | `internal/crawl` (multi-source crawler + first tests) | `cmd/kura` wiring | `internal/web/server_test.go` fixture | README prose. No new provider, no UI change.
**Estimated Changes:** 9 files (5 code, 3 test incl. 1 new, README); ~6 functions changed, 2 added.

---

## Objective

### Problem / Gap / Need

- **Single-source ingestion.** `crawl.Crawler.Source` is one `Source`; `RunAll`, `SweepTarget`, `FetchDetails` all read `c.Source`. `cmd/kura/main.go` `setup` builds one `brandoff.Source` and calls `BackfillGrades(ctx, brandoff.Scheme)` only. Provider #2 (sprint step c) cannot be added without rewriting this.
- **R1 — global epoch.** `radar.go` `Engine.watched` = `it.FirstSeen.After(e.epoch.Add(24 * time.Hour))`, `e.epoch` = earliest `first_seen` over all rows (`Engine.build`). Two failures:
  - A source (or target) onboarded later than the first has its whole initial sweep read as watched → floods 最新上架 / Browse `sort=newest` / advisor `newest` / member 新上架 pick, and enters days-to-sell with fake listing dates.
  - A genuinely new listing inside epoch + 24 h sorts as initial-crawl — today: `brandoff:2101220126889` (CHANEL, first seen 2026-10-12 17:41:20 UTC in run 9, after CHANEL's first complete sweep ended 16:36:31) is position 15,367 of 15,367 in 最新上架 (audit 2610131037 § R1).

### Success Criteria

- [x] `crawl.Crawler` has `Sources []Source`; `RunAll` sweeps every target of every source, a failed sweep of one source/target does not stop the rest, then spends the detail budget per source on that source's own listings only.
- [x] `crawl_runs` gets one row per (source, target) swept, with the right `source` — asserted by new `internal/crawl/crawler_test.go` (two fake sources).
- [x] `cmd/kura` `setup` builds a `[]crawl.Source` (BRAND OFF only), refuses a source without a registered `listing.Scheme` or with a duplicate name, and calls `BackfillGrades(ctx, listing.Schemes()...)`.
- [x] `store.Store.Onboarded` returns, per (source, target), the earliest `finished_at` among `complete=1` runs; incomplete / unfinished runs ignored; `Version` changes when a complete run is recorded.
- [x] Radar: listing is watched ⇔ its (source, target) has an onboarding time and `first_seen` is strictly after it. New `internal/radar` test: two sources onboarded on different dates → later source's initial-sweep listings sort as initial-crawl (not watched); listing first seen after its own pair's onboarding sorts first; pair with no complete sweep → not watched; days-to-sell counts only watched gone listings.
- [x] Existing tests keep their meaning (radar test helper reproduces the old t0 + 24 h boundary via onboarding; `TestMemberAPI` records complete runs) — all green.
- [x] On a copy of today's `kura.db`: `/api/listings?sort=newest` order = `[brandoff:2101220126889] + baseline order minus that id` (the R1 fix is the only delta); `/api/overview` `overview.since` and `sold_stats_ready` unchanged.
- [x] `go vet ./...` + `go test ./...` green.

### Out of Scope

- Any new provider, provider vetting, `dist.js` per-source dots, advisor prompt shop line (sprint steps b–c).
- `Overview.Since` and `SoldStatsReady` / `MinSoldHistory` — stay on the global epoch; per-shop sold-history gating is an open/provisional question (`MinSoldHistory` comment) → deferred by user decision.
- Two-tier realtime scheduler, concurrent per-source sweeping (sprint step e).
- Running `kura crawl` against BRAND OFF (no live crawl this iteration; crawl path proven with fakes).
- UI / copy changes (`pages.js` 找到新主人 copy — "開始記錄時已經在架上的商品" — stays true per shop).
- Cross-source ordering of initial-crawl listings (no shop-recency signal exists across pairs; ascending `first_seen` kept, as across targets today).

---

## Context

### Current State

- `internal/crawl/crawler.go`: `Crawler{Store, Source, Log, MaxPages, DetailBudget, Now}`. `RunAll` → `for t := range c.Source.Targets() { SweepTarget }` → `FetchDetails`. `SweepTarget(ctx, target)` records a `crawl_runs` row via `Store.StartRun` / `FinishRun` with `Source: c.Source.Name()`; `Complete` = source said complete and no error; `SweepMissing` only on complete. `FetchDetails` → `Store.PendingDetails(ctx, c.Source.Name(), c.DetailBudget)` → `c.Source.Detail`. No tests in package.
- `cmd/kura/main.go` `setup`: `BackfillGrades(ctx, brandoff.Scheme)`; `src := &brandoff.Source{Fetcher: crawl.NewFetcher(c.delay), Brands: brands}`; `&crawl.Crawler{Store: st, Source: src, ...}`. Only other `Crawler` user: `deps.crawlOnce` → `RunAll`.
- `internal/store/store.go`: `listings.target` = crawl target that discovered the row (insert-only); `crawl_runs(source, target, started_at, finished_at, complete, …)`. `Version` = `COUNT(listings)/max(last_seen)+max(gone_at)+max(detail_at)/COUNT(price_points)` — blind to `crawl_runs`. `BackfillGrades(ctx, schemes ...listing.Scheme)` already variadic.
- `internal/radar/radar.go`: `Engine.Refresh` → `Store.Version`, `Store.AllRows` → `build(rows, version)`. `build` computes `epoch` (min `first_seen`) + `latest`. `watched(it)` used by `newer` (sort `newest`, `NewestActive`) and `stats` (days-to-sell). `epoch` also feeds `soldStatsReady` and `Overview.Since` (unchanged by this plan).
- `listing.Schemes()` returns every registered scheme (prod: `brandoff` via `init`).
- Today's `kura.db` (main checkout, read-only inspection): 9 runs, all `brandoff`; first complete sweep per target finished 2026-10-12 16:25:28 (GOYARD) … 16:45:53 (LOEWE); run 9 (CHANEL, 17:39:59) unfinished (`finished_at` NULL, complete 0). Every listing first seen ≤ its pair's first complete `finished_at` except `brandoff:2101220126889`. No gone listings.
- Baseline `go vet ./...` + `go test ./...` green at `526ada6`.

### Key Files

| File | Role | Change Summary |
|------|------|----------------|
| `internal/store/store.go` | Persistence | + `Pair`, + `Onboarded`; `Version` counts complete runs |
| `internal/store/store_test.go` | Store tests | + `TestOnboarded` |
| `internal/radar/radar.go` | Signals | per-pair `watched` computed in `build`; `newer` / `stats` use it; `Refresh` loads onboarding |
| `internal/radar/radar_test.go` | Radar tests | helper passes onboarding; + `TestWatchedPerSourceOnboarding` |
| `internal/crawl/crawler.go` | Crawler | `Source` → `Sources []Source`; per-source sweep + details |
| `internal/crawl/crawler_test.go` | **New** | fake sources; multi-source run + per-source details |
| `cmd/kura/main.go` | Wiring | source list, scheme/duplicate check, backfill all schemes, `-details` help |
| `internal/web/server_test.go` | API tests | `TestMemberAPI` records complete runs for its two pairs |
| `README.md` | Shipped doc | sold/newest prose → per-shop-brand first complete sweep; `-details` per source |

### Dependencies

- **Requires:** nothing (sprint branch at `526ada6`).
- **Blocks:** sprint steps (b)–(e): every new provider crawls through this path and relies on per-pair onboarding.

### Constraints

- SKILL.md § Source Hygiene: no projex ids/filenames in code comments; symbols not line numbers; non-obvious decisions carry rationale naming the rejected alternative.
- Crawl invariants unchanged (Fetcher, robots, UA untouched). Sweeps stay sequential (SQLite one writer; per-site politeness lives in each source's own `Fetcher`).
- Never write to the main checkout's `kura.db`; verification uses a copy in a scratch dir.
- No push to any remote. Dev servers stopped by PID only.

### Assumptions

- A pair's first complete sweep took in that pair's whole catalogue → everything first seen at or before its `finished_at` is pre-existing stock. (Same assumption the 24 h window made, now bound to evidence.) Verify early: Step 2 test + Verification § kura.db copy.
- `listings.target` is the same string as `crawl_runs.target` for the same sweep (both from `SweepTarget`'s `target` arg) — verified in code (`RecordSeen(ctx, target, …)`, `StartRun(ctx, …, target, …)`).
- `first_seen` and `finished_at` are unix seconds; a listing first seen in the same second the sweep finished belongs to that sweep → strict `After` is correct.
- Unfinished runs (`finished_at` NULL, process killed) are never complete (`complete` defaults 0) — verified in schema + today's run 9.

### Impact Analysis

- **Direct:** files above.
- **Adjacent:** `internal/web/member.go` (`NewestActive` → 新上架 pick), `internal/web/server.go` listings `sort=newest`, `internal/advisor/tools.go` (`newest` sort, model stats days-to-sell) — behaviour follows the predicate, no code change. `/api/overview` `runs` already carries `source` per run.
- **Downstream:** pairs without a complete sweep (capped `-max-pages` only, or always-failing) now have no watched listings; old rule watched them after 24 h. Intended (fail-closed) — documented in the predicate comment and README.

---

## Implementation

### Overview

Store exposes onboarding times → radar computes one `watched` bit per item at build time → `newer` / `stats` read the bit. Independently, crawler iterates a source list; `cmd/kura` builds that list. Web test fixture and README follow.

### Step 1: Store — onboarding times + version [PATCHED]

> [PATCHED] via 2610131843-kura-multi-source-onboarding-epoch-patch.md

**Objective:** per-(source, target) end of first complete sweep; radar rebuilds when one is recorded.
**Confidence:** High
**Depends on:** None
**Do-Projex:** Encouraged

**Files:**
- `internal/store/store.go`
- `internal/store/store_test.go`

**Changes:**

```go
// Before (store.go, Version):
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(MAX(last_seen),0) +
		COALESCE(MAX(gone_at),0) + COALESCE(MAX(detail_at),0), (SELECT COUNT(*) FROM price_points) FROM listings`).
		Scan(&n, &last, &pts)
	return fmt.Sprintf("%d/%d/%d", n, last, pts), err

// After:
	var runs int64
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(MAX(last_seen),0) +
		COALESCE(MAX(gone_at),0) + COALESCE(MAX(detail_at),0), (SELECT COUNT(*) FROM price_points),
		(SELECT COUNT(*) FROM crawl_runs WHERE complete=1) FROM listings`).
		Scan(&n, &last, &pts, &runs)
	return fmt.Sprintf("%d/%d/%d/%d", n, last, pts, runs), err
```

Update the `Version` doc comment: changes whenever listing data changes or a sweep completes (the radar's watched rule reads complete sweeps).

Add after `RecentRuns`:

```go
// Pair is one crawl target of one source.
type Pair struct{ Source, Target string }

// Onboarded returns, per source and target, when its first complete sweep
// finished. Runs that stopped early (capped, failed, or never finished) are
// not onboarding: they saw only part of the catalogue.
func (s *Store) Onboarded(ctx context.Context) (map[Pair]time.Time, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT source, target, MIN(finished_at) FROM crawl_runs
		WHERE complete=1 AND finished_at IS NOT NULL GROUP BY source, target`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[Pair]time.Time{}
	for rows.Next() {
		var p Pair
		var at int64
		if err := rows.Scan(&p.Source, &p.Target, &at); err != nil {
			return nil, err
		}
		out[p] = time.Unix(at, 0)
	}
	return out, rows.Err()
}
```

`store_test.go` + `TestOnboarded`: on `open(t)`: `Onboarded` empty; `StartRun`/`FinishRun` for (`src`,`HERMES`) incomplete at day 1 → still absent and `Version` unchanged; complete at day 2 → `Pair{"src","HERMES"}` = day 2 and `Version` changed; another complete at day 3 → still day 2 (MIN); a run started but never finished (only `StartRun`) for (`src`,`CHANEL`) → absent; complete (`other`,`HERMES`) at day 4 → separate key.

**Rationale:** `MIN(finished_at)` over complete runs is the earliest moment the whole catalogue had been seen. Adding the complete-run count to `Version` avoids a stale index when a sweep completes without changing any listing row (empty target, or a refresh that read rows mid-sweep).

**Verification:** `go test ./internal/store/`.

**If this fails:** `git -C <work-root> checkout -- internal/store/`.

---

### Step 2: Radar — per-pair watched predicate [PATCHED]

> [PATCHED] via 2610131843-kura-multi-source-onboarding-epoch-patch.md

**Objective:** replace the global-epoch `watched` with a per-item bit from per-pair onboarding; `newer` and `stats` share it.
**Confidence:** High
**Depends on:** Step 1
**Verify-Projex:** Encouraged

**Files:**
- `internal/radar/radar.go`
- `internal/radar/radar_test.go`

**Changes:**

```go
// Before (Item, after History):
	History         []PricePoint `json:"history,omitempty"`
}
// After:
	History         []PricePoint `json:"history,omitempty"`
	// watched: Kura saw the listing appear (see watched).
	watched bool
}
```

```go
// Before (Refresh):
	rows, err := e.Store.AllRows(ctx)
	if err != nil {
		return err
	}
	e.build(rows, v)
// After:
	rows, err := e.Store.AllRows(ctx)
	if err != nil {
		return err
	}
	onboarded, err := e.Store.Onboarded(ctx)
	if err != nil {
		return err
	}
	e.build(rows, onboarded, v)
```

```go
// Before (build):
func (e *Engine) build(rows []store.Row, version string) {
	...
		it := e.toItem(r, now)
// After:
func (e *Engine) build(rows []store.Row, onboarded map[store.Pair]time.Time, version string) {
	...
		it := e.toItem(r, now)
		it.watched = watched(r, onboarded)
```
(`epoch` / `latest` computation and the `e.epoch` field stay — still used by `soldStatsReady` and `Overview.Since`. Update the `epoch` field comment to: `// earliest first_seen = when crawling began (Overview.Since, sold-stats gate)`.)

```go
// Before:
// watched reports whether Kura saw the listing appear: first seen after the
// initial crawl, which takes in the whole catalogue as already listed. Days to
// sell and 最新上架 order both turn on it.
func (e *Engine) watched(it *Item) bool {
	return it.FirstSeen.After(e.epoch.Add(24 * time.Hour))
}

// After:
// watched reports whether Kura saw r appear on its shop: first seen after its
// own source and target finished onboarding — the end of that pair's first
// complete sweep, which takes in the whole catalogue as already listed. Days
// to sell and 最新上架 order both turn on it.
// Rejected: one epoch for the whole database (earliest first_seen + 24 h). A
// shop or brand Kura starts crawling later then has its entire initial sweep
// read as newly listed, flooding 最新上架 and giving days to sell made-up
// listing dates; and a listing that really appeared inside the 24 h sorts as
// old stock. A pair with no complete sweep has no baseline, so nothing in it
// is watched: a capped sweep reaches only the first list pages, and old stock
// rises into them whenever newer listings above it sell.
func watched(r *store.Row, onboarded map[store.Pair]time.Time) bool {
	at, ok := onboarded[store.Pair{Source: r.Source, Target: r.Target}]
	return ok && r.FirstSeen.After(at)
}
```

```go
// Before (newer):
	wa, wb := e.watched(a), e.watched(b)
// After:
	wa, wb := a.watched, b.watched
```
In the `newer` doc comment, "Initial-crawl listings follow…" → "Listings from their pair's initial sweep follow…" (rest unchanged). `newer` may stay a method (no `e` use left) or become a plain func — keep the method to avoid touching callers.

```go
// Before (stats):
			if it.GoneAt != nil && e.watched(it) {
// After:
			if it.GoneAt != nil && it.watched {
```
`ModelStats.MedianDaysToSell` comment: "(first seen after the initial crawl)" → "(first seen after its shop and brand's first complete sweep)".

`radar_test.go`:
- `engine(rows...)` → `e.build(rows, map[store.Pair]time.Time{{Source: "t"}: t0.Add(24 * time.Hour)}, "test")` — reproduces the old boundary (t0 + 24 h) for source `t`, target `""`, so every existing test keeps its meaning. Add `engineAt(onboarded map[store.Pair]time.Time, rows ...store.Row) *Engine` used by the new test; `engine` calls it.
- New `TestWatchedPerSourceOnboarding` (no scheme registration needed; reuse `row`, then set `Source`, `ID` (`"u:"+id` for source `u`), `Target`, and `FirstSeen`; gone rows set `Status`/`GoneAt` as `TestDropsAndStats` does):
  - Onboarding: `{t, LV}` = t0 + 1 h; `{u, LV}` = day 10 + 1 h; `{u, HERMES}` absent (only incomplete sweeps).
  - Rows (all one model): `t:i` (t, LV, first seen t0) | `t:w` (t, LV, day 5) | `u:i1` (u, LV, day 10) | `u:i2` (u, LV, day 10 + 30 min) | `u:w` (u, LV, day 12) | `u:x` (u, HERMES, day 12, same model code) | `u:ig` gone (u, LV, first day 10, gone day 11) | `u:wg` gone (u, LV, first day 12, gone day 15).
  - Assert `Search(Filter{Sort: "newest"})` ids = `u:w, t:w, t:i, u:i1, u:i2, u:x` (watched latest-first; then initial ascending `first_seen`; `u:x` unwatched despite late `first_seen`; gone rows excluded — `Filter.IncludeGone` defaults false).
  - Assert `NewestActive` over the model's items = `u:w`.
  - Assert `Model(...)` `MedianDaysToSell` = 3 (only `u:wg`; `u:ig` excluded — old rule would give median(1, 3) = 2).
  - Test comment states why: under one global epoch (t0 + 24 h) `u:i1`, `u:i2`, `u:ig` would all read as watched — the case this test guards.

**Rationale:** computing the bit once in `build` keeps one predicate for both consumers and needs no `Target` on the API `Item`. Unexported field → no JSON change.

**Verification:** `go test ./internal/radar/` — existing tests unchanged in assertions + new test green. `go vet ./internal/radar/`.

**If this fails:** `git -C <work-root> checkout -- internal/radar/`.

---

### Step 3: Crawler — several sources [PATCHED]

> [PATCHED] via 2610131843-kura-multi-source-onboarding-epoch-patch.md

**Objective:** `Crawler` runs every configured source; details budget per source.
**Confidence:** High
**Depends on:** None (parallel to Steps 1–2)
**Verify-Projex:** Encouraged

**Files:**
- `internal/crawl/crawler.go`
- `internal/crawl/crawler_test.go` (new)

**Changes:**

```go
// Before:
type Crawler struct {
	Store  *store.Store
	Source Source
	Log    *slog.Logger
	...
	// DetailBudget caps product-page fetches per run; list pages already carry
	// everything the radar needs, details only enrich (colour, size, accessories).
	DetailBudget int

// After:
type Crawler struct {
	Store *store.Store
	// Sources are swept one after another, never concurrently: SQLite has one
	// writer, and each source's Fetcher already paces requests to its own site.
	Sources []Source
	Log     *slog.Logger
	...
	// DetailBudget caps product-page fetches per source per run; list pages
	// already carry everything the radar needs, details only enrich (colour,
	// size, accessories).
	DetailBudget int
```

```go
// Before:
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

// After:
// RunAll sweeps every target of every source, then spends each source's
// detail budget. A failed sweep is logged and the rest still run.
func (c *Crawler) RunAll(ctx context.Context) error {
	for _, src := range c.Sources {
		for _, t := range src.Targets() {
			if _, err := c.SweepTarget(ctx, src, t); err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				c.Log.Error("sweep failed", "source", src.Name(), "target", t, "err", err)
			}
		}
	}
	for _, src := range c.Sources {
		if err := c.FetchDetails(ctx, src); err != nil {
			return err
		}
	}
	return nil
}
```

- `SweepTarget(ctx, target)` → `SweepTarget(ctx context.Context, src Source, target string)`; every `c.Source` inside → `src`; `"sweep done"` log gains `"source", run.Source`; returned error `fmt.Errorf("sweep %s %s: %w", run.Source, target, sweepErr)`.
- `FetchDetails(ctx)` → `FetchDetails(ctx context.Context, src Source)`; doc: "enriches up to DetailBudget of src's listings that lack product-page data"; `PendingDetails(ctx, src.Name(), …)`, `src.Detail(…)`; `"details done"` log gains `"source", src.Name()`.

`crawler_test.go` (package `crawl`): fake `Source` — `name`, `targets`, `pages map[string][][]listing.Listing`, `fail map[string]bool` (target → Sweep returns error after first page), records `details []string` URLs; `Detail` returns `listing.Detail{Color: "c"}`. Real store via `store.Open(filepath.Join(t.TempDir(), "kura.db"))`; `Log: slog.New(slog.NewTextHandler(io.Discard, nil))`; fixed `Now`.
- `TestRunAllSweepsEverySource`: sources `fa` (targets `HERMES`, `CHANEL`; `CHANEL` fails) and `fb` (target `HERMES`), `DetailBudget: 10`. After `RunAll`: `RecentRuns` has 3 rows — (`fa`,`HERMES`, complete) | (`fa`,`CHANEL`, not complete, error non-empty) | (`fb`,`HERMES`, complete); listings exist for both sources (`AllRows`, ids `fa:*` and `fb:*`); `fa.details` ⊆ `fa` URLs and `fb.details` ⊆ `fb` URLs, each non-empty; `Onboarded` has `{fa,HERMES}` and `{fb,HERMES}` but not `{fa,CHANEL}`.
- `TestDetailBudgetPerSource`: two sources × 3 listings, `DetailBudget: 2` → each source's fake got exactly 2 detail calls.

**Rationale:** budget per source keeps one shop from starving another and matches per-site politeness; sequential keeps the single-writer store simple (two-tier scheduler is a later sprint step).

**Verification:** `go test ./internal/crawl/`; `go vet ./...` flags the `cmd/kura` caller until Step 4 (expected; do Step 4 before full-suite run).

**If this fails:** `git -C <work-root> checkout -- internal/crawl/crawler.go` and delete `internal/crawl/crawler_test.go`.

---

### Step 4: `cmd/kura` wiring [PATCHED]

> [PATCHED] via 2610131843-kura-multi-source-onboarding-epoch-patch.md

**Objective:** setup builds the source list, validates it, backfills every registered scheme.
**Confidence:** High
**Depends on:** Step 3
**Do-Projex:** Encouraged

**Files:**
- `cmd/kura/main.go`

**Changes:**

```go
// Before:
	if n, err := st.BackfillGrades(ctx, brandoff.Scheme); err != nil {
	...
	src := &brandoff.Source{Fetcher: crawl.NewFetcher(c.delay), Brands: brands}
	cr := &crawl.Crawler{Store: st, Source: src, Log: log, MaxPages: c.maxPages, DetailBudget: c.details}

// After:
	if n, err := st.BackfillGrades(ctx, listing.Schemes()...); err != nil {
	...
	sources := []crawl.Source{
		&brandoff.Source{Fetcher: crawl.NewFetcher(c.delay), Brands: brands},
	}
	names := map[string]bool{}
	for _, src := range sources {
		// Each shop's grades are shown in its own scale with its own
		// definitions; a source without a registered scheme would be
		// relabelled into another shop's or shown without one.
		if _, ok := listing.SchemeFor(src.Name()); !ok {
			return nil, fmt.Errorf("source %q has no registered condition scheme", src.Name())
		}
		// Listing ids are "<source>:<id>": two adapters with one name would
		// overwrite each other's rows.
		if names[src.Name()] {
			return nil, fmt.Errorf("duplicate source %q", src.Name())
		}
		names[src.Name()] = true
	}
	cr := &crawl.Crawler{Store: st, Sources: sources, Log: log, MaxPages: c.maxPages, DetailBudget: c.details}
```
Keep `BackfillGrades` before the source loop (store-level, independent). `-details` flag help: `"max product pages to fetch per source per crawl for colour/size/accessories"`. `crawlOnce` comment: "refreshes FX then sweeps every source's brands".

**Rationale:** the scheme check enforces the sprint invariant (each provider registers its own `listing.Scheme`) at startup rather than at render.

**Verification:** `go build ./...`; `go vet ./...`; `go run ./cmd/kura serve -h` shows new `-details` text.

**If this fails:** `git -C <work-root> checkout -- cmd/kura/main.go`.

---

### Step 5: Web test fixture + README [PATCHED]

> [PATCHED] via 2610131843-kura-multi-source-onboarding-epoch-patch.md

**Objective:** `TestMemberAPI` keeps its meaning under the per-pair rule; shipped docs describe the rule.
**Confidence:** High
**Depends on:** Step 2
**Do-Projex:** Encouraged

**Files:**
- `internal/web/server_test.go`
- `README.md`

**Changes:**

`TestMemberAPI`: after the `hermes(...)` calls, record each pair's first complete sweep ending after its initial-crawl rows (`hold` at t0 + 20 m is the last):

```go
	// The initial crawl of each brand completes at t0+30m: listings first
	// seen later are ones Kura watched appear.
	onboard := func(target string) {
		t.Helper()
		id, err := st.StartRun(ctx, "wm", target, t0)
		if err != nil {
			t.Fatal(err)
		}
		if err := st.FinishRun(ctx, store.Run{ID: id, FinishedAt: t0.Add(30 * time.Minute), Complete: true}); err != nil {
			t.Fatal(err)
		}
	}
	onboard("LOUIS VUITTON")
	onboard("HERMES")
```
Update the existing comment above `t0` ("listings first seen a day or more later…") to "listings first seen after a brand's initial crawl completed…". Assertions unchanged (`wm:new` t0 + 25 h watched; `hnew` vs `hold` initial).

`README.md`:
- § Knowing when something sold: "(first seen after the initial crawl)" → "(first seen after Kura's first complete sweep of that shop's brand)".
- § Newest listings: "Listings Kura saw appear after the initial crawl come first" → "Listings Kura saw appear — first seen after the first complete sweep of their shop and brand — come first"; add one sentence: "A shop or brand added later starts with its own initial sweep, so its existing stock never reads as new; until a shop's brand has had one complete sweep, none of its listings count as seen appearing."
- Flags table `-details`: "Product pages fetched per source per crawl (colour, size, accessories)".

**Rationale:** fixture must express "initial crawl completed" in the store, since the rule now reads `crawl_runs`.

**Verification:** `go test ./internal/web/`.

**If this fails:** `git -C <work-root> checkout -- internal/web/server_test.go README.md`.

---

## Verification Plan

### Automated Checks
- [x] `go vet ./...` clean
- [x] `go test ./...` green; `internal/crawl` no longer `[no test files]`
- [x] New tests present and passing: `TestOnboarded`, `TestWatchedPerSourceOnboarding`, `TestRunAllSweepsEverySource`, `TestDetailBudgetPerSource`
- [x] Source Hygiene grep on changed `.go` files: no `projex`, no `\d{10}-`, no `Step \d`, no audit finding ids (`R1`, `F1`) in comments

### Manual Verification (today's `kura.db`, copy only)
- [x] Copy: `sqlite3 -readonly <main-checkout>/kura.db ".backup <scratch>/kura-copy.db"` (never open the main checkout's db read-write; if `.backup` refuses under `-readonly`, copy `kura.db` + `kura.db-wal` + `kura.db-shm` together while no kura process runs).
- [x] Baseline binary from sprint base: `git -C <work-root> archive 526ada6 | tar -x -C <scratch>/base` → `go build -C <scratch>/base -o <scratch>/kura-base ./cmd/kura`. New binary: `go build -o <scratch>/kura-new ./cmd/kura` from work-root.
- [x] For each binary, against its own fresh copy of the db (the binaries run `BackfillGrades` on start): `serve -db <copy> -addr 127.0.0.1:<port> -advisor mock`, record PID; fetch `/api/listings?sort=newest&limit=5`, `…&limit=2&offset=<total-2>`, `/api/overview`; stop by PID.
- [x] Expected: baseline newest = `[X1..X5]`, baseline last = `brandoff:2101220126889`; new newest = `[brandoff:2101220126889, X1..X4]`, new last = baseline's second-to-last; `total` equal; `overview.since`, `overview.sold_stats_ready`, `overview.active` equal.
- [x] No live crawl against BRAND OFF.

### Acceptance Criteria Validation
| Criterion | How to Verify | Expected Result |
|-----------|---------------|-----------------|
| Multi-source crawler | `TestRunAllSweepsEverySource` | 3 runs across 2 sources; failure isolated; details per source |
| Per-source detail budget | `TestDetailBudgetPerSource` | 2 calls each |
| `setup` wiring | read `cmd/kura/main.go`; `go build` | `[]crawl.Source`, scheme + duplicate check, `Schemes()...` backfill |
| Onboarding query | `TestOnboarded` | MIN complete `finished_at`; incomplete/unfinished ignored; version bumps |
| R1 per-pair watched | `TestWatchedPerSourceOnboarding` | later source's initial sweep not watched; post-onboarding first; no-sweep pair unwatched; days-to-sell 3 |
| Existing semantics | full `go test ./...` | existing assertions unchanged, green |
| kura.db behaviour | manual copy check | only delta = R1 listing moves last → first |
| Hygiene | vet + grep | clean |

---

## Rollback Plan

1. `git -C <work-root> log --oneline 526ada6..HEAD` → identify the execution commit(s).
2. `git -C <work-root> revert <sha>` per execution commit (sprint branch is shared — revert, don't reset).

---

## Notes

### Risks
- **Capped-only deployments lose 新上架 / days-to-sell entirely** (no complete sweep → nothing watched): intended fail-closed; README states it. Sprint step (e) scheduler must keep a periodic full sweep.
- **Pagination artefact:** a listing skipped by the first complete sweep (item sold mid-sweep shifts later pages up) is first seen next sweep and reads as watched. Same exposure as the 24 h rule had after its window; not addressed here.
- **`newer` cross-pair initial order** stays ascending `first_seen` — interleaves shops by crawl date, not shop recency. No signal exists to do better; out of scope.
- **`Item` copies carry the unexported bit** — any code constructing `radar.Item` outside the engine gets `watched=false` (none today: `advisor/tools.go` only copies engine items).

### Split Decision
`No split — single scope, within size budget` (one repo, one `.projex`; ~540 lines / ~31 KB but 5 steps, so the size heuristic — lines/KB AND > 5 steps — does not trip; Steps 1→2→5 and 3→4 are each a coupled chain sized for one patch pass).

### Open Questions
- [ ] (none)
