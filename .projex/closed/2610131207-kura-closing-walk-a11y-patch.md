# Patch: Kura closing walk + a11y pass

> **Status:** Complete
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 8
> **Author:** Claude Opus 5.5 (patch, sprint subagent)
> **Parent:** 2610131159-kura-closing-walk-a11y-plan.md
> **Directive:** Execute 2610131159-kura-closing-walk-a11y-plan.md in full (plan-execution mode): route focus + footer `aria-current`, dark pending-caption contrast (D1 accepted by orchestrator), walk matrix light/dark × 1280/375 with Appendix A probes, ≤ 6 walk fixes, def bookkeeping, Deferred ↔ def 1:1.
> **Source Plan:** 2610131159-kura-closing-walk-a11y-plan.md (all Steps 1–7)
> **Related Projex:** 2610131159-kura-closing-walk-a11y-plan.md | 2610130449-kura-mock-def.md (Canvas gaps + deviations updated) | 2610130524-kura-interactive-mock-sprint.md (Deferred table reconciled, not edited) | 2610130846-kura-advisor-panel-audit.md (router focus carry-over closed) | 2610131104-kura-methodology-footer-plan.md (footer `aria-current` carry-over closed) | 2610130938-kura-browse-page-audit.md (W1 vs its F2 scroll claim)
> **Result:** Success
> **Commits:** `e6a6c58` (Steps 1–2) · `02b68ee` (walk fixes W2, W3) · `2fc5d58` (def)

---

## Summary

Router now focuses the new view's `h1` (error view: `.k-error`) after every in-app navigation and marks the current footer link `aria-current="page"`; the advisor no longer pulls focus back to the dock when a link inside it navigates. Pending-photo caption gets token `--k-mat-note` (dark `#6B6055`, 4.89:1; light unchanged 4.71:1) — canvas deviation D1. Walk ran every Walk State at C1 light 1280 · C2 dark 1280 · C3 light 375 · C4 dark 375 with Appendix A probes on the worktree DB copy. Walk found 3 defects: W2 (account button label-in-name, A) + W3 (通知 panel stays open over the focused control after Tab-out, AA 2.4.11) fixed; W1 (Back to a long list lands short — browser scroll restore clamped by the still-shown short detail view; pre-existing) logged. Final scan: 0 overflow, 0 contrast fails, 0 console errors on every route × 4 combos; remaining small-target hits all under a WCAG 2.5.8 exception. Def: theme × layout gap → checked, 尚無行情估價 → built, 通知設定 listed, D1 under Canvas deviations. Deferred ↔ def § Canvas gaps 1:1.

---

## Changes

### Route-change focus + footer current page (Step 1)

**File:** `internal/web/static/js/app.js` — Modified
- `focusView()` (l. 55): `main.querySelector("h1, .k-error")` → `tabindex="-1"` → `focus({ preventScroll: true })`; rationale comment names rejected `<main>` focus.
- `markFooter()` (l. 64): footer `nav` links whose pathname equals `location.pathname` get `aria-current="page"`, others lose it.
- `render()` (l. 97–98): inside `mine === seq` → `markFooter()`; `if (mine > 1) focusView()` (first load left alone; stale render never focuses).

**File:** `internal/web/static/js/advisor.js` — Modified
- `let leaving` beside `busy` (l. 58) with hazard comment; close handler `if (!leaving) dock.focus(); leaving = false;` (l. 85); in-app-link click sets `leaving = true` before `dlg.close()` (l. 98); existing link comment extended.

**File:** `internal/web/static/kura.css` — Modified
- `#main [tabindex="-1"]:focus { outline: none; }` (l. 198) after `.k-error`, same convention as the 通知 / 顧問 headings.

**Why:** iter-4 carry-over (focus fell to `<body>` after navigation) + iter-7 carry-over (footer `aria-current`). The flag closes the focus race: the dialog `close` event is queued and can land after a synchronous view's render.

### Pending caption contrast in dark (Step 2, deviation D1)

**File:** `internal/web/static/kura.css` — Modified
- Header comment: token list `--k-chart-mark and --k-mat-note` (l. 3).
- `--k-mat-note: #7D7262` light (l. 57); `#6B6055` in both dark blocks (l. 73, 87).
- `.k-card__pending` → `color: var(--k-mat-note)` (l. 233) with rationale comment (ratios; rejected `--k-ink-3` / `--k-mat-ink`; token vs dark-only rule for `[data-theme]` sub-trees).

