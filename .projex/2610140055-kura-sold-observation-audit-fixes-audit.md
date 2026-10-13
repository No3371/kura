# Audit: Kura Sold Observation Audit Fixes (F1–F4)

> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 5
> **Status:** Complete (Accept)
> **Author:** Claude Opus 5.5 (auditor, sprint subagent, uid `2610131822-kura-def-coverage#5`)
> **Parent:** 2610140046-kura-sold-observation-audit-fixes-patch.md
> **Related Projex:** 2610140046-kura-sold-observation-audit-fixes-patch.md (audited patch) | 2610140037-kura-sold-observation-audit.md (findings F1–F4 source) | 2610140023-kura-sold-observation-patch.md | 2610140010-kura-sold-observation-plan.md | 2610131822-kura-def-coverage-sprint.md (not edited)
> **Audit Date:** 2026-10-14 | **Work Period:** 2026-10-14, range `635c90e..0052b99` on `projex/sprint/2610131822-kura-def-coverage` (`61299b0` code + tests, `78831b4` docs/comments/help, `0052b99` patch doc)
> **Subject:** does the patch resolve F1–F4 without regressions? In scope: F1–F4. Out of scope (not counted): F5, F6, F7, owner-held items (ALLU consignments, 第18条 (12) reading, stop-on-403, GALLERY RARE paging).
> **Method:** No request to any shop host. Diff read in full (`.projex/` excluded). `go vet ./...` + `go test -count=1 ./...` + `-v` on `brandoff` in the worktree (read-only; Go cache only). gofmt on committed content of all 44 tracked `.go` files (`git show HEAD:<f> | gofmt -l`). Fixture `list_lv.html` scanned for soldout markers and stock filter state. Tree of `0052b99` exported (`git archive`) to the session scratchpad: 9 mutation probes on `brandoff` there. Scratch removed; worktree left clean.

---

## Audit Summary

**Claim:** F1 BRAND OFF `SOLDOUT` tiles → `ListPage.Sold` → bag sold ids to `visit` (never stored), row count incl. sold tiles; F2 volume wording states cap + steady state; F3 stale capped-sweep text fixed; F4 `MarkGone` flap hazard promoted to code comment.

**Verdict:** Verified — F1, F2, F4 fully resolved; F3 resolved for every named location but one sentence (residual N1, Low). No regression found: full suite green, 7/9 mutations killed, the 2 survivors are the one unpinned branch (N2).

**Assessment:** Completeness: High | Correctness: High | Quality: High | Value: High (BRAND OFF sold tiles now observed → all 3 providers carry a sold signal)

**Top Issues:**
1. **N2** `offset+rows >= total` half of the last-page change is unpinned (mutations A6, A9 survive) and rests on an unstated assumption — BRAND OFF's hit count includes SOLDOUT tiles. Fixture evidence supports it. Low.
2. **N1** README § Knowing when something sold keeps "Kura marks an item **gone** only after it is missing from **two consecutive complete sweeps**" — named in F3, left literal; mitigated two lines later. Low.
3. **N3** Patch doc presents the page-count trap as "beyond the auditor's suggested fix"; the audit's F1 remediation already said "keep the raw row count for `Sweep`'s last-page test". Trap real, handling correct; novelty overstated. Informational.

---

## Claims vs Evidence

