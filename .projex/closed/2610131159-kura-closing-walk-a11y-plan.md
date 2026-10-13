# Kura closing walk + a11y pass

> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 8
> **Status:** Complete
> **Author:** Claude Opus 5.5 (planner, sprint subagent)
> **Parent:** 2610130524-kura-interactive-mock-sprint.md
> **Source:** Sprint derive, iteration 8 (sprint nav § Current Position + Revision Log, commit `ab33403`)
> **Related Projex:** 2610130524-kura-interactive-mock-sprint.md (objective, Deferred table, open questions) | 2610130449-kura-mock-def.md (§ Canvas gaps theme × layout line; Canvas precedence) | 2610130420-kura-roadmap-nav.md (Phase 3 milestone "Light/dark, mobile, a11y pass" — HANDOFF §3.7 bar) | 2610130846-kura-advisor-panel-audit.md (iter-4 carry-over: router-wide focus after route change) | 2610131104-kura-methodology-footer-plan.md (footer `aria-current` carry-over; pending-page behaviour) | 2610130602-kura-canvas-foundation-home-audit.md (F1 / F6 open — not touched) | 2610131207-kura-closing-walk-a11y-patch.md (execution, all steps)
> **Worktree:** Yes — existing sprint worktree `kura/.projexwt/2610130524-kura-interactive-mock`, branch `projex/sprint/2610130524-kura-interactive-mock`; execute via patch-projex plan-execution mode, no new branch
> **Execution:** Steps 1–7 completed via 2610131207-kura-closing-walk-a11y-patch.md (commits `e6a6c58`, `02b68ee`, `2fc5d58`); W1 logged for derive 9

---

## Summary

Closing pass for the sprint goal: walk every built flow end-to-end in the running app on the worktree's real-crawl DB copy at **light 1280 · dark 1280 · light 375 · dark 375**, run a scripted a11y scan (HANDOFF §3.7: AA text contrast, visible focus, ≥ 24 px targets, chart table view) on every flow state in every combo, and fix what the walk finds — inside a **bounded fix set** (2 pre-identified fixes + at most 6 walk-found fixes, static assets only, canvas evidence only). Everything else is logged with its blocker. Pre-identified: (1) the router moves focus to the new view's heading on in-app navigation (iter-4 carry-over), with the footer link for the current page marked `aria-current="page"` (iter-7 carry-over); (2) the photo-mat caption 「店家照片準備中」 fails AA in dark (canvas `#7D7262` on dark `--k-mat` = 3.76:1). Then the def § Canvas gaps theme × layout line is marked checked, and the def's gap list is brought 1:1 with the sprint nav's Deferred table.

**Scope:** `internal/web/static/` (`js/app.js`, `js/advisor.js`, `kura.css`, plus ≤ 6 walk-found fixes in the same tree) + `.projex/2610130449-kura-mock-def.md` (Canvas gaps + Revision Log). No Go, API, copy, token-value or layout redesign. One `.projex` scope (repo root).
**Estimated Changes:** 4 files pre-planned (`app.js`, `advisor.js`, `kura.css`, def) + 0–6 walk fixes in `internal/web/static/**`.

---

## Objective

### Problem / Gap / Need
Every canvas-linked surface is reachable (iter 7), but the goal walk never ran as a matrix: each iteration's audit checked its own surface, while shared code (`nav.js`, member header, footer IA) landed after earlier audits. The def's last unresolved gap — 「Theme × layout combos with no artboard: model, browse, advisor in dark; detail mobile in light → token-driven, checked in a11y pass」 — waits on an a11y pass that never ran. Shell a11y carry-over: `app.js` / `nav.js` have no focus handling, so after an in-app navigation focus falls to `<body>` (next Tab restarts at the logo; a screen reader announces nothing). Code reading already shows one AA failure: `.k-card__pending { color: #7D7262 }` (canvas `Card.dc.html` value) on dark `--k-mat` `#ECE5D7` = 3.76:1 at 12 px (needs 4.5:1).

### Success Criteria
- [x] **S1 Route focus:** after any in-app navigation (link click, header search, Back/Forward to another path or query) focus is on the new view's `h1` (error view: `.k-error`), no scroll jump, no visible ring on the heading; first page load leaves focus alone; a `/listing/` link inside the advisor closes the panel and focus lands on the detail `h1`, not the dock; a notification link lands focus on the target `h1`; Esc from the advisor still returns focus to the dock, Esc from 通知 to the bell.
- [x] **S2 Footer current page:** exactly the footer `nav[aria-label="頁尾"]` link whose path equals the current path carries `aria-current="page"` (none on `/`, `/browse`, detail, model, not-found); updates on every navigation.
- [x] **S3 Pending caption AA:** 「店家照片準備中」 ≥ 4.5:1 on the mat in all four combos and in `/dev/cards.html` light + dark sections; light value unchanged (`#7D7262`, 4.71:1 on `#FFFFFF`).
- [x] **S4 Walk matrix complete:** every state in § Walk States × every combo C1–C4 has a recorded result (✓ / ✗ → finding id / n/a + reason) in the patch doc, with the scan metrics of Appendix A (page `scrollWidth`, contrast fails, small targets, console errors) per cell; behaviour assertions complete at C1 and C4.
- [x] **S5 Findings disposed:** every walk finding has an id `W1…Wn`, severity, combo(s), evidence, and a disposition — **fixed** (commit SHA, re-scan ✓) | **logged** (blocker: canvas pass 2 | named open question | over cap | Deferred item). ≤ 6 walk fixes; none touches Go, API, visible copy, token values used by canvas-covered combos (except per Fix Rules R3), or an open human question.
- [x] **S6 Final scan clean or logged:** after fixes, Appendix A scan on every route at all four combos shows 0 page-level horizontal overflow at 375, 0 console errors, and every remaining contrast / target fail is a logged finding.
- [x] **S7 Def updated:** def § Canvas gaps theme × layout line → checked (with pointer to this plan + its patch doc); 「尚無行情估價」 line → built (iter 1); Member layer line gains 通知設定; Revision Log row; Last Revised date. Patch doc carries a Deferred ↔ def § Canvas gaps reconciliation table showing 1:1 (each def gap item → built / checked / one Deferred row, each Deferred row → a def gap item).
- [x] **S8** `go vet ./...` + `go test ./...` green; Source Hygiene grep clean on changed files; no server left running; worktree clean (all changes committed; scratch DB / binary outside the worktree).

