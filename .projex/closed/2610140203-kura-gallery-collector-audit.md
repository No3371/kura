# Audit: Kura Gallery Collector

> **Status:** Complete (Accept with Conditions)
> **Author:** Claude Opus 5.5 (auditor, orchestrator subagent, depth 2)
> **Parent:** 2610140136-kura-gallery-collector-log.md
> **Related Projex:** 2610140136-kura-gallery-collector-plan.md (audited plan, success criteria) | 2610140136-kura-gallery-collector-log.md (audited execution) | 2610130449-kura-mock-def.md (§ Is Display owner decision "Accept as demo") | 2610140010-kura-sold-observation-plan.md (re-check tiers reused) | 2610140037-kura-sold-observation-audit.md, 2610132339-kura-allu-provider-3-audit.md (precedents)
> **Audit Date:** 2026-10-14 | **Work Period:** 2026-10-14 01:46–02:10, range `master` (`6ffdcb5`)`..e96cc40` on `projex/2610140136-kura-gallery-collector` (worktree `.projexwt/2610140136-kura-gallery-collector`)
> **Subject:** `listing.Detail.Images` | `store.SaveDetail` / `replaceImages` / `SetImages` | `crawl.Crawler.FetchDetails` | `brandoff`, `grare`, `allu` `ParseDetail` + `galleryImages` / `apiAsset.images` | `README.md`
> **Method:** No request to any shop host; no browser check (web/radar/static unchanged, nothing new to render); repo `kura.db` not opened. Diff read in full. `go vet ./...` + `go test -count=1 ./...` in worktree. gofmt on committed content of all 14 touched `.go` files (`git show e96cc40:<f>`, CR stripped). `git grep "Get(ctx"` branch vs `master`. Tree of `e96cc40` exported (`git archive`) to session scratchpad: 16 mutation probes, plus one dump test of GALLERY RARE fixture output. Scratch removed; worktree left clean apart from this file.

---

## Audit Summary

**Claim:** each adapter's product-page parser returns the page's photo URLs in `listing.Detail.Images`; `store.SaveDetail` writes them into `listing_images` in the same tx as other product-page fields (non-empty replaces, empty keeps); zero additional shop requests; web/radar untouched.

**Verdict:** Verified — all 7 plan success criteria hold in code and tests; executor's claims reproduce. One deviation, correctly logged and justified.

**Assessment:** Completeness: High | Correctness: High | Quality: High | Value: Medium (existing DB listings mostly never gain a gallery without backfill — F1)

**Top Issues:**
1. **F1** Backfill gap wider than plan Out of Scope + new README sentence state: `RecheckDue` re-reads only listings the crawl's list pages did *not* show (`last_seen < crawlStart`), in incomplete targets. Already-read active listings still shown on list pages are never re-read on any source → never gain a gallery. Not an execution defect; input to the pending human backfill decision; README overpromises.
2. **F2** GALLERY RARE gallery test asserts count + first + last only; middle-reorder and duplicate mutants survive.
3. **F3** `schema.sql` `listing_images` comment: "Product photos beyond listings.image_url" now inaccurate; edited line rewrapped to 124 cols.

---

## Claims vs Evidence