| Claim | Evidence | Status | Notes |
|-------|----------|--------|-------|
| Fixture holds exactly 7 `.product__item--soldout` tiles among 90; pids = audit list | `grep 'class="[^"]*soldout[^"]*"' list_lv.html` → 7 `product__item--soldout` + 7 nested `--soldout-text`; `TestParseList` 83 + 7, `TestParseListSold` ids | ✓ | class selector matches only the exact token; `-text` spans sit inside the same 7 tiles |
| `ParseList` routes soldout tiles to `Sold`, fully parsed (so `IsBag` works) | `brandoff.ParseList` — check after `applyTitle`, `return` skips `Cards` | ✓ | A4 (split off), A5 (tile in both) killed |
| `Sweep` passes bag sold ids only | `brandoff.Source.Sweep` `lp.Sold` loop | ✓ | A2 (`nil`), A3 (bag filter dropped), A7 (URL in place of id) killed |
| Page-count trap real: cards-only count reads a full page with sold tiles as last page | `Sweep`: `offset := (page-1)*pageSize`, stop on `rows < pageSize`; 83 < 90 → complete after page 1 → `SweepMissing` false misses | ✓ | A1 (`rows := len(lp.Cards)`) and A8 (short-page test dropped) killed by `TestSweepSoldTiles` |
| "Fix counts `len(Cards)+len(Sold)` in both the `< pageSize` and `>= total` tests; a new Sweep test pins it" | A6 (`offset+len(lp.Cards) >= total`) and A9 (total test removed) both survive | ⚠ | test pins the short-page half only (N2) |
| Sold ids map to held listing ids | `crawl.SweepTarget`: `listing.Listing{Source: run.Source, SourceID: sid}.ID()`; `Name()` = `SourceName`; card `SourceID = pid` | ✓ | same id scheme as stored BRAND OFF listings |
| Repeated SOLDOUT tile → no flap | `store.MarkGone` `WHERE … status='active'`; sold ids never reach `RecordSeen` | ✓ | |
| No new request paths | sold tiles come from list pages already fetched; `Detail`, `DetailBudget`, `RecheckAfter` untouched in diff | ✓ | |
| `baseURL` const → var, test-only | `git grep baseURL -- internal/source/brandoff` → def, `listURL`, `source_test.go` only | ✓ | `TestSweepSoldTiles` serves via `httptest`; no shop host contacted |
| No other `ListPage`/`ParseList` consumer outside package | `git grep 'brandoff\.\(ParseList\|ListPage\|Card\)'` → projex docs only | ✓ | |
| F2: "than before" / "(no added requests)" dropped; cap + steady state stated | `Crawler.RecheckAfter`, README § Knowing when something sold, vetting GR crawl limits + ALLU Outcome bullet | ✓ | 150 = `-details` default; 600/day = 4 × 150 at `-crawl-every 6h` ✓ |
| F3: `-max-pages` help, flags row, README capped-sweep sentence, `Crawler.MaxPages` | diff | ⚠ | all fixed; "only after … two consecutive complete sweeps" sentence left (N1) |
| F4: hazard in `store.MarkGone` doc | `MarkGone` comment vs `RecordSeen` (`status='active', gone_at=NULL`), `PendingDetails` (`detail_at IS NULL`), `MarkGone` (`gone_at=?` on active) | ✓ | mechanism stated matches code |
| vet / test / gofmt green | re-run | ✓ | vet exit 0; 10 packages ok; 44/44 gofmt-clean |
| Rationale promoted (F7 of prior audit answered) | patch header `**Rationale promoted:**` line; targets exist in diff | ✓ | |

---

## Objective Verification

### F1 — BRAND OFF `SOLDOUT` tiles
**Evidence:** `61299b0`; `brandoff.ListPage.Sold`, `brandoff.ParseList`, `brandoff.Source.Sweep`; `TestParseList`, `TestParseListSold`, `TestSweepSoldTiles`; README SOLDOUT bullet (`78831b4`).
**Findings:** Actual: as claimed; `crawl.Source.Sweep` contract ("nil when the shop's list pages show none") now true for all 3 adapters. Rule-6 gap on `brandoff.Source.Sweep` closed (names the cards-only alternative + consequence). Missing: pin + rationale for the `>= total` half (N2). Quality: High.
**Verification:** ✓ Verified

### F2 — request-volume wording
**Evidence:** `Crawler.RecheckAfter`; README re-check bullet; `docs/provider-vetting.md` GR crawl limits + ALLU Outcome bullet.
**Findings:** Actual: "stays within `-details` … while re-check candidates exist it can use the whole cap (150 … up to 600 a day at `-crawl-every 6h`)". Accurate; rule-3 framing gone. Owner's 第18条 (12) reading now gets the steady-state number. Quality: High.
**Verification:** ✓ Verified

