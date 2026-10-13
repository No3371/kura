# Audit: Kura model page (型號頁)

> **Status:** Complete (Accepted)
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 3
> **Author:** Claude Opus 5.5 (auditor, sprint subagent)
> **Parent:** 2610130737-kura-model-page-patch.md
> **Related Projex:** 2610130722-kura-model-page-plan.md (audited plan) | 2610130737-kura-model-page-patch.md (execution claims) | 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md (data classes, L4 Day-1 appearance, Canvas gaps) | 2610130624-kura-listing-detail-plan.md (detail mobile layout, F1 origin) | 2610130657-kura-listing-detail-audit.md (F3 malformed-escape precedent) | 2610130757-kura-model-page-audit-fixes-patch.md (F1 F2 F3 H1 fixed)
> **Audit Date:** 2026-10-13 | **Work Period:** 2026-10-13, commits `a97d6a8` (code) `e978aff` (patch doc) on `projex/sprint/2610130524-kura-interactive-mock` (range `9b36213..e978aff`); re-audit 2026-10-13 of fix `b5182c4` + doc `efb0732` (range `1c0fa94..efb0732`)
> **Subject:** plan S1–S9 vs code, tests, live app (worktree `kura.db` served read-only + mutated copy outside repo for sold / days-to-sell / 0-active), full API sweep of all 5,472 models, canvas `project/Main.dc.html` 型號頁 markup + `renderVals` (served version 1791845608-ba84; sprint nav: same project files as pin 1791839497-46ec)

---

## Audit Summary

**Claim:** canvas 型號頁 at `/model/{brand}/{group_key}` on real data — crumb, mat (photo|icon), id block, 2×2 tiles (real | 資料不足), 追蹤這個型號 toggle, model-mode distribution with API `scheme`, 「在售的 {model}」 row capped at 5 with in-place 看全部; model API gains `scheme`; tests green.

**Verdict:** Verified

**Assessment:** Completeness: High | Correctness: High | Quality: High | Value: High

**Top Issues:**
1. F1 (fixable now, out-of-iteration) — listing detail mobile `.k-detail__top` inherits desktop `align-items: start` → photo-less listing mat 220px of 343 at 375 (26 listings). Executor's out-of-scope report holds.
2. H1 (human decision) — `soldKnown` gate is global + count-based: first sweep that marks any listing gone flips all 5,472 models from 資料不足 to `已售出 N 件`, and 平均找到新主人 shows a value from a single sale. Def says real-derived stays 資料不足 "until history suffices" (~2 weeks); def L4 Day-1 appearance still open.
3. F2 (fixable now) — detail and model views share document title shape `{model} · {brand} · Kura 蔵` → identical tab/history titles for a listing and its model.

**Re-audit (fix `b5182c4`):** F1 F2 F3 ✓ fixed; H1 ✓ gate fixed per orchestrator decision (server-owned 14-day span, provisional) — min-sample + per-shop parts stay with def L4. No regressions. H2–H4 = orchestrator carry-overs. → Accepted. See § Re-audit.

---

## Re-audit: fix patch 2610130757-kura-model-page-audit-fixes-patch.md

**Scope:** `git show b5182c4` (`radar.go`, `radar_test.go`, `model.js`, `kura.css`) + `efb0732` audit-doc edits. Evidence: `go vet` clean; `go test -count=1 ./...` ok (advisor radar brandoff store web); `TestSoldStatsReady` PASS; hygiene grep exit 1; built binary served worktree `kura.db` read-only + 3 scratch copies outside repo (deleted after); built-in browser 375 + 1280; console errors none.