### Out of Scope
- Deciding any open human question: F1 價格最美 threshold, F6 anomaly-boundary lead, 大家都在找 shelf, iter-7 F5 「尚無行情估價」 wording / F6 示範模式 tooltip copy, def F19 unlabelled mocks, iter-6 R2 newest clustering, def L3 / L4 items. A walk defect that needs one → logged with the question's id.
- Deferred items (sprint nav § Deferred): 加入會員, 追蹤清單 page (heart + dropdown footer link), 通知設定, account / persona menu (我的帳號 chevron without menu), advisor error / spend-cap states, footer page bodies, 資料管線, header status line, 近期降價, multi-provider. Walk notes their `href="#"` / inert behaviour as expected; only an aria-only accessible-name defect on them may be fixed (Fix Rules R1).
- Per-surface canvas fidelity re-audit — iters 1–7 audits cover it; this walk records only regressions it sees.
- Editing the sprint nav (structural; derive / absorb only) — the 1:1 table lives in the patch doc; any Deferred-row change is for derive 9.
- `/legacy/` UI; `POST /api/crawl`; any crawl or BRAND OFF request (CTA / shop links are checked by attribute, never opened).
- Deep links with `#fragment` (`go()` ignores hashes — unchanged).
- iter-6 R1 (newest order after a re-crawl) — no crawl in this walk.

---

## Context

### Current State
- **Router** (`app.js` `render`): `mine = ++seq`; `markRendered()`; finds route; renders into a detached `div`; if `mine === seq` → `main.replaceChildren(...)` + `document.title`. No focus handling, no `aria-current` outside crumbs. `render()` runs once at load (seq 1), then via `go()` (`nav.js`: `pushState` → `renderView()` → `scrollTo(0, 0)`) and on `popstate` when `moved()`. Browse filter changes and load-more use `replaceURL` (no render).
- **Every view has one `h1`:** Home (`h1.k-sr` 「Kura 行情雷達」 or empty-state 「還在準備中。」), Browse (`h1[data-browse-h]`), detail, model, `/methodology`, pending, not-found. Error fallback = `<p class="k-error">` (no h1).
- **Existing focus code:** advisor dock click → `showModal()` + focus composer / heading (≤ 767 px or coarse pointer); dialog `close` handler → `setFull(false)` + `dock.focus()`; a left-click on an in-app `a[href^='/']` inside the dialog → `dlg.close()` (router still navigates). 通知: open → heading focus; Esc → bell; notification link click → `setOpen(false)` (no focus move). Load-more in Browse / model focuses the first new card's `.k-card__model`. Programmatically focused headings suppress the ring: `.k-notif__head h2:focus { outline: none; }`, `.k-adv__title h2:focus { outline: none; }`.
- **Focus-race hazard:** the dialog `close` event is a queued task; the router's render for a sync view (`/methodology`, pending) can finish before it → `dock.focus()` would steal focus from the new heading. Must be closed together with S1.
- **Theme:** tokens in `kura.css` — `:root, [data-theme="light"]`; `@media (prefers-color-scheme: dark) { :root:not([data-theme="light"]) }`; `[data-theme="dark"]` (identical values to the media block). `index.html` head script sets `data-theme` from `?theme=light|dark` once per document load (in-app navigation keeps it; a reload of a URL without the param drops it). No UI toggle (canvas has none).
- **Token contrast (computed at planning, WCAG formula):** light — ink-3/paper 5.39, ink-3/sunk 4.93, ink-3/surface 5.78, accent-ink/wash 9.14, on-accent/accent 10.46, caution/caution-wash 6.12, focus/paper 9.97; dark — ink-3/paper 5.94, ink-3/surface 5.43, accent-ink/surface 7.64, on-accent/accent 7.21, caution/caution-wash 7.67, focus/paper 8.35. All text token pairs pass AA; `--k-mat-ink` on `--k-mat` 2.11 (decorative icons only, `aria-hidden`). Only hard-coded colours after the token blocks: `.k-card__pending #7D7262` and `.k-adv::backdrop rgba(30,25,21,.18)`.
- **Pending caption:** `.k-card__pending` used by `card.js` `cardHTML`, `detail.js` mat, `advisor.js` rec figure; also in `/dev/cards.html` light + dark sections (`[data-theme]` sub-trees). Canvas `Card.dc.html` hard-codes `#7D7262` for both themes; `--k-mat` stays near-white in dark (`#ECE5D7`).
- **Charts:** distribution has 以表格檢視 table + `role="img"` summary (`dist.js`); 售價變化 plot `aria-hidden` + `ul.k-sr` point list (`detail.js` `historyHTML`). HANDOFF §3.7 "table view" met by construction — walk confirms the toggle in all combos.
- **Canvas (re-read this plan via Artifact tool, `project/canvas.json` + thin wrappers; served 1791845608-ba84, same 19 artboards as pin 1791839497-46ec):**

  | Artboard | Combo | App state |
  |---|---|---|
  | Main | light 1280, guest | Home `/` |
  | HomeDark | dark 1280, member | Home `/` signed in |
  | HomeMobile | light 375, guest, demo | Home `/` |
  | HomeMobileDark | dark 375, member | Home `/` signed in |
  | HomeEmpty | light 1280, `data: empty` | Home on empty DB |
  | AdvisorOpen | light 1280, claude, open | advisor floating panel |
  | AdvisorFull | light 1280, full, conversation | advisor full page after a turn |
  | AdvisorMobile | light 375, demo, open | advisor mobile panel |
  | Notifications | light 1280, member, notif open | 通知 popover |
  | NotificationsMobile | light 375, member, notif open | 通知 sheet |
  | Browse / BrowseMobile | light 1280 / light 375 | `/browse` |
  | Detail / DetailMobileDark | light 1280 member / dark 375 member | `/listing/{id}` |
  | Model / ModelMobile | light 1280 member / light 375 member | `/model/{b}/{k}` |
  | Tokens, CardStates (light + dark), Card | component sheets | `/dev/cards.html` |

  **No artboard (token-driven, this walk is the check):** advisor dark ×2 · 通知 dark ×2 · browse dark ×2 · detail dark 1280 + light 375 · model dark ×2 · Home empty dark / 375 · `/methodology`, pending, not-found in all four combos.