**Why:** canvas `#7D7262` on dark mat `#ECE5D7` = 3.76:1 < 4.5:1. Orchestrator accepted the fix (D1) → logged in def § Canvas deviations for canvas pass 2.

### Walk fixes W2 + W3 (Step 5)

**File:** `internal/web/static/js/notify.js` — Modified
- W2: `show()` sets `.k-acct` `aria-label` = `${m.name} · 我的帳號` (l. 82) — name (visible at 1280; its initial visible at 375) leads; aria only, no visible copy; rationale comment names rejected fixed 「我的帳號」.
- W3: `wrap` `focusout` (l. 165): open panel closes when focus moves to a target outside bell + panel; `relatedTarget` null (window blur) leaves it open; rationale names rejected Tab-trap (non-modal popover).

### Def bookkeeping (Step 6)

**File:** `.projex/2610130449-kura-mock-def.md` — Modified (commit `2fc5d58`)
- Canvas deviations += D1 line. Canvas gaps: Member layer += 通知設定; 「尚無行情估價」 → built; theme × layout → checked (pointer to plan + this patch). Revision Log row. Last Revised unchanged (same day). No Open Question checkbox touched.

---

## Fixtures (Step 3)

Server: scratch binary, `serve -db <scratch copy of worktree kura.db*> -addr 127.0.0.1:8781` (PID recorded, stopped by PID); empty DB `:8782`. `/api/overview` → `advisor: "mock"`. DB: 15,367 listings, 0 sold.

| State | Fixture |
|---|---|
| D1 fair + distribution | `/listing/brandoff%3A2101220216016` (LV M51252 カルトシエールGM, high, 16 comps, 低於行情 50%) |
| D2 no estimate | `/listing/brandoff%3A2101219731339` (GOYARD, 尚無行情估價) |
| D3 anomaly | `/listing/brandoff%3A2101220184483` (異常低價 · 先確認狀況; 80 anomaly listings in DB) |
| D4 low confidence | `/listing/brandoff%3A2100301255937` (可能低於行情 · 低信心) |
| D5 pending photo | `/listing/brandoff%3A2104102531481` (`photo_pending`, 9 in DB) + `/dev/cards.html` |
| M1 model | `/model/LOUIS%20VUITTON/M51252` (25 在售, show-more 5 → 25) |
| B4 zero result | `/browse?q=zzzz-no-match` |
| H2 empty | `:8782` new DB in scratchpad |
| Sold card (mat `opacity:.6`) | none in DB (0 sold) → n/a on real data |

---

## Walk method + tooling notes

- Combos via `resize_window` + `colorScheme`; A.1 probe each combo: C1 `rgb(245, 240, 230)`/1280 · C2 `rgb(27, 24, 21)`/1280 · C3 light/375 · C4 dark/375. **A-1 holds** (emulation flips `prefers-color-scheme`; `data-theme` null throughout).
- Appendix A probes A.2–A.6 verbatim, installed per origin; plus a focus logger (`focusin` → ring / on-screen / zone per stop) for K, an accessible-name probe (no-name + label-in-name), and a route walker (in-app `<a>` click through the router → A.2–A.5 per route).
- Navigation: real keyboard (Tab / Enter / Esc / Shift+Tab via `computer key`) for K, ring and Esc paths; `element.click()` / in-app link clicks for the rest (method deviation, plan-allowed).
- **Hidden-pane artefact:** with the pane not drawing, Chrome defers the dialog `close` event until the next rendering update — Esc/A4 focus assertions read before a frame show the pre-event state. All A4/A5 results below were read **after a forced frame** (screenshot), with a single pending close event. Repeated `Tab` presses (`repeat`) were occasionally dropped under the same condition; K results count actual focus stops, not presses.
- Console: `read_console_messages` per combo; only entry ever seen = the walker's own `kprobe-console-check` capture test → 0 app errors / warnings.
- Screenshots (supplementary): browse dark 1280, advisor dark 1280 (turn + backdrop), 通知 dark 1280, detail dark 375, detail light 375 — all token-correct. Model dark 1280 partial render only (pane scaling); DOM metrics are evidence of record.

---

## Walk Matrix

