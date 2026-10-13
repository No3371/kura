# Patch: Kura listing detail page (商品頁)

> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 2
> **Status:** Complete
> **Author:** Claude Opus 5.5 (executor, sprint subagent)
> **Parent:** 2610130624-kura-listing-detail-plan.md
> **Directive:** Execute 2610130624-kura-listing-detail-plan.md in full (patch-projex plan-execution mode, sprint worktree, no new branch)
> **Source Plan:** 2610130624-kura-listing-detail-plan.md (Steps 1–7, all objectives)
> **Result:** Success
> **Commits:** `766218d` (server: TWD fields, short labels, scheme in API) | `b12c30a` (client: detail view, distribution, CSS, router) | doc commit (this file + plan close)
> **Related Projex:** 2610130624-kura-listing-detail-plan.md | 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md | 2610130535-kura-canvas-foundation-home-plan.md

---

## Summary

`/listing/{id}` now renders the canvas 商品頁 on real BRAND OFF data: crumb, photo mat, info + CTA + track toggle (synced with cards), 鑑定卡 with shop ladder, 商品資料, 行情估價 prose per valuation state, 同型號價格分布 (real group items, server medians, chart ↔ table), 售價變化 (no-change strip; step line for ≥2 points). Server now computes grade-stat and price-history TWD once in `radar`; advisor + UI read those fields. Home card click no longer dead-ends.

---

## Changes

### Server data (commit `766218d`)

**Files:** `internal/listing/condition.go` | `internal/source/brandoff/condition.go` | `internal/radar/radar.go` | `internal/radar/radar_test.go` | `internal/advisor/tools.go` | `internal/web/static/legacy/app.js` | `internal/web/server.go` | `internal/web/server_test.go`
**Change Type:** Modified
**What Changed:**
- `listing.Grade.Short` (`condition.go:15`, `json:"short,omitempty"`); BRAND OFF scheme short labels 新品 / 未使用 / A / B / C.
- `radar.PricePoint{At, PriceJPY, PriceTWD}` (`radar.go:53`); `Item.History []PricePoint`, filled in `toItem` loop (`radar.go:219`). `GradeStat` gains `MedianTWD/MinTWD/MaxTWD` (`radar.go:107`), set in `stats()` (`radar.go:574`). gofmt realigned `Item` field tags (whitespace only).
- `advisor/tools.go:250` `modelStats` grades → `g.MedianTWD, g.MinTWD, g.MaxTWD`; `tools.go:281` `getListing` → `p.PriceTWD`.
- Legacy `historyChart`: `At`/`PriceJPY` → `at`/`price_jpy` (5 sites).
- `server.go:200` listing response adds `scheme` (shop ladder).
- Tests: `TestDropsAndStats` asserts TWD fields; new `TestListingAPI` (`server_test.go:60`) — scheme shop/short, `history[0].price_twd`, `by_grade[0].median_twd`, 404 for unknown id.

**Why:** S10 one source per stat; S3 ladder data; S1 404 path.

### Client (commit `b12c30a`)

**Files:** `internal/web/static/js/util.js` | `app.js` | `card.js` | `detail.js` (new) | `dist.js` (new) | `internal/web/static/kura.css`
**Change Type:** Modified / Created
**What Changed:**
- `util.js`: `jpy`, `md` (m/d Asia/Taipei); `api()` errors carry `.status`.
- `app.js`: route `/^\/listing\/([^/]+)$/` → `renderDetail`; views get path params, may return title; `err.status === 404` → not-found view + `NOT_FOUND_TITLE` (`app.js:16`); title set inside `mine === seq` guard.
- `card.js`: exports `isTracked`, `badgeHTML`, `PENDING_ICON`, `HEART`, `OUT_ICON`; track listener updates `[data-track-text]` text for text-labelled buttons, aria-label/title for icon-only ones.
- `detail.js`: `renderDetail(main, ov, rawId)` per plan Step 5 markup + copy table; group fetch failure tolerated; sold variant (A5); step-line history (A6) with `.k-sr` list.
- `dist.js`: `distHTML({stats, items, scheme, thisId})` per plan Step 6; hover tooltip + table toggle via document-level listeners.
- `kura.css`: `[hidden]` rule (`kura.css:95`); detail (`:196`), 鑑定卡 (`:237`), lower column (`:260`), distribution (`:277`), history (`:316`), mobile overrides in the 767px block; header comment lists "listing detail".