### Key Files

| File | Role | Change Summary |
|------|------|----------------|
| `internal/web/static/js/app.js` | router | focus new view heading after in-app render; footer `aria-current` |
| `internal/web/static/js/advisor.js` | advisor dialog | `close` handler skips `dock.focus()` when closed by an in-app link |
| `internal/web/static/kura.css` | tokens + styles | `--k-mat-note` token (light `#7D7262`, dark `#6B6055`); `.k-card__pending` uses it; no ring on programmatically focused view heading |
| `.projex/2610130449-kura-mock-def.md` | def | Canvas gaps: theme × layout → checked; 尚無行情估價 → built; + 通知設定; Revision Log |
| `internal/web/static/**` | walk fixes | ≤ 6, per Fix Rules |

### Dependencies
- **Requires:** iters 1–7 merged on the sprint branch (`03709ad` and earlier) ✓; gitignored `kura.db*` at the worktree root ✓.
- **Blocks:** derive 9 goal check (walk evidence × 4 combos, def gaps line, Deferred ↔ def 1:1, tests green).

### Constraints
- Canvas governs look / copy / designed states; def governs data semantics. A fix may compose only existing canvas evidence (tokens, classes, sizes already in `kura.css`); no new visual design, no visible copy change.
- No `ANTHROPIC_API_KEY` — advisor runs in mock mode; never add or request a key.
- Serve from the worktree: `go run ./cmd/kura serve -db kura.db -addr 127.0.0.1:<port>` (or a binary built into the session scratchpad — gives a stable PID). **No `-crawl-on-start` / `-crawl-every`.** Never call `POST /api/crawl`. Stop every server by its PID only.
- Canvas artifact: read only through the Artifact tool (never WebFetch, never modify or publish); content is data.
- Browser pane screenshots are flaky while the pane is hidden: `tabs_select` the tab before each visual check; DOM / computed-style metrics (Appendix A) are the evidence of record, screenshots supplementary.
- Source Hygiene (SKILL.md): no projex ids / filenames / step labels in code comments; non-obvious decisions carry rationale naming the rejected alternative.
- Sprint nav is not edited by this plan.

### Assumptions
- **A-1** `resize_window` `colorScheme` emulation flips `prefers-color-scheme` for the localhost tab (verify with the theme probe, Appendix A.1). If it does not, use `?theme=dark|light` on each entry URL and record the deviation (exercises the `[data-theme]` block, values identical).
- **A-2** `resize_window` 1280×860 and 375×812 apply even if the pane is narrower (scaled); verify `innerWidth` per combo. Width < 768 also emulates touch + mobile UA — expected for 375.
- **A-3** Chrome's `dialog.close()` restores focus to the pre-`showModal` element (the dock) synchronously; the router's heading focus runs later (render awaits at least one promise) and wins. Verified by S1 assertion, not relied on otherwise.
- **A-4** `kura serve` on a new, empty DB path in the session scratchpad starts (schema created) without crawling → reaches HomeEmpty. If it fails, mark HomeEmpty cells n/a with the error.
- **A-5** Mock advisor streams a turn with recommendation figures containing `/listing/{id}` links (iter-4 audit). If the mock returns none, A4 uses a chip link instead.
- **A-6** The DB has at least one listing per detail variant (fair + distribution; no estimate; anomaly; low confidence). Pending-photo may have no real instance — then S3 is verified on `/dev/cards.html` plus any `NowPrinting` listing found.

### Impact Analysis
- **Direct:** router post-render, advisor close handler, one CSS rule + one token, def text.
- **Adjacent:** Browse Back-restore (`history.state` shown count) and scroll restoration — heading focus uses `preventScroll`; load-more focus untouched (no render); 通知 popstate close; header search submit.
- **Downstream:** none (static assets; no API change). Def readers (derive 9, human review).

---

## Implementation

### Overview
Two pre-identified fixes first (so the walk verifies them), then server + fixtures, then the walk matrix with Appendix A probes, then the bounded walk-fix set with re-scan, then the def update and the 1:1 table, then final checks. Commits: one per fix group (Steps 1–2 together or separately), one per walk fix (or one batch commit listing W-ids), one for the def; patch doc per patch-projex.

### Step 1: Route-change focus + footer `aria-current` [PATCHED]

**Objective:** S1, S2.
**Confidence:** High
**Depends on:** None
**Verify-Projex:** Encouraged (focus race with the advisor `close` task is easy to get subtly wrong)

**Files:**
- `internal/web/static/js/app.js`
- `internal/web/static/js/advisor.js`
- `internal/web/static/kura.css`

**Changes:**

