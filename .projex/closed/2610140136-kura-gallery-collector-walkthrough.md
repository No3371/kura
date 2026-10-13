# Walkthrough: Kura Gallery Collector — product-page photos into the stored gallery

> **Execution Date:** 2026-10-14
> **Completed By:** Claude Opus 5.5 (executor, auditor, patch, closer — orchestrator subagents)
> **Source Plan:** 2610140136-kura-gallery-collector-plan.md
> **Parent:** 2610140136-kura-gallery-collector-plan.md
> **Log:** 2610140136-kura-gallery-collector-log.md
> **Duration:** execution 01:46–02:10 (~24 min); audit 02:03; fix-forward patch 02:09; close same day
> **Result:** Success (plan criteria 7/7) — value Partial on the existing DB: backfill of already-read listings open, human decision

---

## Summary

Each adapter's product-page parser now returns the page's photo URLs in `listing.Detail.Images` (BRAND OFF 9 / GALLERY RARE 15 / ALLU 16 on fixtures), and `store.SaveDetail` writes them into `listing_images` in the same transaction as the other product-page fields (non-empty replaces, empty keeps). Galleries ride the product-page read already inside each source's `-details` budget → zero new shop requests; `internal/web`, `internal/radar`, static assets untouched. Audit: Verified — Accept with Conditions; fix-forward patch closed F1(a)/F2/F3. Listings already read before this change mostly never gain a gallery (re-checks reach only listings off the list pages in incomplete targets) — backfill = open human decision.

**Human prompts served:** (1) "I want a gallery. Like clickable thumbnails below the main image, when the db have multiple images crawled" — UI in `4b4b8bd`; this branch fills the DB. (2) "The point is our service only cares about images crawled. Decoupled from the crawler." — collector writes only through `store`; service reads `Row.Images`; zero service-side lines changed. (3) "Start a opus as the orchestrator to do it" — orchestrated chain plan → execute → audit → patch → close.

---

## Objectives Completion

| Objective | Status | Notes |
|---|---|---|
| Parsers fill `Detail.Images` (9 / 15 / 16, exact order) | Complete | Exact-slice asserts on all three (GR tightened by patch F2) |
| NowPrinting excluded; ALLU sorted by `order`, blanks dropped | Complete | `TestParseDetailPlaceholderPhoto`, `TestParseDetailImageOrder` |
| `SaveDetail` atomic gallery replace; keep-on-empty | Complete | `TestSaveDetailImages` p1,p2 → kept → q1 |
| Crawl e2e gallery stored | Complete | `TestDetailStoresGallery` |
| No new request | Complete | `Get(ctx` call sites 7 = 7 vs `master` |
| Service side untouched | Complete | 0 files under `internal/web`, `internal/radar` |
| Suite green | Complete | `go vet` clean; `go test -count=1 ./...` all ok |
| Real galleries across the existing catalogue (prompt 1, demo DB) | Partial | Out of Scope: backfill → human decision (F1(b), F4) |

---

## Execution Detail

> Derived from git history (`master..projex/2610140136-kura-gallery-collector`) and the execution log. Deviations called out.

### Step 1: `Detail.Images` persisted by `SaveDetail`; crawler log + test — `520c0ee`

**Planned:** `Detail.Images`; `SaveDetail` → tx + `replaceImages` when non-empty; `SetImages` body → `replaceImages`; schema comment; `TestSaveDetailImages`; `galleries` log counter; `fakeSource.images` + `TestDetailStoresGallery`.
**Actual:** as planned. Plus `grare/parse_test.go` `d != want` → `!reflect.DeepEqual(d, want)` (moved forward from Step 3).
**Deviation:** grare comparison swap pulled from Step 3 — `go vet` failed: `internal\source\grare\parse_test.go:173:5: invalid operation: d != want (struct containing []string cannot be compared)`. Swapped here so every step commit builds.