**Why:** S1–S9.

---

## Verification

**Method:** `go vet ./...`, `go test ./...`, hygiene grep; built-in browser on `127.0.0.1:8090` (worktree `kura.db`, read-only use) and `127.0.0.1:8092` (scratch copy `mut.db` in session scratchpad, deleted after); DOM assertions via `javascript_tool` + screenshots vs canvas `Main.dc.html` detail markup.

**Result:**
```
go vet ./...            clean
go test ./...           ok advisor | radar | source/brandoff | store | web (TestListingAPI, TestStaticRoutes)
hygiene grep            none
console errors (detail) none; only expected 404 fetch logs from D11 unknown-id probe
```

| # | Case | Evidence | ✓/✗ |
|---|---|---|---|
| D1 | `brandoff:2101220216016` same-grade/high | prose 「同一家店、同樣是 ランクB 的 カルトシエールGM，目前有 16 件可以比較，中位數是 NT$20,184。這件比它低 50%，約省 NT$10,092。」; dl 同店同等級中位數 / 16 件 / 高 · 同等級價格集中; ladder aria 「店家等級：新品、未使用品、ランクA、ランクB（此件）、ランクC」; 25 dots = 25 items; table rows ランクA `8, NT$31,285, NT$16,147, NT$38,349`, ランクB `17, NT$20,184, NT$10,092, NT$26,239` = API `by_grade` TWD exactly; tooltip 「ランクA · 在售 NT$16,147」; toggle 以表格檢視 ↔ 以圖表檢視; 1280 light/dark, 375 light/dark, colorScheme dark emulation (paper `rgb(27,24,21)`) | ✓ |
| D2 | `2101220298944` medium / `2101220156220` low | 信心 中; low → dispersion 0.50 line 「同型號的價格非常分散…僅供參考。」 | ✓ |
| D3 | `2101219298115` / `2101219445359` rank-adjusted | A4 sentence 「…不到 3 件，所以把同型號各等級的 30 件換算到 ランクA 來比較…」; 方法 跨等級換算; low → 「這個估計僅供參考。」 | ✓ |
| D4 | `2101218625240` anomaly | badge 異常低價 · 先確認狀況; anomaly sentence, no 約省; 375 scrollWidth 375 | ✓ |
| D5 | `2101219439341` no estimate | 「同型號的比較樣本不足 3 件，暫不估計行情。」; dl = 行情怎麼算 only; chart present (1 dot) | ✓ |
| D6 | CHANEL `name:マトラッセ` (`brandoff:2104102463379`) | 409 dots = 409 items; axis ランクA 在售 289 / ランクB 在售 120; tooltip 「ランクB · 此件 · 在售 NT$90,827」; 375 scrollWidth 375 | ✓ |
| D7 | `2104102531481` photo pending | mat text 店家照片準備中 | ✓ |
| D8 | `2100301207912` specs/notes | 類型/顏色/材質/尺寸/付屬品/店家備註 (original 「持ち手は取り外し可能です。」)/上架 | ✓ |
| D9 | `2600041253809` empty grade | no 鑑定卡, no ladder, no distribution (no graded points), crumb has no tail (no model code either), no error | ✓ |
| D10 | Home → first card | in-app nav to `/listing/brandoff%3A2101220126889`, title 「ニュートラベルラインMM · CHANEL · Kura 蔵」; track on detail → back → same card `aria-pressed=true` | ✓ |
| D11 | unknown / malformed | `/listing/brandoff%3Anope` in-app + direct load → 找不到這個頁面 + title; `/listing/%E0%A4%A` in-app → not-found; `/listing/%E0%A4` direct load → not-found (see Notes for `%E0%A4%A` direct load) | ✓ |
| D12 | :8092 sold `2101220298944` / multi `2101220216016` | sold: 最後售價 + 已找到新主人, no badge/track/#fair, CTA kept, fine print 2nd sentence only, mat faded, dist dot 「此件 · 已售出」; multi: path `M0.00 14.29 H92.12 V85.71 H100`, labels NT$10,092 / NT$9,083, 10/10 上架 / 今天, sr list 2 items, no note; 1280 light + 375 dark | ✓ |