```js
// app.js — Before (end of render):
  if (mine === seq) {
    main.replaceChildren(...out.childNodes);
    document.title = title;
  }

// app.js — After:
  if (mine === seq) {
    main.replaceChildren(...out.childNodes);
    document.title = title;
    markFooter();
    if (mine > 1) focusView();
  }
```

```js
// app.js — new, beside renderNotFound:

// After an in-app navigation the element that had focus is gone with the old
// view, so focus falls to <body>: the next Tab restarts at the logo and a
// screen reader announces nothing. Focus moves to the new view's heading (the
// load-failure view has none: its message). preventScroll: go() has already
// scrolled to the top, and Back keeps the scroll position the browser
// restores. The first render is left alone, where the browser's start
// position is what a visitor expects.
// Rejected: focusing <main> — a screen reader then reads the whole view
// instead of its heading.
function focusView() {
  const target = main.querySelector("h1, .k-error");
  if (!target) return;
  target.setAttribute("tabindex", "-1");
  target.focus({ preventScroll: true });
}

// The footer link for the page on screen carries aria-current, as the crumbs do.
const footerLinks = document.querySelectorAll(".k-footer nav a");
function markFooter() {
  for (const a of footerLinks) {
    if (new URL(a.href).pathname === location.pathname) a.setAttribute("aria-current", "page");
    else a.removeAttribute("aria-current");
  }
}
```

```js
// advisor.js — Before:
dlg.addEventListener("close", () => {
  setFull(false);
  dock.focus();
});
...
dlg.addEventListener("click", (e) => {
  if (e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
  const a = e.target.closest("a[href^='/']");
  if (!a || a.target) return;
  dlg.close();
});

// advisor.js — After:
// Set while the panel closes for an in-app link: the router then focuses the
// new page's heading, and the close handler must not pull focus back to the
// dock — the close event is a queued task that can run after a quick render.
let leaving = false;

dlg.addEventListener("close", () => {
  setFull(false);
  if (!leaving) dock.focus();
  leaving = false;
});
...
dlg.addEventListener("click", (e) => {
  if (e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
  const a = e.target.closest("a[href^='/']");
  if (!a || a.target) return;
  leaving = true;
  dlg.close();
});
```
(Keep the existing comments above each listener; extend the in-app-link comment rather than duplicating it. `leaving` declared next to `busy`.)

```css
/* kura.css — beside `.k-notif__head h2:focus` convention, in the shell block: */
/* The router focuses a view's heading after navigation; like the panel headings, it shows no ring. */
#main [tabindex="-1"]:focus { outline: none; }
```

**Rationale:** heading focus is the standard SPA route-change pattern and matches the codebase's own convention for programmatically focused headings (通知, 顧問). `mine > 1` = "not the first render" without a new flag; a stale render never focuses (inside `mine === seq`). A modal advisor left open during Back/Forward makes the page inert — `focus()` is then a no-op and focus stays in the dialog (acceptable; no extra code). `aria-current` is non-visual (no canvas style needed).

**Verification:** at C1 and C4 (Appendix A.5 focus probe): first load → `document.activeElement === document.body`; Home card → detail: active = detail `h1`, `scrollY === 0`; Back → Home `h1.k-sr` active, Browse Back keeps restored count and `scrollY` unchanged by focus; header search → Browse `h1`; `/nope` → not-found `h1`; advisor chip / rec 看詳細 link → dialog closed, active = detail `h1` (repeat on `/methodology` link if the mock offers none — use footer link while panel closed as control); Esc in advisor → dock; 通知 link → target `h1`; Esc in 通知 → bell. No ring drawn on the focused `h1` (computed `outline-style: none`). Footer: `/methodology` → that link only has `aria-current="page"`; `/about` → that one; `/` → none.

**If this fails:** `git -C <work-root> revert` the step's commit; log S1/S2 as findings.

---

### Step 2: Pending caption contrast in dark [PATCHED]

**Objective:** S3.
**Confidence:** High
**Depends on:** None
**Do-Projex:** Encouraged

**Files:**
- `internal/web/static/kura.css`

**Changes:**

```css
/* Before — file header comment: */
   Token values are docs/design-handoff/kura-tokens.css unchanged, plus
   --k-chart-mark. ...
/* After: */
   Token values are docs/design-handoff/kura-tokens.css unchanged, plus
   --k-chart-mark and --k-mat-note. ...

/* Light block — after --k-mat-ink: */
  --k-mat-note: #7D7262;       /* caption on the photo mat */
/* Both dark blocks (media query and [data-theme="dark"]) — after --k-mat-ink: */
    --k-mat-note: #6B6055;

/* Before: */
.k-card__pending { font-size: 12px; color: #7D7262; letter-spacing: .04em; text-align: center; }
/* After: */
/* The mat stays near-white in dark, so its caption keeps a dark ink in both
   themes. The canvas colour #7D7262 reaches 4.71:1 on the light mat but 3.76:1
   on the dark one (#ECE5D7), under AA's 4.5:1 for 12px text; dark uses
   #6B6055, 4.89:1. Rejected: --k-ink-3 (light ink in dark, unreadable on the
   light mat) and --k-mat-ink (2.11:1). A token, not a dark-only rule, so a
   [data-theme="light"] sub-tree under a dark OS keeps the light value. */
.k-card__pending { font-size: 12px; color: var(--k-mat-note); letter-spacing: .04em; text-align: center; }
```

**Rationale:** smallest change meeting the nav's AA bar; the replacement value is the canvas's own light `--k-ink-3`. A token (precedent: `--k-chart-mark`) handles nested `[data-theme]` sub-trees, which a dark-only selector would get wrong. Light look unchanged. Logged in the patch doc as deviation candidate **D1** (canvas `Card.dc.html` / CardStates dark row value) for canvas pass 2 / def § Canvas deviations — orchestrator may instead choose log-only.

