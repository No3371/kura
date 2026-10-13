# Kura model page (型號頁)

> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 3
> **Status:** Complete
> **Author:** Claude Opus 5.5 (planner, sprint subagent)
> **Parent:** 2610130524-kura-interactive-mock-sprint.md
> **Source:** Sprint derive, iteration 3 (sprint nav § Current Position + Revision Log, commit `a05b9ee`)
> **Related Projex:** 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md (data classes, canvas precedence, Canvas gaps, L4 Day-1 appearance) | 2610130420-kura-roadmap-nav.md (Phase 3 New IA) | 2610130624-kura-listing-detail-plan.md (A1 model URL, `dist.js`, track toggle this builds on) | 2610130535-kura-canvas-foundation-home-plan.md (shell, card, `.k-row`; A1 大家都在找 hidden) | 2610130737-kura-model-page-patch.md (execution)
> Partial Execution: Steps 1–5 (all) completed via 2610130737-kura-model-page-patch.md — `[PATCHED]`, nothing remains. S1 malformed-escape direct load (`%E0%A4%A`) → Go 400, F3 precedent. Deviation: mobile `.k-model__top` gains `align-items: stretch` (Step 3.8 gap).
> **Worktree:** Yes — existing sprint worktree `kura/.projexwt/2610130524-kura-interactive-mock`, branch `projex/sprint/2610130524-kura-interactive-mock`; execute via patch-projex plan-execution mode, no new branch

---

## Summary

Build the canvas 型號頁 (pinned 1791839497-46ec, artboards `Model` desktop light 1280 + `ModelMobile` light 375; markup `project/Main.dc.html` `page: model`) at `/model/{brand}/{group_key}` on real BRAND OFF data: crumb, model photo mat, brand/name/類型 · code, 4 stat tiles (real or 資料不足), 追蹤這個型號 toggle, shared 同型號價格分布 (`dist.js`, model mode), 「在售的 {model}」 card row. Closes the detail crumb dead end; first screen showing the real-derived 資料不足 state the sprint goal names.

**Scope:** `internal/web` only — model API gains `scheme` (+ test), router route, new `model.js`, `card.js` track-label hook, `kura.css` model block. One `.projex` scope.
**Estimated Changes:** 6 files — `server.go`, `server_test.go`, `app.js`, `card.js`, `kura.css` edited; `model.js` added.

---

## Objective

### Problem / Gap / Need
Detail crumb links `/model/{brand}/{group_key}` (iteration 2 A1) → not-found view. Canvas model page is the next unblocked flow (sprint nav blocking map: Browse waits on F1; Model unblocked). Model API (`/api/models/{brand}/{key}`) lacks the grade `scheme` the distribution ladder needs.

### Success Criteria
- [x] **S1 Route:** `/model/{brand}/{group_key}` (both segments `encodeURIComponent`-ed) renders the model view via in-app click (detail crumb) and direct load/reload, incl. keys with `/` (`name:24/24ヴァンキャトル29`), `.` (`name:2.55マトラッセ`) and Japanese; title `{model} · {brand} · Kura 蔵`; unknown brand/key or malformed encoding → not-found view + 「找不到這個頁面 · Kura 蔵」.
- [x] **S2 Hero** matches canvas values (Step 3/4): crumb `首頁 / {BRAND} {model}` (A2); square mat with a real listing photo, else the canvas bag icon (A3); `{BRAND} {中文}`, h1 model name, `{類型} · {code}` (parts present only); no demand line (A4).
- [x] **S3 Stat tiles** (canvas order, 2×2): 在售 `{N} 件` · 各等級合計 | 已售出 `{N} 件` or 資料不足 · `自 {m/d} 起記錄` (A5) | `{grade} 中位數` `NT$…` · `{n} 件比較` (A6) | 平均找到新主人 `{d} 天` or 資料不足 · `自 {m/d} 起記錄` (A7). On the worktree DB: 已售出 + 平均找到新主人 read 資料不足 on every model.
- [x] **S4 Track:** 追蹤這個型號 ↔ 追蹤中 toggle (bell icon, `aria-pressed`) + 「有新上架或低於行情時通知你」; state in the same in-memory set as listing tracks, survives in-app navigation (A8).
- [x] **S5 Distribution:** `distHTML` in model mode (`thisId` null: no 此件 legend/label/highlight) with the shop ladder from the model API `scheme`; dots = group items, tick/table = `by_grade` `*_twd`; section absent when no listing has a grade.
- [x] **S6 在售 row:** 「在售的 {model}」 + up to 5 active cards (`cardHTML`, API order = price ↑); when active > 5, 「看全部 {N} 件 →」 reveals the rest in place and moves focus to the 6th card (A9); no link when ≤ 5; section absent when 0 active (A10).
- [x] **S7 Responsive/themes:** 1280 + 375 × light/dark; no horizontal page scroll at 375 (row scrolls in its own box); targets ≥ 24px; visible focus ring.
- [x] **S8 One source per stat:** every figure from server fields (`active`, `sold`, `median_days_to_sell`, `by_grade[].median_twd/count`, overview `since`/`sold`); page does no JPY→TWD math; no `sell_through`, demand or sample (`KURA_MODEL_STATS`) value rendered.
- [x] **S9** `go vet ./...` + `go test ./...` green (extended tests Step 1); no projex references in source comments.

