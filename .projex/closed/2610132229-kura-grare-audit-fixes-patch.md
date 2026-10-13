# Patch: Kura GALLERY RARE — Audit Fixes (F1 small part, F2–F8)

> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 3
> **Status:** Complete
> **Author:** Claude Opus 5.5 ([patch], sprint iteration 3)
> **Parent:** 2610132219-kura-grare-provider-2-audit.md
> **Directive:** Address findings of 2610132219-kura-grare-provider-2-audit.md (optional `[patch]` step, uid `2610131822-kura-def-coverage#3`): F1 small part only (out-of-stock product page → gone + synthetic test; README, `Sweep` doc, vetting crawl-limit line state the real sold-detection limit) and F2–F8; F6/F7 fix if small, else accept. F1 re-check cadence out of scope (sprint Deferred / next derive).
> **Source Plan:** Direct (audit findings; plan 2610132140-kura-grare-provider-2-plan.md already closed)
> **Related Projex:** 2610132219-kura-grare-provider-2-audit.md | 2610132204-kura-grare-provider-2-patch.md | 2610132140-kura-grare-provider-2-plan.md | 2610131822-kura-def-coverage-sprint.md (not edited)
> **Result:** Success

---

## Summary

GALLERY RARE product pages carrying the shop's out-of-stock notice now read as sold (gone), tested on a synthetic sold page built from the saved live fixture. README, `grare.Source.Sweep`/`Detail` docs, `crawl.Source.Detail` doc and both vetting-doc crawl-limit lines now say sold status is seen only at the one product-page fetch per listing, and an item sold later stays 在售 — no doc claims 404-only detection anymore. F2 (cross-shop comp copy), F3 (caption), F4 (stale `newer` comment), F5 (rationale), F8 (zero/short-count guard) fixed. F6, F7 accepted (reasons below). No network requests; no `kura.db` write.

| Finding | Disposition | Commit |
|---------|-------------|--------|
| F1 small part | Fixed — out-of-stock notice → gone; docs state limit | `1a95cb2` |
| F1 re-check cadence | Out of scope → sprint Deferred / next derive (orchestrator) | — |
| F2 cross-shop comps copy | Fixed — 「同型號在各家店、各等級」 in detail + methodology | `903ded1` |
| F3 caption repeats shop | Fixed — leading shop stripped from `scheme.name` | `903ded1` |
| F4 stale `Engine.newer` comment | Fixed — comment + README § Newest listings | `1a95cb2` |
| F5 thin rationale | Fixed — names the `distHTML` default-param trap | `903ded1` |
| F6 test registry leak | Accepted | — |
| F7 375 px 「未使/用」 break | Accepted | — |
| F8 zero-total guard | Fixed — count without number or below cards → error | `1a95cb2` |

---

## Changes

### F1 — out-of-stock product page → gone

**File:** `internal/source/grare/parse.go`
**Change Type:** Modified
**What Changed:**
- `ParseDetail` (L197) signature → `(d listing.Detail, sold bool, err error)`; returns `sold=true, nil` when an element matches `.fs-c-productNotice--outOfStock` (L202), before the spec-table check. Doc comment (L189–196) states the template evidence and the rejected alternative (404-only).

**File:** `internal/source/grare/source.go`
**Change Type:** Modified
**What Changed:**
- `Source.Detail` (L88) returns `ParseDetail` directly; doc: removed → 404, sold → page stays with notice.
- `Source.Sweep` doc (L47–58): "sold items surface only through a product-page 404" → sold items surface only through `Detail`, one product-page fetch per listing (those without a detail), usually in the first-seeing crawl; later sale stays active; price changes seen only while on page 1.

**File:** `internal/crawl/crawler.go`
**Change Type:** Modified
**What Changed:**
- `Source.Detail` interface doc (L22–23): gone = page no longer exists **or says the item is sold**.

**File:** `internal/source/grare/parse_test.go`
**Change Type:** Modified
**What Changed:**
- `TestParseDetail` (L150): asserts the saved live page is not sold — its scripts name the class in text, so this also proves script text does not count.
- New `TestParseDetailSold` (L184): saved fixture with `<p class="fs-c-productNotice fs-c-productNotice--outOfStock">SOLD OUT</p>` inserted before `<div class="fs-c-productActionButton`; plus a notice-only page without spec table. Both → `sold=true`, `err=nil`. Comment marks text/position synthetic, class the shop's.

**Why:** audit F1 — saved `testdata/detail_bag.html` template reacts to `.fs-c-productNotice--outOfStock` (reveals 「表記価格は最終販売した時点での価格です。」 + 商品お探しサービス button), so a sold page likely answers 200; 404-only read it as live. Evidence source: committed fixture only (no new request; no sold page captured, hence synthetic).

### F1 — docs state the limit

**File:** `README.md`
**Change Type:** Modified
**What Changed:**
- § Knowing when something sold (L115–121): "a product page that answers 404 does" → sold status read only from the product page, fetched once per listing; 404 or out-of-stock notice → gone; nothing re-checks after, so later sales stay 在售; price changes seen only on first list page; "known limit".

**File:** `docs/provider-vetting.md`
**Change Type:** Modified
**What Changed:**
- GALLERY RARE "Crawl limits" bullet (L133–137): consequence — >40-bag brands never complete; sold status seen only at the one product-page fetch (404 or out-of-stock notice); later sale stays active.
- Outcome "GALLERY RARE: Pass" line (L272–274): same limit, one clause.

### F2 — fair-value copy names cross-shop comps

