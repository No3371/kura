# Execution Log: Kura Gallery Collector

> **Status:** Complete
> **Started:** 20261014 01:46
> **Repo Root:** S:/Repos/portfolio/kura
> **Plan File:** .projex/2610140136-kura-gallery-collector-plan.md
> **Parent:** 2610140136-kura-gallery-collector-plan.md
> **Base Branch:** master
> **Worktree Path:** kura/.projexwt/2610140136-kura-gallery-collector

## Pre-Check Results
```
REPO_ROOT=S:/Repos/portfolio/kura
BRANCH=master
PLAN_REL=.projex/2610140136-kura-gallery-collector-plan.md

PASS  Plan is committed (a2a9c3b projex: add plan - kura-gallery-collector)
WARN  Working tree has 2 uncommitted change(s)

PRE-CHECK PASSED
```
WARN resolution: the 2 changes are the human's untracked PNGs in the base checkout (out of bounds per orchestrator; untouched). Worktree mode isolates execution from them.

## Init
- Plan status Ready → In Progress, committed on `master` (6ffdcb5).
- Worktree: `projex-worktree` → branch `projex/2610140136-kura-gallery-collector` at 6ffdcb5.
- Bootstrap: `go mod download` (module cache, outside worktree — nothing to clean). go1.26.0 windows/amd64.
- Baseline: `go vet ./...` clean; `go test -count=1 ./...` all `ok`.
- Orchestrator constraints: no live crawl / no shop requests (fixtures + `httptest` only); backfill out of scope; never touch repo `kura.db`; no push/deploy; web/radar untouched.

## Steps

### 20261014 01:52 - Step 1: `Detail.Images` persisted by `SaveDetail`; crawler log + test
**Action:** Edited per plan: `internal/listing/listing.go` (+ `Images []string` w/ doc comment) | `internal/store/store.go` (`SaveDetail` → tx, `replaceImages(ctx, tx, id, urls)` when `len(d.Images) > 0`; `SetImages` body → `replaceImages`) | `internal/store/schema.sql` (comment names `SaveDetail`) | `internal/store/store_test.go` (+ `TestSaveDetailImages`) | `internal/crawl/crawler.go` (doc sentence; `galleries` counter + log key) | `internal/crawl/crawler_test.go` (`fakeSource.images`; + `TestDetailStoresGallery`). `gofmt -w` on touched files. **Plus** `internal/source/grare/parse_test.go`: `d != want` → `!reflect.DeepEqual(d, want)` (see Deviations).
**Result:** First `go vet ./...` failed: `internal\source\grare\parse_test.go:173:5: invalid operation: d != want (struct containing []string cannot be compared)` — slice field makes `listing.Detail` non-comparable. After the grare comparison swap: `go vet ./...` clean; `go test -count=1 ./...` all `ok`; `TestSetImages` | `TestSaveDetailImages` | `TestSaveDetailOverridesModelCode` | `TestDetailBudgetPerSource` | `TestDetailStoresGallery` PASS.
Self-verification (Verify-Projex Encouraged; done inline): tx scope — inside `SaveDetail` only `tx` used (UPDATE + `replaceImages`), no `s.db` while tx open → no deadlock under `SetMaxOpenConns(1)`. Keep-on-empty ✓ (test: p1,p2 → kept on photo-less read w/ Color updated → q1). `SetImages` behaviour unchanged (`TestSetImages` incl. dedupe/blank/clear/Version ✓). `FetchDetails` budget/target logic untouched — diff adds only counter + log key. Edge noted: with images, `SaveDetail` on a nonexistent id now errors (FK on `listing_images.listing_id`) instead of a silent 0-row UPDATE; crawler ids come only from `PendingDetails`/`RecheckDue` (existing rows) → not reachable in practice. gofmt note: checkout is CRLF (`core.autocrlf=true`, global); `gofmt -w` wrote LF to touched files (+ untouched `internal/crawl/fetch.go`, content-identical, not in `git diff`) — git normalizes, diffs are content-only.
**Status:** Success