Cell = result · `sw` (page `scrollWidth`; 1265 at 1280 = scrollbar) · `f` contrast fails · `s` small targets not excepted (kind) · `e` console errors. Behaviour assertions full at C1 + C4; C2 / C3 reach + scan. C1 walk ran as guest, C2–C4 as member 林小姐.

| Id | State | C1 light 1280 | C2 dark 1280 | C3 light 375 | C4 dark 375 |
|---|---|---|---|---|---|
| H1 | Home | ✓ sw1265 f0 s0 e0 · 最新上架 5 / 價格最美 3; first load focus `BODY` | ✓ sw1265 f0 s0 e0 | ✓ sw375 f0 s0 e0 · rails `overflow-x:auto`, scroll in own rail | ✓ sw375 f0 s0 e0 · first load `BODY`; rails own-scroll |
| H2 | Home empty `:8782` | ✓ sw1280 f0 s0 · 「還在準備中。」, 0 shelves | ✓ sw1280 f0 s0 | ✓ sw375 f0 s0 | ✓ sw375 f0 s0 |
| B1 | 看全部價格最美 → Browse | ✓ 1,471 = Home 1,471; focus `h1` | ✓ f0 s0 | ✓ f0 s34ᵃ | ✓ Home N = 1,471; focus `h1` |
| B2 | filters / sort / budget / deals / 再看 / Back | ✗ W1 · URL per change ✓; 40 → 80, focus first new `.k-card__model` ✓; Back 80 cards ✓, scroll 2400 → 1434 | ✓ f0 s0 (80 cards) | ✓ f0 s59ᵃ (80) | ✗ W1 · 80 cards ✓, scroll 3000 → 2040; more-focus ✓ |
| B3 | 375 篩選與排序 | n/a (desktop form; toggle mobile-only) | n/a | ✓ `aria-expanded` true → false; Tab: deals → 品牌 → 類型 → 成色 → 預算 → 排序 | ✓ `aria-expanded` true → false; sw375 |
| B4 | zero result | ✓ 「沒有符合條件的包。」 f0 s0 | ✓ | ✓ | ✓ |
| B5 | header search | ✓ `/browse?q=Speedy`, focus `h1` | ✓ `/browse?q=Kelly`, `h1` | ✓ icon (`display:flex`, form none) → `/browse`, `h1` | ✓ same as C3 |
| D1 | detail fair + dist | ✓ 鑑定卡 / 行情估價 / 行情怎麼算 → `/methodology` (focus `h1`, footer current) / 以表格檢視 ⇄ 以圖表檢視 / dot tip 「ランクA · 在售NT$16,147」 / 售價變化 (1 point → prose) / CTA `_blank` + `noopener noreferrer` | ✓ tip + table f0 | ✓ tip + table f0 sw375 | ✓ tip + table f0 sw375 |
| D2–D4 | none · anomaly · low | ✓ 尚無行情估價 · 異常低價 · 先確認狀況 · 可能低於行情 · 低信心; prose matches | ✓ f0 | ✓ f0 | ✓ f0 |
| D5 | pending photo (S3) | ✓ `rgb(125, 114, 98)` on `#FFF` 4.71 | ✓ `rgb(107, 96, 85)` on `rgb(236, 229, 215)` 4.89 | ✓ 4.71 | ✓ 4.89 |
| M1 | model | ✓ 4 tiles (在售 25 件 · 已售出 資料不足 · …), 追蹤這個型號 `aria-pressed`, dist table toggle, show-more 5 → 25 + focus first new card | ✓ f0 s0 | ✓ f0 s0 | ✓ f0 s0 |
| A1 | advisor open | ✓ 示範模式 visible; focus composer `TEXTAREA`; f0 | ✓ composer; backdrop `color(srgb .082 .075 .063 / .6)` | ✓ focus heading `H2`; f0 | ✓ heading `H2`; dark backdrop |
| A2 | chip prompt → mock turn | ✓ step line 「找到 684 件…」, 4 figures, 4 position bars, 6 chips; pinned (0 px from end); s6ᵇ | ✓ 4/4 s6ᵇ | ✓ 4/4 s6ᵇ | ✓ 4/4 s6ᵇ |
| A3 | full ⇄ floating | ✓ 放大成整頁 ⇄ 縮小成視窗 | ✓ | ✓ | ✓ |
| A4 | rec link → detail | ✓ 看詳細 / chip / name → closed, detail `h1`; sync-route race (`/methodology` link in panel) → `h1` after close event | (scan only) | (scan only) | ✓ 看詳細 → closed, `h1` after close event |
| A5 | Esc | ✓ dock (`k-adv-dock`) | — | — | ✓ dock |
| N1 | 登入 → 林小姐 | ✓ badge 2, 「通知，2 則未讀」, focus bell | ✓ f0 | ✓ f0 | ✓ f0 |
| N2 | 通知 open / Esc | ✓ heading focused; Esc → bell; ✗ W3 (Tab-out leaves panel open) | ✓ heading f0 | ✓ heading f0 | ✓ heading; Esc → bell; ✗ W3 (focused control under sheet, `elementFromPoint` = `.k-notif__meta`) |
| N3 | 全部標為已讀 | ✓ badge hidden, label 「通知」 | — | — | ✓ |
| N4 | notification link | ✓ `/listing/…315146`, panel closed, `h1` | — | — | ✓ same |
| N5 | track toggles | ✓ detail / card / model `aria-pressed`; 追蹤清單 4 → 5 → 6 → 7; card synced with detail; reload keeps session | — | — | ✓ same counts |
| P1 | 行情怎麼算 | ✓ `h1`, footer current = 行情怎麼算; 5 sections / 6 badge rows | ✓ | ✓ | ✓ |
| P2 | 5 pending pages | ✓ `h1` = link text, footer current = that link (each) | ✓ | ✓ | ✓ |
| P3 | `/nope` | ✓ in-app `h1`; direct load: not-found, focus `BODY` | ✓ | ✓ | ✓ |
| K | keyboard pass | ✓ full: Home 40 / Browse 39 / detail 38 / model 34 / meth 22 / pending 14 stops — all ring `rgb(107, 31, 42)`, on screen, order header → main → footer → dock → header; Enter on footer link → `h1` focused, `:focus-visible` true, outline none | ✓ spot 3 stops × 6 routes, ring `rgb(227, 161, 166)` | ✓ spot × 6 + 篩選 panel, ring `rgb(107, 31, 42)` | ✓ full: Home 40 / Browse 29 / detail 20 / model 31 / meth 18 / pending 16, ring `rgb(227, 161, 166)`; advisor modal: 30 Tabs stay in dialog; 通知 non-modal (Tab leaves → W3) |
| S2 | footer `aria-current` | ✓ 0 on `/`, Browse, detail, model, not-found; 1 on each footer page | ✓ | ✓ | ✓ |
| /dev | `/dev/cards.html` caption | ✓ light section 4.71, dark section 4.89 | ✓ light sub-tree under dark OS keeps `rgb(125, 114, 98)` 4.71; dark 4.89 | — | — |

