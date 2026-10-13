# Audit: Kura closing walk + a11y pass

> **Status:** Complete (Accepted)
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 8
> **Author:** Claude Opus 5.5 (auditor, sprint subagent)
> **Parent:** 2610131207-kura-closing-walk-a11y-patch.md
> **Related Projex:** 2610131159-kura-closing-walk-a11y-plan.md (audited plan, S1–S8) | 2610131207-kura-closing-walk-a11y-patch.md (execution claims, walk matrix, W1–W3) | 2610130449-kura-mock-def.md (§ Canvas gaps / deviations edited in `2fc5d58`) | 2610130524-kura-interactive-mock-sprint.md (Deferred table, not edited) | 2610130938-kura-browse-page-audit.md (F2 scroll claim contradicted by W1)
> **Re-audit:** 2026-10-13 of 2610131307-kura-back-scroll-restore-patch.md — fix `97c670b` + def pointer `e242efb` + doc `9be835c` (range `ccc8f52..9be835c`) → § Re-audit; condition (W1) met, Minor #1 fixed
> **Audit Date:** 2026-10-13 | **Work Period:** 2026-10-13, commits `e6a6c58` `02b68ee` `2fc5d58` `9c8ae54` on `projex/sprint/2610130524-kura-interactive-mock` (range `ab33403..9c8ae54`)
> **Subject:** plan S1–S8 vs diff, tests and the live app — scratchpad build of HEAD serving a scratch copy of the worktree `kura.db*` (127.0.0.1:8791) + empty DB (:8792), no crawl flags, both stopped by PID, scratch deleted, worktree `kura.db` sha1 unchanged. Independent walk sample across C1–C4 (no re-run of the full matrix); canvas `project/Card.dc.html` re-read via Artifact tool (served 1791845608-ba84) for D1.

---

## Audit Summary

**Claim:** Router focuses the new view's `h1` after in-app navigation and marks the current footer link; advisor link no longer pulls focus to the dock; dark pending caption meets AA (D1); walk matrix × 4 combos complete with W2 + W3 fixed and W1 logged; final scan clean; def bookkeeping (theme × layout checked, 尚無行情估價 built, 通知設定 listed, D1); Deferred ↔ def § Canvas gaps 1:1; vet/test green.

**Verdict:** Verified — every sampled claim reproduces; W1 reproduces exactly as logged.

**Assessment:** Completeness: High | Correctness: High | Quality: High | Value: High

**Top Issues:**
1. W1 (Back to a long Browse list lands short) — real, pre-existing, reproduced at C2 + C4; Medium flow defect; fixable in a bounded patch (~20–30 lines, `nav.js` + `app.js`), the "open point" the patch names has a known answer — orchestrator decision, not a human one
2. W1's disposition used a blocker category outside the plan's four (deviation 4, disclosed) while 4 of 6 fix slots were unused
3. Def 「尚無行情估價」 → "built" carries no pointer to the still-open iter-7 F5 wording question (Minor; nothing decided, but a def reader could read it as settled)

---

## Claims vs Evidence

