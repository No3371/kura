# Patch: Kura browse page (價格最美)

> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 5
> **Status:** Complete
> **Author:** Claude Opus 5.5 (patch-projex plan-execution, sprint subagent)
> **Parent:** 2610130913-kura-browse-page-plan.md
> **Directive:** Execute `2610130913-kura-browse-page-plan.md` in full (patch-projex plan-execution mode) on the sprint branch; verify per the plan's own checks.
> **Source Plan:** 2610130913-kura-browse-page-plan.md (Steps 1–5, all objectives)
> **Result:** Success
> **Commits:** `2f15a43` feat(radar) — Step 1 | `265eea0` feat(web) — Steps 2–4 | doc commit (this file + plan close)
> **Related Projex:** 2610130913-kura-browse-page-plan.md | 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md | 2610130602-kura-canvas-foundation-home-audit.md (F3 link contract honoured; F1 / F6 untouched)

---

## Summary

Canvas Browse at `/browse` on real BRAND OFF data: crumb `首頁 / {heading}`, h1 + `{total} 件`, desktop 篩選與排序 row (品牌 / 類型 / 最低成色 / 預算上限（NT$） / 關鍵字 / 排序 / 只看低於行情 switch), mobile form (關鍵字 + 篩選與排序 disclosure + switch), card grid (desktop 4-col stack cards at 1280; mobile row cards with rules), 「再看 N 件」 load-more by 40. URL query = state (`replaceState` per change; reload restores both forms). Server: `/api/listings?offset=` + overview `by_category`. Header search now in-app (`/browse?q=…`, no reload). Closes shell dead ends: Home 看全部 ×2, header search form, mobile search icon; dark × browse built token-driven.

---

## Changes

### Radar paging + category counts; API offset (Step 1, `2f15a43`)

**File:** `internal/radar/radar.go` — Modified
- `Filter.Offset` + comment (l.385–387); `Search` slices `out[min(max(f.Offset, 0), len(out)):]` after `total` (l.430)
- `Overview.ByCategory` `json:"by_category"` (l.633), init (l.643), active count `o.ByCategory[it.Category]++` (l.652)

**File:** `internal/radar/radar_test.go` — Modified: `TestSearchFilters` += offset page = `all[1]`, offset past end → `[]` + total 3, `ByCategory["tote"] == 3` (l.192–202)

**File:** `internal/web/server.go` — Modified: `listings` reads `offset` (l.185)

**File:** `internal/web/server_test.go` — Modified: `TestListingAPI` += `getList` helper; `offset=1` page differs from first, total 2; `offset=9` → no items, total 2 (l.115–138)

**Why:** paging where the sorted match list lives (one sort, one predicate); 類型 options from the same index as counts.

### Shell (Step 2, `265eea0`)

**File:** `internal/web/static/js/app.js` — Modified
- import + route `[/^\/browse$/, renderBrowse, "價格最美 · Kura 蔵"]` (l.5, l.16)
- `go(url)` helper (l.75) shared by link click handler and new header-search `submit` listener (l.95: `form.k-search` → `/browse?q=…` | `/browse`, fresh state; form keeps `action`/`method` as no-JS fallback)

**File:** `internal/web/static/js/home.js` — Modified: `export const EMPTY_HTML` (l.9, same markup) + one-line comment; `renderHome` uses it (l.19)

**File:** `internal/web/static/js/model.js` — Modified: 看全部 comment present tense — reason now "Browse has no model-group filter" (l.115)

### CSS (Step 3, `265eea0`)

**File:** `internal/web/static/kura.css` — Modified
- Header comment lists browse
- Browse block after 型號頁 (l.363–): `.k-field`, `.k-browse__head`/count, `.k-bf` + width classes, `.k-switch`, `.k-bgrid` (auto-fill 240, gap 48/28, busy opacity), `.k-bnone`, `.k-bmore`
- Mobile `@media (max-width: 767px)` additions (l.537–553): h1 24px, wide form hidden / narrow shown (44px fields), toggle button, panel column, list layout with `padding 20px 0` + rule, row-card rules applied by width (comment names the rejected JS-layout alternative)

### Browse view (Step 4, `265eea0`)