| Claim (log / plan) | Evidence | Status | Notes |
|---|---|---|---|
| `Detail.Images` added, parsers fill 9 / 15 / 16 | `listing.go`; `TestParseDetail` ×3 pass; GR dump: `-01`…`-15` distinct, `size=l`, entity-decoded | ✓ | BRAND OFF + ALLU exact-slice asserts; GR weaker (F2) |
| NowPrinting excluded; ALLU sorted by `order`, blanks dropped | `TestParseDetailPlaceholderPhoto`, `TestParseDetailImageOrder`; mutants M4, M7, M8 killed | ✓ | |
| `SaveDetail` atomic; non-empty replaces, empty keeps | `store.SaveDetail` single tx, only `tx` used inside; `TestSaveDetailImages`; M1, M2 killed | ✓ | Atomicity by inspection; no fault-injection test (informational) |
| Crawl e2e: gallery on HERMES-0, nil on HERMES-1 | `TestDetailStoresGallery` pass; M11 killed | ✓ | |
| No new request | `git grep "Get(ctx"` 7 = 7 sites branch vs `master`; `FetchDetails` diff = counter + log key + doc only | ✓ | |
| web / radar / static untouched | `git diff --stat master..e96cc40`: 16 code/doc + 2 `.projex` files, none under `internal/web`, `internal/radar` | ✓ | |
| `go vet` clean, `go test -count=1 ./...` all ok | Re-run here: vet clean; 10 pkgs ok, `fx` no tests | ✓ | |
| gofmt applied | Committed content of 14 touched `.go`: `gofmt -l` empty | ✓ | Worktree CRLF makes a raw `gofmt -l` list everything — noise |
| Deviation: grare `!=` → `reflect.DeepEqual` moved Step 3 → Step 1 | `520c0ee` touches `grare/parse_test.go` 2 lines; `dedc358` adds Images asserts + `d.Images = nil` | ✓ | Needed: `Detail` stops being comparable at Step 1; every step commit builds |
| Two stale comments left | `schema.sql` first line; `radar.Item.Images` names only `store.SetImages` | ✓ | Both real (F3, F5) |
| `fetch.go` LF churn restored; worktree clean | `git status --short --ignored` empty before this audit | ✓ | |
| Base checkout + other worktree/branch untouched | Base `master` @ `6ffdcb5`, only the 2 human PNGs untracked; `home-shelf-hero` @ `4b4b8bd` | ✓ | |

---

## Objective Verification

### Human prompt 1 — clickable thumbnails when the DB has multiple crawled images
**Evidence:** UI from `4b4b8bd` (`detail.js` `photoHTML`: `it.images?.length ? it.images : it.photo_pending ? [] : [it.image_url]`); this branch fills `listing_images` on every product-page read.
- Actual: listings first read after deploy get full galleries (fixtures: 9 / 15 / 16 photos).
- Missing: listings already read keep the list photo — and almost all of them will keep it indefinitely (F1).
- Quality: High
**Verification:** ⚠ Partial — delivered for new reads; existing DB depends on a backfill decision.

### Human prompt 2 — service reads only what is stored; decoupled from the crawler
**Evidence:** collector writes through `store.SaveDetail` → `replaceImages`; radar reads `Row.Images`; zero lines changed under `internal/web`, `internal/radar`.
**Verification:** ✓ Verified

### Crawl invariants (UA, robots/Crawl-delay, no evasion, per-run budget)
**Evidence:** no fetcher, source `Get`, or budget code touched; galleries ride the already-budgeted product-page response body.
**Verification:** ✓ Verified. Viewer-side hotlink load rises from 1 to up to 9–16 full-size photos per detail view — not crawler traffic; covered by owner decision "Accept as demo".

### Provider `listing.Scheme` independence
**Evidence:** no `condition.go` / scheme code in diff.
**Verification:** ✓ Verified

---

## Code/Implementation Inspection

### `internal/store` — `SaveDetail`, `replaceImages`, `SetImages`
**Claimed:** tx + gallery replace; `SetImages` refactored onto shared `replaceImages`.
**Actual:** as claimed — Quality: High. `UPDATE` text unchanged; `replaceImages` uses only `tx` (safe under `SetMaxOpenConns(1)`); `SetImages` behaviour unchanged (`TestSetImages` incl. dedupe; M13 killed).
**Issues Found:**
- Non-empty `d.Images` on a nonexistent id now fails on the FK instead of a silent 0-row `UPDATE`; `FetchDetails` returns the save error → aborts that source's detail pass. Unreachable: ids come only from `PendingDetails` / `RecheckDue`, and no code deletes listings (`git grep "DELETE FROM listings"` empty) — Severity: Low (informational)
**Undocumented:** none.

### `internal/crawl/crawler.go` — `FetchDetails`
**Actual:** `galleries` counter + log key + 2-line doc. Budget/target logic byte-identical. Quality: High.
**Issues Found:** counter is untested (log-only) — Severity: Low (informational).

