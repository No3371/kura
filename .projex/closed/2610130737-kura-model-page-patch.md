# Patch: Kura model page (型號頁)

> **Status:** Complete
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 3
> **Author:** Claude Opus 5.5 (patch executor, sprint subagent)
> **Parent:** 2610130722-kura-model-page-plan.md
> **Directive:** Execute plan 2610130722-kura-model-page-plan.md in full (patch-projex plan-execution mode) on the sprint branch
> **Source Plan:** 2610130722-kura-model-page-plan.md (Steps 1–5, all)
> **Related Projex:** 2610130722-kura-model-page-plan.md | 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md | 2610130705-kura-listing-detail-audit-polish-patch.md (F3 malformed-escape precedent)
> **Result:** Success
> **Code commit:** `a97d6a8` (`feat(web): model page at /model/{brand}/{group_key}`, trailer `Projex: 2610130737-kura-model-page`)

---

## Summary

Canvas 型號頁 built at `/model/{brand}/{group_key}` on real BRAND OFF data: crumb (A2), model mat (photo per A3 or canvas bag icon), id block, 2×2 stat tiles (real or 資料不足), 追蹤這個型號 toggle on the shared track Set, model-mode 同型號價格分布 with the shop ladder from the new model API `scheme`, 「在售的 {model}」 row capped at 5 with in-place 看全部. Detail crumb no longer dead-ends. All 12 matrix rows (M1–M12) pass; one plan CSS gap found in verification and fixed (mobile `align-items`, see Deviations).

---

## Changes

### Model API scheme (Step 1)

**File:** `internal/web/server.go` — Modified
- `model` handler (L223–230): response built as `res`; `scheme` = `listing.SchemeFor(items[0].Source)` when found. Comment names the rejected alternative (single field vs per-shop ladders once several shops are ingested).

**File:** `internal/web/server_test.go` — Modified
- `TestStaticRoutes` (L35–36): `/model/CHANEL/name%3A…マトラッセ` + `/model/HERMES/name%3A24%2F24` → 200 shell.
- `TestListingAPI` (L75–79): second listing `SourceID 2`, `ModelCode c/2`; (L112–143): `GET /api/models/HERMES/C1` → scheme shop テスト店, 1 item, active 1; `C%2F2` → 200, `Items[0].ID == wt:2`; `NOPE` → 404.

**Why:** distribution ladder needs the shop scheme; mirrors detail's `listing` handler.

### Router + track label hook (Step 2)

**File:** `internal/web/static/js/app.js` — Modified
- L4 `import { renderModel }`; L15 route `[/^\/model\/([^/]+)\/([^/]+)$/, renderModel, "型號頁 · Kura 蔵"]`.

**File:** `internal/web/static/js/card.js` — Modified
- Track listener: off label = `text.dataset.trackText || "追蹤這件"`; listener comment extended (model 追蹤這個型號 shares this listener instead of a second toggle).

### Styles (Step 3)

**File:** `internal/web/static/kura.css` — Modified
- Header comment += model page; `.k-detail__info h1` → `.k-detail__info h1, .k-model__info h1` (desktop + mobile).
- New `/* 型號頁 */` block after 售價變化 (L335–362): `.k-model__top`, `.k-model__mat` (+img/svg), `.k-model__info`, `.k-mstats` tiles (`.k-mstat__v`, `.is-na`, `.k-mstat__note`), `.k-mtrack`, `.k-dtrack--model` (44px, 17px bell, unfilled when pressed), `.k-model__dist`, `.k-msale__head`, `.k-more`.
- Mobile (L386–391): `.k-model__top` flex column **with `align-items: stretch`** (deviation, below), `.k-model__mat svg` 110px.

### Model view (Step 4)

**File:** `internal/web/static/js/model.js` — Created (112 lines)
- `renderModel(main, ov, rawBrand, rawKey)`: decode in one try (bad escape → 404), fetch model API, markup per plan (crumb `aria-current="page"`, mat photo `{model} 型號照片` / `aria-hidden` icon, id block, `dl.k-mstats`, track button `data-track="model:{brand}/{group_key}"` + `data-track-text="追蹤這個型號"`, dist wrapper, 在售 section). Returns `{name} · {brand} · Kura 蔵`.
- `soldKnown(ov)` = `ov.overview.sold > 0` with rationale comment (A5); `tilesHTML` per plan table (A6 tie rule via strict `>`, A7 `不到 1 天`, `is-na` 資料不足); `saleHTML` with `CAP = 5`, hidden cards past 5, `<button data-model-more>`.
- Document-level 看全部 listener: unhide, remove button, focus 6th card's model link; comment names rejected link to unbuilt Browse.

