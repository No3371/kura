# Audit: Kura advisor panel (顧問)

> **Status:** Complete (Accepted)
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 4
> **Author:** Claude Opus 5.5 (auditor, sprint subagent)
> **Parent:** 2610130822-kura-advisor-panel-patch.md
> **Related Projex:** 2610130815-kura-advisor-panel-plan.md (audited plan, S1–S11) | 2610130822-kura-advisor-panel-patch.md (execution claims) | 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md (one source per stat, demo label, Canvas gaps, L3 advisor questions) | 2610130748-kura-model-page-audit.md (carry-over H2) | 2610130854-kura-advisor-panel-audit-fixes-patch.md (F1–F7 fixes)
> **Audit Date:** 2026-10-13 | **Work Period:** 2026-10-13, commits `5e65a9a` (Go gate) `912a354` (web) `3220b17` (patch doc) on `projex/sprint/2610130524-kura-interactive-mock` (range `d749a53..3220b17`); re-audit 2026-10-13 of fix `2ab2e28` + doc `51701b1` (range `bf54d3d..51701b1`)
> **Subject:** plan S1–S11 vs code, tests, live app (worktree `kura.db` served read-only on 127.0.0.1, `advisor=mock`, no key), canvas `project/Main.dc.html` 顧問 block + `renderVals` (served 1791845608-ba84; plan records artboards identical to pin 1791839497-46ec). Browser pane hidden during run → screenshots partly flaky; every state the executor verified by geometry only (full 1280, dark 1280, close/focus return) re-checked here directly — screenshots captured for floating dark 1280, full dark + light 1280, 375 light + dark.

---

## Audit Summary

**Claim:** Canvas 顧問 in the shell on every route (dock → modal floating panel ⇄ full page → closed), SSE conversation from real tool events with markdown, chips, ≤ 4 figures (photo, grade, NT$, badge, `by_grade` position bar, 追蹤 / 看詳細 / 商品頁), functional error fallback; H2 closed — advisor tools withhold sold count / sell-through / days-to-sell until `radar.Engine.SoldStatsReady()`.

**Verdict:** Verified — every S1–S11 criterion holds on direct re-test; patch-doc evidence accurate. Two latent layout defects + three small gaps found that the patch did not report.

**Assessment:** Completeness: High | Correctness: Medium-High | Quality: High | Value: High

**Top Issues:**
1. **Dialog itself is a hidden scroll container with ~3,000 px phantom overflow** — `.k-sr` sr-only spans in figures (`position: absolute`) take the `position: fixed` dialog as containing block, escaping the `.k-adv__body` clip. Any scroll-into-view shifts the whole panel content up (header off-screen, blank band under composer). Reproduced. One-line fix verified.
2. **Full-page mode leaves the page scrollbar visible** (panel 1265 wide at 1280; page behind still scrollable) — executor's open issue confirmed visually; fix verified.
3. Small: `navigating` flag is ineffective (UA dialog focus-restore puts focus on the dock anyway); chips wrap mid-border across lines; per-turn `known` map → cross-turn `[[id]]` shows 「查看商品」 with no figure.

**Re-audit (fix `2ab2e28`):** F1 F2 F3 F4 F5 ✓ fixed; F6 ✓ and F7 ✓ implemented per orchestrator decisions (dark scrim token-driven pending canvas pass 2; no composer focus on narrow/coarse pointer). No regressions. Router-wide focus target (F3 part 2) = orchestrator carry-over. → Accepted. See § Re-audit.

---

## Re-audit: fix patch 2610130854-kura-advisor-panel-audit-fixes-patch.md

