# Patch: Kura advisor panel — audit fixes F1–F7

> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 4
> **Status:** Complete
> **Author:** Claude Opus 5.5 (patch-projex direct directive, sprint subagent)
> **Parent:** 2610130846-kura-advisor-panel-audit.md
> **Directive:** Fix audit findings F1–F7 per orchestrator decisions: F1 `.k-adv__body` containing block (verify `scrollHeight` = visible height at 1280 + 375, with figures, after scroll-into-view); F2 full mode hides the page scrollbar, floating unaffected; F3 drop the no-op `navigating` flag only (router-wide focus out of scope); F4 chips atomic (`inline-block; max-width: 100%`); F5 listing map for the page lifetime, covered by a fetch-stub replay; F6 dark backdrop derived from dark tokens, light stays canvas literal, comment says token-driven pending canvas pass 2; F7 no composer autofocus on narrow / coarse-pointer viewports, desktop unchanged, focus still inside the dialog, comment names the rejected alternative.
> **Source Plan:** Direct (findings from 2610130846-kura-advisor-panel-audit.md)
> **Result:** Success
> **Commits:** `2ab2e28` fix(web) — all seven findings | doc commit (this file + audit conditions)
> **Related Projex:** 2610130846-kura-advisor-panel-audit.md | 2610130822-kura-advisor-panel-patch.md | 2610130815-kura-advisor-panel-plan.md (A8 backdrop, `autofocus` choice) | 2610130524-kura-interactive-mock-sprint.md

---

## Summary

All seven audit findings fixed in one code commit across the advisor's CSS, JS, and markup. The dialog has no hidden overflow anymore. Full mode covers the whole viewport with no root scrollbar. Chips no longer split across lines. A later turn can cite a listing returned in an earlier turn. Dark mode gets a visible, token-derived scrim. Phones open the panel without raising the keyboard. The dead `navigating` flag is gone, and behaviour is unchanged. Everything was verified live (mock advisor, copy of the worktree `kura.db`) at 1280 and 375, in light and dark.

---

## Changes

### Dialog overflow (F1)