| Finding | Fix | Evidence | Status |
|---|---|---|---|
| F1 detail mobile mat | mobile `.k-detail__top` + `align-items: stretch`; shared comment | 375: photo-less `brandoff:2120100030382` align `stretch`, mat 343 / info 343 (was 220 / 342), scrollWidth 375; photo listing 343/343; model AS4906 343/343. 1280: detail grid, `align-items: start`, mat 591 / info 538 (desktop unchanged) | ✓ |
| F2 identical titles | model title `{name} 型號 · {brand} · Kura 蔵` | model `マトラッセ 型號 · CHANEL · Kura 蔵`, `ミニマトラッセ 型號 · …`; detail `ミニマトラッセ · CHANEL · Kura 蔵` unchanged | ✓ |
| F3 rule-6 gaps | comments at `MODEL_ICON`, `trackKey`, 平均 tile | each names the rejected alternative (canvas `role="img"` label → would claim a missing photo; separate store per kind → splits 追蹤 list; relabel → canvas copy decision) — no vacuity, no projex refs | ✓ |
| H1 global count gate | `radar.MinSoldHistory` 14 d; `Engine.latest` = max `FirstSeen|LastSeen|GoneAt`; `SoldStatsReady()` = `latest − epoch ≥ 14 d`; overview `sold_stats_ready`; `model.js` gates 已售出 **and** 平均找到新主人 on it | scratch span checks (2 gone on 027633CC, `median_days_to_sell` 3 in all): span 5 d → `sold_stats_ready false`, tiles 資料不足 ×2 with `sold 2` in API; span 14 d − 1 h → false; span 14 d → true, tiles `已售出 2 件` / `3 天`, AS4906 `已售出 0 件` / 平均 資料不足 (no sale). Real DB: span ~1.3 h → false, `sold 0` → 資料不足 everywhere | ✓ (gate) |
| H1 test | `TestSoldStatsReady` | sold row present in every case (rules out count gate); `14d−1h`/`14d`/`15d` via method + `Overview`; `gone_at` alone extends span; empty engine false. `row()` leaves `LastSeen` zero, so cases discriminate | ✓ |

**[PATCHED] marks by fix agent (efb0732):** F1, F2, F3 earned. H1 earned for the gate; its own text notes "min-sample rule not added" — accurate. F1 condition tick earned.

**Residual (not defects; def L4 owns them):**
- After the gate opens, 平均找到新主人 shows a value from 1–2 sales (scratch: `3 天` from 2) — no min-sample rule.
- Gate is global, not per shop / per model — mixed-onboarding question open. Change point: `Engine.soldStatsReady` only.
- `MinSoldHistory` 14 d = provisional (comment says so; no projex ref in source ✓).

**Regressions:** none — Home 1280/375 renders (scrollWidth ≤ innerWidth), detail crumb → `/model/HERMES/070539CK%2F070544CK`, model M1 tiles / 5 visible + 看全部 409 unchanged, detail 追蹤這件 + per-listing 已找到新主人 untouched (per-listing sold state is fact, correctly ungated). Only Go consumer of sold-derived stats outside radar: advisor `tools.go` (`sell_through`, `median_days_to_sell`) — ungated = H2.

**Source Hygiene (fix diff):** rules 1–5 none; rule 6 — `MinSoldHistory` + `SoldStatsReady` comments name rejected "any listing gone" gate and rejected `now − epoch` span ✓. Commit `b5182c4`: typed `fix(web)` (touches `internal/radar` too — scope label narrow, informational), `Projex: 2610130757-kura-model-page-audit-fixes` ✓.

---

## Claims vs Evidence

