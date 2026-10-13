# Audit: Kura multi-source crawl + per-pair onboarding epoch

> **Status:** Complete (Accepted)
> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 1
> **Author:** Claude Opus 5.5 (auditor, sprint subagent)
> **Parent:** 2610131843-kura-multi-source-onboarding-epoch-patch.md
> **Related Projex:** 2610131831-kura-multi-source-onboarding-epoch-plan.md (audited plan, Steps 1–5) | 2610131843-kura-multi-source-onboarding-epoch-patch.md (execution claims) | 2610131822-kura-def-coverage-sprint.md (objective; not edited) | 2610130449-kura-mock-def.md (最新上架 line edited in `514e56d`) | 2610130420-kura-roadmap-nav.md (Known Blockers R1; not edited) | 2610131037-kura-member-notifications-audit.md (R1 origin) | 2610131858-kura-onboarding-epoch-audit-fixes-patch.md (F1–F4 fixed, `3aa09fa`)
> **Audit Date:** 2026-10-13 | **Work Period:** 2026-10-13, commits `072c15d` (code) `514e56d` (patch doc) on `projex/sprint/2610131822-kura-def-coverage` (range `63eb355..514e56d`)
> **Subject:** plan Steps 1–5 vs diff; `go build`/`vet`/`test` + gofmt in worktree; 9 mutants on a scratch `git archive HEAD`; independent SQL over a `sqlite3 -readonly .backup` copy of the main checkout `kura.db` (original `kura.db`/`-wal` mtimes unchanged); scratch build of HEAD served that copy on 127.0.0.1:8097 `-advisor mock`, no crawl flags, stopped by PID; scratch removed.

---

## Audit Summary

**Claim:** (1) crawler + `cmd/kura` setup run several `crawl.Source`s — `RunAll`, `FetchDetails` per source, `BackfillGrades` for every registered `listing.Scheme`; BRAND OFF still only real source. (2) R1 fix: watched = `first_seen` strictly after own (source, target) first complete sweep `finished_at` from `crawl_runs`; 最新上架 (`newer`) + days-to-sell (`stats`) share it.

**Verdict:** Verified

**Assessment:** Completeness: High | Correctness: High | Quality: High | Value: High

**Top Issues:**
1. Strict-`After` boundary untested — mutant `!FirstSeen.Before(at)` survives every test (F1, Minor)
2. README "existing stock never reads as new" overclaims — page-shift skip in first complete sweep reads as watched, now without the old 24 h cushion; no don't-fix note in `watched` (F2, Minor)
3. Per-source detail-budget rationale lives only in plan (F3, Minor); `setup` source validation untested (F4, Minor)

---

## Claims vs Evidence

| Claim | Evidence | Status | Notes |
|-------|----------|--------|-------|
| `Crawler.Sources []Source`; `RunAll` sweeps every source/target, failure isolated, details per source | `crawler.go` diff; `TestRunAllSweepsEverySource`; mutants M7 (return on sweep error) + M8 (details only `Sources[0]`) killed | ✓ | Sweeps-then-details order kept; ctx cancel still aborts |
| One `crawl_runs` row per (source, target) with right `source` | `TestRunAllSweepsEverySource` asserts 3 keyed runs | ✓ | |
| `setup`: `[]crawl.Source`, scheme + duplicate refusal, `BackfillGrades(ctx, listing.Schemes()...)` | `cmd/kura/main.go` diff; `go build` | ✓ | Refusal paths untested (F4) |
| `Onboarded` = MIN complete `finished_at` per pair; incomplete/unfinished ignored; `Version` bumps on complete run | `store.go` diff; `TestOnboarded`; mutants M4 (drop runs from `Version`), M5 (drop `complete=1`), M6 (MAX) killed | ✓ | |
| Radar watched per pair; `newer` + `stats` share bit; no-sweep pair unwatched | `radar.go` diff; `TestWatchedPerSourceOnboarding`; M1 (never watched) killed by radar + web, M2 (fail-open) killed | ✓ | M3 (≥ instead of >) survives (F1) |
| Existing tests keep meaning | `engine` → `engineAt` with `{Source:"t"}: t0+24h`; assertions unchanged in diff | ✓ | `row()` leaves `Target` "", so pair key matches |
| `TestMemberAPI` fixture load-bearing | M1 fails `TestMemberAPI` (equivalent to no `onboard` rows) | ✓ | |
| kura.db copy: only delta = R1 listing last → first; `since` / `sold_stats_ready` / `active` unchanged | SQL: old rule (`first_seen > MIN+86400`) watched = 0 rows; new rule watched = exactly `brandoff:2101220126889` (CHANEL 17:41:20 > CHANEL complete 16:36:31); 0 listings lack a complete pair. Served HEAD: `total 15367`, first 3 = `2101220126889, 2100301255937, 2101219731339`; since 2026-10-13 00:25:25+08, ready false, active 15367 | ✓ | Matches patch `new first5`; baseline side not re-run — implied by old-rule count 0 |
| `go vet` + `go test` green; gofmt clean | Run in worktree, go1.26.0: all ok; gofmt on LF copies of 8 changed `.go`: none listed | ✓ | |
| Hygiene grep clean | Own grep on non-`.projex` diff (`projex`, `\d{10}-`, `Step \d`, `R1`, `F\d`, `:\d+`, banners, "deliberate/by design", changelog words) → no comment hits | ✓ | |
| "No deviations from plan code" | Diff vs plan After blocks: line-for-line match; extra README fetcher bullet declared | ✓ | |
| Worktree left clean | `git status --short --ignored` empty | ✓ | |