### `brandoff.ParseDetail` / `galleryImages`
**Actual:** `#js-main-slider .swiper-slide img` src, skip blank + `/NowPrinting`, `base.Parse`; `source.go` passes `resp.URL`. Exact 9-URL assertion proves malformed first `<img>` neither drops nor duplicates a slide and the zoom-button svg is excluded (M3 killed). Placeholder rule mirrors `radar` `photoPending` (`/NowPrinting` substring). Quality: High.
**Issues Found:** a BRAND OFF listing whose first read shows only the placeholder stores no gallery and, if its brand's sweep completes, is never re-read — keeps the list photo forever even after studio shots appear. Subset of F1 — Severity: Low.

### `grare.ParseDetail` / `galleryImages`
**Actual:** `figure.fs-c-productCarouselMainImage__image img` src (`size=l`), `base.Parse`; sold early return untouched. Output on fixture verified by dump: 15 distinct, `-01`…`-15` in order. Quality: High; test Medium (F2).

### `allu.ParseDetail` / `apiAsset.images`
**Actual:** `app_item_images[]{url, order}` decoded into `apiAsset`; clone + `SortStableFunc` by `order`; blank-after-trim dropped. Signature unchanged. Quality: High.
**Issues Found:** URLs stored as given, not resolved against a base, while `listing.Detail.Images` doc promises absolute URLs. Consistent with ALLU's list photo (`ParseList` stores `it.Image.URL` as given); fixture URLs absolute — Severity: Low (informational).

**Metrics (whole diff):** Readability/Maintainability: High | Test Coverage: every new branch has a test except the log counter | Technical Debt: Low (F3, F5 stale comments)

---

## Testing Validation

**Coverage:** store keep/replace/first-write ✓ | crawl e2e ✓ | each parser on fixture ✓ | placeholder + order + blank ✓ | `allu` `TestDetail` through `httptest` ✓. Edge cases missing: GR middle order/duplicates (F2); tx atomicity under a mid-tx failure (inspection only).
**Execution:** all pass, `-count=1`; no flakes observed.
**Mutation probes (16 applied):** 13 killed | 1 equivalent | 2 survived.
- Killed: clear-on-empty (M1) | `SaveDetail` ignores Images (M2) | BRAND OFF `#js-main-slider img` (M3) | keep NowPrinting (M4) | no URL resolution (M5) | GR reversed order (M6) | ALLU no sort (M7) | ALLU keep blanks (M8) | ALLU drop `Images` (M9) | BRAND OFF `nil` base (M10) | fake source drops Images (M11) | GR `Images: nil` (M12) | `replaceImages` no dedupe (M13).
- Equivalent: BRAND OFF page-wide `img.productimg` (M3b) — class sits only on the 9 slider imgs, same output.
- **Survived:** GR swap photos 2↔3 (M6b) | GR duplicate photo 2 over photo 3 (M6c) → F2.
**Quality Issues:** GR assertion follows the plan's own spec; executor's "15 distinct" cross-check was a grep, not encoded in the test.

---

## Documentation Audit

**Completeness:** README — `-details` bullet, flags table, Limitations bullet: Complete. Migration: none needed (schema unchanged). Rollback: plan § Rollback accurate.
**Accuracy:** Partial — new sentence "Listings whose page was read before galleries were collected keep the list photo until a re-check reads it again" implies eventual convergence; re-checks only reach listings the list pages stopped showing (README § Knowing when something sold says so, the new sentence does not) (F1).
**Quality:** Clarity High | Completeness High | Usability High

---

## Source Hygiene

**Scope:** Diff only (`master..e96cc40`, `.projex/**` excluded)

| Rule | Location (symbol) | Quote | Severity |
|---|---|---|---|
| 3 (describes what is not) | `schema.sql` → `listing_images` comment | "Product photos beyond listings.image_url (the list-page photo)" | Low |

Rules 1, 2, 4, 5: grep of added lines for projex refs / `file:ln` / `:NNN` / `Step N` / `====` / `----` / "deliberate" / "by design" → none.

**Rule 6 gaps** (non-obvious change, no rationale comment): none load-bearing.
- `allu` `apiAsset.images` — no base resolution unlike brandoff/grare — asymmetry follows ALLU `ParseList`; a reader of `listing.Detail.Images` ("absolute URLs") could expect resolution. Low, optional.

