# Audit: Kura canvas foundation + Home (guest)

> **Status:** In Progress (Conditions Open)
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 1
> **Author:** Claude Opus 5.5 (auditor, sprint subagent)
> **Parent:** 2610130552-kura-canvas-foundation-home-patch.md
> **Related Projex:** 2610130535-kura-canvas-foundation-home-plan.md (audited plan) | 2610130552-kura-canvas-foundation-home-patch.md (execution claims) | 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md (data semantics, canvas precedence) | 2610130607-kura-home-audit-routing-fixes-patch.md (F2 F3 F4 F5 fix)
> **Audit Date:** 2026-10-13 | **Work Period:** 2026-10-13, commits `0e472c0` `169953a` `fecb3e2` on `projex/sprint/2610130524-kura-interactive-mock` (range `905cdc6..fecb3e2`); re-audit 2026-10-13 of fix `5a9ede3` + doc `0531c66` (range `89d61d7..0531c66`)
> **Subject:** plan S1–S11 vs code, live app (real `kura.db` + empty temp DB), canvas pinned 1791839497-46ec (`project/Main.dc.html`, `project/Card.dc.html`, `project/CardStates.dc.html`; served as live 1791840357-ac34, same content per sprint nav)

---

## Audit Summary

**Claim:** `/` serves canvas shell (tokens, fonts, light/dark, guest header, footer w/ live FX) + Home on real data (最新上架 5 | 價格最美 3, `trends: hidden`, empty state); card in all states + `/dev/cards.html`; legacy at `/legacy/`; `source_name` / `photo_pending` / `drop.from_twd|to_twd`; vet/test green; S1–S11 ✓.

**Verdict:** Verified — every S1–S11 claim reproduced independently. Two Significant issues sit outside the criteria (one data-semantics decision taken silently, one latent routing trap).
**Re-audit (2026-10-13):** F2 F3 F4 F5 fixed and independently verified, no regression (§ Re-audit). Open: F1, F6 — human decisions.

**Assessment:** Completeness: High | Correctness: High (one latent bug) | Quality: High | Value: High

**Top Issues:**
1. 價格最美 definition — open def L3 question decided without logging; count 1,471 (≥5% `only_deals`) vs canvas/HANDOFF sample 1,020 (= `overview.deals`, ≥10%). Three "below market" thresholds now coexist (card 3% | shelf 5% | overview 10%). **Needs decision.**
2. `spaFallback` 404s any extension-looking path → 104 live model codes contain `.` (`191502BZJ.04LU`) → future model/detail routes keyed by code break on reload/pasted link. **Fixable now.**
3. 「看全部價格最美 1,471 件 →」 links to bare `/browse` — the count's filter is not carried in the URL. **Fixable now** (or pin as browse-iteration contract).

---

## Claims vs Evidence