---

## Objective Verification

### Objective 1: multi-source crawl + setup

**Evidence:** `internal/crawl/crawler.go`, `internal/crawl/crawler_test.go` (new), `cmd/kura/main.go`

**Findings:**
- Actual: `Sources` loop; `SweepTarget(ctx, src, target)`, `FetchDetails(ctx, src)`; logs + error text carry source; `setup` builds list (BRAND OFF only), refuses unregistered scheme / duplicate name, backfills all schemes.
- Missing: tests for `setup` refusals (F4).
- Quality: High

**Verification:** ✓ Verified

### Objective 2: R1 per-(source, target) onboarding epoch

**Evidence:** `internal/store/store.go` (`Pair`, `Onboarded`, `Version`), `internal/radar/radar.go` (`watched`, `build`, `newer`, `stats`), tests, kura.db copy

**Findings:**
- Actual: one bit per item at `build`; both consumers read it; `Overview.Since` / `SoldStatsReady` stay on global epoch (declared out of scope).
- Missing: same-second boundary test (F1).
- Quality: High

**Verification:** ✓ Verified

---

## Code/Implementation Inspection

### `internal/store` — `Onboarded`, `Version`

**Claimed:** onboarding query + version counts complete runs
**Actual:** as claimed — Quality: High

**Issues Found:** none. `Refresh` reads `Version` before rows/onboarding, so a sweep completing mid-refresh leaves a stale-low version → next refresh rebuilds (safe direction).

**Undocumented:** none

**Metrics:** Readability: High | Test Coverage: query branches all mutated + killed | Technical Debt: None

### `internal/radar` — `watched`

**Claimed:** per-pair predicate, shared by `newer` / `stats`
**Actual:** as claimed; rationale comment names rejected global epoch + fail-closed rule — Quality: High

**Issues Found:**
- Boundary (`first_seen == finished_at` → not watched) relied on by plan Assumptions, unasserted — Severity: Low (F1)
- No don't-fix note for page-shift skip in first complete sweep (F2) — Severity: Low

**Undocumented:** none

**Metrics:** Readability: High | Technical Debt: Low

### `internal/crawl` — multi-source `Crawler`

**Claimed:** sequential multi-source sweep, per-source detail budget
**Actual:** as claimed — Quality: High

**Issues Found:**
- `DetailBudget` comment says "per source" but not why per source rather than one shared budget — Severity: Low (F3)
- One source's `FetchDetails` DB error stops later sources' details — acceptable (DB errors are store-wide); not a finding

**Metrics:** Readability: High | Test Coverage: first tests in package; 2 mutants killed | Technical Debt: None

### `cmd/kura` — `setup`

**Claimed:** source list + validation + all-scheme backfill
**Actual:** as claimed; both checks carry constraint comments — Quality: High

**Issues Found:** refusal paths untested (`cmd/kura` has no tests) — Severity: Low (F4)

---

## Testing Validation