**Rule 6 vacuity:** none. `SaveDetail` names two rejected alternatives + consequences; `replaceImages` names the single-connection constraint; both `galleryImages` and `apiAsset.images` name rejected alternatives with consequences.

**Rationale promotion:** walkthrough not yet written (branch open). README carries the shipped claims (no extra request, URLs only, hotlink). Keep-on-empty and tx rationale live in `store.SaveDetail` comment. Backfill rationale exists only in the plan — it is a pending decision, not settled rationale → none needed now; walkthrough should state "none needed — supported by diff" or promote the backfill outcome once decided.

**Commit composition** (informational — fix forward, never rewrite history): 6 commits checked — 5 step commits `projex: step N - …` (ephemeral form allowed) all carry `Projex: 2610140136-kura-gallery-collector` [5/5]; `e96cc40` doc-only (exempt). Landing commit not yet made.
**Violations:** None.

**Trailer survival on `master`:** 1/14 (master holds 14 commits; only `8950a55` carries a trailer; `4b4b8bd` gallery UI commit carries none) — not zero; informational, consistent with prior audits.

---

## Gap Analysis

### Promised But Not Delivered
| Promise | Status | Impact |
|---|---|---|
| (none against the plan) | — | — |
| Human prompt 1 on the existing demo DB | Partial (scoped out: backfill) | Medium — see F1 |

### Undocumented Issues
| Issue | Severity | Affects |
|---|---|---|
| Re-check path rarely backfills: only listings no longer shown on list pages, in incomplete targets; those skew toward sold/delisted (→ `MarkGone`, no gallery) | Medium | Human backfill decision; README accuracy |
| BRAND OFF listing first read with NowPrinting placeholder never gains a gallery in a complete-sweep brand | Low | Detail page of such listings |

### Unhandled Edge Cases
- Shop reshoots a listing still on list pages → gallery stale forever (never re-read); plan Risks says "re-checks refresh it", true only for listings off the list pages. Impact: Low for a demo.
- Mid-tx failure in `SaveDetail` — rolls back by `defer tx.Rollback()`; untested. Impact: none observed.

---

## Quality Assessment

### Completeness: High
**Strengths:** every success criterion met; tests per parser, store, crawl e2e; README updated in the three right places.
**Gaps:** backfill (scoped out); GR test precision.

### Correctness: High
**Works:** galleries parse ✓ | atomic save ✓ | keep-on-empty ✓ | zero new requests ✓ | service untouched ✓
**Bugs:** none found.

### Code Quality: High
**Positive:** smallest viable change — gallery folded into the one existing write per product-page read; `replaceImages` keeps one dedupe/position rule; rationale comments are specific.
**Concerns:** two stale comments (F3, F5).
**Tech Debt:** Low.

### Value Delivered: Medium
**Intended:** real galleries on detail pages.
**Actual:** real galleries for listings first read after deploy; the existing catalogue keeps single photos unless a backfill is approved.
**Impact:** User: Positive (new listings) | Business: demo shows real multi-photo galleries once a fresh crawl reads new listings.

---

## Open Findings

### Undocumented Discoveries
- Changes: none beyond the plan; `gofmt` realigned `apiAsset` struct tags (whitespace only, logged).
- Problems: F1 — Evidence: `store.RecheckDue` `WHERE … AND last_seen < ?` with `seenBefore = crawlStart`.
- Workarounds: none.

### Impact Analysis
- Downstream: `/api/…` `images` populated for newly read listings → thumbnails appear; `Version` moves on each re-check gallery rewrite (already moved by `detail_at`).
- Future: Enabled: any backfill tier only needs to call `src.Detail` + `SaveDetail` (gallery comes free). | Blocked: nothing.
- Risks: markup drift → `galleries` drops in `details done` log, stored galleries kept — Likelihood: Medium over time, impact Low.

### Improvements
- Could be better: GR exact-slice assertion (as BRAND OFF).
- Would make excellent: a `galleries` assertion in a crawl test (log capture) — the counter is the only drift signal.
- Missed: none material.