| Claim | Evidence | Status | Notes |
|-------|----------|--------|-------|
| Model API returns `scheme` | `server.go` `model`; sweep: 5,472/5,472 responses 200 with `scheme` | ✓ | `items[0]` safe — `radar.Engine.stats` already indexes `g[0]`; groups never empty |
| Tests extended + green | `go vet ./...` clean; `go test ./...` ok (web, radar, store, advisor, brandoff) | ✓ | `TestStaticRoutes` +2, `TestListingAPI` model C1 / `C%2F2` / NOPE |
| Route + decode → 404 | `app.js` route; `/model/HERMES/NOPE`, `/model/NOBRAND/X`, in-app `%E0%A4%A`, `/model/HERMES` (1 seg), trailing `/` → not-found + title | ✓ | Direct-load `%E0%A4%A` → Go 400 (accepted F3 precedent) |
| `/` and `.` keys | direct load `/model/HERMES/070539CK%2F070544CK` + 2 `name:24/24…` keys → 200 shell; in-app crumb href keeps `%2F` | ✓ | 0 keys `.`/`..`, 0 asset-extension, 0 `%?#` keys (sweep) |
| Tiles per plan table | M1 マトラッセ: 409 件 / 資料不足 / ランクA NT$100,919 · 289 件比較 / 資料不足 = API `by_grade`; scratch: 027633CC 3 件 / 2 件 · 自 10/3 / ランクA NT$807,348 · 3 / 6 天; 001855CC 0 件 / 1 件 / no tile 3 / 資料不足 | ✓ | NT$100,919 also on AS4906 — coincidence (both ¥500,000), checked |
| Track toggle shared Set | 070539CK: toggle → 追蹤中 `aria-pressed=true`, bell `fill:none`; card → detail (追蹤這件 unaffected) → crumb back → still 追蹤中; off → 追蹤這個型號 | ✓ | |
| Model-mode distribution | M1 409 dots, page text has no 此件; scratch 027633CC 2 rings + 3 dots; 0-active → rings only | ✓ | |
| 在售 row cap + 看全部 | M1: 409 rendered / 5 visible; click → 409 visible, button removed, focus on card index 5 (`.k-card__model`) | ✓ | `[hidden] { display:none !important }` guarantees hide despite `.k-row > *` |
| 0 active → section absent | scratch 001855CC: no `.k-msale`, mat photo from gone item | ✓ | |
| Responsive / themes | 375 dark AS4906: scrollWidth 375 = innerWidth, mat 343², icon 110px, h1 26px, 資料不足 `rgb(160,148,130)` (ink-3), body `rgb(27,24,21)`; 1280 light M1 screenshot ≈ canvas layout | ✓ | |
| Mobile `align-items: stretch` deviation | canvas mobile `modelTop` has no `align-items` (= stretch); verified mat/info 343 | ✓ | Justified; same gap exists on detail (F1) |
| S8 no sell-through / demand / sample | `model.js` reads only `active sold median_days_to_sell by_grade` + overview `sold since`; page text: no 人在找 / 售出率 / 近 30 天 / 此件 | ✓ | |
| Hygiene grep none | `grep -rnE "[0-9]{10}-…" internal cmd` → exit 1 | ✓ | |
| Detail mobile mat (executor note) | `/listing/brandoff%3A2120100030382` @375: `.k-detail__top` flex, `align-items:start`, mat 220 vs info 342; photo listing → mat 343 | ✓ holds | canvas `detailTop` mobile has no `align-items` → F1 |
| Tab order crumb → track → table → 看全部 → cards | not re-run by keyboard; global `:focus-visible` 2px outline confirmed in CSS; programmatic focus move confirmed | ⚠ | Accepted on patch evidence |

---

## Objective Verification

### S1 Route — ✓ Verified
Click (detail crumb) + direct load, `/` + `.` + Japanese keys; title `{model} · {brand} · Kura 蔵`; unknowns + in-app malformed → not-found. Case-insensitive brand (`/model/hermes/027633CC`) renders canonical; key is case-sensitive (`027633cc` → not-found) — hand-typed only, crumb always emits canonical key.

### S2 Hero — ✓ Verified
Crumb `首頁 / {BRAND} {model}` `aria-current` (A2). Photo = first active then any item with real photo; 5 models icon-only (AS4906, M21096, N41441, N51204, `name:マトラッセミニカメラバッグ`); 11 `NowPrinting` stats images bypassed. No demand line. Icon `aria-hidden` vs canvas `role="img" aria-label="型號照片"` — plan A3 decision, uncommented (Source Hygiene).

### S3 Tiles — ✓ Verified
Order + values = API. Real DB: 0 `sold > 0`, 0 `median_days_to_sell` → 已售出 + 平均找到新主人 資料不足 on every model. 1 no-label model → 3 tiles. Tie rule strict `>` → better grade.

### S4 Track — ✓ | S5 Distribution — ✓ | S6 在售 row — ✓ | S7 Responsive/themes — ✓ (1280 light, 375 dark checked; patch covers rest) | S8 One source — ✓ | S9 vet/test/hygiene — ✓

---

## Code/Implementation Inspection

### `internal/web/server.go` `model`
**Actual:** `scheme` from `items[0].Source`; comment names the multi-shop alternative. Quality High.

### `internal/web/static/js/model.js` (new, 112 lines)
**Actual:** matches plan Step 4 markup + tile table. All interpolation via `esc`; figures via `num`/`twd`; no client currency math. `dl > div > dt + dd + dd` valid. Quality High.
**Issues:** none functional. Rule 6 gaps below.

### `internal/web/static/js/card.js`
**Actual:** off label from `data-track-text` (empty → 追蹤這件); detail button unchanged (verified 追蹤這件). Quality High.

### `internal/web/static/kura.css`
**Actual:** model block = canvas values; selector extensions for h1; mobile stretch. Quality High.
**Undocumented:** none beyond patch Deviations.

**Metrics:** Readability High | Technical Debt Low (409-card DOM on largest model, plan-acknowledged until Browse)

---

## Testing Validation