### F3 — stale capped-sweep claims
**Evidence:** `cmd/kura` `-max-pages` help; README flags row; README § Knowing when something sold; `Crawler.MaxPages` (beyond audit's list — good).
**Findings:** Actual: "capped sweeps never count unseen items toward sold"; "(the sold signals below still apply)"; "missing from a sweep never marks their items sold. Other signals do:". Missing: the opening "only after … two consecutive complete sweeps" — explicitly named in the prior audit's Documentation Audit — still reads as the sole gone path (N1). Quality: Medium-High.
**Verification:** ⚠ Partial (residual Low)

### F4 — `MarkGone` flap hazard
**Evidence:** `store.MarkGone` doc.
**Findings:** Actual: names trap (first-read-gone + list still showing → reactivate → refetch → `gone_at` moves forward, one first fetch per crawl), the mitigation (adapters route shop-marked sold tiles to sold ids) and the residual (unmarked sold item still flaps). Matches `RecordSeen` / `PendingDetails` / `MarkGone` SQL. Quality: High.
**Verification:** ✓ Verified

---

## Code/Implementation Inspection

### `internal/source/brandoff`
**Claimed:** split sold tiles, pass bag sold ids, row-count fix, test seam. **Actual:** matches — Quality: High.
**Issues Found:**
- `offset+rows >= total` unpinned; correctness depends on hit count including SOLDOUT tiles (N2) — Severity: Low
**Undocumented:** none. Sold tiles skip the `c.Brand = brand.Name` default — irrelevant (only `SourceID` is used).

### `internal/crawl/crawler.go`, `internal/store/store.go`, `cmd/kura`
**Claimed:** comment/help text only. **Actual:** comment/help text only (no logic lines changed) — Quality: High.

**Metrics:** Readability/Maintainability: High | Test Coverage: 7/9 mutations killed; survivors confined to the `>= total` term | Technical Debt: Low

---

## Testing Validation

**Coverage:** mutation probes on `0052b99` scratch copy (`go test ./internal/source/brandoff/`):

| # | Mutation | Result |
|---|----------|--------|
| A1 | `rows := len(lp.Cards)` | killed — `TestSweepSoldTiles` |
| A2 | `visit(page, bags, nil)` | killed — `TestSweepSoldTiles` |
| A3 | bag filter on sold dropped | killed — `TestSweepSoldTiles` |
| A4 | soldout split disabled | killed — `TestParseList`, `TestParseListSold`, `TestSweepSoldTiles` |
| A5 | sold tile also appended to `Cards` | killed — same three |
| A6 | `offset+len(lp.Cards) >= total` | **survives** |
| A7 | sold ids = `c.URL` | killed — `TestSweepSoldTiles` |
| A8 | `rows < pageSize` dropped | killed — `TestSweepSoldTiles` |
| A9 | `\|\| offset+rows >= total` dropped | **survives** |

A1–A4 reproduce the executor's M1–M4.
**Execution:** All pass: Yes | Flaky: none seen.
**Missing:** a sweep test where the hit count ends the sweep on a full page holding sold tiles (e.g. Total = 90 on `list_lv.html`-shaped page 1) — would kill A6/A9.
**Quality Issues:** none; `TestSweepSoldTiles` swaps a package var without `t.Parallel` — safe as written.

---

## Documentation Audit

**Completeness:** README + vetting + code comments — Complete for F1/F2/F4; F3 one sentence short (N1).
**Accuracy:** Yes, except N1. "BRAND OFF keeps sold items on its list pages for a while" — duration unknown offline, phrased softly; fine.
**Quality:** Clarity High | Completeness High | Usability High. Cosmetic: README lines 131/146 and vetting ALLU Outcome line not re-wrapped (~110–140 cols vs ~90 elsewhere).

---

## Source Hygiene

**Scope:** Diff only (`635c90e..0052b99`, `.projex/` excluded)

| Rule | Location (symbol) | Quote | Severity |
|------|-------------------|-------|----------|
| — | — | no rule 1–5 violation in added/modified comments | — |

No projex ids, `file:ln`, plan banners, changelog narration; "than before" (prior rule-3 finding) removed. `brandoff.baseURL` comment ("only so tests can point the sweeper at a local server; nothing else assigns it") names its constraint — fine.

**Rule 6 gaps:**
- `brandoff.Source.Sweep` — `offset+rows >= total` counts sold tiles against the site's hit count; the comment justifies including sold tiles only for the short-page test. A reader cannot tell the total test assumes `.productlist__itemcount` counts SOLDOUT tiles; if it did not, a sweep could complete one page early and `SweepMissing` would count that page's listings missed (N2). Fixture: both 在庫あり / 在庫なし stock filters unchecked, so the count is of the unfiltered result set that displays the sold tiles — supports the assumption.

**Rule 6 vacuity:** none — `ListPage.Sold`, `Source.Sweep`, `MarkGone`, `RecheckAfter`, `MaxPages` each name a rejected alternative or trap.

**Rationale promotion:** patch header `**Rationale promoted:**` names `brandoff.ListPage.Sold`, `brandoff.Source.Sweep`, README, `crawl.Crawler.RecheckAfter`, vetting, `store.MarkGone` — all present in diff. Supported.

**Commit composition** (informational — fix forward, never rewrite history): 3 landed commits — typed subject where required 2/2 (`fix(brandoff)`, `docs(crawl)`; `0052b99` doc-only, exempt) | type matches diff 2/2 (`78831b4` touches comments + one flag help string — docs fits) | `Projex:` trailer 3/3 (`2610140046-kura-sold-observation-audit-fixes`)
**Violations:** None

**Trailer survival on `master`:** 1/11 (master holds 11 commits; sprint not landed) — not zero; informational, unchanged since prior audit.

---

## Gap Analysis

### Promised But Not Delivered
| Promise | Status | Impact |
|---------|--------|--------|
| F3 — no live text claiming the old sold behaviour | Partial (N1) | Low |
| "a new Sweep test pins" both last-page terms | Partial (N2) | Low |

### Undocumented Issues
| Issue | Severity | Affects |
|-------|----------|---------|
| N2 total-test assumption unstated + unpinned | Low | BRAND OFF sweep completion, `SweepMissing` |
| N4 pre-patch rows stored from SOLDOUT tiles get `gone_at` = first post-patch crawl | Low (informational) | time-to-sell for those rows in an existing DB |

### Unhandled Edge Cases
- BRAND OFF hit count excluding SOLDOUT tiles (unverifiable offline) → early completion by ≤ 1 page → false misses (N2). Pre-patch code (cards-only total test) erred the other way: ≤ 1 extra list request.
- Sold item shown without a SOLDOUT marker still flaps (stated in `MarkGone` comment; F1 fix removes known trigger).

---

## Quality Assessment

### Completeness: High
**Strengths:** all four findings addressed; `Crawler.MaxPages` stale comment caught beyond the audit's list; rationale-promoted line added.
**Gaps:** N1, N2.

### Correctness: High
**Works:** sold split: Yes | bag-only sold ids: Yes | full-page detection with sold tiles: Yes | no flap on repeated SOLDOUT: Yes | no new requests: Yes
**Bugs:** none found. N2 is an assumption risk, not an observed bug.

### Code Quality: High
**Positive:** minimal change on the existing contract; test seam mirrors `allu.baseURL`; comments name rejected alternatives.
**Concerns:** N2.
**Tech Debt:** Low.

### Value Delivered: High
**Intended:** BRAND OFF sold tiles observed; docs accurate on volume and capped sweeps. **Actual:** delivered. **Impact:** User: Positive (6 of 90 fixture bags no longer shown 在售 / used as active comps / radar deals). Business: sprint's ≥3-providers sold-observation clause now backed by code + fixture for BRAND OFF.

---

## Open Findings

### Undocumented Discoveries
- Changes: none beyond the patch doc's list.
- Problems: N4 — an existing DB that stored SOLDOUT tiles as listings before `61299b0` marks them gone on the next crawl that shows the tile, with `gone_at` = that crawl, not the sale. Any such row counted as "seen appearing" yields a wrong time-to-sell once. Self-limiting; no migration needed for a demo DB — Evidence: `SweepTarget` sold loop + `MarkGone` stamps `now`.
- Adjacent (pre-existing, outside F3's list): `allu.Source.Sweep` comment "sweeps never mark its items gone" refers to missed-sweep marking; the same sentence goes on to name sold rows — self-qualifying, not counted.

### Impact Analysis
- Downstream: `radar`/`web` — fewer active BRAND OFF comps (sold tiles excluded); all tests green.
- Future: Enabled: sprint absorb can record sold signals for all 3 providers.
- Risks: N2 hit-count semantics — Likelihood: Low (fixture's unfiltered stock state).

### Improvements
- Could be better: pin the total test (Total = 90 single full page with sold tiles → complete after page 1, one request) and state the hit-count assumption in `Source.Sweep`.

---

## Stakeholder Impact

| Stakeholder | Promised | Reality | Impact |
|-------------|----------|---------|--------|
| Demo viewer | BRAND OFF sold bags stop showing 在售 | Yes — SOLDOUT tiles mark held listings gone, new ones never stored | Positive |
| Shops (crawl politeness) | no new requests | same list pages read; product-page budget untouched | Neutral |
| Owner (第18条 (12) reading) | vetting states real volume | cap + steady state (150/crawl, 600/day at 6h) stated in both GR and ALLU bullets | Positive |

---

## Findings

### Critical (Must Address)
- None.

### Significant (Should Address)
- None.

### Minor (Nice to Fix)
- **N1 — F3 residual sentence** (Low) — README § Knowing when something sold: "Kura marks an item **gone** only after it is missing from **two consecutive complete sweeps**" — literal "only" contradicted by ALLU sold rows, BRAND OFF SOLDOUT tiles and re-checks; next sentence's "(the sold signals below still apply)" mitigates. → reword to scope it to vanishing listings ("A listing that merely vanishes is marked gone only after …").
- **N2 — total-test half unpinned, assumption unstated** (Low) — A6/A9 survive; `brandoff.Source.Sweep` comment justifies sold tiles in the row count only for the short-page test. → add a sweep test where Total ends the sweep on a full page with sold tiles; one comment clause: the hit count covers the unfiltered result set incl. SOLDOUT tiles.
- **N3 — patch doc novelty claim** (informational) — trap verified real and handled; prior audit's F1 remediation already said keep the raw row count. No fix needed.
- **N4 — historic SOLDOUT rows** (informational) — one-time `gone_at` skew in a pre-existing DB. No fix needed for the demo.
- Cosmetic: re-wrap README lines 131/146 and the vetting ALLU Outcome line.

### Positive
- Page-count trap caught and pinned (A1, A8 killed); without it a full BRAND OFF page with sold tiles would end every sweep at page 1 and mark everything past it gone within two crawls.
- 7/9 probes killed; executor's M1–M4 reproduced.
- `Crawler.MaxPages` stale comment fixed though unnamed by the audit.
- `**Rationale promoted:**` line present (answers prior F7).

---

## Recommendations

**Immediate:** none blocking. Prior audit conditions met: F1 fixed (condition 1); F2 fixed before the owner's 第18条 (12) reading (condition 2).
**Future:** N1 + N2 together are a ~10-line follow-up, offline, fixture-backed — fold into a later patch or finalize clean-up.
**Process:** when a fix changes two terms of one predicate, the verification mutation set should probe each term separately.

---

## Final Verdict

**Status:** Accept

**Overall Assessment:**
- Completeness: High
- Correctness: High
- Quality: High
- Value: High

**Conditions:**
- None blocking. Optional: N1, N2.

**Sign-off:** Yes — F1–F4 resolved (F3 with one Low residual sentence); no regression in code, tests, budget or request paths; remaining items Low/informational.
