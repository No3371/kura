# Patch: Kura Back/Forward scroll restoration (W1) + def F5 pointer

> **Status:** Complete
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 8
> **Author:** Claude Opus 5.5 (patcher, sprint subagent)
> **Parent:** 2610131253-kura-closing-walk-a11y-audit.md
> **Directive:** Orchestrator: fix audit W1 — router-owned scroll restoration (`scrollRestoration = "manual"`, per-entry key in `history.state`, in-memory map; `go()` assigns keys, `replaceURL` keeps key + Browse `shown`; restore after popstate render, heading focus with `preventScroll`; `#` steps stay non-rendering). Low focus side-issue only if it falls out naturally, with rationale comment. Plus audit Minor #1: def 「尚無行情估價」 built line → pointer to open iter-7 F5. Leave audit Minor #2–#4 and patch slips alone.
> **Source Plan:** Direct (audit finding W1 + Minor #1)
> **Result:** Success
> **Related Projex:** 2610131253-kura-closing-walk-a11y-audit.md (W1, Minor #1 — source) | 2610131207-kura-closing-walk-a11y-patch.md (W1 logged open) | 2610130938-kura-browse-page-audit.md (F2 "scroll restored" — now true by construction) | 2610130449-kura-mock-def.md (§ Canvas gaps line edited) | 2610131134-kura-methodology-footer-audit.md (F5 — pointed to, not decided)

---

## Summary

W1: Back to a long Browse list landed at the shorter detail page's max (C2 2400 → 1434, C4 3000 → 2040). Cause: browser `auto` restoration runs at `popstate` while the leaving view is still in `<main>`; the async render swaps the long view in after the clamp. Fix: router owns scroll. Each history entry carries `key` in `history.state`; position recorded under the key of the view **on screen** (not the address bar) when history leaves it; restored by the router right after `replaceChildren`, then heading focus with `preventScroll`. Verified 15 Browse Back/Forward round trips over 2 combos (incl. the auditor's exact failing positions) — all exact.

Focus side-issue (audit Low): tested the "body focus when restored below top" variant — does **not** fall out naturally: next Tab from `<body>` goes to the crumb 首頁 and scrolls to 0, same as from the heading, and a screen reader announces nothing. Kept heading focus always; rejection recorded as a rationale comment in `focusView`.

Def: 「尚無行情估價」 built line now points to open iter-7 F5. Nothing decided.

---

## Changes

### Router scroll memory

**File:** `internal/web/static/js/nav.js`
**Change Type:** Modified
**What Changed:**
- L16–34: rationale comment; `history.scrollRestoration = "manual"`; `scrolls` Map; `onScreen` (key of entry whose view is in `<main>`); `newKey()` = `${performance.timeOrigin}-${++keys}` (timeOrigin prefix: keys survive reload in `history.state`, counter does not — avoids a new key matching an older entry's); first entry gets a key via `replaceState({...history.state, key})` (keeps a reloaded Browse `shown`)
- L49: `markRendered()` returns the key being rendered
- L58 `rememberScroll()`: `scrolls.set(onScreen, scrollY)` — keyed by on-screen view; during a render the address bar already names the next entry
- L65 `viewShown(key, restore)`: sets `onScreen`; with `restore` → `scrollTo(0, scrolls.get(key) ?? 0)`
- L83 `go()`: `rememberScroll()` → `pushState({ key: newKey() })` → render; immediate `scrollTo(0,0)` removed (router scrolls to top once the new view is in place; scrolling early would move the old view mid-load and record 0 for it if Back came meanwhile)
- L91 `replaceURL(url, state = {})`: writes `{ ...state, key: history.state?.key }` — key survives filter change and load-more

**Why:** audit Improvements shape; keyed by view-on-screen so a Back during a pending render cannot file the old view's position under the new entry.

### Router wiring + focus rationale

**File:** `internal/web/static/js/app.js`
**Change Type:** Modified
**What Changed:**
- L7: imports `rememberScroll`, `viewShown`
- L46–57 `focusView` comment: `preventScroll` reason updated (router has placed the view); added "Rejected: leaving focus on `<body>` when Back restores below top" with measured consequence
- L78: `const key = markRendered()`
- L101: `viewShown(key, mine > 1)` after `replaceChildren`/title/footer, before `focusView()` — first render untouched (same rule as focus)
- L133–137 popstate: moved → `rememberScroll(); render();`; `#` step → `keepViewState()` unchanged (non-rendering)

### Comment accuracy

**File:** `internal/web/static/js/browse.js`
**Change Type:** Modified
**What Changed:** L198 comment "null state = PAGE cards" → "no shown count = PAGE cards" (state now always carries `key`)

### Def pointer (audit Minor #1)

**File:** `.projex/2610130449-kura-mock-def.md`
**Change Type:** Modified
**What Changed:** § Canvas gaps 「尚無行情估價」 line += "its wording for no-group / unmappable-grade listings remains open — iter-7 F5 (2610131134-kura-methodology-footer-audit.md), tied to L4 Grade ordering"; Revision Log row (record only, nothing decided). No checkbox or Open-Question text touched.

---

## Verification

**Method:** scratchpad build of working tree serving a scratch copy of worktree `kura.db*` at 127.0.0.1:8793 (mock advisor, no crawl flags); built-in browser, real history traversal (`navigate back/forward` tool = browser Back/Forward, plus `history.back()/forward()` loops); stopped by PID (62252, 6252); scratch deleted; worktree `kura.db` sha1 `52f8e6cb…` before = after. `go vet ./...` clean; `go test -count=1 ./...` ok ×6 pkgs. Hygiene grep on `git diff -- internal/` (projex refs, `:N`, step/banner, deliberate/by design): 0 hits. Console errors: 0.

**Result:**
```
C1 light 1280
  Home left 610 → detail → Back(tool) → / y=610 | Forward(tool) → detail y=300 (left at 300)
  detail 700 → Back → Home 610 → Forward → detail 700 → Forward → model 400 (left 400)
  Browse 80 cards (load-more, state {"shown":80,key}):
    left 6000 → Back(tool) → 6000, cards 80
    left 2400/3000/5000/9600/1500/8000/4321 → Back → same value each, cards 80;
      Forward → detail 500 (left 500); Back again → same value
  # step (header href="#"): grid element identical (no render), key kept on # entry
C4 dark 375
  Home 400 → detail 900 → model 600; Back(tool) → detail 900; Back → / 400;
    Forward → detail 900; Forward → model 600; model 1400 → listing → Back → model 1400
  Browse 120 cards (2× load-more):
    left 2400/3000/5000/14000/21967(max)/9000/17777/600 → Back → same value each, cards 120
      (detail max 2010–2103: the pre-fix clamp values 2040/2069 no longer appear);
      Forward → detail 700; Back again → same value
  sort change → URL replaced, key kept, 40 cards; left 3000 → listing → Back(tool) → 3000, 40 cards
  location.hash step → Back/Forward: no render (same grid), y unchanged
Focus: forward nav → H1 at y=0; Back/Forward → H1 focused, viewport stays at restored y
Body-focus variant (tested, rejected): Back to Browse 5000, focus BODY → Tab → crumb 「首頁」, y=0
```

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610131253-kura-closing-walk-a11y-audit.md | Parent (W1, Minor #1) | None (audit record stays as written); condition "W1 disposition decided" → fixed here |
| 2610130449-kura-mock-def.md | Def | Canvas gaps pointer + Revision Log row (`e242efb`) |
| 2610130524-kura-interactive-mock-sprint.md | Sprint | Not edited (orchestrator owns nav) |

---

## Notes

- **Commits:** `97c670b` fix(web) code | `e242efb` docs(def) pointer | this doc.
- **Residual (Low, accepted):** a `#` fragment entry shares its view's key and restores nothing — placeholder `href="#"` links (header 追蹤清單 icon etc.) jump to top natively, and Back from them now stays at top (under `auto` the browser returned to the prior position). Distinct per-fragment keys would conflict with `keepViewState` copying the view's state onto the target entry. Stated in the `nav.js` comment.
- **Residual (by directive):** in-memory map → reload, or a full-page return from another site, starts at top.
- **Focus side-issue not resolved:** after Back to a scrolled Browse, next Tab still goes to the top of the view (from heading or body alike). A real fix would focus the card the user left from — out of scope; follow-up candidate.
- `new-projex.sh --projex-dir .projex/closed` placed the doc in `closed/closed/`; moved to `.projex/closed/` before commit (script auto-routes patch to `closed/`; pass `.projex`).
- Audit Minor #2–#4 (D1 header wording, Deferred 通知設定 tag, patch number slips): untouched per directive.
