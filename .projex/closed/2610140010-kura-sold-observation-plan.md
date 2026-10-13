# Kura Sold Observation for Never-Complete Sources

> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 5
> **Status:** Complete
> **Author:** Claude Opus 5.5 (plan, sprint iteration 5)
> **Parent:** 2610131822-kura-def-coverage-sprint.md
> **Source:** Direct request — sprint derive, iteration 5 (uid `2610131822-kura-def-coverage#5`)
> **Related Projex:** 2610131822-kura-def-coverage-sprint.md (§ Current Position: remainder (d′); § Open Questions: sold re-check) | 2610132219-kura-grare-provider-2-audit.md (F1: product page read once) | 2610132339-kura-allu-provider-3-audit.md (F4: list-visible sold rows discarded) | 2610132140-kura-grare-provider-2-plan.md, 2610132307-kura-allu-provider-3-plan.md (adapter precedents) | 2610131831-kura-multi-source-onboarding-epoch-plan.md (crawler, per-source detail budget)
> Partial Execution: Steps 1–4 (all) completed via 2610140023-kura-sold-observation-patch.md
> **Worktree:** No — runs on the sprint worktree `kura/.projexwt/2610131822-kura-def-coverage`, branch `projex/sprint/2610131822-kura-def-coverage` (strategy `plan → patch! → audit → [patch]`); no further worktree/branch

---

## Summary

GALLERY RARE (page 1 only) and ALLU (newest 10 pages) never complete a sweep, so `SweepMissing` never marks their listings gone. Each product page is read once (`PendingDetails`: `detail_at IS NULL`). A bag that sells later stays 在售 forever. This plan adds two sold signals, both through `store.MarkGone`:
1. **Product-page re-check.** It covers active listings of targets whose sweep did not complete in this crawl and that this crawl's list pages did not show. A listing is re-checked at most once per `RecheckAfter` (default 24 h), oldest read first. Re-checks spend only what first fetches leave of each source's `DetailBudget`.
2. **List-visible sold rows.** `crawl.Source.Sweep`'s `visit` gains `sold []string`. ALLU passes the ids of ALLU公式 rows with `stock 0, sale_status 2`, and the crawler marks those listings gone if it holds them.

**Scope:** crawler + store sold lifecycle, `crawl.Source` contract, 3 adapters (ALLU signal; GALLERY RARE / BRAND OFF pass `nil`), `-recheck-after` flag, README + vetting doc. No UI, no radar logic, no new network requests.
**Estimated Changes:** 13 files: `internal/store/store.go` (+test), `internal/crawl/crawler.go` (+test), `internal/source/{allu/parse.go, allu/source.go, grare/source.go, brandoff/source.go}`, `internal/source/allu/{parse_test.go, source_test.go}`, `internal/web/server_test.go` (call-site only), `cmd/kura/main.go`, `README.md`, `docs/provider-vetting.md`.

**Plan-time requests:** none. Evidence comes from the saved fixtures (`internal/source/allu/testdata/list_lv.html`, `detail_sold.html`). The executor makes no requests to any shop either.

---

## Objective

### Problem / Gap / Need
Def § Content kinds: real listings are "refreshed per Realtime cadence". L1 Realtime: the daily full refresh "catches sold items". For 2 of 3 providers nothing ever marks a listing sold after its first product-page read. The effects:
- cards link out to sold pages;
- 已售出 stays near zero for those shops;
- the radar prices comps against stock that no longer exists.

Evidence (re-verified at plan time):
- `store.PendingDetails`: `WHERE source=? AND status='active' AND detail_at IS NULL`. Nothing selects a listing again once `detail_at` is set.
- `crawl.Crawler.SweepTarget`: `SweepMissing` runs only `if run.Complete`.
- `allu.ParseList`: `if it.Stock <= 0 || it.SaleStatus != 1 { continue }`. The sold row is dropped and nothing marks it gone (F4).
- Fixture `list_lv.html`: 30 rows, all ALLU公式, non-partner: 29 × `stock 1, sale_status 1`, 1 × `stock 0, sale_status 2` (id `2117814`). `detail_sold.html` is item `2117814`'s product page (`button_type` 6 = `soldButton`). So list-row sold and product-page sold agree on one real item.

