# Patch: Kura Sold Observation for Never-Complete Sources

> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 5
> **Status:** Complete
> **Author:** Claude Opus 5.5 (patch, sprint iteration 5)
> **Parent:** 2610140010-kura-sold-observation-plan.md
> **Directive:** `Execute @2610140010-kura-sold-observation-plan.md` (patch step, uid `2610131822-kura-def-coverage#5`). No requests to any shop; owner-held decisions untouched.
> **Source Plan:** 2610140010-kura-sold-observation-plan.md (Steps 1–4, all)
> **Related Projex:** 2610140010-kura-sold-observation-plan.md | 2610131822-kura-def-coverage-sprint.md (not edited) | 2610132219-kura-grare-provider-2-audit.md (F1) | 2610132339-kura-allu-provider-3-audit.md (F4)
> **Result:** Success

---

## Summary

GALLERY RARE and ALLU listings now get marked sold even though their sweeps never complete. Two signals, both via an idempotent `store.MarkGone`:
1. **Product-page re-check** — `Crawler.RecheckAfter` (`-recheck-after`, default 24h): after first fetches, leftover `DetailBudget` re-reads active listings of targets whose sweep did not complete this crawl, not shown by this crawl's list pages, last read ≥ `RecheckAfter` ago, oldest read first (`store.RecheckDue`).
2. **ALLU sold list rows** — `crawl.Source.Sweep`'s `visit` gains `sold []string`; `allu.ParseList` fills `ListPage.Sold` with ALLU公式 non-partner rows `stock ≤ 0 && sale_status == 2`; `SweepTarget` marks held listings gone, counts them in `run.Gone` (`SweepMissing` count added, not overwriting).

Per-run product-page requests stay ≤ `DetailBudget` per source. Crawl invariants, `listing.Scheme`s, UI, radar unchanged. No network requests: fakes + saved fixtures only.

| Commit | Content |
|--------|---------|
| `9003350` | Steps 1–3: store, crawler, adapters, flag, tests |
| `90e04ea` | Step 4: README, `docs/provider-vetting.md` |

---

## Changes

### Step 1 — store primitives

**File:** `internal/store/store.go` — Modified
- `MarkGone` (L251–266): `(ok bool, err error)`; `UPDATE … WHERE id=? AND status='active'`; no `detail_at` write; `ok` ← `RowsAffected > 0`. Doc names rejected alternative (any-row update moves `gone_at` forward) and why `detail_at` is not stamped.
- `RecheckDue` (L200–222): active rows of source in targets, `detail_at IS NOT NULL AND detail_at < checkedBefore AND last_seen < seenBefore`, `ORDER BY detail_at, id LIMIT ?`; empty targets / `limit ≤ 0` → nil.
- `scanTodos` (L224–236): shared row scan (closes rows), used by `PendingDetails` + `RecheckDue`.

**File:** `internal/store/store_test.go` — Modified
- `TestMarkGoneOnlyActive` (L114): true → false on repeat, `GoneAt` = first; unknown id false/nil, no insert; relisted row is pending again.
- `TestRecheckDue` (L151): rows a–g as planned → `[a,b]`, `[a]`, `[a,e,b]`, nil.

**Files:** `internal/crawl/crawler.go` (`FetchDetails` gone branch), `internal/web/server_test.go` (L411), `internal/web/wanted_test.go` (L74), `internal/advisor/tools_test.go` (L73) — call sites take 2 results.

**Why:** ALLU sold rows repeat across crawls → mark must not drift. `detail_at` gives oldest-first rotation without a new column. `HasDetail` readers re-checked: `store.go` + `store_test.go` only → dropping the `detail_at` write has no UI effect.

### Step 2 — `crawl.Source` sold ids + ALLU list-row signal

**File:** `internal/crawl/crawler.go` — Modified
- `Source.Sweep` doc + signature (L18–24): `visit func(page int, items []listing.Listing, sold []string) error`; sold ids only mark held listings gone.
- `SweepTarget` (L109, L129–139): closure takes `sold`; each id → `MarkGone`, `run.Gone++` when marked, log `sold on list page`. Complete branch (L149): `run.Gone += n`.