### 20261014 01:56 - Step 2: BRAND OFF gallery parse
**Action:** `internal/source/brandoff/parse.go`: `ParseDetail(r, base *url.URL)`, doc line updated, `Images: galleryImages(doc, base)`, + `galleryImages` (`#js-main-slider .swiper-slide img` src, skip blank / `/NowPrinting`, `base.Parse`). `source.go`: `ParseDetail(bytes.NewReader(resp.Body), resp.URL)`. `parse_test.go`: `TestParseDetail` → `ParseDetail(f, testBase)` + exact 9-URL assertion; + `TestParseDetailPlaceholderPhoto`; imports + `fmt`, `slices`, `strings`. gofmt.
**Result:** `go vet ./internal/source/brandoff/` clean; `go test -count=1 -v ./internal/source/brandoff/` all PASS incl. `TestParseDetail` (exact `_LL` main + `sub01`…`sub08` absolute under `https://www.brandoff-store.com/`, slider order) and `TestParseDetailPlaceholderPhoto` (Images nil, ModelCode X1). Only caller of `brandoff.ParseDetail` is `source.go` (grep, repo-wide). Self-verification (Verify-Projex Encouraged; inline): exact-slice equality proves malformed first `<img>` neither drops nor duplicates a slide and the zoom-button `button_modalopen.svg` (outside `.swiper-slide`) is excluded — plan's fallback (`img.productimg`) not needed. `resp.URL` non-nil: same `Detail` already dereferences `resp.URL.Path`. No new `Fetcher.Get`.
**Status:** Success

### 20261014 02:00 - Step 3: GALLERY RARE gallery parse
**Action:** `internal/source/grare/parse.go`: `ParseDetail(r, base *url.URL)`, doc sentence appended ("base resolves the carousel's photo URLs (galleryImages)."), `Images: galleryImages(doc, base)`, + `galleryImages` (`figure.fs-c-productCarouselMainImage__image img` src, skip blank, `base.Parse`). Sold early-return untouched. `source.go`: `ParseDetail(bytes.NewReader(resp.Body), resp.URL)`. `parse_test.go`: 3 call sites → `(…, testBase)`; `TestParseDetail` + Images assertions (len 15, exact `-01` URL, `-15.jpg?size=l&` last), `d.Images = nil` before `reflect.DeepEqual` (comparison swap itself landed in Step 1 — see Deviations). gofmt.
**Result:** `go vet ./internal/source/grare/` clean; `go test -count=1 -v ./internal/source/grare/` all PASS (`TestParseDetail`, `TestParseDetailSold`, list tests). Fixture cross-check (grep of the 15 carousel `<img src>`): 15 distinct, `-01`…`-15`, all `size=l`, raw HTML encodes `=` as `&#x3D;` → exact-URL assertion proves entity decoding. Callers of `grare.ParseDetail`: `source.go` only.
**Status:** Success

