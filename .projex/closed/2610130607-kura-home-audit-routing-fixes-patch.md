# Patch: Home audit routing + link + valuation-comment fixes

> **Status:** Complete
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 1
> **Author:** Claude Opus 5.5 (patch, sprint subagent)
> **Parent:** 2610130602-kura-canvas-foundation-home-audit.md
> **Directive:** Fix audit findings F2, F5, F3, F4 of 2610130602-kura-canvas-foundation-home-audit.md. F1 (價格最美 threshold) undecided → keep current ≥5% `only_deals`; F4 comment states card ±3% vs shelf ≥5% divergence factually, not as final. F1 | F6 untouched.
> **Source Plan:** Direct (audit findings; audited plan 2610130535-kura-canvas-foundation-home-plan.md)
> **Result:** Success
> **Related Projex:** 2610130602-kura-canvas-foundation-home-audit.md (source findings) | 2610130552-kura-canvas-foundation-home-patch.md (patched work) | 2610130524-kura-interactive-mock-sprint.md

---

## Summary

`spaFallback` 404'd any path whose last segment held a `.` → future model routes keyed by dotted codes (73 distinct dotted codes in local `kura.db`, e.g. `191502BZJ.04LU`) would break on reload (F2); asset dirs without `index.html` (`/js/`, `/dev/`) returned file listings (F5). Asset test now scoped: `/api/` | path under a top-level embedded dir | known asset extension → 404; dirs served only if they hold `index.html`. 價格最美 link now carries the query that produced its count (F3). `card.js` `valuation` comment now names the three coexisting thresholds and the hazard (F4, Rule 6 gap). Code commit `5a9ede3`.

---

## Changes

### Routing (F2 + F5)

**File:** `internal/web/server.go`
**Change Type:** Modified
**What Changed:**
- `spaFallback` (L73): `fs.Stat` hit on a directory without `index.html` → `http.NotFound` (was `FileServerFS` listing). Dirs with `index.html` (`.`, `legacy`) still go to `FileServerFS` → `/legacy` keeps its 301.
- Miss branch: `path.Ext(name) != ""` → `underAssetDir(root, name) || assetExt[strings.ToLower(path.Ext(name))]` (plus unchanged `/api/` prefix).
- New `assetExt` (L97) — `.js .mjs .css .map .json .html .png .jpg .jpeg .gif .webp .svg .ico .woff .woff2 .ttf .txt`; `exists` (L103); `underAssetDir` (L109) — first segment names an existing embedded directory (derived from FS, no hard-coded `js`/`dev`/`legacy` list).
- Doc comment rewritten: names rejected "has any extension" test + its dotted-model-code trap, and the directory-listing reason.

**Why:** audit F2 suggestion (scope 404 to static asset space) + F5 fold-in. Checked live DB: no dotted model-code suffix (`04lu`, `38no`, `t20`, …) collides with `assetExt`.

### Route tests

**File:** `internal/web/server_test.go`
**Change Type:** Modified
**What Changed:** `TestStaticRoutes` rows added (L23–30): `/model/191502BZJ.04LU` → 200 shell | `/dev/cards.html` → 200 | `/js/missing` → 404 | `/missing.png` → 404 | `/js/` → 404 | `/dev/` → 404 | `/legacy` → 301.

**Why:** audit § Testing Validation "Missing: dotted-path row (F2), directory row (F5)". Under the old `spaFallback` the dotted row 404s and `/js/` `/dev/` 200 — both fail.

### 價格最美 link (F3)

**File:** `internal/web/static/js/home.js`
**Change Type:** Modified
**What Changed:** `renderHome` (L34) `href="/browse"` → `href="/browse?deals=1&sort=deal"` — same filter + sort as the `api("/api/listings?deals=1&sort=deal&limit=3")` call that yields `best.total`.

**Why:** count and link target now describe the same set; browse iteration reads the filter from the URL. Threshold behind `deals=1` unchanged (≥5%, F1 open).

### Valuation comment (F4)

**File:** `internal/web/static/js/card.js`
**Change Type:** Modified
**What Changed:** `valuation` comment (L5–13) — ±3% band = legacy `valBadge` band; 價格最美 shelf, its count, advisor `only_deals` use radar `Search` `OnlyDeals` (`deal_pct >= 0.05`, excl. low-confidence + anomaly); `/api/overview` deals at `>= 0.10`; consequence (card can read 低於行情 3–4% yet be absent from 價格最美); hazard (separate constants — changing one moves none of the others). No logic change.

**Why:** audit Rule 6 gap. Wording states the divergence as fact, asserts no final threshold (F1 undecided).

---

## Verification

**Method:** `go vet ./...` | `go test -count=1 ./...` | `go test -run TestStaticRoutes -v ./internal/web` | live server on worktree `kura.db` (`127.0.0.1:8097`, temp binary removed after) + curl.

**Result:**
```
go vet ./...                         clean
ok  github.com/No3371/kura/internal/advisor
ok  github.com/No3371/kura/internal/radar
ok  github.com/No3371/kura/internal/source/brandoff
ok  github.com/No3371/kura/internal/store
ok  github.com/No3371/kura/internal/web
--- PASS: TestStaticRoutes

/                            200 text/html; charset=utf-8
/browse                      200 text/html; charset=utf-8
/model/191502BZJ.04LU        200 text/html; charset=utf-8
/listing/brandoff:1          200 text/html; charset=utf-8
/js/app.js                   200 text/javascript; charset=utf-8
/js/                         404 text/plain; charset=utf-8
/dev/                        404 text/plain; charset=utf-8
/js                          404 text/plain; charset=utf-8
/dev/cards.html              200 text/html; charset=utf-8
/js/missing.js               404 text/plain; charset=utf-8
/missing.png                 404 text/plain; charset=utf-8
/legacy                      301  http://127.0.0.1:8097/legacy/
/browse?deals=1&sort=deal    200 text/html; charset=utf-8
/api/nope                    404 text/plain; charset=utf-8

served /js/home.js: href="/browse?deals=1&sort=deal">看全部價格最美
/api/listings?deals=1&sort=deal&limit=1 → "total":1471
```

**Status:** PASS

| Finding | Resolved by | Evidence |
|---|---|---|
| F2 dotted-segment 404 | `spaFallback` asset test | `/model/191502BZJ.04LU` 200 shell (test + live) |
| F5 dir listings | dir-without-index → 404 | `/js/` `/dev/` `/js` 404 (test + live); `/legacy` 301 kept |
| F3 link drops filter | `home.js` href | served JS href = counted query; total 1471 |
| F4 Rule 6 gap | `card.js` `valuation` comment | names band source, 3 thresholds, consequence, hazard |

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|---|---|---|
| 2610130602-kura-canvas-foundation-home-audit.md | Source findings | Conditions F2 + F3 ticked w/ link here; F1 condition left open; status unchanged |
| 2610130524-kura-interactive-mock-sprint.md | Sprint | None (orchestrator owns nav) |

---

## Notes

- F1 still open: card ±3% | shelf/count/advisor ≥5% | overview ≥10%. If F1 lands on ≥10%, `card.js` `valuation` comment and `deals=1` semantics need revisiting together.
- F6 (anomaly-boundary items heading 價格最美) untouched.
- 2 live model codes contain `/` — a `/model/{code}` route must still escape them; not in scope here.
- `new-projex --projex-dir .projex/closed` scaffolded into `.projex/closed/closed/` (script appends `closed/` for patch type); moved before commit.
