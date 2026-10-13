# Kura Gallery Collector — product-page photos into the stored gallery

> **Status:** In Progress
> **Author:** Claude Opus 5.5 (plan, under orchestrator)
> **Parent:** Orchestrator
> **Source:** Direct request — human prompts: "I want a gallery. Like clickable thumbnails below the main image, when the db have multiple images crawled" | "The point is our service only cares about images crawled. Decoupled from the crawler." | "Start a opus as the orchestrator to do it" (photo collector)
> **Related Projex:** 2610130449-kura-mock-def.md (§ Is Display: owner decision "Accept as demo" — photos with attribution + link-out; L3 gallery fallback) | 2610130420-kura-roadmap-nav.md (Phase 5 milestone "Photo gallery: all product images per listing (real; detail-page / API image fields; budgeted)") | 2610140010-kura-sold-observation-plan.md (detail budget + re-check tiers this reuses) | 2610132307-kura-allu-provider-3-plan.md, 2610132140-kura-grare-provider-2-plan.md (adapter precedents)
> **Worktree:** Yes — base checkout has untracked files (human's) and a pre-existing `projex/sprint/2610130524-kura-interactive-mock` branch; another worktree (`.claude/worktrees/home-shelf-hero`, branch `home-shelf-hero`) also sits at `4b4b8bd`. Leave all three alone.

---

## Summary

Gallery UI + storage shipped in `4b4b8bd`: `listing_images` table, `store.SetImages`, `Row.Images`, `radar.Item.Images` (`images` in the API), detail-page thumbnails. Nothing writes galleries yet. This plan makes the crawler collect them: each adapter's product-page parser returns the page's photo URLs in `listing.Detail.Images`, and `store.SaveDetail` writes them into `listing_images` in the same transaction as the other product-page fields. The product page is already fetched (first fetch, re-checks) within each source's `-details` budget → **zero additional shop requests**. Photo bytes are never downloaded by Kura; viewers' browsers hotlink them, as with the list photo today.

**Scope:** `listing.Detail` + `store.SaveDetail` + crawler log/test + 3 adapters' `ParseDetail` (+ tests) + README. Web, radar, JS, CSS untouched.
**Estimated Changes:** 16 files — `internal/listing/listing.go`, `internal/store/{store.go, schema.sql (comment), store_test.go}`, `internal/crawl/{crawler.go, crawler_test.go}`, `internal/source/brandoff/{parse.go, source.go, parse_test.go}`, `internal/source/grare/{parse.go, source.go, parse_test.go}`, `internal/source/allu/{parse.go, parse_test.go, source_test.go}`, `README.md`.

**Plan-time requests:** none. All evidence ← saved fixtures under each package's `testdata/`. The executor makes no shop requests either (tests use fixtures + `httptest`).

---

## Objective

### Problem / Gap / Need
Def L3 / nav Phase 5: gallery must be real. The detail page shows thumbnails only when `listing_images` holds ≥2 URLs; no code path fills it, so every listing shows the single list photo. All three shops' product pages carry full galleries (verified on fixtures, plan-time):

| Source | Fixture | Gallery markup | Count |
|---|---|---|---|
| BRAND OFF | `brandoff/testdata/detail_bag.html` | `#js-main-slider .swiper-slide img` src — relative: `/Contents/ProductImages/0/2101220315016_LL.jpg`, then `/Contents/ProductSubImages/0/2101220315016_sub01_LL.jpg` … `_sub08_LL.jpg` | 9 |
| GALLERY RARE | `grare/testdata/detail_bag.html` | `figure.fs-c-productCarouselMainImage__image img` src — absolute `https://galleryrare.itembox.cloud/product/593/000000559316/000000559316-01.jpg?size=l&w=NjAw&t=20261010201553` … `-15.jpg?size=l…` (figure attrs also carry `xs` / `xl` variants) | 15 |
| ALLU | `allu/testdata/detail_bag.html` | `__NUXT_DATA__` → `data["market:asset:…"].app_item_images[]` = `{id, url, order}`; `https://image.nanboya.com/ALLU/9480610/ZZ029445.JPG`, `-1.JPG` … `-15.JPG`, `order` 1…16 | 16 |

Markup notes: BRAND OFF slider's first `<img>` lacks its closing `>` (swallows the next `</div>` → later slides nest inside slide 1); a `.swiper-slide img` descendant selector still matches each img once. The slider also contains the zoom button's `button_modalopen.svg`, outside any `.swiper-slide`. Whole-page matching on `ProductSubImages` finds each photo 4× (slider, thumbs, zoom modal).

### Success Criteria
- [ ] `listing.Detail` has `Images []string`; each adapter's `ParseDetail` fills it from its fixture: BRAND OFF exactly the 9 absolute `_LL` URLs in slider order; GALLERY RARE 15 `size=l` URLs `-01`…`-15`; ALLU 16 URLs `ZZ029445.JPG`, `-1`…`-15`.
- [ ] BRAND OFF `NowPrinting` placeholder never enters `Images`; ALLU images sorted by `order`, blank URLs dropped.
- [ ] `store.SaveDetail` with non-empty `d.Images` replaces the listing's gallery atomically with the detail fields; with empty `d.Images` keeps the stored gallery.
- [ ] Crawl end-to-end (fake source): a product page with photos → `Row.Images` equals them after `RunAll`; a page without → `Row.Images` nil.
- [ ] Product-page request count per run unchanged: no new `Fetcher.Get` call anywhere; `FetchDetails` budget logic untouched.
- [ ] `internal/web`, `internal/radar`, `internal/web/static/**` unchanged (`git diff --stat` shows none).
- [ ] `go vet ./...` and `go test ./...` pass.

### Out of Scope
- **Backfill** of listings already read (`detail_at` set, no gallery). They gain a gallery only when a re-check reads their page again (targets whose sweep did not complete: GALLERY RARE, ALLU, capped/failed BRAND OFF). BRAND OFF listings of completed sweeps keep the list photo. A backfill tier spending leftover `-details` budget would raise actual per-run reads (up to the full cap every crawl) → human decision, not taken here.
- Downloading, caching, resizing, or proxying photos (hotlink only, as today).
- Thumbnail-size variants for the UI (`detail.js` uses one URL per photo; UI untouched).
- `docs/provider-vetting.md` R2 (b) wording ("the product photo (hotlinked)"); def § Is Display already covers "photos" with attribution + link-out.
- Nav / def checkbox updates (close-time).
- Any live crawl.

---

## Context

### Current State
- `crawl.Source.Detail(ctx, url) (listing.Detail, gone bool, err)`; `Crawler.FetchDetails` (`internal/crawl/crawler.go`) spends `DetailBudget` per source: `store.PendingDetails` (never read) then `store.RecheckDue` (re-checks), calls `src.Detail`, then `Store.MarkGone` (gone) or `Store.SaveDetail(ctx, id, d, now)` (ok).
- `store.SaveDetail` (`internal/store/store.go`): single `UPDATE listings SET color, material, size, accessories, notes, detail_at, model_code/group_key …`. No transaction.
- `store.SetImages(ctx, id, urls)`: tx → `DELETE FROM listing_images WHERE listing_id=?` → insert trimmed, deduped, non-blank URLs at `pos` 0…n. Callers: `store_test.go` `TestSetImages`, `web/server_test.go` `TestListingAPI`.
- `store.Open`: `db.SetMaxOpenConns(1)` → statements inside a tx must use the tx, never `s.db` (deadlock).
- `listing_images.listing_id REFERENCES listings(id)`, `PRAGMA foreign_keys = ON`.
- `Version` includes `COUNT(*)` + `MAX(id)` of `listing_images`; `detail_at` already moves it on every product-page save.
- `detail.js` `photoHTML`: `it.images?.length ? it.images : [image_url]` — the stored gallery **replaces** the list photo, so `Images` must be the full ordered set with the main photo first (all three shops lead with it).
- Parsers: `brandoff.ParseDetail(r io.Reader) (listing.Detail, error)`; `grare.ParseDetail(r io.Reader) (d listing.Detail, sold bool, err error)` (returns early on sold); `allu.ParseDetail(r io.Reader) (d listing.Detail, sold bool, err error)` via `node[apiAsset](root, "market:asset:")` (returns early on sold). Each `ParseList` already takes `base *url.URL` and resolves image srcs with `base.Parse`.
- Baseline (plan-time, `4b4b8bd`, go1.26.0 windows/amd64): `go test ./...` all `ok`.

### Key Files

| File | Role | Change Summary |
|------|------|----------------|
| `internal/listing/listing.go` | `Detail` shape | + `Images []string` |
| `internal/store/store.go` | persistence | `SaveDetail` → tx + gallery replace; `replaceImages` shared with `SetImages` |
| `internal/store/schema.sql` | `listing_images` comment | name `SaveDetail` as writer |
| `internal/store/store_test.go` | store tests | + `TestSaveDetailImages` |
| `internal/crawl/crawler.go` | `FetchDetails` | log count of reads that yielded photos; doc line |
| `internal/crawl/crawler_test.go` | fake source | `images` map + `TestDetailStoresGallery` |
| `internal/source/brandoff/parse.go`, `source.go`, `parse_test.go` | BRAND OFF | `ParseDetail(r, base)` + `galleryImages`; tests |
| `internal/source/grare/parse.go`, `source.go`, `parse_test.go` | GALLERY RARE | `ParseDetail(r, base)` + `galleryImages`; tests (`!=` → `reflect.DeepEqual`) |
| `internal/source/allu/parse.go`, `parse_test.go`, `source_test.go` | ALLU | `apiAsset.Images` + `images()`; tests |
| `README.md` | shipped docs | gallery collection, no extra request |

### Dependencies
- **Requires:** `4b4b8bd` (gallery storage + UI) on `master` ✓.
- **Blocks:** nav Phase 5 "Photo gallery" milestone; any backfill decision.

### Constraints
- Crawl invariants (human-confirmed): identified KuraBot UA; robots.txt + Crawl-delay; no CAPTCHA/WAF evasion; per-run product-page reads stay ≤ `-details` per source. This plan adds no request.
- Web/radar side reads only what is stored — no change there.
- Each provider keeps its own `listing.Scheme` (untouched).
- Never touch the repo's `kura.db`; tests use `t.TempDir()`. No network, no push, no deploy.
- Source Hygiene (SKILL.md): no projex refs in code comments; rationale comments name the rejected alternative.

### Assumptions
- Fixtures are representative of live pages (same templates). Verify early: Step 2–4 tests.
- `resp.URL` (`crawl.Response.URL`, `*url.URL`) is the final product-page URL after redirects — already relied on by `brandoff.Source.Detail` (`resp.URL.Path`) and every `ParseList` call.
- A page's first gallery photo is its main photo (true on all three fixtures).
- ALLU `app_item_images[].order` is an integer (fixture: 1…16); hydrated numbers round-trip `float64` → JSON → `int` as other `apiAsset` ints do.

### Impact Analysis
- **Direct:** files above.
- **Adjacent:** `crawl.Source` interface unchanged (adapters' `Detail` signatures unchanged; only exported `ParseDetail` of brandoff/grare gains `base`). `SetImages` behaviour unchanged (refactor only). `Version` moves on gallery writes (already moves on `detail_at`).
- **Downstream:** `/api/…` `images` field becomes populated → detail page thumbnails appear for listings read after deploy. `radar.Item.Images` comment names `store.SetImages` — still true (also `SaveDetail`); leave radar untouched per human requirement.

---

## Implementation

### Overview
Parsers extract → `Detail.Images` → `SaveDetail` persists in one tx. Crawler logic stays; only a log counter + test. Order: store/crawler first (Step 1, testable with fake source), then each adapter independently (Steps 2–4), docs last.

### Step 1: `Detail.Images` persisted by `SaveDetail`; crawler log + test

**Objective:** Any `Detail` carrying photos lands in `listing_images` with the rest of the product-page fields, atomically.
**Confidence:** High
**Depends on:** None
**Verify-Projex:** Encouraged (tx scope with `MaxOpenConns(1)`; keep-vs-clear rule)

**Files:**
- `internal/listing/listing.go`
- `internal/store/store.go`
- `internal/store/schema.sql`
- `internal/store/store_test.go`
- `internal/crawl/crawler.go`
- `internal/crawl/crawler_test.go`

**Changes:**

`listing.go` — `Detail` gains a field (after `Notes`):
```go
// Before:
	Notes       string
	FetchedAt   time.Time
}

// After:
	Notes       string
	// Images is the product page's photo gallery in the shop's order, as
	// absolute URLs with shop placeholders left out; nil when the page shows
	// none.
	Images    []string
	FetchedAt time.Time
}
```

`store.go` — `SaveDetail` in a tx; `SetImages` body moves into `replaceImages`:
```go
// Before:
// SaveDetail stores product-page fields. A model code from the spec table is
// authoritative and replaces the one parsed from the title.
func (s *Store) SaveDetail(ctx context.Context, id string, d listing.Detail, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE listings SET ... WHERE id=?`, ...)
	return err
}
...
func (s *Store) SetImages(ctx context.Context, id string, urls []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	...DELETE + INSERT loop...
	return tx.Commit()
}