**Files Changed (ACTUAL):**
| File | Change Type | Planned? | Details |
|---|---|---|---|
| `internal/listing/listing.go` | Modified | Yes | L115–119: `Images []string` + doc comment (absolute URLs, shop order, placeholders out, nil when none) |
| `internal/store/store.go` | Modified | Yes | L242–270: `SaveDetail` doc (2 rejected alternatives) + tx, `replaceImages` when `len(d.Images) > 0`; L378–386 `SetImages` → thin tx wrapper; L403 `replaceImages(ctx, tx, id, urls)` (single-connection note) |
| `internal/store/schema.sql` | Modified | Yes | `listing_images` comment names `SaveDetail` as writer (reworded again by patch F3) |
| `internal/store/store_test.go` | Modified | Yes | L326–369: `TestSaveDetailImages` |
| `internal/crawl/crawler.go` | Modified | Yes | L176–177 doc sentence; L194 `galleries` counter; L209–211 increment; L219 log key |
| `internal/crawl/crawler_test.go` | Modified | Yes | L28 `images` map; L56 fake `Detail` returns them; L167–182 `TestDetailStoresGallery` |
| `internal/source/grare/parse_test.go` | Modified | No (Step 3) | 2 lines: comparison swap — see Deviation |

**Verification:** `go vet ./...` clean after swap; `go test -count=1 ./...` all ok; `TestSetImages`, `TestSaveDetailImages`, `TestSaveDetailOverridesModelCode`, `TestDetailBudgetPerSource`, `TestDetailStoresGallery` PASS. Tx scope by inspection: only `tx` used inside `SaveDetail` (safe under `SetMaxOpenConns(1)`).
**Issues:** compile failure above → resolved by deviation.

---

### Step 2: BRAND OFF gallery parse — `ae9e23a`

**Planned:** `ParseDetail(r, base)` + `galleryImages` (`#js-main-slider .swiper-slide img`, skip blank / `/NowPrinting`, `base.Parse`); `source.go` passes `resp.URL`; tests.
**Actual:** as planned. Plan fallback (`img.productimg`) not needed — exact 9-URL slice proves malformed first `<img>` neither drops nor duplicates a slide, zoom-button svg excluded.
**Deviation:** None.

| File | Change Type | Planned? | Details |
|---|---|---|---|
| `internal/source/brandoff/parse.go` | Modified | Yes | L202–204 doc + signature; L234 `Images:`; L238–258 `galleryImages` (rejected: page-wide path match; keeping NowPrinting) |
| `internal/source/brandoff/source.go` | Modified | Yes | L121 `ParseDetail(bytes.NewReader(resp.Body), resp.URL)` |
| `internal/source/brandoff/parse_test.go` | Modified | Yes | imports `fmt`, `slices`, `strings`; L150 `testBase`; exact 9-URL assert; L163–182 `TestParseDetailPlaceholderPhoto` |

**Verification:** `go test -count=1 -v ./internal/source/brandoff/` all PASS. Only caller of `brandoff.ParseDetail` = `source.go`.

---

### Step 3: GALLERY RARE gallery parse — `dedc358`

**Planned:** `ParseDetail(r, base)` + `galleryImages` (`figure.fs-c-productCarouselMainImage__image img`, `size=l`); test call sites + Images asserts (len, first, last) + `d.Images = nil` before `reflect.DeepEqual`.
**Actual:** as planned (comparison swap already landed in Step 1).
**Deviation:** comparison swap moved to Step 1 (above). Test precision later raised by patch F2 (exact 15-URL slice).

| File | Change Type | Planned? | Details |
|---|---|---|---|
| `internal/source/grare/parse.go` | Modified | Yes | L196–198 doc + signature; L234 `Images:`; L238–257 `galleryImages` (rejected: `size=xl` enlarged, `size=xs` sub) |
| `internal/source/grare/source.go` | Modified | Yes | L97 passes `resp.URL` |
| `internal/source/grare/parse_test.go` | Modified | Yes | 3 call sites `(…, testBase)`; Images asserts |