| Claim | Evidence | Status | Notes |
|---|---|---|---|
| S1 shell/routing/MIME/legacy | `TestStaticRoutes` pass (re-run `-count=1`); live curl: `/` `/browse` `/listing/brandoff:1` 200 html; `/js/app.js` 200 `text/javascript; charset=utf-8`; `/kura.css` `text/css`; `/js/missing.js` `/api/nope` 404; `/legacy` → 301 `/legacy/` | ✓ | Extra probes: `/listing/a.b` → 404 (F2); `/js/` `/dev/` → 200 directory listing (F5) |
| S2 tokens | Programmatic compare `docs/design-handoff/kura-tokens.css` vs `kura.css`: all 89 `--k-*` declarations identical values; only extra = `--k-chart-mark` light `#7A6B58` / dark `#857868` ×2 | ✓ | Selectors per plan (`:root,[data-theme=light]` / OS-dark `:root:not([data-theme=light])` / `[data-theme=dark]`) |
| S2 theme param | `?theme=light` body `rgb(245,240,230)`; `?theme=dark` `rgb(27,24,21)`; `/dev/cards.html` sections paint per own `data-theme` | ✓ | OS-emulation path not re-run (CSS reviewed, executor evidence) |
| S3 fonts | Google Fonts URL byte-identical to tokens file; `document.fonts` loaded EB Garamond, Noto Sans TC, Noto Serif TC, Shippori Mincho; h2 `Noto Serif TC`, model `Shippori Mincho`, price `EB Garamond` | ✓ | |
| S4 guest header | Markup vs canvas `Main.dc.html` header: seal/wordmark/pill/placeholder/登入/加入會員 + mobile icon identical values; 375: pill + 加入會員 `display:none`, icon flex | ✓ | |
| S5 footer | Live `#fx` = ` · 匯率 1 JPY = 0.2018 TWD（2026/10/12）` ← API `Rate 0.201837`, `AsOf 2026-10-12T08:02:32+08:00`; empty DB `0.2000 TWD（預設匯率）`; 6 links order = canvas; nav 3 cols / 2 at 375; pad-bottom 104/96 | ✓ | Fallback label 「預設匯率」 is plan copy, not canvas (canvas has no fallback state) |
| S6 card states | `/dev/cards.html` both themes: 9 labelled states, badge/drop/sold/pending copy = canvas `Card.dc.html` verbatim; row card present; track toggle → `aria-pressed=true`, label 追蹤中，點一下取消, colour `rgb(227,161,166)` (dark), toggles back | ✓ | Fixture photos hotlink BRAND OFF (lazy-load; all loaded) |
| S7 Home data | Live DOM vs `/api/listings?sort=newest&limit=5` + `?deals=1&sort=deal&limit=3`: ids, prices, badges identical; link 「看全部價格最美 1,471 件 →」 = `total 1471` | ✓ | Data semantics: F1; link target: F3; shelf quality: F6 |
| S8 empty | Temp empty DB :8094 → `main` = 還在準備中。/店裡正在整理，新的包很快就會上架。, 0 `.k-sec` | ✓ | |
| S9 responsive | 1280: pad 40px, 5 cols, 最新上架 `44px 0 64px`, 價格最美 head brass rule + 28px; 375 (dark): pad 16, header `14px 0 0` nowrap, h2 24, flex carousel 200px, `scrollWidth = clientWidth = 375`, no element past viewport outside `.k-row`; no a/button/input <24px either width; Tab → `.k-logo` `:focus-visible` `2px solid` | ✓ | Values = canvas `renderVals` (`padX` `headerPad` `freshPad` `sectionPad` `h2Size` `freshRow` `freshItem` `footPadB` `footCols`) exactly |
| S10 data classes | Every Home number ← API field (price `price_twd`, badge `fair.*`, count `total`, FX `/api/overview fx`); fixtures only under `/dev/` | ✓ | Real class correct; *which* real definition = F1 |
| S11 vet/test/hygiene | `go vet ./...` clean; `go test -count=1 ./...` ok advisor/radar/brandoff/store/web; grep projex refs in `internal cmd` → none | ✓ | |
| Patch doc locators | `server.go` L43/L56–61/L63+, `condition.go` L25, brandoff L11, radar L45/56–59/196–197/219/228, test L20/154/187, `kura.css` 213 L, L95–97/123/158–160 | ✓ | All accurate |
| Deviation 1 (`.k-card{position:relative}`) | Reproduced reasoning: sr-only span absolute → escapes `.k-row`; now no overflow | ✓ | Good catch — plan's own `scrollWidth <= innerWidth` check was a false pass under mobile emulation |

---

## Objective Verification

### Step 1 — server routing, MIME, legacy move
**Evidence:** `internal/web/server.go` → `spaFallback`, `init`; `server_test.go`; `git diff --stat` (legacy `app.js` `style.css` renames 100%, `index.html` moved).
**Findings:** Actual = plan code verbatim. Missing: none. Quality: High. Latent: extension heuristic (F2), directory listings (F5).
**Verification:** ✓ Verified

### Step 2 — card data fields
**Evidence:** `listing.Scheme.Shop`, `brandoff.Scheme`, `radar.Item.PhotoPending|SourceName`, `radar.Drop.FromTWD|ToTWD`, `photoPending`; tests `TestDropsAndStats`, `TestPhotoPending`.
**Findings:** Actual = plan. DB check: 26 `NowPrinting` rows, 0 empty `image_url`; a pending card (`brandoff:2104102531481`) is live in 價格最美. `ToTWD` unused by client (harmless, additive). Quality: High.
**Verification:** ✓ Verified

