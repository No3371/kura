# Patch: Kura browse audit fixes (F1–F5)

> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 5
> **Status:** Complete
> **Author:** Claude Opus 5.5 (patch-projex direct-directive, sprint subagent)
> **Parent:** 2610130938-kura-browse-page-audit.md
> **Directive:** Fix audit F1 (router re-renders Browse on `#` step after `replaceState`), F2 (orchestrator decision: Back restores shown card count via `history.state`, URL contract unchanged, re-fetch respecting API `limit` cap, no history entry per load-more), F3 (`stateQuery` comment), F4 (rationale on 再看 disable during filter fetch), F5 (bound budget digits). F6 left to orchestrator.
> **Source Plan:** Direct (audit findings of 2610130938-kura-browse-page-audit.md)
> **Result:** Success
> **Commits:** `2d58893` fix(web) — F1–F5 | doc commit (this file + plan Risks annotation)
> **Related Projex:** 2610130938-kura-browse-page-audit.md | 2610130913-kura-browse-page-plan.md (Risks bullet marked disproved) | 2610130928-kura-browse-page-patch.md | 2610130524-kura-interactive-mock-sprint.md

---

## Summary

F1: router's "which URL is on screen" record moved into new `nav.js` (no import cycle: `app.js` imports every view, so a view cannot import `app.js`). `nav.js` owns `go(url)` and `replaceURL(url, state)`; Browse filter changes now go through `replaceURL`, so the popstate check (`moved()`) sees a `#` step as a fragment step, not a new view. F2: load-more records `{ shown: n }` on the current entry via `replaceURL(location.href, …)` (no new entry, URL unchanged); `renderBrowse` re-fetches the first n cards in `API_MAX` (100) chunks in parallel. A `#` step adds an entry with null state — popstate on a fragment step copies the on-screen view's state onto it (`keepViewState`), so Back to that entry still restores. F3/F4 comments fixed; F5 budget bounded to 9 digits on both URL and form paths.

---

## Changes

### Router state module

**File:** `internal/web/static/js/nav.js`
**Change Type:** Created
**What Changed:**
- Module comment: why it exists apart from `app.js` (cycle), why every path/query write must go through it (stale `current` → `#` step re-renders, dropping pages / open panel / scroll).
- `current` + `viewState`; `setRenderer` (l.22), `markRendered` (l.27), `moved` (l.33), `keepViewState` (l.38), `go` (l.42, moved from `app.js`, same body), `replaceURL(url, state = null)` (l.49: `history.replaceState` + `current` + `viewState`).

**Why:** F1 root cause — `current` lived in `app.js` and only `render()` updated it.

### Router

**File:** `internal/web/static/js/app.js`
**Change Type:** Modified
**What Changed:**
- l.6 imports `nav.js`; local `current` and `go` removed; `render()` calls `markRendered()` (l.46); `setRenderer(render)` (l.75).
- popstate (l.98–101): `moved()` → `render()`, else `keepViewState()`. The old popstate comment's content now lives in `nav.js`'s module comment.

### Browse