**Coverage:** Unit: store/radar/crawl paths of the change all exercised; integration: `TestMemberAPI` via `Engine.Refresh` → `Store.Onboarded`. Edge cases: incomplete / unfinished / later-complete runs ✓, no-sweep pair ✓, failed sweep isolation ✓, same-second boundary ✗.
**Execution:** All pass? Yes | Flaky: none seen (`-count=1`)
**Mutation sample (9):** killed 8 — M1 never-watched | M2 fail-open | M4 version | M5 incomplete-counted | M6 MAX | M7 stop-on-sweep-error | M8 details-first-source-only; M9 no-op control ok. Survived: **M3** `ok && !r.FirstSeen.Before(at)`.
**Missing:** same-second boundary (F1); `setup` refusals (F4).
**Quality Issues:** none — assertions specific (ids, medians, keyed runs).

---

## Documentation Audit

**Completeness:** README sold + newest sections, fetcher bullet, flags table updated — Complete. Migration/rollback: n/a (no schema change; `Version` format change internal).
**Accuracy:** Matches implementation? Partial — "its existing stock never reads as new" (README § Newest listings) overstates (F2). UI copy `pages.js` 「開始記錄時已經在架上的商品」 stays true per shop; `member.go` comment ("the shop's newest from the initial crawl") still accurate.
**Quality:** Clarity High | Completeness High | Usability High

---

## Source Hygiene

**Scope:** Diff only (`63eb355..514e56d`, `:(exclude,glob)**/.projex/**`)

| Rule | Location (symbol) | Quote | Severity |
|------|-------------------|-------|----------|
| — | — | no rule 1–5 violations | — |

**Rule 6 gaps** (non-obvious change, no rationale comment):
- `crawl` → `Crawler.DetailBudget` / `RunAll` — budget became per source (N sources → N × `-details` fetches) — rejected shared budget and why (one shop starving another; per-site politeness) only in plan Step 3 Rationale (F3).
- `radar` → `watched` — strict boundary drops the old 24 h cushion; a listing skipped by the first complete sweep (page shift as sold items vanish mid-sweep) is first seen next sweep and reads watched. Plan Risks names it; code + README do not (F2).

**Rule 6 vacuity:** none — `watched`, `Crawler.Sources`, `setup` scheme/duplicate checks, `Onboarded` each name a rejected alternative or constraint.

**Rationale promotion:** watched rule + fail-closed no-baseline rule promoted to README § Knowing when something sold / § Newest listings ✓. Unpromoted: per-source detail budget rationale (F3). Patch has no walkthrough → no `**Rationale promoted:**` line; judged from diff.

**Commit composition** (informational — fix forward, never rewrite history): 2 landed commits checked — typed subject where required [1/1] | type matches diff where required [1/1] (`feat(crawl)`; scope narrower than diff — crawl+radar+store+cmd — approximate, acceptable) | `Projex:` trailer [1/1] (`2610131843-kura-multi-source-onboarding-epoch`). `514e56d` doc-only, exempt.
**Violations:** None

**Trailer survival on `master`:** 1/11 recent commits (whole history) carry `Projex:` — not zero → OK. Of 4 commits touching source, 1 has it (`8950a55`); `8ab31cb`, `c4e1187` are direct non-projex feat commits, `90ec1a6` baseline import. Informational.

---

## Gap Analysis

### Promised But Not Delivered
| Promise | Status | Impact |
|---------|--------|--------|
| — | none | — |

### Undocumented Issues
| Issue | Severity | Affects |
|-------|----------|---------|
| Page-shift skip in first complete sweep → watched (old rule masked it for 24 h) | Low | 最新上架 head, days-to-sell for a newly onboarded pair |
| `Overview.Since` / `SoldStatsReady` global: provider #2 inherits readiness from BRAND OFF's epoch → its sell-through shown before 14 d own history | Info (deferred by user; `MinSoldHistory` comment marks it undecided) | sprint step (c) |

### Unhandled Edge Cases
- `first_seen == finished_at` — handled (strict `After`), unasserted (F1).
- Pair whose sweeps always fail or run capped — handled fail-closed, documented README + comment.

---

## Quality Assessment

### Completeness: High
**Strengths:** every success criterion met; tests added in all three packages; README follows.
**Gaps:** F1, F4 test gaps.

### Correctness: High
**Works:** multi-source sweep: Yes | per-source details: Yes | per-pair watched: Yes | R1 on real data: Yes (independent SQL + served binary)
**Bugs:** none found