| Claim | Evidence (auditor, this session) | Status | Notes |
|-------|----------|--------|-------|
| S1 heading focus after in-app nav | C2: Home → 看全部價格最美 → Browse `H1` 「價格最美」; `/about` `/terms` `/nope` model, zero-result Browse, `/` → `H1` view heading, `outline-style: none`, `scrollY` 0. C1: header search → `/browse?q=Speedy` `H1`; 通知 link → detail `H1`, panel hidden; Back → Browse `H1`; Forward → detail `H1`. C3: 3 detail variants + model + `/contact` + `/` → `H1`. First load = `BODY` (C2 `/`, C3 detail direct) | ✓ | |
| S1 advisor link → detail `h1`, not dock | C2 (executor: "scan only"): real click on rec 「看詳細」 → route changed, `dialogOpen` false, active = detail `H1` | ✓ | Sync-route race (`/methodology` link inside panel) → `H1` + footer current; the queued `close` event could not be observed (hidden pane defers it, as patch notes) — `leaving` set synchronously before `dlg.close()` makes the handler order-independent by construction |
| S1 Esc paths | C4: Esc in advisor → `k-adv-dock`; 通知 Esc path not re-run (W3 path covered below) | ✓ | |
| S2 footer `aria-current` | 1 on `/about` `/terms` `/methodology` `/contact`; 0 on `/`, Browse, detail, model, `/nope` (C1–C4 sample) | ✓ | Footer nav has no `href="#"` link, so `pathname === location.pathname` cannot false-match |
| S3 pending caption AA | `/dev/cards.html` (C3): light section `rgb(125,114,98)` on `#FFF` 4.71; dark section `rgb(107,96,85)` on `rgb(236,229,215)` 4.89. Real listing `…2104102531481` at C4: `rgb(107,96,85)` on mat. Canvas `#7D7262` on dark mat = 3.76 (recomputed) | ✓ | Canvas `Card.dc.html` hard-codes `#7D7262` for both themes; `#6B6055` = canvas light `--k-ink-3` — D1 claim true |
| S4 walk matrix complete | Patch matrix: no empty cell; `—` cells = plan's scan-only C2/C3 behaviour columns | ✓ | |
| S5 findings disposed | W2: `.k-acct` `aria-label` 「林小姐 · 我的帳號」 at C2 + C4; visible 「林小姐」 (1280) / 「林」 (375) both contained. W3: C2 real Tab past 通知設定 → focus `k-watch`, panel hidden, `aria-expanded` false; C4 real Shift+Tab out of bell → `k-search-icon`, panel hidden | ✓ / ⚠ | W1 disposition category outside the plan's four (see Findings) |
| S6 final scan | Sample (≈ 25 route × combo cells): 0 contrast fails, 0 page overflow (`scrollWidth` = viewport at 375), 0 non-excepted small targets at 1280 and on detail/model/pages at 375, 0 console errors | ✓ | Browse 375: 34 / 67 `.k-card__model` 211 × 22 — spacing exception re-measured: nearest non-fixed target 68 px from centre (patch says ≥ 93 px; exception still holds, needs ≥ 12) |
| S7 def edits | `2fc5d58` diff: 3 Canvas-gaps lines + D1 deviation + Revision Log row; no Open-Question checkbox touched | ✓ | See Documentation Audit for wording notes |
| S7 Deferred ↔ def 1:1 | Re-derived: def gap items 13 (multi-provider · 4 member · 2 advisor · 行情怎麼算 · 5 footer pages · 資料管線 · §4.A · 尚無行情估價 · theme × layout) → 10 Deferred rows + 2 built + 1 checked; each of the 10 Deferred rows maps back to exactly one item | ✓ | Nit: Deferred row 通知設定 lacks the "— def § Canvas gaps" tag the other rows carry (now true) — derive 9 bookkeeping |
| S8 vet / test / hygiene / clean | `go vet ./...` clean; `go test -count=1 ./...` ok ×6 pkgs; hygiene grep on `git diff ab33403 -- internal/web/static` 0 hits; `git status --porcelain` empty | ✓ | |
| W1 pre-existing, logged | C2: Browse 80 cards, `scrollY` 2400 → card → Back → 1434 = detail max (1434). C4: 3000 → 2040 (= detail max), 5000 → 2069 (= detail max); one earlier C4 run 6000 → 6000 (restored). Count 80 restored every time | ✓ | Intermittent but dominant — the browser restores at `popstate` while the shorter detail view is still on screen |

---

## Objective Verification

### Step 1 — route focus + footer current (S1, S2)
**Evidence:** `e6a6c58` (`app.js` `focusView` / `markFooter` / `render`; `advisor.js` `leaving`; `kura.css` `#main [tabindex="-1"]:focus`) + live sample above.
**Findings:** Actual: as planned, verbatim to plan code. `mine > 1` skips first render; stale render never focuses. Side effect (expected, plan Risks): on detail the crumb links sit before `h1`, so forward Tab from the focused heading skips them (C1: `h1` → `k-detail__fairlink` → … → `k-dtrack`, ring `2px solid rgb(107,31,42)`, `:focus-visible` true); Shift+Tab reaches them. On Back to a scrolled Browse, focus is on the `h1` at the top while the viewport is lower — next Tab scrolls back to the filter form (compounds W1).
**Verification:** ✓ Verified