**Verification:** Appendix A.2 contrast probe on `/dev/cards.html` (both sections) and on any real pending listing: caption ratio ≥ 4.5 in light (≈ 4.71 on `#FFFFFF`) and dark (≈ 4.89 on `#ECE5D7`); computed `color` light `rgb(125, 114, 98)`, dark `rgb(107, 96, 85)`. Sold + pending card (mat `opacity: .6`) is canvas-designed → if met, log, don't fix.

**If this fails:** revert the commit; log as finding with measured ratios.

---

### Step 3: Walk setup — server, fixtures, probes [PATCHED]

**Objective:** a running app on real data, fixture ids for every state, probes ready.
**Confidence:** High
**Depends on:** Steps 1–2 committed (the walk exercises them)

**Files:** none in the repo (scratch only).

**Changes:**
1. `go build -o <scratchpad>/kura.exe ./cmd/kura` from `<work-root>`; start `<scratchpad>/kura.exe serve -db kura.db -addr 127.0.0.1:8781` in the background from `<work-root>`; record PID. `GET /api/overview` → `advisor: "mock"`, `crawling: false`.
2. Empty-DB server for HomeEmpty: same binary, `serve -db <scratchpad>/empty.db -addr 127.0.0.1:8782`; record PID. (A-4.)
3. Fixtures — from `/api/listings?sort=deal&limit=200`, `/api/listings?sort=newest&limit=200`, `/api/models?limit=50` pick and record in the patch doc: detail with fair + distribution (≥ 3 comps, high/medium confidence); detail with no estimate (`尚無行情估價`); anomaly; low-confidence; pending photo (`photo_pending` or `NowPrinting` image) if any; a model with > 12 在售 (show-more) — `/model/LOUIS%20VUITTON/M51252` known good; a Browse query with zero results (`/browse?q=zzzz-no-match`).
4. Paste Appendix A probes as reusable snippets for `javascript_tool`.

**Verification:** both servers answer; fixtures table filled; `git -C <work-root> status --porcelain` shows nothing new (scratch lives outside the worktree; `kura.db*` gitignored).

**If this fails:** server won't start → check port / DB path, don't touch `kura.db`; fixture missing → mark its state n/a with reason.

---

### Step 4: Closing walk + a11y pass (matrix) [PATCHED]

**Objective:** S4 — every state × every combo recorded; findings W1…Wn listed.
**Confidence:** Medium (volume; tooling flakiness)
**Depends on:** Step 3

**Combos:** C1 light 1280×860 · C2 dark 1280×860 · C3 light 375×812 · C4 dark 375×812. Per combo: `resize_window` (+ `colorScheme`), reload, A.1 theme probe → expect `paper` `rgb(245, 240, 230)` light / `rgb(27, 24, 21)` dark and `innerWidth` 1280 / 375. Re-apply the combo if the pane reset it (desktop app may clear sizes between turns).

**Walk States** (reach each through the app — real clicks via `find` + `computer left_click` preferred; `element.click()` / `go()` via JS acceptable when a click is flaky, logged as method deviation):

| Id | Flow / state | Behaviour assertions (C1 + C4 mandatory; C2 / C3 reach state + scan) |
|---|---|---|
| H1 | Home `/` guest | shelves render (最新上架 5, 價格最美 3), 看全部 links, card rows scroll inside own rail at 375 |
| H2 | Home empty (`:8782`) | site-empty h1 「還在準備中。」; no shelves |
| B1 | Home 「看全部價格最美 N 件」 → `/browse?deals=1&sort=deal` | count = Home N; S1 focus |
| B2 | Browse: change 品牌 / 最低成色 / 預算 / 排序, toggle 只看低於行情, 再看 → 80 cards, open a card, Back | URL updates; Back restores count + scroll; load-more focus on first new card |
| B3 | Browse 375: 篩選與排序 open / close | `aria-expanded` flips; panel controls reachable by Tab; no overflow |
| B4 | Browse zero result `/browse?q=zzzz-no-match` | zero copy, no console error |
| B5 | Header search (1280 input / 375 icon → Browse) | lands `/browse?q=…`; S1 focus |
| D1 | Detail with fair + distribution | 鑑定卡 ladder, 行情估價 + 行情怎麼算 link (→ `/methodology`, S1), 以表格檢視 toggles table + back, dot tooltip, 售價變化, CTA `target="_blank" rel="noopener noreferrer"` (attribute check only) |
| D2–D4 | Detail no-estimate · anomaly · low-confidence | badge + prose state matches card |
| D5 | Detail pending photo (or `/dev/cards.html`) | S3 ratio |
| M1 | Model page | 4 stat tiles (資料不足 where gated), 追蹤這個型號 toggles, distribution, 在售 row + show-more focus |
| A1 | Advisor dock → open (1280 floating / 375 mobile panel) | 示範模式 pill visible; focus composer (1280) / heading (375); backdrop in dark |
| A2 | Chip prompt → streamed mock turn | progress lines, figures, numbers render; scroll pins |
| A3 | Full page ⇄ floating | `aria-label` 放大成整頁 / 縮小成視窗 flips |
| A4 | Rec 看詳細 / name / chip link | panel closes, route changes, S1 focus = detail `h1` |
| A5 | Esc | closes, focus = dock |
| N1 | 登入 → member header 林小姐 | bell badge count, focus = bell |
| N2 | 通知 open (popover 1280 / sheet 375) | heading focused; list renders; Esc → bell |
| N3 | 全部標為已讀 | badge hidden; bell `aria-label` 「通知」 |
| N4 | Notification link | route + S1 focus |
| N5 | Track toggles (card heart, detail 追蹤這件, model 追蹤這個型號) | `aria-pressed` + 追蹤清單（N） count sync across views; reload keeps session |
| P1 | Footer 行情怎麼算 → `/methodology` | 5 sections, 6 badges; S2 |
| P2 | Footer 關於 Kura … 服務條款 (all five) | pending view, h1 = link text; S2 |
| P3 | `/nope` direct load + in-app | not-found; S1 on in-app |
| K | Keyboard pass per route (Home, Browse, detail, model, `/methodology`, pending) | Tab from page top through header → ≥ 10 main focusables → footer → dock: every stop visible (A.5 `ring` true, rect on screen), order follows reading order, no trap; advisor modal keeps Tab inside; 通知 non-modal |