// After:
// SaveDetail stores product-page fields. A model code from the spec table is
// authoritative and replaces the one parsed from the title. Photos the page
// yielded (d.Images) replace the stored gallery in the same transaction; a
// page that yielded none keeps it.
// Rejected: clearing the gallery on empty d.Images — a page whose gallery
// markup an adapter stops recognising would silently wipe each gallery as
// re-checks read it. Rejected: the crawler calling SetImages after SaveDetail
// — a failure between the two leaves detail_at set, PendingDetails never
// returns the listing again, and its gallery is never collected.
func (s *Store) SaveDetail(ctx context.Context, id string, d listing.Detail, now time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `UPDATE listings SET
		color=?, material=?, size=?, accessories=?, notes=?, detail_at=?,
		model_code = CASE WHEN ?<>'' THEN ? ELSE model_code END,
		group_key  = CASE WHEN ?<>'' THEN UPPER(?) ELSE group_key END
		WHERE id=?`,
		d.Color, d.Material, d.Size, d.Accessories, d.Notes, now.Unix(),
		d.ModelCode, d.ModelCode, d.ModelCode, d.ModelCode, id); err != nil {
		return err
	}
	if len(d.Images) > 0 {
		if err := replaceImages(ctx, tx, id, d.Images); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// SetImages replaces listing id's gallery with urls, in order. Blank and
// repeated URLs are dropped; an empty list clears the gallery.
func (s *Store) SetImages(ctx context.Context, id string, urls []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := replaceImages(ctx, tx, id, urls); err != nil {
		return err
	}
	return tx.Commit()
}

// replaceImages is SetImages' work inside the caller's transaction (the store
// has one connection, so it must not touch s.db while tx is open).
func replaceImages(ctx context.Context, tx *sql.Tx, id string, urls []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM listing_images WHERE listing_id=?`, id); err != nil {
		return err
	}
	seen := map[string]bool{}
	pos := 0
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		if _, err := tx.ExecContext(ctx, `INSERT INTO listing_images (listing_id, pos, url) VALUES (?, ?, ?)`, id, pos, u); err != nil {
			return err
		}
		pos++
	}
	return nil
}
```
(`database/sql` and `strings` are already imported in `store.go`; the `UPDATE` text is unchanged.)

`schema.sql` — comment on `listing_images`:
```sql
-- Before:
-- shop's order. Written by whatever collects them (SetImages); the service
-- After:
-- shop's order. Written by whatever collects them (SaveDetail on each
-- product-page read, or SetImages); the service
```
(rest of comment unchanged; schema itself unchanged.)

`store_test.go` — new test after `TestSetImages`:
```go
func TestSaveDetailImages(t *testing.T) {
	ctx := context.Background()
	s := open(t)
	l := item("a", 100)
	if _, err := s.RecordSeen(ctx, "HERMES", l, time.Now()); err != nil {
		t.Fatal(err)
	}
	stored := func() Row {
		t.Helper()
		rows, err := s.AllRows(ctx)
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range rows {
			if r.ID == l.ID() {
				return r
			}
		}
		t.Fatalf("no row %s", l.ID())
		return Row{}
	}
	save := func(d listing.Detail) {
		t.Helper()
		if err := s.SaveDetail(ctx, l.ID(), d, time.Now()); err != nil {
			t.Fatal(err)
		}
	}

	save(listing.Detail{Color: "c", Images: []string{"p1", "p2"}})
	if r := stored(); !slices.Equal(r.Images, []string{"p1", "p2"}) || r.Color != "c" {
		t.Errorf("first read: images %v color %q", r.Images, r.Color)
	}
	// A read that yields no photos keeps the gallery but still saves fields.
	save(listing.Detail{Color: "d"})
	if r := stored(); !slices.Equal(r.Images, []string{"p1", "p2"}) || r.Color != "d" {
		t.Errorf("photo-less re-read: images %v color %q", r.Images, r.Color)
	}
	// A read with photos replaces it.
	save(listing.Detail{Color: "d", Images: []string{"q1"}})
	if r := stored(); !slices.Equal(r.Images, []string{"q1"}) {
		t.Errorf("re-read with photos: images %v", r.Images)
	}
}
```

`crawler.go` — `FetchDetails`: doc gains one sentence; count reads with photos:
```go
// Before:
// not complete in this crawl — that no list page showed since crawlStart.
func (c *Crawler) FetchDetails(...) error {
	...
	var ok, gone, failed int
	...
		default:
			ok++
			saveErr = c.Store.SaveDetail(ctx, t.ID, d, c.now())
	...
	c.Log.Info("details done", "source", src.Name(), "first", first, "rechecks", len(todo)-first,
		"fetched", ok, "gone", gone, "failed", failed)

// After:
// not complete in this crawl — that no list page showed since crawlStart.
// A read's photos replace the listing's gallery (store.SaveDetail), so
// galleries cost no request beyond the product page itself.
func (c *Crawler) FetchDetails(...) error {
	...
	var ok, gone, failed, galleries int
	...
		default:
			ok++
			if len(d.Images) > 0 {
				galleries++
			}
			saveErr = c.Store.SaveDetail(ctx, t.ID, d, c.now())
	...
	c.Log.Info("details done", "source", src.Name(), "first", first, "rechecks", len(todo)-first,
		"fetched", ok, "galleries", galleries, "gone", gone, "failed", failed)
```

`crawler_test.go` — `fakeSource` gains `images map[string][]string // product page URL -> its photos`; `Detail` returns `listing.Detail{Color: "c", Images: f.images[url]}, false, nil`. New test:
```go
func TestDetailStoresGallery(t *testing.T) {
	fa := &fakeSource{name: "fa", targets: []string{"HERMES"},
		pages:  map[string][][]listing.Listing{"HERMES": bags("fa", "HERMES", 2)},
		images: map[string][]string{"https://fa/HERMES-0": {"https://img/0a.jpg", "https://img/0b.jpg"}}}
	c, st := newCrawler(t, 5, fa)
	if err := c.RunAll(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := row(t, st, "fa:HERMES-0").Images; !slices.Equal(got, []string{"https://img/0a.jpg", "https://img/0b.jpg"}) {
		t.Errorf("HERMES-0 gallery = %v", got)
	}
	if got := row(t, st, "fa:HERMES-1").Images; got != nil {
		t.Errorf("HERMES-1 gallery = %v, want none", got)
	}
}
```
(`bags` builds URL `"https://" + source + "/" + id` → `https://fa/HERMES-0`; `row(t, st, id)` exists in `crawler_test.go`; `slices` already imported.)

**Rationale:** `SaveDetail` is the one write per product-page read; folding the gallery in keeps the crawler unchanged and atomic. Keep-on-empty protects against silent markup drift. `replaceImages` keeps one implementation of the dedupe/position rules.

**Verification:** `go test ./internal/store/ ./internal/crawl/` — new tests pass; `TestSetImages`, `TestSaveDetailOverridesModelCode`, existing crawl tests unchanged and passing.

**If this fails:** `git -C <work-root> checkout -- internal/listing internal/store internal/crawl`.

---

### Step 2: BRAND OFF gallery parse

**Objective:** `brandoff.ParseDetail` returns the slider's photos as absolute URLs.
**Confidence:** High
**Depends on:** Step 1 (`Detail.Images`)
**Verify-Projex:** Encouraged (malformed first `<img>`; selector scope)

**Files:**
- `internal/source/brandoff/parse.go`
- `internal/source/brandoff/source.go`
- `internal/source/brandoff/parse_test.go`

**Changes:**

`parse.go`:
```go
// Before:
// ParseDetail parses a product page's spec list (<dl><dt>商品品番</dt><dd>…</dd>…).
func ParseDetail(r io.Reader) (listing.Detail, error) {
	...
	return listing.Detail{
		...
		Notes:       specs["備考"],
	}, nil
}

// After:
// ParseDetail parses a product page's spec list (<dl><dt>商品品番</dt><dd>…</dd>…)
// and its photo gallery; base resolves the page's relative photo paths.
func ParseDetail(r io.Reader, base *url.URL) (listing.Detail, error) {
	...
	return listing.Detail{
		...
		Notes:       specs["備考"],
		Images:      galleryImages(doc, base),
	}, nil
}

// galleryImages returns the main slider's photos (#js-main-slider), resolved
// against base, in page order: the main shot, then sub01, sub02, …. Only
// .swiper-slide images count. Rejected: matching ProductImages /
// ProductSubImages paths page-wide — the thumbnail strip and zoom modal repeat
// every photo. The shop's NowPrinting placeholder (shown until studio shots
// exist) is left out; rejected: keeping it — the stored gallery would replace
// the detail page's 店家照片準備中 state with the placeholder image.
func galleryImages(doc *goquery.Document, base *url.URL) []string {
	var out []string
	doc.Find("#js-main-slider .swiper-slide img").Each(func(_ int, img *goquery.Selection) {
		src := strings.TrimSpace(img.AttrOr("src", ""))
		if src == "" || strings.Contains(src, "/NowPrinting") {
			return
		}
		if u, err := base.Parse(src); err == nil {
			out = append(out, u.String())
		}
	})
	return out
}
```

`source.go` `Detail`:
```go
// Before:
	d, err := ParseDetail(bytes.NewReader(resp.Body))
// After:
	d, err := ParseDetail(bytes.NewReader(resp.Body), resp.URL)
```

`parse_test.go` — `TestParseDetail` calls `ParseDetail(f, testBase)` and adds:
```go
	want := []string{"https://www.brandoff-store.com/Contents/ProductImages/0/2101220315016_LL.jpg"}
	for i := 1; i <= 8; i++ {
		want = append(want, fmt.Sprintf("https://www.brandoff-store.com/Contents/ProductSubImages/0/2101220315016_sub%02d_LL.jpg", i))
	}
	if !slices.Equal(d.Images, want) {
		t.Errorf("Images =\n%v\nwant\n%v", d.Images, want)
	}
```
New test:
```go
func TestParseDetailPlaceholderPhoto(t *testing.T) {
	page := `<div id="js-main-slider"><div class="swiper-wrapper">` +
		`<div class="swiper-slide"><img src="/Contents/ProductImages/NowPrinting_LL.jpg"></div>` +
		`</div></div><div id="tab1-content"><dl><dt>商品品番</dt><dd>X1</dd></dl></div>`
	d, err := ParseDetail(strings.NewReader(page), testBase)
	if err != nil {
		t.Fatal(err)
	}
	if d.Images != nil || d.ModelCode != "X1" {
		t.Errorf("detail = %+v; want no images, model code X1", d)
	}
}
```
Imports gain `fmt`, `slices`, `strings`.

**Rationale:** Mirrors `ParseList(r, base)`; resolving in the parser keeps `Detail.Images` absolute for every source.

**Verification:** `go test ./internal/source/brandoff/` — 9 URLs exactly, placeholder test passes, existing spec assertions unchanged.

**If this fails:** If count ≠ 9 because of the malformed first `<img>`, inspect with `doc.Find("#js-main-slider img")` and narrow by `img.productimg` (all 9 slider imgs carry `class="productimg"`; the zoom button img does not). Revert: `git -C <work-root> checkout -- internal/source/brandoff`.

---

### Step 3: GALLERY RARE gallery parse

**Objective:** `grare.ParseDetail` returns the carousel's display-size photos.
**Confidence:** High
**Depends on:** Step 1
**Do-Projex:** Encouraged

**Files:**
- `internal/source/grare/parse.go`
- `internal/source/grare/source.go`
- `internal/source/grare/parse_test.go`

**Changes:**

`parse.go`:
```go
// Before:
func ParseDetail(r io.Reader) (d listing.Detail, sold bool, err error) {
	...
		Notes:       specs["状態"],
	}, false, nil
}

// After:
// (existing doc comment kept; append the sentence:)
// … base resolves the carousel's photo URLs (galleryImages).
func ParseDetail(r io.Reader, base *url.URL) (d listing.Detail, sold bool, err error) {
	...
		Notes:       specs["状態"],
		Images:      galleryImages(doc, base),
	}, false, nil
}

// galleryImages returns the product carousel's photos
// (figure.fs-c-productCarouselMainImage__image img), resolved against base, in
// page order. Each slide's img src is the carousel's display size (size=l).
// Rejected: the figure's data-enlarged-image-url (size=xl) — bigger files for
// a mat that shows them no larger; data-sub-image-url (size=xs) is a
// thumbnail, too small as the main photo.
func galleryImages(doc *goquery.Document, base *url.URL) []string {
	var out []string
	doc.Find("figure.fs-c-productCarouselMainImage__image img").Each(func(_ int, img *goquery.Selection) {
		src := strings.TrimSpace(img.AttrOr("src", ""))
		if src == "" {
			return
		}
		if u, err := base.Parse(src); err == nil {
			out = append(out, u.String())
		}
	})
	return out
}
```
Sold path (early return) unchanged: a sold page yields no `Detail`.

`source.go` `Detail`:
```go
// Before:
	return ParseDetail(bytes.NewReader(resp.Body))
// After:
	return ParseDetail(bytes.NewReader(resp.Body), resp.URL)
```

`parse_test.go`:
- Every `ParseDetail(x)` → `ParseDetail(x, testBase)` (`TestParseDetail` ×2, `TestParseDetailSold` ×1).
- `TestParseDetail`: `listing.Detail` is no longer comparable (slice field) → replace `if d != want` with:
```go
	if len(d.Images) != 15 ||
		d.Images[0] != "https://galleryrare.itembox.cloud/product/593/000000559316/000000559316-01.jpg?size=l&w=NjAw&t=20261010201553" ||
		!strings.Contains(d.Images[14], "/000000559316-15.jpg?size=l&") {
		t.Errorf("Images = %v", d.Images)
	}
	d.Images = nil
	if !reflect.DeepEqual(d, want) {
		t.Errorf("detail\n got %+v\nwant %+v", d, want)
	}
```
(`reflect`, `strings` already imported.)

**Rationale:** Same shape as Step 2; URLs already absolute but resolving keeps parity with `ParseList`.

**Verification:** `go test ./internal/source/grare/` — 15 URLs, entity-decoded (`&`, `=`), sold tests unchanged.

**If this fails:** `git -C <work-root> checkout -- internal/source/grare`.

---

### Step 4: ALLU gallery parse

**Objective:** `allu.ParseDetail` returns `app_item_images` URLs sorted by `order`.
**Confidence:** High
**Depends on:** Step 1
**Do-Projex:** Encouraged

**Files:**
- `internal/source/allu/parse.go`
- `internal/source/allu/parse_test.go`
- `internal/source/allu/source_test.go`

**Changes:**

`parse.go`:
```go
// Before:
type apiAsset struct {
	...
	ButtonType     int    `json:"button_type"`
}
...
		Notes:       cleanText(a.ListingComment),
	}, false, nil

// After:
type apiAsset struct {
	...
	ButtonType     int        `json:"button_type"`
	Images         []apiImage `json:"app_item_images"`
}

// apiImage is one product photo of a product payload.
type apiImage struct {
	URL   string `json:"url"`
	Order int    `json:"order"`
}

// images returns a's photo URLs by the payload's order field, blank ones left
// out. The sort is stable, so equal or missing orders keep payload order.
// Rejected: payload order alone — it matches order on both fixtures, but if it
// ever differs the main photo (order 1) would not lead the gallery, and the
// stored gallery replaces the list photo on the detail page.
func (a apiAsset) images() []string {
	imgs := slices.Clone(a.Images)
	slices.SortStableFunc(imgs, func(x, y apiImage) int { return cmp.Compare(x.Order, y.Order) })
	var out []string
	for _, im := range imgs {
		if u := strings.TrimSpace(im.URL); u != "" {
			out = append(out, u)
		}
	}
	return out
}
...
		Notes:       cleanText(a.ListingComment),
		Images:      a.images(),
	}, false, nil
```
Imports gain `cmp`, `slices`.

`parse_test.go`:
- `TestParseDetail` `want` gains `Images`:
```go
	const img = "https://image.nanboya.com/ALLU/9480610/ZZ029445"
	images := []string{img + ".JPG"}
	for i := 1; i <= 15; i++ {
		images = append(images, img+"-"+strconv.Itoa(i)+".JPG")
	}
	want := listing.Detail{ /* fields unchanged */ Images: images}
```
- New test:
```go
func TestParseDetailImageOrder(t *testing.T) {
	page := `<script type="application/json" id="__NUXT_DATA__">` + mustJSON(devalue(map[string]any{
		"data": map[string]any{"market:asset:1": map[string]any{
			"button_type": 1,
			"app_item_images": []any{
				map[string]any{"url": "https://image.nanboya.com/b.JPG", "order": 2},
				map[string]any{"url": " ", "order": 3},
				map[string]any{"url": "https://image.nanboya.com/a.JPG", "order": 1},
			},
		}},
	})) + `</script>`
	d, sold, err := ParseDetail(strings.NewReader(page))
	if err != nil || sold {
		t.Fatalf("sold %v err %v", sold, err)
	}
	if want := []string{"https://image.nanboya.com/a.JPG", "https://image.nanboya.com/b.JPG"}; !slices.Equal(d.Images, want) {
		t.Errorf("Images = %v, want %v", d.Images, want)
	}
}
```
(`devalue`, `mustJSON`, `strconv`, `slices`, `strings` exist in the test package.)

`source_test.go` `TestDetail`: the `/p/bag` check becomes `d.ModelCode != "M15242" || len(d.Images) != 16`.

**Rationale:** Payload already decoded into `apiAsset`; one field + helper.

**Verification:** `go test ./internal/source/allu/` — 16 URLs in order; order test passes; sold/no-node tests unchanged.

**If this fails:** If `order` fails to decode (non-int), change `Order` to `float64` and `cmp.Compare` stays valid. Revert: `git -C <work-root> checkout -- internal/source/allu`.

---

### Step 5: README

**Objective:** Shipped docs state galleries come from the product-page read already budgeted.
**Confidence:** High
**Depends on:** Steps 1–4
**Do-Projex:** Encouraged

**Files:**
- `README.md`

**Changes:**
- § The fetcher also, `-details` bullet — append after "…come out of the same cap, after first fetches.":
  `The same read records the page's photo gallery (URLs only, in the shop's order), so galleries cost no extra request. Listings whose page was read before galleries were collected keep the list photo until a re-check reads it again.`
- Flags table, `-details` row: `(colour, size, accessories)` → `(colour, size, accessories, photo gallery)`.
- Limitations, last bullet: `Product images are hotlinked` → `Product images (the list photo and each product page's gallery) are hotlinked`.

**Rationale:** README is the shipped record of request behaviour; the gallery must not read as a new request class.

**Verification:** `git -C <work-root> diff README.md` shows only these three edits.

**If this fails:** `git -C <work-root> checkout -- README.md`.

---

## Verification Plan

### Automated Checks
- [ ] `go vet ./...` clean
- [ ] `go test ./...` all `ok` (baseline all `ok` at `4b4b8bd`)
- [ ] `git -C <work-root> diff --stat master` lists only the 16 code/doc files in Key Files plus `.projex/` bookkeeping (no `internal/web/**`, `internal/radar/**`)
- [ ] `git -C <work-root> grep -n "Get(ctx" -- internal/source internal/crawl` vs the same on `master`: same call sites (no new request)

### Manual Verification
- [ ] Source Hygiene: new comments carry no projex refs / line numbers; rationale comments name rejected alternatives.
- [ ] No live crawl, no request to any shop host; no file under the repo's `kura.db*` touched.

### Acceptance Criteria Validation
| Criterion | How to Verify | Expected Result |
|-----------|---------------|-----------------|
| Parsers fill `Images` | `TestParseDetail` in brandoff / grare / allu | 9 / 15 / 16 URLs, exact first + last |
| Placeholder + order rules | `TestParseDetailPlaceholderPhoto`, `TestParseDetailImageOrder` | none / sorted, blank dropped |
| Atomic save, keep-on-empty | `TestSaveDetailImages` | p1,p2 → kept → q1 |
| End-to-end crawl | `TestDetailStoresGallery` | gallery on HERMES-0, none on HERMES-1 |
| No new requests | grep `Get(ctx` vs `master`; `FetchDetails` budget code unchanged in diff | identical call sites |
| Service side untouched | `diff --stat` | no web/radar/static paths |
| Suite green | `go test ./...` | all `ok` |

---

## Rollback Plan

1. Abandon the ephemeral branch: `{projex-scripts}/projex-abandon.{sh|ps1} <repo-root> master projex/2610140136-kura-gallery-collector --worktree`.
2. No data migration to undo: schema unchanged; galleries already written to any scratch DB are harmless (the UI reads them; `SetImages(id, nil)` clears one).

---

## Notes

### Risks
- **Markup drift** (any shop changes its gallery template): parser yields none → stored gallery kept (keep-on-empty), new listings fall back to the list photo. Detectable via the `galleries` count in `details done` logs.
- **Viewer-side load on shop hosts:** a detail view now hotlinks up to 9–16 full-size photos (BRAND OFF from `www.brandoff-store.com` itself, ALLU from `image.nanboya.com`, GALLERY RARE from `galleryrare.itembox.cloud`) instead of 1. Not crawler traffic; covered by the owner's display decision (photos with attribution + link-out). Thumbnails `loading="lazy"` already in `detail.js`.
- **Stale gallery** if a shop reshoots and the listing is never re-read: acceptable for a demo; re-checks refresh it.
- **`AUTOINCREMENT` growth:** each re-check rewrites ≤16 rows; ids grow without bound but stay far below int64 range.

### Split Verdict
No split — single scope (`kura/.projex/`), 5 steps; size heuristic needs > 5 steps as well as > 500 lines, and Steps 2–4 depend on Step 1's `Detail.Images`.

### Open Questions
- None blocking execution. Backfill of already-read listings is a deliberate Out of Scope item pending a human decision on request volume (see Out of Scope).