### Step 2 — pending caption (S3, D1)
**Evidence:** `e6a6c58` `kura.css` (`--k-mat-note` in `:root, [data-theme="light"]` + both dark blocks; `.k-card__pending` uses it) + `/dev/cards.html` + real listing.
**Verification:** ✓ Verified

### Steps 3–5 — walk, fixes, final scan (S4–S6)
**Evidence:** patch matrix + independent sample: C1 light 1280 (search, 通知 link, Back/Forward, keyboard ring); C2 dark 1280 (Browse W1, detail fair+dist incl. table toggle + CTA `rel`, advisor open/turn/link/race, 4 pages + not-found + model + zero-result, member W2/W3 by real keys); C3 light 375 (no-artboard detail ×3 variants, model, pending page, `/dev/cards.html`); C4 dark 375 (Browse W1 + load-more focus + B3 `aria-expanded` false→true→false, detail pending-photo, `/methodology` 5 h2, advisor heading focus, W3 backwards, Home empty `:8792` 「還在準備中。」 0 cards `sw` 375).
**Issues:** none new. W2 / W3 fixes behave as claimed.
**Verification:** ✓ Verified

### Step 6 — def + 1:1 (S7)
**Verification:** ✓ Verified (wording notes below)

### Step 7 — checks (S8)
**Verification:** ✓ Verified

---

## Code/Implementation Inspection

### `app.js` / `advisor.js` / `kura.css` (`e6a6c58`)
**Actual:** matches plan. Quality: High. `leaving` reset in the `close` handler pairs one flag with one close; the double-pending-close edge the patch notes is unreachable on a drawing page.
**Issues Found:** none.

### `notify.js` (`02b68ee`)
**Actual:** W2 `aria-label` set in `show()`; W3 `focusout` on `wrap` closes only when `relatedTarget` is outside bell + panel. `setOpen(false)` without `{focus}` → never steals focus back (verified: focus stays on the Tab target). Clicking a notification link: existing click handler closes first; router focus then fires a `focusout` on an already-closed panel (no-op).
**Issues Found:** none.
**Undocumented:** none — `git diff --stat ab33403..HEAD` = the 4 static files + def + plan + patch doc.

---

## Testing Validation

**Execution:** `go vet ./...` clean; `go test -count=1 ./...` ok (advisor, listing, radar, brandoff, store, web). Static JS has no automated tests — behaviour evidence is the walk (patch) + this sample.
**Missing:** no regression test for route focus / scroll restoration (none exists for the SPA shell; out of plan scope).

---

## Documentation Audit

**Patch doc accuracy:** Yes, with two trivial number slips: `.k-tag__stamp` 「店鑑」 10.68:1 is the light value (dark 7.64; both pass); spacing-exception "≥ 93 px" — measured 68 px at C4 (still passes).
**Def edits (`2fc5d58`) — truth + no decision:**
- Theme × layout → "checked … findings fixed or logged there" — true; the only unfixed finding (W1) is a flow defect, not an AA/A fail in a no-artboard combo, so the plan's "checked — open: W{ids}" qualifier was correctly not triggered.
- 尚無行情估價 → "built (card badge from the Tokens sheet)" — true (badge renders on detail `…2101219731339`, card + detail). Does not decide iter-7 F5 (wording for no-group / unmappable grade, def L4) but does not point to it either → Minor.
- 通知設定 added to Member layer — inventory only (link → `#`), matches sprint Deferred; no L3 notification-type question touched.
- D1 under Canvas deviations — true (canvas re-read). The list's header reads "(def overrides canvas …)"; D1 is an AA override, not a def rule — wording misfit only; orchestrator already chose this placement.
- Revision Log row states "No L2–L4 question decided" — confirmed: no checkbox changed, no Open-Question text edited.

---

## Source Hygiene