ᵃ `.k-card__model` 211 × 22 px in 375 row cards — WCAG 2.5.8 **spacing exception**: nearest other target ≥ 93 px from its centre (heart 131 px, shop link 93 px) → not a finding.
ᵇ `.k-adv__chip` 22 px, `inline-block` inside an answer sentence (line-height 27.75 px) → 2.5.8 **inline exception** → not a finding.
Contrast `manual` = 1 on every detail page: `.k-tag__stamp` 「店鑑」 10.68:1 (translucent layer) → pass by eye/value.

---

## Findings

| Id | Severity | Combos | Evidence | Disposition |
|---|---|---|---|---|
| W1 | Flow (not WCAG) | C1, C4 measured; all (code path) | Browse scrolled to 2400 (C1) / 3000 (C4) → card → Back: 80 cards restored but `scrollY` 1434 / 2040 = detail page's max scroll. `popstate` + `setTimeout 0` sample: already clamped while the detail `h1` is still on screen → browser restores at `popstate`, before the async render. **Pre-existing:** same result on a build of `ab33403` (before this patch): 2400 → 1476 = detail max. Not caused by heading focus (`preventScroll`). Contradicts browse audit F2's 「scroll restored (≈ 12,100px)」. | **Logged** — needs an approach decision (router-owned restoration: `history.scrollRestoration = "manual"` + `y` in `history.state` written on leave, restored after render; Forward entries left via `popstate` carry no `y`) — not a single-approach patch; touches every route. Follow-up objective for derive 9. |
| W2 | A (2.5.3 Label in Name) | C1–C4 (member) | `.k-acct` `aria-label="我的帳號"`; visible 「林小姐」 (1280) / initial 「林」 (375) not in the name | **Fixed** `02b68ee` — re-scan C1–C4: `aria-label` 「林小姐 · 我的帳號」, name probe 0 label-in-name / 0 no-name on Home, Browse, detail, model, `/methodology`. Account menu itself stays Deferred (button still inert). |
| W3 | AA (2.4.11 Focus Not Obscured) | C4 measured; same code all widths | 通知 open → Tab past 通知設定 → focus moves to header then `#main` while the panel stays open; C4: focused main link centre hits `.k-notif__meta` (sheet `position:fixed` 64–392 px) → focused control hidden | **Fixed** `02b68ee` — re-scan: C1 / C2 Tab-out → focus `k-watch`, panel closed; C3 / C4 → `k-acct`, closed, `obscured=false`; bell → heading → bell keeps it open; Esc → bell and notification link → `h1` still pass (C4). |