**File:** `internal/source/allu/parse.go` — Modified
- `soldStatus = 2` (L26–32), evidence (fixture row 2117814 ↔ `detail_sold.html` `soldButton`) + rejected alternative (any non-1 status).
- `ListPage.Sold` (L74). Filter (L108–117): stock/status-dropped rows with `stock ≤ 0 && sale_status == soldStatus` → `Sold`. Seller check stays first → partner/user rows never enter.

**File:** `internal/source/allu/source.go` — Modified — `Sweep` doc (L66–75) states both sold signals; signature; `visit(page, lp.Cards, lp.Sold)` (L103).
**File:** `internal/source/grare/source.go` — Modified — `Sweep` doc (L54–59): re-read after leaving page 1, `visit` gets nil; `visit(1, lp.Cards, nil)` (L81).
**File:** `internal/source/brandoff/source.go` — Modified — signature; `visit(page, bags, nil)` (L88).

**Tests:**
- `internal/crawl/crawler_test.go`: `fakeSource` gains `sold` / `capped` / `goneURL` (L24–27); `Sweep` passes `sold[target]` on page 1, returns `!capped[target]`; `Detail` reports gone for `goneURL`. Helpers `row`, `latestRun` (L167, L182). `TestListSoldRowsMarkGone` (L191) as planned. `TestListSoldRowsAddToSweepMissing` (L238, added): one run marks one item by sold row and one by second missed complete sweep → `Gone == 2`.
- `internal/source/allu/parse_test.go`: `TestParseList` asserts `Sold == ["2117814"]` (`detail_sold.html`'s item); `TestParseListFilters` rows 7 (`stock 0, sale_status 3`) + 8 (partner, `stock 0, sale_status 2`), `total_count` 8 → `Rows 8`, card id 1 only, `Sold == ["4"]`.
- `internal/source/allu/source_test.go`: closures take `[]string`; `TestSweepComplete` serves `total_count 4, limit 2`, each page `row(p)` + sold `row(100+p)` → page p `sold == [100+p]`, 1 card, pages `[1,2]`, complete.

### Step 3 — re-check cadence + flag

**File:** `internal/crawl/crawler.go` — Modified
- `DetailBudget` doc +1 line (L46); `RecheckAfter time.Duration` (L48–58) with rejected alternatives (separate budget | every source | fixed per-source flag).
- `RunAll` (L69–95): `start := c.now()`; collects `incomplete[source]` from `!run.Complete` (incl. failed sweeps); `FetchDetails(ctx, src, incomplete[src.Name()], start)`.
- `FetchDetails` (L168–212): `PendingDetails(DetailBudget)` first; `left > 0 && RecheckAfter > 0` → `RecheckDue(src, recheck, crawlStart, now-RecheckAfter, left)` appended; log gains `first`, `rechecks`.

**File:** `cmd/kura/main.go` — Modified — `config.recheckAfter` (L42, gofmt realigned the struct); `-recheck-after` flag (L62, default `24h`); `setup` wires `RecheckAfter` (L155–156).

**File:** `internal/crawl/crawler_test.go` — `TestRecheckIncompleteTargetsOnly` (L263), `TestRecheckCadenceAndPriority` (L297) as planned.

### Step 4 — shipped docs

**File:** `README.md` — Modified
- Fetcher `-details` bullet: re-checks come out of the same cap, after first fetches.
- § Knowing when something sold: GR/ALLU paragraph replaced by the plan text (re-check + ALLU sold rows bullets, lag statement).
- Flags table: `-recheck-after` row. § Tests: Crawler bullet. § Limitations "Three sources": sold status from re-checks (+ ALLU sold rows), can lag by a re-check cycle; 最新上架 relies on crawl order.

**File:** `docs/provider-vetting.md` — Modified
- GALLERY RARE `Crawl limits` consequence (L211–215): read once, re-read after leaving page 1, ≤ 1 per 24 h per listing (`-recheck-after`), within `-details` (no added requests).
- `## Outcome (2026-10-13, ALLU re-vet)` ALLU bullet: product-page budget + re-read cadence; sold rows mark held listings gone with no extra request.
- Superseded Outcomes and R1 第18条 (12) text untouched.

---

## Verification

**Method:** `go vet ./...`, `go test -count=1 ./...`, `-v` on new/updated tests; gofmt on committed content of all 43 tracked `.go` files; mutation check (`run.Gone += n` → `=` makes `TestListSoldRowsAddToSweepMissing` fail, restored); `kura crawl -h` from a binary built into the session scratchpad (flag parse exits before DB open / network); Step 4 greps.

**Result:**
```
gofmt (HEAD content, 43 .go files) → no unformatted file
go vet ./...            → clean
go test -count=1 ./...
ok  	github.com/No3371/kura/cmd/kura
ok  	github.com/No3371/kura/internal/advisor
ok  	github.com/No3371/kura/internal/crawl
?   	github.com/No3371/kura/internal/fx	[no test files]
ok  	github.com/No3371/kura/internal/listing
ok  	github.com/No3371/kura/internal/radar
ok  	github.com/No3371/kura/internal/source/allu
ok  	github.com/No3371/kura/internal/source/brandoff
ok  	github.com/No3371/kura/internal/source/grare
ok  	github.com/No3371/kura/internal/store
ok  	github.com/No3371/kura/internal/web
-v PASS: TestMarkGoneOnlyActive | TestRecheckDue | TestListSoldRowsMarkGone | TestListSoldRowsAddToSweepMissing |
         TestRecheckIncompleteTargetsOnly | TestRecheckCadenceAndPriority | TestParseList | TestParseListFilters | TestSweepComplete
kura crawl -h → -recheck-after duration … (default 24h0m0s)
grep "once per listing|Nothing re-checks|one product-page fetch" README.md docs/provider-vetting.md internal
  → internal/source/allu/source.go (new text: "runs once per listing and then repeats …")
  → internal/source/grare/source.go (sitemap rejection, still true)
grep "recheck-after" → README.md prose + flag row; docs/provider-vetting.md GR crawl limits
```

**Status:** PASS

Success criteria (plan § Success Criteria): all 9 met — evidence per row above. UI copy checked (`internal/web/static/js/pages.js` methodology): no claim about one product-page read; no UI change, no browser check needed (none done).

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610140010-kura-sold-observation-plan.md | Source plan | Steps 1–4 `[PATCHED]`, criteria checked, Partial Execution line, Status `Complete`, moved to `closed/` |
| 2610131822-kura-def-coverage-sprint.md | Sprint | Not edited (orchestrator absorbs) |
| 2610132219-kura-grare-provider-2-audit.md, 2610132339-kura-allu-provider-3-audit.md | Findings addressed (GR F1, ALLU F4) | Not edited (sprint absorb records) |

---

## Notes

**Deviations from plan:**
- Two more `MarkGone` call sites than the plan listed: `internal/advisor/tools_test.go`, `internal/web/wanted_test.go` (call-site only).
- Added `TestListSoldRowsAddToSweepMissing`: the plan's tests did not exercise `run.Gone += n` (criterion "not overwritten by it"); mutation-checked.
- Step 4 grep: plan expected a match in a superseded vetting Outcome ("one product-page fetch per listing"); no such text exists there. Remaining matches are the new ALLU `Sweep` doc wording and the GR sitemap rejection.
- `gofmt -l .` in this Windows checkout lists ~25 untouched files: `core.autocrlf=true` checks them out CRLF, which gofmt flags. The index/committed content is LF and gofmt-clean. The one real hit (`cmd/kura/main.go` struct alignment from the new field) was fixed before commit.

**For audit:**
- Product-page gone marks no longer stamp `detail_at`. A gone listing that a list page shows again becomes active and pending its first fetch. If a GR / BRAND OFF list page ever kept showing a sold item, each crawl would refetch it (1 request, within budget), mark it gone again, and move `gone_at` forward. This breaks the plan's assumption that list-page cards are for sale; previously the same case left the item active forever.
- A re-check that keeps failing leaves `detail_at` as it was, so it stays first in line (plan Risks). Stop-on-403 (F7) unchanged.
- `sale_status` reading rests on one fixture pair. The strict `stock 0 && 2` rule means a drift can only miss sales, never mark an unsold item gone.

**Owner-held (unchanged, not decided here):** ALLU consignments keep/exclude (F1; an exclusion path must not use `MarkGone`) | 第18条 (12) reading (F3; vetting Outcome now states re-check volume) | stop-on-403 (F7) | GALLERY RARE `?page=` paging (a completed GR brand drops out of re-checks automatically).