**Scope:** Diff only (`ab33403..9c8ae54`, excluding `**/.projex/**`)

| Rule | Location (symbol) | Quote | Severity |
|------|-------------------|-------|----------|
| — | — | no violations | — |

**Rule 6 gaps:** none — `focusView` (rejected `<main>`), `leaving` (queued-close constraint), `.k-card__pending` (rejected `--k-ink-3` / `--k-mat-ink`, token vs dark-only rule), `show()` aria-label (rejected fixed 我的帳號), `focusout` (rejected Tab-trap) all carry rationale.
**Rule 6 vacuity:** none. `#main [tabindex="-1"]:focus` comment names its constraint ("a start point, not a control").
**Rationale promotion:** none needed — rationale lives in the code comments; D1 promoted to def § Canvas deviations.
**Commit composition** (informational): 3 landed code/def commits checked — typed subject 3/3 | type matches diff 3/3 (`fix(web)` ×2, `docs(def)`) | `Projex:` trailer 3/3. `9c8ae54` doc-only, exempt.
**Violations:** None
**Trailer survival on `master`:** 0/6 — `master` holds only 6 pre-sprint commits; the convention lives on the sprint branch (30 trailered commits in `master..HEAD`, every code commit). Not a dead channel; the risk point is the sprint-close squash message — informational.

---

## Gap Analysis

### Promised But Not Delivered
| Promise | Status | Impact |
|---------|--------|--------|
| Screenshots of every no-artboard combo | Partial (disclosed in plan checklist) — auditor added detail dark 1280 (token-correct); further screenshots blocked by hidden pane | Low — DOM metrics are evidence of record |

### Undocumented Issues
| Issue | Severity | Affects |
|-------|----------|---------|
| Heading focus on Back leaves focus at page top while scroll is restored lower; next Tab jumps to the filter form | Low (expected trade-off, plan Risks) | Keyboard users on Browse Back — moot until W1 is fixed |
| Fixed advisor dock can sit over a card link while scrolling at 375 (centre 13 px from dock) | Low — partial overlap only, not 2.4.11 failure | 375 Browse |

### Unhandled Edge Cases
- Two `close` events pending at once (non-rendering pane only) → second runs `dock.focus()` — patch notes it; unreachable in a visible tab.

---

## Quality Assessment

### Completeness: High
**Strengths:** full matrix, honest n/a and `—` cells, A/B build proving W1 pre-existing, DB-copy discipline.
**Gaps:** W1 left open with 4 fix slots unused.

### Correctness: High
**Works:** route focus, footer current, advisor race, D1, W2, W3 — Yes.
**Bugs:** W1 (pre-existing) — Medium.

### Code Quality: High
**Positive:** smallest diffs, plan-verbatim, rationale comments name rejected alternatives.
**Tech Debt:** none added.