### Code Quality: High
**Positive:** small, plan-faithful diff; single predicate computed once; unexported field keeps JSON stable; comments name rejected alternatives.
**Concerns:** F2, F3 comment/doc gaps.
**Tech Debt:** none new.

### Value Delivered: High
**Intended:** make crawl → radar safe for provider #2; fix R1.
**Actual:** both; R1 listing now heads 最新上架 on today's data; no other row moves.
**Impact:** User: Positive | Sprint: unblocks steps (b)–(e)

---

## Open Findings

### Undocumented Discoveries
- Changes: none beyond declared README bullet + `newer` comment re-wrap.
- Problems: page-shift skip exposure (F2) — Evidence: README § gone rule exists because of the same page shift; plan Risks "Pagination artefact". Today's data: HERMES second complete sweep (run 8) surfaced 0 such listings → low frequency.
- Workarounds: none.

### Impact Analysis
- Downstream: `web` `sort=newest`, `member` 新上架, `advisor` `newest` + days-to-sell follow the bit, no code change needed ✓.
- Future: Enabled: provider #2 adapters drop into `sources` | Blocked: none.
- Risks: capped-only pairs never watched — Likelihood: Medium once step (e) scheduler lands; mitigated if daily full sweep kept (patch Notes).

### Improvements
- Could be better: F1–F4 below.
- Missed: none material.

---

## Stakeholder Impact

| Stakeholder | Promised | Reality | Impact |
|-------------|----------|---------|--------|
| Demo viewer | truthful 最新上架 / days-to-sell | R1 listing first; future shops' stock not flooding | Positive |
| Sprint (steps b–e) | multi-source path | `Sources` + per-pair epoch in place | Positive |

---

## Findings

### Critical (Must Address)
- none

### Significant (Should Address)
- none

### Minor (Nice to Fix)
> F1–F4 [PATCHED] via 2610131858-kura-onboarding-epoch-audit-fixes-patch.md (`3aa09fa`); I1–I3 remain open.

1. **F1 — boundary untested** (test-coverage; small patch) → add a row with `FirstSeen` == its pair's onboarding time to `TestWatchedPerSourceOnboarding`, assert it sorts with initial-crawl rows (kills M3).
2. **F2 — page-shift overclaim** (doc accuracy + Rule 6 don't-fix trap; small patch) → README § Newest listings: soften "never reads as new" (a listing a first sweep skips as pages shift reads as new when next seen); add the same hazard to `watched` comment.
3. **F3 — unpromoted per-source budget rationale** (Rule 6 gap; small patch) → `Crawler.DetailBudget` comment: rejected shared budget — one shop's backlog would starve another's; total fetches scale with source count.
4. **F4 — `setup` refusals untested** (test-coverage; small patch, optional) → extract source-list validation in `cmd/kura` to a func + table test (missing scheme, duplicate name).

### Informational (orchestrator)
- **I1** — `Overview.Since` / `SoldStatsReady` global epoch: sprint nav hazard list names R1, `dist.js` dots, advisor shop line, not this. Record against step (c) as deferred (open question per `MinSoldHistory` comment). Not code-patchable this iteration.
- **I2** — 2610130420-kura-roadmap-nav.md Known Blockers R1 still open text; resolved by `072c15d`.
- **I3** — plan header line `> Partial Execution:` lacks bold key; cosmetic.

### Positive
- Plan executed verbatim; patch claims all reproduced independently (tests, gofmt, hygiene, real-data delta).
- First tests in `internal/crawl`; tests strong — 8/9 mutants killed.
- `watched` comment is a model Rule 6 comment: rejected alternative, both failure modes, fail-closed reason.

---

## Recommendations

**Immediate:** optional `[patch]` covering F1–F3 (+F4 if cheap) — one small commit; none blocks acceptance.
**Future:** I1 before provider #2; step (e) scheduler must keep periodic full sweeps (fail-closed rule).
**Process:** plan Verification hygiene grep could include a mutation of the boundary the plan's Assumptions rely on.

---

## Final Verdict

**Status:** Accept

**Overall Assessment:**
- Completeness: High
- Correctness: High
- Quality: High
- Value: High

**Conditions:**
- [ ] none required (F1–F4 recommended)

**Sign-off:** Yes — both objectives delivered as planned, verified against tests, mutants and today's data copy; remaining items are minor test/doc gaps.