**File:** `internal/web/static/js/browse.js`
**Change Type:** Modified
**What Changed:**
- F2: header comment adds the load-more/history rule; `API_MAX = 100` (l.18) with the cap rationale (radar.Search answers `limit > 100` with its default 30, not 100); `apiURL(st, offset, limit = PAGE)`; `shownBefore()` (l.129: `history.state.shown`, integer > PAGE, else PAGE); `firstCards(st, n)` (l.135: `ceil(n/100)` parallel requests, `total` from first); `renderBrowse` uses `firstCards(st, shownBefore())`.
- F2: `loadMore` records `at` (l.231) and returns early if the path/query changed meanwhile (grid stays connected while the next route renders — writing then would put the count on the next page's entry); after append, `replaceURL(location.href, { shown })` (l.247).
- F1: `apply` → `replaceURL(url)` (l.199; null state = PAGE cards for a new filter).
- F3: `stateQuery` comment → "Canonical query: fixed key order; defaults omitted, sort=deal among them." (l.65).
- F4: comment on the 再看 disable in `apply` (l.209): a click during the fetch would fetch the new filter at the old grid's length and append to the old grid; the seq check misses it because `loadMore` reads `view.seq` after `apply` bumped it.
- F5: `readState` strips leading zeros, accepts `/^[1-9]\d{0,8}$/` (l.59) with rationale (22 digits → `1e+21` in field + URL, API ignores it); `apply` strips non-digits, leading zeros, `.slice(0, 9)` (l.188).

**Why:** audit F2–F5; the loadMore URL guard is new hazard handling introduced by F2's state write.

---

## Verification

**Method:** `go build ./... && go vet ./... && go test -count=1 ./...`; `node --check` on the three JS files; live app — scratch-built binary serving the worktree `kura.db` on 127.0.0.1:8097 (no crawl flags), built-in browser, real clicks / typing / keys; server stopped, viewport reset, tab closed after.

**Result:**
```
go test: ok advisor | ok radar | ok source/brandoff | ok store | ok web (others: no test files)

F1 repro (1280): /browse?deals=1&sort=deal → sort 最新上架 → 再看 → 80 cards, state {shown:80}
  → click header 登入 (#) → URL …#, 80 cards, 「再看 1,391 件」, grid element unchanged, state carried {shown:80}
  → Back to non-# entry → same grid element (tag kept), 80 cards
F1 repro (375): sort via mobile panel → 再看 → 80 → 登入 → 80 cards, same grid element,
  panel open (aria-expanded=true), scrollWidth 375
F2: 再看 → 120 (unique 120) → open card 100 (/listing/…2107400236560) → browser Back
  → 120 cards, ids identical in order, 「再看 1,351 件」, no reload (window.__probe), scroll restored;
  restore requests: limit=100&offset=0 + limit=20&offset=100
B3: Home 看全部價格最美 → 再看 ×2 → 80 / 120, focus on card 41 / 81 model link,
  1,391 → 1,351, 120 unique ids; history.length unchanged by load-more (3 → 3)
B9: Home → B1 → load to 120 → brand CHANEL (URL deals=1&brand=CHANEL, 40 cards, state null, 279 件,
  no new entry) → card → detail → #fair (same view, no re-render) → Back → Back → Browse CHANEL,
  switch on → Back → Home; window.__probe intact throughout
B13: keyword → Tab → sort → Tab → switch (:focus-visible, solid 2px) → Space → /browse, 全部在售, 15,367
F5: typed 1000000000000000000000 + Enter → URL max_twd=100000000, both fields 100000000;
  direct /browse?max_twd=1000000000000000000000 → field empty (default); typed 02,0000 → max_twd=20000, 2,082 件
Regression: Home → detail → model → Back → Back → Home (titles correct); advisor panel opens
Console errors: none
```

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610130938-kura-browse-page-audit.md | Parent (F1–F5) | None (record); F1 / F3 / F4 / F5 conditions met, F2 decided + fixed here; F6 still with orchestrator |
| 2610130913-kura-browse-page-plan.md | Plan whose Risks called stale `current` harmless | Risks bullet marked **Disproved** with link; Related Projex += this patch |
| 2610130524-kura-interactive-mock-sprint.md | Sprint | Not edited (orchestrator owns nav) |

---

## Notes

- `href="#"` jumps to page top natively (HTML spec: empty fragment = top), so scroll after 登入 is 0 with or without this fix; what the fix keeps is the grid, loaded pages and open panel.
- `history.state.shown` also survives reload of the same entry, so a reload re-fetches the shown count — consistent with "Back lands where you left".
- Restoring a large count (e.g. 400) issues `ceil(n/100)` parallel requests; bounded only by how far the user paged.
- No JS test harness in repo (as iterations 1–4); fragment-after-filter case is covered by the live repro above.