Not findings (plan's known candidates): `.k-switch` checkbox → label box ≥ 24 px (A.4 clean) | pending pages' 40 px `h1` at 375 → no overflow (sw 375) | native controls in dark → `color-scheme: dark`, ink `rgb(238, 231, 219)` on `rgb(36, 32, 28)` | sold mat `opacity:.6` → no sold listing in DB (n/a; canvas-designed anyway).

**Fix count:** 2 walk fixes (cap 6). **Over cap:** none. **Open human questions touched:** none.

---

## Final Scan (after `02b68ee`)

Every route (H1 `/`, B1, B4, D1–D5, M1, P1 `/methodology`, P2 ×5, P3 `/nope` = 16 routes) per combo, plus A1 / A2 / A3 / N1 / N2 / B2 / B3 / D1 tip+table / D5:

| Combo | Routes | max `sw` / vw | overflow | contrast fails | small (non-excepted) | all view-heading focus | console errors | Open states |
|---|---|---|---|---|---|---|---|---|
| C1 light 1280 | 16 | 1280 / 1280 | 0 | 0 | 0 | ✓ | 0 | A1 / A2 / N1 / N2 / B2 f0 s0 |
| C2 dark 1280 | 16 | 1280 / 1280 | 0 | 0 | 0 | ✓ | 0 | A1 / A2 / A3 / N1 / N2 / B2 f0 s0; D5 4.89 |
| C3 light 375 | 16 | 375 / 375 | 0 | 0 | 0 (Browse 34 = spacing exc.) | ✓ | 0 | W2 / W3 re-check ✓ |
| C4 dark 375 | 16 | 375 / 375 | 0 | 0 | 0 (Browse 34 / B2 59 = spacing exc.; chips 6 = inline exc.) | ✓ | 0 | A1 / A2 / A3 / N1 / N2 / B2 / B3 f0; D5 4.89 |

Remaining logged: W1 only (flow, not a contrast / target / overflow fail).

---

## Acceptance Criteria

| Criterion | Result | Evidence |
|---|---|---|
| S1 Route focus | ✓ | Matrix B1 / B5 / D1 / A4 / N4 / P1–P3 / back + forward (C1 `/methodology` Back → Forward `h1`); first load `BODY` (C1, C4, direct `/nope`); `scrollY` 0 after link; Back keeps restored position (Home 400 / 500); outline none on focused `h1` even with `:focus-visible` (keyboard Enter); A4 sync-route race holds after the queued close event; Esc → dock, Esc → bell |
| S2 Footer current | ✓ | 1 link on each of 6 footer pages, 0 on `/`, Browse, detail, model, not-found; updates per navigation, C1–C4 |
| S3 Pending caption AA | ✓ | Real listing + `/dev/cards.html`: light 4.71 (`rgb(125, 114, 98)` unchanged), dark 4.89 (`rgb(107, 96, 85)`); light sub-tree under dark OS keeps light value |
| S4 Walk matrix | ✓ | Every state × C1–C4 recorded (✓ / ✗ W-id / n/a + reason / — = scan-only combo per plan) |
| S5 Findings disposed | ✓ | W1 logged (approach decision), W2 + W3 fixed + re-scanned; 2 ≤ 6; no Go / API / visible copy / canvas-combo token value touched (D1 per R3 and orchestrator) |
| S6 Final scan | ✓ | Table above: 0 overflow, 0 console errors, 0 contrast fails; residual small targets all excepted |
| S7 Def updated | ✓ | `2fc5d58`; reconciliation table below 1:1 |
| S8 Checks | ✓ | `go vet ./...` clean; `go test -count=1 ./...` ok (advisor, listing, radar, brandoff, store, web); hygiene grep on `git diff ab33403 -- internal/web/static` → 0 hits; servers stopped by PID (ports 8781–8784 free); scratch binary / DB copies deleted; worktree clean besides gitignored `kura.db*` (sha1 unchanged) |

---

## Deferred ↔ def § Canvas gaps reconciliation

| Def § Canvas gaps item | Status |
|---|---|
| Multi-provider (7 sub-items) | Deferred row "Multi-provider canvas gaps" |
| Member layer — watchlist page | Deferred row "watchlist page" |
| Member layer — account / persona menu | Deferred row "account / persona menu + persona switching" |
| Member layer — sign-in-as-persona (登入 / 加入會員) | Deferred row "sign-in-as-persona UI / 加入會員" |
| Member layer — 通知設定 (added this patch) | Deferred row "通知設定" |
| Advisor — error state | Deferred row "Advisor error state" |
| Advisor — spend-cap-reached state | Deferred row "Advisor spend-cap-reached state" |
| Pages — 行情怎麼算 content | built (iter 7, provisional) |
| Pages — 關於 Kura / 常見問題 / 聯絡我們 / 隱私權政策 / 服務條款 content | Deferred row "Footer page content" |
| Pages — 資料管線 | Deferred row "資料管線 pipeline page" |
| HANDOFF §4.A — header status line + 近期降價 | Deferred row "header status line + 近期降價" |
| Card 「尚無行情估價」 | built (iter 1; line updated this patch) |
| Theme × layout combos with no artboard | checked (this patch) |

Reverse: 10 Deferred rows → each maps to exactly one def item above (multi-provider row covers the def's 7-item line). **1:1, no mismatch.** Sprint nav not edited.

---

## Deviations from Plan

1. **Served a scratch copy of `kura.db*`** (scratchpad), not `-db kura.db` in the worktree — orchestrator rule "don't mutate it"; worktree DB sha1 identical before / after.
2. **Extra A/B server** (`ab33403` build, `:8783`, stopped by PID, deleted) to prove W1 pre-existing; and a short empty-DB server (`:8784`) after the main walk to record P3 direct load + P1 section count on the final build.
3. **Def § Canvas deviations edited** (D1 line) — outside the plan's def scope (Canvas gaps + Revision Log); orchestrator: "log it as a canvas deviation".
4. **W1 blocker category** — "needs approach decision (router scroll restoration)", not one of the plan's four (canvas pass 2 | open question | over cap | Deferred): no listed category fits a pre-existing flow defect with more than one viable fix.
5. **C1 walk as guest, C2–C4 as member** — both header states covered across combos rather than per combo.
6. Commit grouping: Steps 1–2 one commit (both touch `kura.css`); W2 + W3 one commit.

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|---|---|---|
| 2610131159-kura-closing-walk-a11y-plan.md | Source plan | All steps `[PATCHED]` → this patch; Status Complete; moved to `closed/` |
| 2610130449-kura-mock-def.md | Def gaps / deviations | Canvas gaps (3 lines) + Canvas deviations (D1) + Revision Log (`2fc5d58`) |
| 2610130524-kura-interactive-mock-sprint.md | Sprint nav | Not edited (absorb owns it); reconciliation table above + W1 for derive 9 |
| 2610130938-kura-browse-page-audit.md | F2 "scroll restored" claim | Not edited — W1 records the contrary measurement |

---

## Notes

- W1 suggested shape (for derive 9, not decided here): `history.scrollRestoration = "manual"`; `go()` writes `{ ...history.state, y: scrollY }` onto the leaving entry (keeping Browse's `shown`); router restores `history.state.y` after a `popstate` render. Open point: entries left via Back/Forward never get `y` written unless a scroll listener updates state.
- `leaving` pairs with the next `close` event. With two `close` events pending at once (only in a non-rendering pane) the second runs `dock.focus()`; unreachable with a drawing page, where each `close` dispatches within a frame.
- Model watch toggle increments 追蹤清單（N） (6 → 7) — existing behaviour (watchlist holds listings and models), not judged here.