### 20261014 02:04 - Step 4: ALLU gallery parse
**Action:** `internal/source/allu/parse.go`: `apiAsset` + `Images []apiImage \`json:"app_item_images"\``; + `apiImage{URL, Order int}`; + `(apiAsset).images()` (clone, `slices.SortStableFunc` by `cmp.Compare(Order)`, drop blank after TrimSpace); `ParseDetail` sets `Images: a.images()` (sold early-return untouched); imports + `cmp`, `slices`. `parse_test.go`: `TestParseDetail` `want.Images` = `ZZ029445.JPG`, `-1`…`-15.JPG`; + `TestParseDetailImageOrder`. `source_test.go` `TestDetail` `/p/bag`: `d.ModelCode != "M15242" || len(d.Images) != 16`. gofmt (realigned `apiAsset` struct tags — whitespace-only on the 7 existing fields).
**Result:** `go vet ./internal/source/allu/` clean; `go test -count=1 -v ./internal/source/allu/` all PASS incl. `TestParseDetail` (16 URLs exact, `reflect.DeepEqual`), `TestParseDetailImageOrder` (order 2,blank,1 → a,b), `TestDetail` (httptest server, 16 images), sold / no-node unchanged. Assumption "`order` decodes as int" ✓ (fixture 1…16 decoded; plan's `float64` fallback unused). `allu.ParseDetail` signature unchanged (no `base` — payload URLs already absolute).
**Status:** Success

### 20261014 02:07 - Step 5: README
**Action:** `README.md` 3 edits per plan: § The fetcher also `-details` bullet + the two plan sentences (gallery from same read, no extra request; pre-gallery listings keep list photo until re-check) | Flags table `-details` row `(colour, size, accessories, photo gallery)` | Limitations last bullet `Product images (the list photo and each product page's gallery) are hotlinked …`.
**Result:** `git diff README.md` shows only these three hunks. Limitations bullet rewrapped to the file's ~95-col wrapping (wording otherwise identical) — cosmetic.
**Status:** Success

### 20261014 02:10 - Post-Execution: full verification
**Action:** Plan § Verification Plan + § Acceptance Criteria, run in worktree.
**Result:**
- `go vet ./...` clean ✓ | `go test -count=1 ./...` all `ok` (cmd/kura, advisor, crawl, listing, radar, allu, brandoff, grare, store, web; fx no tests) ✓
- `git diff --stat master..projex/2610140136-kura-gallery-collector`: exactly the 16 Key Files + this log; `internal/web`, `internal/radar` (incl. `internal/web/static/**`): 0 files ✓
- `git grep -n "Get(ctx" -- internal/source internal/crawl`: identical 7 call sites on branch and `master` ✓ (no new request; `FetchDetails` budget/target code unchanged — diff adds only `galleries` counter + log key)
- Acceptance: parsers 9 / 15 / 16 ✓ | `TestParseDetailPlaceholderPhoto`, `TestParseDetailImageOrder` ✓ | `TestSaveDetailImages` p1,p2 → kept → q1 ✓ | `TestDetailStoresGallery` HERMES-0 gallery, HERMES-1 nil ✓
- Source Hygiene: grep of added non-`.projex` lines for `projex|plan|step N|.go:N|====|----` → none; rationale comments (`SaveDetail`, both `galleryImages`, `images()`) name rejected alternatives ✓
- No live crawl / no shop request (fixtures + `httptest` only) ✓; repo `kura.db*` untouched by this execution (worktree has no `kura.db`; tests use `t.TempDir()`) ✓. Spec compliance (plan as spec): every Success Criterion met; no partial/unmet.
- Quality review: diff re-read per step; no dead code; `SetImages` now a thin tx wrapper over `replaceImages`. Observations left as-is (out of scope / plan said leave): `schema.sql` `listing_images` opening line still says "Product photos beyond listings.image_url" though collected galleries lead with the main photo; `radar.Item.Images` comment names only `store.SetImages` (radar untouched per human requirement).
- Cleanup: `internal/crawl/fetch.go` working copy had LF from `gofmt -w internal/crawl` (no content diff, `git diff --ignore-cr-at-eol` empty) → restored via `git checkout -- internal/crawl/fetch.go`. No servers/containers/temp files started. Worktree `git status --short --ignored` clean (go module cache lives outside the worktree).
**Status:** Success

## Deviations
- **Step 1 — grare test comparison moved forward from Step 3.** Plan schedules `grare/parse_test.go` `d != want` → `reflect.DeepEqual` in Step 3, but Step 1's `Detail.Images` slice breaks compile of that test immediately (`go vet` fails). Swapped the comparison in Step 1 so every step commit builds/tests green; Step 3 still adds the `Images` assertions and `d.Images = nil` reset. Outcome unaffected.

## Issues Encountered
- `go vet` failure after Step 1 edits (non-comparable `listing.Detail`) → resolved per Deviations.

## Data Gathered

## User Interventions