### Steps 3–6 — static UI
**Evidence:** `kura.css`, `index.html`, `js/{util,app,card,home}.js`, `dev/cards.{html,js}`; canvas `Main.dc.html`, `Card.dc.html`, `CardStates.dc.html`.
**Findings:** Card markup/copy/inline values = canvas `Card.dc.html` (adds sr-only 「（另開視窗）」 and 10% mat padding for real photos — both justified). Router: pushState nav keeps document (marker survived `/about` → back), not-found view + title, `popstate` ignores hash-only (登入 `#` → back keeps Home). Quality: High.
**Verification:** ✓ Verified

### Step 7 — browser verification
**Evidence:** independent re-run (this audit) at 1280 light / 375 dark / dev page / empty DB / legacy; console errors: none.
**Verification:** ✓ Verified

---

## Code/Implementation Inspection

### `internal/web/server.go` → `spaFallback`
**Claimed:** fallback for extensionless paths; asset/API 404 preserved.
**Actual:** as claimed — High.
**Issues Found:**
- `path.Ext(name) != ""` ⇒ 404 for any path whose last segment holds a `.`. 104 active listings have dotted model codes (`191502BZJ.04LU`, `101513DPV.38NO`, …; 2 more contain `/`). A model route `/model/{code}` (or group key) renders via in-app nav but 404s on reload / pasted link / new tab — Severity: **Significant (latent)**. Listing ids (`brandoff:<digits>`) unaffected.
- Directories without `index.html` (`/js/`, `/dev/`) answer with `FileServerFS` listings — Severity: Low.
**Undocumented:** none.

### `internal/web/static/js/card.js` → `valuation`
**Actual:** ±3% band (legacy `valBadge` parity, plan A5).
**Issues Found:** card says 低於行情 3–4% for listings the 價格最美 shelf/count (≥5%) and advisor `only_deals` exclude — Severity: Minor, part of F1. No rationale comment for 3% vs radar's 5% (Rule 6 gap).

### `internal/web/static/js/home.js` → `renderHome`
**Actual:** as planned. **Issues Found:** 價格最美 link `href="/browse"` drops the `deals=1&sort=deal` query that produced the count (F3) — Minor.

**Metrics:** Readability High | Test coverage: routing table + radar fields (JS untested — no build step by constraint; browser-verified) | Tech debt: Low

---

## Testing Validation

**Execution:** all pass (`go test -count=1 ./...`); not flaky across executor + audit runs.
**Coverage:** routing table 8 rows; radar new fields; MIME asserted on Windows (where A7 bites).
**Missing:** a dotted-path route row (would have exposed F2); directory request row (F5). JS logic (`valuation` thresholds) unit-untested — acceptable (no-build constraint), covered by fixture page.

---

## Documentation Audit

**Completeness:** plan + patch doc complete; deviations 1–7 listed honestly (incl. commit-grouping and trailer-order flaws).
**Accuracy:** patch doc claims all reproduced; locators accurate.
**Gap:** patch doc lacks a `**Rationale promoted:**` line (patch, not walkthrough — see Source Hygiene).

---

## Source Hygiene

**Scope:** Diff only (`905cdc6..fecb3e2`, `.projex/**` excluded)

| Rule | Location (symbol) | Quote | Severity |
|---|---|---|---|
| — | — | No Rule 1–5 violations: no projex ids/filenames, no `file:ln`, present tense, no step banners, `!important` and `position: relative` reassurances name the masked/escaping alternative | — |

**Rule 6 gaps:**
- `card.js` → `valuation` — ±3% band diverges from radar `Search` `OnlyDeals` (≥5%) and `Overview.Deals` (≥10%); comment states the band, not why 3% or why it differs — next reader will "fix" it to 5% or wonder why a 4% badge isn't in 價格最美.

**Rule 6 vacuity:** none. Rationale comments checked: `server.go` `init` (registry MIME constraint), `spaFallback` (rejected all-index.html shortcut), `photoPending` (source quirk), `card.js` listener (duplicate-listener trap), `app.js` `EXTERNAL` / `popstate`, `home.js` (sample values rejected), `kura.css` `!important` / `.k-card` / 768–1023, `dev/cards.js` header — each names alternative/constraint/trap.