**Verification:** `go test -count=1 -v ./internal/source/grare/` all PASS. Fixture encodes `=` as `&#x3D;` → exact URL proves entity decoding.

---

### Step 4: ALLU gallery parse — `dcd13a1`

**Planned:** `apiAsset.Images []apiImage` (`app_item_images`), `apiImage{URL, Order}`, `images()` stable sort by `order`, blanks dropped; tests.
**Actual:** as planned. `order` decodes as int (plan's `float64` fallback unused). `gofmt` realigned `apiAsset` tags (whitespace only, 7 fields).
**Deviation:** None.

| File | Change Type | Planned? | Details |
|---|---|---|---|
| `internal/source/allu/parse.go` | Modified | Yes | imports `cmp`, `slices`; L211–241 `Images` field, `apiImage`, `images()` (rejected: payload order alone); L276 `Images: a.images()` |
| `internal/source/allu/parse_test.go` | Modified | Yes | `want.Images` 16 URLs; L285–304 `TestParseDetailImageOrder` |
| `internal/source/allu/source_test.go` | Modified | Yes | L168 `/p/bag`: `len(d.Images) != 16` |

**Verification:** `go test -count=1 -v ./internal/source/allu/` all PASS incl. `TestDetail` over `httptest`.

---

### Step 5: README — `a107eab`

**Planned:** 3 edits — `-details` bullet sentence | flags table `photo gallery` | Limitations hotlink wording.
**Actual:** as planned; Limitations bullet rewrapped (wording identical).
**Deviation:** None here; sentence 2 of the bullet later corrected by patch F1(a).

| File | Change Type | Planned? | Details |
|---|---|---|---|
| `README.md` | Modified | Yes | L86–90 `-details` bullet; L243 flags row; L306–308 Limitations |

---

### Post-plan: audit + fix-forward patch

- `1c3f278` audit 2610140203-kura-gallery-collector-audit.md — Verified, Accept with Conditions; 16 mutation probes: 13 killed, 1 equivalent, 2 survived (GR middle swap/duplicate → F2).
- `b16a1a1` patch 2610140209-kura-gallery-collector-audit-fixes-patch.md: **F1(a)** README sentence now states the re-read gap ("… so most such listings never gain a gallery") | **F2** GR exact 15-URL `slices.Equal` (+ `fmt` import) | **F3** `listing_images` comment reworded (gallery replaces list photo on detail page) + rewrapped ≤78 cols. `ea734ca` patch doc.
- No user (human) intervention during execution; orchestrator acted as projex user.

---

## Complete Change Log

> **Derived from:** `git diff --stat master..projex/2610140136-kura-gallery-collector` (code/doc paths; `.projex/` bookkeeping excluded). 16 files, +271 / −37.

### Files Created
| File | Purpose | Lines | In Plan? |
|---|---|---|---|
| — | none (code) | — | — |

### Files Modified
| File | Changes | Lines Affected (+/−) | In Plan? |
|---|---|---|---|
| `README.md` | gallery from budgeted read; re-read gap; flags; hotlink | +9 / −5 | Yes (+ patch F1a) |
| `internal/crawl/crawler.go` | doc + `galleries` counter/log | +7 / −2 | Yes |
| `internal/crawl/crawler_test.go` | fake images + e2e test | +18 / −1 | Yes |
| `internal/listing/listing.go` | `Detail.Images` | +5 / −1 | Yes |
| `internal/source/allu/parse.go` | `app_item_images` → `images()` | +34 / −7 | Yes |
| `internal/source/allu/parse_test.go` | want Images + order test | +26 / −0 | Yes |
| `internal/source/allu/source_test.go` | 16-image check | +1 / −1 | Yes |
| `internal/source/brandoff/parse.go` | `ParseDetail(r, base)` + `galleryImages` | +25 / −2 | Yes |
| `internal/source/brandoff/parse_test.go` | 9-URL assert + placeholder test | +24 / −1 | Yes |
| `internal/source/brandoff/source.go` | pass `resp.URL` | +1 / −1 | Yes |
| `internal/source/grare/parse.go` | `ParseDetail(r, base)` + `galleryImages` | +24 / −2 | Yes |
| `internal/source/grare/parse_test.go` | DeepEqual + exact 15-URL assert | +13 / −4 | Yes (+ patch F2) |
| `internal/source/grare/source.go` | pass `resp.URL` | +1 / −1 | Yes |
| `internal/store/schema.sql` | `listing_images` comment | +6 / −4 | Yes (+ patch F3) |
| `internal/store/store.go` | tx `SaveDetail`, `replaceImages` | +33 / −5 | Yes |
| `internal/store/store_test.go` | `TestSaveDetailImages` | +44 / −0 | Yes |

### Files Deleted
| File | Reason | In Plan? |
|---|---|---|
| — | none | — |

### Planned But Not Changed
| File | Planned Change | Why Not Done |
|---|---|---|
| — | all 16 Key Files changed | — |

---

## Success Criteria Verification

### C1: `Detail.Images`; parsers fill 9 / 15 / 16 exact
**Method:** `TestParseDetail` ×3 (exact-slice). **Evidence:** all PASS; GR exact slice since patch F2. **Result:** PASS

### C2: NowPrinting never enters; ALLU sorted by `order`, blanks dropped
**Method:** `TestParseDetailPlaceholderPhoto`, `TestParseDetailImageOrder`; audit mutants M4/M7/M8 killed. **Result:** PASS

### C3: `SaveDetail` atomic replace; empty keeps
**Method:** `TestSaveDetailImages`; tx scope by inspection; mutants M1/M2 killed. **Result:** PASS (atomicity under mid-tx failure: inspection only, `defer tx.Rollback()`)

### C4: Crawl e2e
**Method:** `TestDetailStoresGallery` (HERMES-0 gallery, HERMES-1 nil). **Result:** PASS

### C5: No new request; budget logic untouched
**Method:** `git grep -n "Get(ctx" -- internal/source internal/crawl` on branch vs `master`.
**Evidence:**
```
branch: 7   master: 7
```
`FetchDetails` diff = doc + counter + log key only. **Result:** PASS

### C6: web / radar / static unchanged
**Method:** `git diff --stat master..projex/2610140136-kura-gallery-collector -- internal/web internal/radar` → empty. **Result:** PASS

### C7: `go vet` + `go test` pass
**Method:** re-run at close in worktree (branch tip `ea734ca`).
**Evidence:**
```
go vet ./...  → clean
ok  cmd/kura | internal/advisor | internal/crawl | internal/listing | internal/radar
ok  internal/source/allu | internal/source/brandoff | internal/source/grare | internal/store | internal/web
?   internal/fx [no test files]
```
**Result:** PASS

### Acceptance Criteria Summary

| Criterion | Method | Result | Evidence |
|---|---|---|---|
| C1 parsers fill Images | `TestParseDetail` ×3 | Pass | log Steps 2–4; patch F2 |
| C2 placeholder + order | 2 dedicated tests | Pass | log Steps 2, 4; audit mutants |
| C3 atomic, keep-on-empty | `TestSaveDetailImages` | Pass | log Step 1 |
| C4 e2e crawl | `TestDetailStoresGallery` | Pass | log Step 1 |
| C5 no new request | call-site parity | Pass | 7 = 7 |
| C6 service untouched | diff --stat | Pass | 0 files |
| C7 suite green | vet + test | Pass | close re-run |

**Overall:** 7/7 criteria passed.

---

## Deviations from Plan

### D1: grare test comparison swap moved Step 3 → Step 1
- **Planned:** Step 3 swaps `d != want` → `reflect.DeepEqual`.
- **Actual:** done in Step 1 (`520c0ee`, 2 lines).
- **Reason:** `Detail.Images` slice makes `listing.Detail` non-comparable → Step 1 would not compile tests.
- **Impact:** none on outcome; every step commit builds.
- **Recommendation:** plans adding a slice/map field to a struct should grep `!=`/`==` comparisons of it in the same step.

### D2: Post-plan fix-forward (patch 2610140209-kura-gallery-collector-audit-fixes-patch.md)
- **Planned:** README sentence "keep the list photo until a re-check reads it again"; GR test len + endpoints; schema comment edit only naming `SaveDetail`.
- **Actual:** README states re-check gap; GR exact slice; `listing_images` comment rewritten to describe what is.
- **Reason:** audit F1(a), F2, F3.
- **Impact:** docs accurate on backfill gap; GR order/duplicate mutants now killed.

---

## Issues Encountered

### I1: `go vet` compile failure after Step 1
- **Description:** `invalid operation: d != want (struct containing []string cannot be compared)` in `grare/parse_test.go`.
- **Severity:** Low — **Resolution:** D1 — **Time Impact:** minutes — **Prevention:** see D1.

### I2: Backfill gap understated by plan (audit F1)
- **Description:** `store.RecheckDue` re-reads only active listings in incomplete targets that list pages did not show since `crawlStart` (`last_seen < crawlStart`). Already-read listings still on list pages are never re-read → never gain a gallery; reachable re-checks skew toward sold/delisted (→ `MarkGone`).
- **Severity:** Medium (value, not correctness) — **Resolution:** README corrected (F1a); backfill itself → human decision (F1(b)) — **Prevention:** quote the gating query when scoping a gap out.

### I3: CRLF checkout vs `gofmt -w`
- **Description:** `core.autocrlf=true`; `gofmt -w` wrote LF to touched files and to untouched `internal/crawl/fetch.go`.
- **Severity:** Low — **Resolution:** git normalizes content; `fetch.go` restored by `git checkout --` — **Prevention:** gofmt only named files; check `gofmt -l` on CR-stripped committed content.

### I4: `close-precheck` could not resolve the log (close-time)
- **Description:** `close-precheck.ps1` derived `…-plan-log.md` (not found); `.sh` reported `execution log is missing or inaccessible` / `cannot resolve plan repository identity` (log lives only on the ephemeral branch in worktree mode).
- **Severity:** Low (report-only tool) — **Resolution:** gates checked manually: base `git status --porcelain --untracked-files=no` empty; worktree `status --porcelain --ignored=matching` empty; base on `master`.

---

## Key Insights

**Rationale promoted:** README § fetcher `-details` bullet (gallery from the budgeted read; re-read gap) and `listing_images` comment in `internal/store/schema.sql` (via patch) | keep-on-empty + tx rationale in `store.SaveDetail` comment | backfill rationale not promoted — pending human decision, not settled.

### Lessons Learned
1. **Frame scoped-out gaps on the real mechanism**
   - Context: plan Out of Scope said old listings gain galleries "when a re-check reads their page again"; audit found re-checks skip listings still on list pages.
   - Insight: a deferred decision is only as good as the stated scope of its gap.
   - Application: quote the gating predicate (`RecheckDue … last_seen < ?`) in Out of Scope notes.
2. **Slice fields break struct equality**
   - Context: D1.
   - Insight: adding `[]T` to a value type silently invalidates `==` call sites in tests.
   - Application: grep comparisons in the step that adds the field.

### Pattern Discoveries
1. **Fold derived data into the one existing write**
   - Observed in: `SaveDetail` + `replaceImages`.
   - Description: gallery persisted in the same tx as the product-page fields → no crawler change, no partial state (`detail_at` set without gallery).
   - Reuse potential: any further product-page field (e.g. sold date, description).
2. **Keep-on-empty for scraped collections**
   - Description: empty parse result never clears stored data → markup drift degrades to stale, not wiped; drift signal = `galleries` count in `details done` log.

### Gotchas / Pitfalls
1. **`SetMaxOpenConns(1)` + tx** — any `s.db` call while a tx is open deadlocks; helpers take `*sql.Tx`.
2. **BRAND OFF malformed first `<img>`** — swallows next `</div>`; descendant selector still exact; page-wide path matching triples photos (thumbs, zoom modal).
3. **`new-projex.sh --projex-dir .projex/closed`** creates `closed/closed/` (patch note).

### Technical Insights
- `SaveDetail` with non-empty Images on a nonexistent id now errors on the FK (was silent 0-row UPDATE) — unreachable today (ids from `PendingDetails` / `RecheckDue`; no listing deletes).
- ALLU URLs stored as given (absolute in payload), unlike brandoff/grare `base.Parse` — consistent with ALLU `ParseList`.
- Viewer-side hotlink load per detail view rises 1 → up to 9–16 photos; covered by owner "Accept as demo".

---

## Recommendations

### Immediate Follow-ups (human decision — not implemented)
- [ ] **F1(b) Backfill decision:** without it, galleries appear essentially only on listings first read after deploy; listings whose product page was already read (`detail_at` set) and that stay on list pages keep the single list photo. Option: a backfill tier spending leftover `-details` budget on `detail_at` set + no gallery → raises actual per-run product-page reads up to the full cap each crawl until drained. If approved, README sentence "most such listings never gain a gallery" must change with it.
- [ ] **F4** BRAND OFF listing first read with only the NowPrinting placeholder never gains a gallery in a complete-sweep brand → fold into F1(b).

### Future Considerations
- F5: `radar.Item.Images` comment names only `store.SetImages` — deliberately not done (`internal/radar` untouched by human requirement); fix if a comment-only radar edit is later allowed.
- F6 (informational): `galleries` log counter untested (only drift signal) | stale gallery if a shop reshoots a listing still on list pages.

### Plan Improvements
- Out of Scope note should have quoted `RecheckDue` conditions (I2).
- Step 1 should have included the grare comparison swap (D1).

---

## Related Projex Updates

### Documents Updated / Moved
| Document | Update |
|---|---|
| 2610140136-kura-gallery-collector-plan.md | Complete; + Completed / Walkthrough fields → `closed/` |
| 2610140136-kura-gallery-collector-log.md | Status Complete (verified) → `closed/` |
| 2610140203-kura-gallery-collector-audit.md | Resolved by this close (conditions: trailer on landing commit; F1(a) via patch; F1(b) raised as open decision) → `closed/` |
| 2610140209-kura-gallery-collector-audit-fixes-patch.md | Born closed — nothing to move |
| 2610130420-kura-roadmap-nav.md | In place: Phase 5 gallery milestone note (collection built; backfill open — box stays unchecked); Recent Progress; Revision Log |
| 2610130449-kura-mock-def.md | In place: L3 gallery question gains *Built* pointer line (no checkbox changed); Revision Log |

### New Projex Suggested
| Type | Description |
|---|---|
| Plan (if human approves F1(b)) | Gallery backfill tier: re-read `detail_at`-set listings lacking a gallery within leftover `-details` budget; README sentence update |

---

## Appendix

### Commits squashed
```
520c0ee projex: step 1 - SaveDetail persists Detail.Images atomically; crawler gallery count
ae9e23a projex: step 2 - BRAND OFF product-page gallery parse
dedc358 projex: step 3 - GALLERY RARE product-page gallery parse
dcd13a1 projex: step 4 - ALLU product-page gallery parse
a107eab projex: step 5 - README documents gallery collection from the budgeted product-page read
e96cc40 projex: complete kura-gallery-collector
1c3f278 projex(audit): add kura-gallery-collector
b16a1a1 fix(gallery): state README re-read gap, pin GR gallery order, reword listing_images comment
ea734ca projex(patch): add patch doc - kura-gallery-collector-audit-fixes
```

### References
- Base: `master` @ `6ffdcb5` (plan `a2a9c3b`, gallery UI `4b4b8bd`)
- Audit 2610140203-kura-gallery-collector-audit.md | Patch 2610140209-kura-gallery-collector-audit-fixes-patch.md