---

## Stakeholder Impact

| Stakeholder | Promised | Reality | Impact |
|---|---|---|---|
| Human / owner | Gallery from crawled images, service decoupled | Delivered for new reads; existing listings need a backfill decision | Positive |
| Shops (crawler traffic) | No additional requests | Verified: 0 new request sites | Neutral |
| Shops (viewer hotlinks) | Covered by "Accept as demo" | Up to 9–16 photos per detail view vs 1 | Neutral (accepted) |
| Recruiter viewing demo | Multi-photo detail pages | Only on listings read after deploy | Positive once a fresh crawl runs |

---

## Findings

### Critical (Must Address)
- None.

### Significant (Should Address)
- **F1 Backfill gap is wider than stated** — `RecheckDue` re-reads only active listings in incomplete targets that the crawl's list pages did *not* show since `crawlStart`. Already-read listings still shown on list pages are never re-read on any source, so they never gain a gallery; the listings re-checks do reach skew toward sold/delisted ones (gone → `MarkGone`, no gallery). Plan Out of Scope ("gain a gallery only when a re-check reads their page again (targets whose sweep did not complete …)") and the new README sentence omit the not-shown condition, understating the gap that the pending human decision is about. Not an execution defect. → (a) Reword the README sentence, e.g. "Listings already read keep the list photo; Kura re-reads a product page only for the re-checks under *Knowing when something sold*." (b) Present the backfill question to the human with the accurate scope: without it, galleries appear essentially only on listings first read after deploy.

### Minor (Nice to Fix)
- **F2 GR gallery test precision** — asserts len 15 + first + last; mutants swapping or duplicating middle photos pass → assert the exact 15-URL slice (build `-01`…`-15` like the BRAND OFF test).
- **F3 `schema.sql` `listing_images` comment** — "Product photos beyond listings.image_url" no longer true (collected galleries lead with the main photo and replace the list photo on the detail page); the edited line runs 124 cols vs ~75 around it → reword ("A listing's product-page photo gallery, in the shop's order …") and rewrap.
- **F4 BRAND OFF placeholder-only first read** — never re-read in a complete-sweep brand → keeps list photo even after studio shots appear. Fold into F1's backfill decision.
- **F5 `radar.Item.Images` comment** names only `store.SetImages` — incomplete, still true. Radar untouched by human requirement → fix only if a comment-only radar edit is allowed.
- **F6 (informational)** `galleries` log counter untested | `SaveDetail` FK error on nonexistent id (unreachable) | ALLU URLs not base-resolved (consistent with its list photo).

### Positive
- Deviation handled right: caught by `go vet` at Step 1, moved forward so every step commit builds, logged with the compiler message.
- Zero-request claim provable and proven (call-site parity + diff of `FetchDetails`).
- Rationale comments name real rejected alternatives (clear-on-empty, crawler-side `SetImages`, page-wide path match, xl/xs variants, payload order) — no padding.
- Strong tests: 13/16 mutants killed, 1 equivalent.

---

## Recommendations

**Immediate:** F1 (a) README wording before close (doc-only); F2 + F3 are cheap and may ride the same fix-forward if the orchestrator chooses.
**Future:** F1 (b) human decision on a backfill tier (request-volume trade-off: up to `-details` per source per crawl until drained).
**Process:** when a plan scopes a gap out "pending human decision", quote the gating query (`RecheckDue` conditions) in the Out of Scope note so the decision is framed on the real mechanism.

---

## Final Verdict

**Status:** Accept with Conditions

**Overall Assessment:**
- Completeness: High
- Correctness: High
- Quality: High
- Value: Medium

**Conditions:**
- [ ] Landing commit (close): conventional type subject (e.g. `feat(crawl): …`) + `Projex: 2610140136-kura-gallery-collector` trailer.
- [ ] F1 (a): README sentence corrected, or orchestrator records the decision to ship as is.
- [ ] F1 (b): backfill question raised to the human with the accurate scope (orchestrator).

**Sign-off:** Yes — plan delivered exactly, verified independently, no regressions, no new crawler requests; remaining items are wording, test precision, and a human decision.