**Rationale promotion:** A1 (大家都在找 hidden), A7 (MIME), fallback design live in code comments ✓. A5 threshold rationale exists only in plan → Rule 6 gap above. Footer slugs (A4), guest track (A2), 登入 `#` (A3) need none.

**Commit composition** (informational — fix forward, never rewrite history): 2 code commits checked — typed subject 2/2 | type matches diff 1/2 (`0e472c0` `feat(radar)` also carries the `internal/web/static` legacy moves — documented deviation 6) | `Projex:` trailer 2/2 (`2610130552-kura-canvas-foundation-home`; parses as trailer in both despite `0e472c0` order). `fecb3e2` doc-only, exempt.
**Violations:** `0e472c0` — scope/type covers radar only; at `0e472c0` alone `/` 404s (no `index.html`) — branch-internal, transient.

**Trailer survival on `master`:** 0/6 — **N/A, not Critical**: all 6 are baseline import + doc-only `projex(...)` commits; no code has landed on master yet. Sprint close must carry the `Projex:` trailer in its landing message.

---

## Gap Analysis

### Promised But Not Delivered
| Promise | Status | Impact |
|---|---|---|
| — | none — all plan steps + criteria delivered | — |

### Undocumented Issues
| Issue | Severity | Affects |
|---|---|---|
| F1 價格最美 definition chosen silently (open def L3); count ≠ canvas sample semantics | Significant (decision) | Home, browse iteration, advisor parity, def |
| F2 dotted path segments 404 on reload | Significant (latent) | model/detail iterations |
| F3 count link carries no filter | Minor | browse iteration |
| F5 `/js/` `/dev/` listings | Minor | hosted build |
| F6 shelf headed by anomaly-boundary items | Minor (decision, pre-existing radar) | reviewer impression |

### Unhandled Edge Cases
- Reload on any future route whose last segment has a `.` — 404 (F2).
- `?theme=` dropped on first in-app navigation (attribute persists until reload) — acceptable for a dev/review param.
- `it.url` / `it.image_url` escaped but not scheme-checked (`javascript:` would pass `esc`) — source is own crawler; Low.

---

## Quality Assessment

### Completeness: High
**Strengths:** every criterion, plus honest deviation log; fixture page mirrors canvas `CardStates` incl. its disclaimer.
**Gaps:** F1 not logged as an assumption beside A1–A5.

### Correctness: High
**Works:** routing, MIME, theme, router history, track sync across shelves, empty state, live FX.
**Bugs:** F2 (latent) — Significant.

### Code Quality: High
**Positive:** small modules, single document-level listeners, stale-render guard, all interpolation escaped.
**Tech Debt:** extension heuristic in `spaFallback` — Medium until fixed.

### Value Delivered: High
**Intended:** first canvas screen on real data. **Actual:** delivered; pixel values match canvas. **Impact:** User Positive.

---

## Open Findings

### Undocumented Discoveries
- **Three thresholds for "below market":** card badge `deal_pct ≥ 0.03` (`card.js` `valuation`, legacy parity) | 價格最美 + advisor `only_deals` `≥ 0.05` (`radar` `Search`) | `/api/overview` `deals` `≥ 0.10` (`radar` `Overview`; legacy KPI 「低於行情 ≥10%」 1,020). HANDOFF § market summary lists 「1,020 低於行情 ≥10%」 and canvas `Main.dc.html` hard-codes 「看全部價格最美 1,020 件」 — the designer's count was the ≥10% stat. Implementation shows 1,471.
- **Shelf content (live):** top 3 價格最美 all `deal_pct 0.500` exactly (`price_jpy 50000` vs `fair_jpy 100000`, same-grade, high) — one tick under `anomalyPct = 0.5` (`> 0.5` flags). Confidence-weighted sort pushes boundary items to the top; `radar` `value`'s own comment says half price is "usually damage or missing parts". Pre-existing radar behaviour, newly prominent as Home's hero shelf.

### Impact Analysis
- Downstream: browse iteration must honour 1,471 = `/browse` default (F3); model iteration hits F2.
- Future: Enabled — shell/router/card/tokens reusable as planned.
- Risks: def/canvas conflict on 價格最美 count settled silently — contrary to def § Canvas precedence ("never settled silently during build") — Likelihood High that a later view (e.g. header status line) reads `overview.deals` and shows 1,020 beside 1,471.

