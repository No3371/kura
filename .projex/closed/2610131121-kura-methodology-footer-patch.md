# Patch: Kura footer IA + 行情怎麼算

> **Status:** Complete
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 7
> **Author:** Claude Opus 5.5 (patch executor, sprint subagent)
> **Parent:** 2610131104-kura-methodology-footer-plan.md
> **Directive:** Execute 2610131104-kura-methodology-footer-plan.md in full (patch-projex plan-execution mode; orchestrator accepted A1, A2, A3, A9)
> **Source Plan:** 2610131104-kura-methodology-footer-plan.md (Steps 1–4, all)
> **Result:** Success
> **Commits:** `c19a8a7` (code) | doc commit (this file, plan close, def gap update)
> **Related Projex:** 2610131104-kura-methodology-footer-plan.md | 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md (Canvas gaps § Pages updated)

---

## Summary

`/methodology` (行情怎麼算) built per plan Step 2 copy: crumb + h1 + lead + `#source` `#fair` `#badges` `#history` `#demo`. Every number on it reads `ov.method` (radar's exported constants, new block on `/api/overview`), `ov.schemes`, `ov.overview` or `card.js` `NEAR_BAND`. `/about` `/faq` `/contact` `/privacy` `/terms` → pending view, h1 = footer link text. Detail 行情估價 prose + advisor prompt now state the same constants instead of literals. Constant values unchanged (F1 / F6 undecided).

---

## Changes

### Radar constants
**File:** `internal/radar/radar.go` — Modified
- `minComps` → exported `MinComps` (l.26–28, doc comment); new `const (DealMinPct = 0.05; AnomalyPct = 0.5)` block (l.33–47) carrying the moved anomaly rationale + "consumers state these from here" note
- `value()` uses `MinComps` (l.319, l.326), `v.Anomaly = v.DealPct > AnomalyPct` (l.354); old `const anomalyPct` removed
- `IsDeal` uses `DealMinPct` (l.465); doc comment names the constant

**Why:** one definition per threshold, read by every consumer that states it.

### Scheme list
**File:** `internal/listing/condition.go` — Modified
- `Schemes()` (l.66–77): all registered schemes sorted by source (imports `slices`, `strings`)

### Overview payload
**File:** `internal/web/server.go` — Modified
- `overview()` (l.177–183): `method {deal_min_pct, anomaly_pct, min_comps, min_sold_history_days}` + `schemes`

### Advisor prompt
**File:** `internal/advisor/prompt.go` — Modified
- systemPrompt anomaly line → 「比行情低超過 {{anomaly_pct}}%」 (l.38); replacer fills it (l.53–54); `pctText` helper (l.58–59); `only_deals` description built from `radar.DealMinPct` (l.81)

### Tests
**Files:** `internal/web/server_test.go`, `internal/advisor/agent_test.go` — Modified
- `TestStaticRoutes` + `/js/pages.js`, `/methodology`, `/about`
- `TestOverviewMethod` (new): `method` = radar constants; `schemes` contains registered `wo` (membership, registry is process-wide)
- `TestPromptStatesThresholds` (new): prompt has 「低超過 50%」-from-constant, no `{{`; `only_deals` description has "at least 5%"-from-constant

### Page module
**File:** `internal/web/static/js/pages.js` — Created
- `renderMethodology(el, ov)` — copy verbatim from plan Step 2; ladders from `ov.schemes` (label / 中文 / 「日文」, `rank` never rendered); badges via `badgeHTML`; `since` only when `Date.parse > 0`; 「…不會出現在價格最美」 sentence only when `band < deal`
- `renderPending(el, ov, slug)` — label from `.k-footer nav a[href="/{slug}"]`, returns `{label} · Kura 蔵`
- Header comments carry rationale + rejected alternatives (typed numbers go stale; not-found reads broken; JS label map = second copy)

### Router / badge / detail / styles
- `app.js`: import + two routes (`/methodology`, `/(about|faq|contact|privacy|terms)`)
- `card.js`: `export const NEAR_BAND = 0.03` used in both comparisons; comment explains client ownership (rejected: server constant → overview threaded into every card render) and names `radar.IsDeal` / `radar.DealMinPct` instead of literal 0.05
- `detail.js`: `fairHTML(it, v, name, m)` with `ov.method` — 「不足 ${m.min_comps} 件」, 「不到 ${m.min_comps} 件」, 「低超過 ${pct(m.anomaly_pct)}%」
- `kura.css`: `.k-method*` + `.k-pending` block after `.k-error` (l.194–209); mobile: h1 26px, dl single column (l.583–584). Additions beyond plan sketch: `.k-method` `padding-bottom: 48px`, `.k-method strong { font-weight: 600 }`, `.k-method__dl dt { min-width: 0 }`

---

## Verification

**Method:** `go vet` + `go test -count=1 ./...`; local serve of worktree DB (built binary in session scratchpad, no crawl flags, stopped by PID); built-in browser at 1280 / 375 × light / dark; greps.

**Result:**
```
go vet ./...                  → clean
go test -count=1 ./...        → ok advisor | radar | source/brandoff | store | web
  TestStaticRoutes PASS | TestOverviewMethod PASS | TestPromptStatesThresholds PASS | TestIsDeal PASS
gofmt (LF-normalized changed .go files) → clean
/api/overview method → {anomaly_pct: 0.5, deal_min_pct: 0.05, min_comps: 3, min_sold_history_days: 14}
/api/overview schemes → [brandoff "BRAND OFF 商品ランク" BRAND OFF 店家鑑定 faq.aspx, 5 grades]
grep -nE "[0-9]+ ?%|[0-9]+ ?件|[0-9]+ ?天" pages.js → none
grep -n "3 件|一半" detail.js → none
grep -n "rank|更新|最後" pages.js → only header comment ("names no internal rank")
hygiene: added lines grep projex|[0-9]{10}-|// Step|==== → none
```

| Criterion | Evidence | Status |
|---|---|---|
| S1 Routes | In-app click on all 6 footer links from `/` (window marker preserved = no reload): titles `關於 Kura · Kura 蔵` … `服務條款 · Kura 蔵`, `行情怎麼算 · Kura 蔵`; h1 = link text; pending link → `/methodology`. Back → `/privacy` → `/contact`, Forward → `/privacy`. Direct load `/methodology`, `/privacy`, `/faq`, `/terms` render. `/nope` → 「找不到這個頁面」, title `找不到這個頁面 · Kura 蔵` | ✓ |
| S2 Detail link | detail 行情估價 「行情怎麼算」 click → `/methodology`, title set, no reload; Back → listing with `#fair` | ✓ |
| S3 Content | crumb 首頁 / 行情怎麼算; h1; lead; ids `source fair badges history demo` with h2 店家資料 / 行情估價 / 標籤怎麼看 / 需要時間累積的數字 / 哪些是示範; BRAND OFF 商品ランク（店家鑑定）+ 店家的等級說明 link; 5 grades 新品 / 未使用品 / ランクA / ランクB / ランクC with 中文 + 「日文」; 6 badges | ✓ |
| S4 Disclosure | `#demo` lists exactly 登入與示範會員 / 追蹤清單 / 通知 / 大家都在找 + closing paragraph (shop data / computed; advisor same computations; 示範模式 meaning). No new label elsewhere | ✓ |
| S5 One source | tests above; page reads 3% band, 5% 價格最美, 50% anomaly, 14 天, 3 件; detail prose 「不足 3 件」 (brandoff:2101219728858), 「不到 3 件」 rank-adjusted (brandoff:2101219298115), 「低超過 50%」 anomaly (brandoff:2101220095338) | ✓ |
| S6 Not stated | no rank codes / factors / mapping; no data-age / last-updated; no contact / legal; `DealMinPct` 0.05, `AnomalyPct` 0.5 unchanged; `since` shown as record start only (`Kura 自 2026/10/13 起記錄`); 目前的記錄還不到 14 天 | ✓ |
| S7 Look | 1280 light + dark: badge column 216 px, widest badge 171 px, all single-line; scrollWidth 1265 ≤ 1280. 375 light + dark: scrollWidth 375 = innerWidth, no element past viewport, dl single column (343 px), h1 26px. Focus: crumb 首頁, 店家的等級說明, pending 行情怎麼算 → `:focus-visible` 2px solid outline, height 24 px. Console: 0 messages | ✓ |
| S8 Build | vet / test / hygiene above | ✓ |

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|---|---|---|
| 2610131104-kura-methodology-footer-plan.md | Source plan | All steps `[PATCHED]` → this patch; Status Complete; moved to `closed/` |
| 2610130449-kura-mock-def.md | Def (Canvas gaps, A4) | § Canvas gaps Pages: 行情怎麼算 built (provisional); five footer pages reachable as pending; Revision Log row |
| 2610130524-kura-interactive-mock-sprint.md | Sprint | Not edited (orchestrator owns nav) |

---

## Notes

- **Deviation (minor, CSS):** three declarations beyond the plan's CSS sketch (`.k-method` bottom padding, `strong` weight, `dt` `min-width: 0`); `13.5em` first column kept — widest badge fits at 1280.
- **Deviation (verification method):** S2 / detail variants driven by `nav.js` `go()` + element `.click()` from JS rather than Home → card mouse path; same in-app router path. Screenshots in the pane intermittently timed out / rendered blank after programmatic scroll; layout checks rest on DOM metrics plus the screenshots that did render (1280 light/dark badges + ladder, 375 light/dark top + badges, pending page).
- Pending pages at 375 keep `.k-empty` h1 40px (existing site-empty style, same as not-found) — no mobile override added; not in plan.
- Observed: brandoff:2101220216016 reads 「這件比它低 50%」 without anomaly (DealPct ≤ 0.5, rounds to 50). Consistent with strict `>` and page copy 「低超過 50%」; display rounding can make the boundary look identical — relevant to F6, not decided here.
- Short pending pages show the footer mid-viewport (pre-existing shell; not-found does the same).
- Worktree cleanup: built binary + scratch JSON/scripts removed from session scratchpad; no untracked files left in the worktree; `kura.db*` untouched.