### Success Criteria
- [x] `store.MarkGone(ctx, id, now) (bool, error)` changes only `status='active'` rows. A repeated signal keeps the first `gone_at` and returns `false`. An unknown id is a no-op returning `false`. It no longer writes `detail_at`.
- [x] `store.RecheckDue` returns active rows of one source in the given targets with `detail_at < checkedBefore` and `last_seen < seenBefore`, ordered `detail_at, id`, capped at `limit`. Empty targets return none.
- [x] `crawl.Source.Sweep`'s `visit` is `func(page int, items []listing.Listing, sold []string) error`. All 3 adapters and the test fake compile against it.
- [x] A sold id from `visit` marks the held active listing gone and is added to `run.Gone` (with `SweepMissing`'s count, not overwritten by it). An id Kura does not hold inserts no row.
- [x] `allu.ParseList` on `list_lv.html` returns `Sold == ["2117814"]`. Synthetic rows show that only ALLU公式 non-partner rows with `stock ≤ 0` **and** `sale_status == 2` enter `Sold`.
- [x] `Crawler.FetchDetails` fetches first-fetch todos first and fills only the remaining budget with re-checks. Re-checks cover only targets whose sweep did not complete this crawl, only listings not seen since crawl start, and only listings whose last read is ≥ `RecheckAfter` old. `RecheckAfter == 0` → no re-checks.
- [x] A re-check whose `Detail` reports gone marks the listing gone. A live one updates detail fields and `detail_at`, so the next crawl rotates to other listings.
- [x] `kura crawl|serve -recheck-after` flag (default `24h`) is wired to `Crawler.RecheckAfter`.
- [x] README (fetcher bullets, § Knowing when something sold, flags table, Tests, Limitations) and `docs/provider-vetting.md` (GALLERY RARE crawl limits, ALLU re-vet Outcome) state the new behaviour. No live text still claims "fetched once per listing" / "Nothing re-checks".
- [x] `gofmt -l .` empty; `go vet ./...` + `go test -count=1 ./...` green.

### Out of Scope
- Owner-held decisions (sprint Open Questions), unchanged: ALLU consignment keep/exclude (F1), 第18条 (12) reading (F3), stop-on-403 scope (F7), GALLERY RARE `?page=` paging.
- Per-provider history gate (`Overview.Since` / `SoldStatsReady`): Deferred (def L4 Day-1 open question).
- Realtime two-tier scheduler; hosting.
- ALLU F2 title-derived model fallback (remainder (e)).
- `sale_status` / stock drift alerting (informational re-audit note).
- Any UI / radar change. Gone listings already flow through `radar` (`Gone`, comps window, member events).
- Status checks in GALLERY RARE / BRAND OFF `Detail`. The fetcher already returns an error for every status ≥ 400 except 404 (`fetch.go`), so error pages never reach `SaveDetail`.

---

## Context

### Current State
- `crawl.Crawler.RunAll`: sweeps every target of every source (`SweepTarget`), then `FetchDetails(ctx, src)` per source.
- `SweepTarget`: dedupes ids per sweep, `RecordSeen` each, and `if run.Complete` → `run.Gone = SweepMissing(...)`. The `visit` closure receives only kept cards.
- `FetchDetails`: `PendingDetails(source, DetailBudget)` → `Detail(url)`:
  - error → `failed++`;
  - gone → `MarkGone`;
  - else → `SaveDetail` (sets `detail_at`, model code override).
- `store.MarkGone`: `UPDATE listings SET status='gone', gone_at=?, detail_at=? WHERE id=?`. It is unconditional, so repeating it moves `gone_at` forward. Callers: `crawl.FetchDetails`, `internal/web/server_test.go` (`st.MarkGone(ctx, "wm:cut", …)`).
- `store.Row.HasDetail` (← `detail_at`) is read by nothing outside `store` (grep: radar/web do not use it). So dropping the `detail_at` write from `MarkGone` has no UI effect.
- `RecordSeen` on an existing row: `status='active', gone_at=NULL, missed_sweeps=0`, `last_seen=now`; `detail_at` is untouched.
- Adapters' `Sweep` call `visit(page, cards)`:
  - `brandoff/source.go`: `visit(page, bags)`;
  - `grare/source.go`: `visit(1, lp.Cards)`;
  - `allu/source.go`: `visit(page, lp.Cards)`.
- Test callers: `allu/source_test.go` × 3 closures plus one `nil` visit; `crawl/crawler_test.go` `fakeSource`.
- `cmd/kura/main.go` `setup`: `&crawl.Crawler{Store: st, Sources: sources, Log: log, MaxPages: c.maxPages, DetailBudget: c.details}`.
- Fetcher (`crawl/fetch.go`): status ≥ 400 except 404 → `statusError` (retry on 429/5xx). Adapters see only 2xx/3xx/404 responses.

### Key Files

| File | Role | Change Summary |
|------|------|----------------|
| `internal/store/store.go` | Listing lifecycle | `MarkGone` active-only `(bool, error)`, no `detail_at`; new `RecheckDue`; shared row scan |
| `internal/store/store_test.go` | Store tests | `TestMarkGoneOnlyActive`, `TestRecheckDue` |
| `internal/crawl/crawler.go` | Source contract + crawler | `visit` gains `sold`; `SweepTarget` marks sold ids gone (`run.Gone +=`); `RecheckAfter`; `RunAll` collects incomplete targets; `FetchDetails(ctx, src, recheck, crawlStart)` |
| `internal/crawl/crawler_test.go` | Crawler tests | fake gains `sold` / `capped` / `goneURL`; sold-row + re-check tests |
| `internal/source/allu/parse.go` | ALLU list parser | `ListPage.Sold`, `soldStatus = 2` |
| `internal/source/allu/source.go` | ALLU sweeper | `visit(page, lp.Cards, lp.Sold)`; doc comment |
| `internal/source/allu/parse_test.go` | ALLU parser tests | fixture `Sold`; filter rows for sold |
| `internal/source/allu/source_test.go` | ALLU sweeper tests | closures take `sold`; sweep passes sold ids |
| `internal/source/grare/source.go` | GR sweeper | `visit(1, lp.Cards, nil)`; doc comment |
| `internal/source/brandoff/source.go` | BRAND OFF sweeper | `visit(page, bags, nil)` |
| `internal/web/server_test.go` | Member API test | `MarkGone` call site takes 2 results |
| `cmd/kura/main.go` | Flags + wiring | `-recheck-after` → `Crawler.RecheckAfter` |
| `README.md`, `docs/provider-vetting.md` | Shipped docs | Sold-observation text, flag row, crawl limits |

### Dependencies
- **Requires:** sprint head `45c172c` (iteration-5 derive), with vet + tests green.
- **Blocks:** sprint finalize judging the ≥3-providers clause (sold observation part).

### Constraints
- Crawl invariants: identified KuraBot UA, robots.txt + Crawl-delay, no evasion. Unchanged: re-checks go through each source's existing `Fetcher`.
- **Request budget:** product-page requests per source per run stay ≤ `DetailBudget` (`-details`, default 150). Re-checks take only what first fetches leave. Per listing: ≤ 1 re-check per `RecheckAfter` (24 h), and only after it leaves the pages read. The README and the vetting doc state this.
- Each provider keeps its own `listing.Scheme`. Untouched.
- Source Hygiene (SKILL.md): no projex ids in code comments; symbols not line numbers; non-obvious choices name the rejected alternative.
- No commits to `master`; no push; no credentials; no network requests to shops.

### Assumptions (verify early in execution)
- A card on a list page is for sale. This holds for ALLU by filter (`stock > 0, sale_status 1`) and is the existing `RecordSeen` premise for GR / BRAND OFF. So a listing seen this crawl is not re-checked.
- `sale_status 2` with `stock 0` means sold. Evidence is one fixture row whose product page carries `soldButton`. Other non-1 statuses (unseen) are dropped as before and are **not** sold evidence.
- `HasDetail` stays unused outside `store` (`grep -rn HasDetail internal cmd` → `store.go` only). If execution finds a reader, keep `detail_at` in `MarkGone`'s product-page path and report it.
- Go 1.26 toolchain (`go version go1.26.0`).

### Impact Analysis
- **Direct:** store sold lifecycle, crawler detail phase, `crawl.Source` interface, 3 adapters.
- **Adjacent:**
  - `radar`: gone rows leave active search, count as comps for `compWindow`, and appear in `Gone` and member events. GR/ALLU rows are never `watched` (no complete sweep → no onboarding), so time-to-sell is unaffected.
  - `store.Version`: `gone_at` / `detail_at` changes already bump it.
  - `crawl_runs.gone` now also counts list-row sold marks.
- **Downstream:** none outside the repo. `crawl.Source` has no external implementers.

---

## Implementation

### Overview
Step 1: store primitives. Step 2: contract change + list-row signal (ALLU). Step 3: re-check cadence in the crawler + flag. Step 4: shipped docs. After each code step `go build ./... && go test -count=1 ./...` must pass.

### Step 1: Store — idempotent `MarkGone`, `RecheckDue` [PATCHED] → 2610140023-kura-sold-observation-patch.md

**Objective:** sold marks that do not drift, and a re-check selector.
**Confidence:** High
**Depends on:** None

**Files:** `internal/store/store.go`, `internal/store/store_test.go`, `internal/crawl/crawler.go` (call site only), `internal/web/server_test.go` (call site only)

**Changes:**

```go
// Before (store.go):
// MarkGone records that a product page no longer exists.
func (s *Store) MarkGone(ctx context.Context, id string, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE listings SET status='gone', gone_at=?, detail_at=? WHERE id=?`,
		now.Unix(), now.Unix(), id)
	return err
}