### Improvements
- Could be better: `spaFallback` 404 only for known static prefixes/extensions.
- Missed: test rows for dotted path + directory.

---

## Stakeholder Impact

| Stakeholder | Promised | Reality | Impact |
|---|---|---|---|
| Reviewer (job application) | canvas Home on real data | delivered, faithful look | Positive |
| Later iterations | reusable shell/router/card | delivered; F2/F3 traps pending | Neutral→Positive once fixed |
| Def owner | data semantics governed by def | 價格最美 definition picked without record | Negative until logged |

---

## Findings

### Critical (Must Address)
- None.

### Significant (Should Address)
- **F1 — 價格最美 definition decided silently [needs decision]** — def L3 「Home sections 大家都在找 / 最新上架 / 價格最美 — definitions of each ranking?」 is open; plan deferred only 大家都在找 (A1) and silently took `only_deals` (≥5%) + `sort=deal` for 價格最美. Canvas sample count = ≥10% `overview.deals`. → Orchestrator/human: pick one definition (a) keep ≥5% advisor predicate (one source with advisor) and log as canvas deviation / sprint decision, or (b) adopt ≥10% everywhere (shelf, count, advisor `only_deals`), or (c) defer; then align card ±3% band rationale and `/api/overview` `deals` so one stat has one computation.
- **F2 — `spaFallback` dotted-segment 404 [fixable now]** — 104 live model codes contain `.`. → Replace `path.Ext(name) != ""` with a check scoped to static asset space (e.g. 404 when the first segment is `js`/`dev`/`legacy` or the extension is a known asset type `.js .css .png .svg .ico .jpg .webp .woff2 .map .json`); add `TestStaticRoutes` rows `/model/191502BZJ.04LU` → 200 shell and `/js/` → 404 (covers F5).

### Minor (Nice to Fix)
- **F3 — count link drops its filter [fixable now]** → `home.js` 價格最美 `href="/browse?deals=1&sort=deal"` (mirrors the counted query; canvas `Browse` artboard = 價格最美 list), or record "`/browse` default = deals" as browse-iteration contract.
- **F4 — Rule 6 gap `card.js` `valuation` [fixable now, wording depends on F1]** → comment naming why ±3% and that it intentionally differs from the shelf predicate (or align).
- **F5 — `/js/` `/dev/` directory listings [fixable now]** → folded into F2 fix.
- **F6 — 價格最美 headed by anomaly-boundary items [needs decision; pre-existing radar]** → consider ranking cap/penalty near `anomalyPct`, or accept; out of iteration-1 scope.
- Informational: `0e472c0` type/scope covers radar only though it moves web files; trailers present; master trailer survival N/A until first code landing.

### Positive
- All 89 token values exact; every canvas `renderVals` layout value matched.
- Executor found and fixed a false-pass in the plan's own overflow check (deviation 1) and documented every deviation with cause.
- Card copy verbatim to canvas incl. 「尚無行情估價」; fixture page carries the canvas disclaimer; no sample value on `/`.

---

## Recommendations

**Immediate:** F2 + F5 + F3 as one `[patch]` (server routing + test rows + one href). Escalate F1 (and F6) to orchestrator for logging/decision before the browse iteration builds the 價格最美 list.
**Future:** browse iteration — honour whichever F1 definition; model iteration — add a dotted-code route test.
**Process:** plan Assumptions should enumerate every def L3 open question the plan touches, not only the one it defers.

---

## Final Verdict

**Status:** Accept with Conditions

**Overall Assessment:**
- Completeness: High
- Correctness: High
- Quality: High
- Value: High

**Conditions:**
- [x] F2 fixed + tested (dotted path → shell; asset dir → 404) — 2610130607-kura-home-audit-routing-fixes-patch.md (`5a9ede3`, F5 folded in) — re-audit verified (§ Re-audit)
- [ ] F1 recorded as decision/deviation (sprint nav or def) or resolved — open; human decision (to be logged as sprint nav Open Question)
- [x] F3 link carries filter, or browse-iteration contract recorded — `href="/browse?deals=1&sort=deal"` via 2610130607-kura-home-audit-routing-fixes-patch.md (F4 comment also landed there) — re-audit verified (§ Re-audit)
- [ ] F6 價格最美 ranking near `anomalyPct` accepted or changed — open; human decision (pre-existing radar behaviour)