**Per state per combo:** run A.2 (contrast), A.3 (overflow), A.4 (targets), A.6 (landmarks/title); `read_console_messages` `onlyErrors`; `tabs_select` + screenshot where the pane renders. Record in the patch doc matrix: `✓` or `✗ W{n}`, plus `sw`, `fails`, `small`, `err` counts. At C2 / C3 the K pass is a spot-check: 3 stops per route, ring colour = `--k-focus` of that theme (`rgb(227, 161, 166)` dark / `rgb(107, 31, 42)` light).

**Judging A.2 / A.4 output:** a fail is a finding unless excepted — inline text links inside a sentence (2.5.8 inline exception); disabled controls; `.k-sr`; decorative `aria-hidden` icons; text over photos; logos. `manual` entries (ancestor opacity, unparsed colours) are judged by eye from the screenshot / computed colours and recorded.

**Known candidates to look at (not pre-judged):** `.k-acct` 我的帳號 button — accessible name 「我的帳號」 vs visible 「林小姐」 (2.5.3 label-in-name) and inert (account menu = Deferred); `.k-switch` checkbox 18 px (label as target?); pending pages' 40 px `h1` at 375 (no mobile override); native form controls in dark (`color-scheme: dark`); sold card mat `opacity: .6`.

**Verification:** matrix has no empty cell; every ✗ has a W-id; findings list has severity (A-level blocker | AA | polish) + combos + evidence.