---

## Verification

**Method:** plan Step 5 — automated checks, hygiene grep, built binary serving worktree `kura.db` on 127.0.0.1:8090 and a mutated scratch copy (`mut.db`, plan Step 5.3 SQL; 2 gone + 3 active on 027633CC, 001855CC gone) on :8092, built-in browser matrix, page-state JS probes vs `/api/models/…` + `/api/overview`.

**Automated:**
```
go vet ./...            → clean
go test ./...           → all ok (internal/web 0.87s incl. extended TestStaticRoutes / TestListingAPI)
grep -rnE "[0-9]{10}-[a-z0-9-]+-(plan|sprint|def|nav|audit|patch)" internal cmd → no matches
```

**Matrix (1280 light unless noted):**

| # | Result | Evidence |
|---|---|---|
| M1 | ✓ | CHANEL `name:マトラッセ`: tiles 在售 409 件 · 各等級合計 / 已售出 資料不足 · 自 10/13 起記錄 / ランクA 中位數 NT$100,919 · 289 件比較 (API by_grade ランクA 289 = largest) / 平均找到新主人 資料不足; 409 dots, 0 此件, legend without 此件; 5 cards + 「看全部 409 件 →」; click → 409 visible, button removed, focus on card 6; keyboard Tab→Enter → same, focus ring visible (`:focus-visible`, 2px solid) |
| M2 | ✓ | HERMES 027633CC: 5 cards, no 看全部; ランクA 中位數 NT$736,705 · 4 件比較 = API; table rows ランクA 4/736,705/484,409/807,348, ランクB 1/343,123 = by_grade; 2 medians, 5 dots |
| M3 | ✓ | HERMES 001855CC: 1 card; sub 手提包 · 001855CC |
| M4 | ✓ | LV M52052 (stats image NowPrinting) → real item photo `…2104102519120_L.jpg`; CHANEL AS4906 (only item pending) → canvas icon `aria-hidden=true`, 140px desktop / 110px mobile |
| M5 | ✓ | HERMES `name:エルメスケリー28cm`: 3 tiles (no median), no distribution, no console error |
| M6 | ✓ | `name:24/24ヴァンキャトル29`: direct load + `location.reload()` (nav type reload) render, title `24/24 ヴァンキャトル29 · HERMES · Kura 蔵`; via detail crumb (href `%2F`) in-app; 375 no overflow |
| M7 | ✓ | `name:2.55マトラッセ` reload: sub = 肩背包 only; ランクB 中位數 NT$60,551 · 10 件比較 (B 10 > A 4) |
| M8 | ✓ | LV M51165 tie ランクA 28 / ランクB 28 → ランクA 中位數 NT$40,367 |
| M9 | ✓ | Home card → detail → crumb → model → card → detail → Back → model → Back → first detail; titles per view; 1 navigation entry (all in-app) |
| M10 | ✓* | `/model/HERMES/NOPE`, `/model/NOBRAND/X` direct → not-found + 「找不到這個頁面 · Kura 蔵」; `/model/HERMES/%E0%A4%A` in-app and `/model/HERMES/%E0%A4` direct → not-found. *Direct load of `%E0%A4%A` → Go `net/http` 400 before any handler — same server-level case accepted in iteration 2 (F3: S1 holds for every URL the server accepts) |
| M11 | ✓ | Toggle → 追蹤中 `aria-pressed=true` (accent ink, bell unfilled); detail 追蹤這件 unaffected and flips 追蹤這件 ↔ 追蹤中; crumb back → still 追蹤中; toggle off → 追蹤這個型號 |
| M12 | ✓ | :8092 027633CC: 在售 3 件 / 已售出 2 件 · 自 10/3 起記錄 / ランクA 中位數 NT$807,348 · 3 件比較 / 平均找到新主人 6 天; 2 rings + 3 dots; 3 cards. 001855CC: 在售 0 件, no tile 3, 已售出 1 件, no 在售 section, distribution 1 ring. AS4906: 已售出 0 件 |