**Execution:** all pass, no flakes observed.
**Coverage:** server — scheme, escaped `/`, 404. Client view untested by automation (no JS test harness in repo; browser matrix is the evidence).
**Missing:** `scheme` absent branch (unknown source) not tested — trivial.

---

## Documentation Audit

Patch doc accurate vs code + live app; every M-row re-checked except keyboard Tab order. Plan Status Complete, moved to `closed/`, `[PATCHED]` + Partial Execution line present. Patch notes `new-projex.sh --projex-dir .projex/closed` → `closed/closed/` quirk; no stray `closed/closed/` remains.

---

## Source Hygiene

**Scope:** Diff only (`git show a97d6a8 -- . ':(exclude,glob)**/.projex/**'`)

| Rule | Location (symbol) | Quote | Severity |
|------|-------------------|-------|----------|
| — | none (rules 1–5) | — | — |

**Rule 6 gaps** (non-obvious change, no rationale comment):
- `model.js` → `MODEL_ICON` — `aria-hidden` while canvas gives the icon `role="img" aria-label="型號照片"`; a reader syncing to canvas would revert it. Low.
- `model.js` → `tilesHTML` 平均找到新主人 tile — label says 平均 (mean), value is `median_days_to_sell`; known canvas-copy mismatch, not marked. Low.
- `model.js` → `trackKey` — `model:` prefix relies on no source being named `model` (ids are `{source}:{id}`). Low.

**Rule 6 vacuity:** none — every rationale comment names its alternative or constraint (Browse link → not-found; per-render listener stacking; time-based rule replacing `soldKnown`; stats `image_url` placeholder; desktop `align-items` shrink).

**Rationale promotion:** none needed — load-bearing rationale (A2 crumb, A4 demand, A5 gate, A9 in-place reveal, A3 photo choice) lives in `model.js` comments; deviations for canvas pass 2 belong in sprint nav / def, not shipped docs.

**Commit composition** (informational): 1 code commit — typed subject 1/1 | type matches diff 1/1 | `Projex:` trailer 1/1. Doc commit `e978aff` exempt.
**Violations:** None. Trailer value `2610130737-kura-model-page` omits `-patch` (iteration 2 `b12c30a` uses `…-patch`, `1dab0e9` omits) — inconsistent naming, informational.

**Trailer survival on `master`:** 0/6 — master holds only baseline import + doc commits (sprint unmerged); sprint branch code commits 7/7 carry `Projex:`. Channel alive; not Critical.

---

## Gap Analysis

### Promised But Not Delivered
None.

### Undocumented Issues
| Issue | Severity | Affects |
|-------|----------|---------|
| F1 detail mobile mat shrink (photo-less listings) | Minor–Significant | listing detail @ ≤767px, 26 of 15,266 grouped listings |
| F2 detail + model identical title shape | Minor | tab/history/SR page announcement |
| H1 global `soldKnown` flips on day 2 | Significant (decision) | every model's 已售出 + 平均找到新主人 |

### Unhandled Edge Cases
- `ランクX 中位數 · 1 件比較` on single-listing models (3,737 models) — "median" of one, truthful but thin; copy question (H4).
- Overview cached for the session (`overviewP` never refreshes) → `soldKnown` + `since` stale in a long-lived tab. Pre-existing, Low.

---

## Quality Assessment

### Completeness: High
All S1–S9 met; matrix rows reproduced.

### Correctness: High
No bugs in iteration-3 code. F1 is iteration-2 code.

### Code Quality: High
Reuses card/dist/detail classes; one listener per concern; comments carry rationale.
**Tech Debt:** full-group card DOM (409 on largest) — Low, Browse supersedes.

### Value Delivered: High
Detail crumb dead end closed; first real 資料不足 surface; model page demonstrably honest on short history.

---

## Open Findings

### Impact Analysis
- Downstream: advisor `model_price_stats` still returns `sold_observed: 0` / `sell_through: 0` where UI says 資料不足 (H2).
- Future: Browse iteration swaps 看全部 button → link (A9).
- Risk: day-2 demo shows `已售出 0 件` / one-sale 平均 on most models (H1) — Likelihood High once a daily sweep runs.

---

## Findings

### Critical (Must Address)
None.