**File:** `internal/web/static/js/detail.js` (L142), `internal/web/static/js/pages.js` (L65)
**Change Type:** Modified
**What Changed:** rank-adjusted sentence 「把同型號各等級的 N 件」 → 「把同型號在各家店、各等級的 N 件」; methodology 行情估價 paragraph same wording.
**Why:** `radar.Engine.value` rank-adjusted comps take every source in the group; copy said only 各等級. Cross-shop comp *labels* stay Deferred (canvas pass 2).

### F3 — ladder caption

**File:** `internal/web/static/js/dist.js`
**Change Type:** Modified
**What Changed:** `distHTML` (L111–114): caption drops a leading `shop` from `scheme.name` → 「BRAND OFF <small>商品ランク · 店家鑑定</small>」.

### F4 — `radar.Engine.newer` comment

**File:** `internal/radar/radar.go` (L519–533), `README.md` § Newest listings (L127–131; line rewrap landed with the doc commit)
**Change Type:** Modified
**What Changed:** newest-first premise scoped to sources that sweep newest first; GALLERY RARE (one page, shop order, never complete for >1-page brands, never watched) gets crawl order only — first crawl's stock stays ahead of later arrivals. README parenthetical says the same.

### F5 — `web.Server.model` rationale

**File:** `internal/web/server.go` (L237–240)
**Change Type:** Modified
**What Changed:** `// [] not null in JSON` → names the trap: nil slice → `null`; `distHTML`'s `schemes = []` default replaces only `undefined`; `null` throws on `schemes.filter`.

### F8 — misread list count guard

**File:** `internal/source/grare/parse.go`, `internal/source/grare/parse_test.go`
**Change Type:** Modified
**What Changed:**
- `ParseList`: count text without digits → error (L56–58); `Total < len(Cards)` → error (L108–110). Doc comment (L41–46) lists all three error cases and the gone-marking hazard.
- `TestParseListNoCount` → table `TestParseListCount` (L89) over helper `listPage` (L73): whole category ✓, empty category ✓, no count ✗, `件中` ✗, `1件中` with 2 cards ✗.
**Why:** `Total=0` with cards made `Complete()` true → would mark the brand's other listings gone. Guard `Total < len(Cards)` covers the audit's `Total == 0 && len(Cards) > 0` and any short misread.

### F6 — accepted (no change)

Every package test that needs a scheme (`radar` "t", `web` "wo"/"wt"/"wm"/"wx"/"wy", `advisor` "ax"/"ay") registers into `listing`'s process-global registry; fixing only F6's two tests would be inconsistent, and a cleanup needs a new exported `listing` unregister fn used only by tests. Leak stays inside each package's test binary; test sources never collide with real ones; `TestPromptStatesThresholds` asserts `Contains`, not the full shop list. Revisit if a test ever asserts the exact registered set outside `listing`.

### F7 — accepted (no change)

`word-break: keep-all` on `.k-dist__axis b` at 9 columns × ~33 px (375 px) would stop 「未使/用」 but make every wider label (中古SA, 中古AB, 未使用 ~39–42 px at 13 px) overflow into neighbours instead of wrapping — a layout change needing a browser pass, not a patch. Ladder layout is provisional (designed multi-provider UI Deferred).

---

## Verification

**Method:** `go vet ./...`; `go test -count=1 ./...`; `node --check` on edited JS modules; `distHTML` rendered in node (document stubbed) with two schemes to read the `h3` captions; gofmt on edited Go files (CRLF-normalised; repo-wide `gofmt -l` hits are CRLF only, HEAD blobs clean).

**Result:**
```
go vet ./...                      exit 0
ok  github.com/No3371/kura/cmd/kura
ok  github.com/No3371/kura/internal/advisor
ok  github.com/No3371/kura/internal/crawl
?   github.com/No3371/kura/internal/fx  [no test files]
ok  github.com/No3371/kura/internal/listing
ok  github.com/No3371/kura/internal/radar
ok  github.com/No3371/kura/internal/source/brandoff
ok  github.com/No3371/kura/internal/source/grare
ok  github.com/No3371/kura/internal/store
ok  github.com/No3371/kura/internal/web
dist.js ok | detail.js ok | pages.js ok
<h3 class="k-dist__shop">BRAND OFF <small>商品ランク · 店家鑑定</small></h3>
<h3 class="k-dist__shop">GALLERY RARE <small>商品ランク · 店家鑑定</small></h3>
```

**Status:** PASS

No browser pass: F2/F3 are copy/caption strings (checked by render above); F5/F4 comments only. No server started; no shop request.

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610132219-kura-grare-provider-2-audit.md | Parent (findings) | Related Projex → this patch; condition 1 `[x]` `[PATCHED]`; Partial Execution line; condition 2 (re-check cadence) left open |
| 2610132204-kura-grare-provider-2-patch.md | Audited execution | None (point-in-time record) |
| 2610132140-kura-grare-provider-2-plan.md | Original plan (closed) | None |
| 2610131822-kura-def-coverage-sprint.md | Sprint nav | Not edited (orchestrator records F1 follow-up in Deferred / next derive) |

---

## Notes

- **Sold-page markup is inferred, not observed.** No sold GALLERY RARE page was captured; the class comes from the live template's own scripts. If the shop renders the notice client-side only, the static HTML lacks it and the check never fires — behaviour then equals the old 404-only path, and the README limit still holds. Confirm on the next acquisition-approved product-page capture of a sold item.
- **F1 follow-up open:** per-source detail re-check for never-complete sources (oldest `detail_at` first, within budget) touches shared `store.PendingDetails` → plan-sized.
- Audit's other edge case (SOLD OUT cards on list pages ingested as active) not addressed — not in F1–F8; no list fixture shows one.
- Rationale promoted: sold-detection limit and newest-order caveat now live in README (shipped doc), not only in the audit.