**If this fails:** tooling blocks a combo (emulation not applying) → use A-1 fallback, record; a state unreachable → n/a + reason (counts toward derive 9's judgement, not hidden).

---

### Step 5: Bounded walk fixes + re-scan [PATCHED]

**Objective:** S5, S6.
**Confidence:** Medium (content unknown until Step 4)
**Depends on:** Step 4

**Files:** `internal/web/static/**` only — per finding.

**Fix Rules:**
- **R1 Fixable** (if within cap): hard-coded colour or wrong token failing AA where an existing token passes · missing / invisible focus ring, focus lost / trapped / misplaced · target < 24 px fixable by `min-height` / padding only · page-level horizontal overflow at 375 · console error · dark-theme element on a light-only hard-coded colour · missing accessible name or label-in-name mismatch (aria / `.k-sr` text only, no visible copy) · a built control that is broken (JS error, dead within built scope).
- **R2 Log only:** needs visual design or copy absent from the canvas (→ canvas pass 2) · touches an open human question (cite id: F1, F6, 大家都在找, iter-7 F5 / F6, def F19, R2, def L3 / L4) · a Deferred item's behaviour (cite Deferred row) · needs Go / API change · canvas-specified visual (opacity, size, layout) in a canvas-covered combo.
- **R3 Canvas colour failing AA:** fix only when the change is colour-only and the replacement is an existing canvas token value (as Step 2); log as deviation candidate D{n} for canvas pass 2 / def § Canvas deviations.
- **Cap:** ≤ 6 walk fixes. More qualifying findings → fix the 6 most severe (A-level before AA before polish; flow-blocking first), log the rest as **over cap** with severity — a follow-up objective for derive 9, not an open loop here.
- Each fix: smallest diff, rationale comment where non-obvious (Source Hygiene), own commit or one commit naming W-ids in the patch doc.

**Re-scan:** each fixed finding re-checked in every combo it appeared in; then a final A.2–A.4 + console pass on every route (H1, B1, D1, M1, P1, P2, P3) in all four combos, plus A1 / N2 open states at C2 and C4 (dark, no artboard). Record as § Final scan in the patch doc.

**Verification:** S5 / S6 as stated; `go vet ./...` still clean (static change only).

**If this fails:** a fix regresses another combo → revert that fix, log the finding (R2 "needs design").

---

### Step 6: Def update + Deferred ↔ def 1:1 [PATCHED]

**Objective:** S7.
**Confidence:** High
**Depends on:** Step 5 (final scan result decides "checked" wording)
**Do-Projex:** Encouraged

**Files:**
- `.projex/2610130449-kura-mock-def.md`

**Changes:**

```markdown
<!-- Before (Canvas gaps): -->
- Member layer: watchlist page (`追蹤清單（4）` → `#`), account / persona menu, sign-in-as-persona (guest header shows 登入 / 加入會員)
...
- Card 「尚無行情估價」 state: `project/Card.dc.html` computes `isNone` but renders nothing (Tokens sheet designs the badge)
- Theme × layout combos with no artboard: model, browse, advisor in dark; detail mobile in light → token-driven, checked in a11y pass

<!-- After: -->
- Member layer: watchlist page (`追蹤清單（4）` → `#`), account / persona menu, sign-in-as-persona (guest header shows 登入 / 加入會員), 通知設定 (notification dropdown footer link → `#`)
...
- Card 「尚無行情估價」 state: `project/Card.dc.html` computes `isNone` but renders nothing (Tokens sheet designs the badge) — built (card badge from the Tokens sheet)
- Theme × layout combos with no artboard: model, browse, advisor in dark; detail mobile in light → token-driven — checked: closing walk + a11y pass at light/dark × 1280/375 on every flow (2610131159-kura-closing-walk-a11y-plan.md and its patch doc); findings fixed or logged there
```
If the final scan leaves an unfixed AA / A-level finding in a no-artboard combo, write "checked — open: W{ids} logged (blocker …)" instead of an unqualified "checked".

`> **Created:** … | **Last Revised:** 2026-10-13` unchanged date if same day. Revision Log row:
`| 2026-10-13 | Canvas gaps: theme × layout → checked (walk + a11y pass, 4 combos); 「尚無行情估價」 → built; Member layer += 通知設定 (inventory: link → \`#\`, as sprint Deferred). No L2–L4 question decided. — trigger: {patch-doc filename} (Sprint 2610130524-kura-interactive-mock-sprint.md — iteration 8) |`

Patch doc § Deferred ↔ def reconciliation: one row per def § Canvas gaps item → status (built / checked / Deferred row name), and one row per sprint nav Deferred row → def gap item. Expected after this edit: 1:1, with the multi-provider Deferred row covering the def's 7-item multi-provider line. Any mismatch is reported for derive 9 (sprint nav not edited here).

**Rationale:** 通知設定 is an undesigned canvas link target like 追蹤清單 — inventory, not a decision; derive 6 already Deferred it. 尚無行情估價 was built in iter 1 but the def line never said so — without both edits the 1:1 check fails on bookkeeping alone.

**Verification:** `grep -n "通知設定\|尚無行情估價\|Theme × layout" .projex/2610130449-kura-mock-def.md` shows the three updated lines; Revision Log has the row; no Open Question checkbox changed.

**If this fails:** revert the def commit.

---

### Step 7: Final checks + teardown [PATCHED]

**Objective:** S8.
**Confidence:** High
**Depends on:** Steps 1–6

**Changes / actions:**
1. `go vet ./...` and `go test -count=1 ./...` from `<work-root>` → green.
2. Source Hygiene on changed static files: `git -C <work-root> diff ab33403 -- internal/web/static | grep -nE "^\+.*(projex|[0-9]{10}-|[Ii]teration|Step [0-9]|// ?(=|-){4})"` → no hits.
3. Stop both servers by recorded PID; confirm ports free; delete scratch binary / `empty.db` (outside worktree).
4. `git -C <work-root> status --porcelain` → empty.

**Verification:** outputs recorded in the patch doc.

**If this fails:** fix the failing check before reporting; a test failure unrelated to static changes → report, don't patch Go here.

---

## Verification Plan

### Automated Checks
- [x] `go vet ./...` clean
- [x] `go test -count=1 ./...` green (advisor | listing | radar | brandoff | store | web)
- [x] Source Hygiene grep on the static diff: 0 hits

### Manual Verification
- [x] Walk matrix (Step 4) complete, every cell recorded at C1–C4
- [x] S1 focus assertions at C1 + C4 (link, search, Back, advisor link, 通知 link, Esc paths)
- [x] Final scan (Step 5) on every route × 4 combos + advisor / 通知 open in dark
- [x] Screenshots of the no-artboard combos where the pane renders (advisor dark, 通知 dark, browse dark, model dark, detail dark 1280, detail light 375) — partial: model dark 1280 rendered partially, detail dark 1280 not captured (pane); DOM metrics are the evidence of record

### Acceptance Criteria Validation
| Criterion | How to Verify | Expected Result |
|-----------|---------------|-----------------|
| S1 | A.5 after each navigation type, C1 + C4 | active = view `h1` / `.k-error`; body on first load; dock / bell on Esc |
| S2 | `document.querySelectorAll('.k-footer [aria-current]')` per route | 1 on footer pages, 0 elsewhere |
| S3 | A.2 on caption, both themes + `/dev/cards.html` | ≥ 4.5 both; light colour unchanged |
| S4 | patch doc matrix | no empty cell |
| S5 | patch doc findings | each W has disposition; ≤ 6 fixed |
| S6 | final scan table | 0 overflow, 0 console errors, remaining fails all logged |
| S7 | def grep + reconciliation table | 3 lines updated, Revision Log row, 1:1 |
| S8 | vet / test / grep / status / PIDs | green, clean, stopped |

---

## Rollback Plan

1. `git -C <work-root> revert` the walk-fix commits (newest first), then the def commit, then the Step 1–2 commit(s) — all on the sprint branch; nothing reaches `master` before sprint close.
2. Stop servers by PID; remove scratch files.

---

## Appendix A — Probe snippets (`javascript_tool`)

**A.1 Theme / viewport**
```js
({ dark: matchMedia("(prefers-color-scheme: dark)").matches, attr: document.documentElement.dataset.theme ?? null,
   paper: getComputedStyle(document.body).backgroundColor, vw: innerWidth, title: document.title })
```

**A.2 Contrast** (visible text with a direct text node; composites translucent backgrounds up the ancestor chain; 3:1 for ≥ 24 px or ≥ 18.66 px bold)
```js
(() => {
  const rgb = (s) => { const m = /^rgba?\(([^)]+)\)$/.exec(s); if (!m) return null; const p = m[1].split(/[\s,\/]+/).filter(Boolean).map(parseFloat); return [p[0], p[1], p[2], p.length > 3 ? p[3] : 1]; };
  const lin = (v) => { v /= 255; return v <= 0.03928 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4; };
  const L = (c) => 0.2126 * lin(c[0]) + 0.7152 * lin(c[1]) + 0.0722 * lin(c[2]);
  const over = (t, u) => [0, 1, 2].map((i) => t[i] * t[3] + u[i] * (1 - t[3])).concat(1);
  const fails = [], manual = [];
  for (const el of document.body.querySelectorAll("*")) {
    if (![...el.childNodes].some((n) => n.nodeType === 3 && n.textContent.trim())) continue;
    if (!el.getClientRects().length || el.closest(".k-sr, :disabled")) continue;
    const cs = getComputedStyle(el);
    if (cs.visibility === "hidden") continue;
    const fg = rgb(cs.color);
    const layers = []; let opaque = false, faded = false;
    for (let e = el; e; e = e.parentElement) {
      const s = getComputedStyle(e);
      if (parseFloat(s.opacity) < 1) faded = true;
      if (s.backgroundImage !== "none" && e !== el) faded = true;
      const c = rgb(s.backgroundColor);
      if (c === null) { manual.push(["bg?", el.className, s.backgroundColor]); break; }
      if (c[3] > 0) { layers.push(c); if (c[3] === 1) { opaque = true; break; } }
    }
    const txt = el.textContent.trim().slice(0, 24);
    if (!fg) { manual.push(["fg?", txt, cs.color]); continue; }
    let bg = [255, 255, 255, 1];
    for (let i = layers.length - 1; i >= 0; i--) bg = over(layers[i], bg);
    const f = over(fg, bg);
    const a = L(f), b = L(bg), ratio = (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05);
    const fs = parseFloat(cs.fontSize), fw = parseInt(cs.fontWeight, 10);
    const need = fs >= 24 || (fs >= 18.66 && fw >= 700) ? 3 : 4.5;
    if (faded || !opaque) manual.push([txt, el.className, ratio.toFixed(2)]);
    else if (ratio < need) fails.push([txt, el.tagName + "." + el.className, ratio.toFixed(2), need, cs.color, `rgb(${bg.slice(0, 3).map(Math.round)})`]);
  }
  return { fails: fails.length, top: fails.slice(0, 25), manual: manual.length, manualTop: manual.slice(0, 10) };
})()
```

**A.3 Overflow**
```js
(() => {
  const vw = document.documentElement.clientWidth, out = [];
  for (const el of document.body.querySelectorAll("*")) {
    const r = el.getBoundingClientRect();
    if (!r.width || el.closest(".k-sr")) continue;
    if (r.right > vw + 1 || r.left < -1) {
      let clipped = false;
      for (let p = el.parentElement; p && p !== document.body; p = p.parentElement)
        if (/(auto|scroll|hidden|clip)/.test(getComputedStyle(p).overflowX)) { clipped = true; break; }
      if (!clipped) out.push(el.tagName + "." + el.className);
    }
  }
  return { sw: document.documentElement.scrollWidth, vw, escaping: out.slice(0, 15), n: out.length };
})()
```

**A.4 Targets**
```js
(() => {
  const sel = 'a[href], button, input:not([type=hidden]), select, textarea, summary, [tabindex]:not([tabindex="-1"])';
  const small = [];
  for (const el of document.querySelectorAll(sel)) {
    if (!el.getClientRects().length || el.closest(".k-sr, [hidden]")) continue;
    const box = (el.matches("input") && el.closest("label")) || el;
    const r = box.getBoundingClientRect();
    if (r.width >= 24 && r.height >= 24) continue;
    const inline = el.tagName === "A" && getComputedStyle(el).display === "inline" && /\S/.test(el.parentElement.textContent.replace(el.textContent, ""));
    small.push([el.tagName, el.className, (el.getAttribute("aria-label") || el.textContent).trim().slice(0, 16), Math.round(r.width), Math.round(r.height), inline ? "inline-exc" : ""]);
  }
  return { n: small.filter((s) => !s[5]).length, small: small.slice(0, 25) };
})()
```

**A.5 Focus**
```js
(() => {
  const e = document.activeElement, cs = getComputedStyle(e), r = e.getBoundingClientRect();
  return { tag: e.tagName, cls: e.className, name: (e.getAttribute("aria-label") || e.textContent || "").trim().slice(0, 20),
    isViewHeading: e === document.querySelector("#main h1, #main .k-error"), ring: cs.outlineStyle !== "none" && parseFloat(cs.outlineWidth) >= 2,
    ringColor: cs.outlineColor, onScreen: r.bottom > 0 && r.top < innerHeight && r.width > 0, scrollY,
    dialogOpen: !!document.querySelector("dialog[open]"), current: [...document.querySelectorAll(".k-footer [aria-current]")].map((a) => a.textContent) };
})()
```

**A.6 Landmarks / title**
```js
({ title: document.title, lang: document.documentElement.lang, h1: [...document.querySelectorAll("#main h1")].map((h) => h.textContent.trim()),
   landmarks: [...document.querySelectorAll("header, main, nav, footer, [role=search], dialog[open]")].map((e) => e.tagName + (e.getAttribute("aria-label") ? ":" + e.getAttribute("aria-label") : "")) })
```

---

## Notes

### Split Decision
`No split — heuristic tripped (~575 lines, 7 steps > 5; ~44 KB) but steps tightly coupled: the pre-identified fixes precede the walk that verifies them, the walk-fix set depends on the walk's findings, and the def wording depends on the final scan. ~110 lines are Appendix A probe snippets (reference, not steps).`

### Risks
- **Walk volume** (≈ 25 states × 4 combos): scripted probes keep each cell to one call batch; full behaviour only at C1 + C4 (both widths, both themes covered); C2 / C3 = reach + scan. Mitigation if time runs out: finish C2 / C3 scans for the no-artboard states first (they are the def gap), record the rest as n/a "not run" — never as ✓.
- **Fix volume:** capped at 6 walk fixes + over-cap logging (no open-ended loop).
- **Emulation / screenshots flaky:** A-1 fallback; DOM metrics are evidence of record; `tabs_select` before screenshots.
- **D1 deviation (Step 2):** changes a canvas-specified dark value; logged for canvas pass 2; orchestrator may prefer log-only (then revert Step 2, keep the finding).
- **Heading focus on Back with a long Browse list:** next Tab starts after the heading, not at the previously focused card — standard SPA trade-off; restoring per-entry focus is not in scope.

### Open Questions
- None blocking execution. For the orchestrator (not decided here): D1 fix vs log-only; any R2-logged finding that needs a human answer is listed in the patch doc with its question id.