**Sign-off:** Yes, conditional — iteration 1 delivers its plan exactly; all fixable-now findings closed; F1 + F6 remain open human decisions and block only the browse iteration's 價格最美 list, not iteration 1.

---

## Re-audit — 2610130607-kura-home-audit-routing-fixes-patch.md

**Scope:** `git diff 89d61d7..0531c66` — code commit `5a9ede3` (`internal/web/server.go`, `server_test.go`, `static/js/card.js`, `static/js/home.js`); doc commit `0531c66` (patch doc + this audit's F2/F3 ticks). Patch doc claims not trusted — each re-checked below.

| Finding | Evidence | Status |
|---|---|---|
| F2 dotted routes | `spaFallback`: `path.Ext != ""` replaced by `underAssetDir` (first segment is an embedded dir) ∪ `assetExt` allow-list (17 exts, case-folded). Live :8095: `/model/191502BZJ.04LU` → 200 shell; `/model/LOUIS%20VUITTON/M40156` → 200 shell; `/listing/brandoff%3A2101220126889` → 200 shell. DB sweep (read-only): dotted `model_code` / `group_key` / `model_name` suffixes (`.04LU` `.38NO` `.38SI` `.02NT` `.Z99` `.V07` `.T20` …) — 0 collide with `assetExt`; 0 values start with `js/` `dev/` `legacy/`. New test row `/model/191502BZJ.04LU` → 200 | ✓ |
| F5 dir listings | Dir without `index.html` → 404: live `/js/` `/dev/` 404; `/legacy/` 200 (has index); `/legacy` 301 → `/legacy/`. Test rows `/js/` `/dev/` 404, `/legacy` 301 | ✓ |
| F3 count link | `home.js` 價格最美 `href="/browse?deals=1&sort=deal"`; live DOM href identical; click → pushState `/browse?deals=1&sort=deal` (not-found view until browse iteration — expected); back → Home, 2 sections | ✓ |
| F4 Rule 6 | `card.js` `valuation` comment names legacy parity (`legacy/app.js` `valBadge`), the 5% `OnlyDeals` predicate, the 10% `/api/overview` count, the visible consequence (3–4% badge absent from 價格最美) and that the constants are independent — names alternative + trap; no projex refs, no `file:ln` | ✓ |

**Regression checks:** `go vet ./...` clean; `go test -count=1 ./...` ok advisor | radar | brandoff | store | web (15-row `TestStaticRoutes`). Live route table: `/` `/browse` `/browse?deals=1&sort=deal` `/listing/*` → shell; `/js/app.js` `/js/card.js` `/legacy/app.js` `/dev/cards.js` → 200 `text/javascript`; `/kura.css` `/legacy/style.css` → `text/css`; `/dev/cards.html` 200; `/js/missing.js` `/js/missing` `/legacy/missing` `/missing.png` `/x.HTML` `/JS/app.js` `/api/nope` → 404; `/api/overview` 200 JSON; `/index.html` 301. Home: 5 + 3 cards, 價格最美 ids = `/api/listings?deals=1&sort=deal&limit=3`, count 1,471 = `total`, FX footer live, track toggle works, 0 console errors. `/legacy/`: tabs 行情雷達 / AI 找包顧問 / 資料管線 + 15,367 tile, 0 console errors.

**Source hygiene (fix diff):** `spaFallback` comment rewritten — names rejected test ("has any extension"), the dotted-code trap and the directory-listing alternative ✓; `underAssetDir` doc ✓; no projex refs / line numbers. Commit `5a9ede3`: `fix(web)` typed, matches diff, `Projex: 2610130607-kura-home-audit-routing-fixes` trailer, `Co-Authored-By` last ✓. `0531c66` doc-only, exempt.

**Residual (Low, informational):** `assetExt` contains `.html` and `.txt` — a future client route whose last segment ends in one of the 17 listed exts would 404 on reload. No live value does; revisit only if a route key gains such a suffix.

**Re-audit verdict:** Verified — F2 F3 F4 F5 closed; ticks earned. Open: F1, F6 (human decisions).