**Responsive / themes:** M1, M2 at 1280 + 375 × `?theme=light|dark`; OS dark via colorScheme emulation (no `?theme`) → body `rgb(27,24,21)`, 資料不足 `--k-ink-3`. At 375: `scrollWidth 375 == innerWidth` on M1, M2, M6, AS4906; 在售 row scrolls in its own box (`scrollWidth 1064 / clientWidth 343`, overflow-x auto); mat + info 343px wide; h1 26px. Track button 44px tall, 看全部 24px min-height.
**Tab order (M2/M1):** crumb → 追蹤這個型號 → 以表格檢視 → (看全部) → cards; every stop `:focus-visible` with 2px solid outline.
**S8 strings:** page text has none of 人在找 / 售出率 / 近 30 天 / 此件 on every probed model.
**Console:** onlyErrors empty on normal views; only the deliberate M10 404/400 resource errors.
**Canvas:** `project/Main.dc.html` 型號頁 markup read (Artifact read) — mat/info/tiles/track/在售 head values match the CSS; remaining differences are the plan's logged deviations (A2 crumb, A3 mat padding, A4 no demand line, A7 note, A9 button).

**Cleanup:** servers stopped, viewport reset, tab closed, `mut.db*` + built binary + logs deleted from scratchpad; worktree `kura.db*` untouched.

**Status:** PASS

### Criteria

| Criterion | Result | Evidence |
|---|---|---|
| S1 Route | ✓ (F3 caveat) | M6, M7, M9, M10 |
| S2 Hero | ✓ | M1/M3/M4 screenshots + probes; crumb A2, photo/icon A3, no demand line |
| S3 Stat tiles | ✓ | M1, M2, M5, M7, M8, M12 = API fields; 資料不足 on every real-DB model |
| S4 Track | ✓ | M11 |
| S5 Distribution | ✓ | M1/M2 model mode (no 此件), ladder 5 cols from `scheme`, table = by_grade `*_twd`; M5 absent |
| S6 在售 row | ✓ | M1 (409, focus card 6), M2 (5, no link), M12 (0 → absent) |
| S7 Responsive/themes | ✓ | after mobile `align-items` fix |
| S8 One source | ✓ | model.js renders only server fields; no JPY math; string grep clean |
| S9 vet/test/hygiene | ✓ | above |

---

## Deviations from Plan

- **Mobile `.k-model__top { align-items: stretch }`** (added to plan Step 3.8). Plan's mobile rule switched `display` to flex column but inherited the desktop `align-items: start`; at 375 an icon-only mat shrank to ~196px and the info column to its content (found on CHANEL AS4906). Canvas mobile `modelTop` sets no `align-items` (= stretch), so the fix restores canvas layout. Verified: mat + info 343px.
- `.k-mstat__v` uses `font-variant-numeric: var(--k-num)` (token = `lining-nums tabular-nums`, identical value).
- Brief CSS comments added (mat padding reason; bell outline vs heart fill; mobile stretch) — Source Hygiene rationale comments.

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|---|---|---|
| 2610130722-kura-model-page-plan.md | Source plan, fully executed | Criteria checked, Status → Complete, `[PATCHED]` + Partial Execution line, moved to `.projex/closed/` |
| 2610130524-kura-interactive-mock-sprint.md | Sprint nav | Not edited (orchestrator owns nav) |

---

## Notes

- **For orchestrator (latent, out of scope):** listing detail has the same mobile flex-column `align-items: start` inheritance — a photo-pending listing's `.k-detail__mat` renders 220px of 343 at 375 (e.g. `/listing/brandoff%3A2120100030382`). One-line fix: `align-items: stretch` in the mobile `.k-detail__top` rule.
- Plan Open Questions (A2 crumb, A9 看全部 in place, A7 note + label, A10 0-active state, A11 資料不足 tile, A5 `soldKnown` gate, A4 demand line) remain for the orchestrator to log; nothing decided here.
- `new-projex.sh --type patch --projex-dir .projex/closed` wrote to `.projex/closed/closed/`; file moved to `.projex/closed/` before staging (script appends `closed/` for born-closed types — pass `.projex`).