// After:
// MarkGone records that listing id is no longer for sale: its product page is
// gone or says sold, or a list page shows it sold. Only an active listing
// changes, and ok reports whether it did. Rejected: updating any row, which
// moves gone_at forward on every crawl that repeats the signal (a sold row
// stays in ALLU's newest pages for days). detail_at is not written: a list-row
// signal reads no product page, and stamping it would skip the first
// product-page read if the item is relisted.
func (s *Store) MarkGone(ctx context.Context, id string, now time.Time) (ok bool, err error) {
	res, err := s.db.ExecContext(ctx, `UPDATE listings SET status='gone', gone_at=? WHERE id=? AND status='active'`,
		now.Unix(), id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}
```

Add `RecheckDue` after `PendingDetails`. Move `PendingDetails`' row-scan loop into a shared `scanTodos(rows *sql.Rows) ([]DetailTodo, error)` (closes rows) used by both:

```go
// RecheckDue returns up to limit active listings of source in targets whose
// product page was last read before checkedBefore and that no list page has
// shown since seenBefore, least recently read first, so successive crawls
// rotate through them. Listings never read are PendingDetails' and are not
// returned.
func (s *Store) RecheckDue(ctx context.Context, source string, targets []string, seenBefore, checkedBefore time.Time, limit int) ([]DetailTodo, error) {
	if len(targets) == 0 || limit <= 0 {
		return nil, nil
	}
	args := []any{source, checkedBefore.Unix(), seenBefore.Unix()}
	for _, t := range targets {
		args = append(args, t)
	}
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, `SELECT id, url FROM listings
		WHERE source=? AND status='active' AND detail_at IS NOT NULL AND detail_at < ? AND last_seen < ?
		  AND target IN (`+strings.TrimSuffix(strings.Repeat("?,", len(targets)), ",")+`)
		ORDER BY detail_at, id LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	return scanTodos(rows)
}
```

`SaveDetail` stays as is: on a re-check the latest read wins, like `RecordSeen`'s "latest grade wins". (Error pages never reach it; see Out of Scope.)

Call sites:
- `crawler.go` `FetchDetails` `case isGone:` → `_, saveErr = c.Store.MarkGone(ctx, t.ID, c.now())`.
- `server_test.go` → `if _, err := st.MarkGone(ctx, "wm:cut", t0.Add(27*time.Hour)); err != nil {`.

Tests (`store_test.go`):
- `TestMarkGoneOnlyActive`:
  - `RecordSeen` a @day1; `MarkGone(a, day2)` → `true`; `MarkGone(a, day3)` → `false`; row `GoneAt == day2`.
  - `MarkGone("src:nosuch", day2)` → `false, nil`, and still one row.
  - `RecordSeen` a @day4 → active, and `PendingDetails` returns `src:a` (no `detail_at` written).
- `TestRecheckDue`, all `source "src"`; `day(n)` as in `TestPriceHistoryAndGoneLifecycle`. Rows:
  - a: HERMES, seen d1, `SaveDetail` d1
  - b: HERMES, seen d1, `SaveDetail` d2
  - c: HERMES, seen d1, no detail
  - d: HERMES, seen d1, `SaveDetail` d1, seen again d3
  - e: CHANEL, seen d1, `SaveDetail` d1
  - f: HERMES, seen d1, `SaveDetail` d1, `MarkGone` d2
  - g: HERMES, seen d1, `SaveDetail` d3

  Expect, with `seenBefore = checkedBefore = day(3)`:
  - `RecheckDue("src", ["HERMES"], d3, d3, 10)` → `[src:a, src:b]`
  - limit 1 → `[src:a]`
  - targets `["HERMES","CHANEL"]` → `[src:a, src:e, src:b]`
  - `nil` targets → empty

**Rationale:** idempotence is required by signal 2: ALLU's sold rows repeat across crawls. Selecting by `detail_at` gives oldest-first rotation with no new column. `last_seen < seenBefore` keeps requests off listings the list pages just showed for sale.

**Verification:** `go test -count=1 ./internal/store ./internal/web ./internal/crawl`. `grep -rn "MarkGone(" internal cmd` → every call takes 2 results.

**If this fails:** `git -C <work-root> checkout -- internal/store internal/crawl/crawler.go internal/web/server_test.go`.

---

### Step 2: `crawl.Source` sold ids + ALLU list-row signal [PATCHED] → 2610140023-kura-sold-observation-patch.md

**Objective:** sold rows in the pages read mark held listings gone.
**Confidence:** High
**Depends on:** Step 1
**Verify-Projex:** Encouraged (`run.Gone` accumulation order; interface change across 3 adapters)

**Files:** `internal/crawl/crawler.go`, `internal/crawl/crawler_test.go`, `internal/source/allu/parse.go`, `internal/source/allu/source.go`, `internal/source/allu/parse_test.go`, `internal/source/allu/source_test.go`, `internal/source/grare/source.go`, `internal/source/brandoff/source.go`

**Changes:**

`crawler.go` interface:
```go
// Before:
	// Sweep pages through target's catalogue, newest first, calling visit with
	// each page's bag listings. complete is true when the last page was reached
	// (as opposed to stopping at maxPages, 0 = unlimited).
	Sweep(ctx context.Context, target string, maxPages int, visit func(page int, items []listing.Listing) error) (complete bool, err error)

// After:
	// Sweep pages through target's catalogue, newest first, calling visit with
	// each page's bag listings and the source ids of listings the page shows as
	// sold (nil when the shop's list pages show none). Sold ids are never
	// stored as listings; they only mark gone a listing already held.
	// complete is true when the last page was reached (as opposed to stopping
	// at maxPages, 0 = unlimited).
	Sweep(ctx context.Context, target string, maxPages int, visit func(page int, items []listing.Listing, sold []string) error) (complete bool, err error)
```

`SweepTarget`: the closure takes `sold []string`. After the items loop:
```go
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
```
and the complete branch: `run.Gone = n` → `run.Gone += n`. Without this, a complete sweep's `SweepMissing` count would overwrite list-row marks.

`allu/parse.go`:
- `ListPage` gains `Sold []string // source ids of ALLU公式 rows shown sold (soldStatus)`.
- Add const next to the filter:

```go
// soldStatus is a list row's sale_status for a sold item. Evidence: the list
// fixture's one row with stock 0 has sale_status 2 (the other 29: stock 1,
// sale_status 1), and that item's product page (detail_sold.html, item
// 2117814) carries soldButton. Only stock 0 with this status is read as sold.
// Rejected: any status other than 1 as sold — one value seen, and a misread
// active item would be marked gone.
const soldStatus = 2
```
Loop, replacing the stock/status filter:
```go
		// Sold rows stay in ALLU's list; storing them would show sold bags as
		// on sale. Their ids go to Sold instead, which marks a listing Kura
		// already holds gone.
		if it.Stock <= 0 || it.SaleStatus != 1 {
			if it.Stock <= 0 && it.SaleStatus == soldStatus {
				p.Sold = append(p.Sold, strconv.FormatInt(it.ID, 10))
			}
			continue
		}
```
The seller check stays before it, so partner and user rows never enter `Sold`.

`allu/source.go`:
- `visit(page, lp.Cards)` → `visit(page, lp.Cards, lp.Sold)`; signature `visit func(int, []listing.Listing, []string) error`.
- In the `Sweep` doc, replace the sentences from "sold items surface only through Detail…" to "…within the first pages." with:
  > sold items surface through sold rows in the pages read (ListPage.Sold) and through Detail, which the crawler runs once per listing and then repeats for listings that left the pages read (crawl.Crawler.RecheckAfter); price changes are seen only while an item is within the first pages.

`grare/source.go`:
- `visit(1, lp.Cards, nil)`; signature updated.
- Doc: replace "Sold items surface only through Detail, and the crawler fetches one product page per listing (those without a detail yet), usually in the crawl that first sees it: an item that sells after that fetch stays active, and its price changes are seen only while it is on page 1." with:
  > Sold items surface only through Detail: the crawler reads each listing's product page once, usually in the crawl that first sees it, then re-reads it for listings that left page 1 (crawl.Crawler.RecheckAfter). List pages show no sold rows, so visit gets nil. Price changes are seen only while an item is on page 1.

`brandoff/source.go`: `visit(page, bags, nil)`; signature updated.

Tests:
- `crawler_test.go` `fakeSource`:
  - add `sold map[string][]string` (target → ids its first page shows sold);
  - add `capped map[string]bool` (target → `Sweep` returns `complete=false`);
  - `Sweep` passes `f.sold[target]` on page 1 and `nil` after, and returns `!f.capped[target], nil` at the end.
- New `TestListSoldRowsMarkGone`:
  - fa HERMES with `bags("fa","HERMES",2)`, budget 0, capped. RunAll @t0.
  - Then pages = only `HERMES-0`, `sold = {"HERMES": {"HERMES-1", "nosuch"}}`, clock t0+1h, RunAll.
    - `fa:HERMES-1` gone with `GoneAt == t0+1h`;
    - no row `fa:nosuch`;
    - latest run (`RecentRuns(ctx,1)[0]`) `Gone == 1`.
  - Third RunAll @t0+2h, same sold: `GoneAt` unchanged, latest run `Gone == 0`.
  - Clock: `now := t0; c.Now = func() time.Time { return now }`.
- `allu/parse_test.go`:
  - `TestParseList`: `reflect.DeepEqual(p.Sold, []string{"2117814"})`. Comment that it is `detail_sold.html`'s item.
  - `TestParseListFilters`: add `row(7, stock 0 + sale_status 3)` and `row(8, is_allu_partner + stock 0 + sale_status 2)`. Set `total_count` 8 and expect `Rows == 8`, cards only id 1, `Sold == ["4"]`.
- `allu/source_test.go`:
  - the three closures take a third `[]string` param;
  - `TestSweepComplete` serves `total_count 4, limit 2`, items `row(page,nil)` + `row(100+page, stock 0/sale_status 2)`, and asserts page p's `sold == [strconv.Itoa(100+p)]`, cards 1, pages `[1,2]`, complete.

**Rationale:** a third positional param is the smallest contract change. Rejected: a `Sold` flag on `listing.Listing`, which makes every sold row look like a listing to `RecordSeen` and its insert path; a page struct, which is more churn than one slice. Stricter `stock 0 && sale_status 2` keeps unseen statuses (reserved, hidden, …) from marking gone.

**Verification:** `go build ./... && go test -count=1 ./internal/crawl ./internal/source/... ./internal/store`. `grep -rn "visit func(" internal` → all 3 adapters + fake show 3 params.

**If this fails:** `git -C <work-root> checkout -- internal/crawl internal/source` (Step 1 files outside those dirs stay).

---

### Step 3: Product-page re-check cadence [PATCHED] → 2610140023-kura-sold-observation-patch.md

**Objective:** listings of never-complete targets get their product page re-read inside the existing budget, first fetches first.
**Confidence:** High
**Depends on:** Steps 1–2
**Verify-Projex:** Encouraged (budget split, cadence boundary, Unix-second comparisons)

**Files:** `internal/crawl/crawler.go`, `internal/crawl/crawler_test.go`, `cmd/kura/main.go`

**Changes:**

`Crawler` gains a field (after `DetailBudget`):
```go
	// RecheckAfter is the least time between two product-page reads of one
	// listing when re-checking (0 = no re-checks). Re-checks cover listings of
	// targets whose sweep did not complete in this crawl, which SweepMissing
	// cannot mark gone, that the crawl's list pages did not show, and spend
	// only what first fetches leave of DetailBudget, so a run makes no more
	// product-page requests than before. Rejected: a separate re-check budget
	// (more requests per run than -details states); re-checking every source
	// (complete sweeps already mark gone after GoneAfterMissedSweeps); a fixed
	// per-source "never completes" flag (GALLERY RARE brands with one page of
	// bags do complete, and a BRAND OFF sweep that is capped or fails does not).
	RecheckAfter time.Duration
```

`RunAll`:
```go
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
```
Update the `RunAll` doc: "…then spends each source's detail budget (first fetches, then re-checks of the targets whose sweep did not complete)."

`FetchDetails`:
```go
// FetchDetails spends src's detail budget: first on listings without
// product-page data, then, with what is left and when RecheckAfter is set, on
// re-checks (store.RecheckDue) of listings in recheck — targets whose sweep did
// not complete in this crawl — that no list page showed since crawlStart.
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
	// loop unchanged (MarkGone call already updated in Step 1)
	c.Log.Info("details done", "source", src.Name(), "first", first, "rechecks", len(todo)-first,
		"fetched", ok, "gone", gone, "failed", failed)
	return nil
}
```
The `DetailBudget` field doc gains one line: "Re-checks (RecheckAfter) come out of the same budget, after first fetches."

`cmd/kura/main.go`:
- `config`: `delay, every time.Duration` → `delay, every, recheckAfter time.Duration`.
- Flag, after `-details`:
```go
	fs.DurationVar(&c.recheckAfter, "recheck-after", 24*time.Hour, "least time between product-page re-checks of one listing, for brands whose sweep did not complete (0 = never; re-checks use what -details leaves)")
```
- `setup`: `&crawl.Crawler{…, DetailBudget: c.details, RecheckAfter: c.recheckAfter}`.

Default 24 h matches L1 Realtime's daily refresh. With `-crawl-every 6h`, a listing gets ≤ 1 re-check per day. When held listings outnumber the leftover budget, oldest-first rotation spaces each one further apart.

Tests (`crawler_test.go`):
- `fakeSource` gains `goneURL map[string]bool`. `Detail` returns `(listing.Detail{}, true, nil)` for those URLs.
- `TestRecheckIncompleteTargetsOnly`:
  - fa targets HERMES (capped, `bags` 2) + CHANEL (complete, `bags` 1). Budget 10, `RecheckAfter` 24h. RunAll @t0 → 3 first fetches.
  - Then HERMES pages = only `HERMES-0`, CHANEL pages = empty, `goneURL[https://fa/HERMES-1]`, reset `details`, clock t0+25h, RunAll.
    - `details == [https://fa/HERMES-1]`: `HERMES-0` was seen this crawl, and CHANEL completed (its missing item is `SweepMissing`'s);
    - `fa:HERMES-1` gone.
- `TestRecheckCadenceAndPriority`, fa HERMES capped, budget 2, `RecheckAfter` 24h:
  - @t0, `bags` 2 → 2 first fetches.
  - @t0+1h, no pages → `details` empty (cadence).
  - @t0+25h, pages = one new listing `NEW-0` → `details == [https://fa/NEW-0, https://fa/HERMES-0]` (first fetch first; oldest read, tie by id).
  - @t0+26h, no new → `details == [https://fa/HERMES-1]` (HERMES-0 read 1 h ago).
  - `c.RecheckAfter = 0` @t0+60h → `details` empty.
- `TestDetailBudgetPerSource` and `TestRunAllSweepsEverySource` stay green. Default `RecheckAfter` 0 means no re-checks there.

**Rationale:** "did not complete this crawl" is the gap `SweepMissing` leaves, measured not declared, so it adjusts by itself:
- GR brands ≤ 40 bags complete → not re-checked;
- BRAND OFF is re-checked only when capped or failed;
- if the owner later allows GR `?page=`, completed brands drop out of re-checks without code change.

`last_seen < crawlStart` compares Unix seconds. Listings seen this crawl have `last_seen ≥ floor(start)`, so they are excluded. Listings from earlier crawls (> 1 s earlier) are included.

**Verification:** `go test -count=1 -run 'Recheck|ListSold|DetailBudget|RunAll' -v ./internal/crawl`. Then `go run ./cmd/kura crawl -h` lists `-recheck-after` with default `24h0m0s` (no crawl runs, no network).

**If this fails:** `git -C <work-root> checkout -- internal/crawl/crawler.go internal/crawl/crawler_test.go cmd/kura/main.go`, then re-apply Step 2's `crawler.go` part.

---

### Step 4: Shipped docs [PATCHED] → 2610140023-kura-sold-observation-patch.md

**Objective:** README and vetting record match the behaviour and state the request budget.
**Confidence:** High
**Depends on:** Steps 2–3
**Do-Projex:** Encouraged

**Files:** `README.md`, `docs/provider-vetting.md`

**Changes:**

README § fetcher bullets (the `-details` bullet): append
> Re-checks of product pages already read (see *Knowing when something sold*) come out of the same cap, after first fetches.

README § Knowing when something sold: replace from "GALLERY RARE sweeps of a brand with more than 40 bags…" through "…(404, or the payload's sold button type) fetched once per listing." with:
```markdown
GALLERY RARE sweeps of a brand with more than 40 bags, and ALLU sweeps (at most 10 pages per
brand, newest first) of a brand with more ALLU公式 bags, are never complete, so sweeps never
mark their items sold. Two other signals do:
- **Product-page re-check.** Kura reads a listing's product page first in the crawl that
  first sees it, budget permitting. For a brand whose sweep did not complete, it re-reads the
  page of an active listing that the crawl's list pages no longer show, at most once per
  `-recheck-after` (24 h) per listing, least recently read first. A page that answers 404,
  shows GALLERY RARE's out-of-stock notice or carries ALLU's sold button type marks the item
  gone. Re-checks use only what first fetches leave of `-details`, so a crawl makes no more
  product-page requests than before; when listings outnumber that remainder, each is
  re-checked less often than daily. A BRAND OFF brand whose sweep was capped or failed gets
  the same re-checks.
- **Sold rows on ALLU list pages.** ALLU keeps sold items in its list (`stock 0`,
  `sale_status 2`). Such a row in the pages read marks the listing gone if Kura holds it; it
  is never stored as a new listing.

An item that sells after leaving the pages read is marked gone at its next re-check, not at
the moment of sale, and its price changes are seen only while it stays in the pages read.
```

README flags table, after the `-details` row:
```markdown
| `-recheck-after` | `24h` | Least time between product-page re-checks of one listing, for brands whose sweep didn't complete (0 = off); uses what `-details` leaves |
```

README § Tests, after the Store bullet:
```markdown
- **Crawler** covers per-source detail budgets, re-check priority and cadence, and sold rows
  on list pages.
```

README § Limitations, "Three sources" bullet: replace "GALLERY RARE and ALLU never complete a sweep of a large brand, so sold status and 最新上架 for their listings rely on one product-page fetch and crawl order." with
> GALLERY RARE and ALLU never complete a sweep of a large brand, so sold status for their listings comes from product-page re-checks (and ALLU's sold list rows) and can lag the sale by a re-check cycle, and 最新上架 relies on crawl order.

`docs/provider-vetting.md`:
- GALLERY RARE `Crawl limits` bullet (R2 re-read block): replace "Consequence: sweeps of a brand with more than 40 bags never complete, so sold status is seen only at the one product-page fetch per listing (404 or the shop's out-of-stock notice); an item sold after that fetch stays active in Kura." with
  > Consequence: sweeps of a brand with more than 40 bags never complete, so sold status comes from the product page (404 or the shop's out-of-stock notice). Kura reads it once, then re-reads it after the listing leaves page 1, at most once per 24 h per listing (`-recheck-after`), within the `-details` cap per crawl (no added requests).
- `## Outcome (2026-10-13, ALLU re-vet)`, ALLU bullet: after "Crawl limit: ≤10 list pages per brand per sweep, newest first." insert
  > Product pages: within `-details` per crawl; a listing's page is re-read at most once per 24 h, and only after it leaves the pages read. Sold rows in the pages read (`stock 0, sale_status 2`) mark held listings gone with no extra request.
- Leave both `— superseded` Outcome sections and the R1 第18条 (12) reasoning text unchanged. The superseded sections are historical; (12) is owner-held (F3), and the Outcome line above gives the owner the re-check volume.

**Rationale:** the crawl invariant needs every request budget stated and justified. The README is the shipped home for sold-detection behaviour.

**Verification:**
- `grep -n -i "once per listing\|Nothing re-checks\|one product-page fetch" README.md docs/provider-vetting.md internal` → only:
  - `docs/provider-vetting.md` superseded Outcome ("sold status seen only at the one product-page fetch per listing");
  - `internal/source/grare/source.go` sitemap rejection ("one product-page fetch per listing per crawl"), which is about sitemap enumeration and still true.
- `grep -n "recheck-after" README.md docs/provider-vetting.md` → flag row + both prose mentions.

**If this fails:** `git -C <work-root> checkout -- README.md docs/provider-vetting.md`.

---

## Verification Plan

### Automated Checks
- [x] `gofmt -l .` → empty
- [x] `go vet ./...` → clean
- [x] `go test -count=1 ./...` → all packages ok (cmd/kura, advisor, crawl, listing, radar, source/allu, source/brandoff, source/grare, store, web)
- [x] New tests present and passing: `TestMarkGoneOnlyActive`, `TestRecheckDue`, `TestListSoldRowsMarkGone`, `TestRecheckIncompleteTargetsOnly`, `TestRecheckCadenceAndPriority`; updated `allu` `TestParseList`, `TestParseListFilters`, `TestSweepComplete`

### Manual Verification
- [x] `go run ./cmd/kura crawl -h` shows `-recheck-after` (default `24h0m0s`) and makes no request
- [x] Source Hygiene on changed Go comments: no projex ids, no line numbers, rejected alternatives named (MarkGone, soldStatus, RecheckAfter)
- [x] No live crawl. Behaviour is covered by fakes + saved fixtures, so no shop request is needed

### Acceptance Criteria Validation
| Criterion | How to Verify | Expected Result |
|-----------|---------------|-----------------|
| Idempotent MarkGone | `TestMarkGoneOnlyActive` | 2nd call false, `GoneAt` = first, relisted row pending first fetch |
| RecheckDue selection | `TestRecheckDue` | `[a,b]`, `[a]`, `[a,e,b]`, empty |
| Contract change | `go build ./...`; grep `visit func(` | 3 params everywhere |
| List-row sold | `TestListSoldRowsMarkGone`; allu parse tests | held row gone once, unknown id not inserted, `Gone` counted; fixture `Sold == ["2117814"]` |
| Re-check scope + priority + cadence | `TestRecheckIncompleteTargetsOnly`, `TestRecheckCadenceAndPriority` | only unseen listings of incomplete targets; first fetches first; 24 h spacing; 0 = off |
| Flag | `kura crawl -h` | `-recheck-after 24h0m0s` |
| Docs | Step 4 greps | stale claims gone, budget stated |
| Green | vet + test | ok |

---

## Rollback Plan

1. `git -C <work-root> log --oneline 45c172c..HEAD` → identify this plan's commits.
2. `git -C <work-root> revert <sha>…` (newest first). Do not reset: the sprint branch is shared by later iterations.

---

## Notes

### Risks
- **Persistent per-listing failures hog the front of re-checks.** A failing read leaves `detail_at` unchanged, so it is retried first each crawl. This is bounded by the number of such listings and is within budget. Stop-on-403 behaviour is owner question F7 and unchanged: re-checks, like first fetches, continue after a 403 within the budget.
- **Re-check capacity < held listings** as ALLU/GR rows accumulate. Re-check spacing then grows past 24 h, and gone dates lag sales more. The README states this. A larger `-details` is the owner's lever.
- **`sale_status` semantics** rest on one fixture pair. The strict `stock 0 && 2` reading means a semantic change can only miss sales, never fake them.
- **Consignments (F1):** a consigned item that sells is marked gone like any ALLU公式 item. If the owner later excludes consignments, that exclusion path must not use `MarkGone` (sprint Open Questions already notes this).

### Owner-held decisions
None needed for this plan. Interactions with the open items:
- GALLERY RARE `?page=` permission would make some GR sweeps complete, which removes them from re-checks automatically;
- the 第18条 (12) reading (F3) gets the re-check volume in the vetting Outcome but is not re-decided.

### Split decision
`No split — single scope, within size budget`: 4 steps, one `.projex` scope, < 500 lines. Steps 1–3 are coupled through the `MarkGone` signature and the `crawl.Source` contract.

### Open Questions
- (none)
