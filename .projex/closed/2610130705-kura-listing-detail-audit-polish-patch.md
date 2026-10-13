# Patch: Listing detail audit polish (dist jitter, confidence fallback, hygiene)

> **Status:** Complete
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 2
> **Author:** Claude Opus 5.5 (patch, sprint subagent)
> **Parent:** 2610130657-kura-listing-detail-audit.md
> **Directive:** Fix findings F1, F5, F4 of 2610130657-kura-listing-detail-audit.md. F2 deferred (multi-provider iteration) — record only; F3 accepted by orchestrator (S1 holds for every URL the server accepts) — record only.
> **Source Plan:** Direct (audit findings; audited plan 2610130624-kura-listing-detail-plan.md)
> **Result:** Success
> **Related Projex:** 2610130657-kura-listing-detail-audit.md (source findings) | 2610130634-kura-listing-detail-patch.md (patched work) | 2610130524-kura-interactive-mock-sprint.md

---

## Summary

Distribution dots used fixed ±13/±26px jitter → at 375 (5 grade cols × ~60px) outer lanes of neighbouring grades overlapped (centres 7.8px apart, 10px marks) (F1). Lanes now fractions `[0, ±0.5, ±1]` of a reach `clamp(0px, (50/cols)% − 8px, 26px)` — pure CSS, no resize re-render; reach = 26px wherever a column ≥ 68px → desktop identical. `CONF` gets same `?? esc(…)` fallback as `CONF_DD`/`METHOD` (F5). Hygiene: `kura.css` forward ref → present tense; `historyHTML` clamp gets rationale (F4). F2 Rule-6 comment added on `scheme.source` stat filter (audit: "add now if touching the file"); keying fix itself deferred. Code commit `1dab0e9`.

---

## Changes

### Distribution jitter (F1) + F2 trap comment

**File:** `internal/web/static/js/dist.js`
**Change Type:** Modified
**What Changed:**
- L12–19: `OFFS` px array → `LANES = [0, -0.5, 0.5, -1, 1]` + `LANE_MAX = 26`; comment names rejected alternatives (fixed px lanes → overlap in narrow cols; per-offset clamp → lanes merge) and why CSS not JS (resize needs no re-render)
- L64–65: `reach = clamp(0px, ${50/cols.length}% - 8px, 26px)` — `%` in `left` resolves against plot width, so `(50/n)%` = half a column; 8px = 7px dot radius (10px mark + 2px halo) + 1px
- L72: `left = calc(cx% + lane * reach)`
- L44–48: Rule-6 comment on `col.stat` `scheme.source` filter vs cross-source `col.pts` match (F2 trap named; `{source, code}` key needed before a second shop)

**Why:** F1 legibility at 375; F2 comment recommended by audit when touching file.

### Confidence fallback (F5) + history clamp rationale (F4)

**File:** `internal/web/static/js/detail.js`
**Change Type:** Modified
**What Changed:**
- L76 `pricingHTML`: `CONF[f.confidence]` → `CONF[f.confidence] ?? esc(f.confidence)` — matches existing `CONF_DD`/`METHOD` fallback pattern (audit suggested `?? ""`; raw value chosen for consistency with the 信心 `dd` in the same page)
- L176–179 `historyHTML`: comment on `Math.min(100, …)` — observation after end time (post-`gone_at` or server clock ahead) pinned to right edge; rejected: stretching `t1`, which moves line end off the day its end label names

### CSS block header (F4)

**File:** `internal/web/static/kura.css`
**Change Type:** Modified
**What Changed:**
- L277: `/* 同型號價格分布 (shared with the model page) */` → `/* 同型號價格分布 (markup from dist.js distHTML) */` (Rule 3: no forward reference to unbuilt page)

---

## Verification

**Method:** `go vet ./...` + `go test -count=1 ./...`; hygiene grep over touched files; built binary served a scratchpad copy of `kura.db` (worktree DB untouched — mtimes 05:25 unchanged); built-in browser on `/listing/brandoff:2106800650082` (CHANEL マトラッセ, 409-dot group, 5-grade ladder, this listing in ランクA), dark theme. In-page script measured every dot centre vs plot, per occupied column the outermost halo/mark extents, and each dot's centre vs the **old** formula `cx% + OFFS[j]px` (lane parsed from raw `style` attr).

**Result:**
```
375x812: plotW 299 (5 cols → 59.8px each), dots 409 = 289 ランクA + 120 ランクB, parseFail 0
  ランクA|ランクB  haloGap 2px  markGap 6px   (before: outer centres 7.8px apart → 10px marks overlapped)
  scrollWidth 375; hover dot → tip 「ランクA · 在售 NT$52,478」, tip centre x = dot centre x (210)
1280x900: plotW 836 (167px cols), dots 409
  max |new centre − old-formula centre| = 0.01px  → desktop rendering unchanged
  ランクA|ランクB haloGap 101px
Fair link: 「行情 NT$100,919 · 低信心」 (known value path intact)
CONF fallback eval: low→低信心 | bogus→bogus | undefined→""
Console errors: none
go vet: clean; go test: ok advisor radar brandoff store web
hygiene grep (projex|<10-digit id>-|file:ln|Step N) on touched files: none (exit 1)
```

Screenshot at 375 (dark): ランクA / ランクB dot clouds visibly separated, 此件 label clear of neighbours.

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|---|---|---|
| 2610130657-kura-listing-detail-audit.md | Source findings | F1/F4/F5 marked resolved w/ link + `1dab0e9`; F2 deferred + comment landed; F3 accepted by orchestrator; condition ticked; Related Projex adds this patch; status left to orchestrator |
| 2610130524-kura-interactive-mock-sprint.md | Sprint | None (orchestrator owns nav) |

---

## Notes

- **F2 (deferred):** dots still match grade codes across sources; harmless with one shop. Fix = column key `{source, code}` in multi-provider iteration; model page inherits same `distHTML`.
- **F3 (accepted, no code):** direct load of a malformed escape (`/listing/%E0%A4%A`) → Go `net/http` 400 before any handler; S1 reads "every URL the server accepts".
- `.k-dot--this` (14px mark + halo → r9) at an outermost lane can touch a neighbour's halo at 375; marks stay ≥ 4px apart (16px centre gap − 7 − 5). Accepted: halo is paper-coloured.
- Audit Info items (tip `role="status"` inside `role="img"`) untouched — out of directive.