### Out of Scope
- Browse (F1) — 「看全部」 does not link there until it exists (A9)
- 「這週 N 人在找」 demand line and crumb 「大家都在找」 (open question 大家都在找; not decided here)
- Member header / notification feed / persisted tracking (member-entry open question); advisor; footer pages incl. 行情怎麼算
- `sell_through` display (not designed on model tiles); new radar stats or windows (e.g. 30-day days-to-sell)
- Entry points to the model page beyond the detail crumb (Home 熱門型號/大家都在找 not built)
- Advisor wording for `sold_observed` / `sell_through` while the UI shows 資料不足 (advisor iteration — Notes)

---

## Context

### Current State
- `GET /api/models/{brand}/{key}` (`server.go` `model`) → `{model: ModelStats, items}`; brand canonicalised case-insensitively (`canonicalBrand`); unknown → 404 `{"error":"model not found"}`. Items: whole group, active first then price ↑ (`radar.Engine.Model`). No `scheme` (detail's `listing` handler adds it via `listing.SchemeFor(it.Source)`).
- `ModelStats` (`radar.go`): `brand group_key model_name model_code category image_url active sold median_jpy min_jpy by_grade[] sell_through median_days_to_sell?`. `by_grade` sorted best grade first, each `{source, grade{code,label,short,desc_*,rank}, count, median/min/max _jpy/_twd}` over active listings. `MedianDaysToSell` nil unless a gone item was first seen > epoch + 24h. `image_url` = first-inserted item's (may be the `NowPrinting` placeholder). No `brand_zh`/`category_zh`.
- `/api/overview` → `overview.{sold, since}` (since = earliest `first_seen`), `brands[{name,zh}]`, `categories[{key,zh}]` — the router passes this as `ov` to every view.
- Client: `app.js` routes `(el, ov, ...encodedParams)` → may return title; 404 status → not-found. `detail.js` crumb already links `/model/${enc brand}/${enc group_key}`. `dist.js` `distHTML({stats, items, scheme, thisId = null})`. `card.js`: `cardHTML`, `isTracked`, track Set + document listener (text buttons: `[data-track-text]` set to 追蹤中 / 追蹤這件). `home.js` `.k-row` grid (5 cols desktop, 3 at 768–1023, carousel ≤ 767).
- Data (worktree `kura.db`, plan-time probe 2026-10-13 via local serve): 15,367 active, 0 sold; overview `since` 2026-10-13T00:25+08:00; 5,472 models — 3,737 with 1 active, 531 ≥ 5, 433 > 5, max 409 (CHANEL `name:マトラッセ`, response 477 KB); 2,281 `name:` keys (no code); 12 keys contain `/` (API + SPA reload both 200 with `%2F`), 109 contain `.`, none end in an asset extension; 11 models' `image_url` is `NowPrinting`; 1 model with no labelled grade; 304 models tie on the largest grade count; `median_days_to_sell` set on 0, `sold > 0` on 0.
- Canvas model (`Main.dc.html` `C · 型號頁` + `renderVals`): crumb `首頁 / 大家都在找` (span, not link) → `modelTop` (desktop `grid 360px | 1fr; gap 48px; align start; padding-bottom 56px`; mobile `flex column; gap 24px; padding-bottom 40px`): mat (`aspect 1/1; mat bg; 1px rule border; inset 0 0 0 8px mat, inset 0 0 0 9px mat-line`; bag icon 140 / 110px) | info (`flex column; gap 18px`): id block (as detail; h1 32 / 26px), demand `<p>` (hand font 19px accent-ink), `dl` tiles (`grid 2 cols; border-top 1px rule`; tile `column; gap 4px; padding 14px 0; border-bottom 1px rule`; dt 12px ink-3; dd font-num 22px lining tabular; note 12px ink-3), track row (`flex center; gap 14px; wrap`; button `height 44px; padding 0 18px; gap 8px; 14px; border 1px; radius 2px`, bell 17px stroke 1.5, pressed = accent-ink / accent-wash / accent border, off = ink / ink-2 border; note 12px ink-3) → shared distribution (`max-width 880px; padding 0 0 48px`) → 在售 section (head `flex wrap space-between baseline; gap 6px 16px; padding-top 24px; border-top 1px brass; margin-bottom 22px`; h2 font-display 400 22px 「在售的 <span ja>{model}</span>」; link 13px ink-2 「看全部 18 件 →」 → Browse; row = Home `freshRow`). Sample block comment: 「示意數據 — sample values for layout only; real values need ~2 weeks of history」.

### Key Files

| File | Role | Change Summary |
|------|------|----------------|
| `internal/web/server.go` | API | model response gains `scheme` |
| `internal/web/server_test.go` | tests | model API + `/model/…` SPA cases |
| `internal/web/static/js/app.js` | router | `/model/{brand}/{key}` route |
| `internal/web/static/js/card.js` | track listener | off-label from `data-track-text` value |
| `internal/web/static/kura.css` | styles | model hero, tiles, track variant, 在售 head, more button, mobile |
| `internal/web/static/js/model.js` | new | model view + 看全部 listener |

### Dependencies
- **Requires:** sprint branch HEAD `a05b9ee` (iterations 1–2: shell, router, card, `dist.js`, detail crumb URL).
- **Blocks:** nothing in-sprint; Browse iteration later swaps 看全部 to a link (A9).

### Constraints
- HANDOFF §3.6 no build step / framework → native ES module; §3.7 a11y (≥ 24px targets, focus, table view — `dist.js` provides).
- Canvas governs look/IA/copy/states; def governs data semantics (real | real-derived | mocked). Model artboards are `user: member`; guest header stays (iteration 2 A2).
- Real-derived stats → value or 資料不足, never sample; mocked demand → not rendered (sprint open question 大家都在找).
- Client never converts JPY → TWD. Source Hygiene: no projex IDs/filenames in comments; non-obvious decisions name the rejected alternative.
- No crawls; commits on the sprint branch in the worktree only.

### Assumptions
- **A1 — URL** = iteration 2 A1 `/model/{encodeURIComponent(brand)}/{encodeURIComponent(group_key)}`; detail crumb unchanged. Router hands encoded params; view decodes (bad escape → 404).
- **A2 — Crumb** `首頁 / {BRAND} {model}` (current page, `aria-current="page"`, detail crumb vocabulary) replaces canvas `首頁 / 大家都在找`: that shelf is hidden (open question) and most visits arrive from a detail page, so naming the shelf would state a false path. Restore when the shelf ships. Orchestrator: log as canvas deviation.
- **A3 — Model photo** = first item (active first) with a real photo (`!photo_pending && image_url`) — not `ModelStats.image_url`, which is the first-inserted item's and can be `NowPrinting` (11 models). None → canvas bag icon, decorative (`aria-hidden`). Photo alt `{model} 型號照片`. Mat gets detail's 10% padding so a photo clears the inset frame (canvas draws only the icon).
- **A4 — Demand line** 「這週 N 人在找」 not rendered (def: mocked demand; sprint open question 大家都在找 governs it). Nothing replaces it.
- **A5 — 已售出 gate:** value shown only when sold detection has observed a sale anywhere (`ov.overview.sold > 0`); else 資料不足. Rationale: with 0 gone across the index no sweep has marked anything sold, so a per-model 0 would be unknown, not zero; the gate is the factual precondition only — it does **not** answer def L4 「Day-1 appearance」 (~2 weeks rule, open). One named function (`soldKnown`) so that answer drops in at one place. Note copy `自 {md(since)} 起記錄` (canvas; real `since`).
- **A6 — Median tile grade** = the labelled `by_grade` entry with the most in-sale listings; ties → better grade (first in `by_grade` order). Label `{grade.label} 中位數`, value `twd(median_twd)`, note `{count} 件比較` (canvas). No labelled grade (1 model) → tile omitted (3 tiles).
- **A7 — 平均找到新主人** value = `median_days_to_sell` (radar median; canvas label says 平均 — label kept verbatim, mismatch flagged for canvas pass 2 copy review). Shown `{round} 天` (`< 1` → 「不到 1 天」, plan-authored) when non-null, else 資料不足. Note `自 {md(since)} 起記錄` replaces canvas 「近 30 天」 — radar does not window by 30 days. Orchestrator: log as copy deviation.
- **A8 — Model track** reuses `card.js` Set with key `model:{brand}/{group_key}` (listing ids are `{source}:{id}`; no source is named `model`). In-memory only, guest-visible like detail's 追蹤這件 — no member-entry decision.
- **A9 — 看全部 in place:** canvas links 看全部 → Browse (unbuilt, F1). Until then the 5-card row renders every active card, those past 5 `hidden`; 「看全部 {N} 件 →」 is a `<button>` styled as the canvas link that unhides them (desktop grid wraps; ≤ 767 carousel lengthens). Hidden cards' lazy images do not load. Orchestrator: log as deviation; Browse iteration turns it into a link.
- **A10 — 0 active** (no real instance; scratch DB): 在售 `0 件`, median tile omitted, 在售 section omitted, distribution shows rings only. Plan-authored state.
- **A11 — 資料不足 tile** = same tile, dd text 資料不足 in `--k-ink-3` (HANDOFF §3.2 "Not enough data: quiet"); no new visual. Orchestrator: candidate for def § Canvas gaps (HANDOFF §4.C asked for it; canvas draws samples only).
- **A12 — `scheme` per group** = `SchemeFor(items[0].Source)` — one shop per group while a single provider is ingested (multi-provider ladders = sprint Deferred).

### Impact Analysis
- **Direct:** new `/model/…` view; model API response adds `scheme`; card track listener reads an optional off-label.
- **Adjacent:** detail page also fetches the model API (ignores the new field); detail/card track buttons (unchanged text: empty `data-track-text` → 追蹤這件); advisor untouched (`modelStats` reads `radar.Model`, not the handler).
- **Downstream:** none outside the repo.

---

## Implementation

`[PATCHED]` Steps 1–5 → 2610130737-kura-model-page-patch.md (code `a97d6a8`).

### Overview
Server field + tests, then client: router line, card hook, CSS, view; then the browser matrix incl. a scratch DB copy for the sold / days-to-sell / 0-active states real data lacks.

### Step 1: Model API returns the grade scheme

**Objective:** S5 ladder data, S1 404/escaped-key coverage. **Confidence:** High. **Depends on:** None.
**Do-Projex:** Encouraged

**Files:** `internal/web/server.go`, `internal/web/server_test.go`.

**Changes:**

```go
// Before (model):
	return map[string]any{"model": stats, "items": items}, 0, nil
// After:
	res := map[string]any{"model": stats, "items": items}
	// A group holds one shop's listings while a single provider is ingested,
	// so its first item names the ladder. With several shops the distribution
	// needs one ladder per shop, not this single field.
	if sc, ok := listing.SchemeFor(items[0].Source); ok {
		res["scheme"] = sc
	}
	return res, 0, nil
```

`server_test.go`:
- `TestStaticRoutes` cases += `{"/model/CHANEL/name%3A%E3%83%9E%E3%83%88%E3%83%A9%E3%83%83%E3%82%BB", "text/html", `id="main"`, 200}`, `{"/model/HERMES/name%3A24%2F24", "text/html", `id="main"`, 200}`.
- `TestListingAPI`: after the existing `RecordSeen`, record a second listing `SourceID: "2"`, `ModelCode: "c/2"` (else as `l`). After the listing assertions:
  - `GET /api/models/HERMES/C1` → 200; decode `{Model radar.ModelStats; Items []radar.Item; Scheme *listing.Scheme}` → `Scheme.Shop == "テスト店"`, `len(Items) == 1`, `Model.Active == 1`.
  - `GET /api/models/HERMES/C%2F2` → 200, `Items[0].ID == "wt:2"` (escaped `/` stays inside the key).
  - `GET /api/models/HERMES/NOPE` → 404.

**Rationale:** mirrors detail's `listing` handler; the client never guesses a ladder.

**Verification:** `go vet ./...`; `go test ./internal/web/`.

**If this fails:** revert both files.

---

### Step 2: Router route + card track label hook

**Objective:** S1, S4 plumbing. **Confidence:** High. **Depends on:** None (Step 4 provides `model.js`).
**Do-Projex:** Encouraged

**Files:** `internal/web/static/js/app.js`, `internal/web/static/js/card.js`.

**Changes:**
- `app.js`: `import { renderModel } from "./model.js";` routes += `[/^\/model\/([^/]+)\/([^/]+)$/, renderModel, "型號頁 · Kura 蔵"]`.
- `card.js` listener:
  ```js
  // Before:
        text.textContent = on ? "追蹤中" : "追蹤這件";
  // After:
        text.textContent = on ? "追蹤中" : text.dataset.trackText || "追蹤這件";
  ```
  Update the listener comment: a text button's off label comes from its `data-track-text` value (empty → 追蹤這件), so the model page's 追蹤這個型號 shares this one listener.

**Rationale:** one tracked Set + one listener (iteration 2 pattern) instead of a second toggle mechanism.

**Verification:** with Step 4: detail 追蹤這件 still flips 追蹤這件 ↔ 追蹤中; model button flips 追蹤這個型號 ↔ 追蹤中.

**If this fails:** revert both files.

---

### Step 3: `kura.css` — model page

**Objective:** S2–S7 look. **Confidence:** High. **Depends on:** None.

**Files:** `internal/web/static/kura.css`.

**Changes** (canvas `Main.dc.html` model markup + `renderVals` `modelTop`, `modelIcon`, `detailH1`, `crumbPad`, `freshRow`). Insert a `/* 型號頁 */` block after the 售價變化 block:
1. `.k-model__top { display:grid; grid-template-columns:360px minmax(0,1fr); gap:48px; align-items:start; padding-bottom:56px; }`
2. `.k-model__mat { aspect-ratio:1/1; box-sizing:border-box; padding:10%; background:var(--k-mat); border:1px solid var(--k-rule); box-shadow:inset 0 0 0 8px var(--k-mat), inset 0 0 0 9px var(--k-mat-line); display:flex; align-items:center; justify-content:center; color:var(--k-mat-ink); }` · `.k-model__mat img { width:100%; height:100%; object-fit:contain; mix-blend-mode:multiply; }` · `.k-model__mat svg { width:140px; height:140px; }`
3. `.k-model__info { display:flex; flex-direction:column; gap:18px; min-width:0; }`; extend the existing `.k-detail__info h1` selector to `.k-detail__info h1, .k-model__info h1` (same type, `overflow-wrap:anywhere` for the 70-char names).
4. Tiles: `.k-mstats { margin:0; display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); border-top:1px solid var(--k-rule); }` · `.k-mstats > div { display:flex; flex-direction:column; gap:4px; padding:14px 0; border-bottom:1px solid var(--k-rule); min-width:0; }` · `.k-mstats dt { font-size:12px; color:var(--k-ink-3); }` · `.k-mstats dd { margin:0; }` · `.k-mstat__v { font-family:var(--k-font-num); font-size:22px; font-variant-numeric:lining-nums tabular-nums; }` · `.k-mstat__v.is-na { color:var(--k-ink-3); }` · `.k-mstat__note { font-size:12px; color:var(--k-ink-3); }`
5. Track: `.k-mtrack { display:flex; align-items:center; gap:14px; flex-wrap:wrap; }` · `.k-mtrack > span { font-size:12px; color:var(--k-ink-3); }` · `.k-dtrack--model { height:44px; font-size:14px; }` · `.k-dtrack--model svg { width:17px; height:17px; }` · `.k-dtrack--model[aria-pressed="true"] svg { fill:none; }` (canvas bell stays outlined; detail heart fills).
6. `.k-model__dist { max-width:880px; }`
7. 在售: `.k-msale__head { display:flex; flex-wrap:wrap; justify-content:space-between; align-items:baseline; gap:6px 16px; padding-top:24px; border-top:1px solid var(--k-brass); margin-bottom:22px; }` · `.k-msale__head h2 { margin:0; font-family:var(--k-font-display); font-weight:400; font-size:22px; }` · `.k-msale__head h2 span { font-family:var(--k-font-ja); }` · `.k-more { font:inherit; font-size:13px; background:none; border:0; padding:0; cursor:pointer; }` (paired with `.k-sec__more` for color / min-height).
8. Mobile (inside existing `@media (max-width: 767px)`): `.k-model__top { display:flex; flex-direction:column; gap:24px; padding-bottom:40px; }` · extend `.k-detail__info h1` → `.k-detail__info h1, .k-model__info h1 { font-size:26px; }` · `.k-model__mat svg { width:110px; height:110px; }`.

**Rationale:** transcribed canvas values; reuses `.k-crumb`, `.k-detail__id/__brand/__sub`, `.k-dtrack`, `.k-sec__more`, `.k-row`, `.k-dist*` instead of duplicates.

**Verification:** visual, Step 5.

**If this fails:** remove the block and the two selector extensions.

---

### Step 4: Model view

**Objective:** S1–S6, S8. **Confidence:** High. **Depends on:** Steps 1–3.
**Verify-Projex:** Encouraged

**Files:** `internal/web/static/js/model.js` (new).

**Changes:** `export async function renderModel(main, ov, rawBrand, rawKey)`; imports `api, esc, num, twd, md` (util), `cardHTML, isTracked` (card), `distHTML` (dist).

1. Decode both params in one `try` (`decodeURIComponent`); `catch` → `throw Object.assign(new Error("bad model path"), { status: 404 })`.
2. `const { model: s, items, scheme } = await api(\`/api/models/${encodeURIComponent(brand)}/${encodeURIComponent(key)}\`)` (404 → router not-found).
3. `name = s.model_name || s.model_code || key`; `brandZH = ov.brands?.find((b) => b.name === s.brand)?.zh`; `catZH = ov.categories?.find((c) => c.key === s.category)?.zh`; `active = items.filter((it) => it.status === "active")`; `hasPhoto = (it) => !it.photo_pending && it.image_url`; `photo = active.find(hasPhoto) ?? items.find(hasPhoto)`.
4. `main.innerHTML` (all interpolation via `esc`; figures via `num`/`twd`), return `` `${name} · ${s.brand} · Kura 蔵` ``.

**Markup (canvas order):**
```html
<div class="k-model">
  <nav aria-label="位置" class="k-crumb"><a href="/" class="k-link">首頁</a><span aria-hidden="true">/</span><span class="k-crumb__model" aria-current="page">{brand} {name}</span></nav>
  <div class="k-model__top">
    <div class="k-model__mat">{photo ? <img src="{photo.image_url}" alt="{name} 型號照片"> : MODEL_ICON}</div>
    <div class="k-model__info">
      <div class="k-detail__id">
        <span class="k-detail__brand">{brand}{brandZH ? ` <span>{brandZH}</span>` : ""}</span>
        <h1>{name}</h1>
        {sub ? <span class="k-detail__sub">{sub}</span>}   <!-- sub = [catZH, s.model_code].filter(Boolean).join(" · ") -->
      </div>
      <dl class="k-mstats">{tilesHTML}</dl>
      <div class="k-mtrack">
        <button type="button" class="k-dtrack k-dtrack--model" data-track="model:{brand}/{group_key}" aria-pressed="{on}">{BELL}<span data-track-text="追蹤這個型號">{on ? "追蹤中" : "追蹤這個型號"}</span></button>
        <span>有新上架或低於行情時通知你</span>
      </div>
    </div>
  </div>
  {dist ? <div class="k-model__dist">{dist}</div>}      <!-- dist = distHTML({ stats: s, items, scheme }) -->
  {active.length ? 在售 section}
</div>
```
- `MODEL_ICON` = canvas bag: `<svg aria-hidden="true" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="0.9"><path d="M12 24h40l-3 30H15z"/><path d="M23 24v-5a9 9 0 0 1 18 0v5"/><path d="M28 31h8v5h-8z"/></svg>` (size from CSS).
- `BELL` = canvas: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" aria-hidden="true"><path d="M6 16.5V11a6 6 0 0 1 12 0v5.5l1.5 2H4.5z"/><path d="M10 20.5a2 2 0 0 0 4 0"/></svg>`.
- `on = isTracked(trackKey)`, `trackKey = \`model:${s.brand}/${s.group_key}\`` (A8).

**`tilesHTML(s, ov)`** — tile = `<div><dt>{k}</dt><dd class="k-mstat__v{na ? " is-na" : ""}">{v}</dd><dd class="k-mstat__note">{note}</dd></div>` (note as a second `dd`: a `span` is not valid inside `dl > div`). `rec = \`自 ${md(ov.overview.since)} 起記錄\``.

| Tile | Condition | k | v | note |
|---|---|---|---|---|
| 1 | always | 在售 | `{num(s.active)} 件` | 各等級合計 |
| 2 | `soldKnown(ov)` | 已售出 | `{num(s.sold)} 件` | rec |
| 2 | else | 已售出 | 資料不足 (`is-na`) | rec |
| 3 | `g` exists | `{g.grade.label} 中位數` | `twd(g.median_twd)` | `{num(g.count)} 件比較` |
| 4 | `s.median_days_to_sell != null` | 平均找到新主人 | `< 1` → 不到 1 天, else `{Math.round(d)} 天` | rec |
| 4 | else | 平均找到新主人 | 資料不足 (`is-na`) | rec |

`g` = `(s.by_grade ?? []).filter((x) => x.grade?.label).reduce((best, x) => (!best || x.count > best.count ? x : best), null)` (strict `>` keeps the earlier, better grade on ties — A6). `soldKnown = (ov) => ov.overview.sold > 0` with a comment: sold counts exist only once a sweep has marked a listing gone; before that a model's 0 is unknown, not zero — showing it would read as "never sells"; a time-based rule would replace this one function (A5).

**在售 section:**
```html
<section class="k-msale" aria-labelledby="ml-h">
  <div class="k-msale__head"><h2 id="ml-h">在售的 <span>{name}</span></h2>
    {active.length > CAP ? <button type="button" class="k-sec__more k-link k-more" data-model-more>看全部 {num(s.active)} 件 →</button>}</div>
  <div role="list" class="k-row">{active.map((it, i) => <div role="listitem"{i >= CAP ? " hidden" : ""}>{cardHTML(it)}</div>)}</div>
</section>
```
`const CAP = 5; // one desktop row of cards, as the canvas draws 在售的 {model}`.

Module comment (top): view at `/model/{brand}/{group_key}`; every figure from the server; 「這週 N 人在找」 not rendered — demand has no real source and sample values on the page would pass for data (as Home's 大家都在找); crumb names this model rather than the hidden 大家都在找 shelf.

Document-level listener at module load (same reason as `card.js` / `dist.js`):
```js
document.addEventListener("click", (e) => {
  const btn = e.target.closest("[data-model-more]");
  if (!btn) return;
  const rest = btn.closest(".k-msale").querySelectorAll("[role=list] > [hidden]");
  for (const li of rest) li.hidden = false;
  btn.remove();
  rest[0]?.querySelector(".k-card__model")?.focus();
});
```
Comment on the listener: 看全部 reveals the cards already rendered instead of linking to Browse, which does not exist yet; a link there would end on the not-found page (A9).

**Rationale:** one view module per route (home/detail pattern); hero figures straight from `ModelStats` + overview (S8); distribution reused unchanged; real-derived tiles honest per def (value or 資料不足).

**Verification:** Step 5 rows M1–M12.

**If this fails:** delete `model.js`, drop the route line in `app.js` (crumb returns to not-found).

---

### Step 5: End-to-end verification

**Objective:** S1–S9 evidence. **Confidence:** Medium (hotlinked photos/fonts need network). **Depends on:** Steps 1–4.
**Verify-Projex:** Encouraged

**Procedure:**
1. `go vet ./...` + `go test ./...`; hygiene grep `grep -rnE "[0-9]{10}-[a-z0-9-]+-(plan|sprint|def|nav|audit|patch)" internal cmd` → none.
2. Background `go run ./cmd/kura serve -db kura.db -addr 127.0.0.1:8090` (worktree root; no crawl flags).
3. Scratch copy: `sqlite3 -readonly kura.db ".backup '<scratchpad>/mut.db'"`; in `mut.db`: `UPDATE listings SET first_seen = first_seen - 864000, last_seen = last_seen - 864000; UPDATE price_points SET observed_at = observed_at - 864000;` (epoch → 10/3); for the two cheapest listings of HERMES `027633CC` (`SELECT id FROM listings WHERE brand='HERMES' AND upper(model_code)='027633CC' ORDER BY price_jpy LIMIT 2`): `first_seen = strftime('%s','now') - 8*86400, status = 'gone', gone_at = strftime('%s','now') - 2*86400`; for HERMES `001855CC` (1 listing): `status = 'gone', gone_at = strftime('%s','now') - 86400`. Serve on `127.0.0.1:8092`.
4. Matrix (browser pane; `/model/{enc brand}/{enc key}`):

| # | Case | Model | Check |
|---|---|---|---|
| M1 | largest group | CHANEL `name:マトラッセ` (409) | tiles: 在售 409 件, 已售出 + 平均找到新主人 資料不足 · 自 10/13 起記錄, median tile = largest grade; 409 dots; 5 cards + 看全部 409 件 → click → 409 cards, focus on card 6, button gone |
| M2 | exactly 5 | HERMES `027633CC` | 5 cards, no 看全部; tile 3 ランクA (4) median = `by_grade` |
| M3 | 1 active, code | HERMES `001855CC` (ケリー28) | 1 card; sub `手提包 · 001855CC`-style |
| M4 | stats photo pending | LOUIS VUITTON `M52052` / CHANEL `AS4906` | real photo from an item / canvas icon (1 active, pending) |
| M5 | no labelled grade | HERMES `name:エルメスケリー28cm` | 3 tiles, no distribution, no crash |
| M6 | `/` in key | HERMES `name:24/24ヴァンキャトル29` | direct load + reload + via detail crumb |
| M7 | `.` in key, no code | CHANEL `name:2.55マトラッセ` | reload; sub = 類型 only |
| M8 | tie on count | LOUIS VUITTON `M51165` | tile 3 = better of tied grades |
| M9 | flow | Home card → detail → crumb → model → card → detail; Back ×2 | in-app nav, titles, back restores each view |
| M10 | not found | `/model/HERMES/NOPE`, `/model/NOBRAND/X`, `/model/HERMES/%E0%A4%A` | not-found view + title |
| M11 | track | M2 toggle → detail → crumb back | 追蹤中 kept (same session); detail 追蹤這件 unaffected |
| M12 | sold / days / 0 active | :8092 HERMES `027633CC`; HERMES `001855CC`; any other model | 已售出 2 件 · 自 10/3 起記錄, 平均找到新主人 6 天, 2 rings, 3 cards; 001855CC: 在售 0 件, no tile 3, no 在售 section, ring only; other model 已售出 0 件 |

   M1 + M2 at 1280 and 375 × `?theme=light|dark` (and once via colorScheme emulation); M12 at 1280 light; rest 1280 light (M6 also 375).
5. Per view: screenshot vs canvas `Model` / `ModelMobile`; `read_console_messages` onlyErrors empty; at 375 `document.documentElement.scrollWidth <= innerWidth`; Tab order crumb → track → dist toggle → 看全部 → cards with visible ring; page text has no 人在找 / 售出率 / 近 30 天 strings (S8).
6. Stop servers; delete `<scratchpad>/mut.db*`; worktree `git status` clean except committed work.

---

## Verification Plan

### Automated Checks
- [x] `go vet ./...` clean
- [x] `go test ./...` green incl. extended `TestStaticRoutes`, `TestListingAPI`

### Manual Verification
- [x] Step 5 matrix M1–M12
- [x] Home + detail unchanged (detail track label, crumb now lands on the model view)
- [x] Source hygiene grep → none

### Acceptance Criteria Validation

| Criterion | How to Verify | Expected Result |
|---|---|---|
| S1 | M6, M7, M9, M10; `document.title` | view renders on click + reload; not-found for bad paths |
| S2 | M1/M3/M4 screenshots vs canvas | canvas layout/values; crumb per A2; photo or icon per A3; no demand line |
| S3 | M1, M2, M5, M8, M12 vs `/api/models/…` + `/api/overview` | tiles equal server fields; 資料不足 on real DB |
| S4 | M11 | label/aria-pressed flip; state survives navigation |
| S5 | M1, M2 dots/tick/table vs API | model mode (no 此件), ladder from `scheme`, `*_twd` figures |
| S6 | M1, M2, M3, M12 | cap 5, in-place 看全部, focus move, 0-active omission |
| S7 | 375 scrollWidth; both themes; focus | no overflow; tokens switch; ring visible |
| S8 | code review of `model.js`; page-text grep | server fields only; no sell-through/demand/sample values |
| S9 | vet/test/grep | green / none |

**Data-class trace (model):** name / code / 類型 / brand / photo ← listing fields (real) | 在售 N ← `active` (real) | `{grade} 中位數` + 件比較 ← `by_grade` (real — current in-sale listings) | ladder ← `scheme` (real) | dots/table ← group items + `by_grade` (real) | cards ← items (real) | 已售出 ← `sold` gated by overview `sold` (real-derived → 資料不足 now) | 平均找到新主人 ← `median_days_to_sell` (real-derived → 資料不足 now) | `自 {m/d} 起記錄` ← overview `since` (real). Mocked: none rendered (demand line omitted). Track toggle = in-memory UI state, no number.

---

## Rollback Plan

1. `git -C <work-root> log --oneline a05b9ee..HEAD` → this plan's commits.
2. `git -C <work-root> revert <sha>…` newest first (sprint branch only); `/model/…` returns to the not-found view.

---

## Notes

### Split Decision
No split — single scope (`internal/web`, one `.projex` folder), within size budget (~370 lines / ~33 KB; 5 steps — heuristic needs > 5 steps and > 500 lines).

### Risks
- 409 cards rendered (5 visible) → ~0.5 MB HTML on the largest model besides the 477 KB API payload: acceptable locally; Browse/pagination later.
- Expanded carousel at ≤ 767 with hundreds of cards is long to swipe — only the few largest groups; Browse supersedes (A9).
- Advisor `model_price_stats` returns `sold_observed: 0` / `sell_through: 0` while the page says 資料不足 — not a second computation (UI shows no number), but the advisor could read 0 as "never sells". Advisor iteration should apply the same gate.
- Hotlinked photos / fonts unreachable → same exposure as Home/detail.

### Open Questions
None blocking execution. For the orchestrator to log (sprint nav / def), not decided here:
- A2 crumb, A9 看全部 in place, A7 「近 30 天」 → 「自 {m/d} 起記錄」 → canvas deviations.
- A7 label 平均 vs median value; A7 「不到 1 天」; A10 0-active state → canvas pass 2 copy/state review.
- A11 資料不足 tile → def § Canvas gaps candidate.
- A5 `soldKnown` gate = factual minimum only; def L4 「Day-1 appearance」 (time-based rule) still open and replaces it if answered.
- A4 demand line + A2 crumb return when the 大家都在找 open question is answered.