**File:** `internal/web/static/kura.css`
**Change Type:** Modified
**What Changed:**
- `.k-adv__body` gains `position: relative` (l.409), with a rationale comment above it (l.405–408). Absolutely positioned descendants (the figures' `.k-sr`（另開視窗） spans) now resolve against the scroller instead of the fixed dialog.

**Why:** Before the fix, those spans escaped the body's clip and gave the dialog itself 2,600–3,400 px of scrollable overflow. Any scroll-into-view (screen-reader cursor, mobile keyboard, find-in-page) then scrolled the whole panel, header included.

### Full-mode page scrollbar (F2)

**File:** `internal/web/static/kura.css`
**Change Type:** Modified
**What Changed:**
- `html:has(.k-adv.is-full[open]) { overflow: hidden; }` (l.391), with a comment (l.389–390).

**Why:** `width: 100%` resolved against the viewport minus the root scrollbar (1265 at 1280), and the page behind still wheel-scrolled. The rule applies only while full mode is open.

### Chips (F4)

**File:** `internal/web/static/kura.css`
**Change Type:** Modified
**What Changed:**
- `.k-adv__chip` gains `display: inline-block; max-width: 100%; box-sizing: border-box; line-height: 1.5` (l.426), with a comment (l.423–425). `line-height: 1.5` keeps the chip at 22 px tall, against 21 px when it was inline. Without it the box would take the answer's 1.85 line-height.

**Why:** Inline `<a>` chips broke mid-border across lines. The canvas chip is a `<button>`, which is atomic inline.

### Dark backdrop (F6)

**File:** `internal/web/static/kura.css`
**Change Type:** Modified
**What Changed:**
- Light keeps `rgba(30,25,21,.18)`, the canvas literal (l.382).
- Dark sets `.k-adv:not(.is-full)::backdrop` to `color-mix(in srgb, var(--k-paper-sunk) 60%, transparent)` under both dark selectors, matching the token blocks' theme selectors (l.383–386).
- `:not(.is-full)` keeps full mode transparent, because the media-query selector out-specifies `.k-adv.is-full::backdrop`.
- The comment (l.378–381) says dark is token-driven until canvas pass 2 draws a dark artboard, and why the ink veil is not reused: it would lighten a dark page.

**Why:** The light literal was invisible on dark paper. No new token was added, because the token set mirrors `docs/design-handoff/kura-tokens.css`.

### Composer focus on phones (F7)

**Files:** `internal/web/static/js/advisor.js`, `internal/web/static/index.html`, `internal/web/static/kura.css`
**Change Type:** Modified
**What Changed:**
- `index.html` l.81: the textarea's `autofocus` is removed. l.62: `<h2 tabindex="-1">Kura 顧問</h2>`.
- `advisor.js` l.46–52: `noComposerFocus = matchMedia("(max-width: 767px), (pointer: coarse)")` and `heading`, with a rationale comment that names the rejected alternative (the textarea's autofocus on every screen).
- `advisor.js` l.72: after `showModal()`, focus moves to the heading when that query matches, else to the textarea.
- `kura.css` l.398–399: `.k-adv__title h2:focus { outline: none; }`, because the heading is a start point, not a control.

**Why:** Canvas `AdvisorMobile` shows the intro and prompts without a keyboard. 767 px matches the stylesheet's phone breakpoint. Desktop keeps focus-on-open.

### Dead `navigating` flag (F3)

**File:** `internal/web/static/js/advisor.js`
**Change Type:** Modified
**What Changed:**
- The flag declaration, its set in the in-panel link handler (l.92 now just `dlg.close()`), and its branch in the close handler are removed. The close handler now always calls `dock.focus()` (l.81).

**Why:** The UA's dialog focus-restore already put focus on the dock after 看詳細, so the flag did nothing. The live re-check below shows the same end state. Choosing a different post-navigation focus target is router-wide and out of scope.

### Cross-turn references (F5)

**File:** `internal/web/static/js/advisor.js`
**Change Type:** Modified
**What Changed:**
- `known` moves from a per-`ask()` local to module scope for the page lifetime, like `session` (l.41–44, with a rationale comment). `listings` events still overwrite entries with the newest data.

**Why:** A later answer citing an earlier turn's `[[id]]` rendered 「查看商品」 and no figure.

---

## Verification

**Method:**
- `go vet ./...`, `go test -count=1 ./...`, `node --check advisor.js`, and a grep for `navigating`.
- Live app: the binary was built from the worktree and served a scratch copy of the worktree `kura.db` on 127.0.0.1 with `advisor=mock`. The worktree `kura.db` was not touched.
- Checks used the built-in browser tab, brought to the front for screenshots, with real clicks and keys for dock open, Tab, 看詳細 and Esc.
- Each fix was A/B-checked by reverting it inline (control).

**Result:**
```
go vet ./...            → clean
go test -count=1 ./...  → advisor, radar, brandoff, store, web ok
node --check advisor.js → OK;  grep navigating → exit 1 (none)

F1  1280 dark, 1 turn, 4 figures:   dlg scrollHeight 638 = clientHeight 638; after scrollIntoView of
                                    .k-sr span / last figure / textarea → 638 = 638, scrollTop 0, header y 137 (dialog y 136)
    1280 light, 2 turns, 6 figures: 638 = 638 after scrollIntoView of last figure and .k-sr span; header y 137
    375 light, 1 turn, 4 figures:   598 = 598 after the same three scrollIntoViews; header y 205 (dialog y 204)
    control (.k-adv__body position: static inline): 1280 → 2623 vs 638;  375 → 3375 vs 598
F2  full 1280: panel 1280 wide, documentElement clientWidth 1280, html overflow hidden, dialog 800 = 800
    floating 1280: panel 680, html overflow visible, clientWidth 1265 (page scrollbar back)
F3  375: real click on 看詳細 → path /listing/brandoff%3A2101219625140, dialog closed, not full, focus on .k-adv-dock
    1280: real Esc → closed, focus .k-adv-dock, 2 turns kept
F4  full 1280: chips split across lines → 0 (control, chips forced inline: 4 split —
    ポパンクール オ · M40007, ネヴァーフルPM · M40155, ネヴァーフルMM · M40156 ×2); 375: 0 split
    chip height 22 px inline-block vs 21 px inline
F5  fetch-stub replay: turn 1 real mock (4 figures); turn 2 stubbed /api/chat SSE, text only (no listings
    event) citing two turn-1 ids + one unknown id →
    chips [バティニョール ヴェルティカル · M51153, ヴァヴァンGM モノグラム · M51170, 查看商品],
    figures for both turn-1 ids, order P UL FIGURE FIGURE P, 2 position bars drawn
F6  dark (prefers-color-scheme) floating backdrop = color(srgb 0.0824 0.0745 0.0627 / 0.6) (= --k-paper-sunk @ 60 %);
    dark full → rgba(0,0,0,0); light (?theme=light) floating → rgba(30, 25, 21, 0.18). Screenshot dark 1280:
    page photos and mats visibly dimmed behind the panel
F7  375 mobile emulation (coarse pointer): real click on dock → :modal, focus H2 「Kura 顧問」 inside dialog,
    not :focus-visible, outline none; real Tab → 放大成整頁 button (still inside)
    1280 fine pointer: real click on dock → focus TEXTAREA
375 page scrollWidth 375; console errors: none
```

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610130846-kura-advisor-panel-audit.md | Parent (findings F1–F7) | Conditions F1, F2 checked with link to this patch; patch added to Related Projex |
| 2610130822-kura-advisor-panel-patch.md, 2610130815-kura-advisor-panel-plan.md | Closed records of the original build | None. They are historical, and their `autofocus` / backdrop-literal / `navigating` descriptions were true at the time. |
| 2610130524-kura-interactive-mock-sprint.md | Sprint | None. The orchestrator owns the nav. |

---

## Notes

- F6 and F7 are orchestrator decisions on what the audit had marked as human decisions. Both stay provisional until canvas pass 2: dark scrim alpha 60 % on `--k-paper-sunk`, and the 767 px / coarse-pointer focus rule. The scrim relies on `::backdrop` inheriting custom properties from its dialog, which current Chromium, Firefox and Safari do. It was verified in the pane's Chromium.
- F7 was checked in emulation only. No real device or virtual keyboard was available, the same limit the audit had.
- F5 grows the map by at most the listings returned per page load, which is negligible for a demo session.