### Significant (Should Address)
- **H1 — `soldKnown` gate global + count-based** (needs human decision) → `[PATCHED]` 2610130757-kura-model-page-audit-fixes-patch.md (orchestrator call: server-owned 14-day history-span gate, provisional pending def L4; min-sample rule not added) — `soldKnown = ov.overview.sold > 0`; scratch DB with 3 gone listings index-wide: unrelated models flip to `已售出 0 件 · 自 10/3 起記錄`, 027633CC shows `平均找到新主人 6 天` from 2 sales. Def § content kinds: real-derived → 資料不足 "until history suffices" (~2 weeks); def L4 Day-1 appearance open. Plan A5 flagged this as factual minimum only. Not visible on today's DB (0 sold). → Human answers L4 (e.g. N days since `since` and/or min sales per model); replace `soldKnown` + add a min-sample rule for days-to-sell at one site.
- **F1 — detail mobile mat shrink** (fixable now; iteration-2 latent, executor report verified) → `[PATCHED]` 2610130757-kura-model-page-audit-fixes-patch.md — `kura.css` mobile `.k-detail__top { display:flex; flex-direction:column; gap:24px }` inherits desktop `align-items:start`; photo-less listing @375 → mat 220×220 vs info 342; canvas `detailTop` mobile omits `align-items` (= stretch). → add `align-items: stretch` to the mobile `.k-detail__top` rule (same fix + comment as `.k-model__top`).

### Minor (Nice to Fix)
- **F2 — identical titles** (fixable now; orchestrator copy call) → `[PATCHED]` 2610130757-kura-model-page-audit-fixes-patch.md — detail `${name} · ${it.brand} · Kura 蔵` = model `${name} · ${s.brand} · Kura 蔵`; listing and its model indistinguishable in tabs/history/SR. Both plan-specified. → e.g. model `${name} 型號 · …` (canvas has no titles; no canvas conflict).
- **F3 — Rule 6 gaps** (fixable now) → `[PATCHED]` 2610130757-kura-model-page-audit-fixes-patch.md → one-line comments at `MODEL_ICON` (decorative: mat is not the model's identity image; name is in h1), 平均找到新主人 tile (label kept verbatim from canvas; value is a median), `trackKey` (`model:` cannot collide while no source is named `model`).
- **H2 — advisor vs UI on sold stats** (advisor iteration) — def "one source per stat"; advisor returns 0 where page says 資料不足. → apply the same gate in advisor tool output when advisor UI is built; orchestrator logs.
- **H3 — deviations to log** (orchestrator, no code) — A2 crumb, A3 icon a11y + mat padding, A7 「自 m/d 起記錄」 note + 平均/median label, A9 in-place 看全部, A10 0-active state, A11 資料不足 tile (def § Canvas gaps candidate) → sprint nav / def per plan Open Questions.
- **H4 — `1 件比較` median** (copy, canvas pass 2) — single-listing models label their only price a 中位數. → canvas pass 2 decides threshold or wording.

### Positive
- API sweep: all 5,472 models 200 with `scheme`; plan data probes (433 > 5, 11 NowPrinting, 1 unlabelled, 12 `/` keys) reproduce exactly.
- Executor caught and fixed the mobile stretch gap from canvas evidence and correctly flagged the same latent defect on detail.
- Comments name rejected alternatives throughout; no projex IDs in source.

---

## Recommendations

**Immediate:** fix patch for F1 (+ F2, F3 at orchestrator's discretion) — `kura.css`, `model.js` (+ `detail.js` if F2 touches detail).
**Future:** H1 answer before any demo past the first daily sweep; H2 in advisor iteration.
**Process:** layout rules switching grid → flex column should restate `align-items` (both detail and model hit this).

---

## Final Verdict

**Status:** Accept (was Accept with Conditions; conditions met on re-audit)

**Overall Assessment:** Completeness High | Correctness High | Quality High | Value High

**Conditions:**
- [x] F1 fixed (mobile `.k-detail__top` stretch) and verified at 375 on a photo-less listing — 2610130757-kura-model-page-audit-fixes-patch.md
- [x] H1 resolved by orchestrator decision (provisional 14-day server gate, `b5182c4`); re-audited ✓ — min-sample / per-shop residual stays with def L4
- Carry-overs (orchestrator-owned, not audit conditions): H2 advisor gate → advisor iteration (reuse `Engine.SoldStatsReady`) | H3 deviations → sprint nav / def § Canvas gaps | H4 `1 件比較` median → canvas pass 2 | H1 residual → def L4

**Sign-off:** Yes — iteration-3 work meets S1–S9; fix patch `b5182c4` closes F1 F2 F3 + H1 gate with no regressions. Final status: Accept.