**File:** `internal/web/static/js/browse.js` — Created (260 lines)
- Constants `PAGE`, `SORTS`, `CONDS` (rank-floor comment), `FILTER_ICON` (l.13–19)
- `brandOpts` / `catOpts` (l.26/32): non-empty, count desc; brand label `{name}（{count}）`
- `readState` (l.42) whitelist → `stateQuery` (l.58, canonical order, `sort=deal` omitted) → `apiURL` (l.68, sort/limit/offset explicit) | `heading` (l.79, A2)
- Markup helpers `selectHTML` / `fieldsHTML` / `keywordHTML` / `switchHTML` / `cellsHTML` / `moreHTML` / `resultsHTML` (l.84–118)
- `renderBrowse` (l.120): site-empty → `EMPTY_HTML`; two forms (comment: rejected CSS `order`); returns `{heading} · Kura 蔵`
- `syncControls` (l.152), `apply` (l.159: URL-equality guard, `replaceState`, heading/title/count, seq-guarded fetch, error line), `loadMore` (l.201: offset = cards shown, focus first new card's model link)
- Document-level listeners (l.230–260): change → apply; Enter (not during IME composition) → apply; submit prevented; toggle + load-more clicks

---

## Verification

**Method:** Step 5 procedure — vet/test/hygiene, S8 diff check, local serve of worktree `kura.db` on :8090 (no crawl flags; binary built to session scratchpad) + fresh scratch DB on :8092; built-in browser matrix B1–B14 driven by real clicks/typing/keys where the case is about input (B7, B5 budget + keyword, B13), dispatched `change` events elsewhere.

**Result:**
```
go vet ./...                       → clean
go test ./...                      → ok advisor | radar | brandoff | store | web
hygiene grep (internal cmd)        → no matches
git diff aa020a0 -- radar.go | grep 0.05|anomalyPct|dealScore|confidenceWeight → no matches
API probe (:8090)                  → deals=1 1471 | min_rank S 831 / A 10152 / B 15322 | max_twd=20000 2082 | category=travel 37 | q=ネヴァーフル 290 / +deals 63 | GOYARD 98 / +deals 0 | offset=99999 → 0 items, total 15367
overview.by_category               → shoulder 6861 handbag 4741 tote 2073 backpack 524 bag 361 boston 297 belt 286 clutch 126 business 61 travel 37
```

| # | Result | Evidence |
|---|---|---|
| B1 | ✓ | Home link click → `/browse?deals=1&sort=deal`; h1/crumb 價格最美, `aria-current`; `1,471 件` = Home 「看全部價格最美 1,471 件」; first 3 ids = Home shelf ids, same order; switch on, sort deal; 40 cards; 「再看 1,431 件」; title `價格最美 · Kura 蔵` |
| B2 | ✓ | 看全部最新上架 → h1 最新上架, switch off, sort newest, `15,367 件`, first 5 ids = Home 最新上架 row |
| B3 | ✓ | 再看 ×2 → 80 / 120 cards, button 1,391 → 1,351, 0 duplicate ids, focus on card 41 / 81 model link |
| B4 | ✓ | `brand=GOYARD` → 40 「再看 58 件」 → 80 「再看 18 件」 → 98, button gone, focus card 81 |
| B5 | ✓ | URLs `?deals=1&brand=CHANEL` (279 = API) · `?category=travel` 37 · `?min_rank=S|A|B` 831 / 10,152 / 15,322 · typed `2,0000` + Enter → `?max_twd=20000`, field shows 20000 in both forms, 2,082 · typed keyword + Enter → `q`; via change 290, +switch 63, off 290 · `price_asc` / `price_desc` card prices monotone · heading 價格最美 ↔ 全部在售 ↔ 最新上架 with switch/sort · reload of 7-param URL → identical controls in both forms, count 2 = API |
| B6 | ✓ | `?deals=1&brand=GOYARD` → 「沒有符合條件的包。」, `0 件`, no load-more |
| B7 | ✓ | real typing ネヴァーフル + Enter in header → `/browse?q=…`, h1 全部在售, keyword filled, 290, `window.__probe === 1` (no reload); empty + Enter → `/browse`, 15,367 |
| B8 | ✓ | 375: header search icon → `/browse`, 全部在售, 15,367 |
| B9 | ✓ | Home → B1 → brand CHANEL → card → detail → Back: `?deals=1&brand=CHANEL`, CHANEL selected, 279 → Back: `/` |
| B10 | ✓ | `?brand=NOPE&sort=x&min_rank=Z&max_twd=-5&deals=yes` → 全部在售, 15,367, all controls default; `/browse/x` → 找不到這個頁面 + its title |
| B11 | ✓ | 375: toggle `aria-expanded` false→true, `aria-controls=bf-panel`, panel shows 5 fields at 44px full width; brand CHANEL + switch off → `?brand=CHANEL`, 3,979, panel stays open; at 1280 wide form shows CHANEL + switch off |
| B12 | ✓ | 1280 light (`?theme=light`) + dark (OS / emulation): 4 columns, field widths 150/120/160/130/flex/150 exact, field 38px, gap 48/28; 375 light + dark: list with `20px 0` cells + 1px rule, row cards `116px | 1fr`, h1 24px, `scrollWidth 375 = innerWidth`; dark tokens on body / fields / rules / load-more |
| B13 | ✓ | Tab order 首頁 → brand → category → min_rank → max_twd → q → sort → deals → cards, every stop `:focus-visible` with 2px solid ring; narrow form not in tab order; Space on switch → off, URL `/browse`; Enter on 再看 → 80 cards, focus card 41 |
| B14 | ✓ | :8092 `/browse?deals=1` → 「還在準備中。」 (Home too) |
| Regr. | ✓ | Home 5 + 3 cards, links unchanged; model マトラッセ 看全部 reveals 404 hidden cards in place; console errors 0 on every view; no canvas sample value (`1,020 件` / `1,015`) on page |

**Status:** PASS

### Success criteria

| Criterion | Result | Evidence |
|---|---|---|
| S1 Route | ✓ | B1, B2, B7, B8, B10 (reload via direct navigate), titles |
| S2 Head | ✓ | B1 count = Home N = API; shelf ids/order; B2 |
| S3 Desktop filters | ✓ | B1 option lists (6 brands count desc, 10 類型 count desc, 4 成色, 4 sorts), placeholders, `role="switch"`; B5; B12 widths |
| S4 Mobile | ✓ | B8, B11, B12 375 |
| S5 Results | ✓ | B1, B3, B4, B6, B14 |
| S6 Server | ✓ | Step 1 tests; API probe; advisor builds its own `Filter` (no `Offset`) — tests green |
| S7 Responsive/themes | ✓ | B12, B13; console 0 errors |
| S8 One source per stat | ✓ | counts from API `total` / overview only; radar predicate + sort diff check clean |
| S9 vet/test/hygiene | ✓ | above |

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|---|---|---|
| 2610130913-kura-browse-page-plan.md | Source plan | Status → Complete; criteria + verification checked; `[PATCHED]` + partial-execution note; moved to `closed/` |
| 2610130524-kura-interactive-mock-sprint.md | Sprint nav | Not edited (orchestrator absorbs) |
| 2610130449-kura-mock-def.md | Def (canvas gaps / deviations) | Not edited — plan's Open Questions list the items for the orchestrator to log (A2, A4, A5, A7, A8, A9, A12) |

---

## Notes

**Deviations from plan (all small, reason given):**
- **Option-set whitelist.** `readState` whitelists `brand` / `category` against the *offered* options (brands/types with active listings), not all of `ov.brands` / `ov.categories`. A catalog brand with 0 active listings in the URL would otherwise filter to 0 results while the select (no such option) shows 全部品牌 — URL and controls disagreeing, which A1 rules out ("invalid values are ignored; control shows the default").
- **`.k-bf__panel[hidden]` rule omitted.** Base `[hidden] { display: none !important; }` already beats `.k-bf__panel { display: flex }`; the plan's extra rule would be redundant.
- **IME guard.** Enter listener skips `e.isComposing`, so the Enter that confirms Japanese/Chinese IME text (e.g. typing ネヴァーフル) does not commit a half-composed keyword. Not in plan; needed for the keyword field's stated use.
- **Load-more disabled during `apply`.** `apply` also sets `disabled` on 再看 while the new first page loads, so a click cannot append the old filter's next page into a grid about to be replaced (plan relied on the seq guard alone, which does not catch this ordering).
- **`EMPTY_HTML` comment** added (one line naming its role); `apiURL` built with `URLSearchParams.set` (no `?&sort=` when the state is empty) — same query as planned.
- **Server binary** built to the session scratchpad and run there (`serve -db <worktree>/kura.db`) instead of `go run` — same code, faster restart; deleted afterwards.

**Observations (not fixed — out of scope):**
- Card model links render ~22px tall at 375 (shared card component since iteration 1; inline-text target, WCAG 2.5.8 inline exception). The switch checkbox is 18px but its label (the click target) is 38/44px.
- `?theme=` is dropped from the URL on the first filter change (`stateQuery` writes only state params); the theme attribute persists until reload. Theme param is a review aid, not product state.
- Scaffold first created with `--projex-dir .projex/closed`; the script adds `closed/` for born-closed types itself, so it landed in `closed/closed/` — moved to `.projex/closed/` before any commit (`--projex-dir .projex` is the right argument).
- The pane screenshot tool tiled the 375 frame once; layout was confirmed by computed styles, not only pixels.