### Value Delivered: High
**Intended:** close the last def gap + shell a11y carry-overs + prove the goal walk.
**Actual:** done; one flow defect surfaced (and an earlier audit's claim corrected).
**Impact:** User: Positive.

---

## Open Findings

### Undocumented Discoveries
- Changes: none beyond the declared files.
- Problems: W1 is intermittent (1 of 4 auditor runs restored correctly) — a likely reason iter-5 browse audit F2 recorded 「scroll restored」; not proven.

### Impact Analysis
- Downstream: derive 9 goal check — walk evidence, def gaps line, 1:1 and tests all hold; W1 is the one open walk defect.
- Risks: W1 visible to a reviewer who scrolls Browse, opens a card, presses Back — Likelihood: Medium.

### Improvements
- W1 bounded fix shape (resolves the patch's open point without a scroll listener): `history.scrollRestoration = "manual"`; with manual restoration `scrollY` at `popstate` still belongs to the leaving entry, and `nav.js` already tracks the on-screen entry (`current` / `viewState`) → store the leaving position in an in-memory map keyed by a per-entry id carried in `history.state` (assigned by `go()`, preserved by `replaceURL` so Browse's `shown` state keeps it); `go()` stores the leaving position the same way; router restores the target entry's position after the `popstate` render (`scrollTo` after `replaceChildren`, before `focusView` with `preventScroll`). First load and in-app links unchanged (`go()` still scrolls to 0). Reload loses the map (acceptable; `sessionStorage` if not). ≈ 20–30 lines, `nav.js` + `app.js`; verify Home / Browse / model / detail Back and Forward at C1 + C4.

---

## Stakeholder Impact

| Stakeholder | Promised | Reality | Impact |
|-------------|----------|---------|--------|
| Reviewer (keyboard / SR) | Focus on new view, current footer link | Delivered | Positive |
| Reviewer (dark mode) | Readable pending caption | 4.89:1 | Positive |
| Reviewer (any) | Back returns to place in Browse | Count restored, position short (W1) | Neutral → Negative until fixed |
| Derive 9 | Evidence for goal check | Complete, 1:1 holds | Positive |

---

## Findings

### Critical (Must Address)
- None.

### Significant (Should Address)
- **W1 — Back to a long list lands short** (pre-existing; C2 2400 → 1434, C4 3000 → 2040 / 5000 → 2069, each = detail page max; 1 of 4 runs restored) — degrades a canvas flow end-to-end, contradicts iter-5 audit F2 → **fixable now** in a bounded patch (shape under Improvements); approach choice is technical → orchestrator decision, no human question involved. Severity Medium.
- **W1 disposition outside the plan's blocker set** (patch deviation 4) — plan S5 allowed only canvas pass 2 | open question | over cap | Deferred; "needs approach decision" with 4 of 6 fix slots free is weak → resolved by the orchestrator's W1 decision above; no separate action.

### Minor (Nice to Fix)
- **Def 「尚無行情估價」 "built" line has no pointer to open iter-7 F5 wording** → append "(wording for no-group / unmappable grade open — iter-7 F5)" via revise; fixable now, decides nothing.
- **D1 sits under a list headed "def overrides canvas"** → optional header tweak ("def or AA bar overrides canvas") at next def revise; orchestrator already accepted placement.
- **Sprint Deferred row 通知設定 lacks the "— def § Canvas gaps" tag** → derive 9 bookkeeping (sprint nav not edited here).
- **Patch doc number slips** (店鑑 ratio is the light value; spacing distance 68 px not 93 px) → no action; conclusions unchanged.

### Positive
- Executor's A/B build on `ab33403` proving W1 pre-existing; disclosed every method deviation and the hidden-pane artefact.
- Advisor focus race closed by construction, not by timing.
- D1 handled as a token (correct for `[data-theme]` sub-trees), rationale names both rejected tokens.

---

## Recommendations

**Immediate:** orchestrator decides W1 — bounded fix patch (recommended) or explicit accept-as-known for derive 9.
**Future:** none beyond Minor list.
**Process:** when a walk finding is fixable within cap, fix it or log it under an existing blocker category; a new category is a signal to escalate the choice to the orchestrator in the patch run, not after.

---

## Final Verdict

**Status:** Accept (was Accept with Conditions; condition met on re-audit)

**Overall Assessment:**
- Completeness: High
- Correctness: High
- Quality: High
- Value: High

**Conditions:**
- [x] W1 disposition decided by orchestrator before derive 9 judges the goal (fix via bounded patch, or accepted with reason recorded in the sprint nav) — fixed `97c670b`, re-audited ✓ (§ Re-audit)

**Sign-off:** Yes — all S1–S8 claims reproduce on independent sampling across C1–C4 including the no-artboard combos; def edits are true and decide no open L2–L4 question; W1 fixed and re-verified.

---

## Re-audit

> 2026-10-13 · subject 2610131307-kura-back-scroll-restore-patch.md · commits `97c670b` (`nav.js` / `app.js` / `browse.js`), `e242efb` (def), `9be835c` (doc) · scratchpad build of HEAD on a scratch copy of `kura.db*` (127.0.0.1:8795, no crawl flags, stopped by PID, scratch deleted; worktree `kura.db` untouched)

**Code read:** `history.scrollRestoration = "manual"`; per-entry `key` in `history.state` (`performance.timeOrigin`-prefixed, so a reloaded document cannot collide with an older entry's key); `rememberScroll()` files `scrollY` under the key of the view **on screen** (correct: at `popstate` under manual mode `scrollY` still belongs to the leaving view, and during a pending render the address bar already names the next entry); `viewShown(key, mine > 1)` restores after `replaceChildren`, before `focusView` (`preventScroll`); `go()` no longer scrolls early; `replaceURL` keeps the key and merges Browse's `shown`; `#` steps still go through `keepViewState` (no render). All `history.*` writers in `static/js` route through `nav.js` (grep: only `browse.js` → `replaceURL`). Rationale comments name the rejected alternatives (browser restoration at `popstate`, early scroll in `go()`, body focus on Back). Hygiene grep on `git diff 9c8ae54..HEAD -- internal`: 0 hits. `go vet` clean; `go test -count=1 ./...` ok ×6.

**Live (real browser Back = `navigate back`, Forward = `history.forward()`):**

| Combo | Case | Left at → after Back / Forward | Shown count | Focus |
|---|---|---|---|---|
| C2 dark 1280 | Browse 80 → detail → Back ×3 | 2400 → 2400 · 3000 → 3000 · 7000 → 7000 (detail max 1434: old clamp value gone) | 80, `{shown:80,key}` kept | `H1`, no scroll jump |
| C2 | detail ⇄ Browse ⇄ model chain | Forward detail 0 (left at 0) · detail 600 → Back Browse 7000 → Forward detail 600 · detail 500 → model → Back 500 → Forward model 400 | — | `H1` each |
| C2 | Home → detail → Back | 700 → 696 (Home max) | — | `H1` |
| C2 | `#` step (加入會員) on Home | no re-render (same DOM), key shared; native jump to top, Back stays top (documented residual) | — | — |
| C2 | 通知 link → detail → Back; advisor rec link → detail → Back | detail at 0, panel / dialog closed; Back → Home at its left position (300) | — | `H1` |
| C4 dark 375 | Browse 120 → detail → Back ×5 | 3000 · 6000 · 12000 · 20000 · 5000 → identical (detail max 2040–2099) | 120 kept | `H1` |
| C4 | Back → Forward → Back | 5000 → detail 0 → 5000 | 120 | `H1` |
| C4 | sort change (replaceURL) → detail → Back | 3000 → 3000; key kept, `shown` dropped → 40 cards | 40 | `H1` |
| C4 | `location.hash` step → Back | same DOM, y 2500 unchanged | 40 | — |
| C4 | model (25) 1400 → detail → Back; Home 678 → detail → Back | 1400 → 1400 · 678 → 678 | 25 | `H1` |
| C4 | advisor rec link from Home 500 → Back | detail 0, dialog closed; Back → 500 | — | `H1` |

0 console errors. W1 was intermittent pre-fix (3 of 4); post-fix 8 / 8 Browse round trips exact, plus Forward legs.

**Observed, not a finding:** opening 通知 focuses its heading in the static header, which scrolls the page to the top (pre-existing; the bell is only reachable at the top anyway).

**Def `e242efb`:** 「尚無行情估價」 line gains a pointer — wording for no-group / unmappable-grade listings "remains open — iter-7 F5 … tied to L4 Grade ordering" — true (sprint nav lists F5 as def L4; L4 Grade-ordering item covers unmappable grades); Revision Log row "Record only; nothing decided"; no checkbox or Open-Question text changed. ✓

**Residuals (accepted, documented in code + patch doc):** `#` placeholder links jump to top and Back from them stays at top (Low; were restored under `auto`) · in-memory map → reload / return from another site starts at top · after Back, next Tab from the focused heading scrolls to the view's top (Low; body-focus variant measured no better).

**Still open from this audit:** Minor #2 (D1 header wording), #3 (sprint Deferred 通知設定 tag → derive 9), #4 (patch number slips) — untouched by directive; none blocking.

**Re-audit verdict:** ✓ W1 fixed, no regression in route focus, Browse `shown`, `#` steps, advisor / 通知 navigation. Audit → Complete (Accepted).