**Scope:** `2ab2e28` (`index.html`, `advisor.js`, `kura.css`) + doc `51701b1` (patch doc; it also ticked this audit's two Conditions — ticks checked below, both earned). Re-run: `go vet ./...` clean | `go test -count=1 ./...` ok (advisor, radar, brandoff, store, web) | `node --check advisor.js` ok | hygiene grep none. Live: worktree `kura.db` read-only, `advisor=mock`, own tab, real clicks/keys; 1280 (pane dark + `?theme=light|dark`) and 375 mobile emulation (`pointer: coarse` true).

| # | Fix | Evidence (independent) | ✓/✗ |
|---|---|---|---|
| F1 | `.k-adv__body { position: relative }` | 1280 dark, 1 turn, 4 figures, 4 bars: dialog `scrollHeight` 638 = `clientHeight` 638 before; after `scrollIntoView({block:'center'})` on every `.k-sr` span + last figure + textarea → 638 = 638, `scrollTop` 0, header y 137 (dialog y 136). Original trigger (tool ref-click on textarea, which produced `scrollTop` 255 before) → `scrollTop` 0, screenshot shows header + composer in place. 375 light, 4 figures: 598 = 598 before/after same scroll-into-views, header y 205 (dialog 204); ref-click on textarea → `scrollTop` 0; screenshot ✓. After cross-turn replay (6 figures) at 1280: 638 = 638 | ✓ |
| F2 | `html:has(.k-adv.is-full[open]) { overflow: hidden }` | full 1280: panel 1280 wide, root `clientWidth` 1280, html `overflow-y: hidden`, dialog 800 = 800; after Esc / shrink: html `visible`, floating 680, page scrollbar back (1265) — no lingering lock | ✓ |
| F3 | `navigating` flag removed; close → `dock.focus()` | real click 看詳細 (1280) → `/listing/brandoff%3A2107601166086`, panel closed, detail rendered, focus `.k-adv-dock` (same as UA restore — behaviour unchanged, code now honest) | ✓ |
| F4 | `.k-adv__chip` inline-block, `max-width: 100%` | full 1280: 0 chips with > 1 client rect (was 1+ split); screenshot: chips wrap whole onto next line; 375: 0 split, 0 chips past body edge | ✓ |
| F5 | `known` at module scope | stubbed 2nd turn (text only, no `listings` event) citing two turn-1 ids + unknown id → chips 「バティニョール ヴェルティカル · M51153」「ネヴァーフルPM · M40155」「查看商品」; `P FIGURE FIGURE`; 2 bars drawn; no step list | ✓ |
| F6 | dark scrim `color-mix(--k-paper-sunk 60 %)` (OS dark + `[data-theme="dark"]`), floating only | OS-dark 1280: `color(srgb .082 .075 .063 / .6)`; `?theme=dark`: same; full: transparent; `?theme=light` (OS dark): canvas literal `rgba(30,25,21,.18)`. Screenshot: page behind visibly dimmed. `::backdrop` resolves the token (inherits from dialog in this engine) | ✓ |
| F7 | no `autofocus`; narrow/coarse → `h2[tabindex=-1]`, else textarea | 375 mobile (coarse): real click dock → focus H2 「Kura 顧問」 inside dialog, not `:focus-visible`, outline none; real Tab → 放大成整頁 (inside). 1280 fine pointer: real click + real Enter on dock → TEXTAREA | ✓ |

**Regressions:** none. Dock on `/`, `/listing/…`, `/model/…`, `/nope` (bottom 28, centred); Home (8 cards, 最新上架 / 價格最美), detail (h1 + 商品資料 / 行情估價), model (h1 + 分布 + 31 cards), not-found render; page `scrollWidth` = viewport minus scrollbar, html overflow `visible` with panel closed; open / max / shrink / Esc / 關閉顧問 / reopen-floating / focus containment ✓; console errors none.

**Source hygiene (fix diff):** no rule 1–5 violations; new comments name the rejected alternative or constraint (`known` lifetime, `noComposerFocus` "Rejected: the textarea's autofocus on every screen", body containing block, scrim, chip). Commit `2ab2e28`: `fix(web)` subject, `Projex:` trailer ✓.

**Observation (environment, not a defect):** with the Browser pane hidden, the dialog `close` event was not dispatched (rendering paused) — UA focus-restore still put focus on the dock and the dock handler's `setFull(false)` resets the panel on reopen, so no visible effect. Executor's original "next frame" caveat is this same artifact.

**Carry-overs (orchestrator):** router-wide focus target after navigation (F3 part 2) → later shell-wide a11y item; dark scrim provisional until canvas pass 2 draws a dark advisor artboard (def § Canvas gaps 「dark × advisor」).

---

## Claims vs Evidence

| Claim | Evidence | Status | Notes |
|-------|----------|--------|-------|
| Sold-stat gate on 3 tools + summary + marker | `tools.go` diff `5e65a9a`; `TestSoldStatsGatedUntilHistory` run green | ✓ | Not-ready + ready paths both asserted; single predicate (`SoldStatsReady` / `Overview().SoldStatsReady`) |
| Prompt + tool description say unknown ≠ 0 | `prompt.go` diff | ✓ | 呈現方式 bullet + `model_price_stats` description |
| `go vet` / `go test` green | re-run `-count=1` | ✓ | advisor, radar, brandoff, store, web ok |
| Hygiene grep none; `node --check` | re-run | ✓ | grep exit 1 |
| Dock on every route, bottom 28 / 16 | live `/`, `/listing/…`, `/model/…` at 1280 + 375 | ✓ | 148 × 56, centred; `/nope` by markup (static in `index.html`) |
| Floating 680 × 640 @ bottom 24, `:modal`, backdrop | live geometry + screenshot (dark) | ✓ | centred on `clientWidth` 1265 (scrollbar), not 1280 — imperceptible |
| Full = inset 0, 720 column, transparent backdrop, label swap | live geometry + screenshots (dark, light) | ⚠ | width 1265: page scrollbar visible beside panel (Top Issue 2) |
| Esc / 關閉顧問 → closed, floating reset, focus to dock | real Esc from full mode; real click on 關閉顧問 | ✓ | close event + focus return fired promptly even with pane hidden — executor's "next frame only" caveat not reproduced |
| Focus contained while modal | 10 real Tabs: cycle 送出 → max → close → 4 prompts → textarea → … | ✓ | never left dialog |
| Focus to textarea on open | real click + real Enter on dock | ✓ | `autofocus` |
| 示範模式 pill + canvas title | live `/api/overview` `advisor: "mock"` → pill visible | ✓ | |
| Prompt chip sends; chips removed | real click | ✓ | |
| One 行情 line w/ server summary ✓ | live: 「找到 684 件符合條件的商品，回傳 6 件」 ✓ | ✓ | |
| Figures ≤ 4 after first-reference block | live: `P P UL FIGURE×4 P` | ✓ | |
| Figure numbers = server fields | 4 figures vs `/api/models/LOUIS VUITTON/{M51153,M51170,M40007,M40155}` `by_grade` {brandoff, ランクB}; badge 43 % vs listing `fair.deal_pct` 0.4286 | ✓ | count/min/median/max/aria-label all equal; dot 0 % (each listing is its grade's min) |
| Composer Enter / Shift+Enter / blank | real keys: Shift+Enter → `"GOYARD\n100 萬以內的腰包"`; Enter sends; Enter on empty → no turn | ✓ | IME by code read (`!e.isComposing`) — no real IME available |
| 追蹤 shared set; 看詳細 closes panel + navigates; conversation kept | real clicks | ✓ | detail 追蹤這件 shows 追蹤中; reopen → 1 turn, 4 figures, still 追蹤中 |
| Error + is_error + unknown tool + escaping (V10) | own fetch-stub replay | ✓ | 行情 ✓ / 成色 ✓ (failed line removed, unknown tool no line); `P(chip) → FIGURE → P → P.k-adv__err`; mid-stream cursor + `aria-busy` + 送出 disabled, after: gone/enabled; `<b>` in text escaped; unknown id chip 「查看商品」 |
| Network failure (V11) | stub reject | ✓ | 「暫時無法載入資料，請稍後再試。」 |
| 375 geometry, no x-scroll | live light + dark | ✓ | panel 8 / 204 / 359 × 600, bottom 8; `scrollWidth` 375; median label on own row, ticks copy `display: none` |
| Console clean | `read_console_messages` onlyErrors | ✓ | none |
| `/legacy/` untouched | `git diff d749a53..3220b17 -- internal/web/static/legacy` empty; `TestStaticRoutes` legacy case green | ✓ | live legacy send not re-run |

---

## Objective Verification

### S1 Dock + states — ⚠ Partial (holds as specified; Top Issues 1–2 are latent / cosmetic)
Geometry, modal, backdrop, focus round-trip, dock hidden while open all verified directly. Defects outside the criterion text: phantom dialog overflow (Significant), full-mode scrollbar (Minor).

### S2 Header ✓ | S3 Intro ✓ | S4 Composer ✓
Copy + SVGs verbatim vs canvas markup (`Main.dc.html` 顧問 block). Header drops canvas's extra column `div` (noted deviation, no visual difference).

### S5 Turn ✓
Real steps only; 風格 never drawn; cursor during stream, removed at end; markdown escaped before markup.

### S6 Figures ✓
Order, cap, active-only, single photo / pending, badge via `valuation`/`badgeHTML`, bar rule (count ≥ 2, max > min). Gap: figures and chip labels only for listings from the **same turn** (Minor F5).

### S7 Navigation ✓
In-app link closes; shop link (`target=_blank`) skipped by handler. `navigating` flag dead (Minor F3).

### S8 Errors ✓ | S9 One source per stat ✓ | S11 ✓

### S10 Responsive/themes/a11y — ✓ with notes
Tokens switch in dark (panel `rgb(36,32,28)` dark / `rgb(251,248,242)` light); 375 no overflow; focus contained; ring global. Dark backdrop = light canvas literal → visually nil on dark ground (known def gap, plan A8). Reduced motion: rule present by code read (no emulation).

---

## Code/Implementation Inspection

### `internal/advisor/tools.go`, `prompt.go`, `tools_test.go`
**Claimed:** gate + marker + prompt. **Actual:** as claimed; `model.Sold` → `*int,omitempty`; `Overview()` hoisted so per-model and top-level gate share one snapshot — Quality: High.
**Issues:** none. `search_listings` / `get_listing` compact still carry `status: "sold"` per listing and `days_listed` — per-listing facts, not sold-derived aggregates; correct to leave.

### `internal/web/static/js/advisor.js` (344 lines)
**Actual:** matches plan Steps 3.1–3.9 — Quality: High (escape-first markdown, `textContent` for server summaries, cached model fetch with failure eviction, A12 pinning incl. late bars).
**Issues Found:**
- `navigating` flag cannot do its job: `dialog.close()` runs the UA focus-restore to the opener (the dock) regardless; live after 看詳細 → `document.activeElement` = `.k-adv-dock`. Behaviour acceptable, code misleading — Severity: Low.
- `known` is per `ask()` call → a later turn citing an earlier turn's `[[id]]` without a new `listings` event renders 「查看商品」 and no figure (Claude multi-turn only; Mock always re-searches) — Severity: Low.
**Undocumented:** none beyond patch Notes.

### `internal/web/static/kura.css` advisor block
**Issues Found:**
- `.k-adv` (`position: fixed; overflow: hidden`) is the containing block for every `position: absolute` descendant whose nearer ancestors are static — the figures' `.k-sr`（另開視窗） spans. Their static position lies deep in the body's scroll content, so the dialog's scrollable overflow = 3,577 px vs 638 client (1280, 2 turns) / 3,375 vs 598 (375, 1 turn, no tool interaction). `overflow: hidden` still scrolls programmatically: a ref-click on the textarea (tool `scrollIntoView`) set dialog `scrollTop` 255 → header at y −118, 256 px blank band under composer (screenshot). Real-user triggers: anything that scrolls an element into view across ancestors — screen-reader virtual cursor landing on the sr-only text, mobile browsers scrolling the focused textarea above the keyboard, find-in-page. **Fix verified live:** `.k-adv__body { position: relative; }` → dialog `scrollHeight` 638 = `clientHeight` — Severity: Medium.
- Full mode `width: 100%` resolves against the viewport minus the root scrollbar → 1265 at 1280; page behind still wheel-scrollable over the header/composer. **Fix verified live:** `html:has(.k-adv.is-full[open]) { overflow: hidden; }` → panel 1280, floating unaffected (680). Layout shift happens only under the opaque full panel — Severity: Low.
- `.k-adv__chip` is an inline `<a>`: at 1280 full the chip 「ネヴァーフルMM · M40156」 breaks across lines with its border split (screenshot); canvas chip is a `<button>` (atomic inline). Fix: `display: inline-block; max-width: 100%` — Severity: Low.

**Metrics:** Readability/Maintainability: High | Test Coverage: Go gate covered both states; JS none (repo has no JS test harness — consistent with iterations 1–3) | Technical Debt: Low

---

## Testing Validation

**Coverage:** Go: gate test asserts absent keys + marker + summary, then ready path with real store writes. JS: browser-only.
**Execution:** All pass? Yes (`go test -count=1 ./...`) | Flaky: none.
**Missing:** no test that the panel's own geometry stays put (a `scrollHeight == clientHeight` assertion would have caught Top Issue 1 — browser-only, so manual).

---

## Documentation Audit

**Completeness:** plan, patch complete; patch Notes list deviations honestly.
**Accuracy:** Matches implementation? Yes. Line refs in patch (`l.135–140` etc.) match `5e65a9a`. Caveat "close event … fired only on the next frame" not reproduced (benign over-caution).
**Quality:** High.

---

## Source Hygiene

**Scope:** Diff only (`d749a53..3220b17`, `.projex/` excluded)

| Rule | Location (symbol) | Quote | Severity |
|------|-------------------|-------|----------|
| — | — | no rule 1–5 violations: no projex IDs/filenames, no `file:ln`, present tense, no step banners, no bare "deliberate" | — |

**Rule 6 gaps** (non-obvious change, no rationale comment):
- `advisor.js` → `navigating` — comment says the panel closes and the router navigates, but not why the dock must not take focus; and the UA overrides it anyway (F3). Either drop the flag or state the intended focus target.

**Rule 6 vacuity:** none — `STEP`, `MAX_FIGURES`, `session`, `PIN_SLACK`, `finalHTML`, `fillPositions` clamp, CSS backdrop literal and `.k-adv` box comments each name the rejected alternative or constraint.

**Rationale promotion:** none needed — patch records no load-bearing rationale outside code comments; A1/A10 provisional choices are sprint-nav/def material, not shipped-doc rationale.

**Commit composition** (informational — fix forward, never rewrite history): 2 code commits checked — typed subject [2/2] (`fix(advisor)`, `feat(web)`) | type matches diff [2/2] | `Projex:` trailer [2/2]. `3220b17` doc-only, exempt.
**Violations:** None.

**Trailer survival on `master`:** 0/6 — all 6 are projex doc commits or the pre-convention baseline import (`90ec1a6`); no code has landed on `master` under the convention yet (sprint unmerged). Not the dead-channel case; re-check after the sprint lands.

---

## Gap Analysis

### Promised But Not Delivered
| Promise | Status | Impact |
|---------|--------|--------|
| — (all S1–S11 delivered) | — | — |

### Undocumented Issues
| Issue | Severity | Affects |
|-------|----------|---------|
| Dialog phantom overflow → content shift on scroll-into-view | Medium | every advisor user with AT / mobile keyboard |
| Full-mode page scrollbar | Low | 1280 full page look vs `AdvisorFull` |
| `navigating` flag ineffective | Low | maintainers |
| Chip line-break splits border | Low | answer readability |
| Cross-turn `[[id]]` → 查看商品, no figure | Low | Claude mode follow-ups |

### Unhandled Edge Cases
- Mobile virtual keyboard: textarea `autofocus` opens the keyboard on every panel open (canvas `AdvisorMobile` shows no keyboard) — unverified on a device (emulation has none); plan chose `autofocus` explicitly.
- Dark backdrop invisible (light literal on dark paper) — def gap, plan A8.

---

## Quality Assessment

### Completeness: High
**Strengths:** every canvas state + both widths + both themes; H2 closed at the source with tests.
**Gaps:** designed error / spend-cap / 風格 / gallery — Deferred per plan, blockers recorded.

### Correctness: Medium-High
**Works:** SSE parse, markdown escape, figures, bar math, errors, focus, nav: Yes.
**Bugs:** phantom dialog overflow — Medium; full-mode scrollbar — Low.

### Code Quality: High
**Positive:** one module, reuses `card.js` / `util.js`, numbers single-sourced, comments name alternatives.
**Concerns:** dead focus flag.
**Tech Debt:** model API fetched whole per figure group (≤ 4, cached) — Low, already in plan Risks.

### Value Delivered: High
**Intended:** largest unbuilt canvas flow + H2. **Actual:** both. **Impact:** User: Positive | Reviewer-facing: advisor now demoable in the new shell (Mock), copy verbatim.

---

## Open Findings

### Undocumented Discoveries
- Problems: phantom overflow — Evidence: `dialog.scrollHeight` 3,375–3,577 vs ~600 client; header at y −118 after one scroll-into-view.
- Workarounds: none used.

### Impact Analysis
- Downstream: future shell overlays (notification dropdown, persona menu) inherit the same trap if they hold `position: absolute` children inside a `position: fixed` + `overflow` box.
- Risks: Claude may still cite sold figures from its own knowledge — prompt-mitigated (plan Risks) — Likelihood: Low.

### Improvements
- Could be better: per-figure `/api/models/…` payload (stats-only endpoint later).
- Missed: none material.

---

## Stakeholder Impact

| Stakeholder | Promised | Reality | Impact |
|-------------|----------|---------|--------|
| Reviewer (hosted demo) | Canvas 顧問 working in demo mode | Delivered; panel can jump if AT/keyboard scrolls it (fixable) | Positive |
| Def owner | One source per stat across UI + advisor | Advisor withholds sold stats exactly while UI shows 資料不足 | Positive |

---

## Findings

### Critical (Must Address)
- None.

### Significant (Should Address)
- **F1 Dialog phantom overflow** — sr-only abs spans escape the body clip → dialog scrolls under scroll-into-view, header vanishes, blank band. → **Fixable now:** `.k-adv__body { position: relative; }` (verified: `scrollHeight` 638 = client).

### Minor (Nice to Fix)
- **F2 Full-mode page scrollbar** → **Fixable now:** `html:has(.k-adv.is-full[open]) { overflow: hidden; }` (verified: panel 1280 wide, floating unchanged).
- **F3 `navigating` flag ineffective** (UA focus-restore wins) → **Fixable now:** drop the flag + its branch (behaviour unchanged). Choosing a different post-navigation focus target (e.g. `main`/`h1`) is a shell-wide router decision → **human/orchestrator**.
- **F4 Chip splits across lines** → **Fixable now:** `.k-adv__chip { display: inline-block; max-width: 100%; }`.
- **F5 Cross-turn references** → **Fixable now:** hoist `known` to module scope (page lifetime) so chips keep labels and figures can render for listings returned in earlier turns.
- **F6 Dark backdrop nil** → **Human decision** (canvas pass 2 / def gap 「token-driven」): needs a backdrop token for dark.
- **F7 Mobile `autofocus` keyboard on open** → **Human decision** (canvas `AdvisorMobile` shows intro without keyboard; unverified on device): keep, or focus the panel/close button on coarse pointers.

### Positive
- Executor's evidence table reproduces: every number, label and state matched on re-test.
- Gate test exercises real store writes for the ready path, not a stub.
- Escape-first markdown + `textContent` summaries: injected `<b>` rendered as text in replay.

---

## Recommendations

**Immediate:** one fix patch for F1–F5 (CSS 3 rules, JS 2 small edits); re-verify dialog `scrollHeight == clientHeight` at 1280 + 375 and full-mode width 1280.
**Future:** F6, F7 to canvas pass 2 / def § Canvas gaps tally; sprint nav to log plan provisional choices A1, A3–A10 (plan § Open Questions) as before.
**Process:** for fixed/modal containers, assert `scrollHeight == clientHeight` in the verification matrix.

---

## Final Verdict

**Status:** Accept (was Accept with Conditions; conditions met on re-audit)

**Overall Assessment:**
- Completeness: High
- Correctness: High (was Medium-High; F1/F2 fixed)
- Quality: High
- Value: High

**Conditions:**
- [x] F1 fixed (`.k-adv__body` containing block) and verified at 1280 + 375 — fix `2ab2e28` (2610130854-kura-advisor-panel-audit-fixes-patch.md); re-audited ✓ (scrollHeight = clientHeight with figures, after scroll-into-view, incl. the original ref-click trigger)
- [x] F2 fixed or explicitly accepted by orchestrator as known deviation — fixed in `2ab2e28`; re-audited ✓ (full 1280 wide, lock released on close)
- [x] F3–F7 — F3 F4 F5 fixed, F6 F7 per orchestrator decisions; re-audited ✓. Residual: router-wide post-navigation focus target (orchestrator carry-over)

**Sign-off:** Yes — S1–S11 verified directly; all audit findings closed on re-audit; no regressions.