Other: `#fair` link scrolls (scrollY 0 → 1033, same h1 node, no re-render); Tab order crumb → CTA → track → 行情怎麼算 → toggle, `:focus-visible` outline 2px solid; at 375 no `main a/button` under 24px tall; table box 343 scrolls 560 internally; `/model/...` crumb target → not-found (expected, A1); page text has no 售出率 / 天找到新主人 / 已售出 N 件; `/legacy/` drawer 售價變化 renders single-point line (8090) and multi-point chart (8092) without NaN/Invalid Date.

**Status:** PASS

### Criteria

| Criterion | Result | Evidence |
|---|---|---|
| S1 Route | ✓ | D10, D11 |
| S2 Top block | ✓ | D1 screenshots vs canvas; track sync D10 |
| S3 鑑定卡 | ✓ | D1 ladder aria + short labels; stamp 店鑑; D9 omitted |
| S4 商品資料 | ✓ | D1, D8 |
| S5 行情估價 | ✓ | D1–D5 prose; D12 sold → absent |
| S6 分布 | ✓ | D1 table = API `by_grade`; D6; D9 absent |
| S7 售價變化 | ✓ | D1 strip + canvas copy; D12 step line |
| S8 Sold | ✓ | D12 |
| S9 Responsive/themes | ✓ | 1280/375 × light/dark; scrollWidth; focus; targets |
| S10 One source per stat | ✓ | `tools.go:250,281` read radar fields; no client JPY→TWD; banned strings absent |
| S11 vet/test/hygiene | ✓ | above |

---

## Deviations from plan

1. **Scratch DB dates (Step 7.3):** plan said `gone_at = first_seen + 2 days` and new price point at `first observed_at + 86400`. Real `first_seen` was ~6 h before now, so both would be future-dated. Instead shifted both listings' `first_seen` + first `price_points.observed_at` back 3 days, then `gone_at = first_seen + 2 days` and the new 10%-lower point (¥45,000) at the original first observation. Same states, timestamps in the past.
2. **History point labels at strip ends:** added `.k-hist__ptlabel--start/--end` (inward alignment when x < 15% / > 85%). Plan's centred labels clipped past the strip edge at 375 (first point at x=0); a drop near today would widen the page.
3. **Extra CSS beyond plan list:** `.k-cta { position: relative }` (contains its `.k-sr` text), `.k-cta svg` 13×13 (canvas CTA icon size; card `OUT_ICON` is 10×10), `.k-crumb > span` 24px min-height, `.k-detail--sold` wrapper dims sold price (card vocabulary), `.k-detail__val` near/above/none badges unpadded (card pattern).
4. **鑑定卡 foot** omitted when both `condition_by` and `condition_scheme` are empty (A11 "parts with no data omitted"; plan's markup always rendered it).
5. **Commit `766218d` trailer order:** `Projex:` trailer landed after `Co-Authored-By` (stage-n-commit `--trailer` appends last); amend blocked by repo git-safety hook, left as is. `b12c30a` and the doc commit carry trailers in the body with `Co-Authored-By` last.

## Impact on Related Projex

| Document | Relationship | Update Made |
|---|---|---|
| 2610130624-kura-listing-detail-plan.md | Source plan | All objectives `[PATCHED]`; status Complete; moved to `closed/` |
| 2610130524-kura-interactive-mock-sprint.md | Sprint nav | Not edited (orchestrator owns) |

## Notes

- **Malformed escape direct load:** `/listing/%E0%A4%A` (invalid `%A`) loaded directly gets Go `net/http` 400 「Bad Request」 before any handler runs — server-level URL parsing, not fixable in `spaFallback`. In-app navigation to it and direct load of a well-formed-but-undecodable id (`%E0%A4`) both reach the not-found view. Orchestrator: S1 holds for every URL the server accepts.
- **Sold + single point:** strip keeps canvas note 「上架以來還沒有調價。店家調價時，會畫在這條線上。」 per plan; end label = gone date (m/d). The second sentence reads forward-looking on a sold item — candidate for canvas pass 2 copy review alongside A4/A5.
- **D6 density:** 289-dot ランクA column hides its median tick under dots (plan Risk, accepted); table carries exact values.
- Orchestrator log items from plan Open Questions (A1, A4, A5, A6, A7) unchanged.
- Canvas read this run: `Main.dc.html` version `1791840357-ac34` (artifact read returned current; pinned `1791839497-46ec` not selectable for file reads) — detail values matched the plan's transcription.
- Worktree: no untracked leftovers; `kura.db` + `-wal` unmodified (mtime unchanged, 05:25; latest `fx_rates` row predates the run); `-shm` touched by WAL readers only.
